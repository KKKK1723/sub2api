<template>
  <div v-if="homeContent" class="min-h-screen bg-white">
    <iframe v-if="isHomeContentUrl" :src="homeContent.trim()" class="h-screen w-full border-0" allowfullscreen></iframe>
    <div v-else v-html="homeContent"></div>
  </div>

  <div v-else class="landing-page flex min-h-screen flex-col text-[#101828]" :data-palette="palette">
    <header class="landing-header">
      <nav class="landing-container flex h-[68px] items-center justify-between gap-3">
        <router-link to="/home" class="flex min-w-0 items-center gap-2.5" aria-label="Home">
          <span class="brand-mark">
            <img v-if="siteLogo" :src="siteLogo" :alt="siteName" class="h-full w-full object-contain" />
            <Network v-else :size="21" :stroke-width="1.9" aria-hidden="true" />
          </span>
          <span class="min-w-0">
            <span class="block truncate text-[15px] font-semibold text-[#101828] sm:text-base">{{ siteName }}</span>
            <span class="hidden items-center gap-1.5 text-[10px] font-medium text-[#667085] sm:flex">
              <span class="status-dot"></span>GATEWAY ONLINE
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
              <Activity :size="16" :stroke-width="2" aria-hidden="true" />
              <span>AI ACCESS CONTROL PLANE</span>
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

            <div class="hero-metrics">
              <div v-for="metric in metrics" :key="metric.label" class="hero-metric">
                <component :is="metric.icon" :size="17" :stroke-width="1.9" aria-hidden="true" />
                <div><strong>{{ metric.value }}</strong><span>{{ metric.label }}</span></div>
              </div>
            </div>
          </div>

          <div class="gateway-console" aria-label="Live API gateway overview">
            <div class="console-toolbar">
              <div class="flex items-center gap-2">
                <SquareTerminal :size="17" :stroke-width="1.8" aria-hidden="true" />
                <span>Gateway pulse</span>
              </div>
              <span class="console-live"><span></span>LIVE</span>
            </div>

            <div class="request-summary">
              <div class="request-method">POST</div>
              <div class="min-w-0 flex-1">
                <p>Unified endpoint</p>
                <code>{{ apiEndpoint }}/responses</code>
              </div>
              <div class="request-result"><span>200</span><small>128 ms</small></div>
            </div>

            <div class="console-section-heading"><span>ACTIVE ROUTES</span><span>4 providers</span></div>
            <div class="route-list">
              <div v-for="provider in providerRoutes" :key="provider.name" class="route-row">
                <span class="route-icon" :class="provider.tone">
                  <component :is="provider.icon" :size="17" :stroke-width="1.8" aria-hidden="true" />
                </span>
                <div class="min-w-0 flex-1">
                  <div class="flex items-center justify-between gap-3">
                    <strong>{{ provider.name }}</strong><span>{{ provider.latency }}</span>
                  </div>
                  <div class="route-progress"><span :style="{ width: provider.load }"></span></div>
                </div>
                <CircleCheck :size="17" :stroke-width="2" class="route-check" aria-hidden="true" />
              </div>
            </div>

            <div class="console-footer">
              <div><Gauge :size="16" :stroke-width="1.9" aria-hidden="true" /><span>99.98% success</span></div>
              <div><ReceiptText :size="16" :stroke-width="1.9" aria-hidden="true" /><span>Metered billing</span></div>
              <div><ShieldCheck :size="16" :stroke-width="1.9" aria-hidden="true" /><span>Session safe</span></div>
            </div>
          </div>
        </div>
      </section>

      <section class="capability-band">
        <div class="landing-container capability-grid">
          <div v-for="item in capabilities" :key="item.label" class="capability-item">
            <component :is="item.icon" :size="20" :stroke-width="1.8" aria-hidden="true" />
            <span>{{ item.label }}</span>
            <ChevronRight :size="16" :stroke-width="1.8" aria-hidden="true" />
          </div>
        </div>
      </section>

      <section class="provider-section">
        <div class="landing-container provider-layout">
          <div class="section-copy">
            <p class="section-label">ROUTING FABRIC</p>
            <h2>{{ t('home.providers.title') }}</h2>
            <p>{{ t('home.providers.description') }}</p>
            <div class="endpoint-note">
              <KeyRound :size="18" :stroke-width="1.9" aria-hidden="true" />
              <div><span>ONE API KEY</span><code>{{ apiEndpoint }}</code></div>
            </div>
          </div>

          <div class="provider-table">
            <div class="provider-table-heading"><span>PROVIDER</span><span>ROUTING</span><span>STATUS</span></div>
            <div v-for="provider in providerRoutes" :key="'table-' + provider.name" class="provider-table-row">
              <div class="flex min-w-0 items-center gap-3">
                <span class="route-icon" :class="provider.tone">
                  <component :is="provider.icon" :size="18" :stroke-width="1.8" aria-hidden="true" />
                </span>
                <strong>{{ provider.name }}</strong>
              </div>
              <span class="provider-mode">{{ provider.mode }}</span>
              <span class="provider-health"><span></span>{{ t('home.providers.supported') }}</span>
            </div>
          </div>
        </div>
      </section>

      <section class="feature-section">
        <div class="landing-container">
          <div class="feature-heading">
            <div><p class="section-label">OPERATIONS LAYER</p><h2>{{ t('home.solutions.title') }}</h2></div>
            <p>{{ t('home.solutions.subtitle') }}</p>
          </div>
          <div class="feature-grid">
            <article v-for="feature in features" :key="feature.title" class="feature-card">
              <div class="feature-card-top">
                <span class="feature-number">{{ feature.number }}</span>
                <component :is="feature.icon" :size="22" :stroke-width="1.8" aria-hidden="true" />
              </div>
              <h3>{{ feature.title }}</h3>
              <p>{{ feature.description }}</p>
            </article>
          </div>
        </div>
      </section>

      <section class="final-cta">
        <div class="landing-container final-cta-inner">
          <div class="min-w-0"><p>{{ siteName }}</p><span>{{ siteSubtitle }}</span></div>
          <router-link :to="isAuthenticated ? dashboardPath : '/login'" class="landing-primary-action">
            <span>{{ isAuthenticated ? t('home.goToDashboard') : t('home.getStarted') }}</span>
            <ArrowRight :size="17" :stroke-width="2" aria-hidden="true" />
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
  Activity, ArrowRight, BookOpen, Bot, ChevronRight, CircleCheck, CircleDollarSign,
  Cloud, Gauge, KeyRound, LayoutDashboard, LogIn, Network, Orbit, ReceiptText,
  Route, ShieldCheck, Sparkles, SquareTerminal, Users, Workflow
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
const palette = ref<LandingPalette>('mint')

