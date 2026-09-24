<template>
  <div class="tab-content tab-fade-in purchases-page">
    <!-- БАННЕР БЛОКИРОВКИ НА МОДЕРАЦИИ -->
    <div v-if="isLocked" class="purchases-lock-banner">
      <AlertCircle class="icon-sm text-warning" />
      <div class="lock-banner-content">
        <div class="lock-banner-title">Товары зафиксированы на время модерации</div>
        <div class="lock-banner-desc">
          Проект находится на проверке у модератора. Добавление, редактирование и удаление товаров заблокировано до решения модератора.
        </div>
      </div>
    </div>

    <!-- Верхняя панель: курс валюты и действие добавления -->
    <div v-if="!loading" class="purchases-toolbar">
      <div class="rate-badge" title="1 WCoin = 1 рубль">
        <Coins class="icon-xs text-primary" />
        <span>1 WCoin = 1 ₽</span>
      </div>
      <button v-if="items.length > 0" class="btn-primary-action" :disabled="isLocked" @click="!isLocked && openCreateModal()">
        <Plus class="icon-sm" />
        <span>{{ t('purchases.actions.addItem') }}</span>
      </button>
    </div>

    <!-- Лоадер загрузки -->
    <div v-if="loading" class="state-container">
      <Loader2 class="spinner-md spin" />
      <p>{{ t('common.loading') }}</p>
    </div>

    <!-- Основной контент -->
    <div v-else class="purchases-table-card">
      <!-- Пустой список -->
      <EmptyState
        v-if="items.length === 0"
        :icon="ShoppingBag"
        :title="t('purchases.empty.title')"
        :action-text="t('purchases.actions.addItem')"
        :action-icon="Plus"
        :action-disabled="isLocked"
        @action="!isLocked && openCreateModal()"
      />

      <!-- Таблица товаров -->
      <div v-else class="table-wrapper">
        <table class="project-table">
          <thead>
            <tr>
              <th class="col-item">{{ t('purchases.table.item') }}</th>
              <th class="col-id">{{ t('purchases.table.itemId') }}</th>
              <th class="col-desc">{{ t('purchases.table.desc') }}</th>
              <th class="col-price">{{ t('purchases.table.price') }}</th>
              <th class="col-status">{{ t('purchases.table.status') }}</th>
              <th class="col-actions"></th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="item in items"
              :key="item.id || item.game_item_id"
              class="table-row table-row-clickable"
              @click="openViewModal(item)"
            >
              <td class="col-item">
                <div class="item-meta">
                  <div class="item-icon-box">
                    <img
                      v-if="item.image_url"
                      :src="getMediaUrl(item.image_url, projectId)"
                      alt="Icon"
                      class="item-icon-img"
                      @error="onImageError"
                    />
                    <div v-else class="item-icon-placeholder">
                      <ShoppingBag class="icon-xs text-muted" />
                    </div>
                  </div>
                  <span class="item-name" :title="item.name">{{ item.name }}</span>
                </div>
              </td>
              <td class="col-id">
                <span class="item-id-text" :title="item.game_item_id">{{ item.game_item_id }}</span>
              </td>
              <td class="col-desc">
                <div v-if="item.description" class="item-desc-cell" :title="item.description">
                  {{ item.description }}
                </div>
                <span v-else class="text-muted">—</span>
              </td>
              <td class="col-price">
                <div class="price-val">
                  <span class="price-num">{{ item.price_coins }}</span>
                  <span class="price-unit">WCoin</span>
                </div>
              </td>
              <td class="col-status">
                <span class="status-text" :class="item.is_active !== false ? 'status-active' : 'status-inactive'">
                  {{ item.is_active !== false ? 'Активен' : 'Скрыт' }}
                </span>
              </td>
              <td class="col-actions" @click.stop>
                <div class="row-actions">
                  <button
                    class="btn-icon"
                    :disabled="isLocked"
                    title="Редактировать товар"
                    @click.stop="!isLocked && openEditModal(item)"
                  >
                    <Edit2 class="icon-xs" />
                  </button>
                  <button
                    class="btn-icon text-danger-hover"
                    :disabled="isLocked"
                    title="Удалить товар"
                    @click.stop="!isLocked && openDeleteConfirm(item)"
                  >
                    <Trash2 class="icon-xs" />
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- МОДАЛЬНОЕ ОКНО: СОЗДАНИЕ / РЕДАКТИРОВАНИЕ -->
    <div v-if="showModal" class="modal-backdrop" @click.self="closeModal">
      <div class="modal-card">
        <div class="modal-header">
          <div class="modal-title-group">
            <ShoppingBag class="icon-sm text-primary" />
            <h3 class="modal-title">
              {{ isEditing ? t('purchases.modal.editTitle') : t('purchases.modal.createTitle') }}
            </h3>
          </div>
          <button class="btn-close" @click="closeModal">
            <X class="icon-sm" />
          </button>
        </div>

        <form @submit.prevent="saveItem" class="modal-body">
          <!-- Ошибка формы -->
          <div v-if="modalError" class="form-error-alert">
            <AlertCircle class="icon-xs" />
            <span>{{ modalError }}</span>
          </div>

          <!-- Загрузка иконки -->
          <div class="form-group">
            <label class="form-label">{{ t('purchases.fields.icon') }}</label>
            <div class="icon-upload-row">
              <div class="icon-preview-box">
                <img
                  v-if="iconPreview"
                  :src="iconPreview"
                  alt="Preview"
                  class="icon-preview-img"
                />
                <div v-else class="icon-preview-empty">
                  <ShoppingBag class="icon-md text-muted" />
                </div>
              </div>
              <div class="icon-upload-ctrl">
                <input
                  ref="fileInputRef"
                  type="file"
                  accept="image/png,image/jpeg,image/webp"
                  class="hidden-file-input"
                  @change="onFileSelected"
                />
                <button
                  type="button"
                  class="btn-secondary-sm"
                  @click="fileInputRef?.click()"
                >
                  <UploadCloud class="icon-xs" />
                  <span>{{ iconPreview ? 'Заменить иконку' : 'Загрузить файл' }}</span>
                </button>
                <div class="upload-hint text-muted">
                  PNG, JPG или WebP, квадратная (рекомендуется 128x128 или 256x256)
                </div>
              </div>
            </div>
          </div>

          <!-- Внутриигровой ID -->
          <div class="form-group">
            <label class="form-label">{{ t('purchases.fields.itemId') }} <span class="req">*</span></label>
            <input
              v-model="form.game_item_id"
              type="text"
              class="form-input"
              :disabled="isEditing"
              placeholder="pack_gold_100"
              required
              maxlength="64"
            />
          </div>

          <!-- Название -->
          <div class="form-group">
            <div class="label-with-counter">
              <label class="form-label">{{ t('purchases.fields.name') }} <span class="req">*</span></label>
              <span class="char-counter" :class="{ 'counter-warn': form.name.length > 45 }">
                {{ form.name.length }}/50
              </span>
            </div>
            <input
              v-model="form.name"
              type="text"
              class="form-input"
              placeholder="100 золотых монет"
              required
              maxlength="50"
            />
          </div>

          <!-- Описание -->
          <div class="form-group">
            <div class="label-with-counter">
              <label class="form-label">{{ t('purchases.fields.desc') }}</label>
              <span class="char-counter" :class="{ 'counter-warn': form.description.length > 180 }">
                {{ form.description.length }}/200
              </span>
            </div>
            <textarea
              v-model="form.description"
              class="form-textarea"
              placeholder="Краткое описание товара для игрока..."
              rows="2"
              maxlength="200"
            ></textarea>
          </div>

          <!-- Цена в монетах -->
          <div class="form-group">
            <label class="form-label">{{ t('purchases.fields.price') }} <span class="req">*</span></label>
            <div class="input-with-suffix">
              <input
                v-model.number="form.price_coins"
                type="number"
                min="1"
                step="1"
                class="form-input"
                placeholder="50"
                required
              />
              <span class="input-suffix">WCoin</span>
            </div>
          </div>

          <!-- Активность товара -->
          <div class="form-group-checkbox">
            <label class="checkbox-row">
              <input v-model="form.is_active" type="checkbox" class="styled-checkbox" />
              <span class="checkbox-title">{{ t('purchases.fields.isActive') }}</span>
            </label>
          </div>

          <!-- Футер модалки -->
          <div class="modal-footer">
            <button type="button" class="btn-cancel" @click="closeModal">
              {{ t('common.cancel') }}
            </button>
            <button type="submit" class="btn-save-primary" :disabled="saving || isLocked">
              <Loader2 v-if="saving" class="icon-xs spin" />
              <span>{{ saving ? t('common.saving') : t('common.save') }}</span>
            </button>
          </div>
        </form>
      </div>
    </div>

    <!-- МОДАЛЬНОЕ ОКНО: ПОДТВЕРЖДЕНИЕ УДАЛЕНИЯ -->
    <div v-if="itemToDelete" class="modal-backdrop" @click.self="itemToDelete = null">
      <div class="modal-card modal-sm">
        <div class="modal-header">
          <h3 class="modal-title">{{ t('purchases.delete.title') }}</h3>
          <button class="btn-close" @click="itemToDelete = null">
            <X class="icon-sm" />
          </button>
        </div>
        <div class="modal-body-p">
          <p>
            {{ t('purchases.delete.confirmText', { name: itemToDelete.name, id: itemToDelete.game_item_id }) }}
          </p>
        </div>
        <div class="modal-footer">
          <button type="button" class="btn-cancel" @click="itemToDelete = null">
            {{ t('common.cancel') }}
          </button>
          <button type="button" class="btn-danger-confirm" :disabled="deleting || isLocked" @click="confirmDelete">
            <Loader2 v-if="deleting" class="icon-xs spin" />
            <span>{{ deleting ? t('common.deleting') : t('common.delete') }}</span>
          </button>
        </div>
      </div>
    </div>

    <!-- МОДАЛЬНОЕ ОКНО: ДЕТАЛИ ТОВАРА (READONLY) -->
    <div v-if="viewItem" class="modal-backdrop" @click.self="closeViewModal">
      <div class="modal-card modal-view-card">
        <div class="modal-header">
          <div class="modal-title-group">
            <ShoppingBag class="icon-sm text-primary" />
            <h3 class="modal-title">{{ t('purchases.modal.viewTitle') }}</h3>
          </div>
          <button class="btn-close" @click="closeViewModal">
            <X class="icon-sm" />
          </button>
        </div>

        <div class="modal-body view-details-body">
          <!-- 1 строка: идентификатор (без синего акцента) + иконка копирования -->
          <div class="view-id-row">
            <span class="view-id-label">{{ t('purchases.fields.itemId') }}:</span>
            <span class="view-id-code">{{ viewItem.game_item_id }}</span>
            <button
              type="button"
              class="btn-copy-icon-only"
              :title="copiedId ? t('common.copied') : t('common.copy')"
              @click="copyItemId(viewItem.game_item_id)"
            >
              <Check v-if="copiedId" class="icon-xs text-success" />
              <Copy v-else class="icon-xs" />
            </button>
          </div>

          <!-- 2-3 строка: слева картинка, справа строка 2 (статус, дата время) и строка 3 (название, стоимость) -->
          <div class="view-hero-card">
            <!-- Картинка (занимает высоту строк 2-3) -->
            <div class="view-hero-media">
              <img
                v-if="viewItem.image_url"
                :src="getMediaUrl(viewItem.image_url, projectId)"
                alt="Icon"
                class="view-media-img"
                @error="onImageError"
              />
              <div v-else class="view-media-placeholder">
                <ShoppingBag class="icon-md text-muted" />
              </div>
            </div>

            <div class="view-hero-content">
              <!-- 2 строка (чуть меньше размером): статус активен или нет, дата время -->
              <div class="view-sub-row">
                <span class="status-indicator-pill" :class="viewItem.is_active !== false ? 'pill-active' : 'pill-inactive'">
                  <span class="status-dot"></span>
                  {{ viewItem.is_active !== false ? 'Активен' : 'Скрыт' }}
                </span>
                <span v-if="viewItem.created_at" class="view-meta-date text-muted">
                  {{ formatDateTime(viewItem.created_at) }}
                </span>
              </div>

              <!-- 3 строка: название, стоимость (у стоимости убрать иконку и в скобках приписку в рублях) -->
              <div class="view-main-row">
                <h4 class="view-item-name" :title="viewItem.name">{{ viewItem.name }}</h4>
                <div class="view-price-clean">
                  <span class="view-price-num">{{ viewItem.price_coins }}</span>
                  <span class="view-price-unit">WCoin</span>
                </div>
              </div>
            </div>
          </div>

          <!-- 4 строка: описание -->
          <div class="view-desc-section">
            <div class="view-desc-label">{{ t('purchases.fields.desc') }}</div>
            <div class="view-desc-panel">
              <p v-if="viewItem.description" class="view-desc-content">{{ viewItem.description }}</p>
              <span v-else class="view-desc-empty text-muted">Описание не указано</span>
            </div>
          </div>
        </div>

        <div class="modal-footer">
          <button type="button" class="btn-cancel" @click="closeViewModal">
            {{ t('common.close') }}
          </button>
          <button type="button" class="btn-primary-action btn-edit-action" @click="editFromView">
            <Edit2 class="icon-xs" />
            <span>{{ t('common.edit') }}</span>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive, inject, computed, type Ref } from 'vue';
