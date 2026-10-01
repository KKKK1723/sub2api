<template>
  <AppLayout class="model-pricing-layout">
    <div class="mx-auto min-w-0 max-w-7xl space-y-5">
      <div v-if="loading" class="flex justify-center py-16"><LoadingSpinner /></div>
      <EmptyState v-else-if="loadError" :title="t('modelPricing.loadFailed')" :description="loadError"
        :action-text="t('modelPricing.retry')" :action-icon="false" @action="loadPricing" />
      <EmptyState v-else-if="!groups.length" :title="t('modelPricing.empty')" :description="t('modelPricing.emptyHint')" />
      <template v-else-if="selectedGroup">
        <section class="rounded-2xl border border-gray-200 bg-white p-4 shadow-sm dark:border-dark-700 dark:bg-dark-900">
          <div ref="platformTabs" role="tablist" :aria-label="t('modelPricing.platform')"
            class="-mx-1 flex gap-1 overflow-x-auto border-b border-gray-100 px-1 pb-3 dark:border-dark-800">
            <button v-for="(platform, index) in platforms" :key="platform.key" type="button" role="tab"
              :aria-selected="platform.key === activePlatform" :tabindex="platform.key === activePlatform ? 0 : -1"
              class="flex shrink-0 items-center gap-2 rounded-lg px-3 py-1.5 text-sm transition"
              :class="platform.key === activePlatform ? 'bg-gray-900 text-white dark:bg-white dark:text-dark-900' : 'text-gray-600 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-dark-800'"
              @click="selectPlatform(platform.key)" @keydown="onPlatformKey($event, index)">
              <span class="h-2 w-2 rounded-full" :class="platformDot(platform.key)"></span>
              {{ platformLabel(platform.key) }}
              <span class="rounded-full px-1.5 text-[11px] tabular-nums"
                :class="platform.key === activePlatform ? 'bg-white/20 dark:bg-black/10' : 'bg-gray-100 text-gray-500 dark:bg-dark-800 dark:text-gray-400'">{{ platform.groups.length }}</span>
            </button>
          </div>
          <div class="mt-3 flex items-center justify-between gap-3">
            <h2 class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('modelPricing.group') }}</h2>
            <div class="flex items-center gap-2">
              <span class="hidden text-xs text-gray-500 dark:text-gray-400 sm:inline">{{ t('modelPricing.groupHint') }}</span>
              <button type="button" :disabled="!canScroll.left" :aria-label="t('modelPricing.scrollLeft')"
                class="rounded-md border border-gray-200 px-1.5 py-1 text-gray-600 disabled:opacity-30 dark:border-dark-700 dark:text-gray-300" @click="scrollGroups(-1)"><Icon name="chevronLeft" size="sm" /></button>
              <button type="button" :disabled="!canScroll.right" :aria-label="t('modelPricing.scrollRight')"
                class="rounded-md border border-gray-200 px-1.5 py-1 text-gray-600 disabled:opacity-30 dark:border-dark-700 dark:text-gray-300" @click="scrollGroups(1)"><Icon name="chevronRight" size="sm" /></button>
            </div>
          </div>
          <div class="relative mt-2">
            <div ref="strip" role="radiogroup" :aria-label="t('modelPricing.group')"
              class="flex snap-x snap-mandatory gap-3 overflow-x-auto scroll-smooth pb-2 [scrollbar-width:thin]" @scroll="updateScroll">
              <button v-for="(group, index) in platformGroups" :key="group.id" :data-gid="group.id"
                type="button" role="radio" :aria-checked="group.id === groupId" :tabindex="group.id === groupId ? 0 : -1"
                class="w-52 shrink-0 snap-start rounded-xl border p-3 text-left transition"
                :class="group.id === groupId ? 'border-primary-500 bg-primary-50/60 ring-2 ring-primary-500/20 dark:bg-primary-900/20' : 'border-gray-200 hover:border-primary-300 dark:border-dark-700 dark:hover:border-primary-700'"
                @click="selectGroup(group)" @keydown="onGroupKey($event, index)">
                <div class="flex items-center justify-between gap-2">
                  <span class="truncate text-sm font-medium text-gray-900 dark:text-white" :title="group.name">{{ group.name }}</span>
                  <span class="flex shrink-0 gap-1">
                    <span v-if="group.is_exclusive" class="rounded bg-sky-100 px-1.5 py-0.5 text-[10px] font-medium text-sky-700 dark:bg-sky-900/40 dark:text-sky-300">{{ t('availableChannels.exclusive') }}</span>
                    <span v-if="group.subscription_type === 'subscription'" class="rounded bg-violet-100 px-1.5 py-0.5 text-[10px] font-medium text-violet-700 dark:bg-violet-900/40 dark:text-violet-300">{{ t('groups.subscription') }}</span>
                  </span>
                </div>
                <div class="mt-2 flex items-end justify-between gap-2">
                  <span class="text-[11px] text-gray-500 dark:text-gray-400">{{ t(group.user_rate_multiplier !== null ? 'modelPricing.personalRate' : 'modelPricing.groupRate') }}</span>
                  <span class="text-right tabular-nums">
                    <span v-if="group.user_rate_multiplier !== null" class="mr-1 text-xs text-gray-400 line-through">{{ fmtRate(group.rate_multiplier) }}x</span>
                    <span class="text-lg font-semibold" :class="rateColor(effectiveGroupRate(group))">{{ fmtRate(effectiveGroupRate(group)) }}x</span>
                  </span>
                </div>
                <div v-if="sample(group)" class="mt-1 truncate text-[11px] tabular-nums text-gray-500 dark:text-gray-400">
                  {{ sample(group)!.name }} {{ t('modelPricing.output') }} {{ formatScaled(sample(group)!.output_price * effectiveGroupRate(group), unit) }}
                </div>
                <div v-if="group.description" class="mt-1 truncate text-[11px] text-gray-400 dark:text-gray-500" :title="group.description">{{ group.description }}</div>
                <div v-if="groupHasPeak(group)" class="mt-1 truncate text-[11px] text-amber-600 dark:text-amber-400" :title="peakWindow(group)">{{ t('modelPricing.peak') }} {{ peakWindow(group) }}</div>
              </button>
            </div>
            <div v-if="canScroll.right" class="pointer-events-none absolute inset-y-0 right-0 w-10 bg-gradient-to-l from-white dark:from-dark-900"></div>
          </div>
        </section>

        <section class="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
          <div class="flex flex-1 flex-wrap items-center gap-3">
            <div class="relative w-full sm:w-72">
              <Icon name="search" size="sm" class="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
              <input v-model="search" type="search" :aria-label="t('modelPricing.search')" :placeholder="t('modelPricing.search')"
                class="w-full rounded-lg border border-gray-200 bg-white py-2 pl-9 pr-3 text-sm text-gray-900 outline-none focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 dark:border-dark-700 dark:bg-dark-900 dark:text-gray-100" />
            </div>
            <label class="flex cursor-pointer items-center gap-2 text-sm text-gray-600 dark:text-gray-300">
              <input v-model="showOfficial" type="checkbox" class="h-4 w-4 rounded accent-primary-600" />{{ t('modelPricing.showOfficial') }}
            </label>
            <label class="flex items-center gap-2 text-sm text-gray-600 dark:text-gray-300" :class="groupHasPeak(selectedGroup) ? 'cursor-pointer' : 'opacity-40'" :title="peakWindow(selectedGroup)">
              <input v-model="peakMode" type="checkbox" :disabled="!groupHasPeak(selectedGroup)" class="h-4 w-4 rounded accent-amber-500" />{{ t('modelPricing.peakMode') }}
            </label>
          </div>
          <div class="flex items-center gap-2 text-sm">
            <div role="group" :aria-label="t('modelPricing.unit')" class="inline-flex rounded-lg border border-gray-200 p-0.5 dark:border-dark-700">
              <button v-for="value in [1_000_000, 1_000]" :key="value" type="button" :aria-pressed="unit === value"
                class="rounded-md px-3 py-1" :class="unit === value ? 'bg-primary-600 text-white' : 'text-gray-600 dark:text-gray-300'" @click="unit = value">{{ value === 1_000_000 ? '1M' : '1K' }}</button>
            </div>
            <button type="button" :title="t('common.refresh')" :aria-label="t('common.refresh')" class="rounded-lg border border-gray-200 p-2 text-gray-500 hover:bg-gray-100 dark:border-dark-700 dark:hover:bg-dark-800" @click="loadPricing"><Icon name="refresh" size="sm" /></button>
          </div>
        </section>

        <div class="flex flex-wrap items-center gap-x-4 gap-y-1 rounded-xl bg-primary-50 px-4 py-2.5 text-sm text-primary-800 dark:bg-primary-900/20 dark:text-primary-200">
          <span>{{ t('modelPricing.current') }} <b>{{ platformLabel(activePlatform) }} / {{ selectedGroup.name }}</b></span>
          <span class="tabular-nums">{{ t('modelPricing.billingRate') }} <b>{{ fmtRate(currentRate) }}x</b><span v-if="peakMode && groupHasPeak(selectedGroup)"> ({{ fmtRate(effectiveGroupRate(selectedGroup)) }} × {{ t('modelPricing.peak') }} {{ fmtRate(selectedGroup.peak_rate_multiplier) }})</span></span>
          <span class="text-xs opacity-75">{{ t(`modelPricing.sources.${selectedGroup.model_source}`) }}</span>
          <span class="tabular-nums">{{ t('modelPricing.modelCount', { count: filteredModels.length }) }}</span>
          <span class="text-xs opacity-75">{{ t('modelPricing.unitHint', { unit: unitLabel }) }}</span>
        </div>

        <section class="overflow-hidden rounded-2xl border border-gray-200 bg-white shadow-sm dark:border-dark-700 dark:bg-dark-900">
          <div v-if="!filteredModels.length" class="p-12 text-center text-sm text-gray-500">{{ t(selectedGroup.models.length ? 'modelPricing.noMatches' : 'modelPricing.noModels') }}</div>
          <div v-else class="relative overflow-x-auto">
            <table class="min-w-full text-sm">
              <thead class="bg-gray-50 text-xs text-gray-500 dark:bg-dark-800/60 dark:text-gray-400">
                <tr>
                  <th scope="col" class="px-4 py-3 text-left font-medium" :aria-sort="ariaSort('name')"><button type="button" @click="toggleSort('name')">{{ t('modelPricing.model') }} {{ sortIcon('name') }}</button></th>
                  <th scope="col" class="whitespace-nowrap px-4 py-3 text-right font-medium">{{ t('modelPricing.contextWindow') }}</th>
                  <th v-for="field in tokenFields" :key="field.key" scope="col" class="whitespace-nowrap px-4 py-3 text-right font-medium" :aria-sort="ariaSort(field.key)">
                    <button type="button" @click="toggleSort(field.key)">{{ t(field.label) }} {{ sortIcon(field.key) }}</button>
                  </th>
                  <th v-if="hasRequestModels" scope="col" class="whitespace-nowrap px-4 py-3 text-right font-medium">{{ t('modelPricing.unitPrice') }}</th>
                  <th scope="col" class="px-4 py-3"><span class="sr-only">{{ t('modelPricing.calculate') }}</span></th>
                </tr>
              </thead>
              <tbody class="divide-y divide-gray-100 dark:divide-dark-800">
                <template v-for="model in sortedModels" :key="model.name">
                  <tr class="hover:bg-gray-50/70 dark:hover:bg-dark-800/40">
                    <td class="px-4 py-3">
                      <div class="flex items-center gap-2">
                        <span class="h-2 w-2 shrink-0 rounded-full" :class="platformDot(activePlatform)"></span>
                        <code class="whitespace-nowrap font-mono text-[13px] font-medium text-gray-900 dark:text-white">{{ model.name }}</code>
                        <span v-if="!model.priced" class="whitespace-nowrap rounded bg-rose-50 px-1.5 text-[10px] text-rose-700 dark:bg-rose-900/30 dark:text-rose-300">{{ t('modelPricing.unpricedShort') }}</span>
                        <span v-else-if="model.pricing_source === 'channel'" class="whitespace-nowrap rounded bg-primary-50 px-1.5 text-[10px] text-primary-700 dark:bg-primary-900/30 dark:text-primary-300">{{ t('modelPricing.channelPrice') }}</span>
                        <button type="button" :aria-label="t('modelPricing.copyModel', { model: model.name })" :title="t('modelPricing.copyModel', { model: model.name })" class="text-gray-400 hover:text-primary-600" @click="copyModel(model.name)"><Icon :name="copiedModel === model.name ? 'check' : 'copy'" size="sm" /></button>
                      </div>
                      <div class="mt-1 flex flex-wrap gap-1 pl-4">
                        <span v-if="model.reasoning" class="rounded bg-sky-50 px-1.5 text-[10px] text-sky-700 dark:bg-sky-900/30 dark:text-sky-300">{{ t('modelPricing.reasoning') }}</span>
                        <span v-if="model.vision" class="rounded bg-emerald-50 px-1.5 text-[10px] text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-300">{{ t('modelPricing.vision') }}</span>
                        <span v-if="model.priced && model.long_context_threshold && !model.intervals?.length"
                          class="rounded bg-amber-50 px-1.5 text-[10px] text-amber-700 dark:bg-amber-900/30 dark:text-amber-300"
                          :title="t('modelPricing.longContextRule', { tokens: model.long_context_threshold, input: model.long_context_input_multiplier, output: model.long_context_output_multiplier })">{{ t('modelPricing.longContextBadge', { threshold: formatContext(model.long_context_threshold) }) }}</span>
                        <span v-if="model.billing_mode !== 'token'" class="rounded px-1.5 text-[10px]" :class="getBillingModeBadgeClass(model.billing_mode)">{{ getBillingModeLabel(model.billing_mode, t) }}</span>
                      </div>
                    </td>
                    <td class="whitespace-nowrap px-4 py-3 text-right tabular-nums text-gray-500 dark:text-gray-400">{{ formatContext(model.context_window) }}</td>
                    <td v-for="field in tokenFields" :key="field.key" class="whitespace-nowrap px-4 py-3 text-right tabular-nums">
                      <template v-if="model.priced && model.billing_mode === 'token'">
                        <div class="font-semibold text-gray-900 dark:text-white">{{ displayPrice(model, tokenPrices(model)[field.key], unit) }}</div>
                        <div v-if="showOfficial && displayRate(model) !== 1" class="text-[11px] text-gray-400"><span class="line-through">{{ formatScaled(tokenPrices(model)[field.key], unit) }}</span> {{ originalLabel(model) }}</div>
                      </template>
                      <span v-else class="text-gray-300 dark:text-gray-600">—</span>
                    </td>
                    <td v-if="hasRequestModels" class="whitespace-nowrap px-4 py-3 text-right tabular-nums">
                      <template v-if="model.priced && model.billing_mode !== 'token'">
                        <div class="font-semibold text-gray-900 dark:text-white">{{ displayPrice(model, model.per_request_price, 1) }} <span class="text-xs font-normal">{{ unitLabelFor(model) }}</span></div>
                        <div v-if="showOfficial && displayRate(model) !== 1" class="text-[11px] text-gray-400"><span class="line-through">{{ formatScaled(model.per_request_price, 1) }}</span> {{ originalLabel(model) }}</div>
                      </template>
                      <span v-else class="text-gray-300 dark:text-gray-600">—</span>
                    </td>
                    <td class="whitespace-nowrap px-4 py-3 text-right"><button v-if="model.priced" type="button" :aria-label="t('modelPricing.calculateModel', { model: model.name })"
                      class="rounded-lg px-2.5 py-1 text-xs text-primary-700 hover:bg-primary-50 dark:text-primary-300 dark:hover:bg-primary-900/30" @click="openCalc(model)">{{ t('modelPricing.calculate') }}</button></td>
                  </tr>
                  <tr v-if="model.priced && model.intervals?.length">
                    <td :colspan="hasRequestModels ? 8 : 7" class="px-4 pb-3">
                      <details class="text-xs text-gray-600 dark:text-gray-300">
                        <summary class="cursor-pointer py-1">{{ t('modelPricing.tierDetails') }}</summary>
                        <div v-for="(interval, index) in model.intervals" :key="index" class="flex flex-wrap gap-x-5 gap-y-1 border-t border-gray-100 py-2 dark:border-dark-700">
                          <span class="min-w-48 font-medium">{{ intervalLabel(interval) }}</span>
                          <template v-if="model.billing_mode === 'token'"><span v-for="field in tokenFields" :key="field.key">{{ t(field.label) }}: {{ displayPrice(model, tokenPrices(model, interval)[field.key], unit) }}</span></template>
                          <span v-else>{{ interval.per_request_price ? displayPrice(model, interval.per_request_price, 1) + ' ' + unitLabelFor(model) : t('modelPricing.defaultPrice') }}</span>
                        </div>
                      </details>
                    </td>
                  </tr>
                </template>
              </tbody>
            </table>
          </div>
        </section>
        <section class="grid gap-3 text-xs text-gray-500 dark:text-gray-400 md:grid-cols-3">
          <div v-for="note in ['actual', 'cache', 'source']" :key="note" class="rounded-xl border border-dashed border-gray-300 p-3 dark:border-dark-700">
            <b class="text-gray-700 dark:text-gray-200">{{ t(`modelPricing.notes.${note}Title`) }}</b> {{ t(`modelPricing.notes.${note}`) }}
          </div>
        </section>
      </template>
    </div>

    <Teleport to="body">
      <div v-if="calcOpen && selectedGroup && calcModel" class="fixed inset-0 z-50 bg-black/30" @click.self="closeCalc">
        <aside ref="drawer" role="dialog" aria-modal="true" aria-labelledby="mp-calc-title" tabindex="-1"
          class="fixed right-0 top-0 flex h-dvh w-full max-w-md flex-col bg-white shadow-2xl dark:bg-dark-900" @keydown="onDrawerKey">
          <div class="flex shrink-0 items-center justify-between border-b border-gray-200 px-5 py-4 dark:border-dark-800">
            <div class="min-w-0"><h3 id="mp-calc-title" class="font-semibold text-gray-900 dark:text-white">{{ t('modelPricing.calculator') }}</h3><code class="break-all text-xs text-gray-500">{{ calcModel.name }}</code></div>
            <button type="button" :aria-label="t('common.close')" class="ml-3 shrink-0 rounded-lg p-1.5 text-gray-400 hover:bg-gray-100 dark:hover:bg-dark-800" @click="closeCalc"><Icon name="x" size="md" /></button>
          </div>
          <div class="min-h-0 flex-1 space-y-4 overflow-y-auto p-5 text-sm">
            <div class="rounded-lg bg-gray-50 px-3 py-2 text-xs text-gray-600 dark:bg-dark-800 dark:text-gray-300">
              {{ t('modelPricing.group') }} <b>{{ selectedGroup.name }}</b> · {{ t('modelPricing.billingRate') }} <b class="tabular-nums">{{ fmtRate(displayRate(calcModel)) }}x</b>
            </div>
            <div v-for="field in calcFields" :key="field.key">
              <label :for="`mp-${field.key}`" class="mb-1 flex justify-between gap-2 text-xs text-gray-600 dark:text-gray-300">
                <span>{{ t('modelPricing.perCallTokens', { field: t(field.label) }) }}</span>
                <span v-if="calcModel.billing_mode === 'token'" class="tabular-nums text-gray-400">{{ displayPrice(calcModel, calcPrices[field.key], 1_000_000) }} / 1M</span>
              </label>
              <input :id="`mp-${field.key}`" v-model.number="usage[field.key]" type="number" min="0" step="1" class="mp-calc-input" />
            </div>
            <div>
              <label for="mp-count" class="mb-1 block text-xs text-gray-600 dark:text-gray-300">{{ t(calcModel.billing_mode === 'image' ? 'modelPricing.imageCount' : 'modelPricing.requestCount') }}</label>
              <input id="mp-count" v-model.number="usage.count" type="number" min="1" step="1" class="mp-calc-input" />
            </div>
            <div v-if="calcTiers.length">
              <label for="mp-tier" class="mb-1 block text-xs text-gray-600 dark:text-gray-300">{{ t('modelPricing.tier') }}</label>
              <select id="mp-tier" v-model="usage.tier" class="mp-calc-input"><option value="">{{ t('modelPricing.autoTier') }}</option><option v-for="tier in calcTiers" :key="tier" :value="tier">{{ tier }}</option></select>
            </div>
            <p class="text-xs leading-relaxed text-gray-500">{{ t('modelPricing.tokenHint') }}</p>
            <p v-if="!calcResult" role="alert" class="text-sm text-red-600">{{ t('modelPricing.invalidInput') }}</p>
            <template v-else>
              <p v-if="calcResult.longContext" class="rounded-lg bg-amber-50 px-3 py-2 text-xs text-amber-700 dark:bg-amber-900/30 dark:text-amber-300">{{ t('modelPricing.longContextApplied') }}</p>
              <p v-if="calcResult.interval" class="text-xs text-gray-500">{{ intervalLabel(calcResult.interval) }}</p>
              <div class="space-y-1.5 rounded-xl border border-gray-200 p-4 dark:border-dark-700" aria-live="polite">
                <div v-for="row in calcRows" :key="row.key" class="flex justify-between gap-3 text-xs tabular-nums text-gray-600 dark:text-gray-300"><span>{{ t(row.label) }}</span><span>{{ formatScaled(row.cost, 1) }}</span></div>
                <div class="flex justify-between gap-3 border-t border-gray-100 pt-2 text-xs tabular-nums text-gray-400 dark:border-dark-800"><span>{{ t('modelPricing.originalTotal') }}</span><span data-testid="base-cost" class="line-through">{{ formatScaled(calcResult.base, 1) }}</span></div>
                <div class="flex flex-wrap items-baseline justify-between gap-2"><span class="font-medium text-gray-900 dark:text-white">{{ t('modelPricing.actualCost') }}</span><span data-testid="actual-cost" class="break-all text-2xl font-semibold tabular-nums text-primary-600 dark:text-primary-400">{{ formatScaled(calcResult.actual, 1) }}</span></div>
              </div>
              <p class="text-xs leading-relaxed text-gray-500">{{ t('modelPricing.estimateHint') }}</p>
            </template>
          </div>
        </aside>
      </div>
    </Teleport>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import Icon from '@/components/icons/Icon.vue'
