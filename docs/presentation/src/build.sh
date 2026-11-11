#!/bin/bash
# Сборка презентации и текста доклада к защите (версия для предзащиты — в истории git, файлы лежат рядом):
#   diagrams/*.puml (PlantUML, крупный кегль для слайдов) -> img/
#   deck.js (pptxgenjs) -> postprocess.py (переходы «Трансформация» и анимации) -> ../../Презентация_ВКР_Защита.pptx
#   speech_docx.js (docx) -> ../../Речь_Защита_ВКР.docx
# Текст доклада один — speech.js: из него же берутся заметки докладчика к слайдам.
# Снимки экранов интерфейса переснимаются отдельно: node shots.cjs img (нужен запущенный стенд http://localhost).
# Перед первой сборкой: npm install (pptxgenjs, docx).
set -e
cd "$(dirname "$0")"
plantuml -tpng -o ../img diagrams/*.puml
node deck.js
python3 postprocess.py deck.pptx "../../Презентация_ВКР_Защита.pptx"
rm -f deck.pptx anim.json hidden.json
node speech_docx.js "../../Речь_Защита_ВКР.docx"
