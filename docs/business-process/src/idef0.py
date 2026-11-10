"""IDEF0-диаграммы (контекстная A-0 и декомпозиция A0) в формате draw.io.

Блоки стоят по диагонали, стрелки прокладываются по правилам IDEF0: входы слева, управление сверху,
выходы справа, механизмы снизу. Запуск: python3 idef0.py  — пишет ../<as-is|to-be>/<context|decompose>.drawio
"""
import pathlib
from xml.sax.saxutils import escape

FONT = 'fontFamily=Times New Roman;fontSize=13;'
BOX = 'rounded=0;whiteSpace=wrap;html=1;spacing=8;fillColor=#FFFFFF;strokeColor=#000000;strokeWidth=1.5;' + FONT
TXT = 'text;html=1;whiteSpace=wrap;labelBackgroundColor=#FFFFFF;verticalAlign=middle;fontFamily=Times New Roman;fontSize=12;'
EDGE = 'edgeStyle=none;rounded=0;html=1;endArrow=block;endFill=1;endSize=7;jumpStyle=arc;jumpSize=8;strokeColor=#000000;'
BW, BH, DX, DY, X0, Y0, LEFT, TOP = 210, 86, 272, 138, 290, 150, 20, 20


class Diagram:
    def __init__(self, title, boxes, numbers):
        self.cells, self.n = [], 0
        self.title, self.boxes = title, boxes
        k = len(boxes)
        if k == 1:
            self.pos = [(380, 250)]
            self.bw, self.bh = 480, 150
        else:
            self.pos = [(X0 + i * DX, Y0 + i * DY) for i in range(k)]
            self.bw, self.bh = BW, BH
        self.right = self.pos[-1][0] + self.bw + 270
        self.bottom = self.pos[-1][1] + self.bh + 190
        self.sides = [{'l': [], 't': [], 'r': [], 'b': []} for _ in boxes]
        self.arrows, self.mechs = [], []
        for i, (name, num) in enumerate(zip(boxes, numbers)):
            x, y = self.pos[i]
            self.v(name, BOX, x, y, self.bw, self.bh)
            self.v(num, TXT.replace('labelBackgroundColor=#FFFFFF;', '') + 'align=right;', x + self.bw - 44, y + self.bh - 20, 40, 18)

    def v(self, value, style, x, y, w, h):
        self.n += 1
        self.cells.append(f'<mxCell id="c{self.n}" value="{escape(value, {chr(34): "&quot;"})}" style="{style}" vertex="1" parent="1">'
                          f'<mxGeometry x="{x}" y="{y}" width="{w}" height="{h}" as="geometry"/></mxCell>')

    def line(self, pts, arrow=True):
        self.n += 1
        mid = ''.join(f'<mxPoint x="{x}" y="{y}"/>' for x, y in pts[1:-1])
        st = EDGE if arrow else EDGE.replace('endArrow=block', 'endArrow=none')
        self.cells.append(f'<mxCell id="c{self.n}" style="{st}" edge="1" parent="1"><mxGeometry relative="1" as="geometry">'
                          f'<mxPoint x="{pts[0][0]}" y="{pts[0][1]}" as="sourcePoint"/><mxPoint x="{pts[-1][0]}" y="{pts[-1][1]}" as="targetPoint"/>'
                          f'<Array as="points">{mid}</Array></mxGeometry></mxCell>')

    def label(self, text, x, y, w=190, align='left', h=30):
        self.v(text, TXT + f'align={align};', x, y - h / 2, w, h)

    # ── описание стрелок ──
    def inp(self, box, text, src=None):
        self.arrows.append(('in', box, text, src)); self.sides[box]['l'].append(len(self.arrows) - 1)
        if src is not None:
            self._out_slot(src, len(self.arrows) - 1, text)

    def ctl(self, box, text, src=None):
        self.arrows.append(('ctl', box, text, src)); self.sides[box]['t'].append(len(self.arrows) - 1)
        if src is not None:
            self._out_slot(src, len(self.arrows) - 1, text)

    def out(self, box, text):
        self.arrows.append(('out', box, text, None)); self._out_slot(box, len(self.arrows) - 1, text)

    def _out_slot(self, box, idx, text):
        """Выходы с одинаковой подписью делят одну точку на правой грани блока."""
        for j in self.sides[box]['r']:
            if self.arrows[j][2] == text:
                self.alias = getattr(self, 'alias', {}); self.alias[idx] = j
                return
        self.sides[box]['r'].append(idx)

    def mech(self, text, boxes):
        self.mechs.append((text, boxes))
        for b in boxes:
            self.sides[b]['b'].append(('m', len(self.mechs) - 1))

    def slot(self, box, side, key):
        lst = self.sides[box][side]
        key = getattr(self, 'alias', {}).get(key, key) if side == 'r' else key
        i, n = lst.index(key), len(lst)
        x, y = self.pos[box]
        if side in 'lr':
            return (x if side == 'l' else x + self.bw, round(y + self.bh * (i + 1) / (n + 1)))
        return (round(x + self.bw * (i + 1) / (n + 1)), y if side == 't' else y + self.bh)

    def build(self):
        chan = {}
        for idx, (kind, box, text, src) in enumerate(self.arrows):
            if kind == 'in' and src is None:
                x, y = self.slot(box, 'l', idx)
                self.line([(LEFT, y), (x, y)]); self.label(text, LEFT + 4, y - 13, 275, h=22)
            elif kind == 'ctl' and src is None:
                x, y = self.slot(box, 't', idx)
                self.line([(x, TOP), (x, y)])
                if len(self.boxes) == 1:
                    self.label(text, x + 6, TOP + 30, self.bw / (len(self.sides[box]['t']) + 1) - 12, h=40)
                else:
                    self.label(text, x + 6, TOP + 22 + 30 * self.sides[box]['t'].index(idx), 170, h=30)
            elif kind == 'out':
                x, y = self.slot(box, 'r', idx)
                self.line([(x, y), (self.right, y)]); self.label(text, self.right - 214, y - 13, 210, 'right', h=22)
            elif kind == 'in':  # выход src -> вход box
                sx, sy = self.slot(src, 'r', idx)
                tx, ty = self.slot(box, 'l', idx)
                k = chan.get(box, 0); chan[box] = k + 1
                cx = tx - 26 - 16 * k
                self.line([(sx, sy), (cx, sy), (cx, ty), (tx, ty)])
                if not any(a[0] == 'out' and a[1] == src and a[2] == text for a in self.arrows):
                    self.v(text, TXT.replace('fontSize=12', 'fontSize=11') + 'align=left;', cx + 5, sy + 6, tx - cx + 52, 46)
            elif kind == 'ctl':
                sx, sy = self.slot(src, 'r', idx)
                tx, ty = self.slot(box, 't', idx)
                if src < box:  # выход -> управление блока ниже
                    self.line([(sx, sy), (tx, sy), (tx, ty)]); self.label(text, tx - 176, sy - 13, 170, 'right', h=22)
                else:  # обратная связь
                    ux, uy = sx + 34, ty - 34
                    self.line([(sx, sy), (ux, sy), (ux, uy), (tx, uy), (tx, ty)]); self.label(text, tx + 10, uy - 14, 160, h=22)
        base = self.pos[-1][1] + self.bh + 40
        for m, (text, boxes) in enumerate(self.mechs):
            by = base + 24 * m
            tail = 150 + m * ((self.right - 300) / max(1, len(self.mechs) - 1)) if len(self.mechs) > 1 else self.pos[0][0] + self.bw / 2
            xs = [self.slot(b, 'b', ('m', m))[0] for b in boxes]
            if len(self.boxes) == 1:
                x, y = self.slot(0, 'b', ('m', m))
                low = 46 if m % 2 else 0
                self.line([(x, self.bottom - 86 + low), (x, y)])
                self.label(text, x - 62, self.bottom - 68 + low, 124, 'center', h=34)
                continue
            lo, hi = min(xs + [tail]), max(xs + [tail])
            self.line([(tail, self.bottom - 30), (tail, by)], arrow=False)
            self.line([(lo, by), (hi, by)], arrow=False)
            for b in boxes:
                x, y = self.slot(b, 'b', ('m', m))
                self.line([(x, by), (x, y)])
            self.label(text, tail - 95, self.bottom - 14, 190, 'center', h=30)
        self.v(self.title, TXT + 'align=center;fontStyle=1;fontSize=13;', LEFT, self.bottom + 14, self.right - LEFT, 22)
        return ('<mxfile><diagram name="IDEF0"><mxGraphModel grid="0" page="0"><root>\n<mxCell id="0"/>\n<mxCell id="1" parent="0"/>\n'
                + '\n'.join(self.cells) + '\n</root></mxGraphModel></diagram></mxfile>')