const metrics = computed(() => [
  { icon: Workflow, value: '04', label: t('home.tags.subscriptionToApi') },
  { icon: ShieldCheck, value: '24/7', label: t('home.tags.stickySession') },
  { icon: CircleDollarSign, value: 'LIVE', label: t('home.tags.realtimeBilling') }
])

const capabilities = computed(() => [
  { icon: Route, label: t('home.tags.subscriptionToApi') },
  { icon: ShieldCheck, label: t('home.tags.stickySession') },
  { icon: ReceiptText, label: t('home.tags.realtimeBilling') }
])

const providerRoutes = computed(() => [
  { name: t('home.providers.claude'), icon: Bot, tone: 'tone-coral', latency: '112 ms', load: '68%', mode: 'Priority' },
  { name: 'GPT', icon: Sparkles, tone: 'tone-green', latency: '128 ms', load: '82%', mode: 'Balanced' },
  { name: t('home.providers.gemini'), icon: Orbit, tone: 'tone-blue', latency: '96 ms', load: '54%', mode: 'Fastest' },
  { name: t('home.providers.antigravity'), icon: Cloud, tone: 'tone-violet', latency: '141 ms', load: '44%', mode: 'Standby' }
])

const features = computed(() => [
  { number: '01', icon: Network, title: t('home.features.unifiedGateway'), description: t('home.features.unifiedGatewayDesc') },
  { number: '02', icon: Users, title: t('home.features.multiAccount'), description: t('home.features.multiAccountDesc') },
  { number: '03', icon: Gauge, title: t('home.features.balanceQuota'), description: t('home.features.balanceQuotaDesc') }
])

