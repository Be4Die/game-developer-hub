"""Строит UML-диаграммы классов PlantUML из исходного кода Go: структуры, интерфейсы и перечисления
пакетов domain и service каждого сервиса.

Три уровня подробности:
  overview.puml, overview-fragment.puml — обзор для презентации (OVERVIEW, FRAGMENT);
  text/*.puml — рисунки для текста записки, под книжный лист А4 (TEXT, SERVICES);
  *.puml — полные листы для приложения (SHEETS).

Запуск из docs/classes:
  python3 src/classes.py && PLANTUML_LIMIT_SIZE=16384 plantuml -tpng -tsvg *.puml text/*.puml
"""
import re, pathlib

ROOT = pathlib.Path(__file__).resolve().parents[3] / 'src'
MAX_FIELDS, MAX_METHODS = 12, 8


def strip_comments(src):
    return re.sub(r'//[^\n]*', '', src)


def simp(t):
    """Упрощает запись типа: без пакетов и указателей."""
    t = t.strip()
    t = re.sub(r'\b(domain|time|io|context|config|slog|service)\.', '', t)
    t = t.replace('*', '')
    return re.sub(r'\s+', ' ', t)


def split_top(s):
    out, depth, cur = [], 0, ''
    for ch in s:
        if ch in '([{':
            depth += 1
        elif ch in ')]}':
            depth -= 1
        if ch == ',' and depth == 0:
            out.append(cur)
            cur = ''
        else:
            cur += ch
    if cur.strip():
        out.append(cur)
    return [x.strip() for x in out]


def signature(name, rest):
    """rest — текст после имени метода: '(args) ret'. Возвращает 'Name(a, b) : Ret'."""
    depth, i = 0, 0
    for i, ch in enumerate(rest):
        if ch == '(':
            depth += 1
        elif ch == ')':
            depth -= 1
            if depth == 0:
                break
    args, ret = rest[1:i], rest[i + 1:].strip().rstrip('{').strip()
    names = []
    for a in split_top(args):
        parts = a.split()
        if not parts or parts[0] in ('ctx', '_') or (len(parts) > 1 and 'Context' in parts[1]):
            continue
        names.append(parts[0].rstrip('.'))
    if ret.startswith('('):
        ret = ret[1:-1]
    rets = []
    for r in split_top(ret):
        r = r.split()[-1] if ' ' in r and not r.startswith(('map', 'func', 'chan')) else r
        if r and r != 'error':
            rets.append(simp(r))
    if len(names) > 4:
        names = names[:3] + ['…']
    return f"{name}({', '.join(names)})" + (f" : {', '.join(rets)}" if rets else '')


class Pkg:
    def __init__(self, *dirs):
        self.structs, self.ifaces, self.enums, self.methods = {}, {}, {}, {}
        for d in dirs:
            for f in sorted((ROOT / d).glob('*.go')):
                if f.name.endswith('_test.go'):
                    continue
                self.parse(strip_comments(f.read_text()))

    def parse(self, src):
        for m in re.finditer(r'^type (\w+) struct \{\n(.*?)^\}', src, re.S | re.M):
            fields = []
            for line in m.group(2).splitlines():
                p = line.split()
                if len(p) >= 2 and re.match(r'^[A-Za-z_]\w*,?$', p[0]):
                    fields.append((p[0].rstrip(','), simp(' '.join(p[1:]).split('`')[0])))
            self.structs[m.group(1)] = fields
        for m in re.finditer(r'^type (\w+) interface \{\n(.*?)^\}', src, re.S | re.M):
            body, meths, buf = m.group(2), [], ''
            for line in body.splitlines():
                buf += ' ' + line.strip()
                if not buf.strip():
                    buf = ''
                    continue
                if buf.count('(') == buf.count(')'):
                    mm = re.match(r'\s*(\w+)(\(.*)$', buf)
                    if mm:
                        meths.append(signature(mm.group(1), mm.group(2)))
                    buf = ''
            self.ifaces[m.group(1)] = meths
        for m in re.finditer(r'^type (\w+) (u?int\d*|string)\s*$', src, re.M):
            self.enums.setdefault(m.group(1), [])
        for block in re.finditer(r'^const \(\n(.*?)^\)', src, re.S | re.M):
            cur = None
            for line in block.group(1).splitlines():
                p = line.split()
                if not p:
                    continue
                if len(p) >= 3 and p[2] == '=' and p[1] in self.enums:
                    cur = p[1]
                    self.enums[cur].append(p[0])
                elif len(p) == 1 and cur:
                    self.enums[cur].append(p[0])
                else:
                    cur = None
        for m in re.finditer(r'^func \(\w+ \*?(\w+)\) ([A-Z]\w*)(\(.*?\)[^{\n]*)\{', src, re.S | re.M):
            self.methods.setdefault(m.group(1), []).append(signature(m.group(2), ' '.join(m.group(3).split())))


