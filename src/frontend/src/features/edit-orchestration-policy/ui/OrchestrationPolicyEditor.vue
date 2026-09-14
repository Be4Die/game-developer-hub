<template>
  <div class="policy-section">
    <div class="policy-header" @click="expanded = !expanded">
      <div class="policy-title">
        <Settings2 class="icon-sm" />
        <strong>Политика оркестрации</strong>
      </div>
      <ChevronDown v-if="expanded" class="icon-sm" />
      <ChevronRight v-else class="icon-sm" />
    </div>

    <div v-if="expanded" class="policy-body">
      <div v-if="loading" class="policy-loading">Загрузка…</div>

      <div v-else-if="!editing && policy" class="policy-read">
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
            <span class="policy-value">{{
              behaviorLabels[policy.scale_behavior] || policy.scale_behavior
            }}</span>
          </div>
          <div v-if="policy.scale_behavior === 'SCALE_BEHAVIOR_QUEUE'" class="policy-item">
            <span class="policy-label">Расположение очереди</span>
            <span class="policy-value">{{
              queueLocationLabels[policy.queue_location] || policy.queue_location
            }}</span>
          </div>
          <div class="policy-item">
            <span class="policy-label">Резервация (сек)</span>
            <span class="policy-value">{{ policy.queue_reservation_seconds ?? 30 }}</span>
          </div>
          <div class="policy-item">
            <span class="policy-label">Heartbeat таймаут (сек)</span>
            <span class="policy-value">{{ policy.queue_heartbeat_timeout ?? 15 }}</span>
          </div>
          <div class="policy-item">
            <span class="policy-label">Нода</span>
            <span class="policy-value">{{ nodePreferenceLabel(policy.node_preference) }}</span>
          </div>
        </div>
        <div class="policy-read-actions">
          <button class="btn-outline btn-sm" @click="startEdit">Изменить</button>
        </div>
      </div>

      <div v-else-if="editing" class="policy-edit">
        <div class="policy-form">
          <label>
            <span class="label-row">
              Режим оркестрации
              <Tooltip position="right"
                >Определяет базовое поведение оркестратора.<br /><br />«Только ручное управление» —
                никакого авто-вмешательства.<br />«Держать запущенным» — поддерживает целевое число
                инстансов.<br />«Экономичный» — останавливает при простое, запускает при обнаружении
                игроков.</Tooltip
              >
            </span>
            <select v-model="draft.mode">
              <option value="ORCHESTRATION_MODE_DISABLED">Только ручное управление</option>
              <option value="ORCHESTRATION_MODE_KEEP_ALIVE">Держать запущенным</option>
              <option value="ORCHESTRATION_MODE_SCALE_TO_ZERO">Экономичный (scale-to-zero)</option>
            </select>
          </label>
          <label>
            <span class="label-row">
              Целевое число инстансов
              <Tooltip position="right"
                >Сколько инстансов должно быть запущено в режиме «Держать запущенным».<br /><br />При
                падении одного из них оркестратор автоматически поднимет новый, чтобы поддержать это
                число.</Tooltip
              >
            </span>
            <input v-model.number="draft.target_instances" type="number" min="0" />
          </label>
          <label class="checkbox">
            <input v-model="draft.auto_restart" type="checkbox" />
            <span class="label-row">
              Авторестарт при падении
              <Tooltip position="right"
                >Если включён, оркестратор автоматически перезапустит инстанс при краше.<br /><br />Если
                контейнер был удалён — запустит новый инстанс взамен.</Tooltip
              >
            </span>
          </label>
          <label>
            <span class="label-row">
              Таймаут простоя (мин)
              <Tooltip position="right"
                >Актуально только в режиме «Экономичный».<br /><br />Через сколько минут без игроков
                инстанс будет автоматически остановлен для экономии ресурсов.</Tooltip
              >
            </span>
            <input v-model.number="draft.scale_to_zero_timeout" type="number" min="1" />
          </label>
          <label>
            <span class="label-row">
              Версия билда по умолчанию
              <Tooltip position="right"
                >Какой билд использовать при авто-старте инстанса.<br /><br />Выберите конкретную
                версию или оставьте «latest» — тогда всегда будет использоваться последний
                загруженный билд.</Tooltip
              >
            </span>
            <select v-model="draft.default_build_version">
              <option value="latest">latest</option>
              <option v-for="b in builds" :key="b.build_version" :value="b.build_version">
                {{ b.build_version }}
              </option>
            </select>
          </label>
          <label>
            <span class="label-row">
              Макс. игроков / инстанс
              <Tooltip position="right"
                >Порог для определения переполнения.<br /><br />Если число игроков достигает или
                превышает это значение, срабатывает выбранное поведение при переполнении: запуск
                нового инстанса или очередь.</Tooltip
              >
            </span>
            <input v-model.number="draft.max_players_per_instance" type="number" min="1" />
          </label>
          <label>
            <span class="label-row">
              Макс. инстансов на игру
              <Tooltip position="right"
                >Абсолютный потолок количества инстансов для защиты от неконтролируемого
                масштабирования.<br /><br />Учитываются все инстансы: запущенные, остановленные и
                упавшие.</Tooltip
              >
            </span>
            <input v-model.number="draft.max_instances_per_game" type="number" min="1" />
          </label>
          <label>
            <span class="label-row">
              При переполнении
              <Tooltip position="right"
                >Что делать, когда инстанс заполнен.<br /><br />«Запускать новый инстанс» —
                оркестратор поднимет дополнительный сервер.<br />«Очередь игроков» — новые игроки
                будут ждать освобождения слотов.</Tooltip
              >
            </span>
            <select v-model="draft.scale_behavior">
              <option value="SCALE_BEHAVIOR_SPAWN">Запускать новый инстанс</option>
              <option value="SCALE_BEHAVIOR_QUEUE">Очередь игроков</option>
            </select>
          </label>
          <label v-if="draft.scale_behavior === 'SCALE_BEHAVIOR_QUEUE'">
            <span class="label-row">
              Расположение очереди
              <Tooltip position="right"
                >Где реализована очередь игроков.<br /><br />«На стороне клиента» — оркестратор
                управляет очередью, игроки polling'ят статус.<br />«На стороне сервера» — игровой
                сервер сам управляет очередью, оркестратор только масштабирует по размеру
                очереди.</Tooltip
              >
            </span>
            <select v-model="draft.queue_location">
              <option value="QUEUE_LOCATION_CLIENT">На стороне клиента</option>
              <option value="QUEUE_LOCATION_SERVER">На стороне сервера</option>
            </select>
          </label>
          <label
            v-if="
              draft.scale_behavior === 'SCALE_BEHAVIOR_QUEUE' &&
              draft.queue_location === 'QUEUE_LOCATION_SERVER'
            "
          >
            <span class="label-row">
              Порог масштабирования по очереди
              <Tooltip position="right"
                >При каком размере очереди на игровом сервере запускать новый инстанс. 0 =
                автоматически (половина max_players).</Tooltip
              >
            </span>
            <input
              v-model.number="draft.queue_scale_up_threshold"
              type="number"
              min="0"
              max="1000"
            />
          </label>
          <label v-if="draft.scale_behavior === 'SCALE_BEHAVIOR_QUEUE'">
            <span class="label-row">
              Резервация слота (сек)
              <Tooltip position="right"
                >Сколько секунд даётся игроку на подключение после выделения слота.</Tooltip
              >
            </span>
            <input
              v-model.number="draft.queue_reservation_seconds"
              type="number"
              min="5"
              max="300"
            />
          </label>
          <label v-if="draft.scale_behavior === 'SCALE_BEHAVIOR_QUEUE'">
            <span class="label-row">
              Heartbeat таймаут (сек)
              <Tooltip position="right">Выкидывание из очереди при отсутствии heartbeat.</Tooltip>
            </span>
            <input v-model.number="draft.queue_heartbeat_timeout" type="number" min="5" max="120" />
          </label>
          <label v-if="draft.scale_behavior === 'SCALE_BEHAVIOR_QUEUE'">
            <span class="label-row">
              Таймаут очереди (сек)
              <Tooltip position="right">Авто-отмена после указанного времени в очереди.</Tooltip>
            </span>
            <input v-model.number="draft.queue_max_wait_seconds" type="number" min="0" max="3600" />
          </label>
          <label>
            <span class="label-row">
              Предпочтительная нода
              <Tooltip position="right"
                >На какой ноде развёртывать инстансы при авто-старте.<br /><br />«Авто» —
                оркестратор сам выберет наименее загруженную онлайн-ноду.<br />«Конкретная нода» —
                все авто-старты будут направлены на выбранный сервер.</Tooltip
              >
            </span>
            <select v-model="draft.node_preference">
              <option value="auto">Авто (наименее загруженная)</option>
              <option v-for="n in onlineNodeList" :key="n.id" :value="n.id">
                {{ n.is_platform ? '⭐ [Платформа] ' : '' }}{{ n.address }} ({{ n.id }})
              </option>
            </select>
          </label>
        </div>

        <div v-if="saveError" class="policy-error">
          {{ saveError }}
        </div>

        <div class="policy-edit-actions">
          <button class="btn-primary btn-sm" :disabled="saving" @click="savePolicy">
            {{ saving ? 'Сохранение…' : 'Сохранить' }}
          </button>
          <button class="btn-outline btn-sm" :disabled="saving" @click="cancelEdit">Отмена</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue';
