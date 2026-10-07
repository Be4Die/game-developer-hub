"""BPMN «как есть»: публикация и обновление веб-игры до внедрения Системы."""
import sys
from bpmn import Diagram

d = Diagram('Публикация и обновление веб-игры — «как есть»', [
    ('dev', 'Разработчик игры', 2),
    ('mod', 'Модератор', 1),
    ('eng', 'Инженер по эксплуатации', 2),
    ('ci', 'GitHub Actions', 1),
    ('mgr', 'Менеджер', 2),
])

d.node('s', 'start', 'dev', 0, label='Игра готова к публикации')
d.node('t1', 'task_user', 'dev', 1, label='Запросить размещение игры в Telegram')
d.node('t2', 'task_user', 'eng', 2, label='Создать репозиторий и настроить конвейер сборки')
d.node('t3', 'task_user', 'dev', 3, label='Загрузить сборку в репозиторий')
d.node('t4', 'task_service', 'ci', 4, label='Развернуть сборку в тестовом окружении')
d.node('t5', 'task_user', 'dev', 5, label='Передать модератору ссылку, описание и промо-материалы')
d.node('t6', 'task_user', 'mod', 6, label='Проверить игру')
d.node('g1', 'xgw', 'mod', 7, label='Соответствует<br>требованиям?', lpos='bottom')
d.node('t7', 'task_user', 'dev', 7, label='Доработать игру по замечаниям')
d.node('t8', 'task_user', 'mod', 8, label='Сообщить о готовности к публикации')
d.node('t9', 'task_user', 'eng', 9, label='Вручную запустить развёртывание в продуктивном окружении')
d.node('t10', 'task_service', 'ci', 10, label='Развернуть игру в продуктивном окружении')
d.node('tm', 'timer', 'mgr', 11, label='Конец месяца')
d.node('t11', 'task_user', 'mgr', 12, label='Сформировать отчёт в настольной программе')
d.node('t12', 'task_user', 'mgr', 13, label='Отправить отчёт разработчику в Telegram')
d.node('t13', 'task_user', 'dev', 14, label='Изучить отчёт за месяц')
d.node('g2', 'xgw', 'dev', 15, label='Нужно<br>обновление?')
d.node('e1', 'end', 'dev', 16, label='Игра сопровождается')

top1, top2 = d.lane_y('dev', 20), d.lane_y('dev', 7)
d.edge('s', 't1')
d.edge('t1', 't2')
d.edge('t2', 't3')
d.edge('t3', 't4')
d.edge('t4', 't5')
d.edge('t5', 't6')
d.edge('t6', 'g1')
d.edge('g1', 't7', 'Нет', exit='top', entry='bottom')
d.edge('t7', 't3', exit='top', entry='top', via=[(d.cx(7), top1), (d.cx(3), top1)])
d.edge('g1', 't8', 'Да', exit='right', entry='left')
d.edge('t8', 't9')
d.edge('t9', 't10')
d.edge('t10', 'tm', exit='right', entry='top', via=[(d.cx(11), d.cy('ci', 0))])
d.edge('tm', 't11')
d.edge('t11', 't12')
d.edge('t12', 't13', exit='right', entry='bottom', via=[(d.cx(14), d.cy('mgr', 0))])
d.edge('t13', 'g2')
d.edge('g2', 'e1', 'Нет', exit='right', entry='left')
d.edge('g2', 't3', 'Да', exit='top', entry=(0.75, 0), via=[(d.cx(15), top2), (d.cx(3) + 30, top2)], lx=-0.7)

d.note('n1', 't2', 'eng', 2, 1, 'Настройка вручную для каждой игры')
d.note('n2', 't5', 'dev', 5.5, 1, 'Переписка в мессенджере, история теряется', w=140)
d.note('n3', 't9', 'eng', 9, 1, 'Выпуск игры зависит от инженера')
d.note('n4', 't11', 'mgr', 12, 1, 'Данные доступны раз в месяц')

open(sys.argv[1], 'w').write(d.render())
