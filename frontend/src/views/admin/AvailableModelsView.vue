<template>
  <AppLayout>
    <div class="available-models-page mx-auto max-w-7xl p-4 md:p-6">
      <header class="catalog-header">
        <div>
          <div class="catalog-eyebrow">Model Catalog</div>
          <h1 class="catalog-title">{{ t('admin.availableModels.title') }}</h1>
          <p class="catalog-intro">{{ t('admin.availableModels.description') }}</p>
        </div>
        <div class="catalog-actions">
          <input v-model="searchTerm" class="catalog-search" type="search" :placeholder="t('admin.availableModels.searchPlaceholder')" :aria-label="t('admin.availableModels.searchPlaceholder')" />
          <button class="btn btn-secondary catalog-refresh" :disabled="loading" :title="t('common.refresh')" @click="load"><Icon name="refresh" size="sm" :class="loading ? 'animate-spin' : ''" /><span>{{ t('common.refresh') }}</span></button>
        </div>
      </header>

      <div v-if="error" class="flex items-start gap-3 rounded-xl border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700 dark:border-red-900/40 dark:bg-red-900/20 dark:text-red-300"><Icon name="exclamationTriangle" size="sm" class="mt-0.5 shrink-0" /><span>{{ error }}</span></div>
      <div v-if="loading && groups.length === 0" class="rounded-xl border border-gray-200 bg-white py-16 text-center text-sm text-gray-500 dark:border-dark-600 dark:bg-dark-800 dark:text-gray-400"><Icon name="refresh" size="lg" class="mx-auto mb-3 animate-spin text-primary-500" /><p>{{ t('common.loading') }}</p></div>
      <div v-else-if="!loading && filteredGroups.length === 0" class="rounded-xl border border-dashed border-gray-300 bg-white py-16 text-center dark:border-dark-600 dark:bg-dark-800"><Icon name="server" size="xl" class="mx-auto mb-3 text-gray-300 dark:text-gray-600" /><p class="text-sm text-gray-500 dark:text-gray-400">{{ groups.length ? t('admin.availableModels.noSearchResults') : t('admin.availableModels.empty') }}</p></div>

      <div v-else class="catalog-groups">
        <section v-for="group in filteredGroups" :key="group.group_id" class="group-panel">
          <div class="group-row">
            <div class="group-identity"><h2 class="group-name">{{ group.group_name }}</h2><div class="group-meta"><span class="status-pill" :class="group.enabled ? 'status-pill-success' : 'status-pill-muted'"><span class="status-dot" :class="group.enabled ? 'bg-emerald-500' : 'bg-gray-400'" />{{ group.enabled ? t('admin.availableModels.enabled') : t('admin.availableModels.disabled') }}</span><span>{{ group.models.length }} {{ t('admin.availableModels.modelsUnit') }}</span><span v-if="group.last_synced_at">{{ formatTime(group.last_synced_at) }}</span></div></div>
            <div class="group-models">
              <div v-if="group.enabled && group.models.length" class="model-list"><div v-for="model in group.models" :key="model" class="model-row"><span class="model-dot" /><span>{{ model }}</span></div></div>
              <div v-else class="flex items-center gap-2 py-2 text-sm text-gray-400 dark:text-gray-500"><Icon name="inbox" size="sm" /><span>{{ t('admin.availableModels.noModels') }}</span></div>
            </div>
            <button v-if="canManage" class="btn btn-secondary group-sync shrink-0 px-2.5 py-1.5 text-xs" :disabled="syncing === group.group_id" :title="t('admin.availableModels.sync')" @click="syncGroup(group.group_id)"><Icon name="refresh" size="xs" :class="syncing === group.group_id ? 'animate-spin' : ''" /><span>{{ syncing === group.group_id ? t('admin.availableModels.syncing') : t('admin.availableModels.sync') }}</span></button>
          </div>
          <div class="px-5 pb-4">
            <div v-if="group.sync_error" class="mt-4 flex items-start gap-2 rounded-lg border border-red-200 bg-red-50 px-3 py-2.5 text-xs text-red-700 dark:border-red-900/40 dark:bg-red-900/20 dark:text-red-300"><Icon name="exclamationTriangle" size="xs" class="mt-0.5 shrink-0" /><span>{{ group.sync_error }}</span></div>
          </div>
        </section>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import { adminAPI } from '@/api'
import type { AvailableModelGroup } from '@/api/admin/availableModels'
import { useAppStore } from '@/stores/app'
import { useAuthStore } from '@/stores/auth'

const { t } = useI18n()
const appStore = useAppStore()
const authStore = useAuthStore()
const groups = ref<AvailableModelGroup[]>([])
const searchTerm = ref('')
const loading = ref(false)
const syncing = ref<number | null>(null)
const error = ref('')
const canManage = computed(() => authStore.isAdmin)
const filteredGroups = computed(() => {
  const keyword = searchTerm.value.trim().toLowerCase()
  if (!keyword) return groups.value
  return groups.value.filter(group => group.group_name.toLowerCase().includes(keyword) || group.models.some(model => model.toLowerCase().includes(keyword)))
})
const formatTime = (value: string) => new Date(value).toLocaleString()

const load = async () => {
  loading.value = true
  error.value = ''
  try { groups.value = await adminAPI.availableModels.list() } catch (e: any) { error.value = e?.response?.data?.message || e?.message || t('admin.availableModels.loadFailed') } finally { loading.value = false }
}

