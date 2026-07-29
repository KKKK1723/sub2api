<template>
  <div v-if="homeContent" class="min-h-screen bg-white">
    <iframe v-if="isHomeContentUrl" :src="homeContent.trim()" class="h-screen w-full border-0" allowfullscreen></iframe>
    <div v-else v-html="homeContent"></div>
  </div>

  <div v-else class="landing-page flex min-h-screen flex-col" :data-palette="palette">
    <header class="landing-header">
      <nav class="landing-container landing-nav">
        <router-link to="/home" class="brand-lockup" aria-label="Home">
          <span class="brand-mark">
            <img v-if="siteLogo" :src="siteLogo" :alt="siteName" class="h-full w-full object-contain" />
            <img v-else src="/logo.svg?v=cobalt-20260729" alt="筏&API" class="h-full w-full object-contain" />
          </span>
          <span class="min-w-0">
            <span class="brand-name">{{ siteName }}</span>
            <span class="brand-caption">UNIFIED AI ACCESS</span>
          </span>
        </router-link>

        <div class="nav-controls">
          <a
            v-if="docUrl"
            :href="docUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="landing-icon-control"
            :title="t('home.viewDocs')"
            :aria-label="t('home.viewDocs')"
          >
            <BookOpen :size="17" :stroke-width="1.9" aria-hidden="true" />
          </a>
          <LandingAppearancePicker v-model="palette" />
          <LandingLocaleSwitcher />
          <router-link :to="isAuthenticated ? dashboardPath : '/login'" class="landing-nav-action">
            <LayoutDashboard v-if="isAuthenticated" :size="16" :stroke-width="1.9" aria-hidden="true" />
            <LogIn v-else :size="16" :stroke-width="1.9" aria-hidden="true" />
            <span>{{ isAuthenticated ? t('home.dashboard') : t('home.login') }}</span>
          </router-link>
        </div>
      </nav>
    </header>

    <main class="flex-1">
      <section class="hero-section">
        <div class="landing-container hero-grid">
          <div class="hero-copy">
            <p class="eyebrow"><span>01</span> AI ACCESS LAYER</p>
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

            <div class="hero-context">
              <div class="hero-endpoint">
                <SquareTerminal :size="18" :stroke-width="1.8" aria-hidden="true" />
                <span><small>UNIFIED ENDPOINT</small><code>{{ apiEndpoint }}</code></span>
              </div>
              <div class="hero-ready"><CircleCheck :size="17" :stroke-width="2" aria-hidden="true" /><span>ROUTING READY</span></div>
            </div>
          </div>

          <aside class="route-workbench" aria-label="API routing overview">
            <div class="workbench-header">
              <div><span class="workbench-index">LIVE</span><strong>Routing desk</strong></div>
              <span class="workbench-status"><i></i>Online</span>
            </div>

            <div class="workbench-entry">
              <span class="entry-icon"><Route :size="18" :stroke-width="1.8" aria-hidden="true" /></span>
              <div class="min-w-0"><small>ONE REQUEST</small><code>POST {{ apiEndpoint }}/responses</code></div>
              <span class="entry-badge">1 key</span>
            </div>

            <div class="workbench-label"><span>ACTIVE ROUTES</span><span>{{ providerRoutes.length }} AVAILABLE</span></div>
            <div class="route-list">
              <div v-for="provider in providerRoutes" :key="provider.name" class="route-row">
                <span class="route-link" aria-hidden="true"></span>
                <span class="route-icon" :class="provider.tone"><component :is="provider.icon" :size="17" :stroke-width="1.8" aria-hidden="true" /></span>
                <div class="route-copy"><strong>{{ provider.name }}</strong><span>{{ provider.mode }} routing</span></div>
                <div class="route-metric"><strong>{{ provider.latency }}</strong><span>{{ provider.load }} load</span></div>
                <CircleCheck :size="17" :stroke-width="2" class="route-check" aria-hidden="true" />
              </div>
            </div>

            <div class="workbench-summary">
              <div><span>SUCCESS RATE</span><strong>99.98%</strong></div>
              <div><span>POLICY</span><strong>Adaptive</strong></div>
              <div><span>USAGE</span><strong>Realtime</strong></div>
            </div>
          </aside>
        </div>
      </section>

      <section class="signal-strip" aria-label="Platform capabilities">
        <div class="landing-container signal-grid">
          <div v-for="signal in signals" :key="signal.label" class="signal-item">
            <span class="signal-number">{{ signal.number }}</span>
            <component :is="signal.icon" :size="19" :stroke-width="1.8" aria-hidden="true" />
            <strong>{{ signal.label }}</strong>
            <span class="signal-note">{{ signal.note }}</span>
          </div>
        </div>
      </section>

      <section class="provider-section">
        <div class="landing-container">
          <div class="section-heading">
            <div>
              <p class="eyebrow"><span>02</span> PROVIDER NETWORK</p>
              <h2>{{ t('home.providers.title') }}</h2>
            </div>
            <p>{{ t('home.providers.description') }}</p>
          </div>

          <div class="provider-board">
            <div class="provider-board-top">
              <div class="provider-origin"><KeyRound :size="18" :stroke-width="1.8" aria-hidden="true" /><span><small>ONE KEY / EVERY ROUTE</small><code>{{ apiEndpoint }}</code></span></div>
              <span class="provider-board-note"><CircleCheck :size="16" :stroke-width="2" aria-hidden="true" />policy applied per request</span>
            </div>
            <div class="provider-table-heading"><span>ROUTE</span><span>POLICY</span><span>LATENCY</span><span>STATE</span></div>
            <div v-for="(provider, index) in providerRoutes" :key="`network-${provider.name}`" class="provider-table-row">
              <div class="provider-identity"><span class="provider-number">0{{ index + 1 }}</span><span class="route-icon" :class="provider.tone"><component :is="provider.icon" :size="18" :stroke-width="1.8" aria-hidden="true" /></span><strong>{{ provider.name }}</strong></div>
              <span class="provider-mode">{{ provider.mode }}</span>
              <span class="provider-latency">{{ provider.latency }}</span>
              <span class="provider-health"><i></i>{{ t('home.providers.supported') }}</span>
            </div>
          </div>
        </div>
      </section>

      <section class="operation-section">
        <div class="landing-container">
          <div class="section-heading section-heading-compact">
            <div>
              <p class="eyebrow"><span>03</span> CONTROL, NOT COMPLEXITY</p>
              <h2>{{ t('home.solutions.title') }}</h2>
            </div>
            <p>{{ t('home.solutions.subtitle') }}</p>
          </div>

          <div class="operation-grid">
            <article v-for="feature in features" :key="feature.title" class="operation-item">
              <div class="operation-top"><span>{{ feature.number }}</span><component :is="feature.icon" :size="22" :stroke-width="1.8" aria-hidden="true" /></div>
              <h3>{{ feature.title }}</h3>
              <p>{{ feature.description }}</p>
              <span class="operation-rule"></span>
            </article>
          </div>
        </div>
      </section>

      <section class="conversion-section">
        <div class="landing-container conversion-inner">
          <div><p class="eyebrow"><span>04</span> READY TO CONNECT</p><h2>把 AI 接入，变成一条稳定的路由。</h2></div>
          <div class="conversion-action"><p>{{ siteSubtitle }}</p><router-link :to="isAuthenticated ? dashboardPath : '/login'" class="landing-primary-action"><span>{{ isAuthenticated ? t('home.goToDashboard') : t('home.getStarted') }}</span><ArrowUpRight :size="17" :stroke-width="2" aria-hidden="true" /></router-link></div>
        </div>
      </section>
    </main>

    <footer class="landing-footer">
      <div class="landing-container landing-footer-inner">
        <p>&copy; {{ currentYear }} {{ siteName }}. {{ t('home.footer.allRightsReserved') }}</p>
        <div><a v-if="docUrl" :href="docUrl" target="_blank" rel="noopener noreferrer">{{ t('home.docs') }}</a><a :href="githubUrl" target="_blank" rel="noopener noreferrer">GitHub</a></div>
      </div>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  ArrowRight, ArrowUpRight, BookOpen, Bot, CircleCheck, Cloud, Gauge, KeyRound,
  LayoutDashboard, LogIn, Orbit, ReceiptText, Route, ShieldCheck, Sparkles,
  SquareTerminal, Users
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
const apiEndpoint = computed(() => `${window.location.origin}/v1`)
const githubUrl = 'https://github.com/Wei-Shaw/sub2api'
const palette = ref<LandingPalette>('cobalt')

