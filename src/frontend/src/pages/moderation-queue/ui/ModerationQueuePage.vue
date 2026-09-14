<template>
  <div class="moderation-page-container">
    <div class="main-content-wrap">
      <!-- Панель фильтрации и поиска (компактный стиль консоли) -->
      <div class="filters-toolbar">
        <!-- Поиск по названию, ID, разработчику или обоснованию -->
        <div class="filter-field field-search">
          <label class="field-label">{{ t('common.search') }}</label>
          <div class="input-wrapper">
            <input
              v-model="searchQuery"
              type="text"
              placeholder="Поиск по названию, ID, разработчику или обоснованию..."
              class="filter-input"
            />
            <button
              v-if="searchQuery"
              class="clear-input-btn"
              title="Очистить"
              @click="searchQuery = ''"
            >
              <X class="icon-xs" />
            </button>
          </div>
        </div>

        <!-- Фильтр по типу заявки (справа от поиска) -->
        <div class="filter-field field-type">
          <label class="field-label">Тип заявки</label>
          <div class="select-wrapper">
            <select v-model="typeFilter" class="filter-select">
              <option value="all">Все типы</option>
              <option value="publication">Публикация</option>
              <option value="server">Серверы</option>
            </select>
            <ChevronDown class="icon-xs select-arrow" />
          </div>
        </div>

        <!-- Фильтр по статусу -->
        <div class="filter-field field-status">
          <label class="field-label">{{ t('common.status') }}</label>
          <div class="select-wrapper">
            <select v-model="statusFilter" class="filter-select">
              <option value="all">{{ t('moderation.allActive') }}</option>
              <option value="pending">{{ t('moderation.pending') }}</option>
              <option value="in_review">{{ t('moderation.inReview') }}</option>
              <option value="my">{{ t('moderation.assignedToMe') }}</option>
            </select>
            <ChevronDown class="icon-xs select-arrow" />
          </div>
        </div>

        <!-- Фильтр по сетевому режиму (для проектов) -->
        <div v-if="typeFilter !== 'server'" class="filter-field field-mode">
          <label class="field-label">{{ t('projects.mode') }}</label>
          <div class="select-wrapper">
            <select v-model="modeFilter" class="filter-select">
              <option value="all">{{ t('common.all') }}</option>
              <option value="online">{{ t('projects.modeOnline') }}</option>
              <option value="offline">{{ t('projects.modeOffline') }}</option>
            </select>
            <ChevronDown class="icon-xs select-arrow" />
          </div>
        </div>

        <!-- Сортировка -->
        <div class="filter-field field-sort">
          <label class="field-label">{{ t('common.actions') }}</label>
          <div class="select-wrapper">
            <select v-model="sortBy" class="filter-select">
              <option value="newest">{{ t('stats.today') }} / {{ t('common.created') }} ↓</option>
              <option value="oldest">{{ t('common.created') }} ↑</option>
              <option value="title">{{ t('common.name') }} (A–Z)</option>
            </select>
            <ChevronDown class="icon-xs select-arrow" />
          </div>
        </div>

        <!-- Кнопка сброса фильтров -->
        <button
          v-if="searchQuery || statusFilter !== 'all' || modeFilter !== 'all' || sortBy !== 'newest' || typeFilter !== 'all'"
          class="btn-reset-filters"
          title="Сбросить фильтры"
          @click="resetFilters"
        >
          <RotateCcw class="icon-xs" />
          <span>{{ t('common.reset') }}</span>
        </button>

        <!-- Кнопка обновления -->
        <button class="btn-refresh" :disabled="loading" title="Обновить" @click="loadQueue">
          <RefreshCw class="icon-xs" :class="{ spin: loading }" />
          <span>{{ t('common.refresh') }}</span>
        </button>
      </div>

      <!-- Состояние загрузки -->
      <div v-if="loading" class="state-container loading-card">
        <div class="spinner-md"></div>
        <p>{{ t('common.loading') }}</p>
      </div>

      <!-- Пустой список заявок -->
      <div
        v-else-if="requests.length === 0 && !searchQuery && statusFilter === 'all'"
        class="state-container empty-card"
      >
        <div class="empty-icon-wrap">
          <CheckCircle2 class="icon-lg text-success" />
        </div>
        <h3>{{ t('moderation.noActiveRequests') }}</h3>
        <p>{{ t('moderation.emptyQueue') }}</p>
        <button class="btn-primary-sm" @click="loadQueue">
          <RefreshCw class="icon-xs" />
          <span>{{ t('common.refresh') }}</span>
        </button>
      </div>

      <!-- Пустой список по результатам поиска/фильтров -->
      <div v-else-if="filteredRequests.length === 0" class="state-container empty-card">
        <Search class="icon-md text-muted" />
        <h3>{{ t('common.empty') }}</h3>
        <p>{{ t('stats.noData') }}</p>
        <button class="btn-reset-filters" @click="resetFilters">
          {{ t('common.reset') }}
        </button>
      </div>

      <!-- Единая таблица активных заявок на модерацию -->
      <div v-else class="table-wrapper">
        <table class="moderation-table">
          <thead>
            <tr>
              <th class="col-game">Проект / Заявка</th>
              <th class="col-details">Предмет заявки</th>
              <th class="col-dev">{{ t('moderation.developerColumn') }}</th>
              <th class="col-date">{{ t('moderation.submittedColumn') }}</th>
              <th class="col-status">{{ t('common.status') }}</th>
              <th class="col-mod">{{ t('moderation.moderator') }}</th>
              <th class="col-actions"></th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="req in paginatedRequests"
              :key="req.type + '-' + req.id"
              class="table-row"
              @click="openProject(req.projectId)"
            >
              <!-- 1 колонка: Игра / Проект + тип заявки -->
              <td class="col-game">
                <div class="game-cell">
                  <div class="game-icon-box">
                    <img
                      v-if="req.iconPath"
                      :src="getMediaUrl(req.iconPath)"
                      alt="Icon"
                      class="game-icon-img"
                    />
                    <Server v-else-if="req.type === 'server'" class="icon-xs text-warning" />
                    <Gamepad2 v-else class="icon-xs text-muted" />
                  </div>
                  <div class="game-text">
                    <div class="game-title-row">
                      <span class="game-title" :title="req.title">
                        {{ req.title }}
                      </span>
                      <!-- Бейдж типа заявки (короткие понятные названия) -->
                      <span
                        class="type-pill"
                        :class="req.type === 'server' ? 'type-server' : 'type-project'"
                      >
                        {{ req.type === 'server' ? 'Серверы' : 'Публикация' }}
                      </span>
                      <!-- Режим онлайн/офлайн для проектов публикации -->
                      <span
                        v-if="req.type === 'project'"
                        class="mode-badge"
                        :class="req.isOnline ? 'mode-online' : 'mode-offline'"
                        :title="req.isOnline ? t('projects.modeOnline') : t('projects.modeOffline')"
                      >
                        <Globe v-if="req.isOnline" class="icon-xxs" />
                        <Gamepad2 v-else class="icon-xxs" />
                        <span>{{ req.isOnline ? t('projects.modeOnline') : t('projects.modeOffline') }}</span>
                      </span>
                    </div>
                  </div>
                </div>
              </td>

              <!-- 2 колонка: Предмет заявки (версия билда или обоснование серверов) -->
              <td class="col-details">
                <template v-if="req.type === 'project'">
                  <span v-if="req.version" class="version-badge">
                    v{{ req.version }}
                  </span>
                  <span v-else class="text-muted text-sm">—</span>
                </template>
                <template v-else>
                  <div class="server-details-cell">
                    <div class="quota-chips-line">
                      <span class="quota-chip" title="Запрошенное число инстансов">
                        {{ req.maxInstances }} инст.
                      </span>
                      <span class="quota-chip" title="Суммарный лимит CPU">
                        {{ formatCpu(req.maxTotalCpuMillis, true) }}
                      </span>
                      <span class="quota-chip" title="Суммарный лимит RAM">
                        {{ formatMemory(req.maxTotalMemoryMb, true) }}
                      </span>
                    </div>
                    <p class="server-reason-snippet" :title="req.reason">
                      «{{ req.reason || 'Запрос доступа к серверам платформы' }}»
                    </p>
                  </div>
                </template>
              </td>

              <!-- 3 колонка: Разработчик -->
              <td class="col-dev">
                <div class="dev-cell" :title="req.ownerId">
                  <User class="icon-xs text-muted" />
                  <span class="dev-name">{{ req.ownerId ? getUserDisplayName(req.ownerId) : '—' }}</span>
                </div>
              </td>

              <!-- 4 колонка: Дата подачи -->
              <td class="col-date">
                <span class="date-text">
                  {{ formatDateTime(req.submittedAt) }}
                </span>
              </td>

              <!-- 5 колонка: Статус -->
              <td class="col-status">
                <span class="status-pill" :class="reqStatusClass(req.status)">
                  {{ reqStatusLabel(req.status) }}
                </span>
              </td>

              <!-- 6 колонка: Модератор / Решение -->
              <td class="col-mod">
                <!-- Для серверов, если одобрено/отклонено -->
                <div v-if="req.type === 'server' && isApproved(req.status)" class="mod-decision-box text-success">
                  <strong>Квота: {{ req.maxInstances }} инст.</strong>
                  <span class="quota-subline">
                    {{ formatCpu(req.maxTotalCpuMillis, true) }} / {{ formatMemory(req.maxTotalMemoryMb, true) }}
                  </span>
                  <span
                    v-if="req.moderatorComment"
                    class="mod-comment-cell"
                    :title="req.moderatorComment"
                  >
                    «{{ req.moderatorComment }}»
                  </span>
                  <span v-if="req.moderatorId" class="sub-mod">{{ getUserDisplayName(req.moderatorId) }}</span>
                </div>
                <div v-else-if="req.type === 'server' && isRejected(req.status)" class="mod-decision-box text-danger">
                  <span class="rejection-text-cell" :title="req.rejectionReason">«{{ req.rejectionReason || 'Отказ' }}»</span>
                  <span v-if="req.moderatorId" class="sub-mod">{{ getUserDisplayName(req.moderatorId) }}</span>
                </div>
                <span v-else-if="req.moderatorId" class="mod-name" :title="req.moderatorId">
                  {{ getUserDisplayName(req.moderatorId) }}
                </span>
                <span v-else class="unassigned-text">
                  {{ t('moderation.notAssigned') }}
                </span>
              </td>

              <!-- 7 колонка: Действия -->
              <td class="col-actions" @click.stop>
                <div class="row-actions">
                  <!-- Действия для проектов -->
                  <template v-if="req.type === 'project'">
                    <button
                      v-if="isAdmin"
                      class="btn-inspect-sm"
                      title="Просмотр проекта"
                      @click="openProject(req.projectId)"
                    >
                      <Eye class="icon-xs" />
                      <span>Просмотр</span>
                    </button>
                    <template v-else>
                      <button
                        v-if="isPending(req.status)"
                        class="btn-claim-sm"
                        :disabled="claimingId === req.id"
                        :title="t('moderation.claimBtn')"
                        @click="claimAndOpen(req)"
                      >
                        <Loader2 v-if="claimingId === req.id" class="icon-xs spin" />
                        <CheckSquare v-else class="icon-xs" />
                        <span>{{ t('moderation.claimBtn') }}</span>
                      </button>

                      <button
                        v-else
                        class="btn-inspect-sm"
                        :title="t('moderation.continueBtn')"
                        @click="openProject(req.projectId)"
                      >
                        <ArrowRight class="icon-xs" />
                        <span>{{ t('moderation.continueBtn') }}</span>
                      </button>
                    </template>
                  </template>

                  <!-- Действия для серверов платформы -->
                  <template v-else>
                    <template v-if="isPending(req.status)">
                      <button
                        class="btn-approve-sm"
                        title="Одобрить доступ"
                        @click="openApproveServerModal(req)"
                      >
                        <Check class="icon-xs" />
                        <span>Одобрить</span>
                      </button>
                      <button
                        class="btn-reject-sm"
                        title="Отклонить заявку"
                        @click="openRejectServerModal(req)"
                      >
                        <X class="icon-xs" />
                        <span>Отклонить</span>
                      </button>
                    </template>
                    <button
                      v-else
                      class="btn-inspect-sm"
                      title="Просмотр проекта"
                      @click="openProject(req.projectId)"
                    >
                      <Eye class="icon-xs" />
                      <span>Просмотр</span>
                    </button>
                  </template>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Пагинация -->
      <div v-if="totalPages > 1" class="pagination-bar">
        <span class="page-info">
          {{ (currentPage - 1) * pageSize + 1 }}–{{
            Math.min(currentPage * pageSize, filteredRequests.length)
          }}
          из {{ filteredRequests.length }}
        </span>

        <div class="page-nav">
          <button class="page-nav-btn" :disabled="currentPage === 1" @click="currentPage = 1">
            <ChevronsLeft class="icon-sm" />
          </button>
          <button class="page-nav-btn" :disabled="currentPage === 1" @click="currentPage--">
            <ChevronLeft class="icon-sm" />
          </button>
          <span class="page-current">{{ currentPage }} / {{ totalPages }}</span>
          <button
            class="page-nav-btn"
            :disabled="currentPage === totalPages"
            @click="currentPage++"
          >
            <ChevronRight class="icon-sm" />
          </button>
          <button
            class="page-nav-btn"
            :disabled="currentPage === totalPages"
            @click="currentPage = totalPages"
          >
            <ChevronsRight class="icon-sm" />
          </button>

          <div class="page-size-wrap">
            <select v-model="pageSize" class="page-size-select">
              <option :value="10">10</option>
              <option :value="25">25</option>
              <option :value="50">50</option>
            </select>
            <ChevronDown class="icon-xs select-caret" />
          </div>
        </div>
      </div>

      <!-- Модальное окно: Одобрение доступа к серверам -->
      <div
        v-if="approvingServerTarget"
        class="modal-overlay"
        @click.self="approvingServerTarget = null"
      >
        <div class="modal-card modal-card-wide">
          <div class="modal-header">
            <h3>Одобрить доступ к серверам платформы</h3>
            <button class="btn-close" @click="approvingServerTarget = null">
              <X class="icon-sm" />
            </button>
          </div>
          <div class="modal-body">
            <div class="project-info-banner">
              <div class="project-info-title">
                Проект: <strong>{{ approvingServerTarget.title }}</strong> (ID: {{ approvingServerTarget.projectId }})
              </div>
              <p v-if="approvingServerTarget.reason" class="modal-reason-quote">
                <strong>Обоснование разработчика:</strong> «{{ approvingServerTarget.reason }}»
              </p>
              <div class="requested-summary-line">
                <span class="req-sum-label">Запрошено разработчиком:</span>
                <span class="req-sum-val">{{ approvingServerTarget.maxInstances }} инст.</span>
                <span class="req-sum-val">Всего CPU: {{ formatCpu(approvingServerTarget.maxTotalCpuMillis) }}</span>
                <span class="req-sum-val">Всего RAM: {{ formatMemory(approvingServerTarget.maxTotalMemoryMb) }}</span>
              </div>
            </div>

            <!-- Форма настройки утверждённой квоты -->
            <div class="moderator-quota-form">
              <div class="form-group mb-12">
                <label class="form-label">Утверждённое число серверов (инстансов) *</label>
                <input
                  v-model.number="approveServerQuota"
                  type="number"
                  min="1"
                  max="50"
                  class="form-input"
                />
                <span class="field-hint">Максимум одновременно работающих серверов на мощностях платформы</span>
              </div>

              <!-- Суммарные ресурсы проекта -->
              <div class="quota-group-card">
                <h4 class="quota-group-title">Суммарный лимит на весь проект</h4>
                <div class="quota-inputs-row">
                  <div class="form-group flex-1">
                    <div class="label-with-toggle">
                      <label class="form-label">Всего CPU (ядер)</label>
                      <label class="toggle-label">
                        <input type="checkbox" v-model="approveUnlimitedTotalCpu" />
                        <span>Без огр.</span>
                      </label>
                    </div>
                    <input
                      v-if="!approveUnlimitedTotalCpu"
                      v-model.number="approveTotalCpu"
                      type="number"
                      step="0.5"
                      min="0.5"
                      max="64"
                      class="form-input"
                    />
                    <div v-else class="unlimited-placeholder">∞ Без ограничений</div>
                  </div>

                  <div class="form-group flex-1">
                    <div class="label-with-toggle">
                      <label class="form-label">Всего RAM (МБ)</label>
                      <label class="toggle-label">
                        <input type="checkbox" v-model="approveUnlimitedTotalRam" />
                        <span>Без огр.</span>
                      </label>
                    </div>
                    <input
                      v-if="!approveUnlimitedTotalRam"
                      v-model.number="approveTotalRam"
                      type="number"
                      step="256"
                      min="256"
                      max="131072"
                      class="form-input"
                    />
                    <div v-else class="unlimited-placeholder">∞ Без ограничений</div>
                  </div>
                </div>
              </div>

              <!-- Лимиты на 1 инстанс -->
              <div class="quota-group-card">
                <h4 class="quota-group-title">Лимит на 1 отдельный инстанс (сервер)</h4>
                <div class="quota-inputs-row">
                  <div class="form-group flex-1">
                    <div class="label-with-toggle">
                      <label class="form-label">CPU на инстанс</label>
                      <label class="toggle-label">
                        <input type="checkbox" v-model="approveUnlimitedInstanceCpu" />
                        <span>Без огр.</span>
                      </label>
                    </div>
                    <input
                      v-if="!approveUnlimitedInstanceCpu"
                      v-model.number="approveInstanceCpu"
                      type="number"
                      step="0.5"
                      min="0.5"
                      max="32"
                      class="form-input"
                    />
                    <div v-else class="unlimited-placeholder">∞ Без ограничений</div>
                  </div>

                  <div class="form-group flex-1">
                    <div class="label-with-toggle">
                      <label class="form-label">RAM на инстанс (МБ)</label>
                      <label class="toggle-label">
                        <input type="checkbox" v-model="approveUnlimitedInstanceRam" />
                        <span>Без огр.</span>
                      </label>
                    </div>
                    <input
                      v-if="!approveUnlimitedInstanceRam"
                      v-model.number="approveInstanceRam"
                      type="number"
                      step="256"
                      min="256"
                      max="32768"
                      class="form-input"
                    />
                    <div v-else class="unlimited-placeholder">∞ Без ограничений</div>
                  </div>
                </div>
              </div>

              <!-- Комментарий модератора -->
              <div class="form-group">
                <label class="form-label">Комментарий модератора (необязательно)</label>
                <textarea
                  v-model="approveComment"
                  rows="2"
                  class="form-textarea"
                  placeholder="Например: Выделено на период закрытого тестирования..."
                ></textarea>
              </div>
            </div>

            <div v-if="serverReviewError" class="alert-error">
              <AlertCircle class="icon-xs" />
              <span>{{ serverReviewError }}</span>
            </div>
          </div>
          <div class="modal-footer">
            <button
              class="btn-secondary-sm"
              :disabled="submittingServerReview"
              @click="approvingServerTarget = null"
            >
              Отмена
            </button>
            <button
              class="btn-primary-sm"
              :disabled="submittingServerReview || approveServerQuota < 1"
              @click="submitApproveServer"
            >
              <Loader2 v-if="submittingServerReview" class="icon-xs spin" />
              <Check v-else class="icon-xs" />
              <span>Подтвердить одобрение</span>
            </button>
          </div>
        </div>
      </div>

      <!-- Модальное окно: Отклонение доступа к серверам -->
      <div
        v-if="rejectingServerTarget"
        class="modal-overlay"
        @click.self="rejectingServerTarget = null"
      >
        <div class="modal-card">
          <div class="modal-header">
            <h3>Отклонить заявку на серверы</h3>
            <button class="btn-close" @click="rejectingServerTarget = null">
              <X class="icon-sm" />
            </button>
          </div>
          <div class="modal-body">
            <p class="modal-intro">
              Проект: <strong>{{ rejectingServerTarget.title }}</strong>
            </p>
            <div class="form-group">
              <label class="form-label">Причина отказа для разработчика:</label>
              <textarea
                v-model="rejectServerReason"
                rows="3"
                class="form-textarea"
                placeholder="Укажите подробную причину отказа или рекомендации по оптимизации..."
              ></textarea>
            </div>
            <div v-if="serverReviewError" class="alert-error">
              <AlertCircle class="icon-xs" />
              <span>{{ serverReviewError }}</span>
            </div>
          </div>
          <div class="modal-footer">
            <button
              class="btn-secondary-sm"
              :disabled="submittingServerReview"
              @click="rejectingServerTarget = null"
            >
              Отмена
            </button>
            <button
              class="btn-danger-sm"
              :disabled="submittingServerReview || !rejectServerReason.trim()"
              @click="submitRejectServer"
            >
              <Loader2 v-if="submittingServerReview" class="icon-xs spin" />
              <X v-else class="icon-xs" />
              <span>Отклонить заявку</span>
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import { useI18n } from 'vue-i18n';
import {
  Search,
  X,
  ChevronDown,
  RotateCcw,
  RefreshCw,
  CheckCircle2,
  User,
  CheckSquare,
  ArrowRight,
  Loader2,
  ChevronLeft,
  ChevronRight,
  ChevronsLeft,
  ChevronsRight,
  Eye,
  Gamepad2,
  Globe,
  Server,
  Check,
  AlertCircle,
} from 'lucide-vue-next';
import {
  moderationApi,
  normalizeRequest,
  formatDateTime,
  formatCpu,
  formatMemory,
  REQUEST_STATUS,
} from '@/entities/moderation';
import { getMediaUrl, listProjects } from '@/entities/project';
import { useAuth, getUserDisplayName } from '@/entities/user';
import { showToast } from '@/shared/lib';

