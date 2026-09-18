<template>
  <div class="docs-viewport-container">
    <div class="docs-split-layout">
      <!-- ─── ЛЕВЫЙ САЙДБАР (ПРИЖАТ К ЛЕВОМУ КРАЮ, ФИКСИРОВАННЫЙ СКРОЛЛ) ─── -->
      <aside class="docs-sidebar-fixed">
        <div class="sidebar-top-brand">
          <div class="brand-badge-row">
            <BookOpen class="icon-sm text-primary" />
            <span class="brand-title">База знаний</span>
          </div>
        </div>

        <!-- Поиск по всем разделам и статьям -->
        <div class="sidebar-search-wrap">
          <Search class="icon-xs text-muted" />
          <input
            v-model="sidebarSearch"
            type="text"
            class="sidebar-search-input"
            placeholder="Поиск по статьям..."
          />
          <button
            v-if="sidebarSearch"
            type="button"
            class="btn-clear-search"
            @click="sidebarSearch = ''"
          >
            <X class="icon-xs" />
          </button>
        </div>

        <!-- Навигационное дерево разделов и вложенных статей -->
        <nav class="sidebar-nav-tree">
          <div
            v-for="section in filteredSections"
            :key="section.id"
            class="nav-root-group"
          >
            <!-- Корневой раздел -->
            <button
              type="button"
              class="root-section-btn"
              :class="{
                active: currentSectionId === section.id,
                'is-expanded': isSectionExpanded(section.id),
              }"
              @click="handleRootSectionClick(section)"
            >
              <component :is="getSectionIcon(section.icon)" class="icon-xs section-icon" />
              <span class="root-title">{{ section.title }}</span>
              <ChevronDown
                class="icon-xs chevron-toggle"
                :class="{ 'chevron-open': isSectionExpanded(section.id) }"
              />
            </button>

            <!-- Вложенные статьи раздела -->
            <div
              v-if="isSectionExpanded(section.id)"
              class="nested-articles-list"
            >
              <button
                v-for="art in section.articles"
                :key="art.id"
                type="button"
                class="nested-art-btn"
                :class="{ active: currentSectionId === section.id && currentArticleId === art.id }"
                @click="navigateToArticle(section.id, art.id)"
              >
                <span class="art-dot"></span>
                <span class="art-label">{{ art.title }}</span>
              </button>
            </div>
          </div>
        </nav>
      </aside>

      <!-- ─── ЦЕНТРАЛЬНЫЙ СКРОЛЛЯЩИЙСЯ КОНТЕНТ ────────────────────────── -->
      <main class="docs-main-scrollable" ref="contentScrollRef">
        <div class="content-inner-wrapper">
          <!-- Верхняя микронавигация: отображается ТОЛЬКО если мы внутри вложенной статьи -->
          <div v-if="isInsideNestedArticle" class="article-breadcrumb-bar">
            <div class="breadcrumbs-trail">
              <button
                type="button"
                class="crumb-root-link"
                @click="navigateToSectionRoot(currentSectionId)"
              >
                {{ currentSection?.title }}
              </button>
              <ChevronRight class="icon-xs text-muted" />
              <span class="crumb-leaf">{{ currentArticle?.title }}</span>
            </div>

            <!-- Метка даты изменения статьи -->
            <div class="last-updated-badge" :title="`Дата актуализации: ${currentArticleLastUpdated}`">
              <Clock class="icon-xs text-muted" />
              <span>Редакция от <strong>{{ formattedArticleDate }}</strong></span>
            </div>
          </div>

          <!-- Если мы в корне раздела без выбранной статьи: Шапка раздела -->
          <div v-else class="section-overview-hero">
            <div class="hero-meta-row">
              <span class="hero-section-kicker">Раздел платформы</span>
              <div class="last-updated-badge">
                <Clock class="icon-xs text-muted" />
                <span>Редакция от <strong>{{ formattedSectionDate }}</strong></span>
              </div>
            </div>
            <h1 class="hero-main-title">{{ currentSection?.title }}</h1>
            <p class="hero-description">{{ currentSection?.description }}</p>
          </div>

          <!-- ─── 1. СПЕЦИАЛЬНЫЙ РЕЖИМ: КАТАЛОГ ПРАВИЛ И РЕГЛАМЕНТА ──── -->
          <div v-if="isRulesCatalogView" class="rules-catalog-view">
            <div v-if="isInsideNestedArticle" class="article-title-block">
              <h1 class="article-h1">{{ currentArticle?.title }}</h1>
              <p class="article-lead">{{ currentSection?.description }}</p>
            </div>

            <!-- Панель поиска и фильтрации правил -->
            <div class="rules-filter-toolbar">
              <div class="rules-search-box">
                <Search class="icon-sm text-primary" />
                <input
                  v-model="rulesSearchQuery"
                  type="text"
                  class="rules-search-input"
                  placeholder="Поиск по коду (SEC-04, SRV-02), ключевым словам или описанию..."
                />
                <button
                  v-if="rulesSearchQuery"
                  type="button"
                  class="btn-clear-search"
                  @click="rulesSearchQuery = ''"
                >
                  <X class="icon-xs" />
                </button>
              </div>

              <!-- Категории правил -->
              <div class="rules-categories-row">
                <button
                  type="button"
                  class="cat-chip-btn"
                  :class="{ active: selectedRuleCategory === 'all' }"
                  @click="selectedRuleCategory = 'all'"
                >
                  Все категории ({{ PLATFORM_RULES.length }})
                </button>
                <button
                  v-for="cat in ruleCategories"
                  :key="cat.id"
                  type="button"
                  class="cat-chip-btn"
                  :class="{ active: selectedRuleCategory === cat.id }"
                  @click="selectedRuleCategory = cat.id"
                >
                  {{ cat.title }} ({{ cat.rules.length }})
                </button>
              </div>
            </div>

            <!-- Список правил платформы -->
            <div class="rules-cards-stack">
              <div v-if="displayedRules.length === 0" class="no-rules-found">
                <AlertCircle class="icon-md text-muted" />
                <h3>Правила не найдены</h3>
                <p>По запросу «{{ rulesSearchQuery }}» совпадений нет.</p>
                <button
                  type="button"
                  class="btn-reset-search"
                  @click="resetRuleFilters"
                >
                  Сбросить фильтр
                </button>
              </div>

              <div
                v-for="rule in displayedRules"
                :id="rule.code"
                :key="rule.code"
                class="rule-box-card"
                :class="{
                  'is-target-highlighted': highlightedRuleCode === rule.code,
                  'is-critical-rule': rule.severity === 'critical',
                }"
              >
                <div class="rule-box-header">
                  <div class="rule-title-wrap">
                    <span class="rule-code-chip" :class="`cat-${rule.category}`">
                      {{ rule.code }}
                    </span>
                    <h3 class="rule-heading">{{ rule.title }}</h3>
                    <span
                      v-if="rule.severity === 'critical'"
                      class="critical-tag"
                    >
                      Критическое
                    </span>
                  </div>

                  <div class="rule-actions-meta">
                    <span class="rule-date-label">
                      <Clock class="icon-xs" />
                      ред. {{ rule.lastUpdated }}
                    </span>
                    <button
                      type="button"
                      class="btn-copy-rule-link"
                      title="Скопировать прямую ссылку на это правило"
                      @click="copyRuleDirectLink(rule.code)"
                    >
                      <Link2 class="icon-xs" />
                      <span>{{ copiedRuleCode === rule.code ? 'Скопировано!' : 'Ссылка' }}</span>
                    </button>
                  </div>
                </div>

                <!-- Краткая выжимка -->
                <div class="rule-summary-banner">
                  <ShieldAlert class="icon-xs text-primary" />
                  <span>{{ rule.summary }}</span>
                </div>

                <!-- Детальное описание -->
                <div class="rule-description-body">
                  <p>{{ rule.description }}</p>
                </div>

                <!-- Рекомендация по исправлению -->
                <div class="rule-fix-guideline">
                  <div class="fix-guideline-header">
                    <CheckCircle2 class="icon-xs text-success" />
                    <strong>Рекомендация по устранению замечания:</strong>
                  </div>
                  <p class="fix-guideline-text">{{ rule.howToFix }}</p>
                </div>
              </div>
            </div>
          </div>

          <!-- ─── 2. СТАНДАРТНАЯ ВЛОЖЕННАЯ СТАТЬЯ ДОКУМЕНТАЦИИ ─────────── -->
          <div v-else-if="isInsideNestedArticle" class="standard-article-view">
            <div class="article-title-block">
              <h1 class="article-h1">{{ currentArticle?.title }}</h1>
              <p class="article-lead">{{ currentArticle?.summary }}</p>
            </div>

            <div class="article-content-wrapper">
              <!-- Текстовое наполнение статьи -->
              <div
                v-if="currentArticle?.content"
                class="article-text-body"
                v-html="renderMarkdown(currentArticle.content)"
              ></div>

              <!-- Одиночный сниппет кода -->
              <div v-if="currentArticle?.codeSnippet" class="code-block-container">
                <div class="code-block-header">
                  <span class="code-language-tag">{{ currentArticle.codeLanguage || 'bash' }}</span>
                  <button
                    type="button"
                    class="btn-copy-snippet"
                    @click="copySnippet(currentArticle.codeSnippet, 'main')"
                  >
                    <Check v-if="copiedSnippetId === 'main'" class="icon-xs text-success" />
                    <Copy v-else class="icon-xs" />
                    <span>{{ copiedSnippetId === 'main' ? 'Скопировано' : 'Копировать' }}</span>
                  </button>
                </div>
                <pre class="code-block-pre"><code>{{ currentArticle.codeSnippet }}</code></pre>
              </div>

              <!-- Вкладки языков (C#, C++, Go и т.д.) -->
              <div
                v-if="currentArticle?.codeTabs && currentArticle.codeTabs.length"
                class="code-tabs-container"
              >
                <div class="code-tabs-header">
                  <div class="tabs-buttons-list">
                    <button
                      v-for="(tab, tIdx) in currentArticle.codeTabs"
                      :key="tIdx"
                      type="button"
                      class="lang-tab-btn"
                      :class="{ active: (activeTabIndices[currentArticle.id] ?? 0) === tIdx }"
                      @click="activeTabIndices[currentArticle.id] = tIdx"
                    >
                      {{ tab.label }}
                    </button>
                  </div>
                  <button
                    type="button"
                    class="btn-copy-snippet"
                    @click="copySnippet(currentArticle.codeTabs[activeTabIndices[currentArticle.id] ?? 0]?.code, 'tab')"
                  >
                    <Check v-if="copiedSnippetId === 'tab'" class="icon-xs text-success" />
                    <Copy v-else class="icon-xs" />
                    <span>{{ copiedSnippetId === 'tab' ? 'Скопировано' : 'Копировать' }}</span>
                  </button>
                </div>
                <pre class="code-block-pre"><code>{{ currentArticle.codeTabs[activeTabIndices[currentArticle.id] ?? 0]?.code }}</code></pre>
              </div>
            </div>

            <!-- Нижняя навигация: Предыдущая / Следующая статья -->
            <div class="article-pagination-footer">
              <button
                v-if="prevArticle"
                type="button"
                class="pagination-btn prev-btn"
                @click="navigateToArticle(currentSectionId, prevArticle.id)"
              >
                <span class="pag-dir">← Назад</span>
                <span class="pag-title">{{ prevArticle.title }}</span>
              </button>
              <div v-else class="pag-spacer"></div>

              <button
                v-if="nextArticle"
                type="button"
                class="pagination-btn next-btn"
                @click="navigateToArticle(currentSectionId, nextArticle.id)"
              >
                <span class="pag-dir">Вперед →</span>
                <span class="pag-title">{{ nextArticle.title }}</span>
              </button>
            </div>
          </div>

          <!-- ─── 3. ОБЗОР КОРНЕВОГО РАЗДЕЛА (С КАРТОЧКАМИ СТАТЕЙ) ─────── -->
          <div v-else class="section-overview-grid">
            <div class="overview-cards-list">
              <div
                v-for="art in currentSection?.articles"
                :key="art.id"
                class="overview-article-card"
                @click="navigateToArticle(currentSectionId, art.id)"
              >
                <div class="card-icon-wrap">
                  <FileText class="icon-sm text-primary" />
                </div>
                <div class="card-text-wrap">
                  <h3 class="card-article-title">{{ art.title }}</h3>
                  <p class="card-article-summary">{{ art.summary }}</p>
                </div>
                <ChevronRight class="icon-sm text-muted card-arrow" />
              </div>
            </div>
          </div>
        </div>
      </main>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch, nextTick } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import {
  BookOpen,
  Sparkles,
  Server,
  Code2,
  ShieldCheck,
  Search,
  X,
  ChevronDown,
  ChevronRight,
  Clock,
  Link2,
  ShieldAlert,
  CheckCircle2,
  AlertCircle,
  Copy,
  Check,
  FileText,
} from 'lucide-vue-next';
import {
  DOCS_SECTIONS,
  PLATFORM_RULES,
  getRulesByCategory,
  getSectionById,
  getArticleById,
  formatDocDate,
} from '@/entities/documentation';
import { showToast } from '@/shared/lib';