const syncGroup = async (id: number) => {
  syncing.value = id
  error.value = ''
  try { await adminAPI.availableModels.sync(id); appStore.showSuccess(t('admin.availableModels.syncSuccess')); await load() } catch (e: any) { error.value = e?.response?.data?.message || e?.response?.data?.detail || e?.message || t('admin.availableModels.syncFailed'); await load() } finally { syncing.value = null }
}

onMounted(load)
</script>

<style scoped>
.available-models-page { --panel-border: rgb(229 231 235); }
.dark .available-models-page { --panel-border: rgb(75 85 99); }
.catalog-header { display: flex; align-items: flex-end; justify-content: space-between; gap: 1.5rem; margin-bottom: 1.625rem; }
.catalog-eyebrow { color: rgb(37 99 235); font-size: .75rem; font-weight: 700; letter-spacing: .08em; text-transform: uppercase; }
.catalog-title { margin: .25rem 0 .375rem; font-size: 1.75rem; line-height: 1.2; letter-spacing: -.02em; font-weight: 700; color: rgb(23 32 51); }
.dark .catalog-title { color: white; }
.catalog-intro { margin: 0; color: rgb(107 114 128); font-size: .875rem; }
.dark .catalog-intro { color: rgb(156 163 175); }
.catalog-actions { display: flex; align-items: center; gap: .625rem; }
.catalog-search { width: 15rem; height: 2.375rem; border: 1px solid var(--panel-border); border-radius: .5rem; background: white; padding: 0 .75rem; color: rgb(23 32 51); outline: none; }
.catalog-search:focus { border-color: rgb(147 197 253); box-shadow: 0 0 0 3px rgb(37 99 235 / .1); }
.dark .catalog-search { background: rgb(31 41 55); color: white; }
.catalog-refresh { height: 2.375rem; white-space: nowrap; }
.catalog-groups { display: grid; gap: .625rem; }
.dark .group-panel { background: rgb(31 41 55 / .85); }
.group-panel { overflow: hidden; border: 1px solid var(--panel-border); border-radius: .625rem; background: white; box-shadow: 0 8px 24px rgb(15 23 42 / .05); }
.group-row { display: grid; grid-template-columns: 13.75rem minmax(0, 1fr) auto; align-items: center; gap: 1.5rem; min-height: 5.5rem; padding: 1rem 1.125rem; }
.group-identity { min-width: 0; }
.group-name { margin: 0; font-size: .9375rem; font-weight: 700; color: rgb(23 32 51); }
.dark .group-name { color: white; }
.group-models { min-width: 0; }
.group-meta { display: flex; align-items: center; flex-wrap: wrap; gap: .5rem; margin-top: .375rem; color: rgb(107 114 128); font-size: .75rem; }
.dark .group-meta { color: rgb(156 163 175); }
.model-list { display: flex; flex-wrap: wrap; gap: .5rem; }
.model-row { display: inline-flex; min-width: 0; align-items: center; gap: .45rem; border: 1px solid rgb(226 232 240); border-radius: .5rem; background: linear-gradient(180deg, #fff 0%, #f8fafc 100%); padding: .42rem .65rem; font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace; font-size: .75rem; font-weight: 600; line-height: 1.35; color: rgb(38 52 73); box-shadow: 0 1px 2px rgb(15 23 42 / .04); }
.model-row span:last-child { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.model-row:hover { border-color: rgb(147 197 253); background: rgb(239 246 255); color: rgb(29 78 216); }
.model-dot { height: .45rem; width: .45rem; flex-shrink: 0; border: 2px solid rgb(96 165 250); border-radius: 9999px; background: rgb(219 234 254); box-shadow: 0 0 0 2px rgb(239 246 255); }
.group-sync { margin-top: .1rem; }
.dark .model-row { border-color: rgb(75 85 99); background: rgb(55 65 81 / .7); color: rgb(226 232 240); }
.status-pill { display: inline-flex; align-items: center; gap: .375rem; border-radius: 9999px; padding: .25rem .5rem; font-weight: 500; }
.status-pill-success { background: rgb(236 253 245); color: rgb(4 120 87); }
.status-pill-muted { background: rgb(243 244 246); color: rgb(107 114 128); }
.dark .status-pill-success { background: rgb(6 78 59 / .35); color: rgb(110 231 183); }
.dark .status-pill-muted { background: rgb(55 65 81); color: rgb(156 163 175); }
.status-dot { height: .375rem; width: .375rem; border-radius: 9999px; }
.model-chip { max-width: 100%; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; border: 1px solid rgb(226 232 240); border-radius: .5rem; background: rgb(248 250 252); padding: .375rem .625rem; font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace; font-size: .75rem; color: rgb(51 65 85); }
.dark .model-chip { border-color: rgb(75 85 99); background: rgb(55 65 81 / .7); color: rgb(226 232 240); }
@media (max-width: 767px) {
  .catalog-header { align-items: stretch; flex-direction: column; gap: 1rem; }
  .catalog-actions { width: 100%; }
  .catalog-search { flex: 1; width: auto; }
  .group-row { grid-template-columns: 1fr; gap: .75rem; }
  .group-models { grid-column: 1 / -1; }
  .group-sync { justify-self: start; }
}
</style>
