"""BPMN «как есть»: управление игровыми серверами до внедрения Системы."""
import sys
from bpmn import Diagram

d = Diagram('Управление игровыми серверами — «как есть»', [
    ('dev', 'Разработчик игры', 2),
    ('eng', 'Инженер по эксплуатации', 2),
    ('ci', 'GitHub Actions', 1),
])

d.node('s', 'start', 'dev', 0, label='Серверная сборка игры готова')
d.node('t1', 'task_user', 'dev', 1, label='Запросить размещение сервера в Telegram')
d.node('t2', 'task_user', 'eng', 2, label='Создать репозиторий и конвейер для сервера')
d.node('t3', 'task_user', 'dev', 3, label='Загрузить серверную сборку в репозиторий')
d.node('t4', 'task_service', 'ci', 4, label='Доставить сборку на выделенный сервер')
d.node('t5', 'task_user', 'dev', 5, label='Запросить запуск сервера в Telegram')
d.node('t6', 'task_user', 'eng', 6, label='Вручную запустить экземпляр')
d.node('t7', 'task_user', 'dev', 7, label='Проверить работу сервера без доступа к журналам')
d.node('g1', 'xgw', 'dev', 8, label='Требуется<br>действие?')
d.node('t8', 'task_user', 'dev', 9, row=1, label='Запросить остановку или перезапуск в Telegram')
d.node('t9', 'task_user', 'eng', 10, label='Вручную остановить или перезапустить экземпляр')
d.node('e1', 'end', 'dev', 10, label='Сервер работает')

top = d.lane_y('dev', 10)
d.edge('s', 't1')
d.edge('t1', 't2')
d.edge('t2', 't3')
d.edge('t3', 't4')
d.edge('t4', 't5')
d.edge('t5', 't6')
d.edge('t6', 't7', exit='right', entry='bottom', via=[(d.cx(7), d.cy('eng', 0))])
d.edge('t7', 'g1')
d.edge('g1', 'e1', 'Нет', exit='right', entry='left')
d.edge('g1', 't8', 'Остановка или перезапуск', exit='bottom', entry='left', via=[(d.cx(8), d.cy('dev', 1))])
d.edge('t8', 't9', exit='right', entry='top', via=[(d.cx(10), d.cy('dev', 1))])
d.edge('t9', 't7', exit='left', entry=(0.8, 1), via=[(d.cx(7) + 25, d.cy('eng', 0))])
d.edge('g1', 't3', 'Новая версия', exit='top', entry='top', via=[(d.cx(8), top), (d.cx(3), top)], lx=-0.85)

d.note('n1', 't2', 'eng', 2, 1, 'Отдельная ручная настройка для каждой игры', w=150)
d.note('n3', 't9', 'eng', 10, 1, 'Каждое действие — через инженера с ожиданием ответа', w=160)

open(sys.argv[1], 'w').write(d.render())
