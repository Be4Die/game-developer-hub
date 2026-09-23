<template>
  <div v-if="projectData" class="moderator-workspace" :class="{ 'chat-open': isChatOpen }">
    <!-- ОСНОВНАЯ РАБОЧАЯ ОБЛАСТЬ (ЛЕВАЯ ЧАСТЬ) -->
    <main class="moderator-main-area">
      <!-- 1. ДВУХСТРОЧНАЯ ВЕРХНЯЯ СТРОКА ИНФОРМАЦИИ О ПРОЕКТЕ -->
      <header class="top-info-bar">
        <!-- Левый блок: Иконка проекта (растягивается на обе строки) -->
        <div class="top-icon-cell">
          <div class="main-icon-box">
            <img
              v-if="projectIconUrl"
              :src="projectIconUrl"
              alt="Icon"
              class="main-icon-img"
            />
            <div v-else class="main-icon-placeholder">
              <Gamepad2 class="icon-md" />
            </div>
          </div>
        </div>

        <!-- Центральный блок: 2 отдельные строки данных -->
        <div class="top-center-cell">
          <!-- Строка 1: Название, ID, Бейджи -->
          <div class="top-row-main">
            <div class="title-meta-group">
              <span class="top-project-title" :title="projectTitle">
                {{ projectTitle }}
              </span>
              <span class="project-id-chip">#{{ projectId }}</span>
            </div>

            <div class="top-badges-group">
              <span
                v-if="activeRequest"
                class="badge-chip"
                :class="isUpdateRequest ? 'badge-update' : 'badge-pub'"
              >
                {{ isUpdateRequest ? 'Обновление' : 'Публикация' }}
              </span>

              <span
                v-if="activeRequest"
                class="badge-chip"
                :class="getStatusBadgeClass(requestStatus)"
              >
                {{ getStatusText(requestStatus) }}
              </span>

              <span
                class="badge-chip"
                :class="isProjectOnline ? 'badge-online' : 'badge-offline'"
                :title="isProjectOnline ? t('projects.modeOnline') : t('projects.modeOffline')"
              >
                <Globe v-if="isProjectOnline" class="icon-xxs" />
                <Gamepad2 v-else class="icon-xxs" />
                <span>{{ isProjectOnline ? 'Онлайн' : 'Оффлайн' }}</span>
              </span>

              <span v-if="projectData.activeBuildVersion" class="badge-chip badge-version">
                v{{ projectData.activeBuildVersion }}
              </span>
            </div>
          </div>

          <!-- Строка 2: Автор, Подано, Модератор -->
          <div class="top-row-meta">
            <div class="meta-inline-item">
              <User class="icon-xxs text-tertiary" />
              <span class="meta-caption">Автор:</span>
              <span class="meta-val" :title="ownerUserId">
                {{ ownerUserId ? getUserDisplayName(ownerUserId) : '—' }}
              </span>
            </div>

            <span class="meta-dot-divider">•</span>

            <div class="meta-inline-item">
              <Calendar class="icon-xxs text-tertiary" />
              <span class="meta-caption">Подано:</span>
              <span class="meta-val">
                {{ activeRequest ? formatDateTime(activeRequest.submittedAt) : '—' }}
              </span>
            </div>

            <span class="meta-dot-divider">•</span>

            <div class="meta-inline-item">
              <Eye class="icon-xxs text-tertiary" />
              <span class="meta-caption">Модератор:</span>
              <span class="meta-val" :title="activeRequest?.moderatorId">
                {{ activeRequest?.moderatorId ? getUserDisplayName(activeRequest.moderatorId) : 'Не назначен' }}
              </span>
            </div>
          </div>
        </div>

        <!-- Правый блок: Кнопка отметки шага + Кнопка чата (обе растянуты на две строки) -->
        <div class="top-actions-cell">
          <!-- Кнопка явной пометки текущего шага "Проверен" (для шагов 1..4) -->
          <button
            v-if="currentStep <= 4"
            type="button"
            class="top-verify-btn-tall"
            :class="{ 'is-verified': verifiedSteps[currentStep] }"
            :title="verifiedSteps[currentStep] ? 'Снять отметку проверки с текущего шага' : 'Отметить текущий шаг как проверенный'"
            @click="toggleStepVerified(currentStep)"
          >
            <CheckCircle2 v-if="verifiedSteps[currentStep]" class="icon-sm text-success" />
            <Circle v-else class="icon-sm text-muted" />
            <span class="top-verify-text">{{ verifiedSteps[currentStep] ? 'Шаг проверен' : 'Отметить проверенным' }}</span>
          </button>

          <!-- Кнопка скрытия / открытия чата -->
          <button
            type="button"
            class="chat-toggle-btn-tall"
            :class="{ active: isChatOpen }"
            :title="isChatOpen ? 'Скрыть чат' : 'Показать чат проекта'"
            @click="isChatOpen = !isChatOpen"
          >
            <MessageSquare class="icon-sm" />
            <span class="chat-toggle-text">Чат</span>
            <span v-if="unreadChatCount > 0" class="chat-unread-badge">
              {{ unreadChatCount }}
            </span>
          </button>
        </div>
      </header>

      <!-- 2. ЦЕНТРАЛЬНАЯ ОБЛАСТЬ: КОНТЕНТ ТЕКУЩЕГО ШАГА -->
      <div class="step-content-viewport" :class="{ 'is-sandbox-step': currentStep === 4 }">
        <!-- ШАГ 1: ОСНОВНАЯ ИНФОРМАЦИЯ -->
        <div v-show="currentStep === 1" class="step-pane step-pane-scrollable">
          <div class="pane-content-wrapper">
            <div class="card">
              <div class="section-head">
                <div class="head-title-row">
                  <FileText class="icon-sm text-primary" />
                  <h3>Названия и текстовые описания</h3>
                </div>
              </div>

              <!-- Названия -->
              <div class="data-row">
                <div class="data-group">
                  <label class="data-label">{{ t('projectDraft.gameTitleRu') || 'Название (RU)' }}</label>
                  <div class="data-box">{{ projectData.titleRu || '—' }}</div>
                </div>
                <div class="data-group">
                  <label class="data-label">{{ t('projectDraft.gameTitleEn') || 'Название (EN)' }}</label>
                  <div class="data-box">{{ projectData.titleEn || '—' }}</div>
                </div>
              </div>

              <!-- SEO описания -->
              <div class="data-row">
                <div class="data-group">
                  <label class="data-label">{{ t('projectDraft.seoDescriptionRu') || 'Краткое SEO описание (RU)' }}</label>
                  <div class="data-box multiline">{{ projectData.seoRu || '—' }}</div>
                </div>
                <div class="data-group">
                  <label class="data-label">{{ t('projectDraft.seoDescriptionEn') || 'Краткое SEO описание (EN)' }}</label>
                  <div class="data-box multiline">{{ projectData.seoEn || '—' }}</div>
                </div>
              </div>

              <!-- Полные описания -->
              <div class="data-row">
                <div class="data-group">
                  <label class="data-label">{{ t('projectDraft.aboutGameRu') || 'Полное описание игры (RU)' }}</label>
                  <div class="data-box multiline-lg">{{ projectData.aboutRu || projectData.about || '—' }}</div>
                </div>
                <div class="data-group">
                  <label class="data-label">{{ t('projectDraft.aboutGameEn') || 'Полное описание игры (EN)' }}</label>
                  <div class="data-box multiline-lg">{{ projectData.aboutEn || '—' }}</div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- ШАГ 2: МЕДИА И ПРОМО-МАТЕРИАЛЫ -->
        <div v-show="currentStep === 2" class="step-pane step-pane-scrollable">
          <div class="pane-content-wrapper">
            <!-- Верхняя строка: Сетка 2-х колонок под Иконку и Обложку -->
            <div class="media-two-col-grid">
              <!-- Иконка игры -->
              <div class="card media-card-item">
                <div class="section-head">
                  <div class="head-title-row">
                    <Image class="icon-sm text-primary" />
                    <h3>Иконка проекта</h3>
                  </div>
                  <span class="media-pill">512×512 px</span>
                </div>

                <div v-if="projectIconUrl" class="media-item-body">
                  <div
                    class="interactive-image-preview icon-aspect"
                    title="Нажмите для полноэкранного просмотра"
                    @click="openLightbox(projectIconUrl, getFileName(projectData?.iconPath || projectData?.icon_path))"
                  >
                    <img :src="projectIconUrl" alt="Game Icon" class="media-fit-img" />
                    <div class="preview-overlay-hover">
                      <Maximize2 class="icon-md" />
                      <span>Увеличить</span>
                    </div>
                  </div>
                  <div class="media-meta-footer">
                    <span class="media-filename" :title="getFileName(projectData?.iconPath || projectData?.icon_path)">
                      {{ getFileName(projectData?.iconPath || projectData?.icon_path) }}
                    </span>
                    <span class="media-hint">Кликните для просмотра</span>
                  </div>
                </div>
                <div v-else class="empty-media-box">
                  <Image class="icon-md text-muted" />
                  <p>Иконка проекта не загружена</p>
                </div>
              </div>

              <!-- Обложка игры -->
              <div class="card media-card-item">
                <div class="section-head">
                  <div class="head-title-row">
                    <Image class="icon-sm text-primary" />
                    <h3>Обложка проекта</h3>
                  </div>
                  <span class="media-pill">800×470 px (16:9)</span>
                </div>

                <div v-if="projectCoverUrl" class="media-item-body">
                  <div
                    class="interactive-image-preview cover-aspect"
                    title="Нажмите для полноэкранного просмотра"
                    @click="openLightbox(projectCoverUrl, getFileName(projectData?.coverPath || projectData?.cover_path))"
                  >
                    <img :src="projectCoverUrl" alt="Game Cover" class="media-fit-img" />
                    <div class="preview-overlay-hover">
                      <Maximize2 class="icon-md" />
                      <span>Увеличить</span>
                    </div>
                  </div>
                  <div class="media-meta-footer">
                    <span class="media-filename" :title="getFileName(projectData?.coverPath || projectData?.cover_path)">
                      {{ getFileName(projectData?.coverPath || projectData?.cover_path) }}
                    </span>
                    <span class="media-hint">Кликните для просмотра</span>
                  </div>
                </div>
                <div v-else class="empty-media-box">
                  <Image class="icon-md text-muted" />
                  <p>Обложка проекта не загружена</p>
                </div>
              </div>
            </div>

            <!-- Промо-видео проекта -->
            <div class="card video-section-card">
              <div class="section-head">
                <div class="head-title-row">
                  <Video class="icon-sm text-primary" />
                  <h3>Промо-видео (MP4)</h3>
                </div>
                <span v-if="projectVideoUrl" class="media-pill">16:9 H.264/MP4</span>
              </div>

              <div v-if="projectVideoUrl" class="video-preview-layout">
                <div class="video-player-frame">
                  <video
                    :src="projectVideoUrl"
                    controls
                    preload="metadata"
                    class="promo-video-element"
                  ></video>
                </div>
                <div class="video-info-aside">
                  <div class="meta-item">
                    <span class="meta-label">Файл:</span>
                    <span class="meta-value">{{ getFileName(projectData?.videoPath || projectData?.video_path) }}</span>
                  </div>
                  <div class="meta-item">
                    <span class="meta-label">Формат:</span>
                    <span class="meta-value">Видеоролик для каталога и карточки</span>
                  </div>
                  <a
                    :href="projectVideoUrl"
                    target="_blank"
                    download
                    class="btn-video-download"
                  >
                    <ExternalLink class="icon-xs" />
                    <span>Скачать видеофайл</span>
                  </a>
                </div>
              </div>
              <div v-else class="empty-video-inline">
                <Video class="icon-sm text-muted" />
                <span>Промо-видео не загружено разработчиком</span>
              </div>
            </div>
          </div>
        </div>

        <!-- ШАГ 3: ВНУТРИИГРОВЫЕ ТОВАРЫ (IAP) -->
        <div v-show="currentStep === 3" class="step-pane step-pane-scrollable">
          <div class="pane-content-wrapper">
            <div class="card iap-card">
              <div class="section-head">
                <div class="head-title-row">
                  <ShoppingBag class="icon-sm text-primary" />
                  <h3>Внутриигровые товары</h3>
                </div>
              </div>

              <div v-if="!snapshotItems.length" class="empty-items-box">
                <div class="empty-icon-wrap">
                  <ShoppingBag class="icon-lg text-muted" />
                </div>
                <h4>Внутриигровые покупки отсутствуют</h4>
                <p class="text-muted">В проекте внутриигровые товары не заведены или игра является бесплатной без доната.</p>
              </div>

              <div v-else class="table-responsive">
                <table class="table-iap">
                  <thead>
                    <tr>
                      <th class="col-icon">Иконка</th>
                      <th class="col-id">ID товара</th>
                      <th class="col-name">Название (RU / EN)</th>
                      <th class="col-desc">Описание</th>
                      <th class="col-price">Стоимость</th>
                      <th class="col-type">Тип товара</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-for="item in snapshotItems" :key="item.id || item.item_id">
                      <td class="col-icon">
                        <img
                          v-if="item.iconUrl || item.icon_url"
                          :src="item.iconUrl || item.icon_url"
                          alt="IAP"
                          class="iap-thumb"
                          @click="openLightbox(item.iconUrl || item.icon_url, item.titleRu || item.name || 'Товар')"
                        />
                        <div v-else class="iap-thumb-mock">
                          <ShoppingBag class="icon-xxs" />
                        </div>
                      </td>
                      <td class="col-id">
                        <code>{{ item.id || item.item_id }}</code>
                      </td>
                      <td class="col-name">
                        <strong class="item-name-ru">{{ item.titleRu || item.title_ru || item.name || '—' }}</strong>
                        <div v-if="item.titleEn || item.title_en" class="sub-name">
                          {{ item.titleEn || item.title_en }}
                        </div>
                      </td>
                      <td class="col-desc">
                        {{ item.descriptionRu || item.description_ru || item.description || '—' }}
                      </td>
                      <td class="col-price">
                        <span class="price-val">{{ item.price || item.priceCoins || 0 }}</span>
                        <span class="price-currency">{{ item.currency || 'монет' }}</span>
                      </td>
                      <td class="col-type">
                        <span class="iap-type-pill">{{ item.type || 'Consumable' }}</span>
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>
          </div>
        </div>

        <!-- ШАГ 4: ПЕСОЧНИЦА ДЛЯ ТЕСТИРОВАНИЯ -->
        <div v-show="currentStep === 4" class="step-pane step-pane-sandbox">
          <div class="sandbox-player-wrapper">
            <GameSandboxPlayer
              v-if="activeSandboxUrl"
              :key="activeSandboxUrl"
              :game-url="activeSandboxUrl"
              :project-id="projectId"
              :show-devtools="true"
            />
            <div v-else class="empty-sandbox-state">
              <AlertTriangle class="icon-lg text-warning" />
              <p>URL сборки игры не определен</p>
            </div>
          </div>
        </div>

        <!-- ШАГ 5: РЕШЕНИЕ МОДЕРАТОРА (ОДОБРЕНИЕ / ОТКЛОНЕНИЕ) -->
        <div v-show="currentStep === 5" class="step-pane step-pane-scrollable">
          <div class="pane-content-wrapper">
            <!-- Если заявка уже одобрена или отклонена ранее -->
            <div v-if="isApproved || isRejected" class="card verdict-result-card">
              <div class="section-head">
                <div class="head-title-row">
                  <CheckCircle2 v-if="isApproved" class="icon-md text-success" />
                  <XCircle v-else class="icon-md text-danger" />
                  <h3>{{ isApproved ? 'Проект проверен и одобрен' : 'Проект проверен и отклонен' }}</h3>
                </div>
                <span class="badge-chip" :class="getStatusBadgeClass(requestStatus)">
                  {{ getStatusText(requestStatus) }}
                </span>
              </div>
              <div v-if="activeRequest?.rejectionReason || activeRequest?.moderatorComment" class="verdict-summary-comment">
                <p class="summary-label">Комментарий решения:</p>
                <div class="data-box multiline">
                  {{ activeRequest?.rejectionReason || activeRequest?.moderatorComment }}
                </div>
              </div>
            </div>

            <!-- Если заявка в статусе Pending (еще не взята в работу) -->
            <div v-else-if="isPending" class="card verdict-pending-card">
              <div class="pending-prompt-content">
                <AlertCircle class="icon-lg text-warning" />
                <h3>Заявка ожидает назначения</h3>
                <p>Перед вынесением решения необходимо взять заявку в работу.</p>
                <button
                  type="button"
                  class="btn-verdict btn-claim"
                  :disabled="actionLoading"
                  @click="handleClaim"
                >
                  <Eye class="icon-sm" />
                  <span>{{ t('moderation.claimBtn') }}</span>
                </button>
              </div>
            </div>

            <!-- Если заявка In Review: Активное вынесение вердикта -->
            <template v-else>
              <!-- РЕЖИМ 1: ОДОБРЕНИЕ -->
              <div v-if="verdictMode === 'approve'" class="card approve-flow-card">
                <div class="section-head">
                  <div class="head-title-row">
                    <CheckCircle2 class="icon-sm text-success" />
                    <h3>Подтверждение публикации проекта</h3>
                  </div>
                </div>

                <div class="approve-banner">
                  <p>
                    Проект будет опубликован в Production каталоге платформы и станет доступен пользователям.
                    Все данные и материалы соответствуют регламенту платформы.
                  </p>
                </div>

                <div class="form-group">
                  <label class="form-label">Комментарий модератора (необязательно):</label>
                  <textarea
                    v-model="approveComment"
                    class="form-textarea"
                    rows="3"
                    placeholder="Например: Проект соответствует всем правилам площадки."
                    @input="saveToStorage"
                  ></textarea>
                </div>

                <div class="verdict-card-actions">
                  <button
                    type="button"
                    class="btn-submit-approve"
                    :disabled="actionLoading"
                    @click="handleApprove"
                  >
                    <Loader2 v-if="actionLoading" class="icon-xs spin" />
                    <CheckCircle2 v-else class="icon-xs" />
                    <span>{{ actionLoading ? 'Одобрение...' : '✓ Подтвердить одобрение и опубликовать' }}</span>
                  </button>
                </div>
              </div>

              <!-- РЕЖИМ 2: ОТКЛОНЕНИЕ (КОНСТРУКТОР НАРУШЕНИЙ) -->
              <div v-else-if="verdictMode === 'reject'" class="reject-flow-container">
                <div class="header-warning-banner">
                  <AlertTriangle class="icon-md text-danger banner-icon" />
                  <div>
                    <h4>Оформление решения об отклонении заявки</h4>
                    <p>
                      Опишите конкретные нарушения регламента платформы и прикрепите фото/видео доказательства.
                      Структурированный отчёт поступит в чат проекта, а черновик вернётся разработчику на доработку.
                    </p>
                  </div>
                </div>

                <!-- Список нарушений -->
                <div class="card">
                  <div class="section-head">
                    <div class="head-title-row">
                      <ShieldAlert class="icon-sm text-danger" />
                      <h3>Пункты нарушений ({{ violations.length }})</h3>
                    </div>
                    <p class="section-subtitle">
                      Разработчику будет проще исправить замечания, если каждое нарушение выделено в отдельный пункт с доказательством.
                    </p>
                  </div>

                  <div class="violations-list">
                    <div
                      v-for="(item, idx) in violations"
                      :key="idx"
                      class="violation-box"
                    >
                      <div class="violation-box-header">
                        <div class="violation-badge">
                          <span class="number-tag">Пункт #{{ idx + 1 }}</span>
                          <span v-if="item.ruleCode" class="rule-preview-tag">{{ item.ruleCode }}</span>
                        </div>
                        <button
                          v-if="violations.length > 1"
                          type="button"
                          class="btn-remove-box"
                          title="Удалить данный пункт нарушения"
                          @click="removeViolation(idx)"
                        >
                          <Trash2 class="icon-xs" />
                          <span>Удалить пункт</span>
                        </button>
                      </div>

                      <!-- Выбор правила из каталога -->
                      <div class="rule-selector-field">
                        <label class="form-label">
                          Выберите пункт из регламента платформы (или введите вручную):
                        </label>
                        <RuleSearchSelect
                          v-model="item.ruleCode"
                          placeholder="Начните вводить: SEC-04, SRV-02, квоты, вызовы, баг..."
                          @select="onRuleSelected(item, $event)"
                          @clear="onRuleCleared(item)"
                        />
                      </div>

                      <div class="form-grid-two">
                        <div class="form-group">
                          <label class="form-label">
                            Код / № правила <span class="req">*</span>
                          </label>
                          <input
                            v-model="item.ruleCode"
                            type="text"
                            class="form-input code-font"
                            placeholder="Например: SEC-04"
                            @input="saveToStorage"
                          />
                        </div>

                        <div class="form-group">
                          <label class="form-label">
                            Название правила платформы <span class="req">*</span>
                          </label>
                          <input
                            v-model="item.ruleTitle"
                            type="text"
                            class="form-input"
                            placeholder="Например: Несанкционированные сетевые запросы"
                            @input="saveToStorage"
                          />
                        </div>
                      </div>

                      <div class="form-group">
                        <label class="form-label">
                          Подробное описание проблемы и где обнаружено <span class="req">*</span>
                        </label>
                        <textarea
                          v-model="item.description"
                          class="form-textarea"
                          rows="3"
                          placeholder="Опишите, в какой сцене, на каком уровне или при каких действиях воспроизводится нарушение, шаги для воспроизведения..."
                          @input="saveToStorage"
                        ></textarea>
                      </div>

                      <!-- Доказательства -->
                      <div class="evidence-block">
                        <label class="form-label">Доказательства нарушения (скриншот или видео бага):</label>
                        <div class="evidence-items-wrap">
                          <div
                            v-for="(att, aIdx) in item.attachments"
                            :key="att.id || aIdx"
                            class="evidence-chip"
                          >
                            <Film v-if="isVideo(att)" class="icon-xs text-primary" />
                            <Image v-else class="icon-xs text-primary" />
                            <span class="chip-name" :title="att.file_name || att.name">
                              {{ att.file_name || att.name }}
                            </span>
                            <span class="chip-size">({{ formatSize(att.file_size || att.size) }})</span>
                            <button
                              type="button"
                              class="btn-delete-chip"
                              title="Удалить файл"
                              @click="removeAttachment(item, aIdx)"
                            >
                              <X class="icon-xs" />
                            </button>
                          </div>

                          <label class="btn-upload-evidence" :class="{ 'is-loading': item.uploading }">
                            <input
                              type="file"
                              accept="image/png,image/jpeg,image/webp,video/mp4,video/webm"
                              class="file-hidden-input"
                              :disabled="item.uploading"
                              @change="handleFileUpload($event, item)"
                            />
                            <Loader2 v-if="item.uploading" class="icon-xs spin" />
                            <Paperclip v-else class="icon-xs" />
                            <span>{{ item.uploading ? 'Загрузка...' : '+ Прикрепить скриншот / видео' }}</span>
                          </label>
                        </div>
                      </div>
                    </div>
                  </div>

                  <div class="add-more-row">
                    <button type="button" class="btn-add-violation" @click="addViolation">
                      <Plus class="icon-sm" />
                      <span>Добавить еще одно нарушение</span>
                    </button>
                  </div>
                </div>

                <!-- Общее заключение -->
                <div class="card">
                  <div class="section-head">
                    <div class="head-title-row">
                      <MessageSquare class="icon-sm text-primary" />
                      <h3>Общее заключение модератора</h3>
                    </div>
                    <p class="section-subtitle">
                      Финальные рекомендации и напутствие разработчику перед повторной отправкой билда на проверку.
                    </p>
                  </div>
                  <div class="form-group">
                    <textarea
                      v-model="generalComment"
                      class="form-textarea"
                      rows="3"
                      placeholder="Например: Пожалуйста, устраните сетевые вызовы и приведите возрастные ограничения в порядок, после чего загрузите обновленный билд."
                      @input="saveToStorage"
                    ></textarea>
                  </div>
                </div>

                <!-- Кнопка отправки решения об отклонении -->
                <div class="reject-submit-bar">
                  <div v-if="!canSubmitReject" class="validation-tip">
                    <AlertCircle class="icon-xs text-muted" />
                    <span>Заполните код, название и описание хотя бы для одного нарушения.</span>
                  </div>

                  <button
                    type="button"
                    class="btn-submit-reject"
                    :disabled="!canSubmitReject || actionLoading"
                    @click="handleReject"
                  >
                    <Loader2 v-if="actionLoading" class="icon-xs spin" />
                    <XCircle v-else class="icon-xs" />
                    <span>{{ actionLoading ? 'Отклонение проекта...' : '✕ Отклонить проект и отправить вердикт' }}</span>
                  </button>
                </div>
              </div>
            </template>
          </div>
        </div>
      </div>

      <!-- 3. ФИКСИРОВАННАЯ НИЖНЯЯ ПАНЕЛЬ НАВИГАЦИИ ПО ШАГАМ И ДЕЙСТВИЙ -->
      <footer class="bottom-nav-bar">
        <!-- Левая часть: Кнопка "Назад" -->
        <div class="nav-left">
          <button
            type="button"
            class="btn-nav btn-secondary"
            :disabled="currentStep <= 1"
            @click="prevStep"
          >
            <ChevronLeft class="icon-sm" />
            <span>Назад</span>
          </button>
        </div>

        <!-- Центральный степпер с 5-ю шагами -->
        <nav class="nav-stepper">
          <template v-for="(s, sIdx) in steps" :key="s.id">
            <button
              type="button"
              class="stepper-step"
              :class="{
                active: currentStep === s.id,
                completed: verifiedSteps[s.id],
              }"
              :title="s.fullTitle || s.title"
              @click="goToStep(s.id)"
            >
              <span class="step-circle">
                <Check v-if="verifiedSteps[s.id]" class="icon-xs" />
                <span v-else>{{ s.id }}</span>
              </span>
              <span class="step-text">{{ s.title }}</span>
            </button>

            <div
              v-if="sIdx < steps.length - 1"
              class="stepper-line"
              :class="{ filled: verifiedSteps[s.id] }"
            ></div>
          </template>
        </nav>

        <!-- Правая часть: действия перехода / решения -->
        <div class="nav-right">

          <!-- Шаги 1-3: Кнопка "Далее" -->
          <button
            v-if="currentStep < 4"
            type="button"
            class="btn-nav btn-primary"
            @click="nextStep"
          >
            <span>Далее</span>
            <ChevronRight class="icon-sm" />
          </button>

          <!-- Шаг 4 (Песочница): Выбор решения "Одобрить" / "Отклонить" -> переход на Шаг 5 -->
          <template v-else-if="currentStep === 4">
            <!-- Если Pending -->
            <button
              v-if="isPending"
              type="button"
              class="btn-verdict btn-claim"
              :disabled="actionLoading"
              @click="handleClaim"
            >
              <Eye class="icon-sm" />
              <span>{{ t('moderation.claimBtn') }}</span>
            </button>

            <!-- Если In Review: Кнопки Одобрить и Отклонить (ведут на шаг 5) -->
            <div v-else-if="isInReview" class="step-verdict-actions">
              <button
                type="button"
                class="btn-verdict btn-reject-action"
                :disabled="actionLoading"
                @click="openVerdict('reject')"
              >
                <XCircle class="icon-sm" />
                <span>Отклонить</span>
              </button>

              <button
                type="button"
                class="btn-verdict btn-approve-action"
                :disabled="actionLoading"
                @click="openVerdict('approve')"
              >
                <CheckCircle2 class="icon-sm" />
                <span>Одобрить</span>
              </button>
            </div>

            <!-- Если уже завершено -->
            <button
              v-else
              type="button"
              class="btn-nav btn-secondary"
              @click="goToStep(5)"
            >
              <span>К решению</span>
              <ChevronRight class="icon-sm" />
            </button>
          </template>

          <!-- Шаг 5: Быстрые действия в футере (дублирование кнопок подтверждения) -->
          <template v-else-if="currentStep === 5">
            <template v-if="isInReview">
              <button
                v-if="verdictMode === 'approve'"
                type="button"
                class="btn-verdict btn-approve-action"
                :disabled="actionLoading"
                @click="handleApprove"
              >
                <Loader2 v-if="actionLoading" class="icon-xs spin" />
                <CheckCircle2 v-else class="icon-xs" />
                <span>Одобрить</span>
              </button>

              <button
                v-else
                type="button"
                class="btn-verdict btn-reject-action"
                :disabled="!canSubmitReject || actionLoading"
                @click="handleReject"
              >
                <Loader2 v-if="actionLoading" class="icon-xs spin" />
                <XCircle v-else class="icon-xs" />
                <span>Отклонить</span>
              </button>
            </template>

            <div
              v-else-if="isApproved || isRejected"
              class="badge-chip badge-verdict-done"
              :class="getStatusBadgeClass(requestStatus)"
            >
              {{ getStatusText(requestStatus) }}
            </div>
          </template>
        </div>
      </footer>
    </main>

    <!-- ПРАВАЯ СКРЫВАЕМАЯ КОЛОНКА: ЧАТ ПРОЕКТА -->
    <aside v-show="isChatOpen" class="moderator-chat-aside">
      <ProjectChat
        :project-id="projectId"
        :readonly="isAdmin"
        :collapsible="true"
        :is-open="isChatOpen"
        @collapse="isChatOpen = false"
        @unread-count-changed="unreadChatCount = $event"
      />
    </aside>

    <!-- Полноэкранный просмотр изображений (Lightbox) -->
    <MediaLightboxModal
      v-if="lightboxData"
      :src="lightboxData.src"
      :file-name="lightboxData.fileName"
      :download-url="lightboxData.downloadUrl"
      @close="lightboxData = null"
    />
  </div>

  <!-- Состояния загрузки и ошибки -->
  <div v-else-if="loading" class="state-loading-screen">
    <div class="spinner-md"></div>
    <p>{{ t('common.loading') }}</p>
  </div>

  <div v-else class="state-loading-screen">
    <AlertTriangle class="icon-lg text-danger" />
    <p>Проект не найден</p>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import {
  Gamepad2,
  Globe,
  CheckCircle2,
  XCircle,
  AlertTriangle,
  Eye,
  Image,
  Video,
  ExternalLink,
  ShoppingBag,
  MessageSquare,
  FileText,
  ChevronLeft,
  ChevronRight,
  Check,
  Maximize2,
  ShieldAlert,
  Trash2,
  Plus,
  Paperclip,
  Film,
  X,
  Loader2,
  AlertCircle,
  User,
  Calendar,
  Circle,
} from 'lucide-vue-next';
import {
  moderationApi,
  moderationStore,
  getStatusText,
  getStatusBadgeClass,
  REQUEST_STATUS,
  formatDateTime,
  ProjectChat,
  MediaLightboxModal,
  RuleSearchSelect,
} from '@/entities/moderation';
import { getProject, getMediaUrl } from '@/entities/project';
import { useAuth, getUserDisplayName } from '@/entities/user';
import { GameSandboxPlayer } from '@/features/game-sandbox';
import { showToast } from '@/shared/lib';

