"""Замеры для статьи на локальном стенде (http://localhost, демонстрационная игра GAME с серверной сборкой).

Серии:
  cold — холодный запуск: от единственного запроса обнаружения сервера при отсутствии экземпляров до состояния
         «Работает» и ответа «Готов»;
  idle — остановка при простое: от готовности экземпляра без игроков до его остановки (тайм-аут простоя 1 мин).
Запрос обнаружения отправляется один раз, дальше опрашивается список экземпляров: повторные запросы обнаружения
во время запуска порождают лишние экземпляры (гонка в сервисе оркестрации, см. README.md).
Скрипт удаляет созданные им экземпляры и возвращает исходную политику игры.

Запуск: python3 measure.py <результат.json> [cold=10] [idle=3]
"""
import json, sys, time, urllib.request

BASE, GAME = 'http://localhost/api/v1', 5
OUT = sys.argv[1]
N = dict(cold=10, idle=3)
for a in sys.argv[2:]:
    k, v = a.split('='); N[k] = int(v)


def call(method, path, body=None, token=None):
    req = urllib.request.Request(BASE + path, method=method, data=json.dumps(body).encode() if body is not None else None)
    req.add_header('Content-Type', 'application/json')
    if token:
        req.add_header('Authorization', 'Bearer ' + token)
    with urllib.request.urlopen(req, timeout=30) as r:
        return json.loads(r.read() or b'{}')


TOKEN = call('POST', '/auth/login', {'email': 'dev@welwise.com', 'password': 'password123'})['tokens']['access_token']
api = lambda m, p, b=None: call(m, p, b, TOKEN)
instances = lambda: api('GET', f'/games/{GAME}/instances').get('instances', [])
active = lambda: [i for i in instances() if i['status'] in ('INSTANCE_STATUS_RUNNING', 'INSTANCE_STATUS_STARTING')]
log = lambda *a: print(time.strftime('%H:%M:%S'), *a, flush=True)


def wait(cond, timeout, step=0.2):
    t0 = time.monotonic()
    while time.monotonic() - t0 < timeout:
        v = cond()
        if v:
            return v
        time.sleep(step)
    return None


def cleanup():
    for i in instances():
        if int(i['id']) not in BASELINE:
            try:
                api('DELETE', f"/games/{GAME}/instances/{i['id']}")
            except Exception as e:
                log('delete failed', i['id'], e)
    wait(lambda: not active(), 60)


def set_policy(**kw):
    p = dict(POLICY); p.update(kw)
    api('POST', f'/games/{GAME}/policy', p)


def start_ready():
    """Один запрос обнаружения при пустом наборе экземпляров; возвращает время до готовности экземпляра."""
    t0 = time.monotonic()
    call('GET', f'/games/{GAME}/discover')
    running = lambda: [i for i in instances() if i['status'] == 'INSTANCE_STATUS_RUNNING']
    if not wait(running, 120, 0.1):
        raise RuntimeError('instance did not start')
    dt = time.monotonic() - t0
    if call('GET', f'/games/{GAME}/discover').get('status') != 'DISCOVERY_STATUS_READY':
        raise RuntimeError('discovery is not ready')
    return dt


POLICY = api('GET', f'/games/{GAME}/policy')['policy']
BASELINE = {15, 16, 17, 20, 21}  # экземпляры, существовавшие на стенде до замеров
res = dict(started=time.strftime('%Y-%m-%d %H:%M:%S'), cold=[], idle=[])
try:
    cleanup()
    # ── Холодный запуск ──
    set_policy(scale_to_zero_timeout=30)
    for k in range(N['cold']):
        dt = start_ready()
        res['cold'].append(round(dt, 2)); log('cold', k + 1, f'{dt:.2f}', 'экземпляров:', len(active()))
        cleanup(); time.sleep(1)
    # ── Остановка при простое ──
    set_policy(scale_to_zero_timeout=1)
    for k in range(N['idle']):
        start_ready()
        t0 = time.monotonic()
        ok = wait(lambda: not active(), 240, 0.5)
        res['idle'].append(round(time.monotonic() - t0, 1) if ok else None); log('idle', k + 1, res['idle'][-1], 'экземпляров:', len(instances()) - len(BASELINE))
        cleanup(); time.sleep(1)
finally:
    try:
        cleanup()
        api('POST', f'/games/{GAME}/policy', POLICY)
    finally:
        json.dump(res, open(OUT, 'w'), ensure_ascii=False, indent=1)
        log('done', json.dumps(res, ensure_ascii=False))
