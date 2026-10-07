// Собирает документ с текстом доклада из speech.js: node speech_docx.js <выходной.docx>
const fs = require('fs');
const { Document, Packer, Paragraph, TextRun, AlignmentType } = require('docx');
const SPEECH = require('./speech.js');

const FONT = 'Times New Roman', SIZE = 28; // 14 пт
const run = (text, o = {}) => new TextRun({ text, font: FONT, size: SIZE, ...o });
const para = (children, o = {}) => new Paragraph({
  children, spacing: { line: 360, after: 120 }, indent: { firstLine: 709 }, alignment: AlignmentType.JUSTIFIED, ...o,
});

const words = SPEECH.reduce((n, sp) => n + [...sp.text, sp.next || ''].join(' ').split(/\s+/).filter(Boolean).length, 0);
const minutes = Math.round(words / 130); // спокойный темп доклада — около 130 слов в минуту

const body = [
  para([run('Текст доклада на предварительной защите выпускной квалификационной работы', { bold: true })],
    { alignment: AlignmentType.CENTER, indent: { firstLine: 0 } }),
  para([run('«Автоматизированная система управления процессом публикации и сопровождения веб-игр»')],
    { alignment: AlignmentType.CENTER, indent: { firstLine: 0 }, spacing: { line: 360, after: 360 } }),
];
SPEECH.forEach((sp, i) => {
  body.push(para([run(`Слайд ${i + 1}. ${sp.title}`, { bold: true })], { indent: { firstLine: 0 }, spacing: { line: 360, before: 240, after: 120 }, keepNext: true }));
  sp.text.forEach((t) => body.push(para([run(t)])));
  if (sp.next) body.push(para([run('Переход: ', { italics: true, bold: true }), run(sp.next, { italics: true })]));
});
body.push(para([run('Примечания', { bold: true })], { indent: { firstLine: 0 }, spacing: { line: 360, before: 360, after: 120 } }));
[
  `Объём доклада — около ${words} слов, это примерно ${minutes} минут спокойной речи; регламент — 5–7 минут, при нехватке времени сокращаются пояснения на слайдах с диаграммами вариантов использования.`,
  'Названия слайдов вслух не повторяются. Текст не зачитывается слово в слово: можно менять порядок слов и падежи, главное — сохранить смысл и цифры.',
  'Фраза-переход произносится до щелчка на следующий слайд: она связывает слайды и удерживает внимание комиссии.',
  'На слайдах с моделями процессов выделения появляются по щелчку — щёлкать по мере рассказа о каждой проблеме или улучшении.',
  'Можно подсматривать в текст, но обязательно поднимать взгляд на членов комиссии.',
].forEach((t) => body.push(para([run('— ' + t)])));

const doc = new Document({
  styles: { default: { document: { run: { font: FONT, size: SIZE } } } },
  sections: [{ properties: { page: { margin: { top: 1134, bottom: 1134, left: 1701, right: 850 } } }, children: body }],
});
Packer.toBuffer(doc).then((buf) => { fs.writeFileSync(process.argv[2], buf); console.log(`written ${process.argv[2]}: ${words} слов, ~${minutes} мин`); });