interface ViolationItem {
  ruleCode: string;
  ruleTitle: string;
  description: string;
  attachments: any[];
  uploading: boolean;
}

const { t } = useI18n();
const route = useRoute();
const router = useRouter();
const { state: authState } = useAuth();

const isAdmin = computed<boolean>(() => {
  const r = authState.user?.role;
  return r === 'USER_ROLE_ADMIN' || r === 'admin' || r === 3;
});

const projectId = computed<string>(() => String(route.params.projectId));

// Определение 5 шагов модерации
const steps = [
  { id: 1, title: 'Информация', fullTitle: 'Основная информация' },
  { id: 2, title: 'Медиа', fullTitle: 'Медиа и материалы' },
  { id: 3, title: 'Товары', fullTitle: 'Внутриигровые товары' },
  { id: 4, title: 'Песочница', fullTitle: 'Песочница игры' },
  { id: 5, title: 'Решение', fullTitle: 'Решение' },
];

// Состояние пошаговой навигации (1..5)
const currentStep = ref<number>(1);

// Независимый учёт проверенности шагов (шаг не сбрасывается при переходе назад)
const verifiedSteps = ref<Record<number, boolean>>({});

// Состояние формы вердикта (Шаг 5)
const verdictMode = ref<'approve' | 'reject'>('approve');
const approveComment = ref<string>('Проект проверен и одобрен к публикации.');
const violations = ref<ViolationItem[]>([
  {
    ruleCode: '',
    ruleTitle: '',
    description: '',
    attachments: [],
    uploading: false,
  },
]);
const generalComment = ref<string>('');