import { useRoute } from 'vue-router';
import {
  ShoppingBag,
  Plus,
  Coins,
  CheckCircle,
  XCircle,
  Edit2,
  Trash2,
  X,
  UploadCloud,
  AlertCircle,
  Loader2,
  Copy,
  Check,
} from 'lucide-vue-next';
import { useI18n } from 'vue-i18n';
import { formatDate, formatDateTime } from '@/shared/lib';
import { getMediaUrl } from '@/entities/project';
import {
  listGameItems,
  createGameItem,
  updateGameItem,
  deleteGameItem,
  uploadItemImage,
  type GameItem,
} from '@/entities/purchases';
import type { Project } from '@/shared/types';
import { EmptyState } from '@/shared/ui';


const route = useRoute();
const projectId = String(route.params.id || '');
const { t } = useI18n();

const sharedProject = inject<Ref<Project | null> | null>('project', null);
const isLocked = computed<boolean>(() => sharedProject?.value?.is_under_review === true);

const items = ref<GameItem[]>([]);
const loading = ref<boolean>(true);
const saving = ref<boolean>(false);
const deleting = ref<boolean>(false);
const showModal = ref<boolean>(false);
const isEditing = ref<boolean>(false);
const modalError = ref<string>('');
const itemToDelete = ref<GameItem | null>(null);
const viewItem = ref<GameItem | null>(null);
const copiedId = ref<boolean>(false);
let copiedTimeout: ReturnType<typeof setTimeout> | null = null;

