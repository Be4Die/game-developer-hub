<template>
  <div class="overview tab-fade-in">
    <div class="overview-header">
      <h1>Игровые сервера</h1>
    </div>

    <!-- Карточки-сводки -->
    <div class="summary-grid">
      <div class="summary-card">
        <span class="summary-label">Билды</span>
        <span class="summary-value">{{ builds.length }}</span>
      </div>
      <div class="summary-card">
        <span class="summary-label">Работающие</span>
        <span class="summary-value">{{ runningCount }}<span class="summary-sub"> / {{ instances.length }}</span></span>
      </div>
      <div class="summary-card">
        <span class="summary-label">Игроков онлайн</span>
        <span class="summary-value">{{ totalPlayers }}</span>
      </div>
      <div class="summary-card">
        <span class="summary-label">Ноды</span>
        <span class="summary-value">{{ onlineNodes }}<span class="summary-sub"> / {{ nodes.length }}</span></span>
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
        <div v-else class="empty-state">Нет запущенных инстансов</div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { Upload, Play } from 'lucide-vue-next'
import StatusBadge from '../../components/orchestrator/StatusBadge.vue'
import { mockBuilds, mockInstances, mockNodes } from '../../data/mock-orchestrator'

defineProps({ gameId: { type: [String, Number], required: true } })

const builds = computed(() => mockBuilds)
const instances = computed(() => mockInstances)
const nodes = computed(() => mockNodes)

const runningCount = computed(() => instances.value.filter(i => i.status === 'running').length)
const totalPlayers = computed(() => instances.value.reduce((sum, i) => sum + (i.player_count ?? 0), 0))
const onlineNodes = computed(() => nodes.value.filter(n => n.status === 'online').length)

function formatDate(ts) {
  return new Date(ts).toLocaleDateString('ru-RU', { day: 'numeric', month: 'short', hour: '2-digit', minute: '2-digit' })
}
</script>

<style scoped>
.tab-fade-in { animation: fadeIn 0.3s ease; }
@keyframes fadeIn { from { opacity: 0; transform: translateY(10px); } to { opacity: 1; transform: translateY(0); } }

.overview-header { margin-bottom: 24px; }
.overview-header h1 { margin: 0; }

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
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 20px;
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: var(--radius-lg);
  text-decoration: none;
  color: var(--text-main);
  transition: 0.15s;
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
  text-align: left;
  padding: 12px 16px;
  font-size: 0.78rem;
  font-weight: 600;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.03em;
  background: var(--bg-secondary);
  border-bottom: 1px solid var(--border);
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
</style>