// Состояние чата
const isChatOpen = ref<boolean>(true);
const unreadChatCount = ref<number>(0);

// Состояние полноэкранного просмотра изображения (Lightbox)
const lightboxData = ref<{ src: string; fileName: string; downloadUrl: string } | null>(null);

function openLightbox(src: string, fileName?: string): void {
  if (!src) return;
  lightboxData.value = {
    src,
    fileName: fileName || 'Изображение',
    downloadUrl: src,
  };
}

const activeRequest = ref<any | null>(null);
const projectData = ref<any | null>(null);
const fullProject = ref<any | null>(null);
const loading = ref<boolean>(true);
const actionLoading = ref<boolean>(false);
const noRequestMode = ref<boolean>(false);

const isUpdateRequest = computed<boolean>(() => {
  const tr = activeRequest.value?.type ?? (activeRequest.value as any)?.request_type;
  return Number(tr) === 3 || tr === 'REQUEST_TYPE_PROJECT_UPDATE' || tr === 'project_update';
});

const snapshotItems = computed<any[]>(() => {
  const s = activeRequest.value?.snapshot;
  return s?.items || s?.Items || projectData.value?.items || [];
});

const activeSandboxUrl = computed<string>(() => {
  return projectData.value?.devUrl || `/games/${projectId.value}/dev/index.html`;
});

