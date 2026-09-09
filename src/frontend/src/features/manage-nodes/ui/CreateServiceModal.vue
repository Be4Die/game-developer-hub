<template>
  <div class="modal-overlay" @click.self="$emit('cancel')">
    <div class="modal card service-modal">
      <div class="modal-header">
        <div>
          <h3>Развернуть сервис хранения</h3>
          <p class="modal-subtitle">Выберите тип сервиса данных или персистентного тома для ноды</p>
        </div>
        <button class="close-btn" @click="$emit('cancel')">&times;</button>
      </div>

      <div class="service-type-grid">
        <div
          v-for="st in serviceTypes"
          :key="st.type"
          class="service-type-card"
          :class="{ active: form.service_type === st.type }"
          @click="selectServiceType(st.type)"
        >
          <div class="type-icon-wrapper" :style="{ backgroundColor: st.bgColor, color: st.color }">
            <component :is="st.icon" class="type-icon" />
          </div>
          <div class="type-info">
            <span class="type-title">{{ st.title }}</span>
            <span class="type-desc">{{ st.desc }}</span>
          </div>
        </div>
      </div>

      <div class="form-content">
        <!-- Имя сервиса -->
        <div class="form-group">
          <label>Имя {{ form.service_type === 'volume' ? 'тома' : 'сервиса' }} *</label>
          <input
            v-model="form.name"
            type="text"
            class="form-input"
            :placeholder="getNamePlaceholder()"
          />
          <p class="hint">
            <template v-if="form.service_type === 'volume'">
              Имя каталога в хранилище ноды (<code>/var/lib/gdh/volumes/{{ form.name || '<имя>' }}</code>).
            </template>
            <template v-else>
              Используется для имени контейнера, тома и обращения в Docker-сети.
            </template>
          </p>
        </div>

        <!-- Подсказка для Volume -->
        <div v-if="form.service_type === 'volume'" class="info-callout">
          <FolderPlus class="callout-icon" />
          <div class="callout-text">
            <strong>Персистентный том на хосте:</strong>
            Игровые сервера получают прямой доступ к каталогу на диске через переменную окружения
            <code>STORAGE_VOLUME_PATH</code>. Идеально для встроенных баз данных (SQLite, RocksDB),
            сохранений мира или бинарных кэшей без накладных расходов на СУБД-контейнеры.
          </div>
        </div>

        <!-- Дополнительные параметры СУБД (сворачиваемые) -->
        <div v-if="form.service_type !== 'volume'" class="collapsible-section">
          <button
            type="button"
            class="collapsible-trigger"
            @click="showAdvanced = !showAdvanced"
          >
            <span>Дополнительные параметры (Пароль, БД, Порт)</span>
            <component :is="showAdvanced ? ChevronUp : ChevronDown" class="icon-xs" />
          </button>

          <div v-show="showAdvanced" class="collapsible-body">
            <div
              v-if="form.service_type === 'postgres' || form.service_type === 'mysql'"
              class="form-group"
            >
              <label>Имя базы данных (опционально)</label>
              <input
                v-model="form.db_name"
                type="text"
                class="form-input"
                placeholder="game_db (по умолчанию)"
              />
            </div>

            <div class="form-group">
              <label>Пароль (опционально)</label>
              <input
                v-model="form.password"
                type="password"
                class="form-input"
                placeholder="Оставьте пустым для автогенерации надежного пароля"
              />
              <p class="hint">
                При пустом поле пароль будет сгенерирован автоматически и безопасно инжектирован в инстансы.
              </p>
            </div>

            <div class="form-group">
              <label>Порт на хосте (опционально)</label>
              <input
                v-model.number="form.port"
                type="number"
                class="form-input"
                placeholder="0 — автовыбор свободного порта"
                min="0"
                max="65535"
              />
            </div>
          </div>
        </div>

        <!-- Ограничение доступа по играм -->
        <div class="form-group">
          <label>Доступ к хранилищу</label>

          <!-- Стилизованный сегментированный переключатель -->
          <div class="segmented-control">
            <button
              type="button"
              class="segment-btn"
              :class="{ active: accessMode === 'all' }"
              @click="accessMode = 'all'"
            >
              <Globe class="segment-icon" />
              <span>Доступно всем играм</span>
            </button>
            <button
              type="button"
              class="segment-btn"
              :class="{ active: accessMode === 'specific' }"
              :disabled="!projects.length"
              :title="!projects.length ? 'У вас пока нет созданных проектов' : ''"
              @click="projects.length && (accessMode = 'specific')"
            >
              <Lock class="segment-icon" />
              <span>Ограничить по играм</span>
            </button>
          </div>

          <!-- Выпадающий список проектов с чекбоксами (мультивыбор) -->
          <div v-if="accessMode === 'specific'" class="game-picker-container">
            <div v-if="loadingProjects" class="picker-loading">
              Загрузка списка проектов...
            </div>

            <div v-else-if="projects.length" ref="dropdownRef" class="multiselect-dropdown-wrap">
              <div
                class="dropdown-field"
                :class="{ active: isDropdownOpen }"
                tabindex="0"
                @click="isDropdownOpen = !isDropdownOpen"
              >
                <div class="dropdown-selected-text">
                  <span v-if="!selectedProjectIds.length" class="placeholder">
                    Выберите проекты...
                  </span>
                  <span v-else class="summary-text">
                    Выбрано проектов: {{ selectedProjectIds.length }}
                  </span>
                </div>
                <ChevronDown class="dropdown-chevron" :class="{ rotated: isDropdownOpen }" />
              </div>

              <!-- Меню выпадающего списка -->
              <div v-if="isDropdownOpen" class="dropdown-popover">
                <div class="dropdown-popover-header">
                  <span class="popover-title">Проекты</span>
                  <div class="popover-actions">
                    <button type="button" class="btn-link" @click.stop="selectAllProjects">Все</button>
                    <span class="divider">|</span>
                    <button type="button" class="btn-link" @click.stop="clearSelectedProjects">Сбросить</button>
                  </div>
                </div>

                <div class="dropdown-options-list">
                  <label
                    v-for="proj in projects"
                    :key="proj.id"
                    class="dropdown-option-row"
                    :class="{ selected: selectedProjectIds.includes(proj.id) }"
                    @click.stop
                  >
                    <input
                      v-model="selectedProjectIds"
                      type="checkbox"
                      :value="proj.id"
                      class="option-checkbox"
                    />
                    <span class="option-title">{{ proj.title || proj.name || 'Без названия' }}</span>
                  </label>
                </div>
              </div>

              <!-- Чипсы выбранных проектов -->
              <div v-if="selectedProjectIds.length > 0" class="selected-chips">
                <span
                  v-for="id in selectedProjectIds"
                  :key="id"
                  class="project-chip"
                >
                  <span class="chip-text">{{ getProjectTitle(id) }}</span>
                  <button type="button" class="chip-remove" @click.stop="removeProject(id)">&times;</button>
                </span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div v-if="error" class="form-error">
        {{ error }}
      </div>

      <div class="modal-actions">
        <button class="btn-outline" @click="$emit('cancel')">Отмена</button>
        <button
          class="btn-primary"
          :disabled="creating || !form.name.trim() || (accessMode === 'specific' && !selectedProjectIds.length)"
          @click="submitCreate"
        >
          {{ creating ? 'Развертывание...' : 'Развернуть' }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, onUnmounted } from 'vue';
