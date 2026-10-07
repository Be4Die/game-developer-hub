#!/bin/bash
# Сборка презентации и текста доклада:
#   deck.js (pptxgenjs) -> postprocess.py (переходы «Трансформация» и анимации) -> ../../Презентация_ВКР_Предзащита.pptx
#   speech_docx.js (docx) -> ../../Речь_Предзащита_ВКР.docx
# Текст доклада один — speech.js: из него же берутся заметки докладчика к слайдам.
# Перед первой сборкой: npm install (pptxgenjs, docx).
set -e
cd "$(dirname "$0")"
node deck.js
python3 postprocess.py deck.pptx "../../Презентация_ВКР_Предзащита.pptx"
rm -f deck.pptx
node speech_docx.js "../../Речь_Предзащита_ВКР.docx"