const { t } = useI18n();
const router = useRouter();
const route = useRoute();
const { state: authState } = useAuth();

const isAdmin = computed(() => {
  const r = authState.user?.role;
  return r === 'USER_ROLE_ADMIN' || r === 'admin' || r === 3;
});

const currentUserId = computed(() => authState.user?.id || authState.user?.email || '');

const requests = ref([]);
const loading = ref(true);
const claimingId = ref(null);

const typeFilter = ref(
  route.query.type === 'server' || route.query.type === 'servers'
    ? 'server'
    : route.query.type === 'project' || route.query.type === 'publication'
      ? 'publication'
      : 'all'
);
const searchQuery = ref('');
const statusFilter = ref('all');
const modeFilter = ref('all');
const sortBy = ref('newest');
const currentPage = ref(1);
const pageSize = ref(10);

// Серверные модальные окна
const approvingServerTarget = ref(null);
const approveServerQuota = ref(2);
const approveUnlimitedTotalCpu = ref(true);
const approveTotalCpu = ref(2.0);
const approveUnlimitedTotalRam = ref(true);
const approveTotalRam = ref(4096);
const approveUnlimitedInstanceCpu = ref(true);
const approveInstanceCpu = ref(1.0);
const approveUnlimitedInstanceRam = ref(true);
const approveInstanceRam = ref(1024);
const approveComment = ref('');
const rejectingServerTarget = ref(null);
const rejectServerReason = ref('');
const submittingServerReview = ref(false);
const serverReviewError = ref('');

