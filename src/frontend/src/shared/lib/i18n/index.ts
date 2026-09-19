import { createI18n, type I18n } from 'vue-i18n';
import ru from './locales/ru';
import en from './locales/en';

export const LOCALE_STORAGE_KEY = 'gdh_locale';

export interface SupportedLocale {
  code: string;
  label: string;
  shortLabel: string;
  flag: string;
}

export const SUPPORTED_LOCALES: SupportedLocale[] = [
  { code: 'ru', label: 'Русский', shortLabel: 'RU', flag: '🇷🇺' },
  { code: 'en', label: 'English', shortLabel: 'EN', flag: '🇬🇧' },
];

export function detectInitialLocale(): string {
  const savedLocale = localStorage.getItem(LOCALE_STORAGE_KEY);
  if (savedLocale && (savedLocale === 'ru' || savedLocale === 'en')) {
    return savedLocale;
  }

  // Detect from browser/system language
  if (typeof navigator !== 'undefined' && navigator.language) {
    const navLang = navigator.language.toLowerCase();
    if (
      navLang.startsWith('ru') ||
      navLang.startsWith('be') ||
      navLang.startsWith('uk') ||
      navLang.startsWith('kk')
    ) {
      return 'ru';
    }
    if (navLang.startsWith('en')) {
      return 'en';
    }
  }

  return 'ru'; // Default to Russian as per specification
}

export const i18n = createI18n({
  legacy: false,
  locale: detectInitialLocale(),
  fallbackLocale: 'ru',
  messages: {
    ru,
    en,
  },
});

export function setLocale(locale: string): void {
  if (locale === 'ru' || locale === 'en') {
    (i18n.global.locale as any).value = locale;
    localStorage.setItem(LOCALE_STORAGE_KEY, locale);
    if (typeof document !== 'undefined') {
      document.documentElement.lang = locale;
    }
  }
}

export function getLocale(): string {
  return (i18n.global.locale as any).value;
}

export function useLocale() {
  return {
    currentLocale: i18n.global.locale,
    setLocale,
    supportedLocales: SUPPORTED_LOCALES,
  };
}

export { useI18n } from 'vue-i18n';
