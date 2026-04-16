<template>
  <div class="settings-page">
    <div class="settings-container">
      <div class="settings-header">
        <button class="back-btn" @click="$router.back()">
          <ArrowLeft class="icon-sm" />
          Назад
        </button>
        <h1>Настройки</h1>
      </div>

      <!-- Табы навигации -->
      <div class="tabs-nav">
        <button :class="{ active: activeTab === 'general' }" @click="activeTab = 'general'">
          <Settings2 class="icon-sm" /> Общие
        </button>
        <button :class="{ active: activeTab === 'account' }" @click="activeTab = 'account'">
          <User class="icon-sm" /> Аккаунт
        </button>
      </div>

      <!-- ОБЩИЕ НАСТРОЙКИ -->
      <div class="settings-content" v-if="activeTab === 'general'">
        <!-- Секция: Язык интерфейса -->
        <section class="settings-section">
          <h2 class="section-title">Язык интерфейса</h2>
          <div class="option-list">
            <label class="option-item" :class="{ active: selectedLang === 'ru' }">
              <input type="radio" v-model="selectedLang" value="ru" disabled />
              <div class="option-content">
                <span class="option-label">Русский</span>
                <span class="option-badge">Текущий</span>
              </div>
            </label>
            <label class="option-item" :class="{ active: selectedLang === 'en' }">
              <input type="radio" v-model="selectedLang" value="en" disabled />
              <div class="option-content">
                <span class="option-label">English</span>
                <span class="option-badge coming-soon">Скоро</span>
              </div>
            </label>
            <label class="option-item" :class="{ active: selectedLang === 'zh' }">
              <input type="radio" v-model="selectedLang" value="zh" disabled />
              <div class="option-content">
                <span class="option-label">中文</span>
                <span class="option-badge coming-soon">Скоро</span>
              </div>
            </label>
          </div>
        </section>

        <!-- Секция: Тема оформления -->
        <section class="settings-section">
          <h2 class="section-title">Тема оформления</h2>
          <div class="theme-grid">
            <div class="theme-card" :class="{ active: theme === 'light' }" @click="theme = 'light'">
              <div class="theme-preview theme-preview-light">
                <div class="preview-header"></div>
                <div class="preview-body">
                  <div class="preview-card"></div>
                  <div class="preview-line"></div>
                  <div class="preview-line short"></div>
                </div>
              </div>
              <div class="theme-info">
                <Sun class="icon-md theme-icon" />
                <span>Светлая</span>
              </div>
            </div>
            <div class="theme-card" :class="{ active: theme === 'dark' }" @click="theme = 'dark'">
              <div class="theme-preview theme-preview-dark">
                <div class="preview-header"></div>
                <div class="preview-body">
                  <div class="preview-card"></div>
                  <div class="preview-line"></div>
                  <div class="preview-line short"></div>
                </div>
              </div>
              <div class="theme-info">
                <Moon class="icon-md theme-icon" />
                <span>Тёмная</span>
              </div>
            </div>
          </div>
        </section>

        <!-- Секция: Местоположение -->
        <section class="settings-section">
          <h2 class="section-title">Местоположение</h2>
          <div class="location-list">
            <label class="location-item" :class="{ active: location === 'moscow' }">
              <input type="radio" v-model="location" value="moscow" disabled />
              <div class="location-content">
                <MapPin class="location-icon" />
                <div class="location-info">
                  <span class="location-name">Москва</span>
                  <span class="location-meta">UTC+3 • По умолчанию</span>
                </div>
                <span class="location-badge">Текущий</span>
              </div>
            </label>
            <label class="location-item" :class="{ active: location === 'petersburg' }">
              <input type="radio" v-model="location" value="petersburg" disabled />
              <div class="location-content">
                <MapPin class="location-icon" />
                <div class="location-info">
                  <span class="location-name">Санкт-Петербург</span>
                  <span class="location-meta">UTC+3</span>
                </div>
                <span class="location-badge coming-soon">Скоро</span>
              </div>
            </label>
            <label class="location-item" :class="{ active: location === 'ekb' }">
              <input type="radio" v-model="location" value="ekb" disabled />
              <div class="location-content">
                <MapPin class="location-icon" />
                <div class="location-info">
                  <span class="location-name">Екатеринбург</span>
                  <span class="location-meta">UTC+5</span>
                </div>
                <span class="location-badge coming-soon">Скоро</span>
              </div>
            </label>
            <label class="location-item" :class="{ active: location === 'novosibirsk' }">
              <input type="radio" v-model="location" value="novosibirsk" disabled />
              <div class="location-content">
                <MapPin class="location-icon" />
                <div class="location-info">
                  <span class="location-name">Новосибирск</span>
                  <span class="location-meta">UTC+7</span>
                </div>
                <span class="location-badge coming-soon">Скоро</span>
              </div>
            </label>
          </div>
        </section>
      </div>

      <!-- АККАУНТ -->
      <div class="settings-content" v-if="activeTab === 'account'">
        <!-- Секция: Dev-среда -->
        <section class="settings-section">
          <h2 class="section-title">Dev-среда</h2>
          <p class="section-desc">Данные для входа в тестовую среду. Эти данные используются во всех проектах.</p>
          <div class="dev-env-box">
            <div class="dev-header"><MonitorPlay class="icon-sm text-primary" /> <strong>Dev-среда</strong></div>
            <p class="text-sm text-muted mb-16">Используйте эти данные для входа в игру:</p>
            <div class="credentials-row">
              <div class="cred-item"><span class="text-muted">Логин:</span> <code class="code-val">welwise</code></div>
              <div class="cred-item"><span class="text-muted">Пароль:</span> <code class="code-val">1txqmYkZ-R</code></div>
            </div>
          </div>
        </section>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, watch } from 'vue'
