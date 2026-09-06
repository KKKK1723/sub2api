<template>
  <button
    type="button"
    class="group text-left p-5 rounded-2xl min-h-[280px] w-full bg-white/70 backdrop-blur-xl border border-gray-200/80 shadow-card dark:bg-dark-800/60 dark:border-dark-700/70 hover:-translate-y-1 hover:shadow-card-hover dark:hover:border-primary-500/30 hover:border-gray-300 transition-all duration-300 ease-out flex flex-col"
    :class="isAdmin ? 'monitor-card-admin overflow-hidden' : ''"
    @click="emit('click')"
    @mouseenter="handlePointerEnter"
    @mouseleave="handlePointerLeave"
    @focus="handlePointerEnter"
    @blur="handlePointerLeave"
  >
    <div
      class="relative flex flex-1 min-w-0 w-full flex-col"
      :class="[
        isAdmin ? 'monitor-card-flip-inner' : '',
        { 'is-flipped': isAdmin && isFlipped },
      ]"
    >
      <div
        class="flex flex-1 min-w-0 w-full flex-col"
        :class="isAdmin ? 'monitor-card-face monitor-card-front' : ''"
      >
        <!-- 头部：图标、渠道名称、状态标签 -->
        <div class="flex items-start gap-3">
          <span
            class="w-9 h-9 rounded-xl ring-1 ring-black/5 dark:ring-white/10 grid place-items-center flex-shrink-0"
            :class="[providerGradient(item.provider), providerTintClass]"
          >
            <ProviderIcon :provider="item.provider" :size="20" />
          </span>
          <div class="flex-1 min-w-0">
            <div class="text-base font-semibold truncate text-gray-900 dark:text-gray-100">
              {{ item.name }}
            </div>
            <div class="mt-0.5 flex items-center gap-1.5 min-w-0">
              <span
                class="inline-flex items-center rounded-md px-1.5 py-0.5 text-[10px] font-medium flex-shrink-0"
                :class="providerBadgeClass(item.provider)"
              >
                {{ providerLabel(item.provider) }}
              </span>
              <span class="font-mono text-xs truncate text-gray-500 dark:text-gray-400">
                {{ item.primary_model }}
              </span>
              <span
                v-if="item.group_name"
                class="inline-flex items-center rounded-md px-1.5 py-0.5 text-[10px] font-medium bg-gray-100 text-gray-600 dark:bg-dark-700 dark:text-gray-300 flex-shrink-0"
              >
                {{ item.group_name }}
              </span>
            </div>
          </div>
          <span
            class="px-2.5 py-1 rounded-full text-xs font-semibold flex-shrink-0"
            :class="statusBadgeClass(item.primary_status)"
          >
            {{ statusLabel(item.primary_status) }}
          </span>
        </div>

        <!-- Metrics -->
        <MonitorMetricPair
          primary-icon="bolt"
          :primary-label="t('monitorCommon.dialogLatency')"
          :primary-value="formatLatency(item.primary_latency_ms)"
          primary-unit="ms"
          secondary-icon="globe"
          :secondary-label="t('monitorCommon.endpointPing')"
          :secondary-value="formatLatency(item.primary_ping_latency_ms)"
          secondary-unit="ms"
        />

        <!-- Divider -->
        <div class="mt-4 border-t border-gray-100 dark:border-dark-700/60"></div>

        <!-- Availability row -->
        <MonitorAvailabilityRow
          :window-label="availabilityLabel"
          :value="availabilityValue"
          :samples-label="extraModelsCountLabel"
        />

        <!-- Timeline -->
        <MonitorTimeline
          :buckets="item.timeline"
          :countdown-seconds="countdownSeconds"
        />
      </div>

      <div
        v-if="isAdmin"
        class="monitor-card-face monitor-card-back flex min-w-0 w-full flex-col rounded-xl bg-white/90 dark:bg-dark-800/90"
        aria-live="polite"
      >
        <div class="flex items-start justify-between gap-3">
          <div class="min-w-0">
            <div class="text-[10px] font-semibold uppercase tracking-wider text-gray-500 dark:text-gray-400">
              {{ t('monitorCommon.probesTitle') }}
            </div>
            <div class="mt-1 truncate text-sm font-semibold text-gray-900 dark:text-gray-100">
              {{ item.name }}
            </div>
          </div>
          <span class="flex-shrink-0 rounded-full bg-gray-100 px-2 py-1 text-[10px] font-semibold text-gray-600 dark:bg-dark-700 dark:text-gray-300">
            {{ t('monitorCommon.probeCount', { n: probes.length }) }}
          </span>
        </div>

        <div
          v-if="probeLoading"
          class="flex flex-1 items-center justify-center gap-2 text-xs text-gray-500 dark:text-gray-400"
        >
          <span class="h-2 w-2 animate-pulse rounded-full bg-primary-500"></span>
          {{ t('monitorCommon.probeLoading') }}
        </div>
        <div
          v-else-if="probes.length === 0"
          class="flex flex-1 items-center justify-center text-xs text-gray-500 dark:text-gray-400"
        >
          {{ t('monitorCommon.probeEmpty') }}
        </div>
        <div v-else class="mt-3 min-h-0 flex-1 space-y-2 overflow-y-auto pr-1">
          <div
            v-for="(probe, index) in probes"
            :key="`${probe.endpoint}-${index}`"
            class="rounded-xl border border-gray-200/80 bg-gray-50/80 p-2.5 dark:border-dark-700/80 dark:bg-dark-900/40"
          >
            <div class="flex items-center justify-between gap-2">
              <div class="flex min-w-0 items-center gap-2">
                <span class="grid h-5 w-5 flex-shrink-0 place-items-center rounded-md bg-gray-200 text-[10px] font-semibold text-gray-600 dark:bg-dark-700 dark:text-gray-300">
                  {{ String(index + 1).padStart(2, '0') }}
                </span>
                <span class="truncate text-xs font-semibold text-gray-800 dark:text-gray-100">
                  {{ probeDisplayName(probe, index) }}
                </span>
              </div>
              <span
                class="flex-shrink-0 rounded-full px-2 py-0.5 text-[10px] font-semibold"
                :class="statusBadgeClass(probeStatus(probe))"
              >
                {{ statusLabel(probeStatus(probe)) }}
              </span>
            </div>
            <div class="mt-2 flex min-w-0 items-center gap-1.5 text-[10px] text-gray-500 dark:text-gray-400">
              <span class="flex-shrink-0 font-medium">{{ t('monitorCommon.probeEndpoint') }}</span>
              <span class="truncate font-mono text-gray-700 dark:text-gray-300" :title="probe.endpoint">
                {{ probe.endpoint || t('monitorCommon.latencyEmpty') }}
              </span>
            </div>
            <div class="mt-1 flex min-w-0 items-center justify-between gap-2 text-[10px] text-gray-500 dark:text-gray-400">
              <span class="min-w-0 truncate font-mono" :title="probe.api_key_masked || undefined">
                {{ t('monitorCommon.probeApiKey') }}: {{ probe.api_key_masked || t('monitorCommon.latencyEmpty') }}
              </span>
              <span class="flex-shrink-0 tabular-nums">
                {{ t('monitorCommon.probeLatency') }} {{ formatLatency(probe.latency_ms) }} ms
              </span>
            </div>
            <div class="mt-1 text-right text-[10px] text-gray-500 dark:text-gray-400">
              {{ probe.enabled ? t('monitorCommon.probeEnabled') : t('monitorCommon.probeDisabled') }}
            </div>
          </div>
        </div>

        <div
          v-if="!probeLoading && probes.length > 0"
          class="mt-3 flex items-center justify-between gap-2 border-t border-gray-200/80 pt-2 text-[10px] dark:border-dark-700/80"
        >
          <span class="text-gray-500 dark:text-gray-400">{{ t('monitorCommon.probeOverall') }}</span>
          <span
            class="rounded-full px-2 py-0.5 font-semibold"
            :class="statusBadgeClass(probeOverallStatus)"
          >
            {{ statusLabel(probeOverallStatus) }}
          </span>
        </div>
      </div>
    </div>
  </button>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import type { UserMonitorView } from '@/api/channelMonitor'
