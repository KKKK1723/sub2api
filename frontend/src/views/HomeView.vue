<template>
  <div v-if="homeContent" class="min-h-screen bg-white">
    <iframe
      v-if="isHomeContentUrl"
      :src="homeContent.trim()"
      class="h-screen w-full border-0"
      allowfullscreen
    ></iframe>
    <div v-else v-html="homeContent"></div>
  </div>

  <div v-else class="landing-page flex min-h-screen flex-col bg-[#f4f7f6] text-gray-950">
    <header class="sticky top-0 z-30 border-b border-gray-200 bg-white/95 backdrop-blur-md">
      <nav class="mx-auto flex h-16 max-w-7xl items-center justify-between gap-4 px-4 sm:px-6 lg:px-8">
        <router-link to="/" class="flex min-w-0 items-center gap-3" aria-label="Home">
          <img
            :src="siteLogo || '/logo.svg'"
            :alt="siteName"
            class="h-9 w-9 flex-none rounded-lg border border-gray-200 bg-white object-contain"
          />
          <span class="truncate text-base font-semibold text-gray-950 sm:text-lg">{{ siteName }}</span>
        </router-link>

        <div class="flex items-center gap-1.5 sm:gap-2">
          <a
            v-if="docUrl"
            :href="docUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="landing-icon-button"
            :title="t('home.viewDocs')"
          >
            <Icon name="book" size="md" />
          </a>
          <LocaleSwitcher />
          <router-link :to="isAuthenticated ? dashboardPath : '/login'" class="landing-nav-action">
            <span>{{ isAuthenticated ? t('home.dashboard') : t('home.login') }}</span>
            <Icon name="arrowRight" size="sm" />
          </router-link>
        </div>
      </nav>
    </header>

    <main class="flex-1">
      <section class="hero-section relative overflow-hidden border-b border-gray-200 bg-white">
        <div class="hero-grid pointer-events-none absolute inset-0" aria-hidden="true"></div>
        <div class="route-line route-line-one" aria-hidden="true"></div>
        <div class="route-line route-line-two" aria-hidden="true"></div>
        <div class="route-line route-line-three" aria-hidden="true"></div>

        <div class="relative mx-auto flex h-full max-w-7xl flex-col justify-center px-4 py-10 sm:px-6 sm:py-12 lg:px-8">
          <div class="max-w-3xl">
            <div class="mb-5 inline-flex items-center gap-2 text-xs font-semibold text-gray-600">
              <span class="h-2 w-2 rounded-full bg-emerald-500"></span>
              <span>AI API GATEWAY</span>
            </div>
            <h1 class="max-w-3xl text-4xl font-semibold leading-[1.08] text-gray-950 sm:text-5xl lg:text-6xl">
              {{ siteName }}
            </h1>
            <p class="mt-5 max-w-2xl text-base leading-7 text-gray-600 sm:text-lg">{{ siteSubtitle }}</p>
            <div class="mt-7 flex flex-wrap items-center gap-3">
              <router-link :to="isAuthenticated ? dashboardPath : '/login'" class="landing-primary-action">
                <span>{{ isAuthenticated ? t('home.goToDashboard') : t('home.getStarted') }}</span>
                <Icon name="arrowRight" size="sm" :stroke-width="2" />
              </router-link>
              <a
                v-if="docUrl"
                :href="docUrl"
                target="_blank"
                rel="noopener noreferrer"
                class="landing-secondary-action"
              >
                <Icon name="book" size="sm" />
                <span>{{ t('home.docs') }}</span>
              </a>
            </div>
          </div>

          <div class="gateway-stage mt-9" aria-label="API gateway routing overview">
            <div class="gateway-source-list">
              <div v-for="provider in providers" :key="provider.name" class="gateway-source">
                <span class="provider-mark" :class="provider.tone">{{ provider.initial }}</span>
                <span class="truncate">{{ provider.name }}</span>
                <span class="ml-auto h-1.5 w-1.5 rounded-full bg-emerald-500"></span>
              </div>
            </div>
            <div class="gateway-flow" aria-hidden="true"><span></span><span></span><span></span></div>
            <div class="gateway-endpoint">
              <div class="flex items-center gap-2 text-xs font-medium text-gray-500">
                <Icon name="terminal" size="sm" />
                <span>OPENAI COMPATIBLE ENDPOINT</span>
              </div>
              <code class="mt-2 block truncate text-sm font-semibold text-gray-950 sm:text-base">{{ apiEndpoint }}</code>
              <div class="mt-3 flex items-center gap-4 text-xs text-gray-500">
                <span class="inline-flex items-center gap-1.5"><span class="h-1.5 w-1.5 rounded-full bg-emerald-500"></span>200 OK</span>
                <span>JSON</span>
                <span>HTTPS</span>
              </div>
            </div>
          </div>
        </div>
      </section>

      <section class="border-b border-gray-200 bg-[#f4f7f6] px-4 py-6 sm:px-6 lg:px-8">
        <div class="mx-auto grid max-w-7xl divide-y divide-gray-200 border-y border-gray-200 sm:grid-cols-3 sm:divide-x sm:divide-y-0">
          <div class="landing-capability"><Icon name="swap" size="md" class="text-teal-700" /><span>{{ t('home.tags.subscriptionToApi') }}</span></div>
          <div class="landing-capability"><Icon name="shield" size="md" class="text-blue-700" /><span>{{ t('home.tags.stickySession') }}</span></div>
          <div class="landing-capability"><Icon name="chart" size="md" class="text-orange-600" /><span>{{ t('home.tags.realtimeBilling') }}</span></div>
        </div>
      </section>

      <section class="bg-white px-4 py-16 sm:px-6 lg:px-8 lg:py-20">
        <div class="mx-auto max-w-7xl">
          <div class="mb-9 max-w-2xl">
            <p class="text-xs font-semibold text-teal-700">GATEWAY CONTROL PLANE</p>
            <h2 class="mt-2 text-2xl font-semibold text-gray-950 sm:text-3xl">{{ t('home.providers.title') }}</h2>
            <p class="mt-3 text-sm leading-6 text-gray-600">{{ t('home.providers.description') }}</p>
          </div>

          <div class="grid gap-4 md:grid-cols-3">
            <article v-for="feature in features" :key="feature.title" class="feature-card">
              <div class="feature-icon" :class="feature.tone"><Icon :name="feature.icon" size="md" /></div>
              <h3 class="mt-5 text-base font-semibold text-gray-950">{{ feature.title }}</h3>
              <p class="mt-2 text-sm leading-6 text-gray-600">{{ feature.description }}</p>
            </article>
          </div>

          <div class="provider-row mt-12">
            <div v-for="provider in providers" :key="`provider-${provider.name}`" class="provider-item">
              <span class="provider-mark" :class="provider.tone">{{ provider.initial }}</span>
              <span class="font-medium text-gray-800">{{ provider.name }}</span>
              <span class="provider-status">{{ t('home.providers.supported') }}</span>
            </div>
            <div class="provider-item text-gray-500">
              <span class="provider-mark bg-gray-100 text-gray-500">+</span>
              <span>{{ t('home.providers.more') }}</span>
              <span class="provider-status provider-status-muted">{{ t('home.providers.soon') }}</span>
            </div>
          </div>
        </div>
      </section>

      <section class="border-y border-gray-200 bg-[#eaf4f1] px-4 py-10 sm:px-6 lg:px-8">
        <div class="mx-auto flex max-w-7xl flex-col items-start justify-between gap-5 sm:flex-row sm:items-center">
          <div><h2 class="text-xl font-semibold text-gray-950">{{ siteName }}</h2><p class="mt-1 text-sm text-gray-600">{{ siteSubtitle }}</p></div>
          <router-link :to="isAuthenticated ? dashboardPath : '/login'" class="landing-primary-action">
            <span>{{ isAuthenticated ? t('home.goToDashboard') : t('home.getStarted') }}</span><Icon name="arrowRight" size="sm" />
          </router-link>
        </div>
      </section>
    </main>

    <footer class="bg-white px-4 py-7 sm:px-6 lg:px-8">
      <div class="mx-auto flex max-w-7xl flex-col items-start justify-between gap-4 text-sm text-gray-500 sm:flex-row sm:items-center">
        <p>&copy; {{ currentYear }} {{ siteName }}. {{ t('home.footer.allRightsReserved') }}</p>
        <div class="flex items-center gap-5">
          <a v-if="docUrl" :href="docUrl" target="_blank" rel="noopener noreferrer" class="footer-link">{{ t('home.docs') }}</a>
          <a :href="githubUrl" target="_blank" rel="noopener noreferrer" class="footer-link">GitHub</a>
        </div>
      </div>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuthStore, useAppStore } from '@/stores'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import Icon from '@/components/icons/Icon.vue'