import modelPricingAPI, { type ModelPricingGroup, type ModelPricingInterval, type ModelPricingItem } from '@/api/modelPricing'
import { GROUP_PLATFORM_OPTIONS } from '@/constants/platforms'
import { useClipboard } from '@/composables/useClipboard'
import { useAppStore } from '@/stores/app'
import { extractApiErrorMessage } from '@/utils/apiError'
import { getBillingModeBadgeClass, getBillingModeLabel } from '@/utils/billingMode'
import { formatPeakRateWindow, hasPeakRate, serverTimezoneLabel } from '@/utils/peak-rate'
import { formatScaled } from '@/utils/pricing'
import { effectiveGroupRate, estimateModelCost, modelRate, tokenPrices } from '@/utils/modelPricing'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const appStore = useAppStore()
const { copyToClipboard } = useClipboard()
const groups = ref<ModelPricingGroup[]>([])
const groupId = ref<number | null>(null)
const activePlatform = ref('')
const loading = ref(true)
const loadError = ref('')
const search = ref('')
const showOfficial = ref(true)
const peakMode = ref(false)
const unit = ref(1_000_000)
const strip = ref<HTMLElement | null>(null)
const platformTabs = ref<HTMLElement | null>(null)
const canScroll = reactive({ left: false, right: false })
const copiedModel = ref('')
let copyTimer: ReturnType<typeof setTimeout> | undefined
let resizeObserver: ResizeObserver | undefined
const calcOpen = ref(false)
const calcModelName = ref('')
const drawer = ref<HTMLElement | null>(null)
let previousFocus: HTMLElement | null = null
const usage = reactive({ input: 20000, output: 2000, cacheWrite: 0, cacheRead: 0, count: 100, tier: '' })
const tokenFields = [
  { key: 'input', label: 'modelPricing.input' },
  { key: 'output', label: 'modelPricing.output' },
  { key: 'cacheWrite', label: 'modelPricing.cacheWrite' },
  { key: 'cacheRead', label: 'modelPricing.cacheRead' },
] as const
type SortKey = 'name' | typeof tokenFields[number]['key']
const sort = reactive<{ key: SortKey; direction: number }>({ key: 'name', direction: 1 })
const platformColors: Record<string, string> = {
  anthropic: 'bg-orange-500', openai: 'bg-emerald-500', gemini: 'bg-blue-500', antigravity: 'bg-fuchsia-500',
  grok: 'bg-zinc-500', kimi: 'bg-cyan-500', zhipu: 'bg-sky-500', deepseek: 'bg-indigo-500',
  minimax: 'bg-rose-500', opencode_go: 'bg-lime-500', composite: 'bg-gray-400',
}
const platforms = computed(() => {
  const byPlatform = new Map<string, ModelPricingGroup[]>()
  for (const group of groups.value) {
    if (!byPlatform.has(group.platform)) byPlatform.set(group.platform, [])
    byPlatform.get(group.platform)!.push(group)
  }
  const rank = (key: string) => {
    const index = GROUP_PLATFORM_OPTIONS.findIndex(option => option.value === key)
    return index < 0 ? GROUP_PLATFORM_OPTIONS.length : index
  }
  return [...byPlatform].map(([key, items]) => ({ key, groups: items.sort((a, b) => effectiveGroupRate(a) - effectiveGroupRate(b) || a.id - b.id) }))
    .sort((a, b) => rank(a.key) - rank(b.key))
})
const platformGroups = computed(() => platforms.value.find(platform => platform.key === activePlatform.value)?.groups ?? [])
const selectedGroup = computed(() => groups.value.find(group => group.id === groupId.value))
// 高峰勾选仅用于价格预览；实际扣费仍由服务器时间决定。
const pricingGroup = computed(() => selectedGroup.value ? {
  ...selectedGroup.value,
  peak_factor_now: peakMode.value && groupHasPeak(selectedGroup.value) ? selectedGroup.value.peak_rate_multiplier : 1,
} : null)
const currentRate = computed(() => pricingGroup.value ? effectiveGroupRate(pricingGroup.value) * pricingGroup.value.peak_factor_now : 1)
const unitLabel = computed(() => unit.value === 1_000_000 ? '1M' : '1K')
const filteredModels = computed(() => selectedGroup.value?.models.filter(model => model.name.toLowerCase().includes(search.value.trim().toLowerCase())) ?? [])
const hasRequestModels = computed(() => selectedGroup.value?.models.some(model => model.billing_mode !== 'token'))
const sortedModels = computed(() => [...filteredModels.value].sort((a, b) => {
  if (sort.key === 'name') return a.name.localeCompare(b.name) * sort.direction
  const x = a.priced && a.billing_mode === 'token' ? tokenPrices(a)[sort.key] : null
  const y = b.priced && b.billing_mode === 'token' ? tokenPrices(b)[sort.key] : null
  if (x === null) return y === null ? 0 : 1
  if (y === null) return -1
  return (x - y) * sort.direction || a.name.localeCompare(b.name)
}))
const calcModel = computed(() => selectedGroup.value?.models.find(model => model.name === calcModelName.value))
const calcFields = computed(() => tokenFields.filter(field => calcModel.value?.billing_mode === 'token' || field.key !== 'output'))
const calcTiers = computed(() => calcModel.value?.billing_mode === 'token' ? [] : [...new Set(calcModel.value?.intervals?.map(interval => interval.tier_label).filter((label): label is string => Boolean(label)) ?? [])])
const calcResult = computed(() => {
  if (!pricingGroup.value || !calcModel.value || !Number.isSafeInteger(usage.count) || usage.count < 1) return null
  const result = estimateModelCost(calcModel.value, pricingGroup.value, usage)
  if (!result) return null
  // Token 区间按单次上下文判断，再乘请求数；按次模式已由计费工具乘数量。
  const count = calcModel.value.billing_mode === 'token' ? usage.count : 1
  return { ...result, base: result.base * count, actual: result.actual * count }
})
const calcPrices = computed(() => {
  if (!calcModel.value) return { input: 0, output: 0, cacheWrite: 0, cacheRead: 0 }
  const prices = tokenPrices(calcModel.value, calcResult.value?.interval)
  if (calcResult.value?.longContext) {
    const inputFactor = calcModel.value.long_context_input_multiplier ?? 1
    prices.input *= inputFactor
    prices.cacheWrite *= inputFactor
    prices.cacheRead *= inputFactor
    prices.output *= calcModel.value.long_context_output_multiplier ?? 1
  }
  return prices
})
const calcRows = computed(() => {
  if (!calcResult.value || !calcModel.value) return []
  if (calcModel.value.billing_mode !== 'token') return [{ key: 'requests', label: 'modelPricing.unitPrice', cost: calcResult.value.actual }]
  return tokenFields.filter(field => usage[field.key] > 0).map(field => ({
    ...field, cost: usage[field.key] * calcPrices.value[field.key] * calcResult.value!.rate * usage.count,
  }))
})