def render(name, title, pkg, types, opts='', extra=(), level='full', max_fields=MAX_FIELDS, max_methods=MAX_METHODS + 4):
    """types — список имён; префикс '~' — показать только имя (без членов).
    level='text' — рисунок для записки: у сущностей только поля, у сервисов только методы, пустые отсеки скрыты."""
    text = level == 'text'
    base = name.rsplit('/', 1)[-1]
    up = '../' * (name.count('/') + 1)
    L = [f'@startuml {base}', f'!include {up}uml/style.iuml', 'hide empty members' if text else '', 'skinparam classAttributeIconSize 0', 'skinparam classBackgroundColor white',
         'skinparam classFontStyle bold', 'skinparam linetype ortho' if 'ortho' in opts else '', 'hide circle',
         'left to right direction' if 'ltr' in opts else '', 'skinparam nodesep 30', 'skinparam ranksep 45']
    brief = {t[1:] for t in types if t.startswith('~')}
    names = [t.lstrip('~') for t in types]
    known = set(names)
    rel = []

    def refs(text, owner, arrow):
        for other in re.findall(r'\w+', text):
            if other in known and other != owner and (owner, other) not in {(a, b) for a, b, _ in rel}:
                rel.append((owner, other, arrow))

    for t in names:
        if t in pkg.enums and t not in pkg.structs:
            L.append(f'enum {t} <<перечисление>> {{')
            pref = t
            for c in pkg.enums[t]:
                if c.endswith('Unspecified') or c == '_':
                    continue
                L.append('  ' + (c[len(pref):] if c.startswith(pref) and len(c) > len(pref) else c))
            L.append('}')
        elif t in pkg.ifaces:
            L.append(f'interface {t} <<интерфейс>> {{')
            ms = [] if t in brief else pkg.ifaces[t]
            for s in ms[:MAX_METHODS]:
                L.append(f'  + {s}')
            if len(ms) > MAX_METHODS:
                L.append(f'  {{method}} … ещё {len(ms) - MAX_METHODS}')
            L.append('}')
            for s in pkg.ifaces[t]:
                refs(s, t, '..>')
        elif t in pkg.structs:
            L.append(f'class {t} {{')
            deps = []
            shown = 0
            for fname, ftype in pkg.structs[t]:
                hit = [o for o in re.findall(r'\w+', ftype) if o in known and o != t]
                if hit:
                    for o in hit:
                        if (o in pkg.ifaces or fname[0].islower()) and o not in pkg.enums:
                            deps.append((o, '-->'))
                        elif o in pkg.enums:
                            deps.append((o, '-->'))
                        else:
                            deps.append((o, '*--' if not ftype.startswith('map') else '-->'))
                if fname[0].islower() or t in brief:
                    continue  # зависимости сервиса показаны связями
                if shown < max_fields:
                    L.append(f'  + {fname} : {ftype}')
                shown += 1
            if shown > max_fields:
                L.append(f'  … ещё {shown - max_fields}')
            ms = [] if t in brief else pkg.methods.get(t, [])
            ms = [s for s in ms if not s.startswith(('With', 'String(', 'Set'))] or ms
            if text:
                ms = [] if shown else [s.split(' : ')[0] for s in ms]
            if ms and not (text and not shown):
                L.append('  --')
            for s in ms[:max_methods]:
                L.append(f'  + {s}')
            if len(ms) > max_methods:
                L.append(f'  {{method}} … ещё {len(ms) - max_methods}')
            L.append('}')
            for o, a in deps:
                if (t, o) not in {(x, y) for x, y, _ in rel}:
                    rel.append((t, o, a))
        else:
            raise SystemExit(f'{name}: тип {t} не найден')
    for a, b, arrow in rel:
        L.append(f'{a} {arrow} {b}')
    L.extend(extra)
    L.append('@enduml')
    out = pathlib.Path(name + '.puml')
    out.parent.mkdir(exist_ok=True)
    out.write_text('\n'.join(x for x in L if x) + '\n')
    print(name, len(names), 'типов')


