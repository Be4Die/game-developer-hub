// Презентация к предзащите ВКР: «Автоматизированная система управления процессом публикации и сопровождения веб-игр».
// Объекты с именем «aN_…» появляются по щелчку (группа N) — анимацию добавляет postprocess.py.
// Схемы BPMN на слайдах — упрощённые версии схем из docs/business-process/bpmn, нарисованные фигурами,
// чтобы текст оставался крупным; полные схемы идут в приложение записки.
const pptxgen = require('pptxgenjs');
const SPEECH = require('./speech.js');

const pres = new pptxgen();
pres.layout = 'LAYOUT_16x9'; // 10 × 5.625 дюйма
pres.author = 'М. В. Наследсков';
pres.title = 'Автоматизированная система управления процессом публикации и сопровождения веб-игр';

const FONT = 'Times New Roman';
// Палитра: белый, чёрный, спокойный синий акцент, зелёный — хорошо, красный — плохо.
const C = {
  black: '000000', white: 'FFFFFF', accent: '2B5C8F', green: '2E7D32', red: 'C62828',
  soft: 'F2F2F2', greenSoft: 'EAF5EA',
};
const W = 10;
const IMG = (n) => `img/${n}`;

pres.defineSlideMaster({ title: 'PLAIN', background: { color: C.white } });

const text = (s, t, o) => s.addText(t, { fontFace: FONT, color: C.black, margin: 0, isTextBox: true, ...o });
const num = (s) => { s.slideNumber = { x: 9.3, y: 5.25, w: 0.45, h: 0.3, fontFace: FONT, fontSize: 12, color: C.black, align: 'right' }; };
// Отрезок: горизонтальный или вертикальный, со стрелкой на конце по желанию
const seg = (s, x1, y1, x2, y2, o = {}) => s.addShape(pres.shapes.LINE, {
  x: Math.min(x1, x2), y: Math.min(y1, y2), w: Math.abs(x2 - x1) || 0.001, h: Math.abs(y2 - y1) || 0.001,
  flipH: x2 < x1, flipV: y2 < y1, line: { color: C.black, width: 1, ...o },
});

// Заметки докладчика — текст речи из speech.js, по порядку слайдов
let speechIndex = 0;
function speechNotes(s) {
  const sp = SPEECH[speechIndex++];
  s.addNotes([...sp.text, ...(sp.next ? ['Переход: ' + sp.next] : [])].join('\n\n'));
}

function slideTitle(title, sub) {
  const s = pres.addSlide({ masterName: 'PLAIN' });
  text(s, title.toUpperCase(), { x: 0.3, y: 0.15, w: 9.4, h: 0.6, fontSize: 28, bold: true, align: 'center', valign: 'middle' });
  if (sub) text(s, sub, { x: 0.3, y: 0.7, w: 9.4, h: 0.3, fontSize: 15, color: C.accent, align: 'center' });
  return s;
}

function marker(s, n, x, y, color, name, d = 0.3) {
  s.addShape(pres.shapes.OVAL, { x, y, w: d, h: d, fill: { color }, line: { color: C.white, width: 1.25 }, objectName: name });
  text(s, String(n), { x, y, w: d, h: d, fontSize: 13, bold: true, color: C.white, align: 'center', valign: 'middle', objectName: name + '_n' });
}

// ── 1. Титульный слайд ─────────────────────────────────────
{
  const s = pres.addSlide({ masterName: 'PLAIN' });
  s.addImage({ path: IMG('agtu.png'), x: 0.35, y: 0.25, w: 1.35, h: 1.35 });
  text(s, [
    { text: 'Федеральное агентство по рыболовству', options: { breakLine: true } },
    { text: 'Федеральное государственное бюджетное учреждение высшего образования', options: { breakLine: true } },
    { text: '«Астраханский государственный технический университет»', options: { breakLine: true } },
    { text: 'Система менеджмента качества в области образования, воспитания, науки и инноваций сертифицирована ООО «ДКС РУС» по международному стандарту ISO 9001:2015', options: { fontSize: 9 } },
  ], { x: 1.9, y: 0.25, w: 7.8, h: 1.2, fontSize: 12, align: 'center', valign: 'middle' });
  text(s, [
    { text: 'Кафедра ' },
    { text: 'Автоматизированные системы обработки информации и управления', options: { underline: { style: 'sng' } } },
  ], { x: 1.9, y: 1.5, w: 7.8, h: 0.3, fontSize: 12, align: 'center' });
  text(s, [
    { text: 'ВЫПУСКНАЯ КВАЛИФИКАЦИОННАЯ РАБОТА', options: { bold: true, fontSize: 22, breakLine: true } },
    { text: 'по теме:', options: { fontSize: 13, breakLine: true } },
    { text: '«Автоматизированная система управления процессом публикации и сопровождения веб‑игр»', options: { fontSize: 20 } },
  ], { x: 0.6, y: 2.0, w: 8.8, h: 1.45, align: 'center', valign: 'middle' });
  text(s, [
    { text: 'Выполнил: обучающийся гр. ДИПРБ-41', options: { breakLine: true } },
    { text: 'Наследсков Михаил Викторович', options: { breakLine: true } },
    { text: ' ', options: { breakLine: true, fontSize: 6 } },
    { text: 'Руководитель ВКР: к.т.н., доцент', options: { breakLine: true } },
    { text: 'Морозов А. В.' },
  ], { x: 5.6, y: 3.6, w: 4.1, h: 1.25, fontSize: 14, valign: 'top' });
  text(s, 'Астрахань 2026 г.', { x: 3.5, y: 5.15, w: 3, h: 0.3, fontSize: 12, align: 'center' });
  speechNotes(s);
}