def save(d, path):
    out = pathlib.Path(__file__).resolve().parents[1] / path
    out.write_text(d.build()); print(out)


# ══ «Как есть» ══════════════════════════════════════════════
d = Diagram('A-0. Публикация и сопровождение веб-игр («как есть»)', ['Публиковать и сопровождать веб-игры'], ['A0'])
for t in ['Клиентская сборка игры', 'Описание и промо-материалы', 'Серверная сборка игры', 'Запросы на управление серверами']:
    d.inp(0, t)
for t in ['Требования площадки к играм', 'Руководство разработчика', 'Регламент выкладки']:
    d.ctl(0, t)
for t in ['Опубликованная игра', 'Работающий игровой сервер', 'Журналы сервера по запросу', 'Ежемесячный отчёт']:
    d.out(0, t)
for t in ['Разработчик игры', 'Модератор', 'Инженер по эксплуатации', 'Менеджер', 'Мессенджер', 'Git-хостинг и сценарии выкладки']:
    d.mech(t, [0])
save(d, 'as-is/context.drawio')

d = Diagram('A0. Публикация и сопровождение веб-игр («как есть»)',
            ['Передать материалы игры', 'Развернуть игру в тестовом окружении', 'Проверить игру',
             'Опубликовать игру и сопровождать серверы', 'Подготовить отчёт'], ['A1', 'A2', 'A3', 'A4', 'A5'])
