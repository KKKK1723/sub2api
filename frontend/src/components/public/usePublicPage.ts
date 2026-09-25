import { onBeforeMount, onUnmounted, ref } from 'vue'
import { getLocale, setLocale } from '@/i18n'
import './public.css'

let mountedPages = 0

export function syncPublicTheme(dark: boolean) {
  const theme = dark ? 'dark' : 'light'
  localStorage.setItem('keepcoding-theme', theme)
  localStorage.setItem('theme', theme)
  document.documentElement.classList.toggle('dark', dark)
}

export function syncPublicLocale(locale: 'zh' | 'en') {
  localStorage.setItem('keepcoding-locale', locale)
  void setLocale(locale)
}

export function usePublicPage() {
  const locale = ref(getLocale())
  const isDark = ref((localStorage.getItem('theme') ?? localStorage.getItem('keepcoding-theme')) === 'dark')
  onBeforeMount(() => {
    mountedPages++
    document.documentElement.classList.add('kc-public')
  })
  onUnmounted(() => {
    if (--mountedPages === 0) document.documentElement.classList.remove('kc-public')
  })
  return { locale, isDark }
}