// ── 2. Анализ предметной области: рынок веб-игр ────────────
{
  const s = slideTitle('Анализ предметной области');
  text(s, [
    { text: 'Веб-игра', options: { bold: true } },
    { text: ' — приложение, которое запускается в браузере без скачивания и установки. Игры распространяются через площадки, которые принимают, проверяют и публикуют их.' },
  ], { x: 0.4, y: 0.95, w: 9.2, h: 0.6, fontSize: 14.5, align: 'center', valign: 'middle' });
  // Два графика рядом, под каждым — подпись с ключевым фактом
  const chart = (x, title, labels, values, fmt) => s.addChart(pres.charts.BAR, [{ name: title, labels, values }], {
    x, y: 1.65, w: 4.3, h: 2.0, barDir: 'bar', chartColors: [C.accent], barGapWidthPct: 45,
    showTitle: true, title, titleFontFace: FONT, titleFontSize: 13, titleColor: C.black,
    showValue: true, dataLabelPosition: 'outEnd', dataLabelFontFace: FONT, dataLabelFontSize: 13, dataLabelColor: C.black, dataLabelFormatCode: fmt,
    catAxisLabelFontFace: FONT, catAxisLabelFontSize: 13, catAxisLabelColor: C.black, valAxisHidden: true,
    valAxisMaxVal: Math.max(...values) * 1.25, catAxisOrientation: 'maxMin', valGridLine: { style: 'none' }, catGridLine: { style: 'none' }, showLegend: false,
  });
  chart(0.4, 'Игроки Яндекс Игр в месяц, млн', ['2023', '2024'], [33, 45], '0');
  // 2025 г. — 15 тыс. по отчёту; 2024 и 2023 гг. — пересчёт по кратностям роста (×2,7 и ×4,9) из того же отчёта
  chart(5.3, 'Новые веб-игры за I полугодие, тыс.', ['2023', '2024', '2025'], [3.1, 5.6, 15], '0.0');
  text(s, 'Около 28 тыс. игр в каталоге; за 2024 г. опубликовано более 29 тыс. новых игр и удалено более 30 тыс., не прошедших требования',
    { x: 0.4, y: 3.7, w: 4.3, h: 0.62, fontSize: 12, align: 'center', valign: 'top' });
  text(s, 'В I полугодии 2025 г. выпущено более 15 тыс. веб-игр — в 2,7 раза больше, чем годом ранее',
    { x: 5.3, y: 3.7, w: 4.3, h: 0.62, fontSize: 12, align: 'center', valign: 'top' });
  s.addShape(pres.shapes.LINE, { x: 0.4, y: 4.42, w: 9.2, h: 0, line: { color: C.accent, width: 1 } });
  text(s, 'Игр кратно больше, требования жёстче — ручной приём, проверка и выпуск не успевают за ростом',
    { x: 0.4, y: 4.48, w: 9.2, h: 0.4, fontSize: 14, bold: true, color: C.accent, align: 'center', valign: 'middle' });
  text(s, 'Источники: App2Top, интервью руководителя развития бизнеса Яндекс Игр об итогах 2024 г. (24.01.2025); Playgama, отчёт о рынке веб-игр за I полугодие 2025 г. (29.07.2025)',
    { x: 0.4, y: 4.95, w: 8.7, h: 0.4, fontSize: 9.5, color: '404040', valign: 'top' });
  num(s);
  speechNotes(s);
}

// ── 3. Заказчик ────────────────────────────────────────────
{
  const s = slideTitle('Заказчик');
  s.addImage({ path: IMG('welwise-hd.png'), x: 0.4, y: 0.98, w: 1.75, h: 1.75 * 367 / 2158 });
  const box = (t, cx, y, w, hl, h = 0.42) => {
    s.addShape(pres.shapes.RECTANGLE, { x: cx - w / 2, y, w, h, fill: { color: C.white }, line: { color: hl ? C.accent : C.black, width: hl ? 2 : 0.75 } });
    text(s, t, { x: cx - w / 2 + 0.03, y, w: w - 0.06, h, fontSize: 9.5, align: 'center', valign: 'middle', color: hl ? C.accent : C.black, bold: !!hl });
  };
  // Связи «родитель — дети»: вниз, горизонталь, вниз
  const fork = (px, py, kids, ky) => {
    const my = (py + ky) / 2;
    seg(s, px, py, px, my, { width: 0.75 });
    if (kids.length > 1 || kids[0] !== px) seg(s, Math.min(px, ...kids), my, Math.max(px, ...kids), my, { width: 0.75 });
    kids.forEach((k) => seg(s, k, my, k, ky, { width: 0.75 }));
  };
  const K = [0.875, 1.895, 2.915, 3.935, 4.955, 5.975], R = 3.6, L = [0.98, 1.55, 2.15, 2.75];
  fork(R, L[0] + 0.42, [R], L[1]);
  fork(R, L[1] + 0.42, [K[1], K[3], 5.465], L[2]);
  fork(K[1], L[2] + 0.42, [K[0], K[1], K[2]], L[3]);
  fork(K[3], L[2] + 0.42, [K[3]], L[3]);
  fork(5.465, L[2] + 0.42, [K[4], K[5]], L[3]);
  box('Основатель ИП Теплов С. Д.', R, L[0], 2.1);
  box('Соучредитель / ведущий менеджер', R, L[1], 2.1, true);
  box('Технический отдел', K[1], L[2], 1.45);
  box('Отдел модерации', K[3], L[2], 0.98);
  box('Студия разработки игр', 5.465, L[2], 1.45);
  ['Разработчик серверной части', 'Разработчик клиентской части', 'Инженер по эксплуатации', 'Специалисты по модерации (2)', 'Unity-разработчик', 'Художники']
    .forEach((t, i) => box(t, K[i], L[3], 0.96, i === 2 || i === 3, 0.52));
  s.addShape(pres.shapes.RECTANGLE, { x: 0.4, y: 3.42, w: 0.26, h: 0.15, fill: { color: C.white }, line: { color: C.accent, width: 2 } });
  text(s, '— сотрудники, обеспечивающие публикацию и сопровождение игр', { x: 0.75, y: 3.36, w: 5.8, h: 0.27, fontSize: 11.5 });
  text(s, [
    { text: 'Welwise Games', options: { bold: true } }, { text: ' — платформа веб-игр, запущена в 2025 г.: ≈60 игр, из них ≈15 онлайн-игр с игровыми серверами, ≈20 сторонних разработчиков.', options: { breakLine: true } },
    { text: 'Бизнес-модель: ', options: { bold: true } }, { text: 'доход от рекламы и внутриигровых покупок делится с авторами игр.', options: { breakLine: true } },
    { text: 'Сопровождение: ', options: { bold: true } }, { text: 'инженер по эксплуатации, два модератора и менеджер.' },
  ], { x: 0.4, y: 3.75, w: 6.15, h: 1.5, fontSize: 13, valign: 'top', paraSpaceAfter: 5 });
  const ih = 4.3, iw = ih * 1440 / 2370;
  s.addImage({ path: IMG('site-tall.jpg'), x: W - 0.4 - iw, y: 0.98, w: iw, h: ih });
  s.addShape(pres.shapes.RECTANGLE, { x: W - 0.4 - iw, y: 0.98, w: iw, h: ih, fill: { type: 'none' }, line: { color: C.black, width: 0.75 } });
  num(s);
  speechNotes(s);
}