const route = useRoute();
const router = useRouter();

const currentSectionId = ref('getting-started');
const currentArticleId = ref('');
const expandedSections = ref({
  'getting-started': true,
  orchestration: true,
  sdk: true,
  rules: true,
});

const sidebarSearch = ref('');
const rulesSearchQuery = ref('');
const selectedRuleCategory = ref('all');
const highlightedRuleCode = ref('');
const copiedRuleCode = ref('');
const copiedSnippetId = ref('');
const activeTabIndices = ref({});
const contentScrollRef = ref(null);

const ruleCategories = computed(() => getRulesByCategory());

const currentSection = computed(() => {
  return getSectionById(currentSectionId.value) || DOCS_SECTIONS[0];
});

const currentArticle = computed(() => {
  if (!currentArticleId.value) return null;
  return getArticleById(currentSectionId.value, currentArticleId.value);
});

const isInsideNestedArticle = computed(() => {
  return !!currentArticle.value;
});

const isRulesCatalogView = computed(() => {
  if (currentSectionId.value === 'rules') {
    if (!currentArticleId.value || currentArticleId.value === 'rules-catalog') {
      return true;
    }
  }
  return false;
});

const formattedSectionDate = computed(() => {
  return formatDocDate(currentSection.value?.lastUpdated);
});