import {
  Database,
  Layers,
  Server,
  FolderPlus,
  Globe,
  Lock,
  ChevronDown,
  ChevronUp,
} from 'lucide-vue-next';
import { createNodeService } from '@/entities/node';
import { listProjects } from '@/entities/project';
import { showToast } from '@/shared/lib';

const props = defineProps({
  nodeId: { type: [Number, String], required: true },
});

const emit = defineEmits(['created', 'cancel']);

const creating = ref(false);
const error = ref(null);
const showAdvanced = ref(false);

const accessMode = ref('all');
const selectedProjectIds = ref([]);
const projects = ref([]);
const loadingProjects = ref(false);
const isDropdownOpen = ref(false);
const dropdownRef = ref(null);

const serviceTypes = [
  {
    type: 'postgres',
    title: 'PostgreSQL 16',
    desc: 'Реляционная БД для профилей игроков, инвентаря и сохранений',
    icon: Database,
    color: '#336791',
    bgColor: 'rgba(51, 103, 145, 0.15)',
  },
  {
    type: 'redis',
    title: 'Redis 7',
    desc: 'Быстрый in-memory кэш, очереди, pub/sub для комнат и лобби',
    icon: Layers,
    color: '#dc382d',
    bgColor: 'rgba(220, 56, 45, 0.15)',
  },
  {
    type: 'volume',
    title: 'Персистентный том',
    desc: 'Каталог на диске хоста для SQLite, RocksDB или кастомных файлов',
    icon: FolderPlus,
    color: '#10b981',
    bgColor: 'rgba(16, 185, 129, 0.15)',
  },
  {
    type: 'mysql',
    title: 'MySQL 8.0',
    desc: 'Классическая реляционная БД для игровых серверов',
    icon: Server,
    color: '#00758f',
    bgColor: 'rgba(0, 117, 143, 0.15)',
  },
];

