"""Добавляет в deck.pptx переходы «Трансформация» и анимации появления по щелчку.

Переход задаётся на слайде, НА который переходят: слайды 3+ получают «Трансформацию» (запасной вариант — «Выцветание»),
переход с титульного слайда на второй остаётся без эффекта.
Анимация: объекты с именем «aN_…» (или с замещающим текстом «anim:aN») появляются с выцветанием,
все объекты одной группы N — по одному щелчку. Группы «bN» запускаются сами сразу после перехода,
группа N — с задержкой (N-1)·0,4 с.
"""
import re, sys, os, zipfile, json, hashlib

src, dst = sys.argv[1], sys.argv[2]
ANIM = json.load(open('anim.json')) if os.path.exists('anim.json') else {}  # номер слайда → {имя объекта: группа «b1»}
MORPH = ('<mc:AlternateContent xmlns:mc="http://schemas.openxmlformats.org/markup-compatibility/2006">'
         '<mc:Choice xmlns:p159="http://schemas.microsoft.com/office/powerpoint/2015/09/main" Requires="p159">'
         '<p:transition spd="slow" xmlns:p14="http://schemas.microsoft.com/office/powerpoint/2010/main" p14:dur="900">'
         '<p159:morph option="byObject"/></p:transition></mc:Choice>'
         '<mc:Fallback><p:transition spd="slow"><p:fade/></p:transition></mc:Fallback></mc:AlternateContent>')


def timing(groups, shape_ids):
    nid = [2]
    def new():
        nid[0] += 1
        return nid[0]
    # Шаги: сначала автоматический (все группы «b»), затем щелчки (группы «a»)
    steps = []
    auto = [(spid, (g[1] - 1) * 400) for g in sorted(k for k in groups if k[0] == 'b') for spid in groups[g]]
    if auto:
        steps.append(('auto', auto))
    for g in sorted(k for k in groups if k[0] == 'a'):
        steps.append(('click', [(spid, 0) for spid in groups[g]]))
    clicks = []
    for kind, members in steps:
        effects = []
        for k, (spid, delay) in enumerate(members):
            a, b, c = new(), new(), new()
            node = 'withEffect' if k else ('afterEffect' if kind == 'auto' else 'clickEffect')
            effects.append(
                f'<p:par><p:cTn id="{a}" presetID="10" presetClass="entr" presetSubtype="0" fill="hold" grpId="0" nodeType="{node}">'
                f'<p:stCondLst><p:cond delay="{delay}"/></p:stCondLst><p:childTnLst>'
                f'<p:set><p:cBhvr><p:cTn id="{b}" dur="1" fill="hold"><p:stCondLst><p:cond delay="0"/></p:stCondLst></p:cTn>'
                f'<p:tgtEl><p:spTgt spid="{spid}"/></p:tgtEl><p:attrNameLst><p:attrName>style.visibility</p:attrName></p:attrNameLst>'
                f'</p:cBhvr><p:to><p:strVal val="visible"/></p:to></p:set>'
                f'<p:animEffect transition="in" filter="fade"><p:cBhvr><p:cTn id="{c}" dur="500"/>'
                f'<p:tgtEl><p:spTgt spid="{spid}"/></p:tgtEl></p:cBhvr></p:animEffect>'
                f'</p:childTnLst></p:cTn></p:par>')
        o, i = new(), new()
        start = '<p:cond delay="indefinite"/>' + ('<p:cond evt="onBegin" delay="0"><p:tn val="2"/></p:cond>' if kind == 'auto' else '')
        clicks.append(f'<p:par><p:cTn id="{o}" fill="hold"><p:stCondLst>{start}</p:stCondLst><p:childTnLst>'
                      f'<p:par><p:cTn id="{i}" fill="hold"><p:stCondLst><p:cond delay="0"/></p:stCondLst><p:childTnLst>'
                      + ''.join(effects) + '</p:childTnLst></p:cTn></p:par></p:childTnLst></p:cTn></p:par>')
    bld = ''.join(f'<p:bldP spid="{sid}" grpId="0" animBg="1"/>' for g in groups for sid in groups[g] if sid in shape_ids)
    return ('<p:timing><p:tnLst><p:par><p:cTn id="1" dur="indefinite" restart="never" nodeType="tmRoot"><p:childTnLst>'
            '<p:seq concurrent="1" nextAc="seek"><p:cTn id="2" dur="indefinite" nodeType="mainSeq"><p:childTnLst>'
            + ''.join(clicks) +
            '</p:childTnLst></p:cTn><p:prevCondLst><p:cond evt="onPrev" delay="0"><p:tgtEl><p:sldTgt/></p:tgtEl></p:cond></p:prevCondLst>'
            '<p:nextCondLst><p:cond evt="onNext" delay="0"><p:tgtEl><p:sldTgt/></p:tgtEl></p:cond></p:nextCondLst></p:seq>'
            '</p:childTnLst></p:cTn></p:par></p:tnLst>' + (f'<p:bldLst>{bld}</p:bldLst>' if bld else '') + '</p:timing>')


zin = zipfile.ZipFile(src)
zout = zipfile.ZipFile(dst, 'w', zipfile.ZIP_DEFLATED)
# Одинаковые картинки (одна схема на всех кадрах приближения) храним один раз
first, alias = {}, {}
for item in zin.infolist():
    if item.filename.startswith('ppt/media/'):
        h = hashlib.sha1(zin.read(item.filename)).hexdigest()
        if h in first:
            alias[item.filename.split('/')[-1]] = first[h].split('/')[-1]
        else:
            first[h] = item.filename
for item in zin.infolist():
    if item.filename.split('/')[-1] in alias and item.filename.startswith('ppt/media/'):
        continue
    data = zin.read(item.filename)
    if item.filename.endswith('.rels') and alias:
        x = data.decode('utf8')
        for a, b in alias.items():
            x = x.replace('/media/' + a + '"', '/media/' + b + '"')
        data = x.encode('utf8')
    m = re.fullmatch(r'ppt/slides/slide(\d+)\.xml', item.filename)
    if m:
        n = int(m.group(1))
        x = data.decode('utf8')
        groups, sps = {}, set()
        for el in re.finditer(r'<p:(sp|pic)>.*?</p:\1>', x, re.S):
            c = re.search(r'<p:cNvPr id="(\d+)" name="([^"]*)"', el.group(0))
            g = re.match(r'([ab])(\d+)_', c.group(2)) or re.fullmatch(r'([ab])(\d+)', ANIM.get(str(n), {}).get(c.group(2), ''))
            if g:
                groups.setdefault((g.group(1), int(g.group(2))), []).append(c.group(1))
                if el.group(1) == 'sp':
                    sps.add(c.group(1))
        extra = (MORPH if n >= 3 else '') + (timing(groups, sps) if groups else '')
        if extra:
            assert '<p:timing>' not in x and '<p:transition' not in x
            x = x.replace('</p:clrMapOvr>', '</p:clrMapOvr>' + extra, 1)
            print(f'slide{n}: morph={n >= 3} clicks={len(groups)}')
        data = x.encode('utf8')
    zout.writestr(item, data)
zout.close()