orc = Pkg('orchestrator/internal/domain', 'orchestrator/internal/service')
pm = Pkg('project-manager/internal/domain', 'project-manager/internal/service')
md = Pkg('moderation/internal/domain', 'moderation/internal/service')
sso = Pkg('sso/internal/domain', 'sso/internal/service')
node = Pkg('game-server-node/internal/domain', 'game-server-node/internal/service')

SHEETS = [
    ('orchestrator-entities', 'Сущности сервиса оркестрации', orc,
     ['Node', 'Instance', 'ServerBuild', 'GamePolicy', 'ManagedService', 'ServiceBackup', 'PlatformGrant', 'QueueEntry',
      'ServerEndpoint', 'DiscoveryResult', 'NodeStatus', 'NodeRole', 'IngressMode', 'InstanceStatus', 'Protocol',
      'OrchestrationMode', 'ScaleBehavior', 'ServiceType', 'ServiceStatus', 'DiscoveryStatus'], '',
     ['Node "1" o-- "*" Instance : размещает', 'ServerBuild "1" <-- "*" Instance : запущен из', 'Node "1" *-- "*" ManagedService : размещает',
      'ManagedService "1" *-- "*" ServiceBackup : имеет', 'GamePolicy "1" ..> "*" Instance : управляет числом',
      'GamePolicy "1" ..> "*" QueueEntry : настраивает очередь', 'QueueEntry "*" --> "0..1" Instance : место закреплено на',
      'PlatformGrant "0..1" ..> "*" Instance : ограничивает', 'ServerEndpoint ..> Instance : адрес для клиента']),
    ('orchestrator-services', 'Сервисы и интерфейсы сервиса оркестрации', orc,
     ['NodeService', 'InstanceService', 'BuildPipeline', 'GamePolicyService', 'DiscoveryService', 'QueueService', 'HeartbeatService',
      'PlatformAccessService', 'NodeRepo', 'InstanceRepo', 'BuildStorage', 'GamePolicyRepo', 'QueueStore', 'QueueEventRepo',
      'InstanceStateStore', 'NodeStateStore', 'NodeClient', 'ManagedServiceRepo', 'BackupRepo', 'PlatformAccessRepo'], 'ltr'),
    ('projects-entities', 'Сущности сервиса управления проектами', pm,
     ['Project', 'Draft', 'DraftMeta', 'Build', 'Release', 'DeploymentRecord', 'Member', 'Invitation', 'UserBlock', 'SharedProject',
      'GameItem', 'ProjectStatus', 'InvitationStatus', 'DeploymentStatus'], '',
     ['Project "1" *-- "*" Build : содержит', 'Project "1" *-- "*" Release : содержит', 'Project "1" *-- "*" DeploymentRecord : журнал',
      'Project "1" *-- "*" Member : участники', 'Project "1" *-- "*" Invitation : приглашения', 'Project "1" *-- "*" GameItem : товары',
      'Release "*" --> "1" Build : публикует', 'Draft "*" --> "0..1" Build : активная сборка', 'Invitation ..> Member : при принятии', 'SharedProject --> Project']),
    ('projects-services', 'Сервис и интерфейсы сервиса управления проектами', pm,
     ['ProjectService', 'ProjectRepo', 'DraftRepo', 'BuildRepo', 'ReleaseRepo', 'DeploymentRepo', 'MemberRepo', 'InvitationRepo',
      'BlockRepo', 'BuildStorage', 'MediaStorage', 'Deployer', 'ModerationClient', 'AnalyticsClient', 'PurchaseClient', 'Locker'], 'ltr'),
    ('moderation', 'Классы сервиса модерации', md,
     ['ModerationService', 'ModerationRequest', 'ProjectSnapshot', 'GameItemSnapshot', 'ModerationSnapshot', 'SnapshotPayload',
      'ChatMessage', 'Attachment', 'ViolationItem', 'ModeratorStats', 'RequestStatus', 'RequestType', 'SenderRole', 'MessageType',
      'RequestRepo', 'MessageRepo', 'AttachmentRepo', 'SnapshotRepo', 'ProjectClient', 'OrchestratorClient'], '',
     ['ModerationRequest "1" *-- "*" ChatMessage : переписка', 'ModerationRequest "1" *-- "0..*" ModerationSnapshot : решения',
      'ModerationSnapshot ..> SnapshotPayload : содержит']),
    ('accounts', 'Классы сервиса учётных записей', sso,
     ['AuthService', 'UserService', 'TokenService', 'User', 'Session', 'TokenPair', 'Claims', 'UserRole', 'UserStatus',
      'UserRepository', 'SessionRepository', 'TokenManager', 'PasswordHasher', 'EmailSender', 'EmailVerificationStore',
      'PasswordResetStore', 'AttemptLimiter', 'SessionCache'], 'ltr',
     ['User "1" *-- "*" Session : сессии']),
    ('node-agent', 'Классы агента узла', node,
     ['DeploymentService', 'DiscoveryService', 'BackupManager', 'AnnouncementService', 'Instance', 'Node',
      'ServiceInfo', 'BackupInfo', 'ContainerOpts', 'ContainerInfo', 'ResourcesUsage', 'ResourcesMax', 'InstanceStatus', 'ServiceType',
      'BackupStatus', 'ContainerRuntime', 'InstanceStorage'], 'ltr'),
]
for name, title, pkg, types, opts, *extra in SHEETS:
    types = [t for t in types if t.lstrip('~') in pkg.structs or t.lstrip('~') in pkg.ifaces or t.lstrip('~') in pkg.enums or print('  нет типа', t)]
    render(name, title, pkg, types, opts, extra[0] if extra else ())


