"""UML-диаграмма развёртывания в формате draw.io.

Координаты заданы явно (пиксели draw.io), связи — ломаными с заданными точками,
чтобы раскладка не «плыла» при правках. Запуск: python3 deployment.py <выход.drawio>
"""
import sys
from xml.sax.saxutils import escape

FONT = 'fontFamily=Times New Roman;fontSize=12;'
CUBE = 'shape=cube;size=10;direction=south;html=1;whiteSpace=wrap;verticalAlign=top;align=center;boundedLbl=1;fillColor=#FFFFFF;' + FONT
COMP = 'shape=module;jettyWidth=10;jettyHeight=6;html=1;whiteSpace=wrap;align=center;verticalAlign=middle;spacingLeft=12;fillColor=#FFFFFF;' + FONT
ART = 'shape=note;size=12;html=1;whiteSpace=wrap;align=center;verticalAlign=middle;fillColor=#FFFFFF;' + FONT
GROUP = 'rounded=0;dashed=1;fillColor=none;html=1;whiteSpace=wrap;verticalAlign=top;align=left;spacingLeft=8;fontStyle=2;' + FONT
EDGE = ('edgeStyle=none;rounded=0;html=1;endArrow=none;jumpStyle=arc;jumpSize=8;'
        'labelBackgroundColor=#FFFFFF;fontFamily=Times New Roman;fontSize=11;')

cells, eid = [], [0]


def v(cid, value, style, x, y, w, h):
    cells.append(f'<mxCell id="{cid}" value="{escape(value, {chr(34): "&quot;"})}" style="{style}" vertex="1" parent="1">'
                 f'<mxGeometry x="{x}" y="{y}" width="{w}" height="{h}" as="geometry"/></mxCell>')


def node(cid, stereo, name, x, y, w, h, bold=True):
    label = f'«{stereo}»<br>' + (f'<b>{name}</b>' if bold else name)
    v(cid, label, CUBE, x, y, w, h)


def comp(cid, name, x, y, w, h=38):
    v(cid, name, COMP, x, y, w, h)


def art(cid, name, x, y, w, h=50):
    v(cid, f'«artifact»<br>{name}', ART, x, y, w, h)