// ── 4. Контекстная диаграмма ───────────────────────────────
{
  const s = slideTitle('Контекстная диаграмма');
  const bx = 3.45, by = 2.25, bw = 3.1, bh = 1.45;
  const arrow = (x1, y1, x2, y2) => seg(s, x1, y1, x2, y2, { width: 1.25, endArrowType: 'triangle' });
  s.addShape(pres.shapes.RECTANGLE, { x: bx, y: by, w: bw, h: bh, fill: { color: C.white }, line: { color: C.black, width: 1.5 } });
  text(s, 'Публикация и сопровождение веб-игр', { x: bx + 0.1, y: by + 0.1, w: bw - 0.2, h: bh - 0.4, fontSize: 17, bold: true, align: 'center', valign: 'middle' });
  text(s, 'A0', { x: bx + bw - 0.5, y: by + bh - 0.3, w: 0.4, h: 0.25, fontSize: 12, align: 'right' });
  ['Сборка веб-игры', 'Описание и промо-материалы', 'Сборка игрового сервера', 'Команды управления сервером'].forEach((t, i) => {
    const y = by + 0.22 + i * 0.34;
    arrow(0.4, y, bx, y);
    text(s, t, { x: 0.4, y: y - 0.27, w: bx - 0.5, h: 0.25, fontSize: 12.5 });
  });
  ['Опубликованная игра', 'Доступный игровой сервер', 'Журнал игрового сервера', 'Ежемесячный отчёт'].forEach((t, i) => {
    const y = by + 0.22 + i * 0.34;
    arrow(bx + bw, y, 9.6, y);
    text(s, t, { x: bx + bw + 0.1, y: y - 0.27, w: 3, h: 0.25, fontSize: 12.5 });
  });
  ['Требования площадки к играм', 'Руководство разработчика', 'Практики непрерывной интеграции'].forEach((t, i) => {
    const x = bx + 0.45 + i * 1.1;
    arrow(x, 1.15 + i * 0.3, x, by);
    text(s, t, { x: x + 0.06, y: 1.05 + i * 0.3, w: 3.3, h: 0.25, fontSize: 12.5 });
  });
  ['Разработчик игры', 'Модераторы', 'Инженер, менеджер', 'Telegram, GitHub Actions, программа отчётов'].forEach((t, i) => {
    const x = bx + 0.35 + i * 0.8;
    arrow(x, 5.15 - i * 0.28, x, by + bh);
    text(s, t, { x: x + 0.06, y: 4.98 - i * 0.28, w: 3.6, h: 0.25, fontSize: 12.5 });
  });
  num(s);
  speechNotes(s);
}

