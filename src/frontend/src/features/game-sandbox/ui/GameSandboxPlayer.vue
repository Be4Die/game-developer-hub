<template>
  <div ref="playerContainer" class="game-sandbox-container" :class="{ 'is-fullscreen': isFullscreen }">
    <!-- ВЕРХНИЙ ТУЛБАР УПРАВЛЕНИЯ -->
    <div class="sandbox-toolbar">
      <div class="toolbar-left">
        <div class="status-indicator">
          <span class="status-dot"></span>
          <span class="status-text">Песочница WelwiseGames</span>
        </div>

        <!-- Переключатель видового экрана (Эмулятор устройств) -->
        <div class="viewport-presets">
          <button
            class="preset-btn"
            :class="{ active: currentViewport === 'fit' }"
            title="Заполнить экран"
            @click="setViewport('fit')"
          >
            <Maximize2 class="icon-xs" />
            <span>Адаптивный</span>
          </button>
          <button
            class="preset-btn"
            :class="{ active: currentViewport === 'desktop' }"
            title="Десктоп 16:9 (1280x720)"
            @click="setViewport('desktop')"
          >
            <Monitor class="icon-xs" />
            <span>Десктоп 16:9</span>
          </button>
          <button
            class="preset-btn"
            :class="{ active: currentViewport === 'mobile-portrait' }"
            title="Мобильный экран 9:16 (375x667)"
            @click="setViewport('mobile-portrait')"
          >
            <Smartphone class="icon-xs" />
            <span>Мобильный (9:16)</span>
          </button>
          <button
            class="preset-btn"
            :class="{ active: currentViewport === 'mobile-landscape' }"
            title="Мобильный альбомный (667x375)"
            @click="setViewport('mobile-landscape')"
          >
            <Tablet class="icon-xs" />
            <span>Альбомный (16:9)</span>
          </button>
        </div>
      </div>

      <div class="toolbar-right">
        <button class="tool-btn" title="Перезагрузить игру" @click="reloadIframe">
          <RefreshCw class="icon-xs" :class="{ spin: isReloading }" />
          <span>Перезапустить</span>
        </button>

        <button class="tool-btn" title="Полноэкранный режим" @click="toggleFullscreen">
          <Minimize2 v-if="isFullscreen" class="icon-xs" />
          <Maximize v-else class="icon-xs" />
          <span>{{ isFullscreen ? 'Свернуть' : 'Во весь экран' }}</span>
        </button>

        <button
          v-if="showDevtools"
          class="tool-btn"
          :class="{ active: isDevtoolsOpen }"
          title="Открыть панель SDK DevTools"
          @click="isDevtoolsOpen = !isDevtoolsOpen"
        >
          <Terminal class="icon-xs" />
          <span>DevTools</span>
          <span v-if="logs.length" class="badge-pill">{{ logs.length }}</span>
        </button>
      </div>
    </div>

    <!-- ОСНОВНАЯ РАБОЧАЯ ОБЛАСТЬ -->
    <div class="sandbox-main">
      <!-- ИГРОВАЯ СЦЕНА С ФРЕЙМОМ -->
      <div class="game-viewport-stage" :class="`viewport-${currentViewport}`">
        <div class="frame-wrapper" :style="viewportStyle">
          <!-- Само игровое окно -->
          <iframe
            v-if="gameUrl"
            :key="iframeKey"
            ref="gameIframe"
            :src="gameUrl"
            sandbox="allow-scripts allow-forms allow-pointer-lock allow-same-origin"
            allow="autoplay; fullscreen; gamepad"
            class="game-iframe"
          ></iframe>

          <div v-else class="empty-url-notice">
            <Gamepad2 class="icon-lg text-muted" />
            <p>Билд игры не найден или не развёрнут</p>
          </div>

          <!-- ОВЕРЛЕЙ ЭМУЛЯЦИИ РЕКЛАМЫ (Ad Simulator Overlay) -->
          <transition name="fade">
            <div v-if="adState.isOpen" class="ad-emulator-overlay">
              <div class="ad-card">
                <div class="ad-badge">
                  {{ adState.type === 'rewarded' ? 'Тестовая реклама с вознаграждением (Rewarded)' : 'Тестовая реклама (Interstitial)' }}
                </div>
                <div class="ad-timer-circle">
                  <span class="ad-timer-num">{{ adState.countdown }}</span>
                </div>
                <p class="ad-desc">
                  Имитация показа рекламного ролика. Игра должна поставить звук на паузу и остановить игровой цикл.
                </p>
                <div class="ad-actions">
                  <button class="btn-skip-ad" @click="skipAd">
                    Пропустить рекламу
                  </button>
                  <button class="btn-fail-ad" @click="failAd">
                    Симулировать ошибку (AdBlock)
                  </button>
                </div>
              </div>
            </div>
          </transition>

          <!-- МОДАЛЬНОЕ ОКНО ЭМУЛЯЦИИ ПЛАТЕЖА (Purchase Simulator Modal) -->
          <transition name="fade">
            <div v-if="purchaseModal.isOpen" class="purchase-emulator-overlay">
              <div class="purchase-card">
                <div class="purchase-head">
                  <Coins class="icon-md text-warning" />
                  <h3>Тестовая внутриигровая покупка</h3>
                </div>
                <div class="purchase-body">
                  <div class="purchase-item-name">
                    {{ purchaseModal.itemDetails?.name || purchaseModal.itemId }}
                  </div>
                  <p class="purchase-item-desc">
                    {{ purchaseModal.itemDetails?.description || 'Тестовый предмет из каталога игры' }}
                  </p>
                  <div class="purchase-item-price">
                    Стоимость: <strong>{{ (purchaseModal.itemDetails?.priceCoins || 10) * purchaseModal.quantity }} монет</strong>
                    <span v-if="purchaseModal.quantity > 1">({{ purchaseModal.quantity }} шт.)</span>
                  </div>
                </div>
                <div class="purchase-actions">
                  <button class="btn-purchase-success" @click="handlePurchaseDecision(true)">
                    Успешно оплатить
                  </button>
                  <button class="btn-purchase-cancel" @click="handlePurchaseDecision(false)">
                    Отклонить покупку
                  </button>
                </div>
              </div>
            </div>
          </transition>
        </div>
      </div>

      <!-- БОКОВАЯ ПАНЕЛЬ DEVTOOLS -->
      <div v-if="showDevtools && isDevtoolsOpen" class="devtools-sidebar">
        <GameSandboxDevtools
          :logs="logs"
          :app-env="appEnv"
          :player-storage="playerStorage"
          :ad-config="adConfig"
          :purchase-catalog="purchaseCatalog"
          @clear-logs="clearLogs"
          @clear-storage="clearStorage"
          @update-env="Object.assign(appEnv, $event)"
          @update-ad-config="Object.assign(adConfig, $event)"
        />
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue';
import {
  Maximize2,
  Minimize2,
  Maximize,
  Monitor,
  Smartphone,
  Tablet,
  RefreshCw,
  Terminal,
  Gamepad2,
  Coins,
} from 'lucide-vue-next';
import { useGameSdkSandbox } from '../model/useGameSdkSandbox';
import GameSandboxDevtools from './GameSandboxDevtools.vue';

