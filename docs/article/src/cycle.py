"""BPMN для статьи: один проход цикла контроля узлов и применения политик в сервисе оркестрации."""
import sys
from compact import Diagram

d = Diagram('Цикл контроля', [('orc', 'Сервис оркестрации', 2.2)])

d.node('s', 'timer', 'orc', 0, label='Каждые 15 секунд')
d.node('t1', 'task_service', 'orc', 1, label='Запросить сигнал присутствия узла')
d.node('g1', 'xgw', 'orc', 2, label='Узел<br>ответил?')
d.node('t2', 'task_service', 'orc', 3, label='Сверить экземпляры, пропавшие — «Сбой»')
d.node('g2', 'xgw', 'orc', 3, row=1, label='Тишина дольше 60 с?', lpos='top')
d.node('t3', 'task_service', 'orc', 4, row=1, label='Узел «Не в сети», экземпляры «Сбой»')
d.node('g3', 'xgw', 'orc', 4)
d.node('t4', 'task_service', 'orc', 5, label='Перезапустить сбойные экземпляры по политике')
d.node('t5', 'task_service', 'orc', 6, label='Применить политики игр и очереди')
d.node('e', 'end', 'orc', 6, row=1, label='Проход завершён')

d.edge('s', 't1')
d.edge('t1', 'g1')
d.edge('g1', 't2', 'Да', exit='right', entry='left')
d.edge('g1', 'g2', 'Нет', exit='bottom', entry='left', via=[(d.cx(2), d.cy('orc', 1))])
d.edge('g2', 't3', 'Да')
d.edge('g2', 'g3', 'Нет', exit='bottom', entry='bottom', via=[(d.cx(3), d.cy('orc', 1) + 50), (d.cx(4) + 70, d.cy('orc', 1) + 50), (d.cx(4) + 70, d.cy('orc', 0) + 42), (d.cx(4), d.cy('orc', 0) + 42)])
d.edge('t2', 'g3')
d.edge('t3', 'g3', exit='top', entry='bottom')
d.edge('g3', 't4')
d.edge('t4', 't5')
d.edge('t5', 'e', exit='bottom', entry='top')


open(sys.argv[1], 'w').write(d.render())
