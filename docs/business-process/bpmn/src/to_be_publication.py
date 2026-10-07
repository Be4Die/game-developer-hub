"""BPMN «как должно быть»: публикация и обновление веб-игры в разрабатываемой системе."""
import sys
from bpmn import Diagram

d = Diagram('Публикация и обновление веб-игры — «как должно быть»', [
    ('dev', 'Разработчик игры', 2),
    ('mod', 'Модератор', 1),
    ('sys', 'Разрабатываемая система', 2),
])

d.node('s', 'start', 'dev', 0, label='Новая игра')
d.node('t1', 'task_user', 'dev', 1, label='Создать проект')
d.node('g1', 'xgw', 'dev', 2, label='Работа<br>в команде?')
d.node('t2', 'task_user', 'dev', 3, row=1, label='Пригласить участников и назначить права')
d.node('g1m', 'xgw', 'dev', 4)
d.node('t3', 'task_user', 'dev', 5, label='Заполнить описание, загрузить промо-материалы')
d.node('t4', 'task_user', 'dev', 6, label='Загрузить клиентскую сборку')
d.node('t5', 'task_service', 'sys', 7, label='Подготовить сборку к проверке')
d.node('t6', 'task_user', 'dev', 8, label='Проверить игру в песочнице')
d.node('t7', 'task_user', 'dev', 9, label='Подать заявку на публикацию или обновление')
d.node('t8', 'task_service', 'sys', 10, label='Проверить данные, сформировать аудит-снимок')
d.node('t9', 'task_user', 'mod', 11, label='Взять заявку в работу')
d.node('t10', 'task_user', 'mod', 12, label='Проверить игру по аудит-снимку')
d.node('g2', 'xgw', 'mod', 13, label='Соответствует<br>требованиям?', lpos='bottom')
d.node('t11', 'task_user', 'dev', 13, label='Доработать игру по замечаниям')
d.node('t12', 'task_service', 'sys', 14, label='Опубликовать версию в каталоге')
d.node('t13', 'task_service', 'sys', 15, label='Получать показатели из сервиса аналитики')
d.node('t14', 'task_user', 'dev', 16, label='Просматривать статистику, выгружать отчёт')
d.node('g3', 'xgw', 'dev', 17, label='Нужно<br>обновление?')
d.node('e1', 'end', 'dev', 18, label='Игра сопровождается')

top1, top2 = d.lane_y('dev', 20), d.lane_y('dev', 7)
d.edge('s', 't1')
d.edge('t1', 'g1')
d.edge('g1', 't2', 'Да', exit='bottom', entry='left', via=[(d.cx(2), d.cy('dev', 1))])
d.edge('g1', 'g1m', 'Нет', exit='right', entry='left')
d.edge('t2', 'g1m', exit='right', entry='bottom', via=[(d.cx(4), d.cy('dev', 1))])
d.edge('g1m', 't3')
d.edge('t3', 't4')
d.edge('t4', 't5')
d.edge('t5', 't6')
d.edge('t6', 't7')
d.edge('t7', 't8')
d.edge('t8', 't9')
d.edge('t9', 't10')
d.edge('t10', 'g2')
d.edge('g2', 't11', 'Нет', exit='top', entry='bottom')
d.edge('t11', 't4', exit='top', entry='top', via=[(d.cx(13), top1), (d.cx(6), top1)])
d.edge('g2', 't12', 'Да', exit='right', entry='top', via=[(d.cx(14), d.cy('mod', 0))])
d.edge('t12', 't13')
d.edge('t13', 't14', exit='right', entry='bottom', via=[(d.cx(16), d.cy('sys', 0))])
d.edge('t14', 'g3')
d.edge('g3', 'e1', 'Нет', exit='right', entry='left')
d.edge('g3', 't4', 'Да', exit='top', entry=(0.75, 0), via=[(d.cx(17), top2), (d.cx(6) + 30, top2)], lx=-0.7)

d.note('n1', 't2', 'dev', 1.1, 1, 'Новая возможность: командная работа', w=130)
d.note('n2', 't5', 'sys', 7, 1, 'Автоматически, без инженера по эксплуатации')
d.note('n3', 't11', 'dev', 12, 1, 'Замечания и вердикт — в сквозном чате проекта')
d.note('n4', 't13', 'sys', 15, 1, 'Данные обновляются ежедневно', w=130)

open(sys.argv[1], 'w').write(d.render())