# --- Рисунки для текста записки: сущности без методов и перечислений, не более восьми полей ---
TEXT = [
    ('text/orchestrator-entities', orc,
     ['Node', 'Instance', 'ServerBuild', 'GamePolicy', 'QueueEntry', 'PlatformGrant', 'ManagedService', 'ServiceBackup'],
     ['Node "1" o-- "*" Instance : размещает', 'ServerBuild "1" <-- "*" Instance : запущен из', 'Node "1" *-- "*" ManagedService : размещает',
      'ManagedService "1" *-- "*" ServiceBackup : имеет', 'GamePolicy "1" ..> "*" Instance : управляет числом',
      'GamePolicy "1" ..> "*" QueueEntry : настраивает очередь', 'QueueEntry "*" --> "0..1" Instance : место закреплено на',
      'PlatformGrant "0..1" ..> "*" Instance : ограничивает']),
    ('text/projects-entities', pm,
     ['Project', 'Draft', 'Build', 'Release', 'Member', 'Invitation', 'DeploymentRecord', 'GameItem'],
     ['Project "1" *-- "*" Build : содержит', 'Project "1" *-- "*" DeploymentRecord : журнал',
      'Project "1" *-- "*" Member : участники', 'Project "1" *-- "*" Invitation : приглашения', 'Project "1" *-- "*" GameItem : товары',
      'Release "*" --> "1" Build : публикует', 'Draft "*" --> "0..1" Build : активная сборка', 'Invitation ..> Member : при принятии']),
    ('text/moderation-entities', md,
     ['ModerationRequest', 'ProjectSnapshot', 'ModerationSnapshot', 'ChatMessage', 'Attachment', 'ViolationItem'],
     ['ModerationRequest "1" *-- "*" ChatMessage : переписка', 'ModerationRequest "1" *-- "0..*" ModerationSnapshot : решения',
      'ModerationRequest "1" ..> "*" ViolationItem : нарушения при отказе']),
    ('text/accounts-entities', sso, ['User', 'Session', 'TokenPair', 'Claims'],
     ['User "1" *-- "*" Session : сессии', 'Session "1" ..> "1" TokenPair : выдаётся при входе', 'Claims ..> User', 'Claims ..> Session']),
]
for name, pkg, types, extra in TEXT:
    render(name, '', pkg, types, '', extra, level='text', max_fields=8)

# --- Рисунки для текста записки: один сервис и его зависимости (только имена) ---
SERVICES = [
    ('text/service-instance', orc, 'InstanceService'), ('text/service-queue', orc, 'QueueService'),
    ('text/service-discovery', orc, 'DiscoveryService'), ('text/service-project', pm, 'ProjectService'),
    ('text/service-moderation', md, 'ModerationService'), ('text/service-auth', sso, 'AuthService'),
]