const fileInputRef = ref<HTMLInputElement | null>(null);
const selectedFile = ref<File | null>(null);
const iconPreview = ref<string>('');

interface FormState {
  game_item_id: string;
  name: string;
  description: string;
  image_url: string;
  price_coins: number;
  is_active: boolean;
}

const form = reactive<FormState>({
  game_item_id: '',
  name: '',
  description: '',
  image_url: '',
  price_coins: 50,
  is_active: true,
});

async function loadItems(): Promise<void> {
  loading.value = true;
  try {
    const data = await listGameItems(projectId);
    items.value = data || [];
  } catch (err) {
    console.error('Failed to load game items:', err);
  } finally {
    loading.value = false;
  }
}

function openCreateModal(): void {
  isEditing.value = false;
  modalError.value = '';
  selectedFile.value = null;
  iconPreview.value = '';
  form.game_item_id = '';
  form.name = '';
  form.description = '';
  form.image_url = '';
  form.price_coins = 50;
  form.is_active = true;
  showModal.value = true;
}

function openEditModal(item: GameItem): void {
  isEditing.value = true;
  modalError.value = '';
  selectedFile.value = null;
  form.game_item_id = item.game_item_id;
  form.name = item.name;
  form.description = item.description || '';
  form.image_url = item.image_url || '';
  form.price_coins = item.price_coins;
  form.is_active = item.is_active !== false;
  iconPreview.value = item.image_url ? (getMediaUrl(item.image_url, projectId) || '') : '';
  showModal.value = true;
}