async function loadQueue() {
  loading.value = true;
  try {
    const [modRes, projRes] = await Promise.allSettled([
      moderationApi.listRequests({ limit: 100, offset: 0 }),
      listProjects({ limit: 100 }),
    ]);

    // Карта проектов для извлечения названий и иконок
    const projectMap = {};
    if (projRes.status === 'fulfilled' && projRes.value?.projects) {
      projRes.value.projects.forEach((p) => {
        projectMap[p.id] = p;
      });
    }

    const unifiedList = [];
    if (modRes.status === 'fulfilled' && modRes.value?.requests) {
      for (const r of modRes.value.requests) {
        const pInfo = projectMap[r.projectId];
        const isServer =
          r.type === 2 ||
          r.type === 'server' ||
          r.type === 'REQUEST_TYPE_SERVER_ACCESS' ||
          r.type === 'SERVER_ACCESS';

        unifiedList.push({
          id: r.id,
          type: isServer ? 'server' : 'project',
          projectId: r.projectId,
          title:
            r.snapshot?.titleRu ||
            r.snapshot?.titleEn ||
            pInfo?.title_ru ||
            pInfo?.title_en ||
            `Проект #${r.projectId}`,
          iconPath: r.snapshot?.iconPath || pInfo?.icon_path,
          isOnline: r.snapshot?.isOnline ?? pInfo?.is_online ?? isServer,
          version: r.snapshot?.activeBuildVersion || r.snapshot?.buildVersion || '',
          reason: r.reason || '',
          ownerId: r.ownerId || pInfo?.owner_id,
          submittedAt: r.submittedAt,
          status: r.status,
          moderatorId: r.moderatorId,
          rejectionReason: r.rejectionReason || '',
          maxInstances: r.maxInstances || 2,
          maxTotalCpuMillis: r.maxTotalCpuMillis || 0,
          maxTotalMemoryMb: r.maxTotalMemoryMb || 0,
          maxInstanceCpuMillis: r.maxInstanceCpuMillis || 0,
          maxInstanceMemoryMb: r.maxInstanceMemoryMb || 0,
          moderatorComment: r.moderatorComment || '',
          raw: r,
        });
      }
    }

    requests.value = unifiedList;
  } catch (err) {
    showToast(t('common.error'), 'danger');
  } finally {
    loading.value = false;
  }
}

