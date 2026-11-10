"""BPMN «как будет» для статьи: самообслуживание разработчика и обслуживание серверов системой (укрупнённая схема)."""
import sys
from compact import Diagram

d = Diagram('Управление игровыми серверами — как будет', [
    ('dev', 'Разработчик игры', 2),
    ('mod', 'Модератор', 1),
    ('sys', 'Разрабатываемая система', 3.3),
])

d.node('s', 'start', 'dev', 0, label='Нужен игровой сервер')
d.node('t1', 'task_user', 'dev', 1, label='Загрузить серверную сборку')
d.node('g1', 'xgw', 'dev', 2, label='Свой узел?')
d.node('t2', 'task_user', 'dev', 3, label='Подключить собственный узел')
d.node('t3', 'task_user', 'dev', 3, row=1, label='Подать заявку на серверы платформы')
d.node('t4', 'task_user', 'mod', 4, label='Одобрить доступ')
d.node('g2', 'xgw', 'dev', 5)
d.node('t5', 'task_user', 'dev', 6, label='Задать политику оркестрации')

d.edge('s', 't1')
d.edge('t1', 'g1')
d.edge('g1', 't2', 'Да', exit='right', entry='left')
d.edge('g1', 't3', 'Нет', exit='bottom', entry='left', via=[(d.cx(2), d.cy('dev', 1))])
d.edge('t3', 't4', exit='right', entry='top', via=[(d.cx(4), d.cy('dev', 1))])
d.edge('t2', 'g2')
d.edge('t4', 'g2', exit='right', entry='bottom', via=[(d.cx(5), d.cy('mod', 0))])
d.edge('g2', 't5')

# Система: запуск по политике, подключение игроков, самовосстановление
d.node('si', 'start', 'sys', 0, label='Политика задана')
d.node('i1', 'task_service', 'sys', 1, label='Запустить экземпляры по политике')
d.node('im', 'msg', 'sys', 2, label='Запрос игрока')
d.node('ig', 'xgw', 'sys', 3, label='Есть свободное<br>место?')
d.node('i2', 'task_service', 'sys', 4, row=1, label='Запустить экземпляр или поставить в очередь')
d.node('i3', 'task_service', 'sys', 4, label='Выдать адрес и порт экземпляра')
d.node('ie', 'end', 'sys', 5, label='Игрок подключён')

d.edge('t5', 'si', exit='bottom', entry='top',
       via=[(d.cx(6), d.lane_y('sys', 8)), (d.cx(0), d.lane_y('sys', 8))])
d.edge('si', 'i1')
d.edge('i1', 'im')
d.edge('im', 'ig')
d.edge('ig', 'i3', 'Да', exit='right', entry='left')
d.edge('ig', 'i2', 'Нет', exit='bottom', entry='left', via=[(d.cx(3), d.cy('sys', 1))])
d.edge('i2', 'i3', exit='right', entry='bottom', via=[(d.cx(5) - 10, d.cy('sys', 1)), (d.cx(5) - 10, d.cy('sys', 0) + 40), (d.cx(4), d.cy('sys', 0) + 40)])
d.edge('i3', 'ie')

# Самовосстановление и экономия ресурсов — отдельные обработчики событий
d.node('r1', 'sub_cond', 'sys', 0, row=2.15, label='Сбой экземпляра или узла', lpos='right')
d.node('r2', 'task_service', 'sys', 2, row=2.15, label='Перезапустить экземпляр')
d.node('r3', 'sub_timer', 'sys', 4, row=2.15, label='Нет игроков дольше таймаута', lpos='right')
d.node('r4', 'task_service', 'sys', 6, row=2.15, label='Остановить экземпляр')
d.subprocess('sp1', 'sys', 0, 2, 2.15, 2.15, 'Восстановление после сбоя')
d.subprocess('sp2', 'sys', 3, 6, 2.15, 2.15, 'Экономия ресурсов')
d.edge('r1', 'r2')
d.edge('r3', 'r4')

d.note('n1', 't2', 'dev', 4.6, 1, 'Новая возможность: собственные узлы', w=120)
d.note('n2', 'i1', 'sys', 1.1, 1, 'Без участия инженера', w=100)

open(sys.argv[1], 'w').write(d.render())