const requestStatus = computed<any>(() => activeRequest.value?.status);

const isPending = computed<boolean>(() => {
  const st = requestStatus.value;
  if (st === null || st === undefined) return false;
  return (
    Number(st) === REQUEST_STATUS.PENDING ||
    st === 'REQUEST_STATUS_PENDING' ||
    st === 1 ||
    st === 'pending'
  );
});

const isInReview = computed<boolean>(() => {
  const st = requestStatus.value;
  if (st === null || st === undefined) return false;
  return (
    Number(st) === REQUEST_STATUS.IN_REVIEW ||
    st === 'REQUEST_STATUS_IN_REVIEW' ||
    st === 2 ||
    st === 'in_review'
  );
});

const isApproved = computed<boolean>(() => {
  const st = requestStatus.value;
  if (st === null || st === undefined) return false;
  return (
    Number(st) === REQUEST_STATUS.APPROVED ||
    st === 'REQUEST_STATUS_APPROVED' ||
    st === 3 ||
    st === 'approved'
  );
});

const isRejected = computed<boolean>(() => {
  const st = requestStatus.value;
  if (st === null || st === undefined) return false;
  return (
    Number(st) === REQUEST_STATUS.REJECTED ||
    st === 'REQUEST_STATUS_REJECTED' ||
    st === 4 ||
    st === 'rejected'
  );
});

const projectIconUrl = computed<string>(() => {
  const path = projectData.value?.iconPath || projectData.value?.icon_path;
  if (!path) return '';
  return getMediaUrl(path, projectId.value);
});

const projectCoverUrl = computed<string>(() => {
  const path = projectData.value?.coverPath || projectData.value?.cover_path;
  if (!path) return '';
  return getMediaUrl(path, projectId.value);
});

const projectVideoUrl = computed<string>(() => {
  const path = projectData.value?.videoPath || projectData.value?.video_path;
  if (!path) return '';
  return getMediaUrl(path, projectId.value);
});

const isProjectOnline = computed<boolean>(() => {
  return !!(
    projectData.value?.isOnline ||
    projectData.value?.is_online ||
    activeRequest.value?.snapshot?.isOnline ||
    activeRequest.value?.snapshot?.is_online
  );
});

const projectTitle = computed<string>(() => {
  return (
    projectData.value?.titleRu ||
    projectData.value?.title_ru ||
    projectData.value?.titleEn ||
    projectData.value?.title_en ||
    `Проект #${projectId.value}`
  );
});

