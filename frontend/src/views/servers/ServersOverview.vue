<template>
  <div class="overview tab-fade-in">
    <div class="overview-header">
      <h1>Игровые сервера</h1>
    </div>

    <!-- Ошибка загрузки -->
    <div v-if="error" class="error-banner">
      <AlertCircle class="icon-sm" /> Не удалось загрузить данные: {{ error }}
      <button class="btn-outline btn-sm" @click="fetchAll">Повторить</button>
    </div>

    <!-- Карточки-сводки -->
    <div class="summary-grid">
      <div class="summary-card">
        <span class="summary-label">Билды</span>
        <span class="summary-value">{{ loading ? '...' : builds.length }}</span>
      </div>
      <div class="summary-card">
        <span class="summary-label">Работающие</span>
        <span class="summary-value">{{ loading ? '...' : runningCount }}<span class="summary-sub"> / {{ instances.length }}</span></span>
      </div>
      <div class="summary-card">
        <span class="summary-label">Игроков онлайн</span>
        <span class="summary-value">{{ loading ? '...' : totalPlayers }}</span>
      </div>
      <div class="summary-card">
        <span class="summary-label">Ноды</span>
        <span class="summary-value">{{ loading ? '...' : onlineNodes }}<span class="summary-sub"> / {{ nodes.length }}</span></span>
      </div>
    </div>

    <!-- Политика оркестрации -->
    <div class="policy-section">
      <div class="policy-header" @click="policyExpanded = !policyExpanded">
        <div class="policy-title">
          <Settings2 class="icon-sm" />
          <strong>Политика оркестрации</strong>
        </div>
        <ChevronDown v-if="policyExpanded" class="icon-sm" />
        <ChevronRight v-else class="icon-sm" />
      </div>

      <div v-if="policyExpanded" class="policy-body">
        <div v-if="policyLoading" class="policy-loading">Загрузка…</div>

        <div v-else-if="!policyEditing && policy" class="policy-read">
          <div class="policy-grid">
            <div class="policy-item">
              <span class="policy-label">Режим</span>
              <span class="policy-value">{{ modeLabels[policy.mode] || policy.mode }}</span>
            </div>
            <div class="policy-item">
              <span class="policy-label">Целевое число инстансов</span>
              <span class="policy-value">{{ policy.target_instances }}</span>
            </div>
            <div class="policy-item">
              <span class="policy-label">Авторестарт при падении</span>
              <span class="policy-value">{{ policy.auto_restart ? 'Включён' : 'Отключён' }}</span>
            </div>
            <div class="policy-item">
              <span class="policy-label">Таймаут простоя (мин)</span>
              <span class="policy-value">{{ policy.scale_to_zero_timeout }}</span>
            </div>
            <div class="policy-item">
              <span class="policy-label">Версия билда по умолчанию</span>
              <span class="policy-value">{{ policy.default_build_version }}</span>
            </div>
            <div class="policy-item">
              <span class="policy-label">Макс. игроков / инстанс</span>
              <span class="policy-value">{{ policy.max_players_per_instance }}</span>
            </div>
            <div class="policy-item">
              <span class="policy-label">Макс. инстансов на игру</span>
              <span class="policy-value">{{ policy.max_instances_per_game }}</span>
            </div>
            <div class="policy-item">
              <span class="policy-label">При переполнении</span>
              <span class="policy-value">{{ behaviorLabels[policy.scale_behavior] || policy.scale_behavior }}</span>
            </div>
            <div class="policy-item">
              <span class="policy-label">Нода</span>
              <span class="policy-value">{{ policy.node_preference }}</span>
            </div>
          </div>
          <div class="policy-read-actions">
            <button class="btn-outline btn-sm" @click="startEdit">
              Изменить
            </button>
          </div>
        </div>

        <div v-else-if="policyEditing" class="policy-edit">
          <div class="policy-form">
            <label>
              Режим оркестрации
              <select v-model="policyDraft.mode">
                <option value="ORCHESTRATION_MODE_DISABLED">Только ручное управление</option>
                <option value="ORCHESTRATION_MODE_KEEP_ALIVE">Держать запущенным</option>
                <option value="ORCHESTRATION_MODE_SCALE_TO_ZERO">Экономичный (scale-to-zero)</option>
              </select>
            </label>
            <label>
              Целевое число инстансов
              <input v-model.number="policyDraft.target_instances" type="number" min="0" />
            </label>
            <label class="checkbox">
              <input v-model="policyDraft.auto_restart" type="checkbox" />
              Авторестарт при падении
            </label>
            <label>
              Таймаут простоя (мин)
              <input v-model.number="policyDraft.scale_to_zero_timeout" type="number" min="1" />
            </label>
            <label>
              Версия билда по умолчанию
              <input v-model="policyDraft.default_build_version" type="text" />
            </label>
            <label>
              Макс. игроков / инстанс
              <input v-model.number="policyDraft.max_players_per_instance" type="number" min="1" />
            </label>
            <label>
              Макс. инстансов на игру
              <input v-model.number="policyDraft.max_instances_per_game" type="number" min="1" />
            </label>
            <label>
              При переполнении
              <select v-model="policyDraft.scale_behavior">
                <option value="SCALE_BEHAVIOR_SPAWN">Запускать новый инстанс</option>
                <option value="SCALE_BEHAVIOR_QUEUE">Очередь игроков</option>
              </select>
            </label>
            <label>
              Нода
              <input v-model="policyDraft.node_preference" type="text" />
            </label>
          </div>
          <div class="policy-actions">
            <button class="btn-primary" @click="savePolicy" :disabled="policyLoading">
              <Save class="icon-sm" /> Сохранить
            </button>
            <button class="btn-outline" @click="cancelEdit">Отмена</button>
          </div>
        </div>
      </div>
    </div>

    <!-- Быстрые действия -->
    <div class="quick-actions">
      <router-link :to="`/projects/${gameId}/servers/builds`" class="action-card">
        <Upload class="action-icon" />
        <div>
          <strong>Загрузить билд</strong>
          <p>Загрузить новую версию серверного билда</p>
        </div>
      </router-link>
      <router-link :to="`/projects/${gameId}/servers/instances`" class="action-card">
        <Play class="action-icon" />
        <div>
          <strong>Запустить инстанс</strong>
          <p>Развернуть новый экземпляр сервера</p>
        </div>
      </router-link>
    </div>

    <!-- Последние инстансы -->
    <div class="recent-section">
      <div class="section-header">
        <h2>Последние инстансы</h2>
        <router-link :to="`/projects/${gameId}/servers/instances`" class="link">Все инстансы →</router-link>
      </div>
      <div class="recent-table-wrap">
        <table class="recent-table" v-if="instances.length">
          <thead>
            <tr>
              <th>Имя</th>
              <th>Версия</th>
              <th>Статус</th>
              <th>Игроки</th>
              <th>Запущен</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="inst in instances.slice(0, 5)" :key="inst.id"
                @click="$router.push(`/projects/${gameId}/servers/instances/${inst.id}`)"
                class="clickable-row">
              <td class="cell-name">{{ inst.name || `Инстанс #${inst.id}` }}</td>
              <td><code>{{ inst.build_version }}</code></td>
              <td><StatusBadge :status="inst.status" type="instance" /></td>
              <td>{{ inst.player_count ?? 0 }} / {{ inst.max_players }}</td>
              <td class="cell-muted">{{ inst.started_at ? formatDate(inst.started_at) : '—' }}</td>
            </tr>
          </tbody>
        </table>
        <div v-else-if="!loading" class="empty-state">Нет запущенных инстансов</div>
      </div>
    </div>
  </div>
