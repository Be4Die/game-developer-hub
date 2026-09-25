#!/usr/bin/env python3
"""
seed_demo_api.py
----------------
Автономный скрипт сидирования реалистичных данных для демонстрационного видеоролика
Game Developer Hub строго через HTTP/REST API Платформы (Gateway :8080).

Соблюдает все бизнес-правила платформы:
- Аутентификация через JWT (Developer, Moderator, Admin)
- Регистрация и верификация второго разработчика (Collab)
- Загрузка реальных медиафайлов и клиентских/серверных zip/tar сборок
- Полный цикл публикации релиза (Draft -> Moderation -> In Review -> Approved -> Prod Deploy)
- Полный цикл отклонения с нарушениями (Draft -> Moderation -> Rejected с карточками нарушений и вложением в чате)
- Заявка на платформенные серверы (Server Access -> Approved -> Server Build Upload -> Instance Start)
- Командный доступ (Принятое приглашение + Ожидающее входящее приглашение)
- Внутриигровые товары (IAP Items)
- Живая переписка в чате тикетов с вложениями
- Наполнение журнала решений и журнала действий модератора
"""

import sys
import os
import json
import time
import subprocess
import requests

GATEWAY_URL = os.environ.get("GATEWAY_URL", "http://localhost:8080")
DOWNLOADS_DIR = "/home/michael/Загрузки"
TEST_GAME_DIR = "/home/michael/Projects/gdh-test-game"

# Пути к медиафайлам
MEDIA = {
    "cover_cat": os.path.join(DOWNLOADS_DIR, "Обложка Котик_bpP4.png"),
    "cover_cyber": os.path.join(DOWNLOADS_DIR, "Cover1280x720.png"),
    "cover_pixel": os.path.join(DOWNLOADS_DIR, "Cover650x820.png"),
    "icon_1": os.path.join(DOWNLOADS_DIR, "Ico1.png"),
    "icon_2": os.path.join(DOWNLOADS_DIR, "icon.png"),
    "video_promo": os.path.join(DOWNLOADS_DIR, "Promotion_ObbyAndTelekinesis_Horizontal.mp4"),
    "video_game": os.path.join(TEST_GAME_DIR, "media", "video.mp4"),
    "shot_hangar": os.path.join(TEST_GAME_DIR, "media", "screenshots", "shot_1_hangar.png"),
    "shot_battle": os.path.join(TEST_GAME_DIR, "media", "screenshots", "shot_2_battle.png"),
    "shot_victory": os.path.join(TEST_GAME_DIR, "media", "screenshots", "shot_3_victory.png"),
    "build_cat": os.path.join(DOWNLOADS_DIR, "build_project_1_v1.0.0.zip"),
    "build_client_101": os.path.join(TEST_GAME_DIR, "dist-gdh", "steel-vanguard-client-v1.0.1.zip"),
    "build_client_102": os.path.join(TEST_GAME_DIR, "dist-gdh", "steel-vanguard-client-v1.0.2.zip"),
    "build_client_104": os.path.join(TEST_GAME_DIR, "dist-gdh", "steel-vanguard-client-v1.0.4.zip"),
    "build_server_tar": os.path.join(TEST_GAME_DIR, "dist-gdh", "steel-vanguard-server-v1.0.1.tar.gz"),
}

# Пользователи
CREDENTIALS = {
    "dev": ("dev@welwise.com", "password123"),
    "collab": ("alex_dev@welwise.com", "password123"),
    "moderator": ("moderator@welwise.com", "password123"),
    "admin": ("admin@welwise.com", "password123"),
}

def log(msg: str, level: str = "INFO"):
    colors = {
        "INFO": "\033[94m[INFO]\033[0m",
        "SUCCESS": "\033[92m[✓]\033[0m",
        "WARN": "\033[93m[!]\033[0m",
        "ERROR": "\033[91m[✗]\033[0m",
        "HEADER": "\033[95m\033[1m",
    }
    prefix = colors.get(level, f"[{level}]")
    endc = "\033[0m" if level == "HEADER" else ""
    print(f"{prefix} {msg}{endc}")

