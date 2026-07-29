<template>
  <div v-if="homeContent" class="min-h-screen bg-white">
    <iframe v-if="isHomeContentUrl" :src="homeContent.trim()" class="h-screen w-full border-0" allowfullscreen></iframe>
    <div v-else v-html="homeContent"></div>
  </div>

  <div v-else class="landing-page flex min-h-screen flex-col text-[#15171a]" :data-palette="palette">
    <header class="landing-header">
      <nav class="landing-container flex h-[74px] items-center justify-between gap-3">
        <router-link to="/home" class="brand-lockup" aria-label="Home">
          <span class="brand-mark">
            <img v-if="siteLogo" :src="siteLogo" :alt="siteName" class="h-full w-full object-contain" />
            <Network v-else :size="22" :stroke-width="1.8" aria-hidden="true" />
          </span>
          <span class="min-w-0">
            <span class="brand-name">{{ siteName }}</span>
            <span class="brand-caption">
              <span class="status-dot"></span>UNIFIED AI GATEWAY
            </span>
          </span>
        </router-link>

        <div class="flex items-center gap-1.5">
          <a
            v-if="docUrl"
            :href="docUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="landing-icon-control"
            :title="t('home.viewDocs')"
            :aria-label="t('home.viewDocs')"
          >
            <BookOpen :size="18" :stroke-width="1.8" aria-hidden="true" />
          </a>
          <LandingAppearancePicker v-model="palette" />
          <LandingLocaleSwitcher />
          <router-link :to="isAuthenticated ? dashboardPath : '/login'" class="landing-nav-action">
            <LayoutDashboard v-if="isAuthenticated" :size="17" :stroke-width="1.9" aria-hidden="true" />
            <LogIn v-else :size="17" :stroke-width="1.9" aria-hidden="true" />
            <span>{{ isAuthenticated ? t('home.dashboard') : t('home.login') }}</span>
          </router-link>
        </div>
      </nav>
    </header>

    <main class="flex-1">
      <section class="hero-section">
        <div class="landing-container hero-layout">
          <div class="hero-copy">
            <div class="hero-kicker">
              <span class="hero-kicker-mark"></span>
              <span>AI ACCESS INFRASTRUCTURE</span>
            </div>
            <h1>{{ siteName }}</h1>
            <p class="hero-lead">{{ t('home.heroSubtitle') }}</p>
            <p class="hero-description">{{ t('home.heroDescription') }}</p>

            <div class="hero-actions">
              <router-link :to="isAuthenticated ? dashboardPath : '/login'" class="landing-primary-action">
                <span>{{ isAuthenticated ? t('home.goToDashboard') : t('home.getStarted') }}</span>
                <ArrowRight :size="17" :stroke-width="2" aria-hidden="true" />
              </router-link>
              <a v-if="docUrl" :href="docUrl" target="_blank" rel="noopener noreferrer" class="landing-secondary-action">
                <BookOpen :size="17" :stroke-width="1.9" aria-hidden="true" />
                <span>{{ t('home.docs') }}</span>
              </a>
            </div>

            <div class="hero-endpoint">
              <div class="endpoint-command">
                <SquareTerminal :size="18" :stroke-width="1.8" aria-hidden="true" />
                <div class="min-w-0">
                  <span>UNIFIED ENDPOINT</span>
                  <code>{{ apiEndpoint }}</code>
                </div>
              </div>
              <span class="endpoint-ready"><CircleCheck :size="16" :stroke-width="2" aria-hidden="true" />READY</span>
            </div>
          </div>

          <div class="gateway-console" aria-label="Live API gateway overview">
            <div class="console-toolbar">
              <div class="console-title">
                <span class="console-index">01</span>
                <div><strong>Route intelligence</strong><small>LIVE CONTROL PLANE</small></div>
              </div>
              <span class="console-live"><span></span>ONLINE</span>
            </div>

            <div class="request-summary">
              <div class="request-copy">
                <span>REQUEST INGRESS</span>
                <div><b>POST</b><code>{{ apiEndpoint }}/responses</code></div>
              </div>
              <div class="request-result"><span>200 OK</span><small>128 ms</small></div>
            </div>

            <div class="console-section-heading"><span>ACTIVE PROVIDER ROUTES</span><span>04 CONNECTED</span></div>
            <div class="route-list">
              <div v-for="provider in providerRoutes" :key="provider.name" class="route-row">
                <span class="route-icon" :class="provider.tone">
                  <component :is="provider.icon" :size="17" :stroke-width="1.8" aria-hidden="true" />
                </span>
                <div class="min-w-0 flex-1">
                  <strong>{{ provider.name }}</strong>
                  <span>{{ provider.mode }} routing</span>
                </div>
                <div class="route-telemetry"><strong>{{ provider.latency }}</strong><span>{{ provider.load }} load</span></div>
                <CircleCheck :size="17" :stroke-width="2" class="route-check" aria-hidden="true" />
              </div>
            </div>

            <div class="console-stats">
              <div><span>SUCCESS RATE</span><strong>99.98%</strong></div>
              <div><span>ROUTING MODE</span><strong>Adaptive</strong></div>
              <div><span>BILLING</span><strong>Realtime</strong></div>
            </div>
          </div>
        </div>
      </section>

      <section class="capability-band">
        <div class="landing-container capability-grid">
          <div v-for="(item, index) in capabilities" :key="item.label" class="capability-item">
            <span class="capability-number">0{{ index + 1 }}</span>
            <component :is="item.icon" :size="20" :stroke-width="1.8" aria-hidden="true" />
            <span>{{ item.label }}</span>
            <ArrowUpRight :size="16" :stroke-width="1.8" aria-hidden="true" />
          </div>
        </div>
      </section>

      <section class="provider-section">
        <div class="landing-container">
          <div class="section-intro">
            <div class="section-copy">
              <p class="section-label"><span>01</span> PROVIDER NETWORK</p>
              <h2>{{ t('home.providers.title') }}</h2>
              <p>{{ t('home.providers.description') }}</p>
            </div>
            <div class="endpoint-note">
              <KeyRound :size="19" :stroke-width="1.8" aria-hidden="true" />
              <div><span>ONE KEY / EVERY ROUTE</span><code>{{ apiEndpoint }}</code></div>
            </div>
          </div>

          <div class="provider-table">
            <div class="provider-table-heading"><span>PROVIDER</span><span>ROUTING POLICY</span><span>LATENCY</span><span>STATUS</span></div>
            <div v-for="(provider, index) in providerRoutes" :key="'table-' + provider.name" class="provider-table-row">
              <div class="flex min-w-0 items-center gap-3">
                <span class="provider-number">0{{ index + 1 }}</span>
                <span class="route-icon" :class="provider.tone">
                  <component :is="provider.icon" :size="18" :stroke-width="1.8" aria-hidden="true" />
                </span>
                <strong>{{ provider.name }}</strong>
              </div>
              <span class="provider-mode">{{ provider.mode }}</span>
              <span class="provider-latency">{{ provider.latency }}</span>
              <span class="provider-health"><span></span>{{ t('home.providers.supported') }}</span>
            </div>
          </div>
        </div>
      </section>

      <section class="feature-section">
        <div class="landing-container">
          <div class="feature-heading">
            <div><p class="section-label"><span>02</span> OPERATIONS LAYER</p><h2>{{ t('home.solutions.title') }}</h2></div>
            <p>{{ t('home.solutions.subtitle') }}</p>
          </div>
          <div class="feature-grid">
            <article v-for="feature in features" :key="feature.title" class="feature-card">
              <div class="feature-card-top">
                <span class="feature-number">{{ feature.number }}</span>
                <span class="feature-icon"><component :is="feature.icon" :size="22" :stroke-width="1.8" aria-hidden="true" /></span>
              </div>
              <h3>{{ feature.title }}</h3>
              <p>{{ feature.description }}</p>
            </article>
          </div>
        </div>
      </section>

      <section class="final-cta">
        <div class="landing-container final-cta-inner">
          <div class="cta-brand"><span>READY TO ROUTE</span><p>{{ siteName }}</p></div>
          <div class="cta-subtitle">{{ siteSubtitle }}</div>
          <router-link :to="isAuthenticated ? dashboardPath : '/login'" class="landing-primary-action">
            <span>{{ isAuthenticated ? t('home.goToDashboard') : t('home.getStarted') }}</span>
            <ArrowUpRight :size="17" :stroke-width="2" aria-hidden="true" />
          </router-link>
        </div>
      </section>
    </main>

    <footer class="landing-footer">
      <div class="landing-container landing-footer-inner">
        <p>&copy; {{ currentYear }} {{ siteName }}. {{ t('home.footer.allRightsReserved') }}</p>
        <div>
          <a v-if="docUrl" :href="docUrl" target="_blank" rel="noopener noreferrer">{{ t('home.docs') }}</a>
          <a :href="githubUrl" target="_blank" rel="noopener noreferrer">GitHub</a>
        </div>
      </div>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  ArrowRight, ArrowUpRight, BookOpen, Bot, CircleCheck, Cloud, Gauge, KeyRound,
  LayoutDashboard, LogIn, Network, Orbit, ReceiptText, Route, ShieldCheck,
  Sparkles, SquareTerminal, Users
} from 'lucide-vue-next'
import { useAuthStore, useAppStore } from '@/stores'
import LandingAppearancePicker, { type LandingPalette } from '@/components/landing/LandingAppearancePicker.vue'
import LandingLocaleSwitcher from '@/components/landing/LandingLocaleSwitcher.vue'
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
const docUrl = computed(() => sanitizeUrl(appStore.cachedPublicSettings?.doc_url || appStore.docUrl || ''))
const homeContent = computed(() => appStore.cachedPublicSettings?.home_content || '')
const isHomeContentUrl = computed(() => /^https?:\/\//i.test(homeContent.value.trim()))
const isAuthenticated = computed(() => authStore.isAuthenticated)
const dashboardPath = computed(() => (authStore.isAdmin ? '/admin/dashboard' : '/dashboard'))
const currentYear = computed(() => new Date().getFullYear())
const apiEndpoint = computed(() => window.location.origin + '/v1')
const githubUrl = 'https://github.com/Wei-Shaw/sub2api'
const palette = ref<LandingPalette>('cobalt')

const capabilities = computed(() => [
  { icon: Route, label: t('home.tags.subscriptionToApi') },
  { icon: ShieldCheck, label: t('home.tags.stickySession') },
  { icon: ReceiptText, label: t('home.tags.realtimeBilling') }
])

const providerRoutes = computed(() => [
  { name: t('home.providers.claude'), icon: Bot, tone: 'tone-coral', latency: '112 ms', load: '68%', mode: 'Priority' },
  { name: 'GPT', icon: Sparkles, tone: 'tone-gold', latency: '128 ms', load: '82%', mode: 'Balanced' },
  { name: t('home.providers.gemini'), icon: Orbit, tone: 'tone-blue', latency: '96 ms', load: '54%', mode: 'Fastest' },
  { name: t('home.providers.antigravity'), icon: Cloud, tone: 'tone-violet', latency: '141 ms', load: '44%', mode: 'Standby' }
])

const features = computed(() => [
  { number: '01', icon: Network, title: t('home.features.unifiedGateway'), description: t('home.features.unifiedGatewayDesc') },
  { number: '02', icon: Users, title: t('home.features.multiAccount'), description: t('home.features.multiAccountDesc') },
  { number: '03', icon: Gauge, title: t('home.features.balanceQuota'), description: t('home.features.balanceQuotaDesc') }
])

function isLandingPalette(value: string | null): value is LandingPalette {
  return value === 'cobalt' || value === 'graphite' || value === 'vermilion'
}

watch(palette, (value) => window.localStorage.setItem('landing-palette', value))

onMounted(() => {
  const storedPalette = window.localStorage.getItem('landing-palette')
  if (isLandingPalette(storedPalette)) palette.value = storedPalette
  authStore.checkAuth()
  if (!appStore.publicSettingsLoaded) appStore.fetchPublicSettings()
})
</script>


<style scoped>
.landing-page {
  --accent: #2454d6;
  --accent-strong: #173ea8;
  --accent-soft: #eef2ff;
  --accent-border: #c9d5ff;
  background: #f4f4f2;
  color: #15171a;
}

.landing-page[data-palette='graphite'] {
  --accent: #24262b;
  --accent-strong: #0e0f11;
  --accent-soft: #efefee;
  --accent-border: #cfd0d2;
}

.landing-page[data-palette='vermilion'] {
  --accent: #d9472b;
  --accent-strong: #b93420;
  --accent-soft: #fff0ec;
  --accent-border: #f3c9bf;
}

.landing-container {
  width: min(100% - 48px, 1320px);
  margin-inline: auto;
}

.landing-header {
  position: sticky;
  top: 0;
  z-index: 40;
  border-bottom: 1px solid #dedfdf;
  border-bottom-color: #dedfdf;
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(16px);
}

.brand-lockup {
  display: flex;
  min-width: 0;
  align-items: center;
  gap: 12px;
}

.brand-mark {
  display: inline-flex;
  width: 42px;
  height: 42px;
  flex: none;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  border: 1px solid var(--accent);
  border-color: var(--accent);
  border-radius: 6px;
  background: var(--accent);
  color: #fff;
}

.brand-name {
  display: block;
  overflow: hidden;
  color: #15171a;
  font-size: 16px;
  font-weight: 700;
  line-height: 1.2;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.brand-caption {
  display: flex;
  align-items: center;
  gap: 7px;
  margin-top: 4px;
  color: #76797f;
  font-size: 9px;
  font-weight: 700;
}

.status-dot {
  width: 5px;
  height: 5px;
  border-radius: 9999px;
  background: var(--accent);
  box-shadow: none;
}

.landing-icon-control,
.landing-nav-action,
.landing-primary-action,
.landing-secondary-action {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border-radius: 6px;
  font-weight: 650;
  transition: border-color 160ms ease, background-color 160ms ease, color 160ms ease;
}

.landing-icon-control {
  width: 40px;
  height: 40px;
  border: 1px solid #d8dadd;
  border-color: #d8dadd;
  background: #fff;
  color: #40434a;
}

.landing-icon-control:hover {
  border-color: #a8abb0;
  background: #f5f5f4;
  color: #15171a;
}

.landing-nav-action {
  min-height: 42px;
  gap: 7px;
  padding-inline: 16px;
  background: #15171a;
  color: #fff;
  font-size: 13px;
}

.landing-nav-action:hover {
  background: #303238;
}

.hero-section {
  border-bottom: 1px solid #dedfdf;
  border-bottom-color: #dedfdf;
  background: #fbfbfa;
}

.hero-layout {
  display: grid;
  min-height: min(760px, calc(100svh - 112px));
  grid-template-columns: minmax(0, 0.94fr) minmax(500px, 1.06fr);
  align-items: center;
  gap: 80px;
  padding-block: 64px;
}

.hero-copy {
  min-width: 0;
  max-width: 620px;
}

.hero-kicker {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  color: #62656b;
  font-size: 10px;
  font-weight: 750;
}

.hero-kicker-mark {
  width: 22px;
  height: 3px;
  background: var(--accent);
}

.hero-copy h1 {
  max-width: none;
  margin-top: 24px;
  color: #111316;
  font-size: 76px;
  font-weight: 680;
  line-height: 1.02;
}

.hero-lead {
  max-width: 560px;
  margin-top: 26px;
  color: #202329;
  font-size: 26px;
  font-weight: 620;
  line-height: 1.42;
}

.hero-description {
  max-width: 560px;
  margin-top: 12px;
  color: #686b72;
  font-size: 16px;
  line-height: 1.8;
}

.hero-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  margin-top: 32px;
}

.landing-primary-action,
.landing-secondary-action {
  min-height: 48px;
  padding-inline: 20px;
  font-size: 14px;
}

.landing-primary-action {
  gap: 8px;
  background: var(--accent);
  color: #fff;
}

.landing-primary-action:hover {
  background: var(--accent-strong);
}

.landing-secondary-action {
  gap: 8px;
  border: 1px solid #cfd1d4;
  border-color: #cfd1d4;
  background: #fff;
  color: #34373d;
}

.landing-secondary-action:hover {
  border-color: #9da0a5;
  background: #f5f5f4;
  color: #15171a;
}

.hero-endpoint {
  display: flex;
  max-width: 560px;
  min-height: 70px;
  align-items: center;
  justify-content: space-between;
  gap: 20px;
  margin-top: 40px;
  border-block: 1px solid #dedfdf;
  color: var(--accent);
}

.endpoint-command {
  display: flex;
  min-width: 0;
  align-items: center;
  gap: 12px;
}

.endpoint-command span,
.endpoint-command code {
  display: block;
}

.endpoint-command span {
  color: #8a8d92;
  font-size: 9px;
  font-weight: 700;
}

.endpoint-command code {
  overflow: hidden;
  margin-top: 4px;
  color: #2c2f35;
  font-size: 12px;
  font-weight: 650;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.endpoint-ready {
  display: inline-flex;
  flex: none;
  align-items: center;
  gap: 6px;
  color: var(--accent);
  font-size: 9px;
  font-weight: 800;
}

.gateway-console {
  min-width: 0;
  overflow: hidden;
  border: 1px solid #cfd1d4;
  border-color: #cfd1d4;
  border-top: 4px solid var(--accent);
  border-radius: 4px;
  background: #fff;
  box-shadow: 0 28px 70px rgba(22, 24, 28, 0.1);
}

.console-toolbar {
  display: flex;
  min-height: 68px;
  align-items: center;
  justify-content: space-between;
  border-bottom: 1px solid #dedfdf;
  border-bottom-color: #dedfdf;
  background: #fff;
  padding-inline: 22px;
  color: #3a3d43;
  font-size: 12px;
  font-weight: 650;
}

.console-title {
  display: flex;
  align-items: center;
  gap: 14px;
}

.console-index {
  display: inline-flex;
  width: 34px;
  height: 34px;
  align-items: center;
  justify-content: center;
  border: 1px solid #d8dadd;
  color: #74777d;
  font-size: 9px;
  font-weight: 800;
}

.console-title strong,
.console-title small {
  display: block;
}

.console-title strong {
  color: #202329;
  font-size: 13px;
  font-weight: 700;
}

.console-title small {
  margin-top: 3px;
  color: #92959a;
  font-size: 8px;
  font-weight: 700;
}

.console-live {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  color: var(--accent);
  font-size: 9px;
  font-weight: 800;
}

.console-live span {
  width: 6px;
  height: 6px;
  border-radius: 9999px;
  background: var(--accent);
  box-shadow: none;
}

.request-summary {
  display: flex;
  min-height: 88px;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  border-bottom: 1px solid #dedfdf;
  border-bottom-color: #dedfdf;
  background: #f4f4f2;
  padding: 16px 22px;
}

.request-copy {
  min-width: 0;
}

.request-copy > span {
  color: #8a8d92;
  font-size: 8px;
  font-weight: 800;
}

.request-copy div {
  display: flex;
  min-width: 0;
  align-items: center;
  gap: 10px;
  margin-top: 8px;
}

.request-copy b {
  flex: none;
  border-radius: 3px;
  background: var(--accent-soft);
  padding: 4px 6px;
  color: var(--accent);
  font-size: 9px;
}

.request-copy code {
  overflow: hidden;
  color: #282b30;
  font-size: 11px;
  font-weight: 650;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.request-result span {
  display: block;
  color: var(--accent);
  font-size: 12px;
  font-weight: 750;
}

.request-result {
  display: flex;
  flex: none;
  flex-direction: column;
  align-items: flex-end;
}

.request-result small {
  margin-top: 3px;
  color: #96999e;
  font-size: 9px;
}

.console-section-heading {
  display: grid;
  grid-template-columns: 1fr auto;
  padding: 18px 22px 8px;
  color: #95989d;
  font-size: 8px;
  font-weight: 800;
}

.route-list {
  padding: 0 14px 14px;
}

.route-row {
  display: flex;
  min-height: 66px;
  align-items: center;
  gap: 13px;
  border-bottom: 1px solid #e7e7e5;
  border-bottom-color: #e7e7e5;
  padding-inline: 8px;
}

.route-row:last-child {
  border-bottom: 0;
}

.route-icon {
  display: inline-flex;
  width: 36px;
  height: 36px;
  flex: none;
  align-items: center;
  justify-content: center;
  border-radius: 5px;
}

.tone-coral {
  background: #fff0eb;
  color: #c43e23;
}

.tone-gold {
  background: #fff5d9;
  color: #9a6700;
}

.tone-blue {
  background: #edf2ff;
  color: #2454d6;
}

.tone-violet {
  background: #f3efff;
  color: #6d45bd;
}

.route-row strong,
.route-row span {
  display: block;
}

.route-row strong {
  color: #2a2d32;
  font-size: 12px;
}

.route-row span {
  margin-top: 3px;
  color: #8a8d92;
  font-size: 9px;
}

.route-telemetry {
  min-width: 68px;
  text-align: right;
}

.route-telemetry strong {
  color: #3a3d43;
  font-size: 11px;
}

.route-telemetry span {
  color: #a0a2a7;
  font-size: 8px;
}

.route-check {
  flex: none;
  color: var(--accent);
}

.console-stats {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  border-top: 1px solid #dedfdf;
  background: #f4f4f2;
}

.console-stats div {
  min-width: 0;
  padding: 15px 16px 16px;
}

.console-stats div + div {
  border-left: 1px solid #dedfdf;
}

.console-stats span,
.console-stats strong {
  display: block;
}

.console-stats span {
  overflow: hidden;
  color: #92959a;
  font-size: 8px;
  font-weight: 700;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.console-stats strong {
  margin-top: 5px;
  color: #202329;
  font-size: 12px;
  font-weight: 700;
}

.capability-band {
  border-bottom: 0;
  background: #17191d;
}

.capability-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.capability-item {
  display: grid;
  min-height: 88px;
  grid-template-columns: auto auto minmax(0, 1fr) auto;
  gap: 14px;
  padding-inline: 28px;
  color: #fff;
  font-size: 13px;
  font-weight: 650;
}

.capability-item + .capability-item {
  border-left: 1px solid #34363b;
  border-left-color: #34363b;
}

.capability-item > svg:first-of-type {
  color: var(--accent-border);
}

.capability-item > svg:last-of-type {
  color: #777a81;
}

.capability-number {
  color: #666970;
  font-size: 9px;
  font-weight: 800;
}

.provider-section {
  border-bottom: 1px solid #dedfdf;
  border-bottom-color: #dedfdf;
  background: #fff;
  padding: 112px 0 116px;
}

.section-intro {
  display: flex;
  align-items: end;
  justify-content: space-between;
  gap: 64px;
}

.section-copy {
  max-width: 700px;
}

.section-label {
  display: flex;
  align-items: center;
  gap: 10px;
  color: #696c72;
  font-size: 10px;
}

.section-label span {
  color: var(--accent);
}

.section-copy h2,
.feature-heading h2 {
  margin-top: 16px;
  color: #15171a;
  font-size: 48px;
  font-weight: 650;
  line-height: 1.14;
}

.section-copy > p:last-of-type,
.feature-heading > p {
  margin-top: 18px;
  color: #6b6e74;
  font-size: 15px;
}

.endpoint-note {
  display: flex;
  min-width: 340px;
  align-items: center;
  gap: 12px;
  margin-top: 0;
  border-top: 0;
  border-left: 3px solid var(--accent);
  background: #f4f4f2;
  padding: 18px 20px;
  color: var(--accent);
}

.endpoint-note span,
.endpoint-note code {
  display: block;
}

.endpoint-note span {
  color: #8c8f94;
  font-size: 9px;
  font-weight: 750;
}

.endpoint-note code {
  overflow: hidden;
  max-width: 300px;
  margin-top: 4px;
  color: #2d3035;
  font-size: 12px;
  font-weight: 650;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.provider-table {
  margin-top: 58px;
  overflow: visible;
  border: 0;
  border-block: 1px solid #bfc1c4;
  border-radius: 0;
}

.provider-table-heading,
.provider-table-row {
  display: grid;
  grid-template-columns: minmax(0, 1.45fr) minmax(120px, 0.7fr) minmax(100px, 0.5fr) minmax(120px, 0.6fr);
  align-items: center;
  padding-inline: 20px;
}

.provider-table-heading {
  min-height: 48px;
  border-bottom-color: #d5d6d7;
  background: transparent;
  color: #8a8d92;
  font-size: 9px;
  font-weight: 750;
}

.provider-table-row {
  min-height: 82px;
  color: #34373d;
}

.provider-table-row + .provider-table-row {
  border-top-color: #e1e2e2;
}

.provider-table-row strong {
  overflow: hidden;
  color: #26292e;
  font-size: 14px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.provider-number {
  width: 22px;
  flex: none;
  color: #999ca1;
  font-size: 9px;
  font-weight: 700;
}

.provider-mode,
.provider-latency {
  color: #676a70;
  font-size: 12px;
}

.provider-health {
  display: inline-flex;
  align-items: center;
  gap: 7px;
  color: var(--accent);
  font-size: 11px;
  font-weight: 650;
}

.provider-health span {
  width: 7px;
  height: 7px;
  border-radius: 9999px;
  background: var(--accent);
}

.feature-section {
  background: #f1f1ef;
  padding: 108px 0 116px;
}

.feature-heading {
  display: flex;
  align-items: end;
  justify-content: space-between;
  gap: 32px;
}

.feature-heading > p {
  max-width: 420px;
  margin: 0;
}

.feature-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 0;
  margin-top: 54px;
  border-block: 1px solid #c8c9ca;
}

.feature-card {
  min-height: 310px;
  border: 0;
  border-radius: 0;
  background: transparent;
  padding: 32px 34px 38px;
  box-shadow: none;
  transition: background-color 160ms ease;
}

.feature-card + .feature-card {
  border-left: 1px solid #c8c9ca;
}

.feature-card:hover {
  transform: none;
  border-color: #c8c9ca;
  background: #fff;
  box-shadow: none;
}

.feature-card-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  color: var(--accent);
}

.feature-number {
  color: #93969b;
  font-size: 10px;
  font-weight: 750;
}

.feature-icon {
  display: inline-flex;
  width: 44px;
  height: 44px;
  align-items: center;
  justify-content: center;
  background: var(--accent-soft);
  color: var(--accent);
}

.feature-card h3 {
  margin-top: 64px;
  color: #202329;
  font-size: 21px;
}

.feature-card p {
  max-width: 330px;
  margin-top: 14px;
  color: #686b71;
  font-size: 14px;
  line-height: 1.8;
}

.final-cta {
  border-block: 0;
  background: var(--accent);
}

.final-cta-inner {
  display: flex;
  min-height: 190px;
  align-items: center;
  justify-content: space-between;
  gap: 40px;
  color: #fff;
}

.cta-brand {
  min-width: 220px;
}

.cta-brand > span {
  display: block;
  color: rgba(255, 255, 255, 0.7);
  font-size: 9px;
  font-weight: 800;
}

.cta-brand p {
  margin-top: 8px;
  color: #fff;
  font-size: 38px;
  font-weight: 680;
  line-height: 1.1;
}

.cta-subtitle {
  min-width: 0;
  flex: 1;
  color: rgba(255, 255, 255, 0.78);
  font-size: 14px;
}

.final-cta .landing-primary-action {
  flex: none;
  background: #fff;
  color: #15171a;
}

.final-cta .landing-primary-action:hover {
  background: #f0f0ee;
}

.final-cta .landing-primary-action span {
  color: #15171a;
}

.landing-footer-inner {
  display: flex;
  min-height: 84px;
  align-items: center;
  justify-content: space-between;
  gap: 20px;
  color: #73767c;
  font-size: 12px;
}

.landing-footer {
  background: #fff;
}

.landing-footer-inner div {
  display: flex;
  gap: 20px;
}

.landing-footer-inner a:hover {
  color: #15171a;
}

@media (max-width: 1150px) {
  .hero-layout {
    grid-template-columns: minmax(0, 0.82fr) minmax(470px, 1.18fr);
    gap: 44px;
  }

  .hero-copy h1 {
    font-size: 64px;
  }
}

@media (max-width: 1023px) {
  .hero-layout {
    min-height: 0;
    grid-template-columns: 1fr;
    gap: 56px;
    padding-block: 72px;
  }

  .hero-copy {
    max-width: 720px;
  }

  .gateway-console {
    width: min(100%, 760px);
  }

  .section-intro {
    display: block;
  }

  .endpoint-note {
    width: min(100%, 500px);
    margin-top: 32px;
  }
}

@media (max-width: 767px) {
  .landing-container {
    width: min(100% - 28px, 1320px);
  }

  .brand-caption {
    display: none;
  }

  .hero-layout {
    gap: 40px;
    padding-block: 52px 44px;
  }

  .hero-copy h1 {
    margin-top: 20px;
    font-size: 50px;
  }

  .hero-lead {
    margin-top: 22px;
    font-size: 21px;
  }

  .hero-description {
    font-size: 14px;
  }

  .hero-endpoint {
    min-height: 64px;
    margin-top: 32px;
  }

  .console-toolbar,
  .request-summary {
    padding-inline: 16px;
  }

  .console-section-heading {
    padding-inline: 16px;
  }

  .route-list {
    padding-inline: 8px;
  }

  .capability-grid {
    grid-template-columns: 1fr;
  }

  .capability-item {
    min-height: 72px;
    padding-inline: 10px;
  }

  .capability-item + .capability-item {
    border-top: 1px solid #34363b;
    border-left: 0;
    border-top-color: #34363b;
  }

  .provider-section,
  .feature-section {
    padding-block: 76px 82px;
  }

  .section-copy h2,
  .feature-heading h2 {
    font-size: 38px;
  }

  .feature-heading {
    display: block;
  }

  .feature-heading > p {
    margin-top: 14px;
  }

  .endpoint-note {
    min-width: 0;
  }

  .provider-table {
    margin-top: 40px;
  }

  .provider-table-heading,
  .provider-table-row {
    grid-template-columns: minmax(0, 1fr) 96px;
    padding-inline: 10px;
  }

  .provider-table-heading span:nth-child(2),
  .provider-table-heading span:nth-child(3),
  .provider-mode,
  .provider-latency {
    display: none;
  }

  .feature-grid {
    grid-template-columns: 1fr;
    margin-top: 38px;
  }

  .feature-card {
    min-height: 250px;
    padding: 28px 20px 32px;
  }

  .feature-card + .feature-card {
    border-top: 1px solid #c8c9ca;
    border-left: 0;
  }

  .feature-card h3 {
    margin-top: 38px;
  }

  .final-cta-inner {
    min-height: 0;
    align-items: flex-start;
    flex-direction: column;
    padding-block: 44px;
  }

  .cta-brand p {
    font-size: 32px;
  }

  .cta-subtitle {
    width: 100%;
  }

  .landing-footer-inner {
    align-items: flex-start;
    flex-direction: column;
    padding-block: 24px;
  }
}

@media (max-width: 520px) {
  .landing-nav-action {
    width: 40px;
    padding: 0;
  }

  .landing-nav-action span {
    display: none;
  }

  .hero-actions {
    display: grid;
    grid-template-columns: 1fr;
  }

  .hero-actions > * {
    width: 100%;
  }

  .hero-copy h1 {
    font-size: 46px;
  }

  .hero-endpoint {
    align-items: flex-start;
    flex-direction: column;
    gap: 12px;
    padding-block: 15px;
  }

  .route-row:nth-child(n + 3) {
    display: flex;
  }

  .console-stats {
    display: grid;
  }

  .console-stats div {
    padding-inline: 10px;
  }

  .provider-number {
    display: none;
  }
}

@media (max-width: 400px) {
  .brand-mark {
    width: 38px;
    height: 38px;
  }

  .brand-name {
    max-width: 78px;
    font-size: 14px;
  }

  .route-telemetry {
    display: none;
  }

  .console-stats strong {
    font-size: 10px;
  }
}

@media (max-width: 380px) {
  .landing-icon-control {
    display: none;
  }
}
</style>
