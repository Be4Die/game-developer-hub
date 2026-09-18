<template>
  <div class="tab-content tab-fade-in purchases-page">
    <!-- Шапка страницы -->
    <div class="page-header-row">
      <div>
        <h1 class="page-title">{{ t('purchases.title') }}</h1>
        <p class="page-subtitle text-muted">{{ t('purchases.subtitle') }}</p>
      </div>
      <button class="btn-primary-action" @click="openCreateModal">
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
      <div v-if="items.length === 0" class="empty-state">
        <div class="empty-icon-wrap">
          <ShoppingBag class="icon-xl text-muted" />
        </div>
        <h3>{{ t('purchases.empty.title') }}</h3>
        <p class="text-muted">{{ t('purchases.empty.desc') }}</p>
        <button class="btn-primary-action" @click="openCreateModal">
          <Plus class="icon-sm" />
          <span>{{ t('purchases.actions.addItem') }}</span>
        </button>
      </div>

      <!-- Таблица товаров -->
      <div v-else class="table-wrapper">
        <table class="data-table">
          <thead>
            <tr>
              <th class="col-item">{{ t('purchases.table.item') }}</th>
              <th class="col-id">{{ t('purchases.table.itemId') }}</th>
              <th class="col-price">{{ t('purchases.table.price') }}</th>
              <th class="col-status">{{ t('purchases.table.status') }}</th>
              <th class="col-actions"></th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="item in items" :key="item.id || item.game_item_id" class="table-row">
              <td class="col-item">
                <div class="item-meta">
                  <div class="item-icon-box">
                    <img
                      v-if="item.image_url"
                      :src="getMediaUrl(item.image_url)"
                      alt="Icon"
                      class="item-icon-img"
                      @error="onImageError"
                    />
                    <div v-else class="item-icon-placeholder">
                      <ShoppingBag class="icon-xs text-muted" />
                    </div>
                  </div>
                  <div class="item-text">
                    <div class="item-name">{{ item.name }}</div>
                    <div v-if="item.description" class="item-desc text-muted" :title="item.description">
                      {{ item.description }}
                    </div>
                  </div>
                </div>
              </td>
              <td class="col-id">
                <code class="item-slug-code">{{ item.game_item_id }}</code>
              </td>
              <td class="col-price">
                <div class="price-val">
                  <Coins class="icon-xs text-primary" />
                  <span class="price-num">{{ item.price_coins }}</span>
                  <span class="price-unit">WCoin</span>
                </div>
              </td>
              <td class="col-status">
                <span v-if="item.is_active" class="badge-status-active">
                  <CheckCircle class="icon-xs" />
                  <span>Активен</span>
                </span>
                <span v-else class="badge-status-inactive">
                  <XCircle class="icon-xs" />
                  <span>Скрыт</span>
                </span>
              </td>
              <td class="col-actions">
                <div class="actions-row">
                  <button
                    class="btn-icon-sm"
                    title="Редактировать товар"
                    @click="openEditModal(item)"
                  >
                    <Edit2 class="icon-xs" />
                  </button>
                  <button
                    class="btn-icon-sm btn-delete"
                    title="Удалить товар"
                    @click="openDeleteConfirm(item)"
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

    <!-- Краткая справка внизу страницы: бейдж и 1 строка пояснения -->
    <div class="purchases-footer-note">
      <div class="rate-badge">
        <Coins class="icon-xs text-primary" />
        <span>1 WCoin = 1 ₽</span>
      </div>
      <span class="footer-note-text text-muted">
        {{ t('purchases.footerNote') }}
      </span>
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
                  @click="$refs.fileInputRef.click()"
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
              <div class="checkbox-text">
                <span class="checkbox-title">{{ t('purchases.fields.isActive') }}</span>
                <span class="checkbox-sub text-muted">{{ t('purchases.hints.isActive') }}</span>
              </div>
            </label>
          </div>

          <!-- Футер модалки -->
          <div class="modal-footer">
            <button type="button" class="btn-cancel" @click="closeModal">
              {{ t('common.cancel') }}
            </button>
            <button type="submit" class="btn-save-primary" :disabled="saving">
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
          <button type="button" class="btn-danger-confirm" :disabled="deleting" @click="confirmDelete">
            <Loader2 v-if="deleting" class="icon-xs spin" />
            <span>{{ deleting ? t('common.deleting') : t('common.delete') }}</span>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, reactive } from 'vue';
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
} from 'lucide-vue-next';
import { useI18n } from 'vue-i18n';
import { getMediaUrl } from '@/entities/project';
import {
  listGameItems,
  createGameItem,
  updateGameItem,
  deleteGameItem,
  uploadItemImage,
} from '@/entities/purchases';

const route = useRoute();
const projectId = route.params.id;
const { t } = useI18n();

const items = ref([]);
const loading = ref(true);
const saving = ref(false);
const deleting = ref(false);
const showModal = ref(false);
const isEditing = ref(false);
const modalError = ref('');
const itemToDelete = ref(null);

const fileInputRef = ref(null);
const selectedFile = ref(null);
const iconPreview = ref('');

const form = reactive({
  game_item_id: '',
  name: '',
  description: '',
  image_url: '',
  price_coins: 50,
  is_active: true,
});