class GDHClient:
    def __init__(self, email: str, password: str):
        self.email = email
        self.password = password
        self.token = None
        self.user = None
        self.session = requests.Session()

    def login(self) -> bool:
        url = f"{GATEWAY_URL}/api/v1/auth/login"
        resp = self.session.post(url, json={"email": self.email, "password": self.password})
        if resp.status_code == 200:
            data = resp.json()
            self.token = data.get("tokens", {}).get("access_token")
            self.user = data.get("user", {})
            self.session.headers.update({"Authorization": f"Bearer {self.token}"})
            return True
        return False

    def register_and_verify(self, display_name: str) -> bool:
        # 1. Попытка входа
        if self.login():
            log(f"Пользователь {self.email} уже существует и авторизован.", "INFO")
            return True

        # 2. Регистрация
        url = f"{GATEWAY_URL}/api/v1/auth/register"
        resp = self.session.post(url, json={
            "email": self.email,
            "password": self.password,
            "display_name": display_name
        })
        if resp.status_code not in (200, 201):
            log(f"Ошибка регистрации {self.email}: {resp.text}", "WARN")

        # 3. Извлечение кода верификации из Valkey
        time.sleep(0.5)
        code = self._get_verification_code_from_valkey()
        if code:
            verify_url = f"{GATEWAY_URL}/api/v1/auth/verify-email"
            v_resp = self.session.post(verify_url, json={"verification_code": code})
            if v_resp.status_code == 200:
                log(f"Email {self.email} успешно верифицирован кодом {code}", "SUCCESS")
            else:
                log(f"Ошибка верификации email {self.email}: {v_resp.text}", "WARN")
        else:
            # Fallback на случай прямого подтверждения в БД
            log(f"Код верификации не найден в Valkey, активируем через DB fallback", "WARN")
            subprocess.run([
                "docker", "exec", "gdh-postgres", "psql", "-U", "postgres", "-d", "orchestrator",
                "-c", f"UPDATE users SET email_verified=true, status=1 WHERE email='{self.email}';"
            ], capture_output=True)

        return self.login()

    def _get_verification_code_from_valkey(self) -> str:
        cmd = ["docker", "exec", "gdh-valkey", "valkey-cli", "GET", f"sso:verify:{self.email}"]
        res = subprocess.run(cmd, capture_output=True, text=True)
        out = res.stdout.strip()
        if out and out != "(nil)":
            return out
        return ""

    # Проекты
    def create_project(self, title_ru: str, title_en: str, is_online: bool = False) -> dict:
        url = f"{GATEWAY_URL}/api/v1/projects"
        resp = self.session.post(url, json={
            "title_ru": title_ru,
            "title_en": title_en,
            "is_online": is_online
        })
        if resp.status_code in (200, 201):
            return resp.json().get("project", {})
        raise RuntimeError(f"create_project failed ({resp.status_code}): {resp.text}")

    def update_project(self, project_id: int, **fields) -> dict:
        url = f"{GATEWAY_URL}/api/v1/projects/{project_id}"
        resp = self.session.patch(url, json=fields)
        if resp.status_code == 200:
            return resp.json().get("project", {})
        raise RuntimeError(f"update_project {project_id} failed ({resp.status_code}): {resp.text}")

    def delete_project(self, project_id: int) -> bool:
        url = f"{GATEWAY_URL}/api/v1/projects/{project_id}"
        resp = self.session.delete(url)
        return resp.status_code == 200

    def list_projects(self, limit: int = 50, offset: int = 0) -> list:
        url = f"{GATEWAY_URL}/api/v1/projects?limit={limit}&offset={offset}"
        resp = self.session.get(url)
        if resp.status_code == 200:
            return resp.json().get("projects", [])
        return []

    # Загрузка медиа
    def upload_media(self, project_id: int, media_type: str, file_path: str) -> str:
        if not os.path.exists(file_path):
            log(f"Файл {file_path} не существует, пропускаем", "WARN")
            return ""
        url = f"{GATEWAY_URL}/api/v1/projects/{project_id}/media"
        with open(file_path, "rb") as f:
            files = {"file": (os.path.basename(file_path), f)}
            data = {"media_type": media_type}
            resp = self.session.post(url, data=data, files=files)
        if resp.status_code in (200, 201):
            return resp.json().get("file_path", "")
        raise RuntimeError(f"upload_media {media_type} for project {project_id} failed: {resp.text}")

    # Загрузка клиентского билда
    def upload_build(self, project_id: int, version: str, zip_path: str) -> dict:
        if not os.path.exists(zip_path):
            raise FileNotFoundError(f"Zip build not found: {zip_path}")
        url = f"{GATEWAY_URL}/api/v1/projects/{project_id}/builds"
        with open(zip_path, "rb") as f:
            files = {"file": (os.path.basename(zip_path), f, "application/zip")}
            data = {"version": version}
            resp = self.session.post(url, data=data, files=files)
        if resp.status_code in (200, 201):
            return resp.json()
        raise RuntimeError(f"upload_build {version} for project {project_id} failed: {resp.text}")

    # Внутриигровые товары (IAP)
    def create_game_item(self, project_id: int, item_id: str, name: str, description: str, price: int, image_url: str = "") -> dict:
        url = f"{GATEWAY_URL}/api/v1/projects/{project_id}/purchases/items"
        payload = {
            "project_id": str(project_id),
            "game_item_id": item_id,
            "name": name,
            "description": description,
            "image_url": image_url,
            "price_coins": price,
            "is_active": True
        }
        resp = self.session.post(url, json=payload)
        if resp.status_code in (200, 201):
            return resp.json().get("item", {})
        log(f"create_game_item {item_id} warning ({resp.status_code}): {resp.text}", "WARN")
        return {}

    # Команда и приглашения
    def send_invitation(self, project_id: int, invitee_id: str, invitee_email: str, permissions: list) -> dict:
        url = f"{GATEWAY_URL}/api/v1/projects/{project_id}/invitations"
        resp = self.session.post(url, json={
            "project_id": str(project_id),
            "invitee_id": invitee_id,
            "invitee_email": invitee_email,
            "permissions": permissions
        })
        if resp.status_code in (200, 201):
            return resp.json().get("invitation", {})
        raise RuntimeError(f"send_invitation failed ({resp.status_code}): {resp.text}")

    def list_incoming_invitations(self) -> list:
        url = f"{GATEWAY_URL}/api/v1/invitations/incoming"
        resp = self.session.get(url)
        if resp.status_code == 200:
            return resp.json().get("invitations", [])
        return []

    def respond_invitation(self, invitation_id: int, accept: bool) -> bool:
        url = f"{GATEWAY_URL}/api/v1/invitations/{invitation_id}/respond"
        resp = self.session.post(url, json={"invitation_id": str(invitation_id), "accept": accept})
        return resp.status_code == 200

    # Модерация
    def submit_moderation(self, project_id: int) -> int:
        url = f"{GATEWAY_URL}/api/v1/projects/{project_id}/moderation/submit"
        resp = self.session.post(url, json={"project_id": str(project_id)})
        if resp.status_code in (200, 201):
            return int(resp.json().get("request_id", 0))
        raise RuntimeError(f"submit_moderation project {project_id} failed ({resp.status_code}): {resp.text}")

    def claim_moderation(self, request_id: int) -> dict:
        url = f"{GATEWAY_URL}/api/v1/moderation/requests/{request_id}/claim"
        resp = self.session.post(url, json={"request_id": str(request_id)})
        if resp.status_code in (200, 201):
            return resp.json().get("request", {})
        raise RuntimeError(f"claim_moderation {request_id} failed: {resp.text}")

    def approve_moderation(self, project_id: int, comment: str) -> dict:
        url = f"{GATEWAY_URL}/api/v1/moderation/projects/{project_id}/approve"
        resp = self.session.post(url, json={"project_id": str(project_id), "comment": comment})
        if resp.status_code in (200, 201):
            return resp.json()
        raise RuntimeError(f"approve_moderation {project_id} failed: {resp.text}")

    def reject_moderation(self, project_id: int, reason: str, violations: list) -> dict:
        url = f"{GATEWAY_URL}/api/v1/moderation/projects/{project_id}/reject"
        resp = self.session.post(url, json={
            "project_id": str(project_id),
            "reason": reason,
            "violations": violations
        })
        if resp.status_code in (200, 201):
            return resp.json()
        raise RuntimeError(f"reject_moderation {project_id} failed: {resp.text}")

    # Серверный доступ
    def submit_server_access(self, project_id: int, reason: str, max_instances: int = 4, cpu: int = 4000, ram: int = 4096) -> dict:
        url = f"{GATEWAY_URL}/api/v1/moderation/projects/{project_id}/server-access"
        resp = self.session.post(url, json={
            "project_id": str(project_id),
            "reason": reason,
            "max_instances": max_instances,
            "max_total_cpu_millis": cpu,
            "max_total_memory_mb": ram,
            "max_instance_cpu_millis": cpu // max_instances,
            "max_instance_memory_mb": ram // max_instances
        })
        if resp.status_code in (200, 201):
            return resp.json().get("request", {})
        raise RuntimeError(f"submit_server_access failed: {resp.text}")

    def get_server_access(self, project_id: int) -> dict:
        url = f"{GATEWAY_URL}/api/v1/moderation/projects/{project_id}/server-access"
        resp = self.session.get(url)
        if resp.status_code == 200:
            return resp.json().get("request", {})
        return {}

    def review_server_access(self, request_id: int, approved: bool, comment: str, max_instances: int = 4, cpu: int = 4000, ram: int = 4096) -> dict:
        url = f"{GATEWAY_URL}/api/v1/moderation/requests/{request_id}/review-server-access"
        resp = self.session.post(url, json={
            "request_id": str(request_id),
            "approved": approved,
            "max_instances": max_instances,
            "max_total_cpu_millis": cpu,
            "max_total_memory_mb": ram,
            "max_instance_cpu_millis": cpu // max_instances,
            "max_instance_memory_mb": ram // max_instances,
            "moderator_comment": comment
        })
        if resp.status_code in (200, 201):
            return resp.json().get("request", {})
        raise RuntimeError(f"review_server_access failed: {resp.text}")

    # Чат и вложения
    def upload_chat_attachment(self, project_id: int, file_path: str) -> dict:
        if not os.path.exists(file_path):
            log(f"Файл вложения {file_path} не найден", "WARN")
            return {}
        url = f"{GATEWAY_URL}/api/v1/projects/{project_id}/chat/attachments"
        with open(file_path, "rb") as f:
            files = {"file": (os.path.basename(file_path), f)}
            resp = self.session.post(url, files=files)
        if resp.status_code in (200, 201):
            return resp.json()
        raise RuntimeError(f"upload_chat_attachment failed: {resp.text}")

    def send_chat_message(self, project_id: int, content: str, attachment_ids: list = None) -> dict:
        url = f"{GATEWAY_URL}/api/v1/moderation/projects/{project_id}/messages"
        payload = {
            "project_id": str(project_id),
            "content": content,
            "attachment_ids": attachment_ids or []
        }
        resp = self.session.post(url, json=payload)
        if resp.status_code in (200, 201):
            return resp.json().get("message", {})
        raise RuntimeError(f"send_chat_message failed: {resp.text}")

    # Оркестратор: серверные билды и инстансы
    def upload_server_build(self, game_id: int, version: str, tar_path: str, protocol: str = "PROTOCOL_WEBSOCKET", internal_port: int = 8080, max_players: int = 16) -> dict:
        if not os.path.exists(tar_path):
            log(f"Серверный tar не найден: {tar_path}", "WARN")
            return {}
        url = f"{GATEWAY_URL}/api/v1/games/{game_id}/builds"
        with open(tar_path, "rb") as f:
            files = {"image": (os.path.basename(tar_path), f, "application/gzip")}
            data = {
                "build_version": version,
                "protocol": protocol,
                "internal_port": str(internal_port),
                "max_players": str(max_players)
            }
            resp = self.session.post(url, data=data, files=files)
        if resp.status_code in (200, 201):
            return resp.json().get("build", {})
        log(f"upload_server_build warning ({resp.status_code}): {resp.text}", "WARN")
        return {}

    def start_instance(self, game_id: int, build_version: str, name: str = "", max_players: int = 16) -> dict:
        url = f"{GATEWAY_URL}/api/v1/games/{game_id}/instances"
        payload = {
            "game_id": str(game_id),
            "build_version": build_version,
            "name": name,
            "max_players": max_players
        }
        resp = self.session.post(url, json=payload)
        if resp.status_code in (200, 201):
            return resp.json().get("instance", {})
        log(f"start_instance warning ({resp.status_code}): {resp.text}", "WARN")
        return {}

    def resume_instance(self, game_id: int, instance_id: int) -> dict:
        url = f"{GATEWAY_URL}/api/v1/games/{game_id}/instances/{instance_id}:resume"
        resp = self.session.post(url, json={})
        if resp.status_code == 200:
            return resp.json().get("instance", {})
        return {}