d.inp(0, 'Клиентская сборка игры'); d.inp(0, 'Описание и промо-материалы'); d.inp(0, 'Серверная сборка игры')
d.ctl(0, 'Руководство разработчика')
d.inp(1, 'Материалы игры в репозитории', src=0); d.ctl(1, 'Регламент выкладки')
d.inp(2, 'Тестовая версия игры', src=1); d.ctl(2, 'Требования площадки к играм')
d.ctl(0, 'Замечания', src=2)
d.inp(3, 'Разрешение на публикацию', src=2); d.inp(3, 'Запросы на управление серверами'); d.ctl(3, 'Регламент выкладки')
d.out(3, 'Опубликованная игра'); d.out(3, 'Работающий игровой сервер'); d.out(3, 'Журналы сервера по запросу')
d.inp(4, 'Опубликованная игра', src=3); d.out(4, 'Ежемесячный отчёт')
d.mech('Разработчик игры', [0]); d.mech('Мессенджер', [0, 2, 3]); d.mech('Git-хостинг и сценарии выкладки', [0, 1, 3])
d.mech('Инженер по эксплуатации', [1, 3]); d.mech('Модератор', [2]); d.mech('Менеджер', [4])
save(d, 'as-is/decompose.drawio')

# ══ «Как должно быть» ═══════════════════════════════════════
d = Diagram('A-0. Публикация и сопровождение веб-игр («как должно быть»)', ['Публиковать и сопровождать веб-игры'], ['A0'])
for t in ['Клиентская сборка игры', 'Описание и промо-материалы', 'Серверная сборка игры', 'Политика оркестрации и команды управления']:
    d.inp(0, t)
for t in ['Требования площадки к играм', 'Руководство разработчика']:
    d.ctl(0, t)
for t in ['Опубликованная игра', 'Доступный игровой сервер', 'Журналы и показатели серверов', 'Аналитика проекта']:
    d.out(0, t)
for t in ['Разработчик игры и участники проекта', 'Модератор', 'Разрабатываемая система', 'Вычислительные узлы']:
    d.mech(t, [0])
save(d, 'to-be/context.drawio')

d = Diagram('A0. Публикация и сопровождение веб-игр («как должно быть»)',
            ['Подготовить проект и проверить игру в песочнице', 'Провести модерацию', 'Опубликовать игру',
             'Управлять игровыми серверами', 'Анализировать показатели игры'], ['A1', 'A2', 'A3', 'A4', 'A5'])
d.inp(0, 'Клиентская сборка игры'); d.inp(0, 'Описание и промо-материалы'); d.ctl(0, 'Руководство разработчика')
d.inp(1, 'Черновик проекта с заявкой', src=0); d.ctl(1, 'Требования площадки к играм')
d.ctl(0, 'Замечания', src=1)
d.ctl(3, 'Квота серверов платформы', src=1)
d.inp(2, 'Решение о публикации', src=1)
d.out(2, 'Опубликованная игра')
d.inp(3, 'Серверная сборка игры'); d.inp(3, 'Политика оркестрации и команды управления')
d.out(3, 'Доступный игровой сервер'); d.out(3, 'Журналы и показатели серверов')
d.inp(4, 'Опубликованная игра', src=2); d.out(4, 'Аналитика проекта')
d.mech('Разработчик игры и участники проекта', [0, 3, 4]); d.mech('Модератор', [1])
d.mech('Разрабатываемая система', [0, 1, 2, 3, 4]); d.mech('Вычислительные узлы', [3])
save(d, 'to-be/decompose.drawio')
