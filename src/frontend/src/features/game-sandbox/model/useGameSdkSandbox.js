import { ref, onMounted, onUnmounted, watch } from 'vue';

/**
 * useGameSdkSandbox
 * Хост-адаптер для встроенного плеера тестирования веб-игр.
 * Полностью реализует протокол WelwiseGames JS-SDK (v1) через window.postMessage.
 */
export function useGameSdkSandbox(options = {}) {
  const {
    projectId = 'test-project',
    iframeRef = null,
    defaultPlayerId = 'dev-player-1',
    defaultLanguage = 'ru',
    defaultDeviceType = 'desktop',
  } = options;

  // Окружение игрока (AppEnvironment)
  const appEnv = ref({
    playerId: defaultPlayerId,
    deviceType: defaultDeviceType, // 'desktop' | 'mobile'
    language: defaultLanguage, // 'ru' | 'en'
  });

  // Лог вызовов и событий SDK для панели DevTools
  const logs = ref([]);
  const MAX_LOGS = 200;

  function addLog(entry) {
    const item = {
      id: `${Date.now()}-${Math.random().toString(36).substring(2, 7)}`,
      timestamp: new Date().toLocaleTimeString('ru-RU', {
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit',
        fractionalSecondDigits: 3,
      }),
      status: 'info',
      ...entry,
    };
    logs.value.unshift(item);
    if (logs.value.length > MAX_LOGS) {
      logs.value.pop();
    }
  }

  // Облачные сохранения (Cloud Saves / Player Data)
  const storageKey = `gdh_sandbox_save_${projectId}`;
  const playerStorage = ref({});

  function loadStorage() {
    try {
      const raw = localStorage.getItem(storageKey);
      playerStorage.value = raw ? JSON.parse(raw) : {};
    } catch {
      playerStorage.value = {};
    }
  }

  function saveStorage(data) {
    playerStorage.value = data;
    try {
      localStorage.setItem(storageKey, JSON.stringify(data));
    } catch (e) {
      addLog({
        direction: 'system',
        actionName: 'STORAGE_ERROR',
        status: 'error',
        payload: { error: e.message },
      });
    }
  }

  function clearStorage() {
    playerStorage.value = {};
    localStorage.removeItem(storageKey);
    addLog({
      direction: 'system',
      actionName: 'CLEAR_STORAGE',
      status: 'warn',
      payload: { message: 'Данные игрока сброшены' },
    });
  }

  loadStorage();

  // Настройки эмуляции рекламы (Ad Simulator)
  const adConfig = ref({
    autoCloseSeconds: 3,
    simulateError: false,
    rewardedGranted: true,
  });

  // Текущее состояние рекламного оверлея
  const adState = ref({
    isOpen: false,
    type: 'midgame', // 'midgame' | 'rewarded'
    countdown: 0,
    timerId: null,
  });

  // Эмулятор внутриигровых покупок (In-App Purchases)
  const purchaseCatalog = ref([
    {
      itemId: 'gold_pack_100',
      name: 'Мешок золота (100 монет)',
      priceCoins: 50,
      description: 'Базовый набор для прокачки',
    },
    {
      itemId: 'energy_boost',
      name: 'Эликсир выносливости',
      priceCoins: 25,
      description: 'Мгновенно восстанавливает 100% энергии',
    },
    {
      itemId: 'vip_pass_30d',
      name: 'VIP-билет на 30 дней',
      priceCoins: 300,
      description: 'Удваивает награды за уровни',
    },
  ]);

  const purchaseModal = ref({
    isOpen: false,
    itemId: '',
    quantity: 1,
    itemDetails: null,
    idempotencyKey: '',
    purchaseId: '',
  });

  /**
   * Отправка ответа в Iframe игры
   */
  function postResponse(messageId, actionName, payload, error = null) {
    if (!iframeRef?.value?.contentWindow) return;

    const response = {
      source: 'WebsiteSDK',
      actionName,
      messageId,
      payload,
    };
    if (error) {
      response.error = typeof error === 'string' ? error : error.message || 'Unknown error';
    }

    const raw = JSON.stringify(response);
    iframeRef.value.contentWindow.postMessage(raw, '*');

    addLog({
      direction: 'out',
      source: 'WebsiteSDK',
      actionName,
      messageId,
      status: error ? 'error' : 'success',
      payload: response,
    });
  }

  /**
   * Отправка push-события в канал (например, adv-manager, purchase-manager)
   */
  function postPush(channel, payload) {
    if (!iframeRef?.value?.contentWindow) return;

    const push = {
      source: 'WebsiteSDK',
      channel,
      payload,
    };

    const raw = JSON.stringify(push);
    iframeRef.value.contentWindow.postMessage(raw, '*');

    addLog({
      direction: 'out',
      source: 'WebsiteSDK',
      channel,
      status: payload?.data?.error ? 'error' : 'success',
      payload: push,
    });
  }

  /**
   * Запуск рекламы
   */
  function startAd(type) {
    if (adState.value.isOpen) return;

    adState.value.isOpen = true;
    adState.value.type = type;
    adState.value.countdown = adConfig.value.autoCloseSeconds;

    // Пуш об открытии рекламы
    const openAction = type === 'rewarded' ? 'rewarded-video-callback-open' : 'adv-callback-open';
    postPush('adv-manager', { action: openAction });

    // Если включена симуляция ошибки/AdBlock
    if (adConfig.value.simulateError) {
      const errAction = type === 'rewarded' ? 'rewarded-video-callback-error' : 'adv-callback-error';
      setTimeout(() => {
        postPush('adv-manager', {
          action: errAction,
          data: { error: { message: 'Симуляция ошибки показа рекламы (AdBlock)' } },
        });
        closeAd();
      }, 800);
      return;
    }

    // Автоматический обратный отсчет
    if (adConfig.value.autoCloseSeconds > 0) {
      adState.value.timerId = setInterval(() => {
        adState.value.countdown -= 1;
        if (adState.value.countdown <= 0) {
          completeAd();
        }
      }, 1000);
    }
  }

  function completeAd() {
    if (!adState.value.isOpen) return;

    if (adState.value.type === 'rewarded' && adConfig.value.rewardedGranted) {
      postPush('adv-manager', { action: 'rewarded-video-callback-rewarded' });
    }

    const closeAction = adState.value.type === 'rewarded'
      ? 'rewarded-video-callback-close'
      : 'adv-callback-close';

    postPush('adv-manager', { action: closeAction });
    closeAd();
  }

  function closeAd() {
    if (adState.value.timerId) {
      clearInterval(adState.value.timerId);
      adState.value.timerId = null;
    }
    adState.value.isOpen = false;
  }

  function skipAd() {
    completeAd();
  }

  function failAd() {
    const errAction = adState.value.type === 'rewarded'
      ? 'rewarded-video-callback-error'
      : 'adv-callback-error';

    postPush('adv-manager', {
      action: errAction,
      data: { error: { message: 'Показ рекламы прерван пользователем в DevTools' } },
    });
    closeAd();
  }

  /**
   * Подтверждение/отклонение покупки
   */
  function handlePurchaseDecision(success) {
    if (!purchaseModal.value.isOpen) return;

    const { itemId, quantity, purchaseId } = purchaseModal.value;
    purchaseModal.value.isOpen = false;

    if (success) {
      postPush('purchase-manager', {
        action: 'purchase-callback-success',
        data: {
          purchaseId,
          itemId,
          quantity,
          debitedCoins: (purchaseModal.value.itemDetails?.priceCoins || 10) * quantity,
        },
      });
    } else {
      postPush('purchase-manager', {
        action: 'purchase-callback-error',
        data: {
          purchaseId,
          reason: 'USER_CANCELLED',
          message: 'Пользователь отменил тестовую покупку',
        },
      });
    }
  }

  /**
   * Обработчик входящих сообщений
   */
  function handleMessage(event) {
    let data;
    try {
      data = typeof event.data === 'string' ? JSON.parse(event.data) : event.data;
    } catch {
      return;
    }

    if (!data || data.source !== 'WelwiseSDK') return;

    const { actionName, messageId, payload } = data;

    addLog({
      direction: 'in',
      source: 'WelwiseSDK',
      actionName,
      messageId,
      payload,
    });

    switch (actionName) {
      case 'GET_IFRAME_ORIGIN_SRC': {
        // Хендшейк
        postResponse(messageId, 'GET_IFRAME_ORIGIN_SRC', {
          origin: window.location.origin,
          features: {
            metaverse: false,
            AppEnvironment: {
              playerId: String(appEnv.value.playerId || 'dev-player-1'),
              deviceType: String(appEnv.value.deviceType || 'desktop'),
              language: String(appEnv.value.language || 'ru'),
            },
          },
        });
        break;
      }

      case 'GET_SERVER_TIME': {
        postResponse(messageId, 'GET_SERVER_TIME', {
          serverTime: new Date().toISOString(),
        });
        break;
      }

      case 'GET_PLAYER_DATA': {
        postResponse(messageId, 'GET_PLAYER_DATA', playerStorage.value || {});
        break;
      }

      case 'SET_PLAYER_DATA': {
        const newData = payload || {};
        saveStorage(newData);
        postResponse(messageId, 'SET_PLAYER_DATA', { success: true });
        break;
      }

      case 'ADV_SHOW_MIDGAME': {
        // Подтверждаем вызов метода игры
        postResponse(messageId, 'ADV_SHOW_MIDGAME', { success: true });
        startAd('midgame');
        break;
      }

      case 'ADV_SHOW_REWARDED': {
        // Подтверждаем вызов метода игры
        postResponse(messageId, 'ADV_SHOW_REWARDED', { success: true });
        startAd('rewarded');
        break;
      }

      case 'GET_AVAILABLE_ITEMS': {
        postResponse(messageId, 'GET_AVAILABLE_ITEMS', purchaseCatalog.value);
        break;
      }

      case 'GET_PENDING_PURCHASES': {
        postResponse(messageId, 'GET_PENDING_PURCHASES', []);
        break;
      }

      case 'PURCHASE': {
        const itemId = payload?.itemId;
        const quantity = payload?.quantity || 1;
        const purchaseId = `purchase_${Date.now()}_${Math.random().toString(36).substring(2, 6)}`;
        const itemDetails = purchaseCatalog.value.find((i) => i.itemId === itemId) || {
          itemId,
          name: `Товар "${itemId}"`,
          priceCoins: 10,
        };

        purchaseModal.value = {
          isOpen: true,
          itemId,
          quantity,
          itemDetails,
          idempotencyKey: payload?.idempotencyKey || '',
          purchaseId,
        };

        // Отвечаем на первоначальный request
        postResponse(messageId, 'PURCHASE', { error: null });
        break;
      }

      case 'PURCHASE_CONFIRM': {
        postResponse(messageId, 'PURCHASE_CONFIRM', { success: true });
        break;
      }

      case 'GO_TO_GAME': {
        addLog({
          direction: 'system',
          actionName: 'GO_TO_GAME',
          status: 'info',
          payload: { targetGameId: payload?.gameId },
        });
        postResponse(messageId, 'GO_TO_GAME', { success: true });
        break;
      }

      case 'GAME_READY': {
        addLog({
          direction: 'system',
          actionName: 'GAME_READY',
          status: 'success',
          payload: { message: 'Игра сообщила о готовности к рендеру' },
        });
        break;
      }

      case 'GET_METAVERSE_PLAYER_DATA':
      case 'GET_METAVERSE_GAME_PLAYER_DATA': {
        postResponse(messageId, actionName, {});
        break;
      }

      case 'SET_METAVERSE_PLAYER_DATA':
      case 'SET_METAVERSE_GAME_PLAYER_DATA': {
        postResponse(messageId, actionName, { success: true });
        break;
      }

      default: {
        addLog({
          direction: 'system',
          actionName: `UNKNOWN: ${actionName}`,
          status: 'warn',
          payload: data,
        });
        postResponse(messageId, actionName, { error: `Action ${actionName} is not handled` });
        break;
      }
    }
  }

  onMounted(() => {
    window.addEventListener('message', handleMessage);
  });

  onUnmounted(() => {
    window.removeEventListener('message', handleMessage);
    if (adState.value.timerId) {
      clearInterval(adState.value.timerId);
    }
  });

  // При смене проекта перезагружаем сохраненный стейт
  watch(
    () => projectId,
    () => loadStorage()
  );

  return {
    appEnv,
    logs,
    playerStorage,
    adConfig,
    adState,
    purchaseCatalog,
    purchaseModal,
    clearStorage,
    clearLogs: () => {
      logs.value = [];
    },
    skipAd,
    failAd,
    handlePurchaseDecision,
  };
}