import type { MonitorProbe, MonitorStatus } from '@/api/admin/channelMonitor'
import {
  useChannelMonitorFormat,
  providerGradient,
} from '@/composables/useChannelMonitorFormat'
import ProviderIcon from './ProviderIcon.vue'
import MonitorMetricPair from './MonitorMetricPair.vue'
import MonitorAvailabilityRow from './MonitorAvailabilityRow.vue'
import MonitorTimeline from './MonitorTimeline.vue'

const PROVIDER_TINT: Record<string, string> = {
  openai: 'text-emerald-600 dark:text-emerald-300',
  anthropic: 'text-orange-600 dark:text-orange-300',
  gemini: 'text-sky-600 dark:text-sky-300',
  grok: 'text-zinc-700 dark:text-zinc-200',
}

const props = defineProps<{
  item: UserMonitorView
  window: '7d' | '15d' | '30d'
  availabilityValue: number | null
  countdownSeconds: number
  isAdmin: boolean
  probes?: MonitorProbe[]
  probeLoading: boolean
}>()

const emit = defineEmits<{
  (e: 'click'): void
  (e: 'hover'): void
}>()

const { t } = useI18n()
const isFlipped = ref(false)
const {
  statusLabel,
  statusBadgeClass,
  providerLabel,
  providerBadgeClass,
  formatLatency,
} = useChannelMonitorFormat()