// ── BPMN: схемы фигурами ───────────────────────────────────
// Дорожки идут сверху вниз от y0; узлы задаются центром по x (дюймы) и дорожкой.
// Именование по практике BPMN: задачи — «глагол + объект», начальные события — что произошло,
// конечные — итоговое состояние, таймеры — по расписанию.
function bpmn(s, lanes, opt = {}) {
  const x0 = 0.3, x1 = 9.7, hw = opt.hw || 0.42, tw = opt.tw || 1.05, th = opt.th || 0.56, fs = opt.fs || 10;
  let y = opt.y0 || 1.08;
  const lane = {};
  lanes.forEach(([id, label, h]) => {
    lane[id] = { y, h, cy: y + h / 2 };
    s.addShape(pres.shapes.RECTANGLE, { x: x0, y, w: x1 - x0, h, fill: { type: 'none' }, line: { color: C.black, width: 0.75 } });
    s.addShape(pres.shapes.LINE, { x: x0 + hw, y, w: 0, h, line: { color: C.black, width: 0.75 } });
    // Кегль подписи дорожки подбирается так, чтобы самое длинное слово не переносилось
    const lfs = Math.max(8, Math.min(11, (h - 0.1) / (Math.max(...label.split(' ').map((w) => w.length)) * 0.0074)));
    text(s, label, { x: x0 + hw / 2 - h / 2, y: y + h / 2 - hw / 2, w: h, h: hw, fontSize: lfs, align: 'center', valign: 'middle', rotate: 270 });
    y += h;
  });
  const N = {};
  const put = (id, cx, l, w, h, dy = 0) => (N[id] = { cx, cy: lane[l].cy + dy, w, h, x: cx - w / 2, y: lane[l].cy + dy - h / 2 });
  const lbl = (t, x, y2, w, align = 'center', size = 9) => text(s, t, { x, y: y2, w, h: 0.3, fontSize: size, align, valign: 'top' });
  const thin = { width: 0.75 };
  const api = {
    lane, N,
    task(id, cx, l, t, o = {}) {
      const n = put(id, cx, l, o.w || tw, o.h || th, o.dy);
      s.addShape(pres.shapes.ROUNDED_RECTANGLE, { x: n.x, y: n.y, w: n.w, h: n.h, rectRadius: 0.06, fill: { color: C.white }, line: { color: C.black, width: 1 } });
      // Значок типа задачи: пользовательская (человек) или сервисная (система)
      const ic = o.kind ? 0.12 : 0;
      if (ic) s.addImage({ path: IMG(o.kind === 'service' ? 'icon-service.png' : 'icon-user.png'), x: n.x + 0.04, y: n.y + 0.035, w: ic, h: ic });
      if (o.collapsed) { // свёрнутый подпроцесс: квадрат с плюсом у нижнего края
        const q = 0.13, qx = n.x + 0.08, qy = n.y + n.h - q - 0.03;
        s.addShape(pres.shapes.RECTANGLE, { x: qx, y: qy, w: q, h: q, fill: { color: C.white }, line: { color: C.black, width: 0.75 } });
        seg(s, qx + 0.025, qy + q / 2, qx + q - 0.025, qy + q / 2, thin); seg(s, qx + q / 2, qy + 0.025, qx + q / 2, qy + q - 0.025, thin);
      }
      const pt = ic ? 0.13 : 0, pb = o.collapsed ? 0.2 : 0;
      text(s, t, { x: n.x + 0.04, y: n.y + pt, w: n.w - 0.08, h: n.h - pt - pb, fontSize: o.fs || fs, align: 'center', valign: 'middle', bold: !!o.collapsed });
    },
    // kind: start | end | timer | msg | cond; o.inter — промежуточное (двойной круг), o.dashed — непрерывающее
    event(id, cx, l, kind, t, o = {}) {
      const d = o.d || 0.28, n = put(id, cx, l, d, d, o.dy);
      const line = { color: C.black, width: kind === 'end' ? 2.5 : 1, dashType: o.dashed ? 'dash' : 'solid' };
      s.addShape(pres.shapes.OVAL, { x: n.x, y: n.y, w: d, h: d, fill: { color: C.white }, line });
      if (o.inter) s.addShape(pres.shapes.OVAL, { x: n.x + 0.035, y: n.y + 0.035, w: d - 0.07, h: d - 0.07, fill: { type: 'none' }, line: { ...line, width: 0.75 } });
      const gx = n.cx, gy = n.cy;
      if (kind === 'msg') {
        s.addShape(pres.shapes.RECTANGLE, { x: gx - 0.06, y: gy - 0.04, w: 0.12, h: 0.08, fill: { type: 'none' }, line: { color: C.black, width: 0.75 } });
        seg(s, gx - 0.06, gy - 0.04, gx, gy + 0.005, thin); seg(s, gx, gy + 0.005, gx + 0.06, gy - 0.04, thin);
      } else if (kind === 'timer') {
        s.addShape(pres.shapes.OVAL, { x: gx - 0.075, y: gy - 0.075, w: 0.15, h: 0.15, fill: { type: 'none' }, line: { color: C.black, width: 0.75 } });
        seg(s, gx, gy, gx, gy - 0.055, thin); seg(s, gx, gy, gx + 0.04, gy, thin);
      } else if (kind === 'cond') {
        s.addShape(pres.shapes.RECTANGLE, { x: gx - 0.045, y: gy - 0.06, w: 0.09, h: 0.12, fill: { type: 'none' }, line: { color: C.black, width: 0.75 } });
        [-0.03, 0, 0.03].forEach((k) => seg(s, gx - 0.028, gy + k, gx + 0.028, gy + k, { width: 0.5 }));
      }
      if (t) {
        const lw = o.lw || 0.7;
        lbl(t, Math.max(x0 + hw + 0.02, n.cx - lw / 2), n.y + d + 0.02, lw);
      }
    },
    gw(id, cx, l, t, o = {}) {
      const d = 0.36, n = put(id, cx, l, d, d, o.dy);
      s.addShape(pres.shapes.DIAMOND, { x: n.x, y: n.y, w: d, h: d, fill: { color: C.white }, line: { color: C.black, width: 1 } });
      text(s, '×', { x: n.x, y: n.y - 0.02, w: d, h: d, fontSize: 20, bold: true, align: 'center', valign: 'middle' });
      if (t) {
        const p = o.lpos || 'top', lw = o.lw || 1.2;
        if (p === 'top') lbl(t, n.cx - lw / 2, n.y - (o.lh || 0.2), lw);
        else if (p === 'bottom') lbl(t, n.cx - lw / 2, n.y + d + 0.02, lw);
        else if (p === 'left') lbl(t, n.x - lw - 0.05, n.cy - 0.1, lw, 'right');
        else lbl(t, n.x + d + 0.05, n.cy - 0.24, lw, 'left');
      }
    },
    // Порт узла: сторона «t/b/l/r» и сдвиг вдоль стороны
    port(id, side, d = 0) {
      const n = N[id];
      return { t: [n.cx + d, n.y], b: [n.cx + d, n.y + n.h], l: [n.x, n.cy + d], r: [n.x + n.w, n.cy + d] }[side];
    },
    // Поток: from/to — «сторона» или «сторона:сдвиг»; via — промежуточные точки; без via — один излом
    flow(a, b, o = {}) {
      const ps = (v, def) => { const [sd, d] = (v || def).split(':'); return [sd, +(d || 0)]; };
      const nA = N[a], nB = N[b];
      const def = Math.abs(nA.cy - nB.cy) < 0.05 ? ['r', 'l'] : Math.abs(nA.cx - nB.cx) < 0.05 ? (nB.cy > nA.cy ? ['b', 't'] : ['t', 'b']) : ['r', nB.cy > nA.cy ? 't' : 'b'];
      const [fs1, fd] = ps(o.from, def[0]), [ts, td] = ps(o.to, def[1]);
      const p0 = api.port(a, fs1, fd), p1 = api.port(b, ts, td);
      let pts = [p0, ...(o.via || []), p1];
      if (!o.via && p0[0] !== p1[0] && p0[1] !== p1[1]) {
        const hor = (sd) => sd === 'l' || sd === 'r';
        if (hor(fs1) && !hor(ts)) pts = [p0, [p1[0], p0[1]], p1];
        else if (!hor(fs1) && hor(ts)) pts = [p0, [p0[0], p1[1]], p1];
        else if (hor(fs1)) { const mx = (p0[0] + p1[0]) / 2; pts = [p0, [mx, p0[1]], [mx, p1[1]], p1]; }
        else { const my = (p0[1] + p1[1]) / 2; pts = [p0, [p0[0], my], [p1[0], my], p1]; }
      }
      for (let i = 0; i < pts.length - 1; i++) {
        seg(s, pts[i][0], pts[i][1], pts[i + 1][0], pts[i + 1][1], i === pts.length - 2 ? { endArrowType: 'triangle' } : {});
      }
      if (o.label) {
        const [lx, ly] = o.lp || [pts[0][0] + 0.05, pts[0][1]];
        text(s, o.label, { x: lx, y: ly - 0.2, w: 0.5, h: 0.2, fontSize: 9, valign: 'bottom' });
      }
    },
    // Подсветка по щелчку: пунктирные рамки вокруг групп узлов, номер у первой рамки,
    // пояснение с тем же номером — в свободном месте схемы (note: [x, y, w, h])
    highlight(items, color) {
      items.forEach(({ ids, text: t, note }, i) => {
        const g = `a${i + 1}`;
        const groups = Array.isArray(ids[0]) ? ids : [ids];
        groups.forEach((grp, k) => {
          const bx = Math.min(...grp.map((q) => N[q].x)) - 0.06, by = Math.min(...grp.map((q) => N[q].y)) - 0.06;
          const bw = Math.max(...grp.map((q) => N[q].x + N[q].w)) + 0.06 - bx, bh = Math.max(...grp.map((q) => N[q].y + N[q].h)) + 0.06 - by;
          s.addShape(pres.shapes.ROUNDED_RECTANGLE, { x: bx, y: by, w: bw, h: bh, rectRadius: 0.06, fill: { type: 'none' }, line: { color, width: 2, dashType: 'dash' }, objectName: `${g}_hl${k}` });
          marker(s, i + 1, bx + bw - 0.15, by - 0.15, color, `${g}_m${k}`, 0.26);
        });
        const [nx, ny, nw, nh] = note;
        marker(s, i + 1, nx, ny + 0.02, color, `${g}_lm`, 0.24);
        text(s, t, { x: nx + 0.3, y: ny, w: nw - 0.3, h: nh, fontSize: 10, valign: 'top', objectName: `${g}_lt` });
      });
    },
  };
  return api;
}

