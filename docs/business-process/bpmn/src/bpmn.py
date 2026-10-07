"""Генератор BPMN-диаграмм в формате draw.io.

Пул с дорожками, узлы раскладываются по сетке (столбец, строка внутри дорожки).
Поддерживаются задачи, события, шлюзы, текстовые аннотации и событийные подпроцессы.
"""
from xml.sax.saxutils import escape

FONT = 'fontFamily=Times New Roman;fontSize=12;'
FONT_SMALL = 'fontFamily=Times New Roman;fontSize=11;'
COLW, ROWH, TOPPAD = 150, 100, 26
POOL_X, POOL_Y, HDR = 20, 20, 30
TASK_W, TASK_H, GW, EV = 122, 64, 46, 34

# Текстовая аннотация: прямоугольник со штриховой линией.
ANNOTATION = 'rounded=0;dashed=1;fillColor=none;html=1;whiteSpace=wrap;align=left;spacingLeft=8;spacingRight=4;fontStyle=2;'

TASK_MARKERS = {'task': 'abstract', 'task_user': 'user', 'task_service': 'service', 'task_manual': 'manual'}

# Вид события: (контур, символ). Контур eventNonint — пунктирный, для стартов событийных подпроцессов.
EVENTS = {
    'start': ('standard', 'general'),
    'end': ('end', 'general'),
    'timer': ('catching', 'timer'),
    'cond': ('catching', 'conditional'),
    'msg': ('catching', 'message'),
    'sub_msg': ('eventNonint', 'message'),
    'sub_timer': ('eventNonint', 'timer'),
    'sub_cond': ('eventNonint', 'conditional'),
    # Граничные события: непрерывающие (пунктир) и прерывающее
    'b_cond': ('boundNonint', 'conditional'),
    'b_timer': ('boundNonint', 'timer'),
    'b_msg': ('boundNonint', 'message'),
    'bi_msg': ('boundInt', 'message'),
}