const form = reactive({
  service_type: 'postgres',
  name: 'game-postgres',
  db_name: '',
  password: '',
  port: null,
});

function getNamePlaceholder() {
  switch (form.service_type) {
    case 'volume':
      return 'shared-saves, game-data';
    case 'redis':
      return 'game-redis';
    case 'mysql':
      return 'game-mysql';
    default:
      return 'game-postgres';
  }
}

function selectServiceType(type) {
  form.service_type = type;
  if (!form.name || serviceTypes.some((s) => form.name === `game-${s.type}`)) {
    if (type === 'volume') {
      form.name = 'game-volume';
    } else {
      form.name = `game-${type}`;
    }
  }
}

function selectAllProjects() {
  selectedProjectIds.value = projects.value.map((p) => p.id);
}

function clearSelectedProjects() {
  selectedProjectIds.value = [];
}

function getProjectTitle(id) {
  const p = projects.value.find((item) => item.id === id);
  return p ? (p.title || p.name || 'Без названия') : 'Проект #' + id;
}

function removeProject(id) {
  selectedProjectIds.value = selectedProjectIds.value.filter((i) => i !== id);
}

function handleClickOutside(event) {
  if (dropdownRef.value && !dropdownRef.value.contains(event.target)) {
    isDropdownOpen.value = false;
  }
}

async function loadUserProjects() {
  loadingProjects.value = true;
  try {
    const res = await listProjects();
    projects.value = res.projects || [];
  } catch (err) {
    console.error('Не удалось загрузить список игр:', err);
    projects.value = [];
  } finally {
    loadingProjects.value = false;
  }
}

async function submitCreate() {
  if (!form.name.trim()) {
    error.value = 'Укажите имя сервиса или тома';
    return;
  }

  if (accessMode.value === 'specific' && !selectedProjectIds.value.length) {
    error.value = 'Выберите хотя бы один проект для ограничения доступа';
    return;
  }

  creating.value = true;
  error.value = null;

  try {
    let allowedGameIds = [];
    if (accessMode.value === 'specific') {
      allowedGameIds = selectedProjectIds.value.map(Number);
    }

    const payload = {
      service_type: form.service_type,
      name: form.name.trim(),
      password: form.password ? form.password.trim() : '',
      db_name: form.db_name ? form.db_name.trim() : '',
      port: form.port ? Number(form.port) : 0,
      allowed_game_ids: allowedGameIds,
    };

    const created = await createNodeService(props.nodeId, payload);
    showToast(
      form.service_type === 'volume' ? 'Персистентный том создан' : 'Сервис успешно развернут',
      'success',
    );
    emit('created', created);
  } catch (err) {
    error.value = err.response?.data?.message || err.message || 'Ошибка развертывания сервиса';
  } finally {
    creating.value = false;
  }
}

onMounted(() => {
  loadUserProjects();
  window.addEventListener('click', handleClickOutside);
});

onUnmounted(() => {
  window.removeEventListener('click', handleClickOutside);
});
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.65);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 16px;
}