// ── 5. «Как есть»: публикация ──────────────────────────────
{
  const s = slideTitle('Модель процесса «Как есть»', 'Публикация и обновление игры');
  const b = bpmn(s, [['dev', 'Разработчик', 1.0], ['eng', 'Инженер по эксплуатации', 0.84], ['ci', 'GitHub Actions', 0.72], ['mod', 'Модератор', 0.8], ['mgr', 'Менеджер', 0.8]],
    { tw: 0.95, th: 0.56, fs: 9.5 });
  const U = { kind: 'user' }, SV = { kind: 'service' }, D = { kind: 'user', dy: 0.1 };
  b.event('s', 1.05, 'dev', 'start', 'Игра готова к публикации', { dy: 0.1, lw: 0.8 });
  b.task('t1', 2.05, 'dev', 'Запросить размещение в Telegram', D);
  b.task('t2', 2.05, 'eng', 'Создать репозиторий и конвейер', U);
  b.task('t3', 3.1, 'dev', 'Загрузить сборку в репозиторий', D);
  b.task('t4', 3.1, 'ci', 'Развернуть в тестовом окружении', SV);
  b.task('t5', 4.2, 'dev', 'Передать ссылку и материалы', D);
  b.task('t6', 4.2, 'mod', 'Проверить игру', { kind: 'user', dy: -0.12 });
  b.gw('g', 5.3, 'mod', 'Соответствует требованиям?', { lpos: 'bottom', lw: 1.05, dy: -0.12 });
  b.task('t7', 5.3, 'dev', 'Доработать по замечаниям', D);
  b.task('t8', 6.55, 'mod', 'Сообщить о готовности', { kind: 'user', dy: -0.12 });
  b.task('t9', 6.55, 'eng', 'Вручную запустить выпуск', U);
  b.task('t10', 7.5, 'ci', 'Развернуть в продуктивном окружении', SV);
  b.event('e1', 8.27, 'ci', 'end', 'Игра опубликована', { lw: 0.84 });
  // Отдельный цикл: ежемесячный отчёт
  b.event('tm', 6.85, 'mgr', 'timer', 'Конец месяца', { lw: 0.95 });
  b.task('t11', 7.8, 'mgr', 'Собрать отчёт в настольной программе', { kind: 'user', h: 0.66 });
  b.task('t12', 8.8, 'mgr', 'Отправить отчёт в Telegram', { kind: 'user', w: 0.88, h: 0.66 });
  b.task('t13', 8.8, 'dev', 'Изучить отчёт', { kind: 'user', dy: 0.1, w: 0.88 });
  b.event('e2', 9.5, 'dev', 'end', '', { dy: 0.1 });
  text(s, 'Отчёт изучен', { x: 9.27, y: b.N.e2.y - 0.36, w: 0.43, h: 0.34, fontSize: 9, align: 'center', valign: 'bottom' });
  const top = b.lane.dev.y + 0.13;
  b.flow('s', 't1'); b.flow('t1', 't2'); b.flow('t2', 't3', { from: 'r', to: 'b:-0.25' }); b.flow('t3', 't4', { from: 'b:0.15', to: 't:0.15' });
  b.flow('t4', 't5', { from: 'r', to: 'b:-0.25' }); b.flow('t5', 't6', { from: 'b:0.15', to: 't:0.15' }); b.flow('t6', 'g');
  b.flow('g', 't7', { label: 'Нет', lp: [5.35, b.lane.ci.y + 0.25] });
  b.flow('t7', 't3', { from: 't', to: 't', via: [[5.3, top], [3.1, top]] });
  b.flow('g', 't8', { label: 'Да' }); b.flow('t8', 't9'); b.flow('t9', 't10', { from: 'r', to: 't' }); b.flow('t10', 'e1');
  b.flow('tm', 't11'); b.flow('t11', 't12'); b.flow('t12', 't13');
  b.flow('t13', 'e2');
  const ny = b.lane.mgr.y;
  b.highlight([
    { ids: [['t2'], ['t4']], text: 'CI/CD для каждой игры: сложная поддержка, привязка к Git-хостингу', note: [0.82, ny + 0.06, 2.85, 0.36] },
    { ids: [['t1'], ['t5', 't7']], text: 'Связь через зарубежный мессенджер: нестабильно, история теряется', note: [3.75, ny + 0.06, 2.75, 0.36] },
    { ids: ['t9'], text: 'Выпуск игры зависит от одного инженера', note: [0.82, ny + 0.44, 2.85, 0.32] },
    { ids: [['tm', 't11', 't12']], text: 'Статистика — раз в месяц, отчёт собирается вручную', note: [3.75, ny + 0.44, 2.75, 0.36] },
  ], C.red);
  num(s);
  speechNotes(s);
}

// ── 6. «Как есть»: серверы ─────────────────────────────────
{
  const s = slideTitle('Модель процесса «Как есть»', 'Управление игровыми серверами');
  const b = bpmn(s, [['dev', 'Разработчик', 1.45], ['eng', 'Инженер по эксплуатации', 1.35], ['ci', 'GitHub Actions', 1.36]], { tw: 1.08, th: 0.72, fs: 9.5 });
  const U = { kind: 'user' }, SV = { kind: 'service' };
  b.event('s', 0.9, 'dev', 'start', 'Нужен игровой сервер', { lw: 0.52 });
  b.task('t1', 1.9, 'dev', 'Запросить сервер в Telegram', U);
  b.task('t2', 1.9, 'eng', 'Создать репозиторий и конвейер', U);
  b.task('t3', 3.1, 'dev', 'Загрузить сборку в репозиторий', U);
  b.task('t4', 3.1, 'ci', 'Доставить сборку на сервер', SV);
  b.task('t5', 4.45, 'dev', 'Запросить запуск в Telegram', U);
  b.task('t6', 4.45, 'eng', 'Вручную запустить экземпляр', U);
  b.task('t7', 5.65, 'dev', 'Проверить сервер без журналов', U);
  b.gw('g', 6.7, 'dev', 'Нужно действие?', { lpos: 'bottom', lw: 0.85 });
  b.task('t8', 7.75, 'dev', 'Запросить остановку или перезапуск', U);
  b.task('t9', 7.75, 'eng', 'Вручную остановить или перезапустить', U);
  b.event('e', 9.3, 'dev', 'end', 'Сервер работает', { lw: 0.62 });
  const top = b.lane.dev.y + 0.14;
  b.flow('s', 't1'); b.flow('t1', 't2'); b.flow('t2', 't3', { from: 'r', to: 'b:-0.25' }); b.flow('t3', 't4', { from: 'b:0.15', to: 't:0.15' });
  b.flow('t4', 't5', { from: 'r', to: 'l', via: [[3.77, b.lane.ci.cy], [3.77, b.lane.dev.cy]] });
  b.flow('t5', 't6'); b.flow('t6', 't7', { from: 'r', to: 'b:-0.25' }); b.flow('t7', 'g');
  b.flow('g', 't8', { label: 'Да' }); b.flow('t8', 't9');
  b.flow('t9', 't7', { from: 'l', to: 'b:0.25' });
  b.flow('g', 'e', { from: 't', to: 't', via: [[6.7, top], [9.3, top]], label: 'Нет', lp: [6.75, top + 0.22] });
  const ny = b.lane.ci.y;
  b.highlight([
    { ids: ['t1', 't2'], text: 'Размещение — запрос в зарубежном мессенджере и ручная настройка конвейера', note: [3.9, ny + 0.2, 2.75, 0.5] },
    { ids: ['t5', 't6'], text: 'Запуск экземпляра — только через инженера', note: [6.85, ny + 0.2, 2.75, 0.5] },
    { ids: ['t7'], text: 'Журналы сервера разработчику недоступны', note: [3.9, ny + 0.75, 2.75, 0.5] },
    { ids: ['t8', 't9'], text: 'Любое действие с сервером — снова через инженера', note: [6.85, ny + 0.75, 2.75, 0.5] },
  ], C.red);
  num(s);
  speechNotes(s);
}

