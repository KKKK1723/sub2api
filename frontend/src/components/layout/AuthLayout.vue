<template>
  <div class="auth-page min-h-screen lg:grid lg:grid-cols-[minmax(360px,42%)_1fr]">
    <aside class="auth-brand-panel relative hidden overflow-hidden border-r border-gray-200 lg:flex lg:flex-col">
      <div class="auth-grid pointer-events-none absolute inset-0" aria-hidden="true"></div>
      <router-link to="/" class="relative z-10 flex items-center gap-3 p-8">
        <img :src="siteLogo || '/logo.svg'" :alt="siteName" class="h-10 w-10 rounded-lg border border-gray-200 bg-white object-contain" />
        <span class="text-lg font-semibold text-gray-950">{{ siteName }}</span>
      </router-link>

      <div class="relative z-10 my-auto px-8 pb-20 xl:px-14">
        <p class="text-xs font-semibold text-teal-700">AI API GATEWAY</p>
        <h1 class="mt-4 max-w-lg text-4xl font-semibold leading-tight text-gray-950">{{ siteName }}</h1>
        <p class="mt-4 max-w-md text-base leading-7 text-gray-600">{{ siteSubtitle }}</p>

        <div class="auth-route-map mt-10" aria-hidden="true">
          <div class="grid gap-2">
            <span class="auth-node"><i class="bg-[#fff1ed] text-orange-700">C</i>Claude</span>
            <span class="auth-node"><i class="bg-[#eaf8f1] text-emerald-700">G</i>GPT</span>
            <span class="auth-node"><i class="bg-[#edf4ff] text-blue-700">G</i>Gemini</span>
          </div>
          <div class="auth-connector"><span></span><span></span><span></span></div>
          <div class="auth-endpoint"><span class="h-2 w-2 rounded-full bg-emerald-500"></span><code>/v1</code></div>
        </div>
      </div>
    </aside>

    <main class="flex min-h-screen flex-col bg-[#f6f8f7]">
      <div class="flex items-center justify-between border-b border-gray-200 bg-white px-4 py-3 lg:hidden">
        <router-link to="/" class="flex min-w-0 items-center gap-3">
          <img :src="siteLogo || '/logo.svg'" :alt="siteName" class="h-9 w-9 rounded-lg border border-gray-200 bg-white object-contain" />
          <span class="truncate font-semibold text-gray-950">{{ siteName }}</span>
        </router-link>
      </div>

      <div class="flex flex-1 items-center justify-center p-4 sm:p-8">
        <div class="w-full max-w-[460px]">
          <div class="auth-card border border-gray-200 bg-white p-6 shadow-sm sm:p-8">
            <slot />
          </div>
          <div class="mt-5 text-center text-sm"><slot name="footer" /></div>
          <div class="mt-7 text-center text-xs text-gray-400">&copy; {{ currentYear }} {{ siteName }}</div>
        </div>
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useAppStore } from '@/stores'
import { sanitizeUrl } from '@/utils/url'

const appStore = useAppStore()
const siteName = computed(() => appStore.siteName || '筏&API')
const siteLogo = computed(() => sanitizeUrl(appStore.siteLogo || '', { allowRelative: true, allowDataUrl: true }))
const siteSubtitle = computed(() => appStore.cachedPublicSettings?.site_subtitle || 'AI API Gateway Platform')
const currentYear = computed(() => new Date().getFullYear())

onMounted(() => appStore.fetchPublicSettings())
</script>

<style scoped>
.auth-brand-panel { background: #edf5f2; }
.auth-grid {
  background-image: linear-gradient(#dfe9e6 1px, transparent 1px), linear-gradient(90deg, #dfe9e6 1px, transparent 1px);
  background-size: 52px 52px;
  opacity: .62;
}
.auth-card, .auth-route-map { border-radius: 8px; }
.auth-route-map { display: grid; grid-template-columns: minmax(110px,1fr) 72px minmax(96px,.8fr); overflow: hidden; border: 1px solid #d8e3e0; background: rgba(255,255,255,.82); box-shadow: 0 16px 38px rgba(31,41,55,.08); }
.auth-route-map > div { padding: 14px; }
.auth-node { display: flex; align-items: center; gap: 8px; border: 1px solid #e5e7eb; border-radius: 6px; background: white; padding: 7px 8px; color: #4b5563; font-size: .75rem; font-weight: 600; }
.auth-node i { display: inline-flex; width: 24px; height: 24px; align-items: center; justify-content: center; border-radius: 5px; font-style: normal; font-size: .625rem; }
.auth-connector { display: flex; flex-direction: column; justify-content: space-around; padding-left: 0 !important; padding-right: 0 !important; }
.auth-connector span { position: relative; height: 1px; background: #9fc4bb; }
.auth-connector span::after { position: absolute; top: -3px; right: -1px; width: 7px; height: 7px; border-radius: 9999px; background: #0f766e; content: ''; }
.auth-endpoint { display: flex; align-items: center; justify-content: center; gap: 8px; border-left: 1px solid #d8e3e0; background: white; color: #111827; font-size: .875rem; font-weight: 700; }
</style>
