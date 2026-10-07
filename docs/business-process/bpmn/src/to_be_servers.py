"""BPMN «как должно быть»: управление игровыми серверами в разрабатываемой системе.

Разработчик готовит инфраструктуру, дальше система сама обслуживает серверы в развёрнутом подпроцессе:
запускает экземпляры по политике и распределяет игроков. События эксплуатации — граничные события
на нижней кромке подпроцесса, каждое со своим обработчиком прямо под ним.
"""
import sys
from bpmn import Diagram

d = Diagram('Управление игровыми серверами — «как должно быть»', [
    ('dev', 'Разработчик игры', 2),
    ('mod', 'Модератор', 1),
    ('sys', 'Разрабатываемая система', 5.0),
])

# Подготовка инфраструктуры
d.node('s', 'start', 'dev', 0, label='Нужен игровой сервер')
d.node('t1', 'task_user', 'dev', 1, label='Загрузить серверную сборку')
d.node('g1', 'xgw', 'dev', 2, label='Собственный<br>сервер?')
d.node('t2', 'task_user', 'dev', 3, label='Установить агент, подтвердить узел')
d.node('t3', 'task_user', 'dev', 3, row=1, label='Подать заявку на доступ к серверам платформы')
d.node('t4', 'task_user', 'mod', 4, label='Одобрить доступ к серверам платформы')
d.node('g1m', 'xgw', 'dev', 5)
d.node('t5', 'task_user', 'dev', 6, label='Настроить политику оркестрации')

d.edge('s', 't1')
d.edge('t1', 'g1')
d.edge('g1', 't2', 'Да', exit='right', entry='left')
d.edge('g1', 't3', 'Нет', exit='bottom', entry='left', via=[(d.cx(2), d.cy('dev', 1))])
d.edge('t3', 't4', exit='right', entry='top', via=[(d.cx(4), d.cy('dev', 1))])
d.edge('t2', 'g1m')
d.edge('t4', 'g1m', exit='right', entry='bottom', via=[(d.cx(5), d.cy('mod', 0))])
d.edge('g1m', 't5')

# Развёрнутый подпроцесс: система обслуживает серверы
d.subprocess('sp', 'sys', 0, 8, 0, 2, 'Обслуживание игровых серверов', expanded=True)
d.node('si', 'start', 'sys', 0)
d.node('i1', 'task_service', 'sys', 1, label='Запустить экземпляры по политике')
d.node('im', 'msg', 'sys', 2, label='Запрос игрока на подключение')
d.node('ig', 'xgw', 'sys', 3, label='Есть свободное<br>место?')
d.node('ig2', 'xgw', 'sys', 3, row=1, label='Поведение<br>при заполнении', lpos='left')
d.node('i4', 'task_service', 'sys', 4, row=1, label='Запустить новый экземпляр')
d.node('i5', 'task_service', 'sys', 4, row=2, label='Поставить игрока в очередь')
d.node('ic', 'cond', 'sys', 5, row=2, label='Место освободилось')
d.node('i3', 'task_service', 'sys', 5, label='Выдать адрес и порт экземпляра')
d.node('ie', 'end', 'sys', 6, label='Игрок подключён')

d.edge('t5', 'sp', exit='bottom', entry=((d.cx(6) - d.cx(0) + 150 / 2 - 8) / (9 * 150 - 16), 0))
d.edge('si', 'i1')
d.edge('i1', 'im')
d.edge('im', 'ig')
d.edge('ig', 'i3', 'Да', exit='right', entry='left')
d.edge('i3', 'ie')
d.edge('ig', 'ig2', 'Нет', exit='bottom', entry='top')
d.edge('ig2', 'i4', 'Запуск', exit='right', entry='left')
d.edge('ig2', 'i5', 'Очередь', exit='bottom', entry='left', via=[(d.cx(3), d.cy('sys', 2))])
d.edge('i4', 'i3', exit='right', entry=(0.3, 1), via=[(d.cx(5) - 18, d.cy('sys', 1))])
d.edge('i5', 'ic')
d.edge('ic', 'i3', exit='top', entry=(0.7, 1), via=[(d.cx(5) + 18, d.cy('sys', 2) - 17)])

# Граничные события на нижней кромке подпроцесса и обработчики под ними
br, hr = d.sub_bottom_row('sp'), d.sub_bottom_row('sp') + 1.05
for col, kind, ev, task, end in [
    (1, 'b_cond', 'Экземпляр аварийно завершился', 'Перезапустить экземпляр', 'Экземпляр восстановлен'),
    (3, 'b_timer', 'Нет игроков дольше таймаута', 'Остановить экземпляр', 'Ресурсы освобождены'),
    (5, 'b_msg', 'Команда из веб-интерфейса', 'Выполнить команду, показать журнал', 'Команда выполнена'),
    (7, 'bi_msg', 'Игра снята с публикации', 'Остановить все экземпляры', 'Серверы остановлены'),
]:
    d.node(f'b{col}', kind, 'sys', col, row=br, label=ev, lpos='right')
    d.node(f'h{col}', 'task_service', 'sys', col, row=hr, label=task)
    d.node(f'h{col}e', 'end', 'sys', col + 1, row=hr, label=end)
    d.edge(f'b{col}', f'h{col}', exit='bottom', entry='top')
    d.edge(f'h{col}', f'h{col}e')

d.note('n1', 't2', 'dev', 4.5, 1, 'Новая возможность: собственные узлы', w=120)
d.note('n2', 'i1', 'sys', 1, 1.1, 'Без участия инженера по эксплуатации', w=130)
d.note('n3', 'ic', 'sys', 6.6, 2, 'Новая возможность: очередь игроков', w=130)

open(sys.argv[1], 'w').write(d.render())