// ── 7. «Как будет»: публикация ─────────────────────────────
{
  const s = slideTitle('Модель процесса «Как будет»', 'Публикация и обновление игры');
  const b = bpmn(s, [['dev', 'Разработчик', 1.3], ['sys', 'Система', 1.45], ['mod', 'Модератор', 1.4]], { tw: 1.0, th: 0.66, fs: 9.5 });
  const up = { kind: 'user', dy: -0.12 }, M = -0.3, R2 = 0.35;
  const SV = (dy = M) => ({ kind: 'service', dy });
  b.event('s', 0.9, 'dev', 'start', 'Игра готова к публикации', { dy: -0.12, lw: 0.7 });
  b.task('t1', 2.0, 'dev', 'Заполнить черновик игры', up);
  b.task('t4', 2.0, 'sys', 'Подготовить сборку к проверке', SV());
  b.task('t5', 3.15, 'dev', 'Проверить игру в песочнице', up);
  b.task('t6', 4.25, 'dev', 'Отправить на модерацию', up);
  b.task('t7', 4.25, 'sys', 'Сохранить аудит-снимок, поставить в очередь', { kind: 'service', dy: -0.27, h: 0.7, w: 1.12 });
  b.task('t8', 5.3, 'mod', 'Проверить игру по аудит-снимку', { kind: 'user', dy: -0.2 });
  b.gw('g', 6.3, 'mod', 'Соответствует требованиям?', { lpos: 'bottom', lw: 1.4, dy: -0.2 });
  b.task('t9', 6.3, 'dev', 'Доработать по замечаниям из чата', up);
  b.task('t10', 7.25, 'sys', 'Опубликовать версию в каталоге', SV());
  b.event('e1', 7.95, 'sys', 'end', 'Игра опубликована', { dy: M, lw: 0.9 });
  // Отдельный цикл: ежедневная статистика
  b.event('tm', 7.95, 'sys', 'timer', 'Ежедневно', { dy: R2, lw: 0.7 });
  b.task('t11', 8.95, 'sys', 'Обновить статистику', { kind: 'service', dy: R2, w: 0.9, h: 0.5 });
  b.task('t12', 8.95, 'dev', 'Просмотреть статистику', { kind: 'user', dy: -0.12, w: 0.9 });
  b.event('e2', 9.5, 'dev', 'end', '', { dy: 0.38 });
  text(s, 'Статистика изучена', { x: 8.2, y: b.N.e2.y + 0.03, w: 1.12, h: 0.22, fontSize: 9, align: 'right' });
  const top = b.lane.dev.y + 0.12;
  b.flow('s', 't1'); b.flow('t1', 't4'); b.flow('t4', 't5', { from: 'r', to: 'b' }); b.flow('t5', 't6'); b.flow('t6', 't7');
  b.flow('t7', 't8', { from: 'r', to: 't' }); b.flow('t8', 'g');
  b.flow('g', 't9', { label: 'Нет', lp: [6.35, b.lane.sys.y + 0.3] });
  b.flow('t9', 't1', { from: 't', to: 't', via: [[6.3, top], [2.0, top]] });
  b.flow('g', 't10', { from: 'r', to: 'b', label: 'Да', lp: [6.52, b.N.g.cy] }); b.flow('t10', 'e1');
  b.flow('tm', 't11'); b.flow('t11', 't12');
  b.flow('t12', 'e2', { from: 'r', to: 't', via: [[9.5, b.N.t12.cy]] });
  const sy = b.lane.sys.y, my = b.lane.mod.y;
  b.highlight([
    { ids: ['t1'], text: 'Черновик ведёт команда с разграничением прав', note: [0.8, sy + 0.92, 2.35, 0.45] },
    { ids: ['t5'], text: 'Проверка в песочнице, без инженера', note: [3.3, sy + 0.92, 1.8, 0.45] },
    { ids: [['t7'], ['t8']], text: 'Заявка с аудит-снимком попадает в очередь модерации', note: [0.8, my + 0.55, 3.6, 0.4] },
    { ids: ['t9'], text: 'Замечания и вердикт — в сквозном чате проекта', note: [6.95, my + 0.12, 2.65, 0.45] },
    { ids: [['t10'], ['t11']], text: 'Публикация без инженера, статистика — ежедневно', note: [6.95, my + 0.62, 2.65, 0.45] },
  ], C.green);
  num(s);
  speechNotes(s);
}