import { ArrowLeft, Sun, Moon, MapPin, Settings2, User, MonitorPlay } from 'lucide-vue-next'

const activeTab = ref('general')
const selectedLang = ref('ru')
const theme = ref('light')
const location = ref('moscow')

watch(theme, (newTheme) => {
  if (newTheme === 'dark') {
    document.documentElement.setAttribute('data-theme', 'dark')
  } else {
    document.documentElement.removeAttribute('data-theme')
  }
})
</script>

<style scoped>
.settings-page {
  min-height: 100vh;
  background: var(--bg-app);
  padding: 32px 24px;
}

.settings-container {
  max-width: 720px;
  margin: 0 auto;
}

.settings-header {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 8px;
  margin-bottom: 32px;
}

.back-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  background: transparent;
  border: none;
  color: var(--text-muted);
  font-weight: 500;
  cursor: pointer;
  padding: 4px 0;
  transition: color 0.2s;
}

.back-btn:hover {
  color: var(--primary);
}

.settings-header h1 {
  font-size: 1.75rem;
  font-weight: 700;
  color: var(--text-main);
  margin: 0;
}

/* Табы навигации */
.tabs-nav {
  display: flex;
  gap: 8px;
  margin-bottom: 24px;
}

.tabs-nav button {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 20px;
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  background: var(--bg-card);
  color: var(--text-muted);
  font-weight: 500;
  font-size: 0.9rem;
  cursor: pointer;
  transition: all 0.2s;
}

.tabs-nav button:hover {
  background: var(--bg-hover);
  color: var(--text-main);
}

.tabs-nav button.active {
  background: var(--bg-secondary);
  border-color: var(--primary);
  color: var(--primary);
}

.settings-content {
  display: flex;
  flex-direction: column;
  gap: 32px;
}

.settings-section {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: var(--radius-lg);
  padding: 24px;
}

.section-title {
  font-size: 1.1rem;
  font-weight: 600;
  color: var(--text-main);
  margin: 0 0 16px 0;
}

/* Язык интерфейса */
.option-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.option-item {
  display: flex;
  align-items: center;
  padding: 12px 16px;
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  cursor: pointer;
  transition: all 0.2s;
  background: var(--bg-card);
}

.option-item input {
  display: none;
}

.option-item:not(.active):hover {
  background: var(--bg-hover);
  border-color: var(--primary);
}

.option-item.active {
  border-color: var(--primary);
  background: var(--bg-secondary);
}

.option-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
}

.option-label {
  font-weight: 500;
  color: var(--text-main);
}

.option-badge {
  padding: 2px 10px;
  border-radius: 12px;
  font-size: 0.75rem;
  font-weight: 600;
  background: var(--success-light);
  color: var(--success);
}

.option-badge.coming-soon {
  background: var(--warning-light);
  color: var(--warning);
}

/* Тема оформления */
.theme-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
}

