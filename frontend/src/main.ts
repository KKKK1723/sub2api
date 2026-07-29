import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'
import i18n, { initI18n } from './i18n'
import { useAppStore } from '@/stores/app'
import { updateFavicon } from '@/utils/branding'
import { isIOSDevice } from '@/utils/device'
import './style.css'
import './styles/light-refresh.css'

function initIOSViewportZoomFix() {
  if (!isIOSDevice()) return

  const viewport = document.querySelector('meta[name="viewport"]')
  if (!viewport) return

  const content = viewport.getAttribute('content') || ''
  if (/maximum-scale/i.test(content)) return
  viewport.setAttribute('content', `${content}, maximum-scale=1.0`)
}

function initLightTheme() {
  document.documentElement.classList.remove('dark')
  document.documentElement.style.colorScheme = 'light'
  localStorage.setItem('theme', 'light')
}

async function bootstrap() {
  initLightTheme()
  initIOSViewportZoomFix()

  const app = createApp(App)
  const pinia = createPinia()
  app.use(pinia)

  const appStore = useAppStore()
  appStore.initFromInjectedConfig()

  document.title = `${appStore.siteName || '筏&API'} - AI API Gateway`
  updateFavicon(appStore.siteLogo)

  await initI18n()

  app.use(router)
  app.use(i18n)

  await router.isReady()
  app.mount('#app')
}

bootstrap()