</template>

    <script setup>
import { ref, computed, onMounted } from 'vue'
import { Upload, Play, AlertCircle, ChevronDown, ChevronRight, Save, Settings2 } from 'lucide-vue-next'
import StatusBadge from '../../components/orchestrator/StatusBadge.vue'
import { listBuilds, listInstances, listNodes, getPolicy, setPolicy } from '../../api/orchestrator'

const props = defineProps({ gameId: { type: [String, Number], required: true } })

const builds = ref([])
const instances = ref([])
const nodes = ref([])
const loading = ref(true)
const error = ref(null)

// Политика оркестрации
const policy = ref(null)
const policyLoading = ref(false)
const policyExpanded = ref(true)
const policyEditing = ref(false)
const policyDraft = ref({})

const runningCount = computed(() => instances.value.filter(i => i.status === 'running').length)
const totalPlayers = computed(() => instances.value.reduce((sum, i) => sum + (i.player_count ?? 0), 0))
const onlineNodes = computed(() => nodes.value.filter(n => n.status === 'online').length)

const modeLabels = {
  ORCHESTRATION_MODE_UNSPECIFIED: 'Не задан',
  ORCHESTRATION_MODE_DISABLED: 'Только ручное управление',
  ORCHESTRATION_MODE_KEEP_ALIVE: 'Держать запущенным',
  ORCHESTRATION_MODE_SCALE_TO_ZERO: 'Экономичный (scale-to-zero)',
}