import { Settings2, ChevronDown, ChevronRight } from 'lucide-vue-next';
import { Tooltip } from '@/shared/ui';
import {
  getPolicy,
  setPolicy,
  modeLabels,
  behaviorLabels,
  queueLocationLabels,
  nodePreferenceLabel,
} from '@/entities/policy';
import { showToast } from '@/shared/lib';

const props = defineProps({
  gameId: { type: [String, Number], required: true },
  builds: { type: Array, default: () => [] },
  nodes: { type: Array, default: () => [] },
});

const expanded = ref(false);
const loading = ref(false);
const editing = ref(false);
const saving = ref(false);
const saveError = ref(null);

const policy = ref(null);
const draft = ref({});

const onlineNodeList = computed(() =>
  props.nodes.filter((n) => n.status === 'online' || n.status === 'NODE_STATUS_ONLINE')
);

function startEdit() {
  draft.value = JSON.parse(JSON.stringify(policy.value || {}));
  editing.value = true;
  saveError.value = null;
}

function cancelEdit() {
  editing.value = false;
  saveError.value = null;
}

async function fetchPolicy() {
  loading.value = true;
  try {
    const data = await getPolicy(props.gameId);
    policy.value = data;
  } catch (e) {
    // silently fail
  } finally {
    loading.value = false;
  }
}