onMounted(loadQueue);

function resetFilters() {
  typeFilter.value = 'all';
  searchQuery.value = '';
  statusFilter.value = 'all';
  modeFilter.value = 'all';
  sortBy.value = 'newest';
  currentPage.value = 1;
}

const filteredRequests = computed(() => {
  let list = [...requests.value];

  // Фильтр по типу заявки
  if (typeFilter.value === 'project' || typeFilter.value === 'publication') {
    list = list.filter((r) => r.type === 'project');
  } else if (typeFilter.value === 'server' || typeFilter.value === 'servers') {
    list = list.filter((r) => r.type === 'server');
  }

  // Фильтр по статусу
  if (statusFilter.value === 'pending') {
    list = list.filter((r) => isPending(r.status));
  } else if (statusFilter.value === 'in_review') {
    list = list.filter((r) => isInReview(r.status));
  } else if (statusFilter.value === 'my') {
    list = list.filter(
      (r) =>
        r.moderatorId &&
        currentUserId.value &&
        (r.moderatorId === currentUserId.value || r.moderatorId === authState.user?.email)
    );
  }

  // Фильтр по сетевому режиму
  if (modeFilter.value === 'online') {
    list = list.filter((r) => r.type === 'server' || r.isOnline);
  } else if (modeFilter.value === 'offline') {
    list = list.filter((r) => r.type === 'project' && !r.isOnline);
  }

  // Поиск
  if (searchQuery.value.trim()) {
    const q = searchQuery.value.toLowerCase().trim();
    list = list.filter((r) => {
      const title = (r.title || '').toLowerCase();
      const pId = String(r.projectId || '');
      const owner = (r.ownerId || '').toLowerCase();
      const reason = (r.reason || '').toLowerCase();
      return title.includes(q) || pId.includes(q) || owner.includes(q) || reason.includes(q);
    });
  }

  // Сортировка
  if (sortBy.value === 'newest') {
    list.sort((a, b) => new Date(b.submittedAt || 0) - new Date(a.submittedAt || 0));
  } else if (sortBy.value === 'oldest') {
    list.sort((a, b) => new Date(a.submittedAt || 0) - new Date(b.submittedAt || 0));
  } else if (sortBy.value === 'title') {
    list.sort((a, b) => (a.title || '').localeCompare(b.title || ''));
  }

  return list;
});

