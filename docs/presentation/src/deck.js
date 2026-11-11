// Презентация к защите ВКР: «Автоматизированная система управления процессом публикации и сопровождения веб-игр».
// Объекты с именем «aN_…» появляются по щелчку (группа N) — анимацию добавляет postprocess.py.
// Схемы BPMN на слайдах — упрощённые версии схем из docs/business-process/bpmn, нарисованные фигурами,
// чтобы текст оставался крупным; полные схемы идут в приложение записки.
const pptxgen = require('pptxgenjs');
const fs = require('fs');
const SPEECH = require('./speech.js');
const HIDDEN = require('./speech_hidden.js');
const { paragraphs } = require('./speech_text.js');

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
// Номер слайда — в правом верхнем углу, крупно: комиссия называет слайд по номеру
const num = (s) => { s.slideNumber = { x: 9.2, y: 0.2, w: 0.55, h: 0.5, fontFace: FONT, fontSize: 20, color: C.black, align: 'right', valign: 'middle' }; };
// Отрезок: горизонтальный или вертикальный, со стрелкой на конце по желанию
const seg = (s, x1, y1, x2, y2, o = {}) => s.addShape(pres.shapes.LINE, {
  x: Math.min(x1, x2), y: Math.min(y1, y2), w: Math.abs(x2 - x1) || 0.001, h: Math.abs(y2 - y1) || 0.001,
  flipH: x2 < x1, flipV: y2 < y1, line: { color: C.black, width: 1, ...o },
});

// Заметки докладчика — текст речи из speech.js, по порядку слайдов
let speechIndex = 0;
function speechNotes(s) {
  const sp = SPEECH[speechIndex++];
  s.addNotes(paragraphs(sp).join('\n\n'));
}

// Заметки скрытых слайдов: вопрос комиссии и короткий ответ
let hiddenIndex = 0;
function hiddenNotes(s) {
  const h = HIDDEN[hiddenIndex++];
  s.addNotes([`Резервный слайд. Если спросят: ${h.ask}`, ...h.text].join('\n\n'));
}

function slideTitle(title, sub) {
  const s = pres.addSlide({ masterName: 'PLAIN' });
  // Заголовок уже слайда на ширину номера; длинные заголовки набираются чуть мельче, чтобы не переноситься
  text(s, title.toUpperCase(), { x: 0.7, y: 0.15, w: 8.6, h: 0.6, fontSize: title.length > 31 ? 24 : 26, bold: true, align: 'center', valign: 'middle' });
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
    { text: 'Морозов Александр Васильевич' },
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
  const uml = !!opt.uml; // диаграмма деятельности UML: решение без знака, начальный узел залит, конечный — «мишень»
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
      if (uml && kind === 'end') line.width = 1;
      s.addShape(pres.shapes.OVAL, { x: n.x, y: n.y, w: d, h: d, fill: { color: uml && kind === 'start' ? C.black : C.white }, line });
      if (uml && kind === 'end') s.addShape(pres.shapes.OVAL, { x: n.x + d * 0.22, y: n.y + d * 0.22, w: d * 0.56, h: d * 0.56, fill: { color: C.black }, line: { color: C.black, width: 0.5 } });
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
      if (!uml) text(s, '×', { x: n.x, y: n.y - 0.02, w: d, h: d, fontSize: 20, bold: true, align: 'center', valign: 'middle' });
      if (t) {
        const p = o.lpos || 'top', lw = o.lw || 1.2;
        if (p === 'top') lbl(t, n.cx - lw / 2, n.y - (o.lh || 0.2), lw);
        else if (p === 'bottom') lbl(t, n.cx - lw / 2, n.y + d + 0.02, lw);
        else if (p === 'left') lbl(t, n.x - lw - 0.05, n.cy - 0.1, lw, 'right');
        else lbl(t, n.x + d + 0.05, n.cy - 0.24, lw, 'left');
      }
    },
    // Точка на линии потока: в неё можно привести стрелку (слияние потоков)
    point(id, x, y) { N[id] = { cx: x, cy: y, w: 0, h: 0, x, y }; },
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
  const head = ['Критерий', 'Game\nDistribution', 'Crazy Games', 'Poki for Developers', 'itch.io', 'Cloud Arcade'];
  const HR = 0.58, RH = 0.42;
  const rows = [
    ['Самостоятельная публикация', '+', '+', '–', '+', '–'],
    ['Многоэтапная модерация', '–', '+', '+', '–', '–'],
    ['Сквозной чат с модератором', '–', '–', '–', '–', '–'],
    ['Показатели монетизации', '+', '+', '+', '–', '–'],
    ['Игровые метрики', '–', '+', '+', '–', '–'],
    ['Управление игровыми серверами', '–', '–', '–', '–', '–'],
    ['Свои серверы разработчика', '–', '–', '–', '±', '+'],
    ['Отечественная поддержка', '–', '–', '–', '–', '–'],
  ];
  const cell = (t, o = {}) => ({ text: t, options: { fontFace: FONT, fontSize: 13.5, align: 'center', valign: 'middle', color: C.black, ...o } });
  const data = [head.map((h, i) => cell(h, { fontSize: 12, bold: true, color: C.black, align: i ? 'center' : 'left', fill: { color: C.soft } }))];
  rows.forEach((r) => data.push(r.map((v, i) => i === 0 ? cell(v, { align: 'left' })
    : cell(v, { bold: true, fontSize: 17, color: v === '+' ? C.green : v === '–' ? C.red : C.black }))));
  const colW = [3.6, 1.25, 1.05, 1.25, 0.95, 1.1];
  s.addTable(data, { x: 0.4, y: 0.95, w: 9.2, colW, rowH: [HR, ...rows.map(() => RH)], border: { type: 'solid', pt: 0.75, color: 'BFBFBF' } });
  // Подсветка возможностей, которых нет ни у одного аналога: по щелчку на каждую, подпись — с последней
  [3, 6, 8].forEach((r, i) => s.addShape(pres.shapes.ROUNDED_RECTANGLE, {
    x: 0.36, y: 0.95 + HR + (r - 1) * RH - 0.02, w: 9.28, h: RH + 0.04, fill: { type: 'none' }, line: { color: C.red, width: 2 }, rectRadius: 0.05, objectName: `a${i + 1}_hl`,
  }));
  text(s, 'Сквозного чата, управления игровыми серверами и отечественной поддержки нет ни у одного аналога', { x: 0.4, y: 4.95, w: 9.2, h: 0.3, fontSize: 14, bold: true, color: C.red, align: 'center', objectName: 'a3_cap' });
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