const props = defineProps({
  gameUrl: {
    type: String,
    required: true,
  },
  projectId: {
    type: [String, Number],
    required: true,
  },
  showDevtools: {
    type: Boolean,
    default: true,
  },
});

const gameIframe = ref(null);
const playerContainer = ref(null);
const iframeKey = ref(0);
const isFullscreen = ref(false);
const isReloading = ref(false);
const isDevtoolsOpen = ref(true);
const currentViewport = ref('fit'); // 'fit' | 'desktop' | 'mobile-portrait' | 'mobile-landscape'

// Подключаем хост-адаптер Sandbox SDK
const {
  appEnv,
  logs,
  playerStorage,
  adConfig,
  adState,
  purchaseCatalog,
  purchaseModal,
  clearStorage,
  clearLogs,
  skipAd,
  failAd,
  handlePurchaseDecision,
} = useGameSdkSandbox({
  projectId: props.projectId,
  iframeRef: gameIframe,
});

function setViewport(preset) {
  currentViewport.value = preset;
}

const viewportStyle = computed(() => {
  switch (currentViewport.value) {
    case 'desktop':
      return { width: '1280px', height: '720px', maxWidth: '100%', maxHeight: '100%' };
    case 'mobile-portrait':
      return { width: '375px', height: '667px' };
    case 'mobile-landscape':
      return { width: '667px', height: '375px' };
    case 'fit':
    default:
      return { width: '100%', height: '100%' };
  }
});

