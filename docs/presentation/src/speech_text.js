// Общее для заметок докладчика и документа с речью.
// MARK — пометка «нажать далее»: шаг анимации (появление выделения) или переход к следующему слайду.
const MARK = '[ДАЛЕЕ]';

// Абзацы речи одного слайда. Пометки внутри текста — шаги анимации; после фразы-перехода
// ставится пометка смены слайда (на последнем слайде перехода нет).
function paragraphs(sp) {
  return [...sp.text, ...(sp.next ? [`Переход: ${sp.next} ${MARK}`] : [])];
}

const countWords = (all) => all.reduce((n, sp) => n + paragraphs(sp).join(' ').split(/\s+/).filter((w) => w && w !== MARK).length, 0);

module.exports = { MARK, paragraphs, countWords };