class Diagram:
    def __init__(self, title, lanes):
        # lanes: [(id, подпись, число строк)]
        self.title, self.lanes = title, lanes
        self.nodes, self.edges, self.notes, self.subs = {}, [], [], []
        self.lane_top, y = {}, POOL_Y
        for lid, _, rows in lanes:
            self.lane_top[lid] = y
            y += TOPPAD + rows * ROWH + 10
        self.bottom = y

    def node(self, nid, kind, lane, col, row=0, label='', **kw):
        self.nodes[nid] = dict(kind=kind, lane=lane, col=col, row=row, label=label, **kw)

    def edge(self, src, dst, label='', exit=None, entry=None, via=None, lx=None):
        self.edges.append(dict(src=src, dst=dst, label=label, exit=exit, entry=entry, via=via or [], lx=lx))

    def note(self, nid, target, lane, col, row, text, w=150):
        self.notes.append(dict(id=nid, target=target, lane=lane, col=col, row=row, text=text, w=w))

    def subprocess(self, sid, lane, col0, col1, row0, row1, label, expanded=False):
        """Рамка подпроцесса вокруг ячеек [col0..col1] × [row0..row1]: событийный — пунктир, развёрнутый — сплошная."""
        self.subs.append(dict(id=sid, lane=lane, c0=col0, c1=col1, r0=row0, r1=row1, label=label, expanded=expanded))

    def sub_bottom_row(self, sid):
        """Строка сетки, центр которой лежит на нижней кромке рамки — для граничных событий."""
        sub = next(x for x in self.subs if x['id'] == sid)
        return sub['r1'] + 0.6

    # Центр ячейки сетки в абсолютных координатах.
    def cx(self, col):
        return POOL_X + 2 * HDR + 10 + col * COLW + COLW / 2

    def cy(self, lane, row):
        return self.lane_top[lane] + TOPPAD + row * ROWH + ROWH / 2

    def lane_y(self, lane, offset=13):
        """Горизонталь у верхней кромки дорожки — для обратных стрелок."""
        return self.lane_top[lane] + offset

    def render(self):
        cols = [n['col'] for n in self.nodes.values()] + [a['col'] for a in self.notes] + [s['c1'] for s in self.subs]
        width = 2 * HDR + 20 + (max(cols) + 1) * COLW
        cells = ['<mxCell id="0"/>', '<mxCell id="1" parent="0"/>']
        cells.append(self._v('pool', self.title, 'swimlane;html=1;horizontal=0;startSize=30;fillColor=none;fontStyle=1;' + FONT,
                             POOL_X, POOL_Y, width, self.bottom - POOL_Y, '1'))
        for lid, label, rows in self.lanes:
            h = TOPPAD + rows * ROWH + 10
            cells.append(self._v('lane_' + lid, label, 'swimlane;html=1;horizontal=0;startSize=30;fillColor=none;fontStyle=0;' + FONT,
                                 HDR, self.lane_top[lid] - POOL_Y, width - HDR, h, 'pool'))
        for s in self.subs:  # рамки подпроцессов рисуются раньше узлов, чтобы быть под ними
            x0, x1 = self.cx(s['c0']) - COLW / 2 + 8, self.cx(s['c1']) + COLW / 2 - 8
            y0 = self.cy(s['lane'], s['r0']) - ROWH / 2 - 16
            y1 = self.cy(s['lane'], s['r1']) + ROWH / 2 + 10
            line = 'strokeWidth=1.5;' if s['expanded'] else 'dashed=1;dashPattern=1 3;'
            style = ('rounded=1;arcSize=4;' + line + 'fillColor=none;html=1;whiteSpace=wrap;'
                     'verticalAlign=top;align=left;spacingLeft=8;spacingTop=0;fontStyle=1;' + FONT_SMALL)
            cells.append(self._v(s['id'], s['label'], style, x0 - POOL_X - HDR, y0 - self.lane_top[s['lane']],
                                 x1 - x0, y1 - y0, 'lane_' + s['lane']))
        for nid, n in self.nodes.items():
            cells.append(self._node(nid, n))
        for a in self.notes:
            x, y = self.cx(a['col']) - a['w'] / 2, self.cy(a['lane'], a['row']) - 26
            cells.append(self._v(a['id'], a['text'], ANNOTATION + FONT_SMALL,
                                 x - POOL_X - HDR, y - self.lane_top[a['lane']], a['w'], 52, 'lane_' + a['lane']))
            cells.append(f'<mxCell id="{a["id"]}_as" style="html=1;dashed=1;endArrow=none;" edge="1" parent="1" '
                         f'source="{a["id"]}" target="{a["target"]}"><mxGeometry relative="1" as="geometry"/></mxCell>')
        for i, e in enumerate(self.edges):
            cells.append(self._edge(f'flow{i}', e))
        body = '\n'.join(cells)
        return (f'<mxfile><diagram name="{escape(self.title)}"><mxGraphModel grid="0" page="0">'
                f'<root>\n{body}\n</root></mxGraphModel></diagram></mxfile>')

    def _v(self, cid, value, style, x, y, w, h, parent):
        return (f'<mxCell id="{cid}" value="{escape(value, {chr(34): "&quot;"})}" style="{style}" vertex="1" parent="{parent}">'
                f'<mxGeometry x="{x:.0f}" y="{y:.0f}" width="{w:.0f}" height="{h:.0f}" as="geometry"/></mxCell>')

    def _node(self, nid, n):
        kind = n['kind']
        x, y = self.cx(n['col']), self.cy(n['lane'], n['row'])
        if kind in TASK_MARKERS:
            w, h = TASK_W, TASK_H
            style = (f'shape=mxgraph.bpmn.task2;whiteSpace=wrap;html=1;rectStyle=rounded;size=10;taskMarker={TASK_MARKERS[kind]};'
                     'spacingTop=12;spacingLeft=4;spacingRight=4;' + FONT)
        elif kind in ('xgw', 'pgw', 'egw'):
            w = h = GW
            gw = {'xgw': 'exclusive', 'pgw': 'parallel', 'egw': 'eventBased'}[kind]
            lab = {'top': 'verticalLabelPosition=top;verticalAlign=bottom;',
                   'bottom': 'verticalLabelPosition=bottom;verticalAlign=top;',
                   'left': 'labelPosition=left;align=right;verticalAlign=middle;',
                   'right': 'labelPosition=right;align=left;verticalAlign=middle;'}[n.get('lpos', 'top')]
            style = f'shape=mxgraph.bpmn.gateway2;html=1;gwType={gw};{lab}' + FONT
        else:
            w = h = EV
            outline, symbol = EVENTS[kind]
            thick = 'strokeWidth=3;' if outline == 'end' else ''
            lab = ('labelPosition=right;verticalLabelPosition=bottom;align=left;verticalAlign=top;spacingLeft=-4;spacingTop=-6;labelWidth=110;' if n.get('lpos') == 'right'
                   else 'verticalLabelPosition=bottom;verticalAlign=top;labelWidth=120;')
            style = (f'shape=mxgraph.bpmn.event;html=1;whiteSpace=wrap;outline={outline};symbol={symbol};{thick}' + lab + FONT)
        return self._v(nid, n['label'], style, x - w / 2 - POOL_X - HDR, y - h / 2 - self.lane_top[n['lane']],
                       w, h, 'lane_' + n['lane'])

    def _edge(self, eid, e):
        side = {'right': (1, 0.5), 'left': (0, 0.5), 'top': (0.5, 0), 'bottom': (0.5, 1)}
        style = 'edgeStyle=orthogonalEdgeStyle;rounded=0;html=1;endArrow=block;endFill=1;jumpStyle=arc;' + FONT_SMALL
        for key, prefix in (('exit', 'exit'), ('entry', 'entry')):
            if e[key]:
                px, py = side[e[key]] if isinstance(e[key], str) else e[key]
                style += f'{prefix}X={px};{prefix}Y={py};{prefix}Dx=0;{prefix}Dy=0;'
        pts = ''.join(f'<mxPoint x="{x:.0f}" y="{y:.0f}"/>' for x, y in e['via'])
        pts = f'<Array as="points">{pts}</Array>' if pts else ''
        off = f' x="{e["lx"]}"' if e['lx'] is not None else ''
        return (f'<mxCell id="{eid}" value="{escape(e["label"])}" style="{style}" edge="1" parent="1" '
                f'source="{e["src"]}" target="{e["dst"]}"><mxGeometry relative="1"{off} as="geometry">{pts}</mxGeometry></mxCell>')