function reloadIframe() {
  isReloading.value = true;
  iframeKey.value += 1;
  setTimeout(() => {
    isReloading.value = false;
  }, 400);
}

function toggleFullscreen() {
  if (!playerContainer.value) return;

  if (!document.fullscreenElement) {
    playerContainer.value.requestFullscreen().then(() => {
      isFullscreen.value = true;
    }).catch(() => {});
  } else {
    document.exitFullscreen().then(() => {
      isFullscreen.value = false;
    }).catch(() => {});
  }
}

document.addEventListener('fullscreenchange', () => {
  isFullscreen.value = !!document.fullscreenElement;
});
</script>

<style scoped>
.game-sandbox-container {
  display: flex;
  flex-direction: column;
  height: 100%;
  min-height: 580px;
  background: var(--bg-surface, #121316);
  border-radius: 8px;
  overflow: hidden;
  border: 1px solid var(--border-color, #2d3139);
}

.game-sandbox-container.is-fullscreen {
  border-radius: 0;
  border: none;
  width: 100vw;
  height: 100vh;
}

/* Верхний тулбар */
.sandbox-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 14px;
  background: var(--bg-card, #1a1c22);
  border-bottom: 1px solid var(--border-color, #2d3139);
  gap: 12px;
  flex-wrap: wrap;
}

.toolbar-left,
.toolbar-right {
  display: flex;
  align-items: center;
  gap: 8px;
}

.status-indicator {
  display: flex;
  align-items: center;
  gap: 6px;
  padding-right: 12px;
  border-right: 1px solid var(--border-color, #2d3139);
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #10b981;
  box-shadow: 0 0 6px rgba(16, 185, 129, 0.6);
}

.status-text {
  font-size: 12px;
  font-weight: 600;
  color: var(--text-main, #fff);
}

.viewport-presets {
  display: flex;
  gap: 4px;
  background: var(--bg-surface, #131418);
  padding: 2px;
  border-radius: 6px;
  border: 1px solid var(--border-color, #2d3139);
}

.preset-btn {
  display: flex;
  align-items: center;
  gap: 5px;
  padding: 4px 8px;
  background: transparent;
  border: none;
  border-radius: 4px;
  color: var(--text-muted, #9ba1ad);
  font-size: 11px;
  cursor: pointer;
  transition: all 0.15s ease;
}

.preset-btn:hover {
  color: var(--text-main, #fff);
}

.preset-btn.active {
  background: var(--primary, #3b82f6);
  color: #fff;
  font-weight: 600;
}

.tool-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 5px 10px;
  background: var(--bg-surface, #131418);
  border: 1px solid var(--border-color, #2d3139);
  color: var(--text-muted, #9ba1ad);
  border-radius: 6px;
  font-size: 12px;
  cursor: pointer;
  transition: all 0.15s ease;
}

.tool-btn:hover {
  color: var(--text-main, #fff);
  border-color: var(--border-color-hover, #424754);
}

.tool-btn.active {
  background: rgba(59, 130, 246, 0.15);
  border-color: var(--primary, #3b82f6);
  color: var(--primary, #3b82f6);
}

.badge-pill {
  padding: 1px 6px;
  background: var(--primary, #3b82f6);
  color: #fff;
  border-radius: 10px;
  font-size: 10px;
  font-weight: 600;
}

/* Основная область */
.sandbox-main {
  display: flex;
  flex: 1;
  position: relative;
  overflow: hidden;
}

.game-viewport-stage {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #0d0e11;
  padding: 16px;
  overflow: auto;
  position: relative;
}

.frame-wrapper {
  position: relative;
  background: #000;
  border-radius: 6px;
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.5);
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: width 0.2s ease, height 0.2s ease;
}

.game-iframe {
  width: 100%;
  height: 100%;
  border: none;
  display: block;
}

.empty-url-notice {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10px;
  color: var(--text-muted, #9ba1ad);
}

.devtools-sidebar {
  width: 360px;
  max-width: 40%;
  min-width: 280px;
  height: 100%;
  z-index: 10;
}

/* Оверлей рекламы */
.ad-emulator-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.85);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 50;
  backdrop-filter: blur(4px);
}

.ad-card {
  background: var(--bg-card, #1c1e24);
  border: 1px solid var(--border-color, #2d3139);
  padding: 24px;
  border-radius: 12px;
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  max-width: 380px;
  gap: 14px;
  box-shadow: 0 16px 40px rgba(0, 0, 0, 0.7);
}

.ad-badge {
  font-size: 11px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  padding: 4px 10px;
  border-radius: 20px;
  background: rgba(245, 158, 11, 0.2);
  color: #fbbf24;
}

.ad-timer-circle {
  width: 60px;
  height: 60px;
  border-radius: 50%;
  border: 3px solid #3b82f6;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24px;
  font-weight: 700;
  color: #fff;
}

.ad-desc {
  font-size: 13px;
  color: var(--text-muted, #9ba1ad);
  line-height: 1.4;
  margin: 0;
}

.ad-actions {
  display: flex;
  flex-direction: column;
  gap: 8px;
  width: 100%;
}

.btn-skip-ad {
  padding: 8px 16px;
  background: var(--primary, #3b82f6);
  color: #fff;
  border: none;
  border-radius: 6px;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
}

.btn-fail-ad {
  padding: 6px 12px;
  background: transparent;
  color: #f87171;
  border: 1px solid rgba(239, 68, 68, 0.3);
  border-radius: 6px;
  font-size: 12px;
  cursor: pointer;
}

/* Оверлей платежей */
.purchase-emulator-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.85);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 55;
  backdrop-filter: blur(4px);
}

.purchase-card {
  background: var(--bg-card, #1c1e24);
  border: 1px solid var(--border-color, #2d3139);
  padding: 24px;
  border-radius: 12px;
  display: flex;
  flex-direction: column;
  max-width: 360px;
  width: 90%;
  gap: 16px;
  box-shadow: 0 16px 40px rgba(0, 0, 0, 0.7);
}

.purchase-head {
  display: flex;
  align-items: center;
  gap: 10px;
}

.purchase-head h3 {
  margin: 0;
  font-size: 15px;
  font-weight: 600;
  color: var(--text-main, #fff);
}

.purchase-body {
  background: var(--bg-surface, #131418);
  padding: 12px;
  border-radius: 8px;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.purchase-item-name {
  font-size: 14px;
  font-weight: 600;
  color: #fff;
}

.purchase-item-desc {
  font-size: 12px;
  color: var(--text-muted, #9ba1ad);
  margin: 0;
}

.purchase-item-price {
  font-size: 13px;
  color: var(--text-main, #f0f2f5);
  margin-top: 4px;
}

.purchase-actions {
  display: flex;
  gap: 8px;
}

.btn-purchase-success {
  flex: 1;
  padding: 8px 12px;
  background: #10b981;
  color: #fff;
  border: none;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
}

.btn-purchase-cancel {
  flex: 1;
  padding: 8px 12px;
  background: transparent;
  color: var(--text-muted, #9ba1ad);
  border: 1px solid var(--border-color, #2d3139);
  border-radius: 6px;
  font-size: 12px;
  cursor: pointer;
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