function closeModal(): void {
  showModal.value = false;
  selectedFile.value = null;
  iconPreview.value = '';
}

function onFileSelected(event: Event): void {
  const target = event.target as HTMLInputElement;
  const file = target.files?.[0];
  if (!file) return;

  if (file.size > 2 * 1024 * 1024) {
    modalError.value = 'Размер файла не должен превышать 2 МБ';
    return;
  }

  selectedFile.value = file;
  iconPreview.value = URL.createObjectURL(file);
}

function onImageError(e: Event): void {
  const target = e.target as HTMLElement;
  if (target) {
    target.style.display = 'none';
  }
}

async function saveItem(): Promise<void> {
  modalError.value = '';

  const idPattern = /^[a-zA-Z0-9_-]+$/;
  if (!form.game_item_id || !idPattern.test(form.game_item_id)) {
    modalError.value = 'Идентификатор может содержать только буквы латиницы, цифры, дефис и подчеркивание';
    return;
  }
  if (!form.name.trim()) {
    modalError.value = 'Название товара обязательно для заполнения';
    return;
  }
  if (form.price_coins <= 0) {
    modalError.value = 'Цена должна быть положительным числом';
    return;
  }

  saving.value = true;
  try {
    let savedItem: GameItem;
    if (isEditing.value) {
      savedItem = await updateGameItem(projectId, form.game_item_id, form);
    } else {
      savedItem = await createGameItem(projectId, form);
    }

    // Если был выбран файл иконки, загружаем его
    if (selectedFile.value) {
      try {
        const uploadRes = await uploadItemImage(projectId, form.game_item_id, selectedFile.value);
        if (uploadRes?.file_path) {
          savedItem.image_url = uploadRes.file_path;
        }
      } catch (uploadErr) {
        console.warn('Failed to upload item icon image:', uploadErr);
      }
    }

    closeModal();
    await loadItems();
  } catch (err: any) {
    console.error('Save item error:', err);
    modalError.value = err.response?.data?.message || err.message || 'Ошибка сохранения товара';
  } finally {
    saving.value = false;
  }
}