def link(points, label='', arrow=False, lpos=None):
    """Связь по точкам [(x, y), ...]; подпись — у середины или в lpos=(x, y)."""
    eid[0] += 1
    (sx, sy), (tx, ty) = points[0], points[-1]
    mid = ''.join(f'<mxPoint x="{x}" y="{y}"/>' for x, y in points[1:-1])
    style = EDGE + ('endArrow=open;endFill=0;dashed=1;' if arrow else '')
    cells.append(f'<mxCell id="l{eid[0]}" value="" style="{style}" edge="1" parent="1"><mxGeometry relative="1" as="geometry">'
                 f'<mxPoint x="{sx}" y="{sy}" as="sourcePoint"/><mxPoint x="{tx}" y="{ty}" as="targetPoint"/>'
                 f'<Array as="points">{mid}</Array></mxGeometry></mxCell>')
    if label:
        if lpos is None:
            a, b = points[len(points) // 2 - 1], points[len(points) // 2]
            lpos = ((a[0] + b[0]) / 2, (a[1] + b[1]) / 2)
        w = 7 * max(len(s) for s in label.split('<br>')) + 10
        h = 16 * (label.count('<br>') + 1)
        v(f'll{eid[0]}', label, 'text;html=1;align=center;verticalAlign=middle;labelBackgroundColor=#FFFFFF;'
          'fontFamily=Times New Roman;fontSize=11;', lpos[0] - w / 2, lpos[1] - h / 2, w, h)


# ── Устройства пользователей и смежные системы ─────────────
node('dev', 'device', 'Устройство разработчика<br>или сотрудника', 20, 60, 240, 180)
node('dev_br', 'execution environment', 'Браузер', 35, 120, 210, 108, bold=False)
art('dev_spa', 'SPA-приложение', 50, 168, 180, 48)

node('pl', 'device', 'Устройство игрока', 20, 280, 240, 250)
node('pl_br', 'execution environment', 'Браузер', 35, 325, 210, 190, bold=False)
node('pl_game', 'execution environment', 'Веб-игра', 50, 370, 180, 130, bold=False)
comp('pl_sdk', 'Модуль обнаружения сервера (SDK)', 60, 425, 160, 60)

node('mail', 'device', 'Почтовый сервер<br><i>(смежная система)</i>', 20, 590, 240, 90)
node('an', 'device', 'Сервисы аналитики и покупок площадки<br><i>(смежные системы)</i>', 20, 730, 240, 90)

# ── Сервер платформы ───────────────────────────────────────
node('srv', 'device', 'Сервер платформы Welwise', 330, 40, 1080, 1010)
node('dock', 'execution environment', 'Docker Compose', 345, 85, 1050, 950, bold=False)

node('ngx', 'execution environment', 'Nginx :80', 370, 150, 320, 205)
comp('ngx1', 'Статический веб-сервер SPA', 390, 200, 280, 38)
comp('ngx2', 'Обратный прокси /api, /game-proxy', 390, 250, 280, 38)
comp('ngx3', 'Раздача игр и медиа /games, /media', 390, 300, 280, 38)

node('gw', 'execution environment', 'API-шлюз :8080', 730, 150, 300, 150)
comp('gw1', 'Преобразование REST в gRPC', 750, 200, 260, 38)
comp('gw2', 'Загрузка файлов и вложений', 750, 248, 260, 38)

node('ing', 'execution environment', 'Входной прокси :8085', 1070, 150, 300, 110)
comp('ing1', 'Маршрутизация игрового трафика', 1090, 200, 260, 45)


services = [
    ('sso', 'Сервис учётных записей :50051', ['Регистрация и вход', 'Выпуск токенов', 'Подтверждение почты']),
    ('pm', 'Сервис управления проектами :50053', ['Черновики и релизы', 'Сборки и медиа', 'Командный доступ']),
    ('mod', 'Сервис модерации :50054', ['Очередь заявок', 'Аудит-снимки', 'Чат проекта']),
    ('orch', 'Сервис оркестрации :50052', ['Управление узлами', 'Управление экземплярами', 'Обнаружение и очередь']),
]
SX = [370, 620, 870, 1120]
for (sid, name, parts), x in zip(services, SX):
    node(sid, 'execution environment', name, x, 560, 220, 195)
    for i, p in enumerate(parts):
        comp(f'{sid}{i}', p, x + 15, 615 + i * 44, 190, 36)

v('data', 'Хранилища данных', GROUP.replace('align=left;spacingLeft=8;', 'align=right;spacingRight=8;'), 370, 815, 970, 205)
node('s3', 'execution environment', 'SeaweedFS (S3) :8333', 390, 850, 290, 150)
art('s3a', 'Сборки, каталог игр, медиа, вложения', 410, 920, 250, 55)
node('pg', 'execution environment', 'PostgreSQL 17 :5432', 705, 850, 290, 150)
art('pga', 'Базы данных сервисов', 725, 920, 250, 55)
node('vk', 'execution environment', 'Valkey 8 :6379', 1020, 850, 290, 150)
art('vka', 'Блокировки, кэш, маршруты', 1040, 920, 250, 55)

# ── Вычислительный узел ────────────────────────────────────
node('nd', 'device', 'Вычислительный узел', 1520, 330, 300, 580)
node('ndd', 'execution environment', 'Docker', 1535, 375, 270, 520, bold=False)
node('inst', 'execution environment', 'Экземпляр игрового сервера', 1550, 425, 240, 65)
node('na', 'execution environment', 'Агент узла :44044', 1550, 520, 240, 120)
comp('na1', 'Объявление узла, отчёты о состоянии', 1562, 565, 216, 30)
comp('na2', 'Управление контейнерами и журналы', 1562, 602, 216, 30)
node('dsv', 'execution environment', 'Сервисы данных игры', 1550, 665, 240, 105)
comp('dsv1', 'PostgreSQL, MySQL, Redis, том', 1562, 715, 216, 38)
art('bk', 'Резервные копии', 1550, 795, 240, 55)

# ── Связи ──────────────────────────────────────────────────
link([(245, 180), (370, 180)], '«protocol»<br>HTTPS через<br>Nginx площадки', lpos=(302, 148))
link([(245, 420), (345, 420), (345, 300), (370, 300)], '«protocol»<br>HTTPS, WSS', lpos=(295, 395))
# Прямой режим: игрок подключается к узлу по его домену
link([(20, 470), (10, 470), (10, 20), (1500, 20), (1500, 445), (1550, 445)], '«protocol» WSS по домену узла (прямой режим)', lpos=(900, 20))

link([(690, 225), (730, 225)], 'HTTP', lpos=(710, 212))
link([(600, 150), (600, 135), (1220, 135), (1220, 150)], '«protocol» WebSocket (/game-proxy)', lpos=(910, 135))
link([(370, 330), (357, 330), (357, 832), (430, 832), (430, 850)], 'HTTP (S3)', lpos=(395, 832))

# Шлюз — сервисы
for x0, x1, yb in [(760, 480, 525), (800, 730, 535), (900, 980, 545), (1000, 1230, 530)]:
    link([(x0, 300), (x0, yb), (x1, yb), (x1, 560)])
v('gwlab', '«protocol» gRPC', 'text;html=1;align=center;verticalAlign=middle;labelBackgroundColor=#FFFFFF;fontFamily=Times New Roman;fontSize=11;', 1065, 507, 100, 16)
link([(855, 300), (855, 800), (535, 800), (535, 850)], '«protocol» S3', lpos=(640, 800))

# Межсервисные вызовы
link([(840, 645), (870, 645)], 'gRPC', lpos=(855, 628))
link([(1090, 645), (1120, 645)], 'gRPC', lpos=(1105, 628))

link([(1180, 560), (1180, 440), (680, 440), (680, 560)], '«protocol» gRPC: проверка прав на проект', lpos=(1040, 440))
# Сервисы — хранилища
for x, lab in [(480, 'SQL, RESP'), (730, 'SQL, RESP, S3'), (980, 'SQL'), (1230, 'SQL, RESP, S3')]:
    link([(x, 755), (x, 815)], lab, lpos=(x, 785))
link([(1355, 260), (1355, 925), (1310, 925)], '«protocol»<br>RESP<br>(маршруты)', lpos=(1355, 420))

# Смежные системы
link([(370, 635), (260, 635)], '«protocol»<br>SMTP', lpos=(300, 612))
link([(650, 755), (650, 768), (260, 768)], '«protocol» HTTPS', lpos=(300, 752))

# Узел: входной прокси -> экземпляр; оркестратор <-> агент узла; агент -> контейнеры
link([(1370, 205), (1470, 205), (1470, 465), (1550, 465)], 'WebSocket<br>(прокси<br>платформы)', lpos=(1440, 300))
link([(1340, 600), (1440, 600), (1440, 555), (1550, 555)], 'gRPC :44044<br>команды, опрос,<br>журналы', lpos=(1490, 520))
link([(1550, 615), (1425, 615), (1425, 700), (1340, 700)], 'gRPC :50052<br>объявление<br>узла', lpos=(1470, 690))
link([(1670, 520), (1670, 490)], 'Docker API; HTTP :44045', lpos=(1735, 505))
link([(1670, 640), (1670, 665)], 'Docker API', lpos=(1720, 652))

body = '\n'.join(cells)
open(sys.argv[1], 'w').write('<mxfile><diagram name="Диаграмма развёртывания"><mxGraphModel grid="0" page="0"><root>\n'
                             '<mxCell id="0"/>\n<mxCell id="1" parent="0"/>\n' + body + '\n</root></mxGraphModel></diagram></mxfile>')