const currentArticleLastUpdated = computed(() => {
  return currentArticle.value?.lastUpdated || currentSection.value?.lastUpdated;
});

const formattedArticleDate = computed(() => {
  return formatDocDate(currentArticleLastUpdated.value);
});

const filteredSections = computed(() => {
  if (!sidebarSearch.value.trim()) return DOCS_SECTIONS;
  const q = sidebarSearch.value.trim().toLowerCase();
  return DOCS_SECTIONS.filter(
    (s) =>
      s.title.toLowerCase().includes(q) ||
      s.description.toLowerCase().includes(q) ||
      s.articles?.some((a) => a.title.toLowerCase().includes(q) || a.summary.toLowerCase().includes(q))
  );
});

const displayedRules = computed(() => {
  let list = PLATFORM_RULES;
  if (selectedRuleCategory.value !== 'all') {
    list = list.filter((r) => r.category === selectedRuleCategory.value);
  }
  if (rulesSearchQuery.value.trim()) {
    const q = rulesSearchQuery.value.trim().toLowerCase();
    list = list.filter(
      (r) =>
        r.code.toLowerCase().includes(q) ||
        r.title.toLowerCase().includes(q) ||
        r.summary.toLowerCase().includes(q) ||
        r.description.toLowerCase().includes(q) ||
        r.categoryTitle.toLowerCase().includes(q)
    );
  }
  return list;
});