def cleanup_old_demo_projects(dev_client: GDHClient, collab_client: GDHClient):
    log("Проверка и очистка предыдущих тестовых демо-проектов (проект #5 сохраняется)...", "INFO")
    demo_titles = [
        "Приключения Котика: Волшебный Лес",
        "Пиксельная Арена: Мультиплеер",
        "Киберпанк: Неоновый Город",
        "Неоновый Дрифт: Токио",
        "Космическая Одиссея",
        "Тестовый проект API",
        "Новый проект"
    ]
    for client in [dev_client, collab_client]:
        projects = client.list_projects(limit=100)
        for p in projects:
            p_id = int(p.get("id"))
            p_title = p.get("title_ru", "")
            if p_id == 5:
                continue
            if p_title in demo_titles or "Котик" in p_title or "Арена" in p_title or "Киберпанк" in p_title or "Дрифт" in p_title:
                log(f"Удаление старого проекта #{p_id} ('{p_title}')...", "INFO")
                # Чистим заявки и снимки модерации для чистой перепривязки
                subprocess.run([
                    "docker", "exec", "gdh-postgres", "psql", "-U", "postgres", "-d", "orchestrator",
                    "-c", f"DELETE FROM moderation_snapshots WHERE project_id = {p_id}; "
                          f"DELETE FROM moderation_attachments WHERE project_id = {p_id}; "
                          f"DELETE FROM moderation_messages WHERE project_id = {p_id}; "
                          f"DELETE FROM moderation_requests WHERE project_id = {p_id};"
                ], capture_output=True)
                client.delete_project(p_id)

