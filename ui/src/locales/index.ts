import { useCookies } from 'vue3-cookies';
const { cookies } = useCookies();
import { message } from './modules';
import { createI18n } from 'vue-i18n';
const storeLang = cookies.get('lang');
const cookieLang = cookies.get('django_language');

const browserLang = navigator.language || (navigator.languages && navigator.languages[0]) || 'en';

export const LanguageCode = cookieLang || storeLang || browserLang || 'en';
import date from './date';

const i18n = createI18n({
  locale: LanguageCode,
  fallbackLocale: 'en',
  legacy: false,
  allowComposition: true,
  silentFallbackWarn: true,
  silentTranslationWarn: true,
  messages: message,
  dateTimeFormats: date
});

export default i18n;