const behaviorLabels = {
  SCALE_BEHAVIOR_UNSPECIFIED: 'Не задан',
  SCALE_BEHAVIOR_SPAWN: 'Запускать новый инстанс',
  SCALE_BEHAVIOR_QUEUE: 'Очередь игроков',
}

function formatDate(ts) {
  return new Date(ts).toLocaleDateString('ru-RU', { day: 'numeric', month: 'short', hour: '2-digit', minute: '2-digit' })
}

async function fetchAll() {
  loading.value = true
  error.value = null
  try {
    const [b, i, n] = await Promise.all([
      listBuilds(props.gameId).catch(() => []),
      listInstances(props.gameId).catch(() => []),
      listNodes().catch(() => []),
    ])
    builds.value = b
    instances.value = i
    nodes.value = n
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

async function loadPolicy() {
  policyLoading.value = true
  try {
    const p = await getPolicy(props.gameId)
    policy.value = p
  } catch (e) {
    // Политика может отсутствовать — это нормально, будет default.
    policy.value = {
      mode: 'ORCHESTRATION_MODE_DISABLED',
      target_instances: 1,
      auto_restart: false,
      scale_to_zero_timeout: 10,
      default_build_version: 'latest',
      max_players_per_instance: 100,
      max_instances_per_game: 1,
      scale_behavior: 'SCALE_BEHAVIOR_SPAWN',
      node_preference: 'auto',
    }
  } finally {
    policyLoading.value = false
  }
}

function startEdit() {
  policyDraft.value = { ...policy.value }
  policyEditing.value = true
}

function cancelEdit() {
  policyEditing.value = false
}

async function savePolicy() {
  policyLoading.value = true
  try {
    const payload = {
      mode: policyDraft.value.mode,
      target_instances: Number(policyDraft.value.target_instances),
      auto_restart: Boolean(policyDraft.value.auto_restart),
      scale_to_zero_timeout: Number(policyDraft.value.scale_to_zero_timeout),
      default_build_version: policyDraft.value.default_build_version,
      max_players_per_instance: Number(policyDraft.value.max_players_per_instance),
      max_instances_per_game: Number(policyDraft.value.max_instances_per_game),
      scale_behavior: policyDraft.value.scale_behavior,
      node_preference: policyDraft.value.node_preference,
    }
    const p = await setPolicy(props.gameId, payload)
    policy.value = p
    policyEditing.value = false
  } catch (e) {
    alert('Не удалось сохранить политику: ' + e.message)
  } finally {
    policyLoading.value = false
  }
}

onMounted(() => {
  fetchAll()
  loadPolicy()
})
</script>

<style scoped>
.tab-fade-in { animation: fadeIn 0.3s ease; }
@keyframes fadeIn { from { opacity: 0; transform: translateY(10px); } to { opacity: 1; transform: translateY(0); } }

.overview-header { margin-bottom: 24px; }
.overview-header h1 { margin: 0; }

.error-banner {
  display: flex; align-items: center; gap: 8px;
  padding: 12px 16px; background: var(--danger-light); color: var(--danger);
  border-radius: var(--radius-md); margin-bottom: 16px; font-size: 0.88rem;
}
.btn-sm { padding: 4px 12px; font-size: 0.82rem; }

.summary-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
  margin-bottom: 24px;
}
.summary-card {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: var(--radius-lg);
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.summary-label { font-size: 0.82rem; color: var(--text-muted); font-weight: 500; }
.summary-value { font-size: 1.8rem; font-weight: 800; color: var(--text-main); }
.summary-sub { font-size: 1rem; font-weight: 400; color: var(--text-muted); }

.quick-actions {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
  margin-bottom: 32px;
}
.action-card {
  display: flex; align-items: center; gap: 16px;
  padding: 20px; background: var(--bg-card); border: 1px solid var(--border);
  border-radius: var(--radius-lg); text-decoration: none; color: var(--text-main); transition: 0.15s;
}
.action-card:hover { border-color: var(--primary); box-shadow: 0 0 0 1px var(--primary); }
.action-icon { width: 36px; height: 36px; color: var(--primary); flex-shrink: 0; }
.action-card strong { font-size: 0.95rem; }
.action-card p { margin: 4px 0 0; font-size: 0.82rem; color: var(--text-muted); }

.recent-section { }
.section-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 12px; }
.section-header h2 { margin: 0; font-size: 1.1rem; }
.link { color: var(--primary); font-size: 0.85rem; font-weight: 500; text-decoration: none; }
.link:hover { text-decoration: underline; }

.recent-table-wrap { background: var(--bg-card); border: 1px solid var(--border); border-radius: var(--radius-lg); overflow: hidden; }
.recent-table { width: 100%; border-collapse: collapse; }
.recent-table th {
  text-align: left; padding: 12px 16px; font-size: 0.78rem; font-weight: 600;
  color: var(--text-muted); text-transform: uppercase; letter-spacing: 0.03em;
  background: var(--bg-secondary); border-bottom: 1px solid var(--border);
}
.recent-table td { padding: 12px 16px; font-size: 0.88rem; border-bottom: 1px solid var(--border); }
.recent-table tr:last-child td { border-bottom: none; }
.clickable-row { cursor: pointer; transition: 0.1s; }
.clickable-row:hover { background: var(--bg-hover); }
.cell-name { font-weight: 600; }
.cell-muted { color: var(--text-muted); }
code { background: var(--bg-secondary); padding: 2px 6px; border-radius: 4px; font-size: 0.82rem; }
.empty-state { padding: 40px; text-align: center; color: var(--text-muted); }

@media (max-width: 768px) {
  .summary-grid { grid-template-columns: repeat(2, 1fr); }
  .quick-actions { grid-template-columns: 1fr; }
}

/* Политика оркестрации */
.policy-section {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: var(--radius-lg);
  margin-bottom: 24px;
  overflow: hidden;
}
.policy-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 20px;
  cursor: pointer;
  user-select: none;
  transition: background 0.15s;
}
.policy-header:hover { background: var(--bg-hover); }
.policy-title {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 0.95rem;
}
.policy-body {
  padding: 0 20px 20px;
  border-top: 1px solid var(--border);
}
.policy-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 16px;
  padding-top: 16px;
}
.policy-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.policy-label {
  font-size: 0.78rem;
  color: var(--text-muted);
  font-weight: 500;
}
.policy-value {
  font-size: 0.9rem;
  color: var(--text-main);
  font-weight: 600;
}
.policy-form {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
  padding-top: 16px;
}
.policy-form label {
  display: flex;
  flex-direction: column;
  gap: 6px;
  font-size: 0.85rem;
  color: var(--text-main);
  font-weight: 500;
}
.policy-form input,
.policy-form select {
  padding: 8px 12px;
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  background: var(--bg-secondary);
  color: var(--text-main);
  font-size: 0.88rem;
}
.policy-form .checkbox {
  flex-direction: row;
  align-items: center;
  gap: 8px;
}
.policy-actions {
  display: flex;
  gap: 12px;
  margin-top: 16px;
}
.policy-loading {
  padding: 20px;
  text-align: center;
  color: var(--text-muted);
}

