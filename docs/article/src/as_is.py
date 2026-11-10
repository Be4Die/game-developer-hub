"""BPMN «как есть» для статьи: управление игровыми серверами через инженера (укрупнённая схема)."""
import sys
from compact import Diagram

d = Diagram('Управление игровыми серверами — как есть', [
    ('dev', 'Разработчик игры', 2),
    ('eng', 'Инженер по эксплуатации', 1.45),
])

d.node('s', 'start', 'dev', 0, label='Нужен игровой сервер')
d.node('t1', 'task_user', 'dev', 1, label='Запросить размещение в Telegram')
d.node('t2', 'task_user', 'eng', 2, label='Создать репозиторий и конвейер')
d.node('t3', 'task_user', 'dev', 3, label='Загрузить сборку в репозиторий')
d.node('t4', 'task_user', 'dev', 4, label='Запросить действие в Telegram')
d.node('t5', 'task_user', 'eng', 5, label='Вручную выполнить действие на узле')
d.node('t6', 'task_user', 'dev', 6, label='Проверить работу без доступа к журналам')
d.node('g1', 'xgw', 'dev', 7, label='')
d.node('e1', 'end', 'dev', 7, row=1, label='Сервер работает')

d.edge('s', 't1')
d.edge('t1', 't2', exit='right', entry='top', via=[(d.cx(2), d.cy('dev', 0))])
d.edge('t2', 't3', exit='right', entry='bottom', via=[(d.cx(3), d.cy('eng', 0))])
d.edge('t3', 't4')
d.edge('t4', 't5', exit='right', entry='top', via=[(d.cx(5), d.cy('dev', 0))])
d.edge('t5', 't6', exit='right', entry='bottom', via=[(d.cx(6), d.cy('eng', 0))])
d.edge('t6', 'g1')
d.edge('g1', 'e1', 'Действий больше нет', exit='bottom', entry='top')
d.edge('g1', 't4', 'Нужна остановка, перезапуск или новая версия', exit='top', entry='top',
       via=[(d.cx(7), d.lane_y('dev', 6)), (d.cx(4), d.lane_y('dev', 6))], lx=-0.1)

d.note('n1', 't2', 'eng', 3.6, 0, 'Отдельная ручная настройка для каждой игры', w=130)
d.note('n2', 't5', 'eng', 7.0, 0, 'Каждое действие — через инженера', w=100)

open(sys.argv[1], 'w').write(d.render())
