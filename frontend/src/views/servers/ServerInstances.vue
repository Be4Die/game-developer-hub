<template>
  <div class="instances-page tab-fade-in">
    <div class="page-header">
      <h1>Инстансы <span class="counter">{{ instances.length }}/4</span></h1>
      <div class="header-actions">
        <select v-model="statusFilter" class="filter-select">
          <option value="all">Все статусы</option>
          <option value="starting">Запускается</option>
          <option value="running">Работает</option>
          <option value="stopping">Останавливается</option>
          <option value="stopped">Остановлен</option>
          <option value="crashed">Авария</option>
        </select>
        <button class="btn-primary" @click="showStartForm = true">
          <Play class="icon-sm" /> Запустить инстанс
        </button>
      </div>
    </div>

    <!-- Таблица инстансов -->
    <div class="table-wrap" v-if="filteredInstances.length">
      <table class="data-table">
        <thead>
          <tr>
            <th>Имя</th>
            <th>Версия</th>
            <th>Статус</th>
            <th>Нода</th>
            <th>Игроки</th>
            <th>Адрес</th>
            <th>Запущен</th>
            <th></th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="inst in filteredInstances" :key="inst.id"
              @click="$router.push(`/projects/${gameId}/servers/instances/${inst.id}`)"
              class="clickable-row">
            <td class="cell-name">{{ inst.name || `Инстанс #${inst.id}` }}</td>
            <td><code>{{ inst.build_version }}</code></td>
            <td><StatusBadge :status="inst.status" type="instance" /></td>
            <td class="cell-muted">{{ inst.node_id }}</td>
            <td>{{ inst.player_count ?? 0 }} / {{ inst.max_players }}</td>
            <td class="cell-muted">{{ inst.server_address }}:{{ inst.host_port }}</td>
            <td class="cell-muted">{{ inst.started_at ? formatDate(inst.started_at) : '—' }}</td>
            <td class="cell-actions" @click.stop>
              <button v-if="inst.status === 'running'" class="btn-stop" @click="stopInstance(inst)" title="Остановить">
                <Square class="icon-sm" />
              </button>
              <button v-else-if="inst.status === 'stopped' || inst.status === 'crashed'" class="btn-start" @click="startExisting(inst)" title="Запустить">
                <Play class="icon-sm" />
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
    <div v-else class="empty-state">Нет инстансов{{ statusFilter !== 'all' ? ' с выбранным статусом' : '' }}</div>

    <!-- Модал запуска нового инстанса -->
    <div v-if="showStartForm" class="modal-overlay" @click.self="showStartForm = false">
      <div class="modal card">
        <h3>Запустить инстанс</h3>
        <div class="form-grid">
          <div class="form-group">
            <label>Версия билда *</label>
            <select v-model="startForm.build_version" class="form-input">
              <option value="" disabled>Выберите билд</option>
              <option v-for="b in builds" :key="b.id" :value="b.build_version">{{ b.build_version }}</option>
            </select>
          </div>
          <div class="form-group">
            <label>Режим работы *</label>
            <div class="radio-group">
              <label class="radio-label">
                <input type="radio" v-model="startForm.server_mode" value="manual" /> Ручной
              </label>
              <label class="radio-label">
                <input type="radio" v-model="startForm.server_mode" value="auto" /> Автоматический
              </label>
            </div>
          </div>
          <div class="form-group">
            <label>Имя (опционально)</label>
            <input type="text" v-model="startForm.name" class="form-input" placeholder="EU-1" maxlength="128" />
          </div>
          <div class="form-group">
            <label>Макс. игроков</label>
            <input type="number" v-model.number="startForm.max_players" class="form-input" min="1" placeholder="Из билда" />
          </div>
          <div class="form-group form-group-wide">
            <label>Переменные окружения</label>
            <KeyValueEditor v-model="startForm.env_vars" />
          </div>
          <div class="form-group form-group-wide">
            <label>Аргументы командной строки</label>
            <div class="args-list">
              <div v-for="(arg, i) in startForm.args" :key="i" class="arg-row">
                <input type="text" v-model="startForm.args[i]" class="form-input" placeholder="--flag value" />
                <button class="arg-remove" @click="startForm.args.splice(i, 1)">&times;</button>
              </div>
              <button class="arg-add" @click="startForm.args.push('')">+ Добавить аргумент</button>
            </div>
          </div>
        </div>
        <div class="modal-actions">
          <button class="btn-primary" @click="submitStart" :disabled="!startForm.build_version || !startForm.server_mode">
            Запустить
          </button>
          <button class="btn-outline" @click="showStartForm = false">Отмена</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, reactive } from 'vue'
import { Play, Square } from 'lucide-vue-next'
import StatusBadge from '../../components/orchestrator/StatusBadge.vue'
import KeyValueEditor from '../../components/orchestrator/KeyValueEditor.vue'
import { mockBuilds, mockInstances } from '../../data/mock-orchestrator'
import { showToast } from '../../store'

const props = defineProps({ gameId: { type: [String, Number], required: true } })

const instances = ref([...mockInstances])
const builds = computed(() => mockBuilds)
const statusFilter = ref('all')
const showStartForm = ref(false)

const startForm = reactive({
  build_version: '',
  server_mode: 'manual',
  name: '',
  max_players: null,
  env_vars: {},
  args: [],
})