.theme-card {
  border: 2px solid var(--border);
  border-radius: var(--radius-md);
  cursor: pointer;
  transition: all 0.2s;
  overflow: hidden;
}

.theme-card:hover {
  border-color: var(--primary);
}

.theme-card.active {
  border-color: var(--primary);
  box-shadow: 0 0 0 1px var(--primary);
}

.theme-preview {
  height: 120px;
  padding: 8px;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.theme-preview-light {
  background: #F3F4F6;
}

.theme-preview-light .preview-header {
  height: 16px;
  background: #FFFFFF;
  border-radius: 4px;
  border: 1px solid #E5E7EB;
}

.theme-preview-light .preview-body {
  flex: 1;
  background: #FFFFFF;
  border-radius: 4px;
  border: 1px solid #E5E7EB;
  padding: 6px;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.theme-preview-light .preview-card {
  height: 16px;
  background: #F3F4F6;
  border-radius: 3px;
}

.theme-preview-light .preview-line {
  height: 6px;
  background: #E5E7EB;
  border-radius: 3px;
}

.theme-preview-light .preview-line.short {
  width: 60%;
}

.theme-preview-dark {
  background: #1F2937;
}

.theme-preview-dark .preview-header {
  height: 16px;
  background: #111827;
  border-radius: 4px;
  border: 1px solid #374151;
}

.theme-preview-dark .preview-body {
  flex: 1;
  background: #111827;
  border-radius: 4px;
  border: 1px solid #374151;
  padding: 6px;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.theme-preview-dark .preview-card {
  height: 16px;
  background: #1F2937;
  border-radius: 3px;
}

.theme-preview-dark .preview-line {
  height: 6px;
  background: #374151;
  border-radius: 3px;
}

.theme-preview-dark .preview-line.short {
  width: 60%;
}

.theme-info {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 12px;
  background: var(--bg-secondary);
  font-weight: 500;
  color: var(--text-main);
}

.theme-icon {
  color: var(--primary);
}

/* Местоположение */
.location-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.location-item {
  display: flex;
  align-items: center;
  padding: 12px 16px;
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  cursor: pointer;
  transition: all 0.2s;
  background: var(--bg-card);
}

.location-item input {
  display: none;
}

.location-item:not(.active):hover {
  background: var(--bg-hover);
  border-color: var(--primary);
}

.location-item.active {
  border-color: var(--primary);
  background: var(--bg-secondary);
}

.location-content {
  display: flex;
  align-items: center;
  gap: 12px;
  width: 100%;
}

.location-icon {
  width: 20px;
  height: 20px;
  color: var(--primary);
  flex-shrink: 0;
}

.location-info {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.location-name {
  font-weight: 500;
  color: var(--text-main);
}

.location-meta {
  font-size: 0.8rem;
  color: var(--text-muted);
}

.location-badge {
  padding: 2px 10px;
  border-radius: 12px;
  font-size: 0.75rem;
  font-weight: 600;
  background: var(--success-light);
  color: var(--success);
  flex-shrink: 0;
}

.location-badge.coming-soon {
  background: var(--warning-light);
  color: var(--warning);
}

/* Dev-среда */
.dev-env-box { background: var(--bg-secondary); border: 1px solid var(--border); border-radius: var(--radius-md); padding: 20px; }
.dev-header { display: flex; align-items: center; gap: 8px; margin-bottom: 8px; font-size: 1.1rem; }
.text-primary { color: var(--primary); }
.text-muted { color: var(--text-muted); }
.text-sm { font-size: 0.85rem; }
.mb-16 { margin-bottom: 16px; margin-top: 0; }
.section-desc { font-size: 0.85rem; color: var(--text-muted); margin: 0 0 16px 0; }

.credentials-row { display: flex; gap: 16px; flex-wrap: wrap; }
.cred-item { background: var(--bg-card); border: 1px solid var(--border); padding: 8px 16px; border-radius: 6px; font-size: 0.9rem; display: flex; align-items: center; gap: 8px;}
.code-val { font-family: monospace; font-weight: 600; color: var(--text-main); background: var(--bg-secondary); padding: 2px 6px; border-radius: 4px; letter-spacing: 0.5px;}

/* Адаптивность */
@media (max-width: 640px) {
  .settings-page {
    padding: 24px 16px;
  }

  .theme-grid {
    grid-template-columns: 1fr;
  }
}
</style>