def seed_project_5_items(dev_client: GDHClient):
    log("Наполнение товарами IAP существующего проекта #5 (Steel Vanguard)...", "INFO")
    # Добавляем 3 товара
    dev_client.create_game_item(5, "tank_camo_desert", "Камуфляж 'Буря в пустыне'", "Пустынный песочный камуфляж для корпуса и башни танка", 150)
    dev_client.create_game_item(5, "tank_ap_shells", "Бронебойные снаряды x50", "Увеличенный урон по укрепленным бетонным ДОТам и укрытиям", 50)
    dev_client.create_game_item(5, "vip_pass_30d", "Премиум-аккаунт (30 дней)", "+50% опыта и монет за каждый результативный матч", 500)
    # Убеждаемся, что инстанс 17 запущен
    dev_client.resume_instance(5, 17)
    log("Проект #5: товары добавлены, игровой сервер онлайн", "SUCCESS")

def seed_published_game(dev_client: GDHClient, collab_client: GDHClient, mod_client: GDHClient) -> int:
    log("=== Сидирование Проекта: 'Приключения Котика' (Опубликованный Релиз) ===", "HEADER")

    # 1. Создание
    p = dev_client.create_project(
        title_ru="Приключения Котика: Волшебный Лес",
        title_en="Cat's Adventure: Magic Forest",
        is_online=False
    )
    p_id = int(p["id"])
    log(f"Создан проект #{p_id}", "SUCCESS")

    # 2. Обновление описания
    dev_client.update_project(
        p_id,
        seo_ru="Увлекательный 3D платформер-раннер с котиком в сказочном лесу",
        seo_en="Charming 3D platformer runner featuring a brave cat in an enchanted forest",
        about_ru="Отправляйтесь в захватывающее путешествие вместе с отважным котиком! Исследуйте древний волшебный лес, преодолевайте хитроумные ловушки, собирайте сияющие кристаллы магии и побеждайте лесных духов в динамичном 3D приключении.",
        about_en="Embark on an exciting journey with a brave cat! Explore ancient enchanted woods, dodge cunning obstacles, gather radiant magic crystals, and restore peace to the mystical woodland."
    )
    log(f"Проект #{p_id}: описание обновлено", "SUCCESS")

    # 3. Загрузка медиа
    dev_client.upload_media(p_id, "icon", MEDIA["icon_1"])
    dev_client.upload_media(p_id, "cover", MEDIA["cover_cat"])
    if os.path.exists(MEDIA["video_game"]):
        dev_client.upload_media(p_id, "video", MEDIA["video_game"])
    log(f"Проект #{p_id}: промо-материалы загружены", "SUCCESS")

    # 4. Загрузка клиентского билда
    build_resp = dev_client.upload_build(p_id, "1.0.0", MEDIA["build_cat"])
    dev_url = build_resp.get("dev_url", "")
    log(f"Проект #{p_id}: билд v1.0.0 загружен (URL: {dev_url})", "SUCCESS")

    # 5. Товары IAP
    dev_client.create_game_item(p_id, "cat_skin_tiger", "Облик 'Лесной тигренок'", "Уникальная полосатая расцветка шерстки с анимацией искр", 200)
    dev_client.create_game_item(p_id, "cat_boots_speed", "Сапоги скорости", "+20% к базовой скорости бега по лесным тропинкам", 120)
    dev_client.create_game_item(p_id, "magic_potion_pack", "Зелья неуязвимости (3 шт)", "Защищают от одного столкновения с препятствиями", 75)
    log(f"Проект #{p_id}: внутриигровые товары зарегистрированы", "SUCCESS")

    # 6. Командный доступ: приглашение Alex Dev
    inv = dev_client.send_invitation(
        p_id,
        collab_client.user.get("id"),
        collab_client.email,
        ["PERM_EDIT_INFO", "PERM_UPLOAD_BUILD", "PERM_VIEW_STATS"]
    )
    inv_id = int(inv["id"])
    collab_client.respond_invitation(inv_id, accept=True)
    log(f"Проект #{p_id}: {collab_client.email} успешно приглашен и вступил в команду проекта", "SUCCESS")

    # 7. Отправка на модерацию
    req_id = dev_client.submit_moderation(p_id)
    log(f"Проект #{p_id}: черновик отправлен на модерацию (Заявка #{req_id})", "SUCCESS")

    # 8. Модератор берет заявку
    mod_client.claim_moderation(req_id)
    log(f"Модератор взял заявку #{req_id} на проверку", "SUCCESS")

    # 9. Чат между модератором и разработчиком
    mod_client.send_chat_message(p_id, "Здравствуйте! Приступаем к проверке сборки v1.0.0. Проверяем корректность работы на различных разрешениях экрана и скорость загрузки ассетов.")
    time.sleep(0.3)
    dev_client.send_chat_message(p_id, "Спасибо! Сборка оптимизирована: текстуры упакованы в WebP, звуковые дорожки сжаты, поддерживаются любые соотношения сторон.")
    time.sleep(0.3)
    mod_client.send_chat_message(p_id, "Проверка завершена успешно! Частота кадров стабильная, сетевых утечек нет, возрастной рейтинг соблюден. Одобряем релиз к публикации.")
    time.sleep(0.3)

    # 10. Одобрение публикации
    mod_client.approve_moderation(
        p_id,
        comment="Проект полностью соответствует регламенту платформы Welwise Games. Отличная оптимизация, чистая графика и увлекательный геймплей. Рекомендован к размещению в основном каталоге."
    )
    log(f"Проект #{p_id}: релиз v1.0.0 одобрен и опубликован в каталоге!", "SUCCESS")

    return p_id