const probes = computed(() => props.probes ?? [])

const probeOverallStatus = computed<MonitorStatus | ''>(() => {
  const statuses = probes.value
    .filter(probe => probe.enabled)
    .map(probe => probe.status || '')

  if (statuses.length === 0) return ''
  if (statuses.includes('operational')) return 'operational'
  if (statuses.includes('degraded')) return 'degraded'
  if (statuses.includes('failed')) return 'failed'
  if (statuses.includes('error')) return 'error'
  return ''
})

const providerTintClass = computed(() =>
  PROVIDER_TINT[props.item.provider] ?? 'text-gray-500 dark:text-gray-300'
)

const availabilityLabel = computed(() => {
  const win = t(`channelStatus.windowTab.${props.window}`)
  return `${t('monitorCommon.availabilityPrefix')} · ${win}`
})

const extraModelsCountLabel = computed(() => {
  const count = props.item.extra_models?.length ?? 0
  if (count === 0) return undefined
  return t('monitorCommon.extraModelsCount', { n: count })
})

function handlePointerEnter() {
  if (!props.isAdmin) return
  isFlipped.value = true
  emit('hover')
}

function handlePointerLeave() {
  isFlipped.value = false
}

function probeStatus(probe: MonitorProbe): MonitorStatus | '' {
  return probe.enabled ? (probe.status || '') : ''
}

function probeDisplayName(probe: MonitorProbe, index: number): string {
  return probe.name?.trim() || `${t('monitorCommon.probesTitle')} ${index + 1}`
}
</script>

<style scoped>
.monitor-card-admin {
  perspective: 1200px;
  min-height: 380px;
}

.monitor-card-flip-inner {
  min-height: 340px;
  transform-style: preserve-3d;
  transition: transform 420ms cubic-bezier(0.2, 0.7, 0.2, 1);
}

.monitor-card-flip-inner.is-flipped {
  transform: rotateY(180deg);
}

.monitor-card-face {
  position: absolute;
  inset: 0;
  backface-visibility: hidden;
  overflow: hidden;
}

.monitor-card-back {
  transform: rotateY(180deg);
}

@media (prefers-reduced-motion: reduce) {
  .monitor-card-flip-inner {
    transition: none;
  }
}
</style>