const totalPages = computed(() =>
  Math.max(1, Math.ceil(filteredRequests.value.length / pageSize.value))
);

const paginatedRequests = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value;
  return filteredRequests.value.slice(start, start + pageSize.value);
});

function isPending(status) {
  return (
    status === REQUEST_STATUS.PENDING ||
    status === 1 ||
    status === 'REQUEST_STATUS_PENDING' ||
    status === 'pending'
  );
}

function isInReview(status) {
  return (
    status === REQUEST_STATUS.IN_REVIEW ||
    status === 2 ||
    status === 'REQUEST_STATUS_IN_REVIEW' ||
    status === 'in_review'
  );
}

function isApproved(status) {
  return (
    status === REQUEST_STATUS.APPROVED ||
    status === 3 ||
    status === 'REQUEST_STATUS_APPROVED' ||
    status === 'approved'
  );
}

function isRejected(status) {
  return (
    status === REQUEST_STATUS.REJECTED ||
    status === 4 ||
    status === 'REQUEST_STATUS_REJECTED' ||
    status === 'rejected'
  );
}

function reqStatusLabel(status) {
  if (isPending(status)) return t('moderation.pending');
  if (isInReview(status)) return t('moderation.inReview');
  if (isApproved(status)) return t('projects.approved');
  if (isRejected(status)) return t('projects.rejected');
  return t('common.unknown');
}

function reqStatusClass(status) {
  if (isPending(status)) return 'status-pending';
  if (isInReview(status)) return 'status-in-review';
  if (isApproved(status)) return 'status-approved';
  if (isRejected(status)) return 'status-rejected';
  return 'status-neutral';
}

function openProject(projectId) {
  router.push(`/moderator/projects/${projectId}`);
}

async function claimAndOpen(req) {
  claimingId.value = req.id;
  try {
    await moderationApi.claimRequest(req.id);
    showToast(t('moderation.claimBtn') + ' — успешно', 'success');
    openProject(req.projectId);
  } catch (err) {
    showToast(err.response?.data?.message || t('common.error'), 'danger');
  } finally {
    claimingId.value = null;
  }
}

function openApproveServerModal(req) {
  approvingServerTarget.value = req;
  approveServerQuota.value = req.maxInstances || 2;

  approveUnlimitedTotalCpu.value = !req.maxTotalCpuMillis || req.maxTotalCpuMillis <= 0;
  approveTotalCpu.value = req.maxTotalCpuMillis ? req.maxTotalCpuMillis / 1000 : 2.0;

  approveUnlimitedTotalRam.value = !req.maxTotalMemoryMb || req.maxTotalMemoryMb <= 0;
  approveTotalRam.value = req.maxTotalMemoryMb || 4096;

  approveUnlimitedInstanceCpu.value = !req.maxInstanceCpuMillis || req.maxInstanceCpuMillis <= 0;
  approveInstanceCpu.value = req.maxInstanceCpuMillis ? req.maxInstanceCpuMillis / 1000 : 1.0;

  approveUnlimitedInstanceRam.value = !req.maxInstanceMemoryMb || req.maxInstanceMemoryMb <= 0;
  approveInstanceRam.value = req.maxInstanceMemoryMb || 1024;

  approveComment.value = '';
  serverReviewError.value = '';
}

function openRejectServerModal(req) {
  rejectingServerTarget.value = req;
  rejectServerReason.value = '';
  serverReviewError.value = '';
}

async function submitApproveServer() {
  if (!approvingServerTarget.value) return;
  submittingServerReview.value = true;
  serverReviewError.value = '';
  try {
    await moderationApi.reviewServerAccess(approvingServerTarget.value.id, {
      approved: true,
      maxInstances: Number(approveServerQuota.value) || 2,
      maxTotalCpuMillis: approveUnlimitedTotalCpu.value ? 0 : Math.round((Number(approveTotalCpu.value) || 0) * 1000),
      maxTotalMemoryMb: approveUnlimitedTotalRam.value ? 0 : Number(approveTotalRam.value) || 0,
      maxInstanceCpuMillis: approveUnlimitedInstanceCpu.value ? 0 : Math.round((Number(approveInstanceCpu.value) || 0) * 1000),
      maxInstanceMemoryMb: approveUnlimitedInstanceRam.value ? 0 : Number(approveInstanceRam.value) || 0,
      moderatorComment: approveComment.value.trim(),
    });
    showToast('Доступ к серверам успешно одобрен', 'success');
    approvingServerTarget.value = null;
    await loadQueue();
  } catch (err) {
    serverReviewError.value = err.response?.data?.message || err.message || 'Ошибка одобрения';
  } finally {
    submittingServerReview.value = false;
  }
}

async function submitRejectServer() {
  if (!rejectingServerTarget.value) return;
  if (!rejectServerReason.value.trim()) {
    serverReviewError.value = 'Укажите причину отказа';
    return;
  }
  submittingServerReview.value = true;
  serverReviewError.value = '';
  try {
    await moderationApi.reviewServerAccess(rejectingServerTarget.value.id, {
      approved: false,
      rejectionReason: rejectServerReason.value.trim(),
    });
    showToast('Заявка на доступ отклонена', 'info');
    rejectingServerTarget.value = null;
    await loadQueue();
  } catch (err) {
    serverReviewError.value = err.response?.data?.message || err.message || 'Ошибка отклонения';
  } finally {
    submittingServerReview.value = false;
  }
}
</script>