.policy-read-actions {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
}

/* Скрываем стрелки у числовых полей */
.policy-form input[type="number"]::-webkit-outer-spin-button,
.policy-form input[type="number"]::-webkit-inner-spin-button {
  -webkit-appearance: none;
  margin: 0;
}
.policy-form input[type="number"] {
  -moz-appearance: textfield;
  appearance: textfield;
}

/* Кастомный чекбокс */
.policy-form .checkbox {
  flex-direction: row;
  align-items: center;
  gap: 10px;
  cursor: pointer;
}
.policy-form .checkbox input[type="checkbox"] {
  appearance: none;
  -webkit-appearance: none;
  width: 18px;
  height: 18px;
  border: 1.5px solid var(--border);
  border-radius: 4px;
  background: var(--bg-secondary);
  cursor: pointer;
  position: relative;
  transition: 0.15s;
  flex-shrink: 0;
}
.policy-form .checkbox input[type="checkbox"]:checked {
  background: var(--primary);
  border-color: var(--primary);
}
.policy-form .checkbox input[type="checkbox"]:checked::after {
  content: "";
  position: absolute;
  left: 5px;
  top: 1px;
  width: 5px;
  height: 10px;
  border: solid #fff;
  border-width: 0 2px 2px 0;
  transform: rotate(45deg);
}
.policy-form .checkbox input[type="checkbox"]:focus {
  outline: none;
  box-shadow: 0 0 0 2px var(--primary-light);
}

/* Кнопка Изменить */
.btn-outline {
  padding: 6px 14px;
  font-size: 0.85rem;
  font-weight: 500;
  color: var(--text-main);
  background: transparent;
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  cursor: pointer;
  transition: 0.15s;
}
.btn-outline:hover {
  border-color: var(--primary);
  color: var(--primary);
  background: var(--primary-light);
}
.btn-sm {
  padding: 4px 12px;
  font-size: 0.82rem;
}

@media (max-width: 768px) {
  .policy-grid { grid-template-columns: 1fr; }
  .policy-form { grid-template-columns: 1fr; }
}
</style>