const providerRoutes = computed(() => [
  { name: t('home.providers.claude'), icon: Bot, tone: 'tone-coral', latency: '112 ms', load: '68%', mode: 'Priority' },
  { name: 'GPT', icon: Sparkles, tone: 'tone-gold', latency: '128 ms', load: '82%', mode: 'Balanced' },
  { name: t('home.providers.gemini'), icon: Orbit, tone: 'tone-blue', latency: '96 ms', load: '54%', mode: 'Fastest' },
  { name: t('home.providers.antigravity'), icon: Cloud, tone: 'tone-violet', latency: '141 ms', load: '44%', mode: 'Standby' }
])

const signals = computed(() => [
  { number: '01', icon: Route, label: t('home.tags.subscriptionToApi'), note: 'Single endpoint' },
  { number: '02', icon: ShieldCheck, label: t('home.tags.stickySession'), note: 'Policy aware' },
  { number: '03', icon: ReceiptText, label: t('home.tags.realtimeBilling'), note: 'Usage visible' }
])

const features = computed(() => [
  { number: '01', icon: Route, title: t('home.features.unifiedGateway'), description: t('home.features.unifiedGatewayDesc') },
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
  --accent: #2856d8;
  --accent-strong: #173fae;
  --accent-soft: #edf2ff;
  --accent-border: #cbd8ff;
  --ink: #131722;
  --muted: #697386;
  --line: #e3e7ed;
  --surface: #ffffff;
  --canvas: #f7f8fa;
  min-width: 0;
  background: var(--canvas);
  color: var(--ink);
}

.landing-page[data-palette='graphite'] {
  --accent: #252932;
  --accent-strong: #111318;
  --accent-soft: #f0f2f4;
  --accent-border: #d3d7dd;
}

.landing-page[data-palette='vermilion'] {
  --accent: #c9422e;
  --accent-strong: #a92f20;
  --accent-soft: #fff0ed;
  --accent-border: #f1c9c1;
}

.landing-container {
  width: min(100% - 48px, 1240px);
  margin-inline: auto;
}

.landing-header {
  position: sticky;
  top: 0;
  z-index: 40;
  border-bottom: 1px solid var(--line);
  background: rgba(255, 255, 255, 0.94);
  backdrop-filter: blur(14px);
}

.landing-nav {
  display: flex;
  min-height: 72px;
  align-items: center;
  justify-content: space-between;
  gap: 20px;
}

.brand-lockup,
.nav-controls,
.hero-actions,
.hero-context,
.workbench-header,
.workbench-entry,
.workbench-summary,
.provider-board-top,
.provider-identity,
.landing-footer-inner,
.landing-footer-inner div {
  display: flex;
  align-items: center;
}

.brand-lockup {
  min-width: 0;
  gap: 10px;
}

.brand-mark {
  display: inline-flex;
  width: 38px;
  height: 38px;
  flex: none;
  align-items: center;
  justify-content: center;
}

.brand-name {
  display: block;
  overflow: hidden;
  color: var(--ink);
  font-size: 15px;
  font-weight: 750;
  line-height: 1.2;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.brand-caption {
  display: block;
  margin-top: 3px;
  color: var(--muted);
  font-size: 9px;
  font-weight: 750;
  line-height: 1;
}

.nav-controls {
  flex: none;
  gap: 8px;
}

.landing-icon-control,
.landing-nav-action,
.landing-primary-action,
.landing-secondary-action {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border-radius: 6px;
  font-size: 13px;
  font-weight: 700;
  transition: background-color 160ms ease, border-color 160ms ease, color 160ms ease, transform 160ms ease;
}

.landing-icon-control {
  width: 38px;
  height: 38px;
  border: 1px solid var(--line);
  background: var(--surface);
  color: #596174;
}

.landing-icon-control:hover,
.landing-secondary-action:hover {
  border-color: #b8c0cd;
  background: #f8f9fb;
  color: var(--ink);
}

.landing-nav-action {
  min-height: 38px;
  gap: 7px;
  padding-inline: 14px;
  background: var(--accent);
  color: #fff;
}

.landing-nav-action:hover,
.landing-primary-action:hover {
  background: var(--accent-strong);
}

.hero-section {
  border-bottom: 1px solid var(--line);
  background: var(--surface);
}

.hero-grid {
  display: grid;
  min-height: min(680px, calc(100svh - 190px));
  grid-template-columns: minmax(0, 0.91fr) minmax(520px, 1.09fr);
  align-items: center;
  gap: 76px;
  padding-block: 76px;
}

.hero-copy {
  min-width: 0;
  max-width: 520px;
}

.eyebrow {
  display: flex;
  align-items: center;
  gap: 10px;
  color: var(--muted);
  font-size: 10px;
  font-weight: 800;
  line-height: 1;
}

.eyebrow span {
  color: var(--accent);
}

.hero-copy h1 {
  margin-top: 26px;
  color: var(--ink);
  font-size: 62px;
  font-weight: 760;
  line-height: 1.04;
}

.hero-lead {
  max-width: 490px;
  margin-top: 22px;
  color: #242b38;
  font-size: 25px;
  font-weight: 680;
  line-height: 1.42;
}

.hero-description {
  max-width: 500px;
  margin-top: 12px;
  color: var(--muted);
  font-size: 15px;
  line-height: 1.8;
}

.hero-actions {
  flex-wrap: wrap;
  gap: 10px;
  margin-top: 30px;
}

.landing-primary-action,
.landing-secondary-action {
  min-height: 46px;
  gap: 8px;
  padding-inline: 18px;
}

.landing-primary-action {
  background: var(--accent);
  color: #fff;
}

.landing-primary-action:hover {
  transform: translateY(-1px);
}

.landing-secondary-action {
  border: 1px solid var(--line);
  background: var(--surface);
  color: #354052;
}

.hero-context {
  min-height: 62px;
  justify-content: space-between;
  gap: 18px;
  margin-top: 36px;
  border-top: 1px solid var(--line);
  border-bottom: 1px solid var(--line);
}

.hero-endpoint,
.hero-ready,
.provider-origin,
.provider-board-note {
  display: flex;
  align-items: center;
}

.hero-endpoint {
  min-width: 0;
  gap: 10px;
  color: var(--accent);
}

.hero-endpoint span,
.hero-endpoint small,
.hero-endpoint code,
.provider-origin span,
.provider-origin small,
.provider-origin code {
  min-width: 0;
}

.hero-endpoint small,
.provider-origin small {
  display: block;
  color: var(--muted);
  font-size: 9px;
  font-weight: 750;
  line-height: 1;
}

.hero-endpoint code,
.provider-origin code {
  display: block;
  overflow: hidden;
  margin-top: 5px;
  color: #2c3544;
  font-size: 12px;
  font-weight: 700;
  line-height: 1;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.hero-ready {
  flex: none;
  gap: 6px;
  color: var(--accent);
  font-size: 10px;
  font-weight: 800;
}

.route-workbench {
  min-width: 0;
  overflow: hidden;
  border: 1px solid #dce2ea;
  border-top: 3px solid var(--accent);
  border-radius: 8px;
  background: #fff;
  box-shadow: 0 24px 60px rgba(25, 34, 52, 0.09);
}

.workbench-header {
  min-height: 68px;
  justify-content: space-between;
  gap: 16px;
  padding-inline: 22px;
  border-bottom: 1px solid var(--line);
}

.workbench-header > div {
  display: flex;
  align-items: center;
  gap: 10px;
}

.workbench-index {
  display: inline-flex;
  min-width: 34px;
  min-height: 22px;
  align-items: center;
  justify-content: center;
  border: 1px solid var(--accent-border);
  background: var(--accent-soft);
  color: var(--accent);
  font-size: 9px;
  font-weight: 800;
}

.workbench-header strong {
  color: #242c3a;
  font-size: 14px;
  font-weight: 750;
}

.workbench-status {
  display: inline-flex;
  align-items: center;
  gap: 7px;
  color: #4b576c;
  font-size: 10px;
  font-weight: 750;
}

.workbench-status i,
.provider-health i {
  width: 7px;
  height: 7px;
  flex: none;
  border-radius: 999px;
  background: var(--accent);
}

.workbench-entry {
  min-height: 82px;
  gap: 12px;
  padding: 17px 22px;
  border-bottom: 1px solid var(--line);
  background: #fbfcff;
}

.entry-icon {
  display: inline-flex;
  width: 34px;
  height: 34px;
  flex: none;
  align-items: center;
  justify-content: center;
  border: 1px solid var(--accent-border);
  border-radius: 5px;
  background: var(--accent-soft);
  color: var(--accent);
}

.workbench-entry div {
  min-width: 0;
  flex: 1;
}

.workbench-entry small,
.workbench-label,
.workbench-summary span,
.route-copy span,
.route-metric span,
.provider-table-heading,
.provider-mode,
.provider-latency {
  color: var(--muted);
  font-size: 9px;
  font-weight: 750;
}

.workbench-entry small {
  display: block;
}

.workbench-entry code {
  display: block;
  overflow: hidden;
  margin-top: 5px;
  color: #283244;
  font-size: 12px;
  font-weight: 700;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.entry-badge {
  flex: none;
  color: var(--accent);
  font-size: 10px;
  font-weight: 800;
}

.workbench-label {
  display: flex;
  justify-content: space-between;
  padding: 17px 22px 10px;
}

.route-list {
  padding: 0 14px 12px 22px;
}

.route-row {
  position: relative;
  display: flex;
  min-height: 55px;
  align-items: center;
  gap: 10px;
  padding: 8px 8px 8px 0;
}

.route-row + .route-row {
  border-top: 1px solid #edf0f4;
}

.route-link {
  width: 12px;
  height: 1px;
  flex: none;
  background: #c9d2e2;
}

.route-icon {
  display: inline-flex;
  width: 32px;
  height: 32px;
  flex: none;
  align-items: center;
  justify-content: center;
  border-radius: 5px;
}

.tone-coral { background: #fff1ee; color: #bd4a35; }
.tone-gold { background: #fff7df; color: #a06d0c; }
.tone-blue { background: #eef4ff; color: #2a5cc1; }
.tone-violet { background: #f4efff; color: #7044b4; }

.route-copy {
  min-width: 0;
  flex: 1;
}

.route-copy strong,
.route-copy span,
.route-metric strong,
.route-metric span {
  display: block;
}

.route-copy strong {
  overflow: hidden;
  color: #273042;
  font-size: 12px;
  font-weight: 750;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.route-copy span,
.route-metric span {
  margin-top: 3px;
  font-size: 9px;
}

.route-metric {
  flex: none;
  text-align: right;
}

.route-metric strong {
  color: #4a5669;
  font-size: 10px;
  font-weight: 750;
}

.route-check {
  flex: none;
  color: var(--accent);
}

.workbench-summary {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  border-top: 1px solid var(--line);
  background: #fbfcff;
}

.workbench-summary div {
  min-width: 0;
  padding: 15px 18px;
}

.workbench-summary div + div {
  border-left: 1px solid var(--line);
}

.workbench-summary span,
.workbench-summary strong {
  display: block;
}

.workbench-summary strong {
  overflow: hidden;
  margin-top: 6px;
  color: #263044;
  font-size: 12px;
  font-weight: 760;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.signal-strip {
  border-bottom: 1px solid var(--line);
  background: #f9fafc;
}

.signal-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
}

.signal-item {
  display: grid;
  min-height: 118px;
  grid-template-columns: 26px 24px minmax(0, 1fr);
  grid-template-rows: min-content min-content;
  align-content: center;
  column-gap: 12px;
  row-gap: 7px;
  padding: 22px 24px;
  color: var(--accent);
}

.signal-item + .signal-item {
  border-left: 1px solid var(--line);
}

.signal-number {
  grid-row: span 2;
  color: #9098a8;
  font-size: 10px;
  font-weight: 800;
}

.signal-item strong {
  color: #253045;
  font-size: 13px;
  font-weight: 750;
}

.signal-note {
  grid-column: 3;
  color: var(--muted);
  font-size: 11px;
}

.provider-section,
.operation-section {
  padding-block: 112px;
}

.provider-section { background: #fff; }
.operation-section { border-top: 1px solid var(--line); background: #f8f9fb; }

.section-heading {
  display: flex;
  align-items: end;
  justify-content: space-between;
  gap: 48px;
}

.section-heading > div { max-width: 600px; }

.section-heading h2,
.conversion-inner h2 {
  margin-top: 18px;
  color: var(--ink);
  font-size: 42px;
  font-weight: 750;
  line-height: 1.16;
}

.section-heading > p {
  max-width: 370px;
  margin-bottom: 4px;
  color: var(--muted);
  font-size: 15px;
  line-height: 1.75;
}

.provider-board {
  overflow: hidden;
  margin-top: 48px;
  border: 1px solid #dce2ea;
  border-radius: 8px;
  background: #fff;
}

.provider-board-top {
  min-height: 82px;
  justify-content: space-between;
  gap: 24px;
  padding-inline: 24px;
  border-bottom: 1px solid var(--line);
  background: #fbfcff;
}

.provider-origin {
  min-width: 0;
  gap: 11px;
  color: var(--accent);
}

.provider-board-note {
  flex: none;
  gap: 7px;
  color: var(--accent);
  font-size: 11px;
  font-weight: 750;
}

.provider-table-heading,
.provider-table-row {
  display: grid;
  grid-template-columns: minmax(0, 1.7fr) 1fr 0.7fr 0.7fr;
  align-items: center;
  gap: 20px;
  padding-inline: 24px;
}

.provider-table-heading {
  min-height: 46px;
  border-bottom: 1px solid var(--line);
  background: #f7f9fc;
}

.provider-table-row {
  min-height: 74px;
}

.provider-table-row + .provider-table-row { border-top: 1px solid #edf0f4; }

.provider-identity { min-width: 0; gap: 11px; }

.provider-number {
  width: 24px;
  flex: none;
  color: #9aa2b0;
  font-size: 10px;
  font-weight: 800;
}

.provider-identity strong {
  overflow: hidden;
  color: #293346;
  font-size: 13px;
  font-weight: 750;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.provider-mode,
.provider-latency { color: #556177; font-size: 12px; }

.provider-health {
  display: inline-flex;
  align-items: center;
  justify-self: start;
  gap: 7px;
  color: var(--accent);
  font-size: 11px;
  font-weight: 750;
}

.section-heading-compact { align-items: end; }

.operation-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  margin-top: 52px;
  border-top: 1px solid #cfd6e0;
  border-bottom: 1px solid #cfd6e0;
}

.operation-item {
  position: relative;
  min-height: 330px;
  padding: 28px 28px 32px;
  background: transparent;
}

.operation-item + .operation-item { border-left: 1px solid #cfd6e0; }

.operation-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  color: var(--accent);
}

.operation-top span {
  color: #929cab;
  font-size: 10px;
  font-weight: 800;
}

.operation-item h3 {
  margin-top: 78px;
  color: #202a3a;
  font-size: 21px;
  font-weight: 750;
}

.operation-item p {
  max-width: 315px;
  margin-top: 13px;
  color: var(--muted);
  font-size: 14px;
  line-height: 1.8;
}

.operation-rule {
  position: absolute;
  right: 28px;
  bottom: 30px;
  left: 28px;
  height: 2px;
  background: var(--accent);
  transform: scaleX(0.18);
  transform-origin: left;
  transition: transform 180ms ease;
}

.operation-item:hover .operation-rule { transform: scaleX(1); }

.conversion-section {
  border-top: 1px solid var(--line);
  border-bottom: 1px solid var(--line);
  background: #fff;
}

.conversion-inner {
  display: flex;
  min-height: 300px;
  align-items: center;
  justify-content: space-between;
  gap: 64px;
}

.conversion-inner > div:first-child { max-width: 660px; }

.conversion-action {
  display: flex;
  width: min(100%, 310px);
  flex: none;
  align-items: flex-start;
  flex-direction: column;
  gap: 24px;
}

.conversion-action > p {
  color: var(--muted);
  font-size: 14px;
  line-height: 1.75;
}

.landing-footer { background: #f8f9fb; }

.landing-footer-inner {
  min-height: 82px;
  justify-content: space-between;
  gap: 20px;
  color: #7a8493;
  font-size: 12px;
}

.landing-footer-inner div { gap: 18px; }
.landing-footer-inner a:hover { color: var(--ink); }

@media (max-width: 1120px) {
  .hero-grid { grid-template-columns: minmax(0, 0.84fr) minmax(460px, 1.16fr); gap: 48px; }
  .hero-copy h1 { font-size: 56px; }
}

@media (max-width: 960px) {
  .hero-grid { min-height: 0; grid-template-columns: 1fr; gap: 52px; padding-block: 64px; }
  .hero-copy { max-width: 650px; }
  .route-workbench { width: min(100%, 680px); }
  .section-heading { display: block; }
  .section-heading > p { max-width: 540px; margin-top: 16px; }
  .conversion-inner { min-height: 0; align-items: flex-start; flex-direction: column; padding-block: 74px; }
  .conversion-action { width: min(100%, 520px); }
}

@media (max-width: 767px) {
  .landing-container { width: min(100% - 28px, 1240px); }
  .landing-nav { min-height: 64px; gap: 12px; }
  .brand-caption { display: none; }
  .hero-grid { gap: 38px; padding-block: 50px 44px; }
  .hero-copy h1 { margin-top: 20px; font-size: 48px; }
  .hero-lead { margin-top: 18px; font-size: 21px; }
  .hero-description { font-size: 14px; }
  .hero-context { align-items: flex-start; flex-direction: column; gap: 13px; padding-block: 15px; }
  .signal-grid { grid-template-columns: 1fr; }
  .signal-item + .signal-item { border-top: 1px solid var(--line); border-left: 0; }
  .provider-section, .operation-section { padding-block: 78px; }
  .section-heading h2, .conversion-inner h2 { font-size: 34px; }
  .provider-board { margin-top: 36px; }
  .provider-board-top { align-items: flex-start; flex-direction: column; gap: 12px; padding: 18px; }
  .provider-table-heading, .provider-table-row { grid-template-columns: minmax(0, 1fr) 94px; gap: 10px; padding-inline: 18px; }
  .provider-table-heading span:nth-child(2), .provider-table-heading span:nth-child(3), .provider-mode, .provider-latency { display: none; }
  .operation-grid { grid-template-columns: 1fr; margin-top: 38px; }
  .operation-item { min-height: 250px; padding: 26px 20px 30px; }
  .operation-item + .operation-item { border-top: 1px solid #cfd6e0; border-left: 0; }
  .operation-item h3 { margin-top: 46px; }
  .operation-rule { right: 20px; bottom: 26px; left: 20px; }
}

@media (max-width: 520px) {
  .landing-icon-control { display: none; }
  .landing-nav-action { width: 38px; padding: 0; }
  .landing-nav-action span { display: none; }
  .hero-actions { display: grid; grid-template-columns: 1fr; }
  .hero-actions > * { width: 100%; }
  .route-workbench { margin-inline: -2px; }
  .workbench-header, .workbench-entry { padding-inline: 16px; }
  .workbench-label { padding-inline: 16px; }
  .route-list { padding-right: 8px; padding-left: 16px; }
  .workbench-summary div { padding: 13px 10px; }
  .workbench-summary strong { font-size: 10px; }
  .signal-item { padding-inline: 8px; }
  .provider-origin code { max-width: 210px; }
  .provider-board-note { font-size: 10px; }
  .landing-footer-inner { align-items: flex-start; flex-direction: column; padding-block: 24px; }
}

@media (max-width: 380px) {
  .brand-mark { width: 34px; height: 34px; }
  .brand-name { max-width: 84px; font-size: 14px; }
  .route-metric { display: none; }
  .provider-number { display: none; }
}
</style>