<style scoped>
.moderation-page-container {
  width: 100%;
  min-height: calc(100vh - 60px);
  background: var(--bg-app, #0d1117);
  padding: 24px 32px 48px;
  box-sizing: border-box;
}

.main-content-wrap {
  width: 100%;
  max-width: 100%;
  margin: 0;
  display: flex;
  flex-direction: column;
}

/* Панель фильтрации в стиле Яндекс.Игр / Developer Hub */
.filters-toolbar {
  display: flex;
  align-items: flex-end;
  gap: 16px;
  margin-bottom: 24px;
  width: 100%;
}

.filter-field {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.field-search {
  flex: 1;
  min-width: 220px;
}

.field-status {
  width: 180px;
  flex-shrink: 0;
}

.field-sort {
  width: 210px;
  flex-shrink: 0;
}

.field-label {
  font-size: 12px;
  font-weight: 500;
  color: var(--text-tertiary, #8b949e);
  letter-spacing: 0.1px;
}

.input-wrapper,
.select-wrapper {
  position: relative;
  display: flex;
  align-items: center;
  width: 100%;
}

.filter-input {
  width: 100%;
  height: 36px;
  padding: 0 32px 0 12px;
  background: var(--bg-secondary, #161b22);
  border: 1px solid var(--border, #30363d);
  border-radius: var(--radius-sm, 6px);
  color: var(--text-main, #f0f6fc);
  font-size: 13px;
  outline: none;
  transition: border-color 0.15s;
  box-sizing: border-box;
}

.filter-input:focus {
  border-color: var(--primary, #58a6ff);
}

.filter-input::placeholder {
  color: var(--text-tertiary, #6e7681);
}

.clear-input-btn {
  position: absolute;
  right: 8px;
  background: transparent;
  border: none;
  color: var(--text-tertiary, #8b949e);
  cursor: pointer;
  padding: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.clear-input-btn:hover {
  color: var(--text-main, #f0f6fc);
}

.filter-select {
  width: 100%;
  height: 36px;
  padding: 0 30px 0 12px;
  background: var(--bg-secondary, #161b22);
  border: 1px solid var(--border, #30363d);
  border-radius: var(--radius-sm, 6px);
  color: var(--text-main, #f0f6fc);
  font-size: 13px;
  outline: none;
  cursor: pointer;
  appearance: none;
  -webkit-appearance: none;
  transition: border-color 0.15s;
  box-sizing: border-box;
}

.filter-select:focus {
  border-color: var(--primary, #58a6ff);
}

.select-arrow {
  position: absolute;
  right: 10px;
  pointer-events: none;
  color: var(--text-tertiary, #8b949e);
}

.btn-reset-filters {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  height: 36px;
  padding: 0 12px;
  background: transparent;
  border: 1px solid var(--border, #30363d);
  border-radius: var(--radius-sm, 6px);
  color: var(--text-muted, #b0b8c4);
  font-size: 13px;
  cursor: pointer;
  flex-shrink: 0;
  transition: all 0.15s;
}

.btn-reset-filters:hover {
  background: var(--bg-tertiary, #21262d);
  color: var(--text-main, #f0f6fc);
  border-color: var(--border-secondary, #484f58);
}

.btn-refresh {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  height: 36px;
  padding: 0 14px;
  background: var(--bg-secondary, #161b22);
  border: 1px solid var(--border, #30363d);
  border-radius: var(--radius-sm, 6px);
  color: var(--text-main, #f0f6fc);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  flex-shrink: 0;
  transition: all 0.15s;
}

.btn-refresh:hover:not(:disabled) {
  background: var(--bg-tertiary, #21262d);
  border-color: var(--border-secondary, #484f58);
}

.btn-refresh:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

/* Табличный вид */
.table-wrapper {
  width: 100%;
  background: var(--bg-card, #161b22);
  border: 1px solid var(--border, #30363d);
  border-radius: var(--radius-md, 8px);
  overflow-x: auto;
}

.moderation-table {
  width: 100%;
  border-collapse: collapse;
  text-align: left;
}

.moderation-table thead {
  background: var(--bg-card, #161b22);
  border-bottom: 1px solid var(--border, #30363d);
}

.moderation-table th {
  padding: 12px 16px;
  font-size: 12px;
  font-weight: 600;
  color: var(--text-tertiary, #8b949e);
  letter-spacing: 0.2px;
}

.col-game {
  width: 32%;
}

.col-version {
  width: 10%;
}

.col-dev {
  width: 16%;
}

.col-date {
  width: 14%;
}

.col-status {
  width: 12%;
}

.col-mod {
  width: 10%;
}

.col-actions {
  width: 6%;
  text-align: right;
  padding-right: 16px;
}

.table-row {
  cursor: pointer;
  border-bottom: 1px solid var(--border, #21262d);
  transition: background-color 0.15s ease;
}

.table-row:hover {
  background: var(--bg-secondary, #161b22);
}

.table-row td {
  padding: 12px 16px;
  vertical-align: middle;
}

.table-row td.col-actions {
  text-align: right;
  padding-right: 16px;
}

.game-cell {
  display: flex;
  align-items: center;
  gap: 14px;
}

.game-icon-box {
  width: 40px;
  height: 40px;
  border-radius: var(--radius-sm, 8px);
  background: var(--bg-tertiary, #21262d);
  border: 1px solid var(--border, #30363d);
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  flex-shrink: 0;
}

.game-icon-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.game-icon-mock {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--bg-tertiary, #21262d);
  color: var(--text-tertiary, #8b949e);
  font-size: 11px;
  font-weight: 600;
}

.game-text {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.game-type-label {
  font-size: 12px;
  font-weight: 400;
  color: var(--text-tertiary, #8b949e);
}

.game-title-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.game-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-main, #f0f6fc);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 260px;
}

.mode-badge {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 11px;
  font-weight: 500;
  padding: 2px 7px;
  border-radius: 4px;
  line-height: 1.3;
  white-space: nowrap;
}

.mode-badge.mode-online {
  color: #38bdf8;
  background: rgba(14, 165, 233, 0.12);
  border: 1px solid rgba(14, 165, 233, 0.25);
}

.mode-badge.mode-offline {
  color: #94a3b8;
  background: rgba(148, 163, 184, 0.1);
  border: 1px solid rgba(148, 163, 184, 0.2);
}

.icon-xxs {
  width: 12px;
  height: 12px;
}

.version-badge {
  display: inline-flex;
  align-items: center;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 600;
  background: var(--bg-tertiary, #21262d);
  border: 1px solid var(--border, #30363d);
  color: var(--primary, #58a6ff);
}

.dev-cell {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  color: var(--text-muted, #b0b8c4);
  max-width: 160px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.date-text {
  font-size: 13px;
  color: var(--text-muted, #b0b8c4);
}

.status-pill {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  height: 24px;
  padding: 0 10px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;
  white-space: nowrap;
}

.status-pending {
  background: rgba(245, 176, 39, 0.12);
  border: 1px solid rgba(245, 176, 39, 0.35);
  color: #f5b027;
}

.status-in-review {
  background: rgba(88, 166, 255, 0.12);
  border: 1px solid rgba(88, 166, 255, 0.35);
  color: #58a6ff;
}

.status-neutral {
  background: var(--bg-tertiary, #21262d);
  border: 1px solid var(--border, #30363d);
  color: var(--text-muted, #b0b8c4);
}

.mod-name {
  font-size: 13px;
  color: var(--text-main, #f0f6fc);
  font-weight: 500;
}

.unassigned-text {
  font-size: 13px;
  color: var(--text-tertiary, #6e7681);
}

.row-actions {
  display: inline-flex;
  align-items: center;
  gap: 8px;
}

.btn-claim-sm {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  height: 30px;
  padding: 0 12px;
  background: var(--primary, #58a6ff);
  color: #ffffff;
  border: none;
  border-radius: var(--radius-sm, 6px);
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
  transition: background-color 0.15s;
  white-space: nowrap;
}

.btn-claim-sm:hover:not(:disabled) {
  background: var(--primary-hover, #79c0ff);
}

.btn-claim-sm:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-inspect-sm {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  height: 30px;
  padding: 0 12px;
  background: var(--bg-secondary, #21262d);
  border: 1px solid var(--border, #30363d);
  color: var(--text-main, #f0f6fc);
  border-radius: var(--radius-sm, 6px);
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s;
  white-space: nowrap;
}

.btn-inspect-sm:hover {
  border-color: var(--primary, #58a6ff);
  color: var(--primary, #58a6ff);
}

/* Пустые состояния */
.state-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 20px;
  gap: 12px;
  color: var(--text-muted, #b0b8c4);
}

.empty-card,
.loading-card {
  background: var(--bg-card, #161b22);
  border: 1px solid var(--border, #30363d);
  border-radius: var(--radius-md, 8px);
}

.loading-card {
  min-height: 280px;
}

.empty-icon-wrap {
  width: 56px;
  height: 56px;
  border-radius: 50%;
  background: var(--bg-secondary, #21262d);
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 8px;
}

.btn-primary-sm {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  height: 34px;
  padding: 0 16px;
  background: var(--primary, #58a6ff);
  color: #ffffff;
  border: none;
  border-radius: var(--radius-sm, 6px);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  margin-top: 12px;
}

/* Пагинация */
.pagination-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 0;
  font-size: 13px;
  color: var(--text-muted, #b0b8c4);
}

.page-nav {
  display: flex;
  align-items: center;
  gap: 6px;
}

.page-nav-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  background: var(--bg-secondary, #161b22);
  border: 1px solid var(--border, #30363d);
  border-radius: var(--radius-sm, 6px);
  color: var(--text-main, #f0f6fc);
  cursor: pointer;
  transition: all 0.15s;
}

.page-nav-btn:hover:not(:disabled) {
  border-color: var(--primary, #58a6ff);
  color: var(--primary, #58a6ff);
}

.page-nav-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.page-current {
  padding: 0 8px;
  font-weight: 500;
}

.page-size-wrap {
  position: relative;
  display: flex;
  align-items: center;
  margin-left: 8px;
}

.page-size-select {
  height: 32px;
  padding: 0 24px 0 8px;
  background: var(--bg-secondary, #161b22);
  border: 1px solid var(--border, #30363d);
  border-radius: var(--radius-sm, 6px);
  color: var(--text-main, #f0f6fc);
  font-size: 12px;
  outline: none;
  cursor: pointer;
  appearance: none;
  -webkit-appearance: none;
}

.select-caret {
  position: absolute;
  right: 6px;
  pointer-events: none;
  color: var(--text-tertiary, #8b949e);
}

.spinner-md {
  width: 32px;
  height: 32px;
  border: 3px solid var(--border, #30363d);
  border-top-color: var(--primary, #58a6ff);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
  margin: 0 auto;
  display: block;
  flex-shrink: 0;
}

.spin {
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

.audit-mode-banner {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 16px;
  background: rgba(59, 130, 246, 0.1);
  border: 1px solid rgba(59, 130, 246, 0.25);
  border-radius: var(--radius-md, 8px);
  color: #60a5fa;
  font-size: 0.84rem;
  font-weight: 500;
}

/* Вкладки типов очереди */
.queue-tabs {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 20px;
  border-bottom: 1px solid var(--border, #30363d);
  padding-bottom: 12px;
}

.queue-tab-btn {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  background: var(--bg-card, #161b22);
  border: 1px solid var(--border, #30363d);
  border-radius: var(--radius-sm, 6px);
  color: var(--text-muted, #8b949e);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s ease;
}

.queue-tab-btn:hover {
  background: var(--bg-secondary, #21262d);
  color: var(--text-main, #f0f6fc);
  border-color: var(--border-hover, #484f58);
}

.queue-tab-btn.active {
  background: rgba(59, 130, 246, 0.15);
  border-color: var(--primary, #58a6ff);
  color: var(--primary, #58a6ff);
}

.tab-count {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 20px;
  height: 20px;
  padding: 0 6px;
  border-radius: 10px;
  font-size: 11px;
  font-weight: 600;
  background: var(--bg-secondary, #21262d);
  color: var(--text-main, #f0f6fc);
}

.queue-tab-btn.active .tab-count {
  background: var(--primary, #58a6ff);
  color: #ffffff;
}

/* Бейджи типов заявок */
.type-pill {
  display: inline-flex;
  align-items: center;
  padding: 2px 7px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.3px;
}

.type-project {
  background: rgba(88, 166, 255, 0.15);
  color: #58a6ff;
  border: 1px solid rgba(88, 166, 255, 0.3);
}

.type-server {
  background: rgba(245, 158, 11, 0.15);
  color: #f59e0b;
  border: 1px solid rgba(245, 158, 11, 0.3);
}

.server-details-cell {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.quota-chips-line {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-wrap: wrap;
}

.quota-chip {
  display: inline-flex;
  align-items: center;
  font-size: 11px;
  font-weight: 600;
  padding: 1px 6px;
  border-radius: 4px;
  background: var(--bg-secondary, #21262d);
  color: var(--text-secondary, #c9d1d9);
  border: 1px solid var(--border, #30363d);
}

.quota-subline {
  font-size: 11px;
  color: var(--text-secondary, #8b949e);
}

.mod-comment-cell {
  font-size: 11px;
  color: #60a5fa;
  font-style: italic;
  max-width: 140px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.server-reason-snippet {
  font-size: 12px;
  color: var(--text-secondary, #c9d1d9);
  max-width: 240px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  margin: 0;
  font-style: italic;
}

.mod-decision-box {
  display: flex;
  flex-direction: column;
  gap: 2px;
  font-size: 12px;
}

.sub-mod {
  font-size: 11px;
  color: var(--text-tertiary, #8b949e);
}

.rejection-text-cell {
  max-width: 140px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  font-size: 11px;
}

/* Кнопки действий для серверов */
.btn-approve-sm {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  height: 28px;
  padding: 0 10px;
  background: rgba(34, 197, 94, 0.15);
  color: #4ade80;
  border: 1px solid rgba(34, 197, 94, 0.35);
  border-radius: var(--radius-sm, 6px);
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s ease;
}

.btn-approve-sm:hover {
  background: rgba(34, 197, 94, 0.25);
  border-color: #4ade80;
}

.btn-reject-sm {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  height: 28px;
  padding: 0 10px;
  background: rgba(239, 68, 68, 0.15);
  color: #f87171;
  border: 1px solid rgba(239, 68, 68, 0.35);
  border-radius: var(--radius-sm, 6px);
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s ease;
}

.btn-reject-sm:hover {
  background: rgba(239, 68, 68, 0.25);
  border-color: #f87171;
}

/* Модальные окна */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.7);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 9999;
  padding: 20px;
}

.modal-card {
  background: var(--bg-card, #161b22);
  border: 1px solid var(--border, #30363d);
  border-radius: var(--radius-md, 12px);
  width: 100%;
  max-width: 480px;
  box-shadow: 0 12px 36px rgba(0, 0, 0, 0.4);
  display: flex;
  flex-direction: column;
  overflow: hidden;
  animation: modal-fade-in 0.15s ease-out;
}

.modal-card.modal-card-wide {
  max-width: 580px;
}

.project-info-banner {
  background: var(--bg-secondary, #0d1117);
  border: 1px solid var(--border, #30363d);
  border-radius: var(--radius-sm, 6px);
  padding: 12px 14px;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.project-info-title {
  font-size: 13px;
  color: var(--text-main, #f0f6fc);
}

.requested-summary-line {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 4px;
  font-size: 11px;
}

.req-sum-label {
  color: var(--text-tertiary, #8b949e);
  font-weight: 500;
}

.req-sum-val {
  padding: 2px 6px;
  background: rgba(88, 166, 255, 0.12);
  border: 1px solid rgba(88, 166, 255, 0.25);
  border-radius: 4px;
  color: #58a6ff;
  font-weight: 600;
}

.moderator-quota-form {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.quota-group-card {
  background: var(--bg-secondary, #0d1117);
  border: 1px solid var(--border, #30363d);
  border-radius: var(--radius-sm, 6px);
  padding: 12px 14px;
}

.quota-group-title {
  margin: 0 0 8px;
  font-size: 11px;
  font-weight: 600;
  color: var(--text-tertiary, #8b949e);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.quota-inputs-row {
  display: flex;
  gap: 12px;
}

.flex-1 {
  flex: 1;
}

.mb-12 {
  margin-bottom: 12px;
}

.label-with-toggle {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 4px;
}

.label-with-toggle .form-label {
  margin-bottom: 0 !important;
}

.toggle-label {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 11px;
  color: var(--text-tertiary, #8b949e);
  cursor: pointer;
  user-select: none;
}

.toggle-label input {
  cursor: pointer;
}

.unlimited-placeholder {
  padding: 8px 12px;
  background: var(--bg-card, #161b22);
  border: 1px dashed var(--border, #30363d);
  border-radius: var(--radius-sm, 6px);
  font-size: 12px;
  color: var(--text-tertiary, #8b949e);
  text-align: center;
}

@keyframes modal-fade-in {
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
}

.modal-header h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: var(--text-main, #f0f6fc);
}

.btn-close {
  background: transparent;
  border: none;
  color: var(--text-muted, #8b949e);
  cursor: pointer;
  padding: 4px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border-radius: 4px;
}

.btn-close:hover {
  color: var(--text-main, #f0f6fc);
  background: var(--bg-secondary, #21262d);
}

.modal-body {
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.modal-intro {
  margin: 0;
  font-size: 14px;
  color: var(--text-main, #f0f6fc);
}

.modal-reason-quote {
  margin: 0;
  font-size: 13px;
  font-style: italic;
  color: var(--text-secondary, #8b949e);
  background: var(--bg-secondary, #0d1117);
  padding: 10px 14px;
  border-left: 3px solid var(--primary, #58a6ff);
  border-radius: 4px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.form-label {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-secondary, #c9d1d9);
}

.form-input,
.form-textarea {
  width: 100%;
  box-sizing: border-box;
  background: var(--bg-secondary, #0d1117);
  border: 1px solid var(--border, #30363d);
  border-radius: var(--radius-sm, 6px);
  padding: 8px 12px;
  color: var(--text-main, #f0f6fc);
  font-size: 13px;
  font-family: inherit;
  transition: border-color 0.15s;
}

.form-input:focus,
.form-textarea:focus {
  border-color: var(--primary, #58a6ff);
  outline: none;
}

.form-textarea {
  resize: vertical;
  min-height: 70px;
}

.field-hint {
  font-size: 11px;
  color: var(--text-tertiary, #8b949e);
}

.alert-error {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  background: rgba(239, 68, 68, 0.15);
  border: 1px solid rgba(239, 68, 68, 0.3);
  border-radius: var(--radius-sm, 6px);
  color: #f87171;
  font-size: 12px;
}

.modal-footer {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 10px;
  padding: 14px 20px;
  border-top: 1px solid var(--border, #30363d);
  background: var(--bg-card, #161b22);
}

.btn-secondary-sm {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  height: 32px;
  padding: 0 14px;
  background: var(--bg-secondary, #21262d);
  border: 1px solid var(--border, #30363d);
  color: var(--text-main, #f0f6fc);
  border-radius: var(--radius-sm, 6px);
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s;
}

.btn-secondary-sm:hover:not(:disabled) {
  background: var(--bg-hover, #30363d);
}

.btn-danger-sm {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  height: 32px;
  padding: 0 14px;
  background: #dc2626;
  border: none;
  color: #ffffff;
  border-radius: var(--radius-sm, 6px);
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
  transition: background 0.15s;
}

.btn-danger-sm:hover:not(:disabled) {
  background: #b91c1c;
}

.btn-danger-sm:disabled,
.btn-secondary-sm:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

@media (max-width: 800px) {
  .moderation-page-container {
    padding: 16px;
  }
  .filters-toolbar {
    flex-wrap: wrap;
  }
  .field-search {
    width: 100%;
    min-width: 100%;
  }
}
</style>