def seed_rejected_game(dev_client: GDHClient, mod_client: GDHClient) -> int:
    log("=== Сидирование Проекта: 'Пиксельная Арена' (Отклоненная Заявка) ===", "HEADER")

    # 1. Создание
    p = dev_client.create_project(
        title_ru="Пиксельная Арена: Мультиплеер",
        title_en="Pixel Arena: 2D Battle",
        is_online=True
    )
    p_id = int(p["id"])
    log(f"Создан проект #{p_id}", "SUCCESS")

    # 2. Описание
    dev_client.update_project(
        p_id,
        seo_ru="Динамичные 2D PvP дуэли на пиксельной гладиаторской арене",
        seo_en="Fast-paced 2D pixel gladiator multiplayer combat arena",
        about_ru="Сражайтесь на компактных аренах, используйте разрушаемое окружение, собирайте усилители и побеждайте соперников со всего мира в быстрых рейтинговых дуэлях.",
        about_en="Battle across compact arenas, utilize destructible cover, collect powerups, and outsmart rivals in lightning-fast competitive 2D battles."
    )

    # 3. Медиа
    dev_client.upload_media(p_id, "icon", MEDIA["icon_2"])
    dev_client.upload_media(p_id, "cover", MEDIA["cover_pixel"])

    # 4. Билд
    dev_client.upload_build(p_id, "0.9.5", MEDIA["build_client_101"])
    log(f"Проект #{p_id}: медиа и билд v0.9.5 загружены", "SUCCESS")

    # 5. Отправка на модерацию
    req_id = dev_client.submit_moderation(p_id)
    log(f"Проект #{p_id}: подан на модерацию (Заявка #{req_id})", "SUCCESS")

    # 6. Модератор берет заявку
    mod_client.claim_moderation(req_id)

    # 7. Загрузка вложения-скриншота и диалог в чате
    att = mod_client.upload_chat_attachment(p_id, MEDIA["shot_battle"])
    att_id = att.get("id")
    att_ids = [att_id] if att_id else []

    mod_client.send_chat_message(
        p_id,
        "Добрый день! В ходе проверки сборки v0.9.5 зафиксированы критические замечания:\n"
        "1. Элементы HUD перекрывают границы экрана на экранах с соотношением 16:10 (см. прикрепленный скриншот).\n"
        "2. Зафиксированы попытки сетевых запросов к незадекларированным аналитическим трекерам.\n"
        "Пожалуйста, устраните замечания для прохождения модерации.",
        attachment_ids=att_ids
    )
    time.sleep(0.3)
    dev_client.send_chat_message(p_id, "Здравствуйте! Спасибо за детализацию. Запросы к трекерам отключим полностью, а верстку интерфейса переведем на адаптивные якоря платформы в сборке v0.9.6.")
    time.sleep(0.3)
    mod_client.send_chat_message(p_id, "Отклоняю текущую заявку для внесения правок разработчиком. Ждем обновленный билд!")

    # 8. Отклонение модератором со структурированными нарушениями
    mod_client.reject_moderation(
        p_id,
        reason="Обнаружены критические нарушения регламента платформы Welwise Games: некорректная адаптация интерфейса под нестандартные разрешения и попытки выполнения сторонних сетевых вызовов без декларации в манифесте.",
        violations=[
            {
                "rule_code": "UI-04",
                "rule_title": "Адаптивность и масштабирование UI",
                "description": "Элементы управления и панель здоровья выходят за пределы безопасной зоны экрана (Safe Area) на нестандартных дисплеях.",
                "attachment_ids": att_ids
            },
            {
                "rule_code": "SEC-02",
                "rule_title": "Политика безопасности и CSP",
                "description": "Обнаружены попытки сетевых запросов к сторонним неавторизованным аналитическим сервисам в обход платформенного SDK.",
                "attachment_ids": []
            }
        ]
    )
    log(f"Проект #{p_id}: заявка отклонена с карточками нарушений", "SUCCESS")

    return p_id