function openDeleteConfirm(item: GameItem): void {
  itemToDelete.value = item;
}

async function confirmDelete(): Promise<void> {
  if (!itemToDelete.value) return;
  deleting.value = true;
  try {
    await deleteGameItem(projectId, itemToDelete.value.game_item_id);
    itemToDelete.value = null;
    await loadItems();
  } catch (err) {
    console.error('Delete item error:', err);
  } finally {
    deleting.value = false;
  }
}

function openViewModal(item: GameItem): void {
  viewItem.value = item;
  copiedId.value = false;
}

function closeViewModal(): void {
  viewItem.value = null;
  copiedId.value = false;
  if (copiedTimeout) {
    clearTimeout(copiedTimeout);
    copiedTimeout = null;
  }
}

function copyItemId(id: string): void {
  navigator.clipboard.writeText(id).then(() => {
    copiedId.value = true;
    if (copiedTimeout) clearTimeout(copiedTimeout);
    copiedTimeout = setTimeout(() => {
      copiedId.value = false;
    }, 2000);
  });
}

function editFromView(): void {
  if (!viewItem.value) return;
  const item = viewItem.value;
  closeViewModal();
  openEditModal(item);
}

onMounted(() => {
  loadItems();
});
</script>

<style scoped>
.purchases-page {
  padding: 24px 32px;
  display: flex;
  flex-direction: column;
  gap: 16px;
  max-width: 1100px;
  margin: 0 auto;
  width: 100%;
  box-sizing: border-box;
}

.purchases-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  min-height: 38px;
}

.rate-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  background: rgba(88, 166, 255, 0.12);
  border: 1px solid rgba(88, 166, 255, 0.28);
  color: var(--primary, #58a6ff);
  font-size: 13px;
  font-weight: 600;
  padding: 4px 12px;
  border-radius: 20px;
  white-space: nowrap;
}