function platformLabel(key: string) { return GROUP_PLATFORM_OPTIONS.find(option => option.value === key)?.label ?? key }
function platformDot(key: string) { return platformColors[key] ?? 'bg-gray-400' }
function fmtRate(rate: number) { return Number(rate.toFixed(4)) }
function rateColor(rate: number) {
  return rate < 1 ? 'text-emerald-600 dark:text-emerald-400' : rate > 1 ? 'text-amber-600 dark:text-amber-400' : 'text-gray-900 dark:text-white'
}
function groupHasPeak(group: ModelPricingGroup) { return group.subscription_type === 'subscription' && hasPeakRate(group) }
function peakWindow(group: ModelPricingGroup) {
  return formatPeakRateWindow(group, serverTimezoneLabel(appStore.cachedPublicSettings?.server_utc_offset) || t('modelPricing.serverTimezone'))
}
function sample(group: ModelPricingGroup) {
  const preferred = ['claude-sonnet-5', 'claude-sonnet-4-6', 'gpt-5.5', 'grok-4.6', 'gpt-5.4', 'gemini-2.5-pro', 'deepseek-chat']
  const models = group.models.filter(model => model.priced && model.billing_mode === 'token')
  return preferred.map(name => models.find(model => model.name === name)).find(Boolean) ?? models[0]
}
function displayRate(model: ModelPricingItem) { return pricingGroup.value ? modelRate(pricingGroup.value, model) : 1 }
function displayPrice(model: ModelPricingItem, price: number, scale: number) { return formatScaled(price * displayRate(model), scale) }
function originalLabel(model: ModelPricingItem) { return t(model.pricing_source === 'channel' ? 'modelPricing.channelOriginal' : model.pricing_source === 'litellm' ? 'modelPricing.official' : 'modelPricing.baseOriginal') }
function unitLabelFor(model: ModelPricingItem) { return t(model.billing_mode === 'image' ? 'modelPricing.perImage' : 'modelPricing.perRequest') }
function formatContext(tokens?: number) {
  if (!tokens) return '—'
  if (tokens < 1000) return String(tokens)
  const base = tokens % 1000 !== 0 && tokens % 1024 === 0 ? 1024 : 1000
  const thousands = tokens / base
  if (thousands < 1000) return `${Math.round(thousands)}K`
  const millions = thousands / base
  const rounded = Math.round(millions)
  return `${Math.abs(millions - rounded) / rounded < 0.06 ? rounded : Number(millions.toFixed(1))}M`
}
function intervalLabel(interval: ModelPricingInterval) {
  const range = interval.max_tokens === null ? `> ${interval.min_tokens.toLocaleString()}` : `(${interval.min_tokens.toLocaleString()}, ${interval.max_tokens.toLocaleString()}]`
  return interval.tier_label ? `${interval.tier_label} · ${range} Token` : `${range} Token`
}
function toggleSort(key: SortKey) { sort.direction = sort.key === key ? -sort.direction : 1; sort.key = key }
function sortIcon(key: SortKey) { return sort.key === key ? sort.direction === 1 ? '↑' : '↓' : '' }
function ariaSort(key: SortKey): 'none' | 'ascending' | 'descending' { return sort.key === key ? sort.direction === 1 ? 'ascending' : 'descending' : 'none' }
function updateScroll() {
  const element = strip.value
  canScroll.left = Boolean(element && element.scrollLeft > 4)
  canScroll.right = Boolean(element && element.scrollLeft + element.clientWidth < element.scrollWidth - 4)
}
function scrollGroups(direction: number) { strip.value?.scrollBy({ left: direction * strip.value.clientWidth * 0.8, behavior: 'smooth' }) }
async function selectGroup(group: ModelPricingGroup, syncUrl = true) {
  const changed = groupId.value !== group.id
  groupId.value = group.id
  activePlatform.value = group.platform
  if (!groupHasPeak(group)) peakMode.value = false
  if (changed) { closeCalc(); search.value = '' }
  if (syncUrl && route.query.g !== String(group.id)) await router.replace({ query: { ...route.query, g: String(group.id) } })
  await nextTick()
  strip.value?.querySelector<HTMLElement>(`[data-gid="${group.id}"]`)?.scrollIntoView?.({ block: 'nearest', inline: 'nearest' })
  updateScroll()
}
function selectPlatform(key: string) {
  const group = platforms.value.find(platform => platform.key === key)?.groups[0]
  if (group) void selectGroup(group)
}
function nextIndex(event: KeyboardEvent, index: number, length: number) {
  if (!['ArrowLeft', 'ArrowRight', 'Home', 'End'].includes(event.key)) return null
  event.preventDefault()
  return event.key === 'Home' ? 0 : event.key === 'End' ? length - 1 : (index + (event.key === 'ArrowRight' ? 1 : -1) + length) % length
}
async function onPlatformKey(event: KeyboardEvent, index: number) {
  const next = nextIndex(event, index, platforms.value.length)
  if (next === null) return
  selectPlatform(platforms.value[next].key)
  await nextTick()
  platformTabs.value?.querySelectorAll<HTMLButtonElement>('[role=tab]')[next]?.focus()
}
async function onGroupKey(event: KeyboardEvent, index: number) {
  const next = nextIndex(event, index, platformGroups.value.length)
  if (next === null) return
  await selectGroup(platformGroups.value[next])
  strip.value?.querySelectorAll<HTMLButtonElement>('[role=radio]')[next]?.focus()
}
async function copyModel(name: string) {
  if (!await copyToClipboard(name)) return
  copiedModel.value = name
  clearTimeout(copyTimer)
  copyTimer = setTimeout(() => { copiedModel.value = '' }, 1500)
}
function openCalc(model: ModelPricingItem) {
  previousFocus = document.activeElement as HTMLElement | null
  calcModelName.value = model.name
  usage.tier = ''
  calcOpen.value = true
}
function closeCalc() { calcOpen.value = false }
function onDrawerKey(event: KeyboardEvent) {
  if (event.key === 'Escape') { event.preventDefault(); closeCalc(); return }
  if (event.key !== 'Tab') return
  const controls = drawer.value?.querySelectorAll<HTMLElement>('button:not(:disabled), input:not(:disabled), select:not(:disabled)')
  if (!controls?.length) return
  const first = controls[0], last = controls[controls.length - 1]
  if (event.shiftKey && document.activeElement === first) { event.preventDefault(); last.focus() }
  else if (!event.shiftKey && document.activeElement === last) { event.preventDefault(); first.focus() }
}
watch(calcOpen, async open => {
  document.body.classList.toggle('modal-open', open)
  await nextTick()
  if (open) drawer.value?.querySelector<HTMLButtonElement>('button')?.focus()
  else previousFocus?.focus()
})
watch(strip, element => {
  resizeObserver?.disconnect()
  if (element && typeof ResizeObserver !== 'undefined') { resizeObserver = new ResizeObserver(updateScroll); resizeObserver.observe(element) }
  updateScroll()
})
watch(() => route.query.g, value => {
  const group = groups.value.find(item => item.id === Number(value))
  if (group && group.id !== groupId.value) void selectGroup(group, false)
})
async function loadPricing() {
  loading.value = true
  loadError.value = ''
  closeCalc()
  try {
    groups.value = await modelPricingAPI.list()
    const preferred = groups.value.find(group => group.id === Number(route.query.g))
      ?? groups.value.find(group => group.id === groupId.value) ?? platforms.value[0]?.groups[0]
    if (preferred) await selectGroup(preferred, false)
    else { groupId.value = null; activePlatform.value = '' }
  } catch (error: unknown) {
    loadError.value = extractApiErrorMessage(error, t('modelPricing.loadFailed'))
    appStore.showError(loadError.value)
  } finally {
    loading.value = false
    await nextTick()
    if (groupId.value !== null) strip.value?.querySelector<HTMLElement>(`[data-gid="${groupId.value}"]`)?.scrollIntoView?.({ block: 'nearest', inline: 'nearest' })
    updateScroll()
  }
}
onMounted(() => { void loadPricing(); window.addEventListener('resize', updateScroll) })
onUnmounted(() => {
  resizeObserver?.disconnect()
  clearTimeout(copyTimer)
  window.removeEventListener('resize', updateScroll)
  if (calcOpen.value) { document.body.classList.remove('modal-open'); previousFocus?.focus() }
})
</script>

<style scoped>
.model-pricing-layout :deep(.bg-mesh-gradient) { display: none; }
.model-pricing-layout :deep(main) { padding: 1rem; }
@media (min-width: 1024px) {
  .model-pricing-layout :deep(main) { padding: 1.5rem; }
}
.mp-calc-input {
  @apply w-full rounded-lg border border-gray-200 bg-white px-3 py-2 tabular-nums text-gray-900 outline-none focus:border-primary-500 dark:border-dark-700 dark:bg-dark-900 dark:text-gray-100;
}
</style>