function isLandingPalette(value: string | null): value is LandingPalette {
  return value === 'mint' || value === 'sky' || value === 'coral'
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
  --accent: #0f766e;
  --accent-strong: #0b5f59;
  --accent-soft: #e7f5f2;
  --accent-border: #b9ddd5;
  background: #f5f7f9;
}
.landing-page[data-palette='sky'] {
  --accent: #2563eb;
  --accent-strong: #1d4ed8;
  --accent-soft: #eaf1ff;
  --accent-border: #bfd1ff;
}
.landing-page[data-palette='coral'] {
  --accent: #c2410c;
  --accent-strong: #9a3412;
  --accent-soft: #fff0e8;
  --accent-border: #f2c6b1;
}
.landing-container {
  width: min(100% - 32px, 1200px);
  margin-inline: auto;
}
.landing-header {
  position: sticky;
  top: 0;
  z-index: 40;
  border-bottom: 1px solid #e4e7ec;
  background: rgba(255, 255, 255, 0.96);
  backdrop-filter: blur(12px);
}
.brand-mark {
  display: inline-flex;
  width: 38px;
  height: 38px;
  flex: none;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  border: 1px solid var(--accent-border);
  border-radius: 8px;
  background: var(--accent-soft);
  color: var(--accent);
}
.status-dot {
  width: 6px;
  height: 6px;
  border-radius: 9999px;
  background: #12b76a;
  box-shadow: 0 0 0 2px #d1fadf;
}
.landing-icon-control,
.landing-nav-action,
.landing-primary-action,
.landing-secondary-action {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border-radius: 8px;
  font-weight: 600;
  transition: border-color 160ms ease, background-color 160ms ease, color 160ms ease;
}
.landing-icon-control {
  width: 40px;
  height: 40px;
  border: 1px solid #dfe3e8;
  background: #fff;
  color: #344054;
}
.landing-icon-control:hover {
  border-color: #b9c0ca;
  background: #f8fafb;
  color: #101828;
}
.landing-nav-action {
  min-height: 40px;
  gap: 7px;
  padding: 0 14px;
  background: #101828;
  color: #fff;
  font-size: 13px;
}
.landing-nav-action:hover {
  background: #293056;
}
.hero-section {
  border-bottom: 1px solid #e4e7ec;
  background: #fff;
}
.hero-layout {
  display: grid;
  min-height: 650px;
  grid-template-columns: minmax(0, 0.92fr) minmax(480px, 1.08fr);
  align-items: center;
  gap: 64px;
  padding-block: 56px;
}
.hero-copy {
  min-width: 0;
}
.hero-kicker {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  color: var(--accent);
  font-size: 11px;
  font-weight: 700;
}
.hero-copy h1 {
  max-width: 660px;
  margin-top: 20px;
  color: #101828;
  font-size: 58px;
  font-weight: 650;
  letter-spacing: 0;
  line-height: 1.06;
}
.hero-lead {
  margin-top: 22px;
  color: #1d2939;
  font-size: 21px;
  font-weight: 600;
  line-height: 1.45;
}
.hero-description {
  max-width: 590px;
  margin-top: 10px;
  color: #667085;
  font-size: 15px;
  line-height: 1.75;
}
.hero-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  margin-top: 28px;
}
.landing-primary-action,
.landing-secondary-action {
  min-height: 44px;
  gap: 8px;
  padding: 0 17px;
  font-size: 14px;
}
.landing-primary-action {
  background: var(--accent);
  color: #fff;
}
.landing-primary-action:hover {
  background: var(--accent-strong);
}
.landing-secondary-action {
  border: 1px solid #d0d5dd;
  background: #fff;
  color: #344054;
}
.landing-secondary-action:hover {
  border-color: #98a2b3;
  background: #f9fafb;
  color: #101828;
}
.hero-metrics {
  display: grid;
  max-width: 590px;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  margin-top: 38px;
  border-top: 1px solid #e4e7ec;
}
.hero-metric {
  display: flex;
  min-width: 0;
  align-items: flex-start;
  gap: 10px;
  padding: 17px 14px 0 0;
  color: var(--accent);
}
.hero-metric + .hero-metric {
  border-left: 1px solid #e4e7ec;
  padding-left: 16px;
}
.hero-metric strong,
.hero-metric span {
  display: block;
}
.hero-metric strong {
  color: #101828;
  font-size: 14px;
  line-height: 1.2;
}
.hero-metric span {
  margin-top: 3px;
  overflow: hidden;
  color: #667085;
  font-size: 11px;
  line-height: 1.35;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.gateway-console {
  min-width: 0;
  overflow: hidden;
  border: 1px solid #d6dae1;
  border-radius: 8px;
  background: #fff;
  box-shadow: 0 24px 60px rgba(16, 24, 40, 0.1);
}
.console-toolbar {
  display: flex;
  min-height: 48px;
  align-items: center;
  justify-content: space-between;
  border-bottom: 1px solid #e4e7ec;
  background: #f8fafb;
  padding: 0 16px;
  color: #344054;
  font-size: 12px;
  font-weight: 600;
}
.console-live {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  color: #027a48;
  font-size: 10px;
  font-weight: 700;
}
.console-live span {
  width: 7px;
  height: 7px;
  border-radius: 9999px;
  background: #12b76a;
  box-shadow: 0 0 0 3px #d1fadf;
}
.request-summary {
  display: flex;
  align-items: center;
  gap: 14px;
  border-bottom: 1px solid #e4e7ec;
  padding: 18px 20px;
}
.request-method {
  display: inline-flex;
  height: 28px;
  align-items: center;
  border-radius: 5px;
  background: var(--accent-soft);
  padding: 0 8px;
  color: var(--accent);
  font-size: 10px;
  font-weight: 800;
}
.request-summary p {
  color: #667085;
  font-size: 10px;
  font-weight: 600;
}
.request-summary code {
  display: block;
  overflow: hidden;
  margin-top: 3px;
  color: #101828;
  font-size: 12px;
  font-weight: 600;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.request-result {
  display: flex;
  flex: none;
  flex-direction: column;
  align-items: flex-end;
}
.request-result span {
  color: #027a48;
  font-size: 13px;
  font-weight: 700;
}
.request-result small {
  margin-top: 2px;
  color: #98a2b3;
  font-size: 10px;
}
.console-section-heading {
  display: grid;
  grid-template-columns: 1fr auto;
  padding: 16px 20px 8px;
  color: #98a2b3;
  font-size: 9px;
  font-weight: 700;
}
.route-list {
  padding: 0 12px 12px;
}
.route-row {
  display: flex;
  min-height: 58px;
  align-items: center;
  gap: 12px;
  border-bottom: 1px solid #edf0f3;
  padding: 0 8px;
}
.route-row:last-child {
  border-bottom: 0;
}
.route-icon {
  display: inline-flex;
  width: 34px;
  height: 34px;
  flex: none;
  align-items: center;
  justify-content: center;
  border-radius: 7px;
}
.tone-coral {
  background: #fff1ed;
  color: #c4320a;
}
.tone-green {
  background: #eaf8f2;
  color: #087a5b;
}
.tone-blue {
  background: #edf4ff;
  color: #175cd3;
}
.tone-violet {
  background: #f4f0ff;
  color: #6941c6;
}
.route-row strong {
  color: #344054;
  font-size: 12px;
}
.route-row span {
  color: #667085;
  font-size: 10px;
}
.route-progress {
  height: 3px;
  overflow: hidden;
  margin-top: 7px;
  border-radius: 9999px;
  background: #edf0f3;
}
.route-progress span {
  display: block;
  height: 100%;
  border-radius: inherit;
  background: var(--accent);
}
.route-check {
  flex: none;
  color: #12b76a;
}
.console-footer {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  border-top: 1px solid #e4e7ec;
  background: #f8fafb;
}
.console-footer div {
  display: flex;
  min-width: 0;
  align-items: center;
  justify-content: center;
  gap: 6px;
  padding: 13px 8px;
  color: #667085;
  font-size: 10px;
  white-space: nowrap;
}
.console-footer div + div {
  border-left: 1px solid #e4e7ec;
}
.capability-band {
  border-bottom: 1px solid #dfe3e8;
  background: #eef1f4;
}
.capability-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
}
.capability-item {
  display: grid;
  min-height: 72px;
  grid-template-columns: auto minmax(0, 1fr) auto;
  align-items: center;
  gap: 10px;
  padding: 0 24px;
  color: #344054;
  font-size: 13px;
  font-weight: 600;
}
.capability-item + .capability-item {
  border-left: 1px solid #d6dae1;
}
.capability-item > svg:first-child {
  color: var(--accent);
}
.capability-item > svg:last-child {
  color: #98a2b3;
}
.provider-section {
  border-bottom: 1px solid #e4e7ec;
  background: #fff;
  padding: 84px 0;
}
.provider-layout {
  display: grid;
  grid-template-columns: minmax(260px, 0.72fr) minmax(0, 1.28fr);
  align-items: start;
  gap: 72px;
}
.section-label {
  color: var(--accent);
  font-size: 11px;
  font-weight: 700;
}
.section-copy h2,
.feature-heading h2 {
  margin-top: 10px;
  color: #101828;
  font-size: 32px;
  font-weight: 650;
  letter-spacing: 0;
  line-height: 1.2;
}
.section-copy > p:last-of-type,
.feature-heading > p {
  margin-top: 14px;
  color: #667085;
  font-size: 14px;
  line-height: 1.7;
}
.endpoint-note {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  margin-top: 30px;
  border-top: 1px solid #e4e7ec;
  padding-top: 18px;
  color: var(--accent);
}
.endpoint-note span,
.endpoint-note code {
  display: block;
}
.endpoint-note span {
  color: #98a2b3;
  font-size: 9px;
  font-weight: 700;
}
.endpoint-note code {
  overflow: hidden;
  max-width: 280px;
  margin-top: 4px;
  color: #344054;
  font-size: 12px;
  font-weight: 600;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.provider-table {
  overflow: hidden;
  border: 1px solid #dfe3e8;
  border-radius: 8px;
}
.provider-table-heading,
.provider-table-row {
  display: grid;
  grid-template-columns: minmax(0, 1.2fr) minmax(90px, 0.7fr) minmax(110px, 0.65fr);
  align-items: center;
}
.provider-table-heading {
  min-height: 40px;
  border-bottom: 1px solid #e4e7ec;
  background: #f8fafb;
  padding: 0 18px;
  color: #98a2b3;
  font-size: 9px;
  font-weight: 700;
}
.provider-table-row {
  min-height: 68px;
  padding: 0 18px;
  color: #344054;
}
.provider-table-row + .provider-table-row {
  border-top: 1px solid #edf0f3;
}
.provider-table-row strong {
  overflow: hidden;
  font-size: 13px;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.provider-mode {
  color: #667085;
  font-size: 12px;
}
.provider-health {
  display: inline-flex;
  align-items: center;
  gap: 7px;
  color: #027a48;
  font-size: 11px;
  font-weight: 600;
}
.provider-health span {
  width: 7px;
  height: 7px;
  border-radius: 9999px;
  background: #12b76a;
}
.feature-section {
  background: #f5f7f9;
  padding: 82px 0 88px;
}
.feature-heading {
  display: flex;
  align-items: end;
  justify-content: space-between;
  gap: 32px;
}
.feature-heading > p {
  max-width: 390px;
  margin: 0;
}
.feature-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 16px;
  margin-top: 34px;
}
.feature-card {
  min-height: 246px;
  border: 1px solid #dfe3e8;
  border-radius: 8px;
  background: #fff;
  padding: 24px;
  transition: border-color 160ms ease, box-shadow 160ms ease, transform 160ms ease;
}
.feature-card:hover {
  transform: translateY(-2px);
  border-color: var(--accent-border);
  box-shadow: 0 14px 34px rgba(16, 24, 40, 0.07);
}
.feature-card-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  color: var(--accent);
}
.feature-number {
  color: #98a2b3;
  font-size: 11px;
  font-weight: 700;
}
.feature-card h3 {
  margin-top: 46px;
  color: #101828;
  font-size: 17px;
  font-weight: 650;
}
.feature-card p {
  margin-top: 10px;
  color: #667085;
  font-size: 13px;
  line-height: 1.7;
}
.final-cta {
  border-block: 1px solid var(--accent-border);
  background: var(--accent-soft);
}
.final-cta-inner {
  display: flex;
  min-height: 118px;
  align-items: center;
  justify-content: space-between;
  gap: 24px;
}
.final-cta-inner p {
  overflow: hidden;
  color: #101828;
  font-size: 18px;
  font-weight: 650;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.final-cta-inner span {
  display: block;
  margin-top: 4px;
  color: #667085;
  font-size: 13px;
}
.final-cta-inner .landing-primary-action span {
  margin: 0;
  color: #fff;
}
.landing-footer {
  background: #fff;
}
.landing-footer-inner {
  display: flex;
  min-height: 76px;
  align-items: center;
  justify-content: space-between;
  gap: 20px;
  color: #667085;
  font-size: 12px;
}
.landing-footer-inner div {
  display: flex;
  gap: 20px;
}
.landing-footer-inner a:hover {
  color: #101828;
}
@media (max-width: 1023px) {
  .hero-layout {
    grid-template-columns: 1fr;
    gap: 44px;
    padding-block: 58px;
  }
  .hero-copy {
    max-width: 700px;
  }
  .gateway-console {
    width: min(100%, 700px);
  }
  .provider-layout {
    grid-template-columns: 1fr;
    gap: 40px;
  }
  .section-copy {
    max-width: 620px;
  }
}
@media (max-width: 767px) {
  .landing-container {
    width: min(100% - 28px, 1200px);
  }
  .hero-layout {
    min-height: 0;
    padding-block: 46px 36px;
  }
  .hero-copy h1 {
    font-size: 42px;
  }
  .hero-lead {
    font-size: 18px;
  }
  .capability-grid {
    grid-template-columns: 1fr;
  }
  .capability-item {
    min-height: 62px;
    padding: 0 12px;
  }
  .capability-item + .capability-item {
    border-top: 1px solid #d6dae1;
    border-left: 0;
  }
  .provider-section,
  .feature-section {
    padding-block: 58px;
  }
  .feature-heading {
    display: block;
  }
  .feature-heading > p {
    margin-top: 12px;
  }
  .feature-grid {
    grid-template-columns: 1fr;
  }
  .feature-card {
    min-height: 210px;
  }
  .feature-card h3 {
    margin-top: 34px;
  }
  .final-cta-inner,
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
  .hero-layout {
    gap: 28px;
    padding-bottom: 28px;
  }
  .hero-metrics,
  .route-row:nth-child(n + 3),
  .console-footer {
    display: none;
  }
  .provider-table-heading,
  .provider-table-row {
    grid-template-columns: minmax(0, 1fr) 90px;
  }
  .provider-table-heading span:nth-child(2),
  .provider-mode {
    display: none;
  }
  .provider-table-row,
  .provider-table-heading {
    padding-inline: 12px;
  }
}
@media (max-width: 380px) {
  .landing-icon-control {
    display: none;
  }
  .hero-copy h1 {
    font-size: 38px;
  }
}
</style>