const articleNavList = computed(() => {
  return currentSection.value?.articles || [];
});

const currentArticleIndex = computed(() => {
  if (!currentArticleId.value) return -1;
  return articleNavList.value.findIndex((a) => a.id === currentArticleId.value);
});

const prevArticle = computed(() => {
  const idx = currentArticleIndex.value;
  if (idx > 0) return articleNavList.value[idx - 1];
  return null;
});

const nextArticle = computed(() => {
  const idx = currentArticleIndex.value;
  if (idx >= 0 && idx < articleNavList.value.length - 1) {
    return articleNavList.value[idx + 1];
  }
  return null;
});

function isSectionExpanded(sectionId) {
  return expandedSections.value[sectionId] !== false;
}

function getSectionIcon(iconName) {
  switch (iconName) {
    case 'Sparkles':
      return Sparkles;
    case 'Server':
      return Server;
    case 'Code2':
      return Code2;
    case 'ShieldCheck':
      return ShieldCheck;
    default:
      return BookOpen;
  }
}

function handleRootSectionClick(section) {
  currentSectionId.value = section.id;
  expandedSections.value[section.id] = !expandedSections.value[section.id];
  // При клике на раздел переходим к первому материалу или разделу
  const firstArt = section.articles?.[0];
  if (firstArt) {
    navigateToArticle(section.id, firstArt.id);
  } else {
    navigateToSectionRoot(section.id);
  }
}

function navigateToSectionRoot(sectionId) {
  currentSectionId.value = sectionId;
  currentArticleId.value = '';
  router.push(`/docs/${sectionId}`);
  scrollToTop();
}

function navigateToArticle(sectionId, articleId) {
  currentSectionId.value = sectionId;
  currentArticleId.value = articleId;
  expandedSections.value[sectionId] = true;
  router.push(`/docs/${sectionId}/${articleId}`);
  scrollToTop();
}

function scrollToTop() {
  nextTick(() => {
    if (contentScrollRef.value) {
      contentScrollRef.value.scrollTop = 0;
    }
  });
}

function resetRuleFilters() {
  rulesSearchQuery.value = '';
  selectedRuleCategory.value = 'all';
}

function copyRuleDirectLink(code) {
  const url = `${window.location.origin}/docs/rules/rules-catalog#${code}`;
  navigator.clipboard.writeText(url);
  copiedRuleCode.value = code;
  showToast(`Ссылка на правило ${code} скопирована`, 'success');
  setTimeout(() => {
    if (copiedRuleCode.value === code) {
      copiedRuleCode.value = '';
    }
  }, 2000);
}