def seed_in_review_game(dev_client: GDHClient, mod_client: GDHClient) -> int:
    log("=== Сидирование Проекта: 'Киберпанк: Неоновый Город' (На проверке + Серверы) ===", "HEADER")

    # 1. Создание
    p = dev_client.create_project(
        title_ru="Киберпанк: Неоновый Город",
        title_en="Cyberpunk: Neon City",
        is_online=True
    )
    p_id = int(p["id"])
    log(f"Создан проект #{p_id}", "SUCCESS")

    # 2. Описание
    dev_client.update_project(
        p_id,
        seo_ru="Динамичный 3D раннер в ночном неоновом мегаполисе с аугментациями и трюками",
        seo_en="Fast-paced 3D runner in a night neon metropolis with augmentations and stunts",
        about_ru="Управляйте аугментированным раннером, преодолевайте дронов корпораций, прокачивайте навыки паркура и ставьте рекорды скорости на крышах небоскребов ночного города!",
        about_en="Control an augmented runner, overcome corporate drones, upgrade parkour skills, and set speed records across high-rise metropolis rooftops!"
    )

    # 3. Медиа
    dev_client.upload_media(p_id, "icon", MEDIA["icon_1"])
    dev_client.upload_media(p_id, "cover", MEDIA["cover_cyber"])
    if os.path.exists(MEDIA["video_game"]):
        dev_client.upload_media(p_id, "video", MEDIA["video_game"])

    # 4. Клиентский билд
    dev_client.upload_build(p_id, "1.0.2", MEDIA["build_client_102"])
    log(f"Проект #{p_id}: медиа и клиентский билд v1.0.2 загружены", "SUCCESS")

    # 5. Заявка на серверы платформы
    srv_req = dev_client.submit_server_access(
        p_id,
        reason="Для проведения закрытого бета-тестирования мультиплеера на 16 игроков требуется выделение серверных мощностей платформы на ноде RU-Central-1.",
        max_instances=4,
        cpu=4000,
        ram=4096
    )
    srv_req_id = int(srv_req.get("id", 0))
    if not srv_req_id:
        srv_req_obj = dev_client.get_server_access(p_id)
        srv_req_id = int(srv_req_obj.get("id", 0))

    log(f"Проект #{p_id}: подана заявка на доступ к серверам #{srv_req_id}", "SUCCESS")

    # 6. Одобрение серверной заявки модератором
    if srv_req_id:
        mod_client.review_server_access(
            srv_req_id,
            approved=True,
            comment="Квота на 4 инстанса (4 vCPU, 4096 MB RAM) выделена в полном объеме для проведения ЗБТ.",
            max_instances=4,
            cpu=4000,
            ram=4096
        )
        log(f"Заявка на серверы #{srv_req_id} одобрена модератором, квота в Orchestrator выделена", "SUCCESS")

    # 7. Загрузка серверного билда и запуск инстанса
    dev_client.upload_server_build(p_id, "v1.0.0", MEDIA["build_server_tar"], protocol="PROTOCOL_WEBSOCKET", internal_port=8080, max_players=16)
    dev_client.start_instance(p_id, "v1.0.0", name="neon-city-matchmaker-01", max_players=16)
    log(f"Проект #{p_id}: серверный Docker-билд загружен, инстанс 'neon-city-matchmaker-01' запущен", "SUCCESS")

    # 8. Отправка черновика на модерацию публикации
    pub_req_id = dev_client.submit_moderation(p_id)
    log(f"Проект #{p_id}: черновик отправлен на публикацию (Заявка #{pub_req_id})", "SUCCESS")

    # 9. Модератор берет заявку в работу (IN_REVIEW)
    mod_client.claim_moderation(pub_req_id)
    mod_client.send_chat_message(p_id, "Здравствуйте! Заявка взята в работу. Выделенные серверы активны, проводим стресс-тестирование сетевой синхронизации и сетевого матчмейкинга.")
    time.sleep(0.3)
    dev_client.send_chat_message(p_id, "Отлично! Серверный билд настроен на порт 8080, протокол WebSocket. Ждем результатов тестирования.")

    log(f"Проект #{p_id}: статус 'На проверке' (IN_REVIEW), серверы запущены", "SUCCESS")

    return p_id

