import { createI18n } from 'vue-i18n'

export type Locale = 'en' | 'de'

const SUPPORTED_LOCALES: Locale[] = ['en', 'de']

function detectLocale(): Locale {
  const saved = localStorage.getItem('locale')
  if (saved && SUPPORTED_LOCALES.includes(saved as Locale)) return saved as Locale
  const browser = navigator.language.split('-')[0]
  if (SUPPORTED_LOCALES.includes(browser as Locale)) return browser as Locale
  return 'en'
}

const i18n = createI18n({
  legacy: false,
  locale: detectLocale(),
  fallbackLocale: 'en',
  messages: {},
})

export function setLocale(locale: Locale) {
  ;(i18n.global.locale as { value: string }).value = locale
  localStorage.setItem('locale', locale)
  document.documentElement.lang = locale
}

export function getLocale(): Locale {
  return (i18n.global.locale as { value: string }).value as Locale
}

export { SUPPORTED_LOCALES }
export default i18n
