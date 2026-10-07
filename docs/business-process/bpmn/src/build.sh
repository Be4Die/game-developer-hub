#!/bin/bash
# Пересобирает BPMN-схемы: сценарии src/*.py -> *.drawio -> *.svg и *.png (draw.io в режиме экспорта).
# Внимание: ручные правки в .drawio перезаписываются.
set -e
cd "$(dirname "$0")"
for name in as_is_publication as_is_servers to_be_publication to_be_servers; do
  out=../${name//_/-}
  python3 $name.py $out.drawio
  drawio -x -f svg -o $out.svg $out.drawio --no-sandbox >/dev/null 2>&1
  drawio -x -f png -s 2 -o $out.png $out.drawio --no-sandbox >/dev/null 2>&1
  echo "$out.drawio"
done