// ── 11–14. Диаграммы вариантов использования ───────────────
// Первый рисунок — иерархия акторов; дальше функции по контекстам.
function ucSlide(pairs) {
  const s = slideTitle('Диаграмма вариантов использования');
  const maxW = pairs.length === 1 ? 9.2 : 4.45, maxH = 3.9, top = 0.92;
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
ucSlide([['actors.png', 'Рис. 1 — Иерархия акторов', 829, 950], ['accounts.png', 'Рис. 2 — Учётные записи и администрирование', 2055, 1292]]);
ucSlide([['projects.png', 'Рис. 3 — Проекты и аналитика', 1428, 1408], ['collaboration.png', 'Рис. 4 — Совместная работа', 1990, 1099]]);
ucSlide([['moderation.png', 'Рис. 5 — Модерация', 1466, 1800], ['servers.png', 'Рис. 6 — Игровые серверы', 1796, 1571]]);
ucSlide([['nodes.png', 'Рис. 7 — Узлы и сервисы данных', 1725, 1674]]);

// ── Номера рисунков: первые семь — диаграммы вариантов использования, дальше счётчик ──
let figNo = 7;
const fig = (t) => `Рис. ${++figNo} — ${t}`;
// Размер PNG в пикселях — из заголовка файла
const pngSize = (f) => { const b = fs.readFileSync(IMG(f)); return [b.readUInt32BE(16), b.readUInt32BE(20)]; };

// ── Слайды с готовыми диаграммами (PlantUML) ──
// Элемент: [файл, подпись]. opts.uniform — один масштаб для всех картинок (одинаковый кегль, для диаграмм состояний);
// opts.cols — колонки равной ширины (для моделей данных); иначе у картинок общая высота.
// opts.main — слайд основного доклада (заметки из speech.js); иначе резервный (speech_hidden.js).
function figSlide(title, items, opts = {}) {
  const s = slideTitle(title);
  const gap = opts.gap ?? 0.35, maxH = opts.maxH || 3.85, top = opts.top || 0.92, n = items.length, AW = opts.w || 9.3;
  const px = items.map(([f]) => pngSize(f));
  const colW = (AW - gap * (n - 1)) / n;
  let dims;
  if (opts.uniform) {
    const k = Math.min((AW - gap * (n - 1)) / px.reduce((a, [w]) => a + w, 0), maxH / Math.max(...px.map(([, h]) => h)));
    dims = px.map(([w, h]) => ({ w: w * k, h: h * k }));
  } else if (opts.cols) {
    dims = px.map(([w, h]) => { const ih = Math.min(maxH, colW * h / w); return { w: ih * w / h, h: ih }; });
  } else {
    const H = Math.min(maxH, (AW - gap * (n - 1)) / px.reduce((a, [w, h]) => a + w / h, 0));
    dims = px.map(([w, h]) => ({ w: H * w / h, h: H }));
  }
  const boxH = opts.uniform ? Math.max(...dims.map((d) => d.h)) : maxH;
  let x = (W - dims.reduce((a, d) => a + d.w, 0) - gap * (n - 1)) / 2;
  items.forEach(([f, cap], i) => {
    s.addImage({ path: IMG(f), x, y: top + (boxH - dims[i].h) / 2, w: dims[i].w, h: dims[i].h });
    const cw = Math.max(dims[i].w + 0.3, 2.4);
    text(s, fig(cap), { x: x + dims[i].w / 2 - cw / 2, y: top + boxH + 0.06, w: cw, h: 0.42, fontSize: 11, align: 'center', valign: 'top' });
    x += dims[i].w + gap;
  });
  num(s);
  if (opts.main) speechNotes(s); else hiddenNotes(s);
}
const MAIN = { main: true };

// ── Модели данных: ER-диаграммы фигурами ───────────────────
// Сущность — прямоугольник с заголовком и списком атрибутов; первая строка — первичный ключ.
// Связь идёт от строки первичного ключа родителя к строке внешнего ключа потомка: «один» — поперечная черта,
// «много» — «воронья лапка». Пунктир — ссылка без внешнего ключа (между таблицами разных сервисов).
const ER_W = 1.4, ER_GAP = 0.585, ER_X = [0, 1, 2, 3, 4].map((i) => 0.38 + i * (ER_W + ER_GAP));
function erDiagram(s, o = {}) {
  const fsz = o.fs || 9, rowH = o.rowH || 0.15, headH = o.headH || 0.27, E = {};
  const thin = { width: 1 };
  const api = {
    E,
    // rows: строка либо массив фрагментов { text, options }
    entity(id, x, y, w, title, rows) {
      const h = headH + rows.length * rowH + 0.07;
      E[id] = { x, y, w, h, cx: x + w / 2 };
      s.addShape(pres.shapes.RECTANGLE, { x, y, w, h, fill: { color: C.white }, line: { color: C.black, width: 1 } });
      s.addShape(pres.shapes.RECTANGLE, { x, y, w, h: headH, fill: { color: C.soft }, line: { color: C.black, width: 1 } });
      text(s, title, { x, y, w, h: headH, fontSize: o.tfs || fsz + 0.5, bold: true, align: 'center', valign: 'middle' });
      const runs = [];
      rows.forEach((r, i) => {
        const parts = Array.isArray(r) ? r : [{ text: r }];
        parts.forEach((p, k) => runs.push({ text: p.text, options: { ...(p.options || {}), breakLine: k === parts.length - 1 && i < rows.length - 1 } }));
      });
      text(s, runs, { x: x + 0.06, y: y + headH + 0.035, w: w - 0.08, h: rows.length * rowH, fontSize: fsz, valign: 'top', lineSpacing: rowH * 72 });
    },
    // Точка на боковой стороне сущности на уровне строки i
    at(id, i, side) { const e = E[id]; return [side === 'l' ? e.x : e.x + e.w, e.y + headH + 0.035 + (i + 0.5) * rowH]; },
    mark(p, side, kind) {
      const dx = side === 'l' ? -1 : 1;
      if (kind === 'N') {
        seg(s, p[0] + dx * 0.1, p[1], p[0], p[1] - 0.05, thin);
        seg(s, p[0] + dx * 0.1, p[1], p[0], p[1] + 0.05, thin);
      } else seg(s, p[0] + dx * 0.06, p[1] - 0.045, p[0] + dx * 0.06, p[1] + 0.045, thin);
    },
    line(pts, dash) {
      for (let i = 0; i < pts.length - 1; i++) seg(s, pts[i][0], pts[i][1], pts[i + 1][0], pts[i + 1][1], { width: 1, dashType: dash ? 'dash' : 'solid' });
    },
  };
  return api;
}

// Общая расстановка обеих моделей: пять колонок, три ряда, в центре — проект. Линии связей идут по промежуткам
// между колонками и рядами и подходят сбоку к строке ключа; дорожки подобраны так, чтобы линии не пересекались.
// id — имена сущностей по ролям; logical — рисовать ссылки между сервисами пунктиром (даталогическая модель).
function erLayout(er, id, Y, logical) {
  const at = (k, i, side) => er.at(id[k], i, side);
  const G = ER_X.map((x) => x + ER_W); // левые края промежутков между колонками
  const bot = (r) => Math.max(...Object.values(er.E).filter((e) => Math.abs(e.y - Y[r]) < 0.01).map((e) => e.y + e.h));
  const La = (bot(0) + Y[1]) / 2, Lb = (bot(1) + Y[2]) / 2; // горизонтали между рядами
  const m0 = 0.14, tu = G[0] + 0.2, tn = G[0] + 0.4, tl = G[1] + 0.17, ts = G[1] + 0.28, tp = G[1] + 0.39, tr = G[2] + 0.29, gp = G[3] + 0.18, tq = G[3] + 0.4;
  const marked = new Set();
  // parent/child: [роль, строка, сторона]; via — точки между ними; kind — кратность потомка: «N» или «1»
  const R = (parent, child, kind, via, cross) => {
    const p = at(...parent), c = at(...child);
    er.line([p, ...via, c], logical && cross);
    const key = parent.join();
    if (!marked.has(key)) { er.mark(p, parent[2], '1'); marked.add(key); }
    er.mark(c, child[2], kind);
  };
  const y = (k, i) => at(k, i, 'l')[1], P = y('project', 0), U = y('user', 0), Q = y('request', 0);
  R(['project', 0, 'l'], ['draft', 0, 'r'], '1', [[tl, P], [tl, y('draft', 0)]]);
  R(['project', 0, 'l'], ['member', 1, 'r'], 'N', [[tl, P], [tl, y('member', 1)]]);
  R(['project', 0, 'r'], ['build', 1, 'r'], 'N', [[tr, P], [tr, y('build', 1)]]);
  R(['project', 0, 'r'], ['release', 1, 'l'], 'N', [[tr, P], [tr, y('release', 1)]]);
  R(['project', 0, 'r'], ['request', 1, 'l'], 'N', [[tr, P], [tr, y('request', 1)]], true);
  R(['project', 0, 'r'], ['sbuild', 1, 'r'], 'N', [[tr, P], [tr, y('sbuild', 1)]], true);
  R(['project', 0, 'r'], ['policy', 0, 'l'], '1', [[tr, P], [tr, y('policy', 0)]], true);
  R(['project', 0, 'r'], ['grant', 1, 'l'], '1', [[tr, P], [tr, La], [gp, La], [gp, y('grant', 1)]], true);
  R(['project', 0, 'r'], ['message', 2, 'l'], 'N', [[tr, P], [tr, Lb], [gp, Lb], [gp, y('message', 2)]], true);
  R(['request', 0, 'r'], ['snapshot', 1, 'l'], '1', [[tq, Q], [tq, y('snapshot', 1)]]);
  R(['request', 0, 'r'], ['message', 1, 'l'], 'N', [[tq, Q], [tq, y('message', 1)]]);
  R(['user', 0, 'r'], ['member', 2, 'l'], 'N', [[tu, U], [tu, y('member', 2)]], true);
  R(['user', 0, 'r'], ['project', 1, 'l'], 'N', [[tu, U], [tu, Lb], [tp, Lb], [tp, y('project', 1)]], true);
  R(['user', 0, 'l'], ['node', 1, 'l'], 'N', [[m0, U], [m0, y('node', 1)]], true);
  R(['node', 0, 'r'], ['instance', 1, 'l'], 'N', [[tn, y('node', 0)], [tn, y('instance', 1)]]);
  R(['sbuild', 0, 'l'], ['instance', 2, 'r'], 'N', [[ts, y('sbuild', 0)], [ts, y('instance', 2)]]);
}
// Обозначения — в свободной ячейке первого ряда
function erLegend(s, er, logical) {
  text(s, 'Обозначения:', { x: 0.38, y: 0.95, w: 1.4, h: 0.2, fontSize: 9.5, bold: true });
  [['1', 'один'], ['N', 'много']].forEach(([k, t], i) => {
    const yy = 1.3 + i * 0.22;
    seg(s, 0.43, yy, 0.88, yy);
    er.mark([0.88, yy], 'l', k);
    seg(s, 0.88, yy - 0.06, 0.88, yy + 0.06);
    text(s, `— ${t}`, { x: 0.96, y: yy - 0.1, w: 0.85, h: 0.2, fontSize: 9, valign: 'middle' });
  });
  if (!logical) return;
  seg(s, 0.43, 1.74, 0.88, 1.74, { dashType: 'dash' });
  text(s, '— ссылка на таблицу другого сервиса, без внешнего ключа', { x: 0.96, y: 1.62, w: 0.9, h: 0.55, fontSize: 8, valign: 'top' });
}

// ── 15. Инфологическая модель: сущности предметной области, их атрибуты и связи; без типов и имён таблиц ──
{
  const s = slideTitle('Инфологическая модель');
  const er = erDiagram(s, { fs: 8.7, rowH: 0.15 });
  const Y = [0.9, 2.49, 4.08];
  const PK = (t) => `${t} «PK»`, FK = (t) => `${t} «FK»`;
  const grid = [
    [null,
      ['draft', 'Черновик', ['Код проекта «PK, FK»', 'название и описание', 'значок, обложка, видео', 'активная сборка']],
      ['build', 'Клиентская сборка', [PK('Код сборки'), FK('Код проекта'), 'версия', 'файл и размер']],
      ['release', 'Релиз', [PK('Код релиза'), FK('Код проекта'), 'версия', 'адрес игры', 'дата публикации']],
      ['grant', 'Квота платформы', [PK('Код квоты'), FK('Код проекта'), 'предел экземпляров', 'пределы ресурсов']]],
    [['user', 'Пользователь', [PK('Код пользователя'), 'электронная почта', 'отображаемое имя', 'роль', 'состояние']],
      ['member', 'Участник проекта', [PK('Код участника'), FK('Код проекта'), FK('Код пользователя'), 'права в проекте']],
      ['project', 'Проект', [PK('Код проекта'), FK('Код владельца'), 'состояние', 'признак онлайн-игры']],
      ['request', 'Заявка на модерацию', [PK('Код заявки'), FK('Код проекта'), 'вид заявки', 'состояние', 'причина отклонения']],
      ['snapshot', 'Аудит-снимок', [PK('Код снимка'), FK('Код заявки'), 'решение', 'содержимое проекта']]],
    [['node', 'Вычислительный узел', [PK('Код узла'), FK('Код владельца'), 'адрес', 'состояние', 'ресурсы']],
      ['instance', 'Экземпляр сервера', [PK('Код экземпляра'), FK('Код узла'), FK('Код сборки'), 'состояние', 'адрес и порт']],
      ['sbuild', 'Серверная сборка', [PK('Код сборки'), FK('Код проекта'), 'версия', 'образ контейнера', 'предел игроков']],
      ['policy', 'Политика оркестрации', ['Код проекта «PK, FK»', 'режим', 'число экземпляров', 'автоперезапуск', 'при заполнении']],
      ['message', 'Сообщение чата', [PK('Код сообщения'), FK('Код заявки'), FK('Код проекта'), 'отправитель', 'текст']]],
  ];
  const id = {};
  grid.forEach((row, r) => row.forEach((e, k) => { if (e) { er.entity(e[0], ER_X[k], Y[r], ER_W, e[1], e[2]); id[e[0]] = e[0]; } }));
  erLayout(er, id, Y, false);
  erLegend(s, er, false);
  text(s, fig('Инфологическая модель предметной области'), { x: 0.4, y: 5.25, w: 9.2, h: 0.25, fontSize: 11, align: 'center' });
  num(s);
  speechNotes(s);
}

// ── 16. Диаграммы состояний: общий масштаб, чтобы кегль на всех трёх был одинаковым ──
figSlide('Диаграммы состояний', [['state-project.png', 'Состояния проекта'], ['state-request.png', 'Состояния заявки на модерацию'], ['state-node.png', 'Состояния вычислительного узла']],
  { uniform: true, main: true, gap: 0.25, w: 9.4, maxH: 3.95 });

// ── 17. Алгоритм оркестрации: диаграмма деятельности UML, один проход цикла ──
{
  const s = slideTitle('Алгоритм оркестрации');
  const b = bpmn(s, [['mon', 'Контроль узлов', 1.62], ['pol', 'Применение политик', 1.5]], { uml: true, y0: 0.93, tw: 1.0, th: 0.5, fs: 9.5 });
  const yes = '[да]', no = '[нет]';
  // Раздел 1: контроль узлов
  const r1 = -0.24, r2 = 0.42, y1 = b.lane.mon.cy + r1, y2 = b.lane.mon.cy + r2;
  b.event('s', 0.95, 'mon', 'start', '', { dy: r1, d: 0.2 });
  b.task('a0', 1.85, 'mon', 'Опросить состояние узла', { dy: r1, w: 1.05 });
  b.gw('d1', 2.95, 'mon', 'Узел ответил?', { dy: r1, lw: 1.0 });
  b.task('a1', 4.3, 'mon', 'Сверить экземпляры, пропавшие отметить сбойными', { dy: r1, w: 1.55 });
  b.gw('d2', 2.95, 'mon', 'Молчит дольше 60 с?', { dy: r2, lpos: 'left', lw: 1.3 });
  b.task('a2', 4.3, 'mon', 'Узел — «Не в сети», его экземпляры — «Сбой»', { dy: r2, w: 1.55 });
  b.gw('d3', 6.4, 'mon', 'Есть сбойные экземпляры и включён автоперезапуск?', { dy: r1, lw: 1.75, lh: 0.33 });
  b.task('a3', 7.7, 'mon', 'Перезапустить экземпляры', { dy: r1, w: 1.15 });
  b.point('j1', 5.35, y1); b.point('j1b', 5.52, y1); b.point('j2', 8.65, y1);
  b.flow('s', 'a0'); b.flow('a0', 'd1'); b.flow('d1', 'a1', { label: yes }); b.flow('a1', 'd3');
  b.flow('d1', 'd2', { label: no, lp: [3.0, y1 + 0.42] });
  b.flow('d2', 'a2', { label: yes });
  b.flow('a2', 'j1', { from: 'r', to: 'b', via: [[5.35, y2]] });
  b.flow('d2', 'j1b', { from: 'b', to: 'b', via: [[2.95, y2 + 0.3], [5.52, y2 + 0.3]], label: no, lp: [2.58, y2 + 0.4] });
  b.flow('d3', 'a3', { label: yes });
  b.flow('d3', 'j2', { from: 'b', to: 'b', via: [[6.4, y1 + 0.52], [8.65, y1 + 0.52]], label: no, lp: [6.45, y1 + 0.4] });
  // Раздел 2: применение политики игры
  const q = 0.05, yq = b.lane.pol.cy + q, yb = yq + 0.47, top2 = b.lane.pol.y + 0.12;
  const D = [1.5, 3.65, 5.8, 7.95], A = [2.55, 4.7, 6.85, 9.05];
  const dl = ['Экземпляров меньше целевого числа?', 'Экземпляр пуст дольше заданного времени?', 'Все экземпляры заполнены?', 'Есть очередь и свободное место?'];
  const al = ['Запустить недостающие', 'Остановить экземпляр', 'Запустить ещё один экземпляр', 'Закрепить место за игроком очереди'];
  D.forEach((x, i) => {
    b.gw(`p${i}`, x, 'pol', dl[i], { dy: q, lw: 1.1, lh: i < 2 ? 0.47 : 0.33 }); // первые две подписи — в три строки
    b.task(`q${i}`, A[i], 'pol', al[i], { dy: q, w: i === 3 ? 1.0 : 0.95, h: 0.56, fs: 9 });
  });
  b.event('e', 9.05, 'pol', 'end', '', { dy: q + 0.52, d: 0.22 });
  b.flow('a3', 'p0', { from: 'r', to: 'l', via: [[9.5, y1], [9.5, top2], [0.85, top2], [0.85, yq]] });
  D.forEach((x, i) => {
    b.flow(`p${i}`, `q${i}`, { label: yes });
    if (i < 3) {
      const jx = (A[i] + 0.475 + D[i + 1] - 0.18) / 2;
      b.point(`k${i}`, jx, yq);
      b.flow(`q${i}`, `p${i + 1}`);
      b.flow(`p${i}`, `k${i}`, { from: 'b', to: 'b', via: [[x, yb], [jx, yb]], label: no, lp: [x + 0.05, yq + 0.4] });
    }
  });
  b.flow('q3', 'e', { from: 'b', to: 't' });
  b.flow('p3', 'e', { from: 'b', to: 'l', label: no, lp: [D[3] + 0.05, yq + 0.4] });
  const dy = 0.93 + 3.12;
  text(s, fig('Диаграмма деятельности: один проход цикла оркестрации, выполняется каждые 15 секунд'), { x: 0.4, y: dy + 0.04, w: 9.2, h: 0.25, fontSize: 11, align: 'center' });
  const B = (t) => ({ text: t, options: { bold: true } });
  const T = (t, o = {}) => ({ text: t, options: o });
  text(s, [
    T('Политику оркестрации разработчик задаёт для каждой игры:', { bold: true, color: C.accent, breakLine: true }),
    B('Поддержание'), T(' — система держит целевое число работающих экземпляров.', { breakLine: true }),
    B('Масштабирование до нуля'), T(' — экземпляр запускается по запросу игрока и останавливается при простое.', { breakLine: true }),
    B('При заполнении'), T(' — запуск дополнительного экземпляра либо очередь игроков с закреплением места.'),
  ], { x: 0.5, y: dy + 0.34, w: 9.0, h: 1.1, fontSize: 12.5, valign: 'top', paraSpaceAfter: 2 });
  num(s);
  speechNotes(s);
}

// ── 18. Диаграмма последовательности: во всю ширину слайда ──
figSlide('Диаграмма последовательности', [['sequence-publish.png', 'Диаграмма последовательности: публикация игры']], { main: true, w: 9.4, maxH: 4.32, top: 0.86 });

// ── 19. Диаграмма развёртывания UML: устройства, среды выполнения, артефакты, каналы связи ──
// Показывает только физическое размещение; что работает внутри и как взаимодействует — на следующем слайде.
// Полная схема с портами — docs/deployment.
{
  const s = slideTitle('Диаграмма развёртывания');
  // Узел UML — «кубик»: передняя грань и две видимые боковые
  const node3d = (x, y, w, h, stereo, title, o = {}) => {
    const d = 0.08;
    s.addShape(pres.shapes.RECTANGLE, { x, y: y + d, w: w - d, h: h - d, fill: { color: C.white }, line: { color: C.black, width: 1 } });
    [[x, y + d, x + d, y], [x + d, y, x + w, y], [x + w, y, x + w - d, y + d], [x + w, y, x + w, y + h - d], [x + w, y + h - d, x + w - d, y + h]]
      .forEach(([a, b2, c, e]) => seg(s, a, b2, c, e, { width: 1 }));
    text(s, [{ text: `«${stereo}»`, options: { fontSize: 8.5, breakLine: true } }, { text: title, options: { bold: true, fontSize: o.fs || 10.5 } }],
      { x, y: y + d + 0.03, w: w - d, h: 0.36, align: 'center', valign: 'top' });
  };
  // Артефакт UML: прямоугольник со значком документа и стереотипом
  const artifact = (x, y, w, h, title, sub) => {
    s.addShape(pres.shapes.RECTANGLE, { x, y, w, h, fill: { color: C.white }, line: { color: C.black, width: 1 } });
    const ix = x + w - 0.16, iy = y + 0.05;
    s.addShape(pres.shapes.RECTANGLE, { x: ix, y: iy, w: 0.1, h: 0.13, fill: { color: C.white }, line: { color: C.black, width: 0.75 } });
    seg(s, ix + 0.055, iy, ix + 0.1, iy + 0.045, { width: 0.75 });
    text(s, [{ text: '«artifact» ', options: { fontSize: 8 } }, { text: title, options: { bold: true, fontSize: 10, breakLine: !!sub } }, ...(sub ? [{ text: sub, options: { fontSize: 9 } }] : [])],
      { x: x + 0.06, y, w: w - 0.26, h, align: 'center', valign: 'middle' });
  };
  const path = (pts) => pts.slice(1).forEach((p, i) => seg(s, pts[i][0], pts[i][1], p[0], p[1], { width: 1 }));
  const proto = (t, x, y, w, o = {}) => text(s, t, { x, y, w, h: 0.17, fontSize: 9, align: 'center', valign: 'middle', ...o });
  const ENV = 'execution environment';
  // Устройства пользователей и смежные системы
  node3d(0.3, 1.15, 2.45, 1.25, 'device', 'Устройство игрока');
  node3d(0.45, 1.62, 2.1, 0.7, ENV, 'Веб-браузер', { fs: 10 });
  node3d(0.3, 2.58, 2.45, 1.25, 'device', 'Рабочее место пользователя');
  node3d(0.45, 3.05, 2.1, 0.7, ENV, 'Веб-браузер', { fs: 10 });
  node3d(0.3, 4.02, 2.45, 0.5, 'device', 'Почтовый сервер площадки');
  node3d(0.3, 4.66, 2.45, 0.5, 'device', 'Сервисы аналитики и покупок');
  // Сервер платформы
  node3d(3.65, 1.15, 3.2, 4.01, 'device', 'Сервер платформы');
  node3d(3.8, 1.62, 2.9, 3.46, ENV, 'Docker', { fs: 10 });
  artifact(3.97, 2.1, 2.5, 0.82, 'Веб-сервер и шлюзы', 'Nginx, API-шлюз, входной прокси');
  artifact(3.97, 3.06, 2.5, 0.82, 'Сервисы системы', 'учётные записи, проекты, модерация, оркестрация');
  artifact(3.97, 4.02, 2.5, 0.82, 'Хранилища данных', 'PostgreSQL, Valkey, SeaweedFS');
  // Вычислительный узел
  node3d(7.75, 1.75, 2.0, 2.75, 'device', 'Вычислительный узел');
  node3d(7.88, 2.22, 1.74, 2.2, ENV, 'Docker', { fs: 10 });
  artifact(8.0, 2.7, 1.42, 0.6, 'Агент узла');
  artifact(8.0, 3.42, 1.42, 0.86, 'Игровые серверы', 'экземпляры, сервисы данных');
  // Каналы связи между устройствами
  path([[2.75, 1.8], [3.65, 1.8]]); proto('«HTTPS, WSS»', 2.75, 1.61, 0.9);
  path([[2.75, 3.25], [3.65, 3.25]]); proto('«HTTPS»', 2.75, 3.06, 0.9);
  path([[2.75, 4.31], [3.65, 4.31]]); proto('«SMTP»', 2.75, 4.12, 0.9);
  path([[2.75, 4.95], [3.65, 4.95]]); proto('«HTTPS»', 2.75, 4.76, 0.9);
  path([[6.85, 2.9], [7.75, 2.9]]); proto('«gRPC»', 6.85, 2.71, 0.9);
  path([[6.85, 3.6], [7.75, 3.6]]); proto('«WSS»', 6.85, 3.41, 0.9);
  text(s, '1', { x: 6.88, y: 2.93, w: 0.2, h: 0.16, fontSize: 9 }); text(s, '1..*', { x: 7.4, y: 2.93, w: 0.33, h: 0.16, fontSize: 9, align: 'right' });
  path([[1.5, 1.15], [1.5, 0.98], [8.75, 0.98], [8.75, 1.75]]);
  proto('«WSS» — прямое подключение игрока к узлу', 6.95, 1.03, 1.75, { h: 0.36, align: 'right' });
  text(s, fig('Диаграмма развёртывания'), { x: 0.4, y: 5.22, w: 9.2, h: 0.25, fontSize: 11, align: 'center' });
  num(s);
  speechNotes(s);
}

// ── 20. Архитектура системы: диаграмма контейнеров C4 (уровень 2) — детализация сервера платформы и узла ──
{
  const s = slideTitle('Архитектура системы');
  // Контейнер C4: название, [технология], назначение
  const box = (x, y, w, h, name, tech, desc, o = {}) => {
    s.addShape(pres.shapes.RECTANGLE, { x, y, w, h, fill: { color: o.fill || C.white }, line: { color: C.black, width: 1 } });
    text(s, [{ text: name, options: { bold: true, fontSize: o.fs || 10, breakLine: true } }, { text: `[${tech}]`, options: { fontSize: 8, breakLine: !!desc } },
      ...(desc ? [{ text: desc, options: { fontSize: 8.5 } }] : [])], { x: x + 0.04, y, w: w - 0.08, h, align: 'center', valign: 'middle' });
    return { x, y, w, h, cx: x + w / 2, cy: y + h / 2, b: y + h, r: x + w };
  };
  // Пользователь C4: «голова» и подпись
  const person = (cx, y, w, name) => {
    s.addShape(pres.shapes.OVAL, { x: cx - 0.11, y, w: 0.22, h: 0.22, fill: { color: C.white }, line: { color: C.black, width: 1 } });
    s.addShape(pres.shapes.ROUNDED_RECTANGLE, { x: cx - w / 2, y: y + 0.2, w, h: 0.42, rectRadius: 0.08, fill: { color: C.white }, line: { color: C.black, width: 1 } });
    text(s, [{ text: name, options: { bold: true, fontSize: 10, breakLine: true } }, { text: '[Пользователь]', options: { fontSize: 8 } }], { x: cx - w / 2, y: y + 0.2, w, h: 0.42, align: 'center', valign: 'middle' });
    return { cx, b: y + 0.62 };
  };
  const boundary = (x, y, w, h, name, top) => {
    s.addShape(pres.shapes.RECTANGLE, { x, y, w, h, fill: { type: 'none' }, line: { color: C.black, width: 1, dashType: 'dash' } });
    text(s, name, { x: x + 0.08, y: top ? y + 0.03 : y + h - 0.21, w: w - 0.16, h: 0.18, fontSize: 9, bold: true, valign: 'middle' });
  };
  const arrow = (pts) => pts.slice(1).forEach((p, i) => seg(s, pts[i][0], pts[i][1], p[0], p[1], { width: 1, ...(i === pts.length - 2 ? { endArrowType: 'triangle' } : {}) }));
  const lab = (t, x, y, w, al = 'left') => text(s, t, { x, y, w, h: 0.15, fontSize: 8, align: al, valign: 'middle' });
  boundary(0.3, 1.77, 7.36, 3.37, 'Разрабатываемая система');
  boundary(7.98, 1.77, 1.77, 2.5, 'Вычислительный узел', true);
  const pu = person(2.65, 0.88, 3.0, 'Разработчик, модератор, администратор');
  const pg = person(5.95, 0.88, 1.3, 'Игрок');
  const web = box(1.55, 1.92, 2.2, 0.62, 'Веб-приложение', 'Vue 3, TypeScript', 'личный кабинет');
  const cli = box(4.85, 1.92, 2.2, 0.62, 'Клиент игры', 'JavaScript', 'модуль обнаружения сервера');
  const gw = box(1.55, 2.84, 2.2, 0.54, 'API-шлюз', 'Go, grpc-gateway', 'REST в gRPC');
  const px = box(4.85, 2.84, 2.2, 0.54, 'Входной прокси', 'Go', 'маршруты игрового трафика');
  const SW = 1.66, SX = [0, 1, 2, 3].map((i) => 0.45 + i * (SW + 0.15)), SY = 3.68, SH = 0.5;
  const sv = ['Сервис учётных записей', 'Сервис управления проектами', 'Сервис модерации', 'Сервис оркестрации'].map((t, i) => box(SX[i], SY, SW, SH, t, 'Go, gRPC', '', { fs: 9.5 }));
  const DW = 2.1, DX = [0.55, 2.92, 5.29], DY = 4.46, DH = 0.42;
  const db = [['База данных', 'PostgreSQL 17'], ['Хранилище состояния', 'Valkey 8'], ['Объектное хранилище', 'SeaweedFS, S3']].map(([t, tech], i) => box(DX[i], DY, DW, DH, t, tech, '', { fs: 9.5, fill: C.soft }));
  const inst = box(8.09, 2.08, 1.56, 0.64, 'Экземпляр игрового сервера', 'Контейнер Docker', '', { fs: 9.5 });
  const ag = box(8.09, 3.1, 1.56, 0.62, 'Агент узла', 'Go, gRPC', 'управление контейнерами', { fs: 9.5 });
  // Связи: кто кого вызывает и по какому протоколу
  arrow([[pu.cx, pu.b], [pu.cx, web.y]]); lab('работает в кабинете [HTTPS]', pu.cx + 0.06, 1.56, 1.9);
  arrow([[pg.cx, pg.b], [pg.cx, cli.y]]); lab('играет [HTTPS]', pg.cx + 0.06, 1.56, 1.2);
  arrow([[web.cx, web.b], [web.cx, gw.y]]); lab('вызывает [REST, HTTPS]', web.cx - 1.46, 2.62, 1.4, 'right');
  arrow([[cli.x + 0.4, cli.b], [cli.x + 0.4, 2.71], [gw.r - 0.25, 2.71], [gw.r - 0.25, gw.y]]); lab('ищет сервер [REST]', 3.72, 2.55, 1.4, 'center');
  arrow([[cli.cx + 0.3, cli.b], [cli.cx + 0.3, px.y]]); lab('игровой трафик [WebSocket]', cli.cx + 0.36, 2.62, 1.5);
  seg(s, gw.cx, gw.b, gw.cx, 3.54, { width: 1 }); seg(s, sv[0].cx, 3.54, sv[3].cx, 3.54, { width: 1 });
  sv.forEach((v) => arrow([[v.cx, 3.54], [v.cx, SY]])); lab('[gRPC]', gw.cx + 0.06, 3.385, 0.6);
  sv.forEach((v) => seg(s, v.cx, v.b, v.cx, 4.32, { width: 1 })); seg(s, sv[0].cx, 4.32, sv[3].cx, 4.32, { width: 1 });
  db.forEach((d) => arrow([[d.cx, 4.32], [d.cx, DY]])); lab('[SQL, RESP, S3]', sv[3].cx + 0.06, 4.2, 0.9);
  arrow([[px.r, px.cy], [7.82, px.cy], [7.82, inst.cy], [inst.x, inst.cy]]); lab('[WebSocket]', px.r + 0.02, px.cy - 0.17, 0.62, 'center');
  arrow([[sv[3].r, sv[3].cy], [ag.cx, sv[3].cy], [ag.cx, ag.b]]); lab('управляет [gRPC]', 8.02, sv[3].cy + 0.03, 1.0);
  arrow([[ag.cx, ag.y], [ag.cx, inst.b]]); lab('[Docker API]', ag.cx + 0.05, 2.83, 0.75);
  // Обозначения
  text(s, 'Обозначения:', { x: 7.98, y: 4.4, w: 1.8, h: 0.18, fontSize: 9, bold: true });
  s.addShape(pres.shapes.RECTANGLE, { x: 8.0, y: 4.64, w: 0.3, h: 0.16, fill: { color: C.white }, line: { color: C.black, width: 1 } });
  text(s, '— контейнер', { x: 8.36, y: 4.62, w: 1.4, h: 0.2, fontSize: 8.5, valign: 'middle' });
  s.addShape(pres.shapes.RECTANGLE, { x: 8.0, y: 4.88, w: 0.3, h: 0.16, fill: { color: C.soft }, line: { color: C.black, width: 1 } });
  text(s, '— хранилище данных', { x: 8.36, y: 4.86, w: 1.45, h: 0.2, fontSize: 8.5, valign: 'middle' });
  text(s, fig('Диаграмма контейнеров (C4, уровень 2)'), { x: 0.4, y: 5.22, w: 9.2, h: 0.25, fontSize: 11, align: 'center' });
  num(s);
  speechNotes(s);
}

// ── 21. Даталогическая модель: таблицы и поля — как в init.sql сервисов ──
// Типы и число полей читаются из src/<сервис>/migrations/init.sql, поэтому слайд не расходится со схемой.
function sqlTables(svc) {
  const sql = fs.readFileSync(`../../../src/${svc}/migrations/init.sql`, 'utf8').replace(/--[^\n]*/g, '');
  const T = {};
  for (const m of sql.matchAll(/CREATE TABLE(?: IF NOT EXISTS)?\s+(\w+)\s*\((.*?)\n\);/gs)) {
    const cols = [];
    m[2].split(/,\s*\n/).forEach((l) => {
      l = l.trim().replace(/\s+/g, ' ');
      if (!l || /^(PRIMARY KEY|FOREIGN KEY|UNIQUE|CONSTRAINT|CHECK)\b/i.test(l)) return;
      const c = l.match(/^(\w+) ([A-Za-z]+(?:\(\d+\))?(?:\[\])?(?: WITH TIME ZONE)?)/);
      if (!c) return;
      const type = c[2].toLowerCase().replace(/^bigserial$/, 'bigint').replace(/^integer$/, 'int').replace('timestamp with time zone', 'timestamptz');
      cols.push({ name: c[1], type, pk: /PRIMARY KEY/i.test(l), fk: /REFERENCES/i.test(l) });
    });
    T[m[1]] = cols;
  }
  return T;
}
{
  const s = slideTitle('Даталогическая модель');
  const er = erDiagram(s, { fs: 8.3, rowH: 0.15, tfs: 9 });
  const Y = [0.9, 2.49, 4.08]; // высота таблицы 1,09: пять строк
  const DB = { sso: sqlTables('sso'), pm: sqlTables('project-manager'), mod: sqlTables('moderation'), orc: sqlTables('orchestrator') };
  // [роль в расстановке, сервис, таблица, показанные поля, поля-ссылки без внешнего ключа].
  // Порядок полей — как в erLayout: первое — первичный ключ, следом ссылочные.
  const grid = [
    [null,
      ['draft', 'pm', 'project_drafts', ['project_id', 'title_ru', 'about_ru', 'icon_path'], []],
      ['build', 'pm', 'project_builds', ['id', 'project_id', 'version', 'file_path'], []],
      ['release', 'pm', 'project_releases', ['id', 'project_id', 'version', 'prod_url'], []],
      ['grant', 'orc', 'platform_grants', ['id', 'game_id', 'max_instances', 'is_active'], ['game_id']]],
    [['user', 'sso', 'users', ['id', 'email', 'role', 'status'], []],
      ['member', 'pm', 'project_members', ['id', 'project_id', 'user_id', 'permissions'], ['user_id']],
      ['project', 'pm', 'projects', ['id', 'owner_id', 'status', 'is_online'], ['owner_id']],
      ['request', 'mod', 'moderation_requests', ['id', 'project_id', 'status', 'request_type'], ['project_id']],
      ['snapshot', 'mod', 'moderation_snapshots', ['id', 'request_id', 'status', 'snapshot_json'], []]],
    [['node', 'orc', 'nodes', ['id', 'owner_id', 'address', 'status'], ['owner_id']],
      ['instance', 'orc', 'instances', ['id', 'node_id', 'server_build_id', 'status'], []],
      ['sbuild', 'orc', 'server_builds', ['id', 'game_id', 'version', 'image_tag'], ['game_id']],
      ['policy', 'orc', 'game_policies', ['game_id', 'mode', 'target_instances', 'auto_restart'], ['game_id']],
      ['message', 'mod', 'moderation_messages', ['id', 'request_id', 'project_id', 'content'], ['project_id']]],
  ];
  const id = {};
  let shown = 0;
  grid.forEach((row, r) => row.forEach((e, k) => {
    if (!e) return;
    const [key, svc, table, show, logical] = e, cols = DB[svc][table];
    const rows = show.map((name) => {
      const c = cols.find((q) => q.name === name);
      if (!c) throw new Error(`нет поля ${table}.${name}`);
      const pre = [c.pk && 'PK', c.fk && 'FK'].filter(Boolean).join(', ');
      return [...(pre ? [{ text: pre + ' ', options: { bold: true } }] : []), { text: `${c.name} : ${c.type}`, options: { italic: logical.includes(name) } }];
    });
    rows.push(`… ещё ${cols.length - show.length}`);
    er.entity(table, ER_X[k], Y[r], ER_W, table, rows);
    id[key] = table; shown++;
  }));
  erLayout(er, id, Y, true);
  erLegend(s, er, true);
  const total = Object.values(DB).reduce((a2, t) => a2 + Object.keys(t).length, 0);
  text(s, fig(`Даталогическая модель: основные таблицы (${shown} из ${total}), база данных PostgreSQL`), { x: 0.4, y: 5.25, w: 9.2, h: 0.25, fontSize: 11, align: 'center' });
  num(s);
  speechNotes(s);
}

// Текстовый слайд из колонок: [заголовок колонки, [[жирное начало, продолжение] | строка, …]]
function listSlide(title, cols, o = {}) {
  const s = slideTitle(title);
  const gap = o.gap || 0.4, cw = (9.0 - gap * (cols.length - 1)) / cols.length, fz = o.fs || 15;
  cols.forEach(([head, items], i) => {
    const x = 0.5 + i * (cw + gap);
    // Подзаголовок — отдельной надписью вплотную к списку: без пустой строки между ними
    text(s, head, { x, y: 1.0, w: cw, h: 0.36, fontSize: fz + 2, bold: true, color: C.accent, valign: 'middle' });
    const runs = [];
    items.forEach((it, k) => {
      const last = k === items.length - 1;
      if (Array.isArray(it)) runs.push({ text: it[0], options: { bold: true, bullet: o.bullet ? { indent: 14 } : undefined } }, { text: it[1], options: { breakLine: !last } });
      else runs.push({ text: it, options: { bullet: o.bullet ? { indent: 14 } : undefined, breakLine: !last } });
    });
    text(s, runs, { x, y: 1.42, w: cw, h: 3.9, fontSize: fz, valign: 'top', paraSpaceAfter: o.after || 7 });
  });
  num(s);
  speechNotes(s);
}

// ── 22. Входные и выходные данные ──
listSlide('Входные и выходные данные', [
  ['Входные данные', [
    'Клиентская сборка веб-игры',
    'Описание игры, медиа-материалы, внутриигровые товары',
    'Серверная сборка игры',
    'Политика оркестрации и команды управления серверами',
    'Сведения о вычислительных узлах: адрес, ключ, ресурсы',
    'Решения и замечания модератора',
    'Запросы клиентов игр на подключение к серверу',
  ]],
  ['Выходные данные', [
    'Опубликованная игра в каталоге площадки',
    'Заявки на модерацию и аудит-снимки решений',
    'Адрес экземпляра игрового сервера или место в очереди',
    'Состояние узлов и экземпляров, журналы серверов',
    'Ежедневная аналитика игры с выгрузкой в CSV',
    'Сообщения в чате проекта и письма с кодами подтверждения',
  ]],
], { bullet: true });

// ── 23. Разработанные интерфейсы: один слайд, снимки экранов каскадом ──
// Первый снимок виден сразу, каждый следующий по щелчку ложится поверх со сдвигом; подпись меняется вместе с ним.
// Снимки делает shots.cjs со стенда (1440×810, светлая тема). Тонкая рамка добавлена в сами файлы снимков (magick -border 3):
// отдельная фигура-рамка в части программ показа оставалась видна поверх ещё не появившихся снимков.
{
  const s = slideTitle('Разработанные интерфейсы');
  const shots = [
    ['ui-draft.jpg', 'Черновик игры'], ['ui-stats.jpg', 'Аналитика игры'], ['ui-servers.jpg', 'Игровые серверы проекта'],
    ['ui-node.jpg', 'Вычислительный узел разработчика'], ['ui-queue.jpg', 'Очередь заявок модератора'], ['ui-review.jpg', 'Проверка игры в песочнице и чат проекта'],
    ['ui-catalog.jpg', 'Каталог опубликованных игр'], ['ui-admin.jpg', 'Управление разработчиками'],
  ];
  const n = shots.length, dx = 0.12, dy = 0.045, top = 0.9, ih = 4.28 - dy * (n - 1), iw = ih * 2166 / 1221, x0 = (W - iw - dx * (n - 1)) / 2;
  shots.forEach(([f, cap], i) => {
    const x = x0 + i * dx, y = top + i * dy, g = i ? `a${i}_` : 'ui0_';
    s.addImage({ path: IMG(f), x, y, w: iw, h: ih, objectName: `${g}shot` });
    s.addText(fig(cap), { x: 0.4, y: 5.22, w: 9.2, h: 0.28, fontFace: FONT, fontSize: 11.5, color: C.black, align: 'center', valign: 'middle', margin: 0, fill: { color: C.white }, objectName: `${g}cap` });
  });
  num(s);
  speechNotes(s);
}

// ── 24. Системные требования ──
listSlide('Системные требования', [
  ['Сервер платформы', [
    'Linux, архитектура x86-64',
    'Docker Engine 24 и новее, Docker Compose 2.20 и новее',
    'Процессор — от 2 ядер',
    'Оперативная память — от 4 ГБ',
    'Диск — от 100 ГБ',
    'Два доменных имени с сертификатами TLS',
  ]],
  ['Вычислительный узел', [
    'Linux, архитектура x86-64',
    'Docker Engine 24 и новее',
    'Агент узла; порт 44044 открыт для сервиса оркестрации',
    'Процессор и память — по числу экземпляров: пределы задаются для каждого',
  ]],
  ['Пользователь и игрок', [
    'Google Chrome 120 и новее',
    'Яндекс Браузер 24.1 и новее',
    'Mozilla Firefox 121 и новее',
    'Microsoft Edge 120, Safari 17 и новее',
    'Экран шириной от 1280 точек для личного кабинета',
    'Подключение к сети Интернет',
  ]],
], { bullet: true, fs: 13.5, after: 6, gap: 0.3 });

// ── 25. Сведения о проекте ──
listSlide('Сведения о проекте', [
  ['Серверная часть', [
    ['8', ' сервисов'],
    ['162', ' метода программного интерфейса'],
    ['≈ 41 000', ' строк кода на Go, без тестов'],
    ['433', ' автоматических теста, ≈ 25 300 строк'],
  ]],
  ['База данных', [
    ['23', ' таблицы в PostgreSQL'],
    ['4', ' набора таблиц — у каждого сервиса свой'],
  ]],
  ['Веб-интерфейс', [
    ['26', ' экранов для трёх ролей'],
    ['64', ' компонента Vue'],
    ['≈ 54 700', ' строк кода'],
  ]],
], { fs: 17, after: 10 });

// ── 26. Средства разработки: версии — из go.mod, package.json, файлов Docker и рабочего места ──
listSlide('Средства разработки', [
  ['Разработка', [
    ['Серверная часть: ', 'Go 1.25, gRPC 1.80, Protocol Buffers 3 (protoc 36.1), grpc-gateway 2.29'],
    ['Веб-интерфейс: ', 'Vue 3.5, TypeScript 5.8, Vite 7.3, Node.js 22'],
    ['Тестирование: ', 'testify 1.11, Playwright 1.63'],
    ['Контроль версий: ', 'Git 2.56'],
  ]],
  ['Хранение и развёртывание', [
    ['СУБД: ', 'PostgreSQL 17'],
    ['Состояние и блокировки: ', 'Valkey 8'],
    ['Объектное хранилище: ', 'SeaweedFS 4.46'],
    ['Контейнеры: ', 'Docker 29.9, Docker Compose 5.7'],
    ['Веб-сервер: ', 'Nginx 1.27'],
  ]],
  ['Рабочее место', [
    ['Операционная система: ', 'CachyOS, ядро Linux 7.3'],
    ['Редактор кода: ', 'Zed 1.27'],
    ['Проектирование: ', 'draw.io 31.7, PlantUML 1.2026.8'],
    ['Документация: ', 'ONLYOFFICE Desktop Editors 10.0'],
  ]],
], { fs: 13.5, after: 7, gap: 0.3 });

// ── 27. Заключение ─────────────────────────────────────────
// Справа — место под фотографию подписанного акта о внедрении: достаточно положить файл img/act.jpg.
{
  const s = slideTitle('Заключение');
  const done = [
    'Проанализированы предметная область и деятельность заказчика, построены модели процессов «как есть» и «как будет».',
    'Проведено сравнение с аналогами, разработано техническое задание по ГОСТ 34.602-2020.',
    'Спроектированы архитектура, модели данных и алгоритмы оркестрации игровых серверов.',
    'Реализована система из восьми сервисов и веб-интерфейса, код проверен 433 автоматическими тестами.',
  ];
  const LW = 6.0;
  text(s, [
    { text: 'В ходе работы:', options: { bold: true, breakLine: true } },
    ...done.map((t, i) => ({ text: t, options: { bullet: { type: 'number', style: 'arabicParenR' }, breakLine: i < done.length - 1 } })),
  ], { x: 0.4, y: 0.95, w: LW, h: 2.2, fontSize: 13, valign: 'top', align: 'justify', paraSpaceAfter: 3 });
  // Сравнение моделей процессов: подсчёт по схемам BPMN (docs/business-process/bpmn/README.md)
  const cell = (t, o = {}) => ({ text: t, options: { fontFace: FONT, fontSize: 12, align: 'center', valign: 'middle', color: C.black, margin: [1, 6, 1, 6], ...o } });
  const rows = [
    ['Участников-людей в процессе', '4', '2'],
    ['Ручных операций инженера и менеджера', '7', '0'],
    ['Передач через мессенджер', '8', '0'],
    ['Автоматических операций', '3', '12'],
    ['Задержка получения статистики', 'до 30 дней', 'до суток'],
  ];
  const data = [['Критерий', 'Как есть', 'Как будет'].map((h, i) => cell(h, { bold: true, align: i ? 'center' : 'left', fill: { color: C.soft } }))];
  rows.forEach((r) => data.push(r.map((v, i) => cell(v, i === 0 ? { align: 'left' } : { bold: true, color: i === 1 ? C.red : C.green }))));
  s.addTable(data, { x: 0.4, y: 3.3, w: LW, colW: [3.5, 1.25, 1.25], rowH: data.map(() => 0.3), border: { type: 'solid', pt: 0.75, color: 'BFBFBF' } });
  // Акт о внедрении: лист А4 в книжной ориентации
  const ah = 4.0, aw = ah / 1.414, ax = 6.4 + (3.2 - aw) / 2, ay = 0.95;
  if (fs.existsSync(IMG('act.jpg'))) s.addImage({ path: IMG('act.jpg'), x: ax, y: ay, w: aw, h: ah, sizing: { type: 'contain', w: aw, h: ah } });
  else text(s, 'Место для фотографии подписанного акта о внедрении', { x: ax + 0.2, y: ay, w: aw - 0.4, h: ah, fontSize: 13, color: '7F7F7F', align: 'center', valign: 'middle' });
  s.addShape(pres.shapes.RECTANGLE, { x: ax, y: ay, w: aw, h: ah, fill: { type: 'none' }, line: { color: C.black, width: 0.75, dashType: fs.existsSync(IMG('act.jpg')) ? 'solid' : 'dash' } });
  text(s, fig('Акт о внедрении'), { x: 6.4, y: ay + ah + 0.06, w: 3.2, h: 0.25, fontSize: 11, align: 'center' });
  num(s);
  speechNotes(s);
}

// ══ Скрытые резервные слайды (после заключения): показываются только при вопросах комиссии ══
const MAIN_SLIDES = pres.slides.length;

figSlide('Диаграмма состояний', [['instance-state.png', 'Состояния экземпляра игрового сервера']]);
figSlide('Диаграмма последовательности', [['instance-start-sequence.png', 'Запуск экземпляра игрового сервера']], { maxH: 4.2, top: 0.88 });
figSlide('Диаграмма последовательности', [['discovery-sequence.png', 'Обнаружение сервера и подключение игрока']], { maxH: 4.2, top: 0.88 });
figSlide('Диаграмма деятельности', [['node-select-activity.png', 'Выбор узла для экземпляра']], { maxH: 4.2, top: 0.88 });
figSlide('Диаграмма классов', [['class-orchestrator.png', 'Классы предметной области сервиса оркестрации']], { maxH: 4.2, top: 0.88 });
figSlide('Даталогическая модель', [['er-orchestrator.png', 'Таблицы сервиса оркестрации: узлы и экземпляры'], ['er-orchestrator-policies.png', 'Таблицы сервиса оркестрации: политики и доступ']], { cols: true });
figSlide('Даталогическая модель', [['er-projects.png', 'Таблицы сервиса управления проектами']]);
figSlide('Даталогическая модель', [['er-moderation.png', 'Таблицы сервиса модерации'], ['er-accounts.png', 'Таблицы сервиса учётных записей']], { cols: true });

// Список скрытых слайдов для postprocess.py
fs.writeFileSync('hidden.json', JSON.stringify(Array.from({ length: pres.slides.length - MAIN_SLIDES }, (_, i) => MAIN_SLIDES + i + 1)));

pres.writeFile({ fileName: 'deck.pptx' }).then(() => console.log(`written deck.pptx: ${MAIN_SLIDES} основных слайдов, ${pres.slides.length - MAIN_SLIDES} резервных`));