function copySnippet(text, id) {
  if (!text) return;
  navigator.clipboard.writeText(text);
  copiedSnippetId.value = id;
  showToast('Сниппет скопирован в буфер', 'success');
  setTimeout(() => {
    if (copiedSnippetId.value === id) {
      copiedSnippetId.value = '';
    }
  }, 2000);
}

function renderMarkdown(text) {
  if (!text) return '';
  return text
    .replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>')
    .replace(/`(.*?)`/g, '<code class="inline-code">$1</code>')
    .replace(/\n\n/g, '</p><p>')
    .replace(/\n- /g, '<br/>• ')
    .replace(/\n(\d+)\. /g, '<br/>$1. ');
}

function syncRouteState() {
  const sectionParam = route.params.section;
  const articleParam = route.params.article;

  if (sectionParam) {
    currentSectionId.value = sectionParam;
    expandedSections.value[sectionParam] = true;
  } else if (route.path.includes('/rules')) {
    currentSectionId.value = 'rules';
    expandedSections.value['rules'] = true;
  }

  if (articleParam) {
    currentArticleId.value = articleParam;
  } else if (currentSectionId.value === 'rules') {
    currentArticleId.value = 'rules-catalog';
  } else {
    // Если статья не указана в URL, выбираем первую статью раздела
    const sec = getSectionById(currentSectionId.value);
    if (sec && sec.articles?.length) {
      currentArticleId.value = sec.articles[0].id;
    } else {
      currentArticleId.value = '';
    }
  }

  // Обработка якоря правила (напр. #SEC-04)
  if (route.hash) {
    const targetCode = route.hash.replace('#', '').toUpperCase();
    currentSectionId.value = 'rules';
    currentArticleId.value = 'rules-catalog';
    highlightedRuleCode.value = targetCode;

    nextTick(() => {
      setTimeout(() => {
        const el = document.getElementById(targetCode);
        if (el) {
          el.scrollIntoView({ behavior: 'smooth', block: 'center' });
        }
      }, 250);
    });
  }
}

watch(
  () => [route.params.section, route.params.article, route.hash],
  () => {
    syncRouteState();
  }
);

onMounted(() => {
  syncRouteState();
});
</script>