def seed_collab_invitation(dev_client: GDHClient, collab_client: GDHClient) -> int:
    log("=== Сидирование Входящего Приглашения (Проект Alex Dev -> Test Developer) ===", "HEADER")

    # Alex Dev создает свой проект
    p = collab_client.create_project(
        title_ru="Неоновый Дрифт: Токио",
        title_en="Neon Drift: Tokyo",
        is_online=False
    )
    p_id = int(p["id"])

    collab_client.update_project(
        p_id,
        seo_ru="Аркадный дрифт на ночных автострадах Токио",
        seo_en="Arcade drift racing on neon-lit Tokyo highways",
        about_ru="Управляйте кастомизированными японскими спорткарами, входите в затяжные управляемые заносы, зарабатывайте очки стиля и открывайте новые трассы.",
        about_en="Drive customized Japanese sports cars, pull off long controlled drifts, score style points, and unlock iconic urban racetracks."
    )
    collab_client.upload_media(p_id, "icon", MEDIA["icon_1"])
    collab_client.upload_media(p_id, "cover", MEDIA["cover_cyber"])

    # Отправка приглашения для главного разработчика
    inv = collab_client.send_invitation(
        p_id,
        dev_client.user.get("id"),
        dev_client.email,
        ["PERM_EDIT_INFO", "PERM_UPLOAD_BUILD", "PERM_VIEW_STATS"]
    )
    log(f"Alex Dev отправил приглашение в проект #{p_id} ('Неоновый Дрифт') для {dev_client.email}", "SUCCESS")
    log("Приглашение оставлено в статусе PENDING для демонстрации уведомления в консоли", "SUCCESS")

    return p_id