import { sanitizeUrl } from '@/utils/url'

const { t } = useI18n()
const authStore = useAuthStore()
const appStore = useAppStore()

const siteName = computed(() => appStore.siteName || '筏&API')
const siteLogo = computed(() =>
  sanitizeUrl(appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '', {
    allowRelative: true,
    allowDataUrl: true
  })
)
const siteSubtitle = computed(() => appStore.cachedPublicSettings?.site_subtitle || 'AI API Gateway Platform')
const docUrl = computed(() =>
  sanitizeUrl(appStore.cachedPublicSettings?.doc_url || appStore.docUrl || '')
)
const homeContent = computed(() => appStore.cachedPublicSettings?.home_content || '')
const isHomeContentUrl = computed(() => /^https?:\/\//i.test(homeContent.value.trim()))
const isAuthenticated = computed(() => authStore.isAuthenticated)
const dashboardPath = computed(() => (authStore.isAdmin ? '/admin/dashboard' : '/dashboard'))
const currentYear = computed(() => new Date().getFullYear())
const apiEndpoint = computed(() => `${window.location.origin}/v1`)
const githubUrl = 'https://github.com/Wei-Shaw/sub2api'

const providers = computed(() => [
  { name: t('home.providers.claude'), initial: 'C', tone: 'provider-coral' },
  { name: 'GPT', initial: 'G', tone: 'provider-green' },
  { name: t('home.providers.gemini'), initial: 'G', tone: 'provider-blue' },
  { name: t('home.providers.antigravity'), initial: 'A', tone: 'provider-orange' }
])

const features = computed(() => [
  { icon: 'server' as const, title: t('home.features.unifiedGateway'), description: t('home.features.unifiedGatewayDesc'), tone: 'feature-icon-teal' },
  { icon: 'users' as const, title: t('home.features.multiAccount'), description: t('home.features.multiAccountDesc'), tone: 'feature-icon-blue' },
  { icon: 'dollar' as const, title: t('home.features.balanceQuota'), description: t('home.features.balanceQuotaDesc'), tone: 'feature-icon-orange' }
])

onMounted(() => {
  authStore.checkAuth()
  if (!appStore.publicSettingsLoaded) appStore.fetchPublicSettings()
})
</script>

<style scoped>
.hero-section { height: min(680px, calc(100svh - 112px)); min-height: 510px; }
.hero-grid {
  background-image: linear-gradient(#eef1f0 1px, transparent 1px), linear-gradient(90deg, #eef1f0 1px, transparent 1px);
  background-size: 48px 48px;
  mask-image: linear-gradient(to right, transparent 0, black 45%, black 100%);
}
.route-line { position: absolute; right: -4rem; height: 1px; width: min(48vw, 680px); background: #cbded9; transform-origin: right center; }
.route-line::before, .route-line::after { position: absolute; top: -3px; width: 7px; height: 7px; border-radius: 9999px; background: #0f766e; content: ''; }
.route-line::before { left: 18%; }
.route-line::after { right: 28%; background: #f97316; }
.route-line-one { top: 25%; transform: rotate(-8deg); }
.route-line-two { top: 49%; transform: rotate(4deg); }
.route-line-three { top: 72%; transform: rotate(-3deg); }
.landing-icon-button, .landing-nav-action, .landing-primary-action, .landing-secondary-action {
  display: inline-flex; align-items: center; justify-content: center; border-radius: 8px; font-weight: 600;
  transition: background-color 160ms ease, border-color 160ms ease, color 160ms ease;
}
.landing-icon-button { width: 38px; height: 38px; color: #4b5563; }
.landing-icon-button:hover { background: #f3f4f6; color: #111827; }
.landing-nav-action { min-height: 38px; gap: 7px; padding: 0 13px; background: #111827; color: white; font-size: 0.8125rem; }
.landing-nav-action:hover, .landing-primary-action:hover { background: #1f2937; }
.landing-primary-action, .landing-secondary-action { min-height: 42px; gap: 8px; padding: 0 17px; font-size: 0.875rem; }
.landing-primary-action { background: #111827; color: white; }
.landing-secondary-action { border: 1px solid #d1d5db; background: white; color: #374151; }
.landing-secondary-action:hover { border-color: #9ca3af; background: #f9fafb; color: #111827; }
.gateway-stage {
  display: grid; grid-template-columns: minmax(150px, .85fr) minmax(80px, .5fr) minmax(260px, 1.35fr);
  width: min(760px, 100%); overflow: hidden; border: 1px solid #dfe5e3; border-radius: 8px;
  background: rgba(255,255,255,.94); box-shadow: 0 16px 36px rgba(31,41,55,.08);
}
.gateway-source-list { display: grid; border-right: 1px solid #e5e7eb; }
.gateway-source { display: flex; min-width: 0; align-items: center; gap: 9px; padding: 8px 12px; border-bottom: 1px solid #f0f2f1; font-size: .75rem; color: #4b5563; }
.gateway-source:last-child { border-bottom: 0; }
.provider-mark { display: inline-flex; width: 26px; height: 26px; flex: none; align-items: center; justify-content: center; border-radius: 6px; font-size: .6875rem; font-weight: 700; }
.provider-coral { background: #fff1ed; color: #c2410c; }
.provider-green { background: #eaf8f1; color: #047857; }
.provider-blue { background: #edf4ff; color: #1d4ed8; }
.provider-orange { background: #fff5e7; color: #c2410c; }
.gateway-flow { display: flex; flex-direction: column; justify-content: space-around; padding: 16px 0; }
.gateway-flow span { position: relative; display: block; height: 1px; background: #b8d2cc; }
.gateway-flow span::after { position: absolute; top: -3px; right: -1px; width: 7px; height: 7px; border-radius: 9999px; background: #0f766e; content: ''; }
.gateway-endpoint { display: flex; min-width: 0; flex-direction: column; justify-content: center; border-left: 1px solid #e5e7eb; padding: 18px 20px; }
.landing-capability { display: flex; align-items: center; justify-content: center; gap: 10px; padding: 18px 20px; font-size: .875rem; font-weight: 600; color: #374151; }
.feature-card { min-height: 220px; border: 1px solid #e2e7e5; border-radius: 8px; background: #fbfcfc; padding: 24px; transition: border-color 160ms ease, box-shadow 160ms ease, transform 160ms ease; }
.feature-card:hover { transform: translateY(-2px); border-color: #b7c8c4; box-shadow: 0 12px 28px rgba(31,41,55,.07); }
.feature-icon { display: inline-flex; width: 42px; height: 42px; align-items: center; justify-content: center; border-radius: 8px; }
.feature-icon-teal { background: #e5f4f0; color: #0f766e; }
.feature-icon-blue { background: #eaf2ff; color: #1d4ed8; }
.feature-icon-orange { background: #fff1e6; color: #c2410c; }
.provider-row { display: flex; flex-wrap: wrap; gap: 10px; border-top: 1px solid #e5e7eb; padding-top: 28px; }
.provider-item { display: inline-flex; align-items: center; gap: 9px; border: 1px solid #e1e6e4; border-radius: 8px; background: white; padding: 9px 11px; font-size: .8125rem; }
.provider-status { border-radius: 4px; background: #e8f5f1; padding: 2px 6px; color: #047857; font-size: .625rem; font-weight: 700; }
.provider-status-muted { background: #f3f4f6; color: #6b7280; }
.footer-link { color: #6b7280; transition: color 160ms ease; }
.footer-link:hover { color: #111827; }
@media (max-width: 767px) {
  .hero-section { height: auto; min-height: calc(100svh - 112px); }
  .hero-grid { mask-image: none; opacity: .55; }
  .route-line { display: none; }
  .gateway-stage { grid-template-columns: 1fr; }
  .gateway-source-list { grid-template-columns: repeat(4, minmax(0,1fr)); border-right: 0; border-bottom: 1px solid #e5e7eb; }
  .gateway-source { justify-content: center; padding: 9px 5px; border-right: 1px solid #f0f2f1; border-bottom: 0; }
  .gateway-source > span:nth-child(2), .gateway-source > span:last-child, .gateway-flow { display: none; }
  .gateway-endpoint { border-left: 0; padding: 15px 16px; }
}
@media (max-width: 420px) { .landing-nav-action { width: 38px; padding: 0; } .landing-nav-action span { display: none; } }
</style>