async function savePolicy() {
  saving.value = true;
  saveError.value = null;
  try {
    const data = await setPolicy(props.gameId, draft.value);
    policy.value = data;
    editing.value = false;
    showToast('Политика обновлена', 'success');
  } catch (e) {
    saveError.value = e.response?.data?.message ?? 'Ошибка сохранения политики';
  } finally {
    saving.value = false;
  }
}

onMounted(() => {
  fetchPolicy();
});
</script>

<style scoped>
.policy-section {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: var(--radius-lg);
  margin-bottom: 24px;
  overflow: hidden;
}

.policy-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  cursor: pointer;
  user-select: none;
}

.policy-header:hover {
  background: var(--bg-hover);
}

.policy-title {
  display: flex;
  align-items: center;
  gap: 8px;
  color: var(--text-main);
}

.policy-body {
  padding: 0 20px 20px;
  border-top: 1px solid var(--border);
}

.policy-loading {
  padding: 20px 0;
  color: var(--text-muted);
  font-size: 0.9rem;
}

.policy-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
  gap: 16px;
  padding: 16px 0;
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
  font-weight: 600;
  color: var(--text-main);
}

.policy-read-actions,
.policy-edit-actions {
  display: flex;
  gap: 10px;
  margin-top: 12px;
}

.policy-form {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
  gap: 16px;
  padding: 16px 0;
}

.policy-form label {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.policy-form label.checkbox {
  flex-direction: row;
  align-items: center;
  gap: 8px;
  margin-top: 24px;
}

.label-row {
  display: flex;
  align-items: center;
  font-size: 0.8rem;
  font-weight: 600;
  color: var(--text-muted);
}

.policy-form select,
.policy-form input[type='number'],
.policy-form input[type='text'] {
  width: 100%;
  padding: 8px 12px;
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  background: var(--bg-app);
  color: var(--text-main);
  font-size: 0.88rem;
  outline: none;
  box-sizing: border-box;
}

.policy-error {
  margin-top: 12px;
  padding: 8px 12px;
  background: var(--danger-light);
  color: var(--danger);
  border-radius: var(--radius-md);
  font-size: 0.85rem;
}
</style>