// ── 8. «Как будет»: серверы ────────────────────────────────
// Обслуживание серверов — развёрнутый подпроцесс на всю ширину дорожки системы; граничные события
// стоят на его нижней кромке строго над своими обработчиками.
{
  const s = slideTitle('Модель процесса «Как будет»', 'Управление игровыми серверами');
  const b = bpmn(s, [['dev', 'Разработчик', 1.27], ['mod', 'Модератор', 0.6], ['sys', 'Система', 2.28]], { tw: 1.0, th: 0.56, fs: 9 });
  const r0 = -0.25, r1 = 0.33;
  const U = (dy = r0) => ({ kind: 'user', dy });
  b.event('s', 0.9, 'dev', 'start', 'Нужен игровой сервер', { dy: r0, lw: 0.62 });
  b.task('t1', 1.8, 'dev', 'Загрузить серверную сборку', U());
  b.gw('g1', 2.65, 'dev', 'Свой сервер?', { dy: r0 });
  b.task('t2', 3.55, 'dev', 'Подключить свой узел', U());
  b.task('t3', 3.55, 'dev', 'Запросить серверы платформы', U(r1));
  b.task('t4', 4.6, 'mod', 'Одобрить доступ', { kind: 'user' });
  b.gw('g2', 5.45, 'dev', '', { dy: r0 });
  b.task('t5', 6.3, 'dev', 'Настроить политику оркестрации', U());
  b.flow('s', 't1'); b.flow('t1', 'g1');
  b.flow('g1', 't2', { label: 'Да', lp: [2.86, b.N.g1.cy + 0.2] });
  b.flow('g1', 't3', { from: 'b', to: 'l', label: 'Нет', lp: [2.7, b.N.t3.cy - 0.05] });
  b.flow('t3', 't4', { from: 'r', to: 't' }); b.flow('t2', 'g2');
  b.flow('t4', 'g2', { from: 'r', to: 'b' }); b.flow('g2', 't5');

  // Развёрнутый подпроцесс «Обслуживание игровых серверов»
  const sy = b.lane.sys.y, scy = b.lane.sys.cy;
  const F = { x: 0.85, y: sy + 0.08, w: 8.7, h: 1.32 };
  s.addShape(pres.shapes.ROUNDED_RECTANGLE, { x: F.x, y: F.y, w: F.w, h: F.h, rectRadius: 0.08, fill: { color: C.white }, line: { color: C.black, width: 1.25 } });
  text(s, 'Обслуживание серверов', { x: F.x + 0.12, y: F.y + 0.04, w: 3, h: 0.22, fontSize: 10, bold: true });
  b.N.F = { cx: 6.3, cy: F.y + F.h / 2, w: F.w, h: F.h, x: F.x, y: F.y };
  b.flow('t5', 'F', { from: 'b', to: 't:0' });
  const ra = F.y + 0.5 - scy, rb = F.y + 1.0 - scy, SV = (dy, w = 1.25) => ({ kind: 'service', dy, w, h: 0.44 });
  b.event('si', 1.2, 'sys', 'start', '', { dy: ra, d: 0.24 });
  b.task('i1', 2.3, 'sys', 'Запустить экземпляры по политике', SV(ra));
  b.event('im', 3.4, 'sys', 'msg', 'Запрос игрока', { dy: ra, inter: true, d: 0.26, lw: 0.9 });
  b.gw('ig', 4.3, 'sys', 'Есть свободное место?', { dy: ra, lpos: 'top', lw: 1.5, lh: 0.19 });
  b.task('i3', 5.7, 'sys', 'Выдать адрес экземпляра', SV(ra));
  b.event('ie', 7.05, 'sys', 'end', 'Игрок подключён', { dy: ra, d: 0.24, lw: 0.95 });
  b.task('i4', 5.7, 'sys', 'Запустить новый экземпляр или поставить в очередь', SV(rb, 1.6));
  b.flow('si', 'i1'); b.flow('i1', 'im'); b.flow('im', 'ig');
  b.flow('ig', 'i3', { label: 'Да', lp: [4.52, b.N.ig.cy] }); b.flow('i3', 'ie');
  b.flow('ig', 'i4', { from: 'b', to: 'l', label: 'Нет', lp: [4.35, b.N.i4.cy - 0.04] });
  b.flow('i4', 'i3', { from: 't', to: 'b' });

  // Граничные события на нижней кромке и обработчики прямо под ними
  const fb = F.y + F.h, hy = sy + 1.82 - scy;
  [['cond', 'Сбой экземпляра', 'Перезапустить экземпляр', true],
   ['timer', 'Простой экземпляра', 'Остановить экземпляр', true],
   ['msg', 'Команда разработчика', 'Выполнить команду, показать журнал', true],
   ['msg', 'Игра снята с публикации', 'Остановить все экземпляры', false],
  ].forEach(([kind, ev, task, nonint], i) => {
    const x = [1.7, 3.85, 6.75, 8.6][i], id = `h${i + 1}`;
    b.event(`be${i + 1}`, x, 'sys', kind, '', { dy: fb - scy, inter: true, dashed: nonint, d: 0.26 });
    b.task(id, x, 'sys', task, { kind: 'service', dy: hy, w: 1.3, h: 0.44 });
    b.event(id + 'e', x + 0.95, 'sys', 'end', '', { dy: hy, d: 0.24 });
    b.flow(`be${i + 1}`, id); b.flow(id, id + 'e');
    text(s, ev, { x: x - 0.65, y: sy + 2.06, w: 1.7, h: 0.2, fontSize: 9, italic: true, align: 'left' });
  });

  const dy0 = b.lane.dev.y, my = b.lane.mod.y;
  b.highlight([
    { ids: ['t2'], text: 'Новое: собственные узлы разработчика', note: [7.0, dy0 + 0.12, 2.6, 0.4] },
    { ids: ['i1'], text: 'Запуск по политике, без инженера', note: [7.75, F.y + 0.3, 1.75, 0.42] },
    { ids: ['i4'], text: 'Мест нет — новый экземпляр или очередь', note: [7.75, F.y + 0.76, 1.75, 0.45] },
    { ids: ['h1'], text: 'Сбой — автоматический перезапуск', note: [6.9, my + 0.04, 2.7, 0.26] },
    { ids: ['h3'], text: 'Команды и журналы — в интерфейсе', note: [6.9, my + 0.31, 2.7, 0.26] },
  ], C.green);
  num(s);
  speechNotes(s);
}

