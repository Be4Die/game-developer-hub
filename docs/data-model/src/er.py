"""Строит ER-диаграммы PlantUML из src/<сервис>/migrations/init.sql: ключевые поля, внешние ключи и до 6 основных полей на таблицу.

Запуск из docs/data-model: python3 src/er.py  — создаёт <сервис>.puml, затем plantuml -tpng -tsvg *.puml
"""
import re, pathlib

ROOT = pathlib.Path(__file__).resolve().parents[3] / 'src'
NAMES = {
    'nodes': 'Узлы', 'server_builds': 'Серверные сборки', 'instances': 'Экземпляры', 'game_policies': 'Политики оркестрации',
    'queue_events': 'События очереди', 'node_services': 'Сервисы данных узла', 'node_service_backups': 'Резервные копии',
    'platform_access_requests': 'Заявки на серверы платформы', 'platform_grants': 'Доступ к серверам платформы',
    'users': 'Пользователи', 'sessions': 'Сессии',
    'projects': 'Проекты', 'project_drafts': 'Черновики', 'project_builds': 'Клиентские сборки', 'project_releases': 'Релизы',
    'deployments': 'Развёртывания', 'project_members': 'Участники проектов', 'project_invitations': 'Приглашения',
    'user_access_blocks': 'Блокировки приглашений',
    'moderation_requests': 'Заявки на модерацию', 'moderation_messages': 'Сообщения чата', 'moderation_attachments': 'Вложения',
    'moderation_snapshots': 'Аудит-снимки',
}
TYPES = {'BIGSERIAL': 'bigint', 'BIGINT': 'bigint', 'SERIAL': 'int', 'INTEGER': 'int', 'INT': 'int', 'SMALLINT': 'smallint', 'TEXT': 'text',
         'TIMESTAMP': 'timestamp', 'TIMESTAMPTZ': 'timestamp', 'JSONB': 'jsonb', 'BOOLEAN': 'bool', 'UUID': 'uuid', 'BYTEA': 'bytea',
         'DOUBLE': 'float', 'REAL': 'float', 'NUMERIC': 'numeric'}
SKIP = {'created_at', 'updated_at'}
# Ссылки по идентификатору на таблицы других сервисов: внешнего ключа нет, потому что у каждого сервиса свой набор таблиц
# и сервисы не обращаются к чужим таблицам. На диаграмме такие связи показаны пунктиром, чужая таблица — пунктирной рамкой.
LOGICAL = {
    'projects': {'owner_id': 'users'}, 'project_members': {'user_id': 'users'}, 'project_invitations': {'invitee_id': 'users'},
    'user_access_blocks': {'user_id': 'users'},
    'moderation_requests': {'project_id': 'projects'}, 'moderation_messages': {'project_id': 'projects'},
    'moderation_attachments': {'project_id': 'projects'},
    'nodes': {'owner_id': 'users'}, 'server_builds': {'game_id': 'projects'}, 'game_policies': {'game_id': 'projects'},
    'queue_events': {'game_id': 'projects', 'instance_id': 'instances'}, 'platform_access_requests': {'project_id': 'projects'},
    'platform_grants': {'game_id': 'projects'},
}
OWNER = {'users': 'сервис учётных записей', 'projects': 'сервис управления проектами', 'instances': 'сервис оркестрации'}
MAX_OTHER = 6