<style scoped>
/* ─── ПОЛНОЭКРАННЫЙ КОНТЕЙНЕР БЕЗ ВНЕШНИХ ПОЛЕЙ ───────────────── */
.docs-viewport-container {
  width: 100%;
  height: calc(100vh - 60px);
  background: var(--bg-app, #0d1117);
  color: var(--text-main, #f0f6fc);
  overflow: hidden;
  box-sizing: border-box;
}

.docs-split-layout {
  display: flex;
  width: 100%;
  height: 100%;
}

/* ─── ЛЕВЫЙ САЙДБАР: ФИКСИРОВАН, ПРИЖАТ К ЛЕВОМУ КРАЮ ────────── */
.docs-sidebar-fixed {
  width: 290px;
  min-width: 290px;
  height: 100%;
  background: var(--bg-secondary, #161b22);
  border-right: 1px solid var(--border, #30363d);
  display: flex;
  flex-direction: column;
  padding: 16px 14px;
  box-sizing: border-box;
  overflow-y: auto;
  flex-shrink: 0;
}

.sidebar-top-brand {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 14px;
  padding: 0 4px;
}

.brand-badge-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.brand-title {
  font-size: 14px;
  font-weight: 700;
  color: var(--text-main, #f0f6fc);
  letter-spacing: -0.2px;
}

.sidebar-search-wrap {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 7px 10px;
  background: var(--bg-app, #0d1117);
  border: 1px solid var(--border, #30363d);
  border-radius: var(--radius-sm, 6px);
  margin-bottom: 16px;
}

.sidebar-search-input {
  flex: 1;
  background: transparent;
  border: none;
  color: var(--text-main, #f0f6fc);
  font-size: 12px;
  outline: none;
}

.btn-clear-search {
  background: transparent;
  border: none;
  color: var(--text-muted, #8b949e);
  cursor: pointer;
  padding: 0;
  display: flex;
  align-items: center;
}

.sidebar-nav-tree {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.nav-root-group {
  display: flex;
  flex-direction: column;
}

.root-section-btn {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 10px;
  border-radius: 6px;
  background: transparent;
  border: none;
  color: var(--text-muted, #8b949e);
  font-size: 13px;
  font-weight: 600;
  text-align: left;
  cursor: pointer;
  transition: all 0.15s ease;
  width: 100%;
}

.root-section-btn:hover {
  background: var(--bg-hover, #21262d);
  color: var(--text-main, #f0f6fc);
}

.root-section-btn.active {
  color: var(--primary, #58a6ff);
}

.section-icon {
  flex-shrink: 0;
}

.root-title {
  flex: 1;
}

.chevron-toggle {
  transition: transform 0.2s ease;
  opacity: 0.6;
}

.chevron-toggle.chevron-open {
  transform: rotate(180deg);
}

.nested-articles-list {
  display: flex;
  flex-direction: column;
  padding-left: 18px;
  margin: 2px 0 6px;
  gap: 1px;
}

.nested-art-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 10px;
  background: transparent;
  border: none;
  color: var(--text-muted, #8b949e);
  font-size: 12px;
  text-align: left;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.12s;
}

.nested-art-btn:hover {
  color: var(--text-main, #f0f6fc);
  background: rgba(255, 255, 255, 0.04);
}

.nested-art-btn.active {
  color: var(--primary, #58a6ff);
  background: rgba(88, 166, 255, 0.1);
  font-weight: 600;
}

.art-dot {
  width: 4px;
  height: 4px;
  border-radius: 50%;
  background: var(--text-muted, #8b949e);
  opacity: 0.5;
}

.nested-art-btn.active .art-dot {
  background: var(--primary, #58a6ff);
  opacity: 1;
}

/* ─── ЦЕНТРАЛЬНЫЙ СКРОЛЛЯЩИЙСЯ КОНТЕНТ ────────────────────────── */
.docs-main-scrollable {
  flex: 1;
  height: 100%;
  overflow-y: auto;
  padding: 32px 48px 80px;
  box-sizing: border-box;
  min-width: 0;
}

.content-inner-wrapper {
  max-width: 980px;
  margin: 0 auto;
}

/* ─── МИКРОНАВИГАЦИЯ (ХЛЕБНЫЕ КРОШКИ) ─────────────────────────── */
.article-breadcrumb-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 24px;
  padding-bottom: 14px;
  border-bottom: 1px solid var(--border, #30363d);
}

.breadcrumbs-trail {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
}

.crumb-root-link {
  background: transparent;
  border: none;
  color: var(--text-muted, #8b949e);
  padding: 0;
  cursor: pointer;
  font-size: 13px;
  transition: color 0.15s;
}

.crumb-root-link:hover {
  color: var(--primary, #58a6ff);
}

.crumb-leaf {
  color: var(--text-main, #f0f6fc);
  font-weight: 600;
}

.last-updated-badge {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: var(--text-muted, #8b949e);
  background: var(--bg-secondary, #161b22);
  padding: 4px 10px;
  border-radius: 14px;
  border: 1px solid var(--border, #30363d);
}

/* ─── ЗАГОЛОВОК СТАТЬИ ────────────────────────────────────────── */
.article-title-block {
  margin-bottom: 28px;
}

.article-h1 {
  font-size: 26px;
  font-weight: 800;
  letter-spacing: -0.4px;
  margin: 0 0 8px;
  color: var(--text-main, #f0f6fc);
}

.article-lead {
  font-size: 14px;
  color: var(--text-muted, #8b949e);
  margin: 0;
  line-height: 1.5;
}

.article-text-body {
  font-size: 14px;
  line-height: 1.7;
  color: #c9d1d9;
  margin-bottom: 24px;
}

:deep(.inline-code) {
  font-family: monospace;
  font-size: 12px;
  background: rgba(255, 255, 255, 0.08);
  padding: 2px 5px;
  border-radius: 4px;
  color: #ff7b72;
}

/* ─── ШАПКА КОРНЕВОГО РАЗДЕЛА ─────────────────────────────────── */
.section-overview-hero {
  margin-bottom: 32px;
}

.hero-meta-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}

.hero-section-kicker {
  font-size: 11px;
  font-weight: 700;
  text-transform: uppercase;
  color: var(--primary, #58a6ff);
  letter-spacing: 0.5px;
}

.hero-main-title {
  font-size: 28px;
  font-weight: 800;
  margin: 0 0 8px;
  color: var(--text-main, #f0f6fc);
}

.hero-description {
  font-size: 15px;
  color: var(--text-muted, #8b949e);
  margin: 0;
  line-height: 1.5;
}

.overview-cards-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.overview-article-card {
  display: flex;
  align-items: center;
  gap: 16px;
  background: var(--bg-secondary, #161b22);
  border: 1px solid var(--border, #30363d);
  border-radius: var(--radius-md, 8px);
  padding: 16px 20px;
  cursor: pointer;
  transition: all 0.15s ease;
}

.overview-article-card:hover {
  border-color: var(--primary, #58a6ff);
  background: var(--bg-hover, #21262d);
  transform: translateY(-1px);
}

.card-icon-wrap {
  width: 36px;
  height: 36px;
  border-radius: 8px;
  background: rgba(88, 166, 255, 0.1);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.card-text-wrap {
  flex: 1;
}

.card-article-title {
  font-size: 15px;
  font-weight: 700;
  margin: 0 0 4px;
  color: var(--text-main, #f0f6fc);
}

.card-article-summary {
  font-size: 12px;
  color: var(--text-muted, #8b949e);
  margin: 0;
  line-height: 1.4;
}

.card-arrow {
  opacity: 0.5;
  transition: transform 0.15s;
}

.overview-article-card:hover .card-arrow {
  opacity: 1;
  transform: translateX(3px);
  color: var(--primary, #58a6ff);
}

/* ─── КАТАЛОГ ПРАВИЛ МОДЕРАЦИИ ───────────────────────────────── */
.rules-filter-toolbar {
  background: var(--bg-secondary, #161b22);
  border: 1px solid var(--border, #30363d);
  border-radius: var(--radius-md, 8px);
  padding: 16px;
  margin-bottom: 24px;
}

.rules-search-box {
  display: flex;
  align-items: center;
  gap: 10px;
  background: var(--bg-app, #0d1117);
  border: 1px solid var(--border, #30363d);
  padding: 9px 12px;
  border-radius: 6px;
  margin-bottom: 12px;
}

.rules-search-input {
  flex: 1;
  background: transparent;
  border: none;
  color: var(--text-main, #f0f6fc);
  font-size: 13px;
  outline: none;
}

.rules-categories-row {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.cat-chip-btn {
  background: transparent;
  border: 1px solid var(--border, #30363d);
  color: var(--text-muted, #8b949e);
  padding: 5px 12px;
  border-radius: 16px;
  font-size: 12px;
  cursor: pointer;
  transition: all 0.15s;
}

.cat-chip-btn:hover {
  background: var(--bg-hover, #21262d);
  color: var(--text-main, #f0f6fc);
}

.cat-chip-btn.active {
  background: rgba(88, 166, 255, 0.15);
  border-color: var(--primary, #58a6ff);
  color: var(--primary, #58a6ff);
  font-weight: 600;
}

.rules-cards-stack {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.no-rules-found {
  padding: 40px;
  text-align: center;
  background: var(--bg-secondary, #161b22);
  border: 1px dashed var(--border, #30363d);
  border-radius: 8px;
  color: var(--text-muted, #8b949e);
}

.btn-reset-search {
  margin-top: 12px;
  background: var(--primary, #58a6ff);
  border: none;
  color: #0d1117;
  padding: 6px 14px;
  font-weight: 600;
  border-radius: 6px;
  cursor: pointer;
}

.rule-box-card {
  background: var(--bg-secondary, #161b22);
  border: 1px solid var(--border, #30363d);
  border-radius: var(--radius-md, 8px);
  padding: 18px 22px;
  transition: all 0.25s ease;
  scroll-margin-top: 24px;
}

.rule-box-card:hover {
  border-color: #484f58;
}

.rule-box-card.is-target-highlighted {
  border-color: var(--primary, #58a6ff);
  background: rgba(88, 166, 255, 0.05);
  box-shadow: 0 0 24px rgba(88, 166, 255, 0.25);
  animation: pulseGlow 2s ease;
}

@keyframes pulseGlow {
  0% {
    transform: scale(0.99);
  }
  50% {
    transform: scale(1.005);
    box-shadow: 0 0 30px rgba(88, 166, 255, 0.4);
  }
  100% {
    transform: scale(1);
  }
}

.rule-box-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
  flex-wrap: wrap;
  gap: 8px;
}

.rule-title-wrap {
  display: flex;
  align-items: center;
  gap: 10px;
}

.rule-code-chip {
  font-family: monospace;
  font-weight: 700;
  font-size: 13px;
  padding: 2px 7px;
  border-radius: 4px;
}

.rule-code-chip.cat-security {
  color: #58a6ff;
  background: rgba(88, 166, 255, 0.15);
  border: 1px solid rgba(88, 166, 255, 0.3);
}

.rule-code-chip.cat-servers {
  color: #3fb950;
  background: rgba(63, 185, 80, 0.15);
  border: 1px solid rgba(63, 185, 80, 0.3);
}

.rule-code-chip.cat-builds {
  color: #d29922;
  background: rgba(210, 153, 34, 0.15);
  border: 1px solid rgba(210, 153, 34, 0.3);
}

.rule-code-chip.cat-content {
  color: #bc8cff;
  background: rgba(188, 140, 255, 0.15);
  border: 1px solid rgba(188, 140, 255, 0.3);
}

.rule-heading {
  font-size: 15px;
  font-weight: 700;
  color: var(--text-main, #f0f6fc);
  margin: 0;
}

.critical-tag {
  font-size: 10px;
  font-weight: 700;
  padding: 1px 6px;
  border-radius: 4px;
  background: rgba(248, 81, 73, 0.15);
  color: var(--danger, #f85149);
  border: 1px solid rgba(248, 81, 73, 0.3);
}

.rule-actions-meta {
  display: flex;
  align-items: center;
  gap: 10px;
}

.rule-date-label {
  font-size: 11px;
  color: #6e7681;
  display: flex;
  align-items: center;
  gap: 4px;
}

.btn-copy-rule-link {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  background: transparent;
  border: 1px solid var(--border, #30363d);
  color: var(--text-muted, #8b949e);
  font-size: 11px;
  padding: 2px 7px;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.15s;
}

.btn-copy-rule-link:hover {
  color: var(--primary, #58a6ff);
  border-color: var(--primary, #58a6ff);
  background: var(--bg-hover, #21262d);
}

.rule-summary-banner {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  background: rgba(88, 166, 255, 0.08);
  border-left: 3px solid var(--primary, #58a6ff);
  border-radius: 4px;
  margin-bottom: 12px;
  font-size: 13px;
  color: #c9d1d9;
}

.rule-description-body {
  font-size: 13px;
  line-height: 1.6;
  color: #b0b8c4;
  margin-bottom: 14px;
}

.rule-fix-guideline {
  background: rgba(63, 185, 80, 0.08);
  border: 1px solid rgba(63, 185, 80, 0.2);
  border-radius: 6px;
  padding: 10px 14px;
}

.fix-guideline-header {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: #3fb950;
  margin-bottom: 4px;
}

.fix-guideline-text {
  font-size: 12px;
  color: #c9d1d9;
  line-height: 1.5;
  margin: 0;
}

/* ─── СНИППЕТЫ КОДА И ВКЛАДКИ ────────────────────────────────── */
.code-block-container,
.code-tabs-container {
  background: #090d12;
  border: 1px solid var(--border, #30363d);
  border-radius: var(--radius-sm, 6px);
  overflow: hidden;
  margin: 16px 0 24px;
}

.code-block-header,
.code-tabs-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 12px;
  background: var(--bg-secondary, #161b22);
  border-bottom: 1px solid var(--border, #30363d);
}

.code-language-tag {
  font-family: monospace;
  font-size: 11px;
  font-weight: 600;
  color: var(--text-muted, #8b949e);
  text-transform: uppercase;
}

.tabs-buttons-list {
  display: flex;
  gap: 4px;
}

.lang-tab-btn {
  background: transparent;
  border: none;
  color: var(--text-muted, #8b949e);
  font-size: 12px;
  font-weight: 500;
  padding: 4px 10px;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.12s;
}

.lang-tab-btn:hover {
  color: var(--text-main, #f0f6fc);
}

.lang-tab-btn.active {
  background: rgba(88, 166, 255, 0.15);
  color: var(--primary, #58a6ff);
  font-weight: 600;
}

.btn-copy-snippet {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  background: transparent;
  border: 1px solid var(--border, #30363d);
  color: var(--text-muted, #8b949e);
  font-size: 11px;
  padding: 3px 8px;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.15s;
}

.btn-copy-snippet:hover {
  color: var(--text-main, #f0f6fc);
  background: var(--bg-hover, #21262d);
}

.code-block-pre {
  margin: 0;
  padding: 16px;
  overflow-x: auto;
  font-family: 'JetBrains Mono', 'Fira Code', monospace;
  font-size: 12px;
  line-height: 1.5;
  color: #e6edf3;
}

/* ─── НАВИГАЦИЯ ПО СТАТЬЯМ (ПАГИНАЦИЯ) ────────────────────────── */
.article-pagination-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 48px;
  padding-top: 24px;
  border-top: 1px solid var(--border, #30363d);
}

.pagination-btn {
  display: flex;
  flex-direction: column;
  background: var(--bg-secondary, #161b22);
  border: 1px solid var(--border, #30363d);
  border-radius: 8px;
  padding: 12px 18px;
  cursor: pointer;
  transition: all 0.15s;
  min-width: 180px;
}

.pagination-btn.prev-btn {
  align-items: flex-start;
  text-align: left;
}

.pagination-btn.next-btn {
  align-items: flex-end;
  text-align: right;
}

.pagination-btn:hover {
  border-color: var(--primary, #58a6ff);
  background: var(--bg-hover, #21262d);
}

.pag-dir {
  font-size: 11px;
  color: var(--text-muted, #8b949e);
  margin-bottom: 4px;
}

.pag-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-main, #f0f6fc);
}

.pag-spacer {
  flex: 1;
}
</style>