// ── 9. Сравнительная таблица аналогов ──────────────────────
{
  const s = slideTitle('Сравнительная таблица аналогов');
  const head = ['Критерий', 'Game\nDistribution', 'Crazy Games', 'Poki for Developers', 'itch.io', 'Cloud Arcade', 'Разрабатываемая система'];
  const HR = 0.62, RH = 0.46;
  const rows = [
    ['Самостоятельная публикация', '+', '+', '–', '+', '–', '+'],
    ['Многоэтапная модерация', '–', '+', '+', '–', '–', '+'],
    ['Сквозной чат с модератором', '–', '–', '–', '–', '–', '+'],
    ['Показатели монетизации', '+', '+', '+', '–', '–', '+'],
    ['Игровые метрики', '–', '+', '+', '–', '–', '+'],
    ['Управление игровыми серверами', '–', '–', '–', '–', '–', '+'],
    ['Свои серверы разработчика', '–', '–', '–', '±', '+', '+'],
  ];
  const cell = (t, o = {}) => ({ text: t, options: { fontFace: FONT, fontSize: 13.5, align: 'center', valign: 'middle', color: C.black, ...o } });
  const data = [head.map((h, i) => cell(h, { fontSize: 12, bold: true, color: i === 6 ? C.green : C.black, align: i ? 'center' : 'left', fill: { color: i === 6 ? C.greenSoft : C.soft } }))];
  rows.forEach((r) => data.push(r.map((v, i) => i === 0 ? cell(v, { align: 'left' })
    : cell(v, { bold: true, fontSize: 17, color: v === '+' ? C.green : v === '–' ? C.red : C.black, fill: i === 6 ? { color: C.greenSoft } : undefined }))));
  const colW = [3.05, 1.1, 0.85, 1.05, 0.72, 0.88, 1.55];
  s.addTable(data, { x: 0.4, y: 0.95, w: 9.2, colW, rowH: [HR, ...rows.map(() => RH)], border: { type: 'solid', pt: 0.75, color: 'BFBFBF' } });
  // Подсветка возможностей, которых нет ни у одного аналога
  [3, 6].forEach((r, i) => s.addShape(pres.shapes.ROUNDED_RECTANGLE, {
    x: 0.36, y: 0.95 + HR + (r - 1) * RH - 0.02, w: 9.28, h: RH + 0.04, fill: { type: 'none' }, line: { color: C.red, width: 2 }, rectRadius: 0.05, objectName: `a1_hl${i}`,
  }));
  text(s, 'Сквозной чат и управление игровыми серверами не предоставляет ни один аналог', { x: 0.4, y: 4.88, w: 9.2, h: 0.3, fontSize: 15, bold: true, color: C.red, align: 'center', objectName: 'a1_cap' });
  num(s);
  speechNotes(s);
}

// ── 10. Цель и задачи ВКР ──────────────────────────────────
{
  const s = slideTitle('Цель и задачи ВКР');
  const tasks = [
    'Провести анализ предметной области, бизнес-процессов платформы Welwise Games и существующих аналогов.',
    'Построить модели процессов «как есть» и «как будет», выявить проблемы текущего процесса.',
    'Разработать техническое задание на систему.',
    'Спроектировать архитектуру системы и модель данных.',
    'Реализовать подсистемы публикации, модерации, управления игровыми серверами и аналитики.',
    'Провести тестирование системы.',
  ];
  text(s, [
    { text: 'Цель работы', options: { bold: true } },
    { text: ' — повышение эффективности процесса публикации и сопровождения веб‑игр на платформе Welwise Games за счёт внедрения автоматизированной системы, переводящей разработчиков на самообслуживание и исключающей ручные операции сотрудников.', options: { breakLine: true } },
    { text: ' ', options: { breakLine: true, fontSize: 8 } },
    { text: 'Задачи:', options: { bold: true, breakLine: true } },
    ...tasks.map((t, i) => ({ text: t, options: { bullet: { type: 'number', style: 'arabicParenR' }, breakLine: i < tasks.length - 1 } })),
  ], { x: 0.5, y: 0.95, w: 9.0, h: 4.3, fontSize: 16, valign: 'top', align: 'justify', paraSpaceAfter: 3 });
  num(s);
  speechNotes(s);
}

// ── 11–13. Диаграммы вариантов использования ───────────────
function ucSlide(pairs) {
  const s = slideTitle('Диаграмма вариантов использования');
  const maxW = 4.45, maxH = 3.95, top = 0.92;
  let x = 0.4;
  pairs.forEach(([f, cap, w, h]) => {
    let iw = maxW, ih = iw * h / w;
    if (ih > maxH) { ih = maxH; iw = ih * w / h; }
    const cx = x + maxW / 2;
    s.addImage({ path: IMG(f), x: cx - iw / 2, y: top + (maxH - ih) / 2, w: iw, h: ih });
    text(s, cap, { x, y: top + maxH + 0.06, w: maxW, h: 0.3, fontSize: 12, align: 'center' });
    x += maxW + 0.3;
  });
  num(s);
  speechNotes(s);
}
ucSlide([['projects.png', 'Рис. 1 — Проекты и аналитика', 1428, 1408], ['collaboration.png', 'Рис. 2 — Совместная работа', 1990, 1099]]);
ucSlide([['moderation.png', 'Рис. 3 — Модерация', 1466, 1800], ['servers.png', 'Рис. 4 — Игровые серверы', 1796, 1571]]);
ucSlide([['nodes.png', 'Рис. 5 — Узлы и сервисы данных', 1725, 1674], ['accounts.png', 'Рис. 6 — Учётные записи и администрирование', 2055, 1292]]);

// ── 14. Заключение ─────────────────────────────────────────
{
  const s = slideTitle('Заключение');
  const done = [
    'Проанализирована предметная область и деятельность заказчика — игровой платформы Welwise Games.',
    'Построены модели процессов «как есть» для публикации игр и управления серверами, выявлены проблемы.',
    'Разработаны модели процессов «как будет» с использованием автоматизированной системы.',
    'Проведено сравнение с аналогами: GameDistribution, CrazyGames, Poki for Developers, itch.io, CloudArcade.',
    'Сформулированы цель и задачи, разработано техническое задание по ГОСТ 34.602-2020.',
    'Построены диаграммы вариантов использования для четырёх ролей: разработчик, модератор, администратор, клиент игры.',
    'Спроектирована и реализована система из восьми сервисов, программный код покрыт 413 автоматическими тестами.',
  ];
  text(s, [
    { text: 'В ходе работы:', options: { bold: true, breakLine: true } },
    ...done.map((t, i) => ({ text: t, options: { bullet: { type: 'number', style: 'arabicParenR' }, breakLine: i < done.length - 1 } })),
  ], { x: 0.5, y: 0.95, w: 9.0, h: 3.55, fontSize: 15, valign: 'top', align: 'justify', paraSpaceAfter: 2 });
  s.addShape(pres.shapes.RECTANGLE, { x: 0.5, y: 4.6, w: 9.0, h: 0.62, fill: { color: C.white }, line: { color: C.accent, width: 1.5 } });
  text(s, [
    { text: 'Текущий этап — внедрение: ', options: { bold: true, color: C.accent } },
    { text: 'система проходит тестирование на реальном стенде платформы, завершение — в течение месяца.' },
  ], { x: 0.65, y: 4.6, w: 8.7, h: 0.62, fontSize: 15, valign: 'middle' });
  num(s);
  speechNotes(s);
}

pres.writeFile({ fileName: 'deck.pptx' }).then(() => console.log('written deck.pptx'));
