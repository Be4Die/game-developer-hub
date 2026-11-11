#!/bin/bash
# Рисунки статьи: BPMN (as_is, to_be) — генераторы as_is.py и to_be.py; остальное — PlantUML.
# Диаграмма деятельности собирается из двух частей (а, б), поставленных рядом.
set -e
cd "$(dirname "$0")"
plantuml -tpng -o .. containers.puml node_state.puml instance_state.puml
T=$(mktemp -d)
plantuml -tpng -o "$T" cycle_nodes.puml cycle_policy.puml
H=$(identify -format '%h\n' "$T"/cycle_*.png | sort -n | tail -1)
lab() { magick "$1" -background white -gravity north -extent "$(identify -format %w "$1")x$H" -gravity south -splice 0x100 -font Times-New-Roman -pointsize 15 -density 200 -annotate +0+25 "$2" "$3"; }
lab "$T/cycle_nodes.png" 'а)' "$T/a.png"; lab "$T/cycle_policy.png" 'б)' "$T/b.png"
magick "$T/a.png" \( -size 110x$((H + 100)) xc:white \) "$T/b.png" +append ../cycle_activity.png
rm -rf "$T"