const filteredInstances = computed(() => {
  if (statusFilter.value === 'all') return instances.value
  return instances.value.filter(i => i.status === statusFilter.value)
})

function submitStart() {
  const newInst = {
    id: Date.now(),
    game_id: Number(props.gameId),
    node_id: 1,
    build_version: startForm.build_version,
    name: startForm.name || `Инстанс #${Date.now()}`,
    protocol: 'websocket',
    host_port: 30000 + Math.floor(Math.random() * 1000),
    internal_port: 8080,
    status: 'starting',
    player_count: 0,
    max_players: startForm.max_players || 16,
    developer_payload: {},
    server_address: '192.168.1.100',
    started_at: new Date().toISOString(),
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
  }
  instances.value.unshift(newInst)
  showStartForm.value = false
  Object.assign(startForm, { build_version: '', server_mode: 'manual', name: '', max_players: null, env_vars: {}, args: [] })
  showToast('Инстанс запускается...')
  // Имитация перехода в running
  setTimeout(() => {
    const inst = instances.value.find(i => i.id === newInst.id)
    if (inst) { inst.status = 'running'; inst.started_at = new Date().toISOString() }
  }, 2000)
}

function stopInstance(inst) {
  inst.status = 'stopping'
  showToast(`Инстанс ${inst.name} останавливается...`)
  setTimeout(() => { inst.status = 'stopped'; inst.player_count = 0 }, 1500)
}

function startExisting(inst) {
  inst.status = 'starting'
  showToast(`Инстанс ${inst.name} запускается...`)
  setTimeout(() => { inst.status = 'running'; inst.started_at = new Date().toISOString() }, 2000)
}

function formatDate(ts) {
  return new Date(ts).toLocaleDateString('ru-RU', { day: 'numeric', month: 'short', hour: '2-digit', minute: '2-digit' })
}
</script>

<style scoped>
.tab-fade-in { animation: fadeIn 0.3s ease; }
@keyframes fadeIn { from { opacity: 0; transform: translateY(10px); } to { opacity: 1; transform: translateY(0); } }

.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 24px; }
.page-header h1 { margin: 0; }
.counter { font-size: 0.9rem; font-weight: 400; color: var(--text-muted); }
.header-actions { display: flex; gap: 12px; align-items: center; }
.filter-select {
  padding: 8px 12px; border: 1px solid var(--border); border-radius: var(--radius-sm);
  background: var(--bg-input); color: var(--text-main); font-size: 0.88rem;
}

.table-wrap { background: var(--bg-card); border: 1px solid var(--border); border-radius: var(--radius-lg); overflow: hidden; }
.data-table { width: 100%; border-collapse: collapse; }
.data-table th {
  text-align: left; padding: 12px 16px; font-size: 0.78rem; font-weight: 600;
  color: var(--text-muted); text-transform: uppercase; letter-spacing: 0.03em;
  background: var(--bg-secondary); border-bottom: 1px solid var(--border);
}
.data-table td { padding: 12px 16px; font-size: 0.88rem; border-bottom: 1px solid var(--border); }
.data-table tr:last-child td { border-bottom: none; }
.clickable-row { cursor: pointer; transition: 0.1s; }
.clickable-row:hover { background: var(--bg-hover); }
.cell-name { font-weight: 600; }
.cell-muted { color: var(--text-muted); }
.cell-actions { display: flex; gap: 4px; }
code { background: var(--bg-secondary); padding: 2px 6px; border-radius: 4px; font-size: 0.82rem; }

.btn-stop, .btn-start {
  background: none; border: 1px solid var(--border); padding: 4px 8px;
  border-radius: var(--radius-sm); cursor: pointer; display: flex; align-items: center; color: var(--text-muted);
}
.btn-stop:hover { color: var(--danger); border-color: var(--danger); }
.btn-start:hover { color: var(--success); border-color: var(--success); }
.empty-state { padding: 40px; text-align: center; color: var(--text-muted); }

.modal-overlay { position: fixed; inset: 0; background: rgba(0,0,0,0.4); z-index: 100; display: flex; align-items: center; justify-content: center; }
.modal { max-width: 560px; width: 90%; max-height: 90vh; overflow-y: auto; }
.modal h3 { margin: 0 0 16px; }
.form-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; }
.form-group { display: flex; flex-direction: column; gap: 4px; }
.form-group-wide { grid-column: 1 / -1; }
.form-group label { font-size: 0.82rem; font-weight: 600; color: var(--text-muted); }
.form-input {
  padding: 8px 12px; border: 1px solid var(--border); border-radius: var(--radius-sm);
  background: var(--bg-input); color: var(--text-main); font-size: 0.88rem;
}
.radio-group { display: flex; gap: 16px; padding-top: 4px; }
.radio-label { display: flex; align-items: center; gap: 6px; font-size: 0.88rem; cursor: pointer; }

.args-list { display: flex; flex-direction: column; gap: 8px; }
.arg-row { display: flex; gap: 8px; }
.arg-row .form-input { flex: 1; }
.arg-remove { background: none; border: none; color: var(--danger); font-size: 1.2rem; cursor: pointer; padding: 0 4px; }
.arg-add { background: none; border: 1px dashed var(--border); color: var(--primary); padding: 4px 12px; border-radius: var(--radius-sm); cursor: pointer; font-size: 0.85rem; }

.modal-actions { display: flex; gap: 8px; margin-top: 20px; }
</style>