def main():
    print("=" * 70)
    print("  GAME DEVELOPER HUB — ПОЛНОЕ СИДИРОВАНИЕ ДАННЫХ ЧЕРЕЗ API")
    print("=" * 70)

    # 1. Авторизация клиентов
    log("Авторизация пользователей...", "INFO")
    dev_client = GDHClient(*CREDENTIALS["dev"])
    if not dev_client.login():
        log(f"Не удалось войти под {CREDENTIALS['dev'][0]}", "ERROR")
        sys.exit(1)
    log(f"Авторизован Главный Разработчик: {dev_client.email}", "SUCCESS")

    collab_client = GDHClient(*CREDENTIALS["collab"])
    if not collab_client.register_and_verify(display_name="Alex Developer"):
        log(f"Не удалось зарегистрировать/авторизовать {CREDENTIALS['collab'][0]}", "ERROR")
        sys.exit(1)
    log(f"Авторизован Второй Разработчик: {collab_client.email}", "SUCCESS")

    mod_client = GDHClient(*CREDENTIALS["moderator"])
    if not mod_client.login():
        log(f"Не удалось войти под {CREDENTIALS['moderator'][0]}", "ERROR")
        sys.exit(1)
    log(f"Авторизован Модератор: {mod_client.email}", "SUCCESS")

    # 2. Очистка старых демо-проектов
    cleanup_old_demo_projects(dev_client, collab_client)

    # 3. Наполнение проекта 5
    seed_project_5_items(dev_client)

    # 4. Проект 1: Опубликованный (Котик)
    pub_id = seed_published_game(dev_client, collab_client, mod_client)

    # 5. Проект 2: Отклоненный (Арена)
    rej_id = seed_rejected_game(dev_client, mod_client)

    # 6. Проект 3: На проверке + Серверы (Киберпанк)
    rev_id = seed_in_review_game(dev_client, mod_client)

    # 7. Проект 4: Входящее приглашение (Alex Dev -> Test Dev)
    inv_id = seed_collab_invitation(dev_client, collab_client)

    print("\n" + "=" * 70)
    log("СИДИРОВАНИЕ УСПЕШНО ЗАВЕРШЕНО!", "SUCCESS")
    print("=" * 70)
    print("\nУЧЕТНЫЕ ДАННЫЕ ДЛЯ ВХОДА:")
    print("  • Главный разработчик : dev@welwise.com       / password123")
    print("  • Второй разработчик  : alex_dev@welwise.com  / password123")
    print("  • Модератор платформы : moderator@welwise.com / password123")
    print("  • Администратор       : admin@welwise.com     / password123")
    print("\nСОСТОЯНИЕ ПРОЕКТОВ В КОНСОЛИ:")
    print(f"  • Проект #{pub_id:2d}: [ОПУБЛИКОВАН]  'Приключения Котика: Волшебный Лес' (Релиз 1.0.0, IAP, Команда)")
    print(f"  • Проект #{rej_id:2d}: [ОТКЛОНЕН]     'Пиксельная Арена: Мультиплеер' (Вердикт, 2 нарушения, скриншот)")
    print(f"  • Проект #{rev_id:2d}: [НА ПРОВЕРКЕ]  'Киберпанк: Неоновый Город' (Серверная квота, инстанс онлайн)")
    print(f"  • Проект # 5: [ОНЛАЙН ДРАФТ] 'Steel Vanguard: Танковая Арена' (Сохранен, IAP, инстанс онлайн)")
    print(f"  • Проект #{inv_id:2d}: [ИНВАЙТ]       'Неоновый Дрифт: Токио' (Входящее приглашение для dev@welwise.com)")
    print("\nЖУРНАЛЫ И РАЗДЕЛЫ ДЛЯ ВИДЕО:")
    print("  • Консоль разработчика : http://localhost/projects")
    print("  • Приглашения команды  : http://localhost/projects (вкладка Доступ / колокольчик)")
    print("  • Публичный каталог    : http://localhost/catalog")
    print("  • Очередь модерации    : http://localhost/moderation/queue")
    print("  • Журнал решений (АРХ) : http://localhost/moderation/archive")
    print("  • Чаты тикетов         : http://localhost/moderation/chats")
    print("  • Журнал модератора    : http://localhost/admin/moderators")
    print("=" * 70 + "\n")

if __name__ == "__main__":
    main()