.btn-primary-action {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  background: var(--primary, #58a6ff);
  color: #ffffff;
  font-weight: 500;
  font-size: 13px;
  padding: 8px 16px;
  border-radius: var(--radius-sm, 6px);
  border: none;
  cursor: pointer;
  transition: background-color 0.15s ease;
}

.btn-primary-action:hover {
  background: var(--primary-hover, #79c0ff);
}

/* Карточка таблицы внутри проекта */
.purchases-table-card {
  background: var(--bg-card, #161b22);
  border: 1px solid var(--border, #30363d);
  border-radius: var(--radius-md, 8px);
  overflow: hidden;
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 20px;
  text-align: center;
  gap: 12px;
}

.empty-icon-wrap {
  width: 56px;
  height: 56px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.04);
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 4px;
}

.table-wrapper {
  overflow-x: auto;
  width: 100%;
}

.project-table {
  width: 100%;
  border-collapse: collapse;
  text-align: left;
  table-layout: fixed;
  min-width: 820px;
}

.project-table th {
  background: rgba(255, 255, 255, 0.02);
  color: var(--text-tertiary, #8b949e);
  font-size: 13px;
  font-weight: 500;
  padding: 12px 16px;
  border-bottom: 1px solid var(--border, #30363d);
  white-space: nowrap;
}

.project-table th.col-item { width: 22%; }
.project-table th.col-id { width: 16%; }
.project-table th.col-desc { width: 24%; }
.project-table th.col-price { width: 14%; }
.project-table th.col-status { width: 14%; white-space: nowrap; }
.project-table th.col-actions { width: 10%; text-align: right; padding-right: 16px; white-space: nowrap; }

.project-table td {
  padding: 12px 16px;
  border-bottom: 1px solid var(--border, #21262d);
  vertical-align: middle;
}

.project-table td.col-status {
  white-space: nowrap;
}

.project-table td.col-actions {
  text-align: right;
  padding-right: 16px;
  white-space: nowrap;
}

.table-row {
  transition: background-color 0.15s ease;
}

.table-row-clickable {
  cursor: pointer;
}

.table-row-clickable:hover {
  background: rgba(255, 255, 255, 0.035);
}

.project-table tr:last-child td {
  border-bottom: none;
}

.item-meta {
  display: flex;
  align-items: center;
  gap: 12px;
  min-width: 0;
}

.item-icon-box {
  width: 36px;
  height: 36px;
  border-radius: var(--radius-sm, 6px);
  background: var(--bg-tertiary, #21262d);
  border: 1px solid var(--border, #30363d);
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  flex-shrink: 0;
}

.item-icon-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.item-icon-placeholder {
  display: flex;
  align-items: center;
  justify-content: center;
}

.item-name {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-main, #f0f6fc);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.item-desc-cell {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  text-overflow: ellipsis;
  word-break: break-word;
  line-height: 1.4;
  font-size: 13px;
  color: var(--text-muted, #8b949e);
}

.item-id-text {
  font-size: 13px;
  color: var(--text-muted, #8b949e);
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.price-val {
  display: inline-flex;
  align-items: baseline;
  gap: 4px;
  font-weight: 600;
  color: var(--text-main, #f0f6fc);
  font-size: 13px;
}

.price-num {
  font-size: 14px;
}

.price-unit {
  font-size: 12px;
  color: var(--text-muted, #8b949e);
  font-weight: 400;
}

/* Статус: чистая типографика (единый стиль) */
.status-text {
  font-size: 13px;
  font-weight: 500;
  white-space: nowrap;
}

.status-text.status-active {
  color: #2ecc71;
}

.status-text.status-inactive {
  color: var(--text-muted, #8b949e);
}

.row-actions {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 4px;
}

.btn-icon {
  background: transparent;
  border: none;
  color: var(--text-muted, #8b949e);
  cursor: pointer;
  padding: 6px;
  border-radius: 4px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  transition: all 0.15s ease;
}

.btn-icon:hover {
  color: var(--text-main, #f0f6fc);
  background: rgba(255, 255, 255, 0.06);
}

.btn-icon.text-danger-hover:hover {
  color: #f85149;
  background: rgba(248, 81, 73, 0.1);
}

/* Модальное окно */
.modal-backdrop {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.7);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 999;
  padding: 20px;
}

.modal-card {
  width: 100%;
  max-width: 580px;
  background: var(--bg-card, #161b22);
  border: 1px solid var(--border, #30363d);
  border-radius: var(--radius-lg, 10px);
  overflow: hidden;
  box-shadow: 0 20px 40px rgba(0, 0, 0, 0.5);
  animation: modal-pop 0.2s ease-out;
}

.modal-card.modal-sm {
  max-width: 420px;
}

@keyframes modal-pop {
  from {
    opacity: 0;
    transform: scale(0.96);
  }
  to {
    opacity: 1;
    transform: scale(1);
  }
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 20px;
  border-bottom: 1px solid var(--border, #30363d);
  background: var(--bg-card, #161b22);
}

.modal-title-group {
  display: flex;
  align-items: center;
  gap: 10px;
}

.modal-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-main, #f0f6fc);
  margin: 0;
}

.btn-close {
  background: transparent;
  border: none;
  color: var(--text-muted, #8b949e);
  cursor: pointer;
  padding: 4px;
  border-radius: 4px;
  transition: all 0.15s ease;
}

.btn-close:hover {
  color: var(--text-main, #f0f6fc);
  background: var(--bg-tertiary, #21262d);
}

.modal-body {
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.modal-body-p {
  padding: 20px;
  font-size: 14px;
  line-height: 1.5;
  color: var(--text-secondary, #9ca3af);
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.form-label {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-main, #f0f6fc);
}

.req {
  color: var(--danger, #f85149);
}

.label-with-counter {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.char-counter {
  font-size: 11px;
  color: var(--text-muted, #8b949e);
}

.counter-warn {
  color: var(--danger, #f85149);
}

.form-input,
.form-textarea {
  background: var(--bg-secondary, #0d1117);
  border: 1px solid var(--border, #30363d);
  border-radius: var(--radius-sm, 6px);
  padding: 9px 12px;
  color: var(--text-main, #f0f6fc);
  font-size: 13px;
  font-family: inherit;
  transition: border-color 0.15s ease, box-shadow 0.15s ease;
}

.form-textarea {
  resize: none;
  min-height: 76px;
  max-width: 100%;
}

.form-input:focus,
.form-textarea:focus {
  outline: none;
  border-color: var(--primary, #58a6ff);
  box-shadow: 0 0 0 2px rgba(88, 166, 255, 0.2);
}

.form-input:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

/* Скрываем стандартные стрелки number input для чистого вида */
input[type="number"]::-webkit-inner-spin-button,
input[type="number"]::-webkit-outer-spin-button {
  -webkit-appearance: none;
  margin: 0;
}
input[type="number"] {
  -moz-appearance: textfield;
}

.input-with-suffix {
  display: flex;
  position: relative;
  align-items: center;
}

.input-with-suffix .form-input {
  width: 100%;
  padding-right: 70px;
}

.input-suffix {
  position: absolute;
  right: 12px;
  font-size: 12px;
  color: var(--text-muted, #8b949e);
  font-weight: 500;
  pointer-events: none;
}

/* Загрузка иконки */
.icon-upload-row {
  display: flex;
  align-items: center;
  gap: 16px;
}

.icon-preview-box {
  width: 64px;
  height: 64px;
  border-radius: var(--radius-sm, 6px);
  background: var(--bg-secondary, #0d1117);
  border: 1px solid var(--border, #30363d);
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  flex-shrink: 0;
}

.icon-preview-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.icon-preview-empty {
  display: flex;
  align-items: center;
  justify-content: center;
}

.icon-upload-ctrl {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.hidden-file-input {
  display: none;
}

.btn-secondary-sm {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  background: var(--bg-tertiary, #21262d);
  border: 1px solid var(--border, #30363d);
  color: var(--text-main, #f0f6fc);
  font-size: 12px;
  padding: 6px 12px;
  border-radius: var(--radius-sm, 6px);
  cursor: pointer;
  align-self: flex-start;
  transition: all 0.15s ease;
}

.btn-secondary-sm:hover {
  background: var(--bg-hover, #30363d);
  border-color: var(--border-secondary, #484f58);
}

.upload-hint {
  font-size: 11px;
  color: var(--text-muted, #8b949e);
}

/* Чекбокс активности */
.form-group-checkbox {
  padding-top: 4px;
}

.checkbox-row {
  display: flex;
  align-items: center;
  gap: 10px;
  cursor: pointer;
  user-select: none;
}

.styled-checkbox {
  appearance: none;
  -webkit-appearance: none;
  width: 18px;
  height: 18px;
  border: 1px solid var(--border, #30363d);
  border-radius: 4px;
  background: var(--bg-secondary, #0d1117);
  cursor: pointer;
  outline: none;
  transition: all 0.15s ease;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  position: relative;
}

.styled-checkbox:hover {
  border-color: var(--primary, #58a6ff);
}

.styled-checkbox:focus-visible {
  border-color: var(--primary, #58a6ff);
  box-shadow: 0 0 0 2px rgba(88, 166, 255, 0.25);
}

.styled-checkbox:checked {
  background: var(--primary, #58a6ff);
  border-color: var(--primary, #58a6ff);
}

.styled-checkbox:checked::after {
  content: '';
  width: 5px;
  height: 9px;
  border: solid #ffffff;
  border-width: 0 2px 2px 0;
  transform: rotate(45deg) translate(-1px, -1px);
}

.checkbox-title {
  display: block;
  font-size: 13px;
  font-weight: 500;
  color: var(--text-main, #f0f6fc);
}

.form-error-alert {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 14px;
  border-radius: var(--radius-sm, 6px);
  background: rgba(239, 68, 68, 0.1);
  border: 1px solid rgba(239, 68, 68, 0.3);
  color: #ef4444;
  font-size: 12px;
}

.modal-footer {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 10px;
  padding: 14px 20px;
  border-top: 1px solid var(--border, #30363d);
}

.modal-body .modal-footer {
  padding: 14px 0 0 0;
}

.btn-cancel {
  background: transparent;
  border: 1px solid var(--border, #30363d);
  color: var(--text-muted, #8b949e);
  padding: 8px 16px;
  border-radius: var(--radius-sm, 6px);
  font-size: 13px;
  cursor: pointer;
  transition: all 0.15s ease;
}

.btn-cancel:hover {
  background: var(--bg-tertiary, #21262d);
  color: var(--text-main, #f0f6fc);
}

.btn-save-primary {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  background: var(--primary, #58a6ff);
  color: #ffffff;
  font-weight: 500;
  padding: 8px 18px;
  border-radius: var(--radius-sm, 6px);
  border: none;
  font-size: 13px;
  cursor: pointer;
  transition: background-color 0.15s ease;
}

.btn-save-primary:hover:not(:disabled) {
  background: var(--primary-hover, #79c0ff);
}

.btn-danger-confirm {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  background: #ef4444;
  color: #ffffff;
  font-weight: 600;
  padding: 8px 18px;
  border-radius: 6px;
  border: none;
  font-size: 13px;
  cursor: pointer;
}

.btn-danger-confirm:hover:not(:disabled) {
  opacity: 0.9;
}

/* Модальное окно: Детали товара (readonly) */
.modal-card.modal-view-card {
  max-width: 520px;
}

.view-details-body {
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 14px;
}

/* 1 строка: идентификатор (нейтральный, без синего) + иконка копирования */
.view-id-row {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  background: rgba(255, 255, 255, 0.025);
  border: 1px solid var(--border, #30363d);
  border-radius: var(--radius-sm, 6px);
}

.view-id-label {
  font-size: 12px;
  font-weight: 500;
  color: var(--text-muted, #8b949e);
  white-space: nowrap;
}

.view-id-code {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 13px;
  color: var(--text-secondary, #c9d1d9);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  flex: 1;
}

.btn-copy-icon-only {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 26px;
  height: 26px;
  background: transparent;
  border: 1px solid transparent;
  color: var(--text-muted, #8b949e);
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.15s ease;
  flex-shrink: 0;
  margin-left: auto;
}

.btn-copy-icon-only:hover {
  background: rgba(255, 255, 255, 0.08);
  color: var(--text-main, #f0f6fc);
  border-color: var(--border, #30363d);
}

/* 2-3 строка: слева картинка, справа статус+дата и название+стоимость */
.view-hero-card {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 14px 16px;
  background: rgba(255, 255, 255, 0.025);
  border: 1px solid var(--border, #30363d);
  border-radius: var(--radius-md, 8px);
}

.view-hero-media {
  width: 60px;
  height: 60px;
  border-radius: var(--radius-sm, 6px);
  background: var(--bg-tertiary, #21262d);
  border: 1px solid var(--border, #30363d);
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  flex-shrink: 0;
}

.view-media-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.view-media-placeholder {
  display: flex;
  align-items: center;
  justify-content: center;
}

.view-hero-content {
  display: flex;
  flex-direction: column;
  justify-content: center;
  gap: 6px;
  min-width: 0;
  flex: 1;
}

/* 2 строка (чуть меньше размером): статус, дата/время */
.view-sub-row {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 12px;
}

.view-meta-date {
  font-size: 12px;
  color: var(--text-muted, #8b949e);
  white-space: nowrap;
}

.status-indicator-pill {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  font-size: 11px;
  font-weight: 500;
  padding: 2px 8px;
  border-radius: 12px;
  width: fit-content;
  white-space: nowrap;
}

.status-indicator-pill.pill-active {
  background: rgba(46, 204, 113, 0.12);
  color: #2ecc71;
}

.status-indicator-pill.pill-inactive {
  background: rgba(139, 148, 158, 0.12);
  color: #8b949e;
}

.status-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: currentColor;
}

/* 3 строка: название и стоимость */
.view-main-row {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 12px;
}

.view-item-name {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: var(--text-main, #f0f6fc);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.view-price-clean {
  display: inline-flex;
  align-items: baseline;
  gap: 4px;
  font-weight: 700;
  color: var(--text-main, #f0f6fc);
  white-space: nowrap;
  flex-shrink: 0;
}

.view-price-num {
  font-size: 16px;
  font-weight: 700;
}

.view-price-unit {
  font-size: 12px;
  font-weight: 400;
  color: var(--text-muted, #8b949e);
}

/* 4 строка: описание */
.view-desc-section {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.view-desc-label {
  font-size: 12px;
  font-weight: 500;
  color: var(--text-muted, #8b949e);
}

.view-desc-panel {
  background: rgba(255, 255, 255, 0.02);
  border: 1px solid var(--border, #30363d);
  border-radius: var(--radius-sm, 6px);
  padding: 10px 12px;
  min-height: 48px;
  max-height: 160px;
  overflow-y: auto;
}

.view-desc-content {
  margin: 0;
  font-size: 13px;
  line-height: 1.5;
  color: var(--text-main, #f0f6fc);
  white-space: pre-wrap;
  word-break: break-word;
}

.view-desc-empty {
  font-size: 13px;
  font-style: italic;
}

.btn-edit-action {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 7px 16px;
  font-size: 13px;
}

.state-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px;
  gap: 12px;
  color: var(--text-secondary, #9ca3af);
}

.spin {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}
.purchases-lock-banner {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 14px 18px;
  margin-bottom: 20px;
  background: rgba(210, 153, 34, 0.12);
  border: 1px solid rgba(210, 153, 34, 0.35);
  border-radius: var(--radius-md, 8px);
  color: var(--text-main, #f0f6fc);
}

.lock-banner-title {
  font-weight: 600;
  font-size: 0.95rem;
  color: #d29922;
  margin-bottom: 4px;
}

.lock-banner-desc {
  font-size: 0.86rem;
  color: var(--text-muted, #8b949e);
  line-height: 1.4;
}
</style>