def deps_of(pkg, svc):
    """Интерфейсы и сервисы, от которых зависит сервис, в порядке объявления полей."""
    out = []
    for fname, ftype in pkg.structs[svc]:
        for o in re.findall(r'\w+', ftype):
            if o[0].isupper() and o != svc and o not in out and (o in pkg.ifaces or o.endswith('Service')) and (o in pkg.ifaces or o in pkg.structs):
                out.append(o)
    return out


for name, pkg, svc in SERVICES:
    render(name, '', pkg, [svc] + ['~' + d for d in deps_of(pkg, svc)], 'ltr', level='text', max_methods=10)

# Таблица «сервис — зависимости» вместо листов с десятками пересекающихся стрелок
rows = ['| Сервис | Число операций | Используемые интерфейсы и сервисы |', '|---|---|---|']
for title, pkg, svcs in [('Оркестрация', orc, ['NodeService', 'InstanceService', 'BuildPipeline', 'GamePolicyService', 'DiscoveryService',
                                              'QueueService', 'HeartbeatService', 'PlatformAccessService']),
                         ('Управление проектами', pm, ['ProjectService']), ('Модерация', md, ['ModerationService']),
                         ('Учётные записи', sso, ['AuthService', 'UserService', 'TokenService']),
                         ('Агент узла', node, ['DeploymentService', 'DiscoveryService', 'BackupManager', 'AnnouncementService'])]:
    rows.append(f'| **{title}** | | |')
    for s in svcs:
        rows.append(f"| `{s}` | {len([m for m in pkg.methods.get(s, []) if not m.startswith(('With', 'String(', 'Set'))])} | {', '.join(f'`{d}`' for d in deps_of(pkg, s)) or '–'} |")
pathlib.Path('text/dependencies.md').write_text('\n'.join(rows) + '\n')

# --- Обзор для презентации: ключевые сущности всех сервисов, только имена, связи и кратности ---
HEAD = ['!include ../uml/style.iuml', 'skinparam classAttributeIconSize 0', 'skinparam classBackgroundColor white', 'hide circle',
        'hide empty members', 'skinparam defaultFontSize 16', 'skinparam packageStyle rectangle', 'skinparam packageFontStyle plain']
OVERVIEW = {
    'Учётные записи': [('User', 'Пользователь')],
    'Управление проектами': [('Project', 'Проект'), ('Draft', 'Черновик'), ('Build', 'Клиентская сборка'), ('Release', 'Релиз'),
                             ('Member', 'Участник проекта')],
    'Модерация': [('ModerationRequest', 'Заявка на модерацию'), ('ModerationSnapshot', 'Аудит-снимок')],
    'Оркестрация': [('ServerBuild', 'Серверная сборка игры'), ('GamePolicy', 'Политика оркестрации'), ('Instance', 'Экземпляр сервера'),
                    ('Node', 'Вычислительный узел'), ('QueueEntry', 'Место в очереди')],
}
OVERVIEW_REL = [
    'User "1" -- "*" Project : владеет', 'User "1" -- "*" Member : является', 'User "1" -- "*" Node : подключает',
    'Project *-- "1" Draft', 'Project *-- "*" Build', 'Project *-- "*" Release', 'Project *-- "*" Member',
    'Release "*" --> "1" Build : публикует',
    'ModerationRequest "*" --> "1" Project : проверяет', 'ModerationRequest *-- "*" ModerationSnapshot',
    'Project "1" -- "*" ServerBuild : серверные сборки', 'Project "1" -- "0..1" GamePolicy : политика',
    'ServerBuild "1" <-- "*" Instance : запущен из', 'Node "1" o-- "*" Instance : размещает',
    'GamePolicy "1" ..> "*" Instance : управляет числом', 'QueueEntry "*" --> "0..1" Instance : закреплено за',
]
L = ['@startuml overview'] + HEAD + ['skinparam nodesep 45', 'skinparam ranksep 55']
for pkg_title, items in OVERVIEW.items():
    L.append(f'package "{pkg_title}" {{')
    for cls, ru in items:
        L.append(f'  class "<b>{cls}</b>\\n{ru}" as {cls}')
    L.append('}')
pathlib.Path('overview.puml').write_text('\n'.join(L + OVERVIEW_REL + ['@enduml']) + '\n')

# --- Фрагмент крупно: ядро оркестрации с полями ---
render('overview-fragment', '', orc, ['Node', 'Instance', 'ServerBuild'], 'ltr',
       ['Node "1" o-- "*" Instance : размещает', 'ServerBuild "1" <-- "*" Instance : запущен из'], level='text', max_fields=10)