const ownerUserId = computed<any>(() => {
  return activeRequest.value?.ownerId || projectData.value?.ownerId || projectData.value?.owner_id;
});

const canSubmitReject = computed<boolean>(() => {
  if (violations.value.length === 0) return false;
  return violations.value.every(
    (v) => v.ruleCode.trim() && v.ruleTitle.trim() && v.description.trim()
  );
});

function getFileName(filePath?: string): string {
  if (!filePath) return '';
  return filePath.split('/').pop() || filePath;
}

// Управление отметками проверенности шагов
function toggleStepVerified(stepId: number): void {
  verifiedSteps.value[stepId] = !verifiedSteps.value[stepId];
  saveToStorage();
}

function nextStep(): void {
  if (currentStep.value < 5) {
    currentStep.value++;
    syncStepState();
  }
}

function prevStep(): void {
  if (currentStep.value > 1) {
    currentStep.value--;
    syncStepState();
  }
}

function goToStep(step: number): void {
  if (step >= 1 && step <= 5) {
    currentStep.value = step;
    syncStepState();
  }
}

function openVerdict(mode: 'approve' | 'reject'): void {
  verdictMode.value = mode;
  currentStep.value = 5;
  syncStepState();
}

// Конструктор нарушений (режим отклонения)
function onRuleSelected(item: ViolationItem, rule: any): void {
  item.ruleCode = rule.code;
  item.ruleTitle = rule.title;
  if (!item.description.trim()) {
    item.description = rule.summary;
  }
  saveToStorage();
}

function onRuleCleared(item: ViolationItem): void {
  item.ruleCode = '';
  item.ruleTitle = '';
  saveToStorage();
}

function addViolation(): void {
  violations.value.push({
    ruleCode: '',
    ruleTitle: '',
    description: '',
    attachments: [],
    uploading: false,
  });
  saveToStorage();
}

function removeViolation(idx: number): void {
  if (violations.value.length > 1) {
    violations.value.splice(idx, 1);
    saveToStorage();
  }
}

function removeAttachment(item: ViolationItem, idx: number): void {
  item.attachments.splice(idx, 1);
  saveToStorage();
}

function isVideo(att: any): boolean {
  const mt = att.mime_type || att.type || '';
  const n = att.file_name || att.name || '';
  return mt.startsWith('video/') || n.endsWith('.mp4') || n.endsWith('.webm');
}