.service-modal {
  max-width: 640px;
  width: 100%;
  max-height: 90vh;
  overflow-y: auto;
  padding: 24px;
  border-radius: var(--radius-lg, 12px);
  background: var(--bg-card, #161b22);
  border: 1px solid var(--border, #30363d);
  box-shadow: 0 20px 40px rgba(0, 0, 0, 0.4);
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 20px;
}

.modal-header h3 {
  margin: 0;
  font-size: 1.25rem;
  font-weight: 700;
  color: var(--text-main, #f0f6fc);
}

.modal-subtitle {
  margin: 4px 0 0;
  font-size: 0.82rem;
  color: var(--text-muted, #8b949e);
}

.close-btn {
  background: transparent;
  border: none;
  font-size: 1.5rem;
  color: var(--text-muted, #8b949e);
  cursor: pointer;
  line-height: 1;
  padding: 4px;
  border-radius: 4px;
}

.close-btn:hover {
  color: var(--text-main, #f0f6fc);
}

.service-type-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 10px;
  margin-bottom: 20px;
}

.service-type-card {
  display: flex;
  gap: 12px;
  padding: 12px;
  border-radius: var(--radius-md, 8px);
  border: 1px solid var(--border, #30363d);
  background: var(--bg-app, #0d1117);
  cursor: pointer;
  transition: all 0.15s ease;
}

.service-type-card:hover {
  border-color: var(--primary, #58a6ff);
  background: var(--bg-hover, #21262d);
}

.service-type-card.active {
  border-color: var(--primary, #58a6ff);
  background: rgba(88, 166, 255, 0.08);
  box-shadow: 0 0 0 1px var(--primary, #58a6ff);
}

.type-icon-wrapper {
  width: 38px;
  height: 38px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.type-icon {
  width: 20px;
  height: 20px;
}

.type-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
  overflow: hidden;
}

.type-title {
  font-size: 0.88rem;
  font-weight: 600;
  color: var(--text-main, #f0f6fc);
}

.type-desc {
  font-size: 0.73rem;
  color: var(--text-muted, #8b949e);
  line-height: 1.25;
}

.form-content {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.form-group label {
  display: block;
  font-size: 0.82rem;
  font-weight: 600;
  color: var(--text-muted, #8b949e);
  margin-bottom: 6px;
}

.form-input {
  width: 100%;
  padding: 9px 12px;
  border: 1px solid var(--border, #30363d);
  border-radius: var(--radius-md, 6px);
  background: var(--bg-app, #0d1117);
  color: var(--text-main, #f0f6fc);
  font-size: 0.9rem;
  outline: none;
  box-sizing: border-box;
  transition: border-color 0.15s;
}

.form-input:focus {
  border-color: var(--primary, #58a6ff);
}

.hint {
  font-size: 0.78rem;
  color: var(--text-muted, #8b949e);
  margin-top: 5px;
  margin-bottom: 0;
  line-height: 1.35;
}

.hint code {
  background: rgba(255, 255, 255, 0.08);
  padding: 1px 4px;
  border-radius: 4px;
}

.info-callout {
  display: flex;
  gap: 12px;
  padding: 12px 14px;
  border-radius: var(--radius-md, 8px);
  background: rgba(16, 185, 129, 0.08);
  border: 1px solid rgba(16, 185, 129, 0.25);
  color: var(--text-main, #f0f6fc);
}

.callout-icon {
  width: 20px;
  height: 20px;
  color: #10b981;
  flex-shrink: 0;
  margin-top: 2px;
}

.callout-text {
  font-size: 0.8rem;
  line-height: 1.4;
}

.callout-text strong {
  display: block;
  margin-bottom: 2px;
  color: #10b981;
}

.callout-text code {
  background: rgba(255, 255, 255, 0.08);
  padding: 1px 4px;
  border-radius: 4px;
  color: #34d399;
}

/* Сворачиваемые настройки */
.collapsible-section {
  border: 1px solid var(--border, #30363d);
  border-radius: var(--radius-md, 8px);
  overflow: hidden;
  background: var(--bg-app, #0d1117);
}

.collapsible-trigger {
  width: 100%;
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 14px;
  background: transparent;
  border: none;
  color: var(--text-main, #f0f6fc);
  font-size: 0.84rem;
  font-weight: 600;
  cursor: pointer;
  transition: background 0.15s;
}

.collapsible-trigger:hover {
  background: var(--bg-hover, #21262d);
}

.collapsible-body {
  padding: 14px;
  border-top: 1px solid var(--border, #30363d);
  display: flex;
  flex-direction: column;
  gap: 12px;
}

/* Сегментированный переключатель доступа */
.segmented-control {
  display: flex;
  gap: 4px;
  padding: 4px;
  background: var(--bg-app, #0d1117);
  border: 1px solid var(--border, #30363d);
  border-radius: var(--radius-md, 8px);
  margin-bottom: 8px;
}

.segment-btn {
  flex: 1;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 8px 12px;
  background: transparent;
  border: none;
  border-radius: var(--radius-sm, 6px);
  color: var(--text-muted, #8b949e);
  font-size: 0.85rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.15s ease;
}

.segment-btn:hover:not(:disabled) {
  color: var(--text-main, #f0f6fc);
}

.segment-btn.active {
  background: var(--primary, #58a6ff);
  color: #ffffff;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.25);
}

.segment-btn:disabled {
  opacity: 0.45;
  cursor: not-allowed;
}

.segment-icon {
  width: 15px;
  height: 15px;
}

/* Выпадающий список мультивыбора */
.game-picker-container {
  margin-top: 10px;
}

.picker-loading {
  padding: 16px;
  text-align: center;
  font-size: 0.85rem;
  color: var(--text-muted, #8b949e);
}

.multiselect-dropdown-wrap {
  position: relative;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.dropdown-field {
  width: 100%;
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 9px 12px;
  background: var(--bg-app, #0d1117);
  border: 1px solid var(--border, #30363d);
  border-radius: var(--radius-md, 6px);
  cursor: pointer;
  box-sizing: border-box;
  transition: border-color 0.15s;
}

.dropdown-field:hover,
.dropdown-field.active {
  border-color: var(--primary, #58a6ff);
}

.dropdown-selected-text .placeholder {
  color: var(--text-muted, #8b949e);
  font-size: 0.88rem;
}

.dropdown-selected-text .summary-text {
  color: var(--text-main, #f0f6fc);
  font-size: 0.88rem;
  font-weight: 500;
}

.dropdown-chevron {
  width: 16px;
  height: 16px;
  color: var(--text-muted, #8b949e);
  transition: transform 0.2s ease;
}

.dropdown-chevron.rotated {
  transform: rotate(180deg);
}

.dropdown-popover {
  position: absolute;
  top: calc(100% - 4px);
  left: 0;
  right: 0;
  z-index: 50;
  background: var(--bg-card, #161b22);
  border: 1px solid var(--border, #30363d);
  border-radius: var(--radius-md, 8px);
  box-shadow: 0 12px 28px rgba(0, 0, 0, 0.45);
  overflow: hidden;
}

.dropdown-popover-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 12px;
  background: var(--bg-secondary, #21262d);
  border-bottom: 1px solid var(--border, #30363d);
}

.popover-title {
  font-size: 0.78rem;
  font-weight: 600;
  color: var(--text-muted, #8b949e);
  text-transform: uppercase;
  letter-spacing: 0.3px;
}

.popover-actions {
  display: flex;
  align-items: center;
  gap: 6px;
}

.btn-link {
  background: none;
  border: none;
  color: var(--primary, #58a6ff);
  font-size: 0.78rem;
  cursor: pointer;
  padding: 0;
}

.btn-link:hover {
  text-decoration: underline;
}

.divider {
  color: var(--border, #30363d);
  font-size: 0.78rem;
}

.dropdown-options-list {
  max-height: 180px;
  overflow-y: auto;
  padding: 4px;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.dropdown-option-row {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 7px 10px;
  border-radius: 4px;
  cursor: pointer;
  transition: background 0.1s;
  user-select: none;
}

.dropdown-option-row:hover {
  background: var(--bg-hover, #21262d);
}

.dropdown-option-row.selected {
  background: rgba(88, 166, 255, 0.1);
}

.option-checkbox {
  cursor: pointer;
  accent-color: var(--primary, #58a6ff);
  width: 15px;
  height: 15px;
}

.option-title {
  font-size: 0.85rem;
  color: var(--text-main, #f0f6fc);
  flex: 1;
}

/* Чипсы выбранных проектов */
.selected-chips {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-top: 4px;
}

.project-chip {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 3px 8px;
  background: rgba(88, 166, 255, 0.1);
  border: 1px solid rgba(88, 166, 255, 0.25);
  border-radius: 6px;
  font-size: 0.78rem;
  color: var(--text-main, #f0f6fc);
}

.chip-text {
  max-width: 180px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.chip-remove {
  background: none;
  border: none;
  color: var(--text-muted, #8b949e);
  cursor: pointer;
  font-size: 1rem;
  line-height: 1;
  padding: 0;
  display: flex;
  align-items: center;
}

.chip-remove:hover {
  color: #f85149;
}

.form-error {
  margin-top: 16px;
  padding: 10px 14px;
  background: rgba(248, 81, 73, 0.12);
  color: #f85149;
  border: 1px solid rgba(248, 81, 73, 0.3);
  border-radius: var(--radius-md, 6px);
  font-size: 0.85rem;
}

.modal-actions {
  display: flex;
  gap: 12px;
  justify-content: flex-end;
  margin-top: 24px;
}

.icon-xs {
  width: 14px;
  height: 14px;
}

@media (max-width: 600px) {
  .service-type-grid {
    grid-template-columns: 1fr;
  }
}
</style>
