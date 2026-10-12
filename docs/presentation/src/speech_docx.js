// Собирает документ с текстом доклада из speech.js: node speech_docx.js <выходной.docx>
const fs = require('fs');
const { Document, Packer, Paragraph, TextRun, AlignmentType } = require('docx');
const SPEECH = require('./speech.js');
const { MARK, paragraphs, countWords } = require('./speech_text.js');

const FONT = 'Times New Roman', SIZE = 28; // 14 пт
const run = (text, o = {}) => new TextRun({ text, font: FONT, size: SIZE, ...o });
const para = (children, o = {}) => new Paragraph({
  children, spacing: { line: 360, after: 120 }, indent: { firstLine: 709 }, alignment: AlignmentType.JUSTIFIED, ...o,
});

const words = countWords(SPEECH);
const minutes = Math.round(words / 130); // спокойный темп доклада — около 130 слов в минуту

// Абзац речи: пометки «[ДАЛЕЕ]» выделяются цветом и жирным, фраза-переход — курсивом
const speechPara = (t) => {
  const isTransition = t.startsWith('Переход: ');
  const parts = t.split(MARK);
  const children = [];
  parts.forEach((p, i) => {
    if (i > 0) children.push(run(MARK, { bold: true, color: '2B5C8F' }));
    if (!p) return;
    if (isTransition && i === 0) {
      children.push(run('Переход: ', { italics: true, bold: true }), run(p.slice('Переход: '.length), { italics: true }));
    } else children.push(run(p));
  });
  return para(children);
};

const body = [
  para([run('Текст доклада на предварительной защите выпускной квалификационной работы', { bold: true })],
    { alignment: AlignmentType.CENTER, indent: { firstLine: 0 } }),
  para([run('«Автоматизированная система управления процессом публикации и сопровождения веб-игр»')],
    { alignment: AlignmentType.CENTER, indent: { firstLine: 0 }, spacing: { line: 360, after: 360 } }),
];
SPEECH.forEach((sp, i) => {
  body.push(para([run(`Слайд ${i + 1}. ${sp.title}`, { bold: true })], { indent: { firstLine: 0 }, spacing: { line: 360, before: 240, after: 120 }, keepNext: true }));
  paragraphs(sp).forEach((t) => body.push(speechPara(t)));
});
body.push(para([run('Примечания', { bold: true })], { indent: { firstLine: 0 }, spacing: { line: 360, before: 360, after: 120 } }));
[
  `Объём доклада — около ${words} слов, это примерно ${minutes} минут спокойной речи; регламент — 5–7 минут, при нехватке времени сокращаются пояснения на слайдах с диаграммами вариантов использования.`,
  'Названия слайдов вслух не повторяются. Текст не зачитывается слово в слово: можно менять порядок слов и падежи, главное — сохранить смысл и цифры.',
  `Пометка ${MARK} означает «нажать далее». Это не всегда новый слайд: на слайдах с моделями процессов и на сравнительной таблице так появляются выделения — каждое нужно сопроводить своим абзацем.`,
  'Фраза-переход произносится до нажатия на следующий слайд: она связывает слайды и удерживает внимание комиссии. Пометка после неё означает смену слайда.',
  'Можно подсматривать в текст, но обязательно поднимать взгляд на членов комиссии.',
].forEach((t) => body.push(para([run('— ' + t)])));

const doc = new Document({
  styles: { default: { document: { run: { font: FONT, size: SIZE } } } },
  sections: [{ properties: { page: { margin: { top: 1134, bottom: 1134, left: 1701, right: 850 } } }, children: body }],
});
Packer.toBuffer(doc).then((buf) => { fs.writeFileSync(process.argv[2], buf); console.log(`written ${process.argv[2]}: ${words} слов, ~${minutes} мин`); });