async function loadItems() {
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

function openCreateModal() {
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

function openEditModal(item) {
  isEditing.value = true;
  modalError.value = '';
  selectedFile.value = null;
  form.game_item_id = item.game_item_id;
  form.name = item.name;
  form.description = item.description || '';
  form.image_url = item.image_url || '';
  form.price_coins = item.price_coins;
  form.is_active = item.is_active;
  iconPreview.value = item.image_url ? getMediaUrl(item.image_url) : '';
  showModal.value = true;
}

function closeModal() {
  showModal.value = false;
  selectedFile.value = null;
  iconPreview.value = '';
}

function onFileSelected(event) {
  const file = event.target.files?.[0];
  if (!file) return;

  if (file.size > 2 * 1024 * 1024) {
    modalError.value = 'Размер файла не должен превышать 2 МБ';
    return;
  }

  selectedFile.value = file;
  iconPreview.value = URL.createObjectURL(file);
}

function onImageError(e) {
  e.target.style.display = 'none';
}

async function saveItem() {
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
    let savedItem;
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
  } catch (err) {
    console.error('Save item error:', err);
    modalError.value = err.response?.data?.message || err.message || 'Ошибка сохранения товара';
  } finally {
    saving.value = false;
  }
}

function openDeleteConfirm(item) {
  itemToDelete.value = item;
}

async function confirmDelete() {
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

onMounted(() => {
  loadItems();
});
</script>

<style scoped>
.purchases-page {
  padding: 24px 32px;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.page-header-row {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
}

.page-title {
  font-size: 22px;
  font-weight: 700;
  color: var(--text-main, #f0f6fc);
  margin: 0;
}

.page-subtitle {
  font-size: 13px;
  margin-top: 4px;
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

/* Подвал: краткая справка */
.purchases-footer-note {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 16px;
  background: var(--bg-card, #161b22);
  border: 1px solid var(--border, #30363d);
  border-radius: var(--radius-md, 8px);
  font-size: 13px;
}

.rate-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  background: rgba(88, 166, 255, 0.12);
  border: 1px solid rgba(88, 166, 255, 0.28);
  color: var(--primary, #58a6ff);
  font-size: 12px;
  font-weight: 600;
  padding: 3px 10px;
  border-radius: 20px;
  white-space: nowrap;
}

.footer-note-text {
  color: var(--text-muted, #8b949e);
  font-size: 12px;
}

/* Карточка таблицы */
.purchases-table-card {
  background: var(--bg-card, #161b22);
  border: 1px solid var(--border, #30363d);
  border-radius: var(--radius-lg, 10px);
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
}

.data-table {
  width: 100%;
  border-collapse: collapse;
  text-align: left;
  font-size: 13px;
}

.data-table th {
  background: var(--bg-tertiary, #21262d);
  color: var(--text-muted, #8b949e);
  font-weight: 600;
  padding: 12px 16px;
  border-bottom: 1px solid var(--border, #30363d);
}

.data-table td {
  padding: 14px 16px;
  border-bottom: 1px solid var(--border, #30363d);
  vertical-align: middle;
}

.table-row:hover {
  background: rgba(255, 255, 255, 0.02);
}

.item-meta {
  display: flex;
  align-items: center;
  gap: 12px;
}

.item-icon-box {
  width: 40px;
  height: 40px;
  border-radius: var(--radius-sm, 6px);
  background: var(--bg-secondary, #0d1117);
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

.item-text {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.item-name {
  font-weight: 600;
  color: var(--text-main, #f0f6fc);
}

.item-desc {
  font-size: 12px;
  color: var(--text-muted, #8b949e);
  max-width: 320px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.item-slug-code {
  background: rgba(88, 166, 255, 0.08);
  border: 1px solid rgba(88, 166, 255, 0.2);
  padding: 3px 8px;
  border-radius: 4px;
  font-size: 12px;
  color: var(--primary, #58a6ff);
  font-family: monospace;
}

.price-val {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-weight: 600;
  color: var(--text-main, #f0f6fc);
}

.price-num {
  font-size: 14px;
}

.price-unit {
  font-size: 11px;
  color: var(--text-muted, #8b949e);
  font-weight: 500;
}

.badge-status-active {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  background: rgba(16, 185, 129, 0.12);
  color: #10b981;
  padding: 3px 8px;
  border-radius: 12px;
  font-size: 11px;
  font-weight: 600;
}

.badge-status-inactive {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  background: rgba(156, 163, 175, 0.12);
  color: #9ca3af;
  padding: 3px 8px;
  border-radius: 12px;
  font-size: 11px;
  font-weight: 500;
}

.actions-row {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 6px;
}

.btn-icon-sm {
  background: transparent;
  border: none;
  color: var(--text-muted, #8b949e);
  cursor: pointer;
  padding: 6px;
  border-radius: 4px;
  transition: all 0.15s ease;
}

.btn-icon-sm:hover {
  background: var(--bg-tertiary, #21262d);
  color: var(--text-main, #f0f6fc);
}

.btn-icon-sm.btn-delete:hover {
  background: rgba(239, 68, 68, 0.15);
  color: #ef4444;
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
  max-width: 520px;
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
  align-items: flex-start;
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
  margin-top: 2px;
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
  font-weight: 600;
  color: var(--text-main, #f0f6fc);
}

.checkbox-sub {
  display: block;
  font-size: 11px;
  margin-top: 2px;
  color: var(--text-muted, #8b949e);
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
  padding-top: 14px;
  border-top: 1px solid var(--border, #30363d);
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
</style>