def parse(sql):
    sql = re.sub(r'--[^\n]*', '', sql)
    tables = {}
    for m in re.finditer(r'CREATE TABLE(?: IF NOT EXISTS)?\s+(\w+)\s*\((.*?)\n\);', sql, re.S):
        name, body = m.groups()
        cols, fks, pks = [], {}, set()
        for line in re.split(r',\s*\n', body):
            line = ' '.join(line.split())
            if not line:
                continue
            mt = re.match(r'PRIMARY KEY \(([^)]*)\)', line, re.I)
            if mt:
                pks |= {c.strip() for c in mt.group(1).split(',')}
                continue
            mt = re.match(r'FOREIGN KEY \((\w+)\) REFERENCES (\w+)', line, re.I)
            if mt:
                fks[mt.group(1)] = mt.group(2)
                continue
            if re.match(r'(UNIQUE|CONSTRAINT|CHECK)', line, re.I):
                continue
            mc = re.match(r'(\w+)\s+([A-Za-z]+)', line)
            if not mc:
                continue
            col, typ = mc.groups()
            if re.search(r'PRIMARY KEY', line, re.I):
                pks.add(col)
            mr = re.search(r'REFERENCES (\w+)', line, re.I)
            if mr:
                fks[col] = mr.group(1)
            cols.append((col, TYPES.get(typ.upper(), typ.lower())))
        tables[name] = dict(cols=cols, fks=fks, pks=pks)
    return tables


def render(svc, tables, out, ltr=False):
    lines = [f'@startuml {out}', '!include ../uml/style.iuml', 'hide circle', 'skinparam linetype ortho', 'skinparam classBackgroundColor white']
    if ltr:
        lines.append('left to right direction')
    for t, d in tables.items():
        key = [(c, ty) for c, ty in d['cols'] if c in d['pks']]
        fk = [(c, ty) for c, ty in d['cols'] if c in d['fks'] and c not in d['pks']]
        other = [(c, ty) for c, ty in d['cols'] if c not in d['pks'] and c not in d['fks'] and c not in SKIP]
        lines.append(f'entity "{NAMES.get(t, t)}\\n({t})" as {t} {{')
        for c, ty in key:
            lines.append(f'  * {c} : {ty}')
        for c, ty in fk:
            lines.append(f'  + {c} : {ty}')
        lines.append('  --')
        for c, ty in other[:MAX_OTHER]:
            lines.append(f'  {c} : {ty}')
        if len(other) > MAX_OTHER:
            lines.append(f'  … ещё {len(other) - MAX_OTHER}')
        lines.append('}')
    for t, d in tables.items():
        for col, parent in sorted(d['fks'].items()):
            if parent in tables:
                # внешний ключ, он же первичный, даёт связь «один к одному»
                lines.append(f'{parent} ||--o| {t}' if col in d['pks'] else f'{parent} ||--o{{ {t}')
    # Таблицы других сервисов, на которые ссылаются по идентификатору
    ext = sorted({p for t in tables for p in LOGICAL.get(t, {}).values() if p not in tables})
    for p in ext:
        lines.append(f'entity "{NAMES.get(p, p)}\\n({p})\\n<i>{OWNER[p]}</i>" as {p} #line.dashed {{')
        lines.append('  * id')
        lines.append('}')
    for t in tables:
        for col, parent in sorted(LOGICAL.get(t, {}).items()):
            one = col in tables[t]['pks'] or (t, col) in {('platform_grants', 'game_id')}
            lines.append(f'{parent} ||..o| {t} : {col}' if one else f'{parent} ||..o{{ {t} : {col}')
    lines.append('@enduml')
    pathlib.Path(out + '.puml').write_text('\n'.join(lines) + '\n')


GROUPS = [
    ('orchestrator', 'orchestrator', ['nodes', 'server_builds', 'instances', 'node_services', 'node_service_backups'], True),
    ('orchestrator', 'orchestrator-policies', ['game_policies', 'queue_events', 'platform_access_requests', 'platform_grants'], False),
    ('project-manager', 'projects', None, False),
    ('moderation', 'moderation', None, True),
    ('sso', 'accounts', None, True),
]
for svc, out, names, ltr in GROUPS:
    t = parse((ROOT / svc / 'migrations' / 'init.sql').read_text())
    if names:
        t = {k: t[k] for k in names}
    print(out, len(t), 'таблиц')
    render(svc, t, out, ltr)