function formatSize(bytes?: number): string {
  if (!bytes) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(1))} ${sizes[i]}`;
}

async function handleFileUpload(e: any, item: ViolationItem): Promise<void> {
  const file: File | undefined = e.target.files?.[0];
  if (!file) return;

  const isVid = file.type.startsWith('video/');
  const maxBytes = isVid ? 50 * 1024 * 1024 : 15 * 1024 * 1024;
  if (file.size > maxBytes) {
    showToast(
      `Файл слишком велик (${formatSize(file.size)}). Максимум: ${isVid ? '50 МБ' : '15 МБ'}`,
      'danger'
    );
    e.target.value = '';
    return;
  }

  item.uploading = true;
  try {
    const res = await moderationApi.uploadAttachment(projectId.value, file);
    if (res && res.attachment) {
      item.attachments.push({
        id: res.attachment.id,
        project_id: projectId.value,
        file_name: res.attachment.file_name,
        file_size: res.attachment.file_size,
        mime_type: res.attachment.mime_type,
        url: res.attachment.url,
        download_url: res.attachment.download_url,
      });
      showToast(`Файл "${file.name}" прикреплен`, 'success');
      saveToStorage();
    }
  } catch (err) {
    console.error('Failed to upload evidence attachment:', err);
    showToast('Не удалось загрузить файл', 'danger');
  } finally {
    item.uploading = false;
    e.target.value = '';
  }
}

// ПЕРСИСТЕНТНОСТЬ (LocalStorage + URL Query)
function getStorageKeys() {
  const pid = projectId.value;
  const reqId = activeRequest.value?.id || 'active';
  return {
    step: `gdh_mod_step_${pid}`,
    checks: `gdh_mod_checks_${pid}_${reqId}`,
    draft: `gdh_mod_draft_${pid}_${reqId}`,
  };
}

function saveToStorage(): void {
  try {
    const keys = getStorageKeys();
    localStorage.setItem(keys.step, String(currentStep.value));
    localStorage.setItem(keys.checks, JSON.stringify(verifiedSteps.value));
    localStorage.setItem(
      keys.draft,
      JSON.stringify({
        verdictMode: verdictMode.value,
        approveComment: approveComment.value,
        violations: violations.value.map((v) => ({ ...v, uploading: false })),
        generalComment: generalComment.value,
      })
    );
  } catch (e) {
    console.error('Failed to save moderation state to localStorage:', e);
  }
}

function loadFromStorage(): void {
  try {
    const keys = getStorageKeys();

    // 1. Отметки проверенности
    const savedChecks = localStorage.getItem(keys.checks);
    if (savedChecks) {
      verifiedSteps.value = JSON.parse(savedChecks);
    }

    // 2. Черновик вердикта
    const savedDraft = localStorage.getItem(keys.draft);
    if (savedDraft) {
      const parsed = JSON.parse(savedDraft);
      if (parsed.verdictMode === 'approve' || parsed.verdictMode === 'reject') {
        verdictMode.value = parsed.verdictMode;
      }
      if (typeof parsed.approveComment === 'string') {
        approveComment.value = parsed.approveComment;
      }
      if (Array.isArray(parsed.violations) && parsed.violations.length > 0) {
        violations.value = parsed.violations;
      }
      if (typeof parsed.generalComment === 'string') {
        generalComment.value = parsed.generalComment;
      }
    }

    // 3. Шаг и режим вердикта из URL Query или localStorage
    const urlStep = Number(route.query.step);
    const urlVerdict = route.query.verdict;
    if (urlVerdict === 'reject' || urlVerdict === 'approve') {
      verdictMode.value = urlVerdict;
    }

    if (urlStep >= 1 && urlStep <= 5) {
      currentStep.value = urlStep;
    } else {
      const savedStep = Number(localStorage.getItem(keys.step));
      if (savedStep >= 1 && savedStep <= 5) {
        currentStep.value = savedStep;
      }
    }
  } catch (e) {
    console.error('Failed to load moderation state from localStorage:', e);
  }
}

function clearDraftStorage(): void {
  try {
    const keys = getStorageKeys();
    localStorage.removeItem(keys.step);
    localStorage.removeItem(keys.checks);
    localStorage.removeItem(keys.draft);
  } catch (e) {
    console.error('Failed to clear moderation draft storage:', e);
  }
}

function syncStepState(): void {
  saveToStorage();
  router.replace({
    query: {
      ...route.query,
      step: String(currentStep.value),
      verdict: currentStep.value === 5 ? verdictMode.value : undefined,
    },
  });
}

watch(
  () => route.query.step,
  (newStep) => {
    const s = Number(newStep);
    if (s >= 1 && s <= 5 && s !== currentStep.value) {
      currentStep.value = s;
    }
  }
);

watch(
  () => route.query.verdict,
  (newVerdict) => {
    if ((newVerdict === 'approve' || newVerdict === 'reject') && newVerdict !== verdictMode.value) {
      verdictMode.value = newVerdict;
    }
  }
);

async function loadProjectInfo(): Promise<void> {
  loading.value = true;
  noRequestMode.value = false;
  try {
    const data = await moderationApi.getLatestByProject(projectId.value);
    if (data && data.request) {
      activeRequest.value = data.request;
      projectData.value = {
        ...(data.request.snapshot || {}),
        isOnline: Boolean(data.request.snapshot?.isOnline ?? data.request.snapshot?.is_online),
      };
      try {
        fullProject.value = await getProject(projectId.value);
      } catch {
        // non-critical
      }
    } else {
      noRequestMode.value = true;
      activeRequest.value = null;
      try {
        const p = await getProject(projectId.value);
        projectData.value = {
          titleRu: p.title_ru,
          titleEn: p.title_en,
          seoRu: p.seo_ru,
          seoEn: p.seo_en,
          aboutRu: p.about_ru || (p as any).about,
          aboutEn: p.about_en,
          about: p.about_ru || (p as any).about,
          iconPath: p.icon_path,
          coverPath: p.cover_path,
          videoPath: p.video_path,
          devUrl: p.dev_url,
          activeBuildVersion: p.active_build_version || '1.0.0',
          ownerId: p.owner_id || (p as any).ownerId,
          isOnline: Boolean(p.is_online ?? (p as any).isOnline),
        };
      } catch {
        showToast('Проект не найден', 'warning');
        router.push('/moderator/queue');
      }
    }
  } catch (err) {
    console.error('Failed to load project request:', err);
    showToast('Не удалось загрузить данные проекта', 'danger');
  } finally {
    loading.value = false;
    loadFromStorage();
  }
}

async function handleClaim(): Promise<void> {
  if (!activeRequest.value) return;
  actionLoading.value = true;
  try {
    const updated = await moderationStore.claimRequest(activeRequest.value.id);
    activeRequest.value = updated;
    showToast('Проект взят в работу', 'success');
  } catch {
    showToast('Ошибка при взятии проекта в работу', 'danger');
  } finally {
    actionLoading.value = false;
  }
}

async function handleApprove(): Promise<void> {
  actionLoading.value = true;
  try {
    await moderationStore.approveRequest(projectId.value, approveComment.value.trim());
    showToast('Проект одобрен и опубликован', 'success');
    clearDraftStorage();
    await loadProjectInfo();
  } catch (err: any) {
    const msg = err.response?.data?.message || err.message || 'Ошибка при одобрении проекта';
    showToast(msg, 'danger');
  } finally {
    actionLoading.value = false;
  }
}

async function handleReject(): Promise<void> {
  if (!canSubmitReject.value || actionLoading.value) return;

  actionLoading.value = true;
  try {
    const violationItems = violations.value.map((v) => ({
      rule_code: v.ruleCode.trim(),
      rule_title: v.ruleTitle.trim(),
      description: v.description.trim(),
      attachment_ids: v.attachments.map((a) => a.id).filter(Boolean),
      attachments: v.attachments,
    }));

    const primaryReason =
      violations.value[0]?.ruleTitle ||
      violations.value[0]?.ruleCode ||
      'Замечания по модерации';

    await moderationApi.reject(projectId.value, primaryReason, violationItems);

    showToast('Проект отклонен, вердикт с замечаниями отправлен разработчику', 'success');
    clearDraftStorage();
    await loadProjectInfo();
  } catch (err: any) {
    console.error('Failed to reject project:', err);
    const msg = err.response?.data?.message || err.message || 'Ошибка отправки решения';
    showToast(`Не удалось отклонить проект: ${msg}`, 'danger');
  } finally {
    actionLoading.value = false;
  }
}

onMounted(() => {
  loadProjectInfo();
});
</script>

<style scoped>
.moderator-workspace {
  display: flex;
  width: 100%;
  height: calc(100vh - 60px);
  background: var(--bg-app);
  overflow: hidden;
  box-sizing: border-box;
}

/* ЛЕВАЯ / ОСНОВНАЯ КОЛОНКА */
.moderator-main-area {
  flex: 1;
  min-width: 0;
  height: 100%;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  background: var(--bg-app);
  container-type: inline-size;
  container-name: mainarea;
}

/* 1. ФИКСИРОВАННАЯ ВЕРХНЯЯ СТРОКА (ДВЕ СТРОКИ) */
.top-info-bar {
  height: 68px;
  flex-shrink: 0;
  display: flex;
  align-items: stretch;
  padding: 8px 20px;
  background: var(--bg-card);
  border-bottom: 1px solid var(--border);
  z-index: 10;
  gap: 16px;
}

/* Левая колонка: Иконка (растянута на обе строки) */
.top-icon-cell {
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.main-icon-box {
  width: 52px;
  height: 52px;
  border-radius: var(--radius-sm, 8px);
  overflow: hidden;
  background: var(--bg-secondary);
  border: 1px solid var(--border);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.main-icon-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.main-icon-placeholder {
  color: var(--text-tertiary);
  display: flex;
  align-items: center;
  justify-content: center;
}

/* Центральная колонка: 2 строки */
.top-center-cell {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  justify-content: center;
  gap: 6px;
}

.top-row-main {
  display: flex;
  align-items: center;
  gap: 12px;
  min-width: 0;
}

.title-meta-group {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
}

.top-project-title {
  font-size: 15px;
  font-weight: 700;
  color: var(--text-main);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 320px;
}

.project-id-chip {
  font-size: 11px;
  font-family: monospace;
  color: var(--text-tertiary);
  background: var(--bg-secondary);
  border: 1px solid var(--border);
  padding: 1px 6px;
  border-radius: 4px;
  white-space: nowrap;
}

.top-badges-group {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-shrink: 0;
}

.badge-chip {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 11px;
  font-weight: 600;
  padding: 2px 8px;
  border-radius: 12px;
  white-space: nowrap;
}

.badge-pub {
  background: rgba(59, 130, 246, 0.15);
  color: #60a5fa;
  border: 1px solid rgba(59, 130, 246, 0.3);
}

.badge-update {
  background: rgba(168, 85, 247, 0.15);
  color: #c084fc;
  border: 1px solid rgba(168, 85, 247, 0.3);
}

.badge-online {
  background: rgba(14, 165, 233, 0.12);
  color: #38bdf8;
  border: 1px solid rgba(14, 165, 233, 0.25);
}

.badge-offline {
  background: rgba(148, 163, 184, 0.1);
  color: #94a3b8;
  border: 1px solid rgba(148, 163, 184, 0.2);
}

.badge-version {
  background: var(--bg-secondary);
  color: var(--text-muted);
  border: 1px solid var(--border);
}

/* Строка 2 мета-информации */
.top-row-meta {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 12px;
  min-width: 0;
}

.meta-inline-item {
  display: flex;
  align-items: center;
  gap: 5px;
  font-size: 12px;
  white-space: nowrap;
}

.meta-caption {
  color: var(--text-tertiary);
}

.meta-val {
  color: var(--text-main);
  font-weight: 500;
  max-width: 180px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.meta-dot-divider {
  color: var(--border);
  font-size: 11px;
}

/* Правая колонка: Кнопки проверки шага и чата (растянуты на обе строки) */
.top-actions-cell {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-shrink: 0;
}

.top-verify-btn-tall {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  height: 52px;
  padding: 0 16px;
  border-radius: var(--radius-sm, 8px);
  background: var(--bg-secondary);
  border: 1px solid var(--border);
  color: var(--text-muted);
  font-size: 13.5px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.15s ease;
  white-space: nowrap;
}

.top-verify-btn-tall:hover {
  background: var(--bg-tertiary);
  color: var(--text-main);
  border-color: var(--border-hover, #4b5563);
}

.top-verify-btn-tall.is-verified {
  background: rgba(46, 160, 67, 0.12);
  border-color: var(--success, #2f9e44);
  color: var(--success, #2f9e44);
}

.top-verify-btn-tall.is-verified:hover {
  background: rgba(46, 160, 67, 0.2);
}

.chat-toggle-btn-tall {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  height: 52px;
  padding: 0 16px;
  border-radius: var(--radius-sm, 8px);
  background: var(--bg-secondary);
  border: 1px solid var(--border);
  color: var(--text-muted);
  font-size: 13.5px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.15s ease;
  position: relative;
}

.chat-toggle-btn-tall:hover {
  background: var(--bg-tertiary);
  color: var(--text-main);
  border-color: var(--primary);
}

.chat-toggle-btn-tall.active {
  background: rgba(88, 166, 255, 0.12);
  color: var(--primary);
  border-color: var(--primary);
}

.chat-unread-badge {
  background: var(--danger, #f85149);
  color: #fff;
  font-size: 10px;
  font-weight: 700;
  padding: 2px 6px;
  border-radius: 10px;
  line-height: 1;
}

/* 2. ЦЕНТРАЛЬНАЯ ОБЛАСТЬ КОНТЕНТА ШАГОВ */
.step-content-viewport {
  flex: 1;
  min-height: 0;
  overflow: hidden;
  position: relative;
  display: flex;
  flex-direction: column;
}

.step-pane {
  width: 100%;
  height: 100%;
}

.step-pane-scrollable {
  overflow-y: auto;
  padding: 24px 28px 48px;
  box-sizing: border-box;
}

.pane-content-wrapper {
  max-width: 1000px;
  margin: 0 auto;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

/* КАРТОЧКИ ИНФОРМАЦИИ ПО ДИЗАЙН-СИСТЕМЕ */
.card {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: var(--radius-md, 8px);
  padding: 20px;
  box-sizing: border-box;
  box-shadow: var(--shadow-sm);
}

.section-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 16px;
}

.head-title-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.section-head h3 {
  margin: 0;
  font-size: 15px;
  font-weight: 600;
  color: var(--text-main);
}

.count-tag {
  font-size: 12px;
  background: var(--bg-secondary);
  color: var(--text-muted);
  padding: 3px 10px;
  border-radius: 12px;
  border: 1px solid var(--border);
  font-weight: 500;
}

.media-pill {
  font-size: 11px;
  color: var(--text-tertiary);
  background: var(--bg-secondary);
  border: 1px solid var(--border);
  padding: 2px 8px;
  border-radius: 4px;
}

/* ПОЛЯ И СЕТКА ДАННЫХ (СТИЛЬ MODERATION SNAPSHOT) */
.data-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
  margin-bottom: 14px;
}

.data-row:last-child {
  margin-bottom: 0;
}

.data-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.data-label {
  font-size: 12px;
  font-weight: 500;
  color: var(--text-tertiary);
}

.data-box {
  padding: 9px 12px;
  background: var(--bg-secondary);
  border: 1px solid var(--border);
  border-radius: var(--radius-sm, 6px);
  color: var(--text-main);
  font-size: 13.5px;
  min-height: 36px;
  display: flex;
  align-items: center;
  box-sizing: border-box;
}

.data-box.multiline {
  min-height: 60px;
  align-items: flex-start;
  line-height: 1.45;
  white-space: pre-wrap;
}

.data-box.multiline-lg {
  min-height: 96px;
  align-items: flex-start;
  line-height: 1.5;
  white-space: pre-wrap;
}

/* ШАГ 2: ЭФФЕКТИВНЫЙ МЕДИА LAYOUT (ДВЕ КОЛОНКИ ДЛЯ ИКОНКИ И ОБЛОЖКИ) */
.media-two-col-grid {
  display: grid;
  grid-template-columns: 1fr 1.6fr;
  gap: 20px;
}

@media (max-width: 900px) {
  .media-two-col-grid {
    grid-template-columns: 1fr;
  }
}

.media-card-item {
  display: flex;
  flex-direction: column;
}

.media-item-body {
  display: flex;
  flex-direction: column;
  gap: 12px;
  align-items: center;
}

/* Интерактивное превью с эффектом увеличения при наведении */
.interactive-image-preview {
  position: relative;
  border-radius: var(--radius-sm, 6px);
  overflow: hidden;
  border: 1px solid var(--border);
  background: var(--bg-secondary);
  cursor: zoom-in;
  transition: all 0.2s ease;
  width: 100%;
}

.interactive-image-preview:hover {
  border-color: var(--primary);
  box-shadow: 0 0 0 2px rgba(88, 166, 255, 0.2);
}

.interactive-image-preview.icon-aspect {
  max-width: 180px;
  aspect-ratio: 1 / 1;
}

.interactive-image-preview.cover-aspect {
  max-width: 480px;
  aspect-ratio: 800 / 470;
}

.media-fit-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.preview-overlay-hover {
  position: absolute;
  inset: 0;
  background: rgba(13, 17, 23, 0.6);
  backdrop-filter: blur(2px);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 6px;
  color: #fff;
  font-size: 12px;
  font-weight: 500;
  opacity: 0;
  transition: opacity 0.2s ease;
}

.interactive-image-preview:hover .preview-overlay-hover {
  opacity: 1;
}

.media-meta-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  padding-top: 6px;
  border-top: 1px solid var(--border);
}

.media-filename {
  font-family: monospace;
  font-size: 12px;
  color: var(--text-main);
  max-width: 220px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.media-hint {
  font-size: 11px;
  color: var(--text-tertiary);
}

.empty-media-box {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px 16px;
  text-align: center;
  gap: 8px;
  color: var(--text-tertiary);
  background: var(--bg-secondary);
  border-radius: var(--radius-sm, 6px);
  border: 1px dashed var(--border);
}

/* ПРОМО-ВИДЕО */
.video-section-card {
  display: flex;
  flex-direction: column;
}

.video-preview-layout {
  display: grid;
  grid-template-columns: 1fr 260px;
  gap: 20px;
  align-items: flex-start;
}

@media (max-width: 860px) {
  .video-preview-layout {
    grid-template-columns: 1fr;
  }
}

.video-player-frame {
  width: 100%;
  aspect-ratio: 16 / 9;
  border-radius: var(--radius-sm, 6px);
  overflow: hidden;
  background: #000;
  border: 1px solid var(--border);
}

.promo-video-element {
  width: 100%;
  height: 100%;
  display: block;
}

.video-info-aside {
  display: flex;
  flex-direction: column;
  gap: 12px;
  background: var(--bg-secondary);
  border: 1px solid var(--border);
  border-radius: var(--radius-sm, 6px);
  padding: 16px;
}

.meta-item {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.meta-label {
  font-size: 11px;
  font-weight: 500;
  color: var(--text-tertiary);
}

.meta-value {
  font-size: 13px;
  color: var(--text-main);
  word-break: break-all;
}

.btn-video-download {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  height: 34px;
  padding: 0 14px;
  border-radius: var(--radius-sm, 6px);
  background: var(--bg-card);
  border: 1px solid var(--border);
  color: var(--text-main);
  font-size: 12.5px;
  font-weight: 500;
  text-decoration: none;
  margin-top: 6px;
  transition: all 0.15s;
}

.btn-video-download:hover {
  background: var(--bg-tertiary);
  border-color: var(--primary);
  color: var(--primary);
}

.empty-video-inline {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 14px 16px;
  background: var(--bg-secondary);
  border-radius: var(--radius-sm, 6px);
  color: var(--text-tertiary);
  font-size: 13px;
}

/* ШАГ 3: ТАБЛИЦА ВНУТРИИГРОВЫХ ТОВАРОВ */
.iap-card {
  padding: 0;
  overflow: hidden;
}

.iap-card .section-head {
  padding: 16px 20px;
  margin-bottom: 0;
  border-bottom: 1px solid var(--border);
}

.table-responsive {
  width: 100%;
  overflow-x: auto;
}

.table-iap {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
  text-align: left;
}

.table-iap th {
  background: var(--bg-secondary);
  color: var(--text-tertiary);
  font-size: 12px;
  font-weight: 600;
  padding: 12px 16px;
  border-bottom: 1px solid var(--border);
}

.table-iap td {
  padding: 12px 16px;
  border-bottom: 1px solid var(--border);
  color: var(--text-main);
  vertical-align: middle;
}

.table-iap tbody tr:last-child td {
  border-bottom: none;
}

.table-iap tbody tr:hover {
  background: var(--bg-secondary);
}

.iap-thumb {
  width: 40px;
  height: 40px;
  border-radius: var(--radius-sm, 6px);
  object-fit: cover;
  border: 1px solid var(--border);
  cursor: zoom-in;
  transition: transform 0.15s;
}

.iap-thumb:hover {
  transform: scale(1.05);
  border-color: var(--primary);
}

.iap-thumb-mock {
  width: 40px;
  height: 40px;
  border-radius: var(--radius-sm, 6px);
  background: var(--bg-secondary);
  border: 1px solid var(--border);
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--text-tertiary);
}

.item-name-ru {
  font-size: 13.5px;
  font-weight: 600;
  color: var(--text-main);
}

.sub-name {
  font-size: 12px;
  color: var(--text-tertiary);
  margin-top: 2px;
}

.col-id code {
  background: var(--bg-secondary);
  border: 1px solid var(--border);
  padding: 2px 6px;
  border-radius: 4px;
  font-family: monospace;
  font-size: 12px;
  color: var(--text-muted);
}

.price-val {
  font-weight: 700;
  color: #eab308;
  margin-right: 4px;
  font-size: 14px;
}

.price-currency {
  font-size: 12px;
  color: var(--text-tertiary);
}

.iap-type-pill {
  font-size: 11px;
  font-weight: 500;
  padding: 3px 8px;
  border-radius: 12px;
  background: var(--bg-secondary);
  color: var(--text-muted);
  border: 1px solid var(--border);
}

.empty-items-box {
  padding: 48px 20px;
  text-align: center;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10px;
}

.empty-items-box h4 {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: var(--text-main);
}

.empty-items-box p {
  margin: 0;
  font-size: 13.5px;
  color: var(--text-tertiary);
  max-width: 440px;
}

/* ШАГ 4: ПЕСОЧНИЦА */
.step-pane-sandbox {
  display: flex;
  flex-direction: column;
  overflow: hidden;
  height: 100%;
  padding: 12px;
  box-sizing: border-box;
}

.sandbox-player-wrapper {
  flex: 1;
  min-height: 0;
  height: 100%;
  border-radius: var(--radius-sm, 6px);
  overflow: hidden;
  border: 1px solid var(--border);
  background: #000;
}

.empty-sandbox-state {
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12px;
  color: var(--text-tertiary);
}

/* 3. ФИКСИРОВАННАЯ НИЖНЯЯ ПАНЕЛЬ НАВИГАЦИИ */
.bottom-nav-bar {
  height: 60px;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 16px;
  background: var(--bg-card);
  border-top: 1px solid var(--border);
  z-index: 10;
  gap: 12px;
  box-sizing: border-box;
}

.nav-left {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}

.nav-right {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 8px;
  flex-shrink: 0;
}

.btn-nav {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  height: 38px;
  padding: 0 16px;
  border-radius: var(--radius-sm, 6px);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s ease;
  white-space: nowrap;
  flex-shrink: 0;
}

.btn-nav:disabled {
  opacity: 0.35;
  cursor: not-allowed;
}

.btn-secondary {
  background: var(--bg-secondary);
  border: 1px solid var(--border);
  color: var(--text-main);
}

.btn-secondary:hover:not(:disabled) {
  background: var(--bg-tertiary);
  border-color: var(--text-tertiary);
}

.btn-primary {
  background: var(--success, #2f9e44);
  border: 1px solid transparent;
  color: #fff;
}

.btn-primary:hover:not(:disabled) {
  opacity: 0.9;
}

/* СТЕППЕР В ЦЕНТРЕ */
.nav-stepper {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  flex: 1;
  min-width: 0;
  padding: 0 4px;
}

.stepper-step {
  display: flex;
  align-items: center;
  gap: 6px;
  background: transparent;
  border: none;
  cursor: pointer;
  padding: 4px 6px;
  border-radius: var(--radius-sm, 6px);
  color: var(--text-tertiary);
  transition: all 0.15s;
  white-space: nowrap;
  flex-shrink: 0;
}

.stepper-step:hover {
  color: var(--text-main);
}

.step-circle {
  width: 24px;
  height: 24px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 11px;
  font-weight: 700;
  background: var(--bg-secondary);
  border: 1px solid var(--border);
  color: var(--text-muted);
  transition: all 0.15s;
  flex-shrink: 0;
}

.step-text {
  font-size: 12px;
  font-weight: 500;
  white-space: nowrap;
}

.stepper-step.active .step-circle {
  background: var(--primary);
  border-color: var(--primary);
  color: #fff;
}

.stepper-step.active .step-text {
  color: var(--text-main);
  font-weight: 600;
}

.stepper-step.completed .step-circle {
  background: rgba(46, 160, 67, 0.15);
  border-color: var(--success, #2f9e44);
  color: var(--success, #2f9e44);
}

.stepper-line {
  width: 16px;
  height: 2px;
  background: var(--border);
  border-radius: 1px;
  transition: background 0.15s;
  flex-shrink: 1;
  min-width: 4px;
}

.stepper-line.filled {
  background: var(--success, #2f9e44);
}


/* КНОПКИ РЕШЕНИЯ НА 4-М ШАГЕ */
.step-verdict-actions {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}

.btn-verdict {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  height: 38px;
  padding: 0 14px;
  border-radius: var(--radius-sm, 6px);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s ease;
  white-space: nowrap;
  flex-shrink: 0;
}

.btn-verdict:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-claim {
  background: rgba(59, 130, 246, 0.15);
  border: 1px solid rgba(59, 130, 246, 0.35);
  color: #60a5fa;
}

.btn-claim:hover:not(:disabled) {
  background: rgba(59, 130, 246, 0.25);
}

.btn-approve-action {
  background: var(--success, #2f9e44);
  border: 1px solid transparent;
  color: #fff;
}

.btn-approve-action:hover:not(:disabled) {
  opacity: 0.9;
}

.btn-reject-action {
  background: rgba(239, 68, 68, 0.12);
  border: 1px solid rgba(239, 68, 68, 0.35);
  color: #f87171;
}

.btn-reject-action:hover:not(:disabled) {
  background: rgba(239, 68, 68, 0.25);
}

.badge-verdict-done {
  padding: 6px 12px;
  font-size: 13px;
}

/* ПРАВАЯ СКРЫВАЕМАЯ КОЛОНКА ЧАТА */
.moderator-chat-aside {
  width: 520px;
  min-width: 420px;
  max-width: 640px;
  flex-shrink: 0;
  height: 100%;
  border-left: 1px solid var(--border);
  background: var(--bg-card);
  transition: width 0.2s ease;
  overflow: hidden;
}

@media (min-width: 1700px) {
  .moderator-chat-aside {
    width: 600px;
    max-width: 720px;
  }
}

@media (max-width: 1366px) {
  .moderator-chat-aside {
    width: 440px;
    min-width: 380px;
  }
}

/* ЗАГРУЗКА */
.state-loading-screen {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: calc(100vh - 60px);
  gap: 16px;
  color: var(--text-tertiary);
}

.spinner-md {
  width: 36px;
  height: 36px;
  border: 3px solid rgba(255, 255, 255, 0.1);
  border-top-color: var(--primary);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}


/* ШАГ 5: ВЕРДИКТ И РЕШЕНИЕ МОДЕРАТОРА */

.approve-flow-card {
  display: flex;
  flex-direction: column;
  gap: 18px;
}

.approve-banner {
  background: rgba(46, 160, 67, 0.08);
  border: 1px solid rgba(46, 160, 67, 0.25);
  border-radius: var(--radius-sm, 6px);
  padding: 14px 18px;
}

.approve-banner p {
  margin: 0;
  font-size: 13.5px;
  color: var(--text-main);
  line-height: 1.5;
}

.verdict-card-actions {
  display: flex;
  justify-content: flex-end;
  margin-top: 8px;
}

.btn-submit-approve {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  height: 42px;
  padding: 0 24px;
  border-radius: var(--radius-sm, 6px);
  background: var(--success, #2f9e44);
  border: none;
  color: #fff;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: opacity 0.15s;
}

.btn-submit-approve:hover:not(:disabled) {
  opacity: 0.9;
}

.btn-submit-approve:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* КОНСТРУКТОР НАРУШЕНИЙ ДЛЯ ОТКЛОНЕНИЯ */
.reject-flow-container {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.header-warning-banner {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  background: rgba(248, 81, 73, 0.08);
  border: 1px solid rgba(248, 81, 73, 0.25);
  padding: 14px 18px;
  border-radius: var(--radius-sm, 6px);
}

.banner-icon {
  margin-top: 2px;
  flex-shrink: 0;
}

.header-warning-banner h4 {
  margin: 0 0 4px;
  font-size: 14px;
  font-weight: 600;
  color: #f85149;
}

.header-warning-banner p {
  margin: 0;
  font-size: 13px;
  color: var(--text-muted);
  line-height: 1.4;
}

.violations-list {
  display: flex;
  flex-direction: column;
  gap: 18px;
}

.violation-box {
  background: var(--bg-secondary);
  border: 1px solid var(--border);
  border-left: 3px solid #f85149;
  border-radius: var(--radius-sm, 6px);
  padding: 18px;
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.violation-box-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.violation-badge {
  display: flex;
  align-items: center;
  gap: 8px;
}

.number-tag {
  font-size: 12px;
  font-weight: 700;
  background: rgba(248, 81, 73, 0.2);
  color: #f85149;
  padding: 2px 8px;
  border-radius: 4px;
}

.rule-preview-tag {
  font-size: 12px;
  font-family: monospace;
  background: var(--bg-app);
  color: var(--text-main);
  padding: 2px 6px;
  border-radius: 4px;
  border: 1px solid var(--border);
}

.btn-remove-box {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  background: transparent;
  border: none;
  color: #f85149;
  font-size: 12px;
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 4px;
  transition: background 0.15s;
}

.btn-remove-box:hover {
  background: rgba(248, 81, 73, 0.15);
}

.rule-selector-field {
  margin-bottom: 4px;
}

.form-grid-two {
  display: grid;
  grid-template-columns: 1fr 2fr;
  gap: 14px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.form-label {
  font-size: 12px;
  font-weight: 500;
  color: var(--text-tertiary);
}

.req {
  color: #f85149;
}

.form-input {
  height: 38px;
  background: var(--bg-secondary);
  border: 1px solid var(--border);
  border-radius: var(--radius-sm, 6px);
  padding: 0 12px;
  color: var(--text-main);
  font-size: 13.5px;
  font-family: inherit;
  box-sizing: border-box;
}

.form-input.code-font {
  font-family: monospace;
}

.form-input:focus,
.form-textarea:focus {
  outline: none;
  border-color: var(--primary);
}

.form-textarea {
  width: 100%;
  background: var(--bg-secondary);
  border: 1px solid var(--border);
  border-radius: var(--radius-sm, 6px);
  padding: 10px 12px;
  color: var(--text-main);
  font-size: 13.5px;
  font-family: inherit;
  line-height: 1.45;
  resize: vertical;
  box-sizing: border-box;
}

.evidence-block {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.evidence-items-wrap {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 10px;
}

.evidence-chip {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  background: var(--bg-app);
  border: 1px solid var(--border);
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 12px;
}

.chip-name {
  max-width: 160px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: var(--text-main);
}

.chip-size {
  color: var(--text-tertiary);
  font-size: 11px;
}

.btn-delete-chip {
  background: transparent;
  border: none;
  color: var(--text-tertiary);
  cursor: pointer;
  padding: 0;
  display: flex;
  align-items: center;
}

.btn-delete-chip:hover {
  color: #f85149;
}

.btn-upload-evidence {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  background: var(--bg-app);
  border: 1px dashed var(--border);
  color: var(--text-muted);
  padding: 6px 12px;
  border-radius: var(--radius-sm, 6px);
  font-size: 12.5px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s;
}

.btn-upload-evidence:hover:not(.is-loading) {
  border-color: var(--primary);
  color: var(--primary);
}

.btn-upload-evidence.is-loading {
  opacity: 0.6;
  cursor: wait;
}

.file-hidden-input {
  display: none;
}

.add-more-row {
  display: flex;
  justify-content: flex-start;
  margin-top: 14px;
}

.btn-add-violation {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  background: transparent;
  border: 1px dashed var(--border);
  color: var(--text-muted);
  padding: 8px 14px;
  border-radius: var(--radius-sm, 6px);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s;
}

.btn-add-violation:hover {
  border-color: var(--primary);
  color: var(--primary);
}

.reject-submit-bar {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 16px;
  margin-top: 4px;
}

.validation-tip {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: var(--text-tertiary);
}

.btn-submit-reject {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  height: 42px;
  padding: 0 24px;
  border-radius: var(--radius-sm, 6px);
  background: #f85149;
  border: none;
  color: #fff;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: opacity 0.15s;
}

.btn-submit-reject:hover:not(:disabled) {
  opacity: 0.9;
}

.btn-submit-reject:disabled {
  opacity: 0.45;
  cursor: not-allowed;
}

.verdict-result-card,
.verdict-pending-card {
  padding: 24px;
}

.pending-prompt-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  gap: 12px;
  padding: 32px 0;
}

.pending-prompt-content h3 {
  margin: 0;
  font-size: 16px;
  color: var(--text-main);
}

.pending-prompt-content p {
  margin: 0;
  font-size: 13.5px;
  color: var(--text-tertiary);
}

.verdict-summary-comment {
  margin-top: 16px;
}

.summary-label {
  margin: 0 0 6px;
  font-size: 12px;
  color: var(--text-tertiary);
}
</style>
