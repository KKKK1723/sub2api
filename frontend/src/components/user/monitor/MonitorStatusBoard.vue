<template>
  <section class="status-board">
    <div class="mb-5 flex flex-wrap items-center justify-between gap-3">
      <p class="text-sm text-gray-500 dark:text-dark-400">
        {{ t('channelStatus.board.subtitle', { minutes: Math.round(intervalSeconds / 60) }) }}
      </p>
      <div v-if="isAdmin" class="flex items-center gap-2">
        <button
          type="button"
          class="flex h-9 w-9 items-center justify-center rounded-xl border border-gray-200 bg-white text-gray-500 transition-colors hover:text-gray-900 dark:border-dark-700 dark:bg-dark-900 dark:text-dark-400 dark:hover:text-white"
          :title="t('channelStatus.board.sortOrder')"
          :aria-label="t('channelStatus.board.sortOrder')"
          @click="emit('sort-order')"
        >
          <Icon name="arrowsUpDown" size="md" />
        </button>
        <div class="inline-flex rounded-xl border border-gray-200 bg-white p-1 dark:border-dark-700 dark:bg-dark-900" role="tablist">
          <button
            v-for="view in viewOptions"
            :key="view"
            type="button"
            role="tab"
            :aria-selected="activeView === view"
            class="rounded-lg px-3.5 py-1.5 text-sm font-medium transition-colors"
            :class="activeView === view
              ? 'bg-gray-900 text-white dark:bg-white dark:text-dark-950'
              : 'text-gray-500 hover:text-gray-800 dark:text-dark-400 dark:hover:text-dark-200'"
            @click="selectView(view)"
          >
            {{ t(`channelStatus.board.views.${view}`) }}
          </button>
        </div>
      </div>
    </div>

    <!-- 指标卡：整体状态 / 正常分组 / 平均可用性 / 平均延迟 / 刷新倒计时 -->
    <div class="mb-6 grid grid-cols-2 gap-3 md:grid-cols-3 xl:grid-cols-[1.4fr_repeat(4,1fr)]">
      <div
        class="board-kpi board-kpi--overall col-span-2 flex items-center gap-3.5 md:col-span-3 xl:col-span-1"
        :class="`tone-${summary.level}`"
      >
        <span class="board-pulse" aria-hidden="true"></span>
        <div class="relative min-w-0">
          <small>{{ overallLabel }}</small>
          <strong class="text-lg">{{ t(`channelStatus.board.overall.${summary.overallKey}`) }}</strong>
        </div>
      </div>
      <div class="board-kpi">
        <small>{{ t('channelStatus.board.healthyGroups') }}</small>
        <strong class="font-mono">{{ summary.green }}<em>/ {{ summary.total }}</em></strong>
        <div class="mt-2.5 flex h-1.5 overflow-hidden rounded-full bg-gray-100 dark:bg-dark-800">
          <span class="bg-emerald-500 dark:bg-emerald-400" :style="{ width: ratioWidth(summary.green) }"></span>
          <span class="bg-amber-500 dark:bg-amber-400" :style="{ width: ratioWidth(summary.yellow) }"></span>
          <span class="bg-red-500 dark:bg-red-400" :style="{ width: ratioWidth(summary.red) }"></span>
        </div>
      </div>
      <div class="board-kpi">
        <small>{{ t('channelStatus.board.avgAvailability', { window: windowLabel }) }}</small>
        <strong class="font-mono">{{ formatPercent(summary.avgAvailability) }}</strong>
      </div>
      <div class="board-kpi">
        <small>{{ t('channelStatus.board.avgLatency') }}</small>
        <strong class="font-mono">{{ formatLatency(summary.avgLatency) }}<em>ms</em></strong>
      </div>
      <div class="board-kpi flex items-center gap-3">
        <span class="board-ring" :style="{ '--ratio': countdownRatio }" aria-hidden="true"></span>
        <div>
          <small>{{ t('channelStatus.board.nextRefresh') }}</small>
          <strong class="font-mono">{{ countdownSeconds }}<em>s</em></strong>
        </div>
      </div>
    </div>

    <div class="mb-1 flex flex-wrap items-center justify-between gap-3">
      <nav v-if="platformSections.length" class="board-platforms flex flex-wrap gap-2" :aria-label="t('channelStatus.board.allPlatforms')">
        <button
          v-for="option in platformOptions"
          :key="option.value"
          type="button"
          class="board-chip"
          :class="{ active: selectedPlatform === option.value }"
          :aria-pressed="selectedPlatform === option.value"
          @click="selectedPlatform = option.value"
        >
          <ProviderIcon v-if="option.value !== ALL_PLATFORMS" :provider="option.value" :size="15" />
          {{ option.label }}
          <small>{{ option.count }}</small>
        </button>
      </nav>
      <div
        v-if="activeView === 'groups'"
        class="inline-flex rounded-xl border border-gray-200 bg-white p-[3px] dark:border-dark-700 dark:bg-dark-900"
        role="tablist"
      >
        <button
          v-for="option in windowOptions"
          :key="option.value"
          type="button"
          role="tab"
          :aria-selected="window === option.value"
          class="rounded-lg px-3 py-1 text-xs transition-colors"
          :class="window === option.value
            ? 'bg-gray-100 font-semibold text-gray-900 dark:bg-dark-800 dark:text-white'
            : 'text-gray-500 hover:text-gray-800 dark:text-dark-400 dark:hover:text-dark-200'"
          @click="emit('update:window', option.value)"
        >
          {{ option.label }}
        </button>
      </div>
    </div>

    <!-- 分组状态 -->
    <template v-if="activeView === 'groups'">
      <div
        v-if="initialLoading"
        class="mt-6 grid grid-cols-1 gap-3.5 md:grid-cols-2 xl:grid-cols-3"
        aria-busy="true"
        :aria-label="t('channelStatus.board.loading')"
      >
        <div
          v-for="n in 6"
          :key="n"
          class="h-[264px] animate-pulse rounded-2xl border border-gray-200 bg-white dark:border-dark-800 dark:bg-dark-900"
        ></div>
      </div>
      <div v-else-if="items.length === 0" class="board-empty">{{ t('channelStatus.board.empty') }}</div>
      <template v-else>
        <section v-for="platform in visibleSections" :key="platform.value" class="board-section">
          <h3 class="mb-3 mt-6 flex items-center gap-2 text-[15px] font-semibold text-gray-900 dark:text-white">
            <ProviderIcon :provider="platform.value" :size="16" />
            {{ platform.label }}
            <small class="text-xs font-normal text-gray-500 dark:text-dark-400">{{ t('channelStatus.board.groupCount', { count: platform.items.length }) }}</small>
          </h3>
          <div class="grid grid-cols-1 gap-3.5 md:grid-cols-2 xl:grid-cols-3">
            <article
              v-for="item in platform.items"
              :key="item.id"
              class="board-card"
              :class="`tone-${toneOf(item.primary_status)}`"
              role="button"
              tabindex="0"
              @click="emit('cardClick', item)"
              @keydown.enter="emit('cardClick', item)"
              @mouseenter="emit('cardHover', item)"
            >
              <div class="flex items-center gap-2.5">
                <span class="flex h-9 w-9 flex-shrink-0 items-center justify-center rounded-[10px] bg-gray-100 text-gray-700 dark:bg-dark-800 dark:text-dark-200">
                  <ProviderIcon :provider="item.provider" :size="18" />
                </span>
                <div class="min-w-0">
                  <strong class="block truncate text-sm font-semibold text-gray-900 dark:text-white">{{ item.group_name || item.name }}</strong>
                  <code class="block truncate text-[11px] text-gray-400 dark:text-dark-500">{{ item.primary_model }}</code>
                </div>
                <span class="board-pill"><i></i>{{ statusLabel(item.primary_status) }}</span>
              </div>

              <div class="mb-2 mt-[18px] flex items-end justify-between gap-3">
                <div>
                  <strong class="block font-mono text-[26px] font-semibold tracking-tight text-gray-900 dark:text-white">{{ formatAvailability(item) }}</strong>
                  <small class="text-[11px] text-gray-500 dark:text-dark-400">{{ t('channelStatus.board.availability', { window: windowLabel }) }}</small>
                </div>
                <div class="text-right">
                  <strong class="block font-mono text-lg font-semibold text-gray-900 dark:text-white">
                    {{ formatLatency(item.primary_latency_ms) }}<small class="text-[11px] font-normal text-gray-400"> ms</small>
                  </strong>
                  <small class="text-[11px] text-gray-500 dark:text-dark-400">
                    {{ t('channelStatus.board.latency') }} · {{ t('channelStatus.board.ping', { value: formatLatency(item.primary_ping_latency_ms) }) }}
                  </small>
                </div>
              </div>

              <!-- 最近 5 小时延迟曲线，失败时段铺红底 -->
              <svg class="board-spark" :viewBox="`0 0 ${SPARK_WIDTH} ${SPARK_HEIGHT}`" preserveAspectRatio="none" aria-hidden="true">
                <defs>
                  <linearGradient :id="`spark-${item.id}`" x1="0" x2="0" y1="0" y2="1">
                    <stop offset="0" stop-color="currentColor" stop-opacity=".25" />
                    <stop offset="1" stop-color="currentColor" stop-opacity="0" />
                  </linearGradient>
                </defs>
                <rect
                  v-for="band in sparkOf(item).fails"
                  :key="band.x"
                  class="board-spark__fail"
                  :x="band.x"
                  y="0"
                  :width="band.width"
                  :height="SPARK_HEIGHT"
                />
                <path :d="sparkOf(item).area" :fill="`url(#spark-${item.id})`" />
                <path class="board-spark__line" :d="sparkOf(item).line" vector-effect="non-scaling-stroke" />
              </svg>

              <div class="board-strip" :aria-label="t('channelStatus.board.timelineLabel')">
                <span
                  v-for="(point, index) in timelinePoints(item)"
                  :key="`${item.id}-${index}`"
                  class="board-cell"
                  :class="[`cell-${toneOf(point.status)}`, tooltipAlign(index)]"
                >
                  <span v-if="point.checked_at" class="board-tip">
                    <strong>{{ formatCheckedAt(point.checked_at) }}</strong>
                    <span>{{ t('channelStatus.board.tooltip.status', { value: statusLabel(point.status) }) }}</span>
                    <span>{{ t('channelStatus.board.tooltip.latency', { value: formatLatency(point.latency_ms) }) }}</span>
                    <span>{{ t('channelStatus.board.tooltip.ping', { value: formatLatency(point.ping_latency_ms) }) }}</span>
                  </span>
                </span>
              </div>
              <div class="mt-1.5 flex justify-between text-[10px] text-gray-400 dark:text-dark-500">
                <span>{{ t('channelStatus.board.hoursAgo') }}</span>
                <span>{{ t('channelStatus.board.now') }}</span>
              </div>

              <div
                v-if="item.extra_models?.length"
                class="mt-3 flex flex-wrap gap-1.5 border-t border-dashed border-gray-200 pt-3 dark:border-dark-700"
              >
                <span
                  v-for="model in modelsOf(item)"
                  :key="model.model"
                  class="inline-flex items-center gap-1 rounded-md bg-gray-100 px-1.5 py-0.5 text-[11px] text-gray-500 dark:bg-dark-800 dark:text-dark-400"
                >
                  <i class="board-dot" :class="`tone-${toneOf(model.status)}`"></i>{{ model.model }}
                </span>
              </div>
            </article>
          </div>
        </section>
      </template>
    </template>

    <!-- 探针状态（管理员） -->
    <div v-else class="mt-5">
      <div class="mb-4 flex items-center justify-between gap-3">
        <div>
          <h2 class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('channelStatus.board.probes.title') }}</h2>
          <p class="mt-1 text-xs text-gray-500 dark:text-dark-400">{{ t('channelStatus.board.probes.description') }}</p>
        </div>
        <span class="rounded-full bg-emerald-50 px-2.5 py-1 text-xs font-semibold text-emerald-600 dark:bg-emerald-500/10 dark:text-emerald-400">
          {{ t('channelStatus.board.probes.healthy', { healthy: probeHealthyCount, total: probeTotalCount }) }}
        </span>
      </div>

      <div v-if="items.length === 0" class="board-empty">{{ t('channelStatus.board.empty') }}</div>
      <template v-else>
        <section v-for="platform in visibleSections" :key="platform.value" class="mb-6 grid gap-3">
          <h3 class="flex items-center gap-2 text-[15px] font-semibold text-gray-900 dark:text-white">
            <ProviderIcon :provider="platform.value" :size="16" />
            {{ platform.label }}
            <small class="text-xs font-normal text-gray-500 dark:text-dark-400">{{ t('channelStatus.board.groupCount', { count: platform.items.length }) }}</small>
          </h3>
          <article
            v-for="item in platform.items"
            :key="item.id"
            class="board-probe-group overflow-hidden rounded-2xl border border-gray-200 bg-white dark:border-dark-800 dark:bg-dark-900"
          >
            <header class="flex items-center justify-between gap-3 border-b border-gray-100 px-[18px] py-3.5 dark:border-dark-800">
              <div class="flex min-w-0 items-center gap-2.5">
                <span class="flex h-8 w-8 flex-shrink-0 items-center justify-center rounded-[10px] bg-gray-100 text-gray-700 dark:bg-dark-800 dark:text-dark-200">
                  <ProviderIcon :provider="item.provider" :size="16" />
                </span>
                <strong class="truncate text-sm text-gray-900 dark:text-white">{{ item.group_name || item.name }}</strong>
              </div>
              <span class="text-xs text-gray-400 dark:text-dark-500">{{ t('channelStatus.board.probes.count', { count: probesFor(item.id).length }) }}</span>
            </header>

            <div v-if="probeLoading[item.id]" class="board-empty !py-8">{{ t('channelStatus.board.probes.loading') }}</div>
            <div v-else-if="probesFor(item.id).length === 0" class="board-empty !py-8">{{ t('channelStatus.board.probes.empty') }}</div>
            <div v-else class="px-[18px] py-1">
              <div
                v-for="(probe, index) in probesFor(item.id)"
                :key="`${probe.endpoint}-${index}`"
                class="board-probe"
                :class="{ 'opacity-60': !probe.enabled }"
              >
                <div class="flex min-w-0 items-center gap-2.5">
                  <span class="flex h-6 w-6 flex-shrink-0 items-center justify-center rounded-md bg-gray-100 font-mono text-[10px] font-bold text-gray-500 dark:bg-dark-800 dark:text-dark-400">
                    {{ String(index + 1).padStart(2, '0') }}
                  </span>
                  <strong class="truncate text-xs text-gray-900 dark:text-white">{{ probe.name || t('channelStatus.board.probes.unnamed', { index: index + 1 }) }}</strong>
                </div>
                <div class="min-w-0 font-mono text-[11px]" :title="probe.endpoint">
                  <span class="block break-all text-gray-600 dark:text-dark-300">{{ probe.endpoint || t('channelStatus.board.probes.noEndpoint') }}</span>
                  <small class="mt-0.5 block break-all text-gray-400 dark:text-dark-500">
                    {{ t('channelStatus.board.probes.apiKey', { value: probe.api_key || probe.api_key_masked || t('channelStatus.board.probes.noKey') }) }}
                  </small>
                </div>
                <div class="text-right">
                  <strong class="block font-mono text-xs text-gray-900 dark:text-white">{{ formatLatency(probe.latency_ms) }} ms</strong>
                  <small class="text-[10px] text-gray-400 dark:text-dark-500">{{ t('channelStatus.board.latency') }}</small>
                </div>
                <span class="board-pill !ml-0 justify-self-end" :class="`tone-${toneOf(probe.status || '')}`"><i></i>{{ statusLabel(probe.status || '') }}</span>
                <span
                  class="justify-self-end text-[11px]"
                  :class="probe.enabled ? 'text-gray-500 dark:text-dark-400' : 'text-red-500 dark:text-red-400'"
                >
                  {{ probe.enabled ? t('channelStatus.board.probes.enabled') : t('channelStatus.board.probes.disabled') }}
                </span>
              </div>
            </div>

            <footer
              v-if="!probeLoading[item.id] && probesFor(item.id).length"
              class="flex items-center justify-between gap-2 border-t border-gray-100 px-[18px] py-2.5 text-xs text-gray-500 dark:border-dark-800 dark:text-dark-400"
            >
              <span>{{ t('channelStatus.board.probes.overall') }}</span>
              <span class="board-pill !ml-0" :class="`tone-${toneOf(probeOverallStatus(item.id))}`"><i></i>{{ statusLabel(probeOverallStatus(item.id)) }}</span>
            </footer>
          </article>
        </section>
      </template>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import type { MonitorProbe, MonitorStatus } from '@/api/admin/channelMonitor'
import type { MonitorTimelinePoint, UserMonitorDetail, UserMonitorView } from '@/api/channelMonitor'
import ProviderIcon from './ProviderIcon.vue'
import type { MonitorWindow } from './MonitorHero.vue'
import { groupMonitorPlatforms } from '@/utils/monitorPlatforms'
import { useChannelMonitorFormat } from '@/composables/useChannelMonitorFormat'

type Tone = 'green' | 'yellow' | 'red' | 'empty'
type ViewKey = 'groups' | 'probes'

const props = defineProps<{
  items: UserMonitorView[]
  window: MonitorWindow
  countdownSeconds: number
  intervalSeconds: number
  loading: boolean
  detailCache: Record<number, UserMonitorDetail>
  isAdmin: boolean
  probeCache: Record<number, MonitorProbe[]>
  probeLoading: Record<number, boolean>
}>()

const emit = defineEmits<{
  (event: 'update:window', value: MonitorWindow): void
  (event: 'cardClick', item: UserMonitorView): void
  (event: 'cardHover', item: UserMonitorView): void
  (event: 'probe-view'): void
  (event: 'sort-order'): void
}>()

const { t, locale } = useI18n()
const { statusLabel, formatLatency, formatPercent } = useChannelMonitorFormat()

// 时间线固定 60 个点（每 5 分钟一次，覆盖最近 5 小时），不足时左侧补空
const TIMELINE_POINTS = 60
const SPARK_WIDTH = 300
const SPARK_HEIGHT = 56
const ALL_PLATFORMS = 'all'

const viewOptions: ViewKey[] = ['groups', 'probes']
const activeView = ref<ViewKey>('groups')

const windowOptions = computed<{ value: MonitorWindow; label: string }[]>(() => [
  { value: '7d', label: t('channelStatus.windowTab.7d') },
  { value: '15d', label: t('channelStatus.windowTab.15d') },
  { value: '30d', label: t('channelStatus.windowTab.30d') },
])
const windowLabel = computed(() => t(`channelStatus.windowTab.${props.window}`))

const platformSections = computed(() => groupMonitorPlatforms(props.items))
const selectedPlatform = ref(ALL_PLATFORMS)
watch(platformSections, sections => {
  if (selectedPlatform.value !== ALL_PLATFORMS && !sections.some(section => section.value === selectedPlatform.value)) {
    selectedPlatform.value = ALL_PLATFORMS
  }
})
const platformOptions = computed(() => [
  { value: ALL_PLATFORMS, label: t('channelStatus.board.allPlatforms'), count: props.items.length },
  ...platformSections.value.map(section => ({ value: section.value, label: section.label, count: section.items.length })),
])
const visibleSections = computed(() =>
  selectedPlatform.value === ALL_PLATFORMS
    ? platformSections.value
    : platformSections.value.filter(section => section.value === selectedPlatform.value)
)
const visibleItems = computed(() => visibleSections.value.flatMap(section => section.items))

// 选中具体平台时，整体状态标题带上平台名，提示指标卡的统计范围
const overallLabel = computed(() => {
  const platform = platformSections.value.find(section => section.value === selectedPlatform.value)
  return platform
    ? t('channelStatus.board.platformOverallLabel', { platform: platform.label })
    : t('channelStatus.board.overallLabel')
})

// 首次加载且尚无数据时显示骨架屏，避免把“加载中”误显示为“暂无监控数据”
const initialLoading = computed(() => props.loading && props.items.length === 0)

// 汇总跟随平台筛选：选「全部平台」统计所有分组，选某个平台只统计该平台分组
const summary = computed(() => {
  const scoped = visibleItems.value
  const counts: Record<Tone, number> = { green: 0, yellow: 0, red: 0, empty: 0 }
  scoped.forEach(item => counts[toneOf(item.primary_status)]++)
  const availabilities = scoped.map(availabilityValue).filter((v): v is number => v != null)
  const latencies = scoped.map(item => item.primary_latency_ms).filter((v): v is number => v != null)
  const level: Tone = scoped.length === 0 ? 'empty' : counts.red ? 'red' : counts.yellow ? 'yellow' : 'green'
  const overallKey = initialLoading.value ? 'loading' : { green: 'operational', yellow: 'degraded', red: 'outage', empty: 'empty' }[level]
  return {
    ...counts,
    total: scoped.length,
    level,
    overallKey,
    avgAvailability: availabilities.length ? availabilities.reduce((a, b) => a + b, 0) / availabilities.length : null,
    avgLatency: latencies.length ? latencies.reduce((a, b) => a + b, 0) / latencies.length : null,
  }
})

const countdownRatio = computed(() => (props.intervalSeconds > 0 ? Math.min(1, props.countdownSeconds / props.intervalSeconds) : 0))

function ratioWidth(count: number): string {
  return `${summary.value.total ? (count / summary.value.total) * 100 : 0}%`
}

const probeTotalCount = computed(() => visibleItems.value.reduce((sum, item) => sum + probesFor(item.id).length, 0))
const probeHealthyCount = computed(() => visibleItems.value.reduce((sum, item) => {
  return sum + probesFor(item.id).filter(probe => probe.enabled && (!probe.status || probe.status === 'operational')).length
}, 0))

function selectView(view: ViewKey) {
  activeView.value = view
  if (view === 'probes') emit('probe-view')
}

function probesFor(id: number): MonitorProbe[] {
  return props.probeCache[id] ?? []
}

function availabilityValue(item: UserMonitorView): number | null {
  if (props.window === '7d') return item.availability_7d ?? null
  const detail = props.detailCache[item.id]
  const primary = detail?.models.find(model => model.model === item.primary_model)
  return props.window === '15d' ? primary?.availability_15d ?? null : primary?.availability_30d ?? null
}

function formatAvailability(item: UserMonitorView): string {
  return formatPercent(availabilityValue(item))
}

// 接口返回的时间线是新到旧，展示时改为旧到新
function timelinePoints(item: UserMonitorView): MonitorTimelinePoint[] {
  const points = [...(item.timeline ?? [])].slice(0, TIMELINE_POINTS).reverse()
  while (points.length < TIMELINE_POINTS) points.unshift({ status: '' as MonitorStatus, latency_ms: null, ping_latency_ms: null, checked_at: '' })
  return points
}

function toneOf(status: MonitorStatus | ''): Tone {
  if (status === 'failed' || status === 'error') return 'red'
  if (status === 'degraded') return 'yellow'
  if (status === 'operational') return 'green'
  return 'empty'
}

function modelsOf(item: UserMonitorView) {
  return [{ model: item.primary_model, status: item.primary_status }, ...(item.extra_models ?? [])]
}

// 靠两端的格子把提示框向内对齐，避免超出卡片可视范围
function tooltipAlign(index: number): string {
  if (index < 8) return 'tip-start'
  if (index >= TIMELINE_POINTS - 8) return 'tip-end'
  return ''
}

interface Spark {
  line: string
  area: string
  fails: { x: number; width: number }[]
}

// 按 60 个时间点生成延迟折线；无延迟（失败或未采样）的点断开，失败点铺红底
function buildSpark(item: UserMonitorView): Spark {
  const points = timelinePoints(item)
  const values = points.map(point => point.latency_ms).filter((v): v is number => v != null)
  const step = SPARK_WIDTH / (TIMELINE_POINTS - 1)
  const fails = points.flatMap((point, index) =>
    toneOf(point.status) === 'red' ? [{ x: Math.max(0, index * step - step / 2), width: step }] : []
  )
  if (!values.length) return { line: '', area: '', fails }

  const max = Math.max(...values) * 1.1
  const min = Math.min(...values) * 0.9
  const range = max - min || 1
  const segments: string[][] = [[]]
  points.forEach((point, index) => {
    if (point.latency_ms == null) {
      if (segments[segments.length - 1].length) segments.push([])
      return
    }
    const x = (index * step).toFixed(1)
    const y = (SPARK_HEIGHT - 4 - ((point.latency_ms - min) / range) * (SPARK_HEIGHT - 8)).toFixed(1)
    segments[segments.length - 1].push(`${x},${y}`)
  })
  const filled = segments.filter(segment => segment.length)
  const line = filled.map(segment => `M${segment.join('L')}`).join('')
  const area = filled.map(segment => {
    const first = segment[0].split(',')[0]
    const last = segment[segment.length - 1].split(',')[0]
    return `M${first},${SPARK_HEIGHT}L${segment.join('L')}L${last},${SPARK_HEIGHT}Z`
  }).join('')
  return { line, area, fails }
}

const sparkCache = computed(() => {
  const cache: Record<number, Spark> = {}
  props.items.forEach(item => { cache[item.id] = buildSpark(item) })
  return cache
})

function sparkOf(item: UserMonitorView): Spark {
  return sparkCache.value[item.id] ?? { line: '', area: '', fails: [] }
}

function probeOverallStatus(id: number): MonitorStatus | '' {
  const statuses = probesFor(id).filter(probe => probe.enabled).map(probe => probe.status || '')
  if (statuses.includes('operational')) return 'operational'
  if (statuses.includes('degraded')) return 'degraded'
  if (statuses.includes('failed')) return 'failed'
  if (statuses.includes('error')) return 'error'
  return ''
}

function formatCheckedAt(value: string): string {
  const date = new Date(value)
  return Number.isNaN(date.getTime())
    ? value
    : date.toLocaleString(locale.value, { month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' })
}
</script>

<style scoped>
/* 状态色：卡片、胶囊、圆点共用 --c */
.tone-green { --c: #10b981; }
.tone-yellow { --c: #f59e0b; }
.tone-red { --c: #ef4444; }
.tone-empty { --c: #94a3b8; }
:global(.dark) .tone-green { --c: #34d399; }
:global(.dark) .tone-yellow { --c: #fbbf24; }
:global(.dark) .tone-red { --c: #f87171; }
:global(.dark) .tone-empty { --c: #64748b; }

.board-kpi {
  position: relative;
  overflow: hidden;
  padding: 16px 18px;
  border: 1px solid #e2e8f0;
  border-radius: 16px;
  background: #fff;
}
:global(.dark) .board-kpi {
  border-color: #1e293b;
  background: #0f172a;
}
.board-kpi small {
  display: block;
  color: #64748b;
  font-size: 12px;
}
.board-kpi strong {
  display: block;
  margin-top: 6px;
  color: #0f172a;
  font-size: 26px;
  font-weight: 650;
  letter-spacing: -0.02em;
}
.board-kpi strong em {
  margin-left: 2px;
  color: #94a3b8;
  font-size: 14px;
  font-style: normal;
  font-weight: 500;
}
:global(.dark) .board-kpi small {
  color: #94a3b8;
}
:global(.dark) .board-kpi strong {
  color: #f1f5f9;
}
.board-kpi--overall::before {
  content: '';
  position: absolute;
  inset: 0;
  background: linear-gradient(120deg, color-mix(in srgb, var(--c) 12%, transparent), transparent 70%);
}
.board-kpi--overall strong {
  margin-top: 2px;
  font-size: 18px;
}

/* 整体状态呼吸点 */
.board-pulse {
  position: relative;
  width: 44px;
  height: 44px;
  flex-shrink: 0;
}
.board-pulse::before,
.board-pulse::after {
  content: '';
  position: absolute;
  border-radius: 50%;
  background: var(--c);
}
.board-pulse::before {
  inset: 12px;
}
.board-pulse::after {
  inset: 0;
  opacity: 0.18;
  animation: board-ring 2s ease-out infinite;
}

/* 刷新倒计时圆环 */
.board-ring {
  width: 44px;
  height: 44px;
  flex-shrink: 0;
  border-radius: 50%;
  background: conic-gradient(#0d9488 calc(var(--ratio) * 360deg), #f1f5f9 0);
  -webkit-mask: radial-gradient(circle, transparent 15px, #000 16px);
  mask: radial-gradient(circle, transparent 15px, #000 16px);
}
:global(.dark) .board-ring {
  background: conic-gradient(#2dd4bf calc(var(--ratio) * 360deg), #1e293b 0);
}

.board-chip {
  display: inline-flex;
  align-items: center;
  gap: 7px;
  padding: 7px 13px;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  background: #fff;
  color: #64748b;
  font-size: 13px;
  font-weight: 500;
  transition: border-color 0.15s, box-shadow 0.15s;
}
.board-chip small {
  padding: 0 6px;
  border-radius: 999px;
  background: #f1f5f9;
  font-size: 11px;
}
.board-chip.active {
  border-color: #0d9488;
  color: #0f172a;
  box-shadow: 0 0 0 3px rgba(13, 148, 136, 0.15);
}
:global(.dark) .board-chip {
  border-color: #1e293b;
  background: #0f172a;
  color: #94a3b8;
}
:global(.dark) .board-chip small {
  background: #1e293b;
}
:global(.dark) .board-chip.active {
  border-color: #2dd4bf;
  color: #f1f5f9;
  box-shadow: 0 0 0 3px rgba(45, 212, 191, 0.15);
}

/* 分组卡片：顶部状态色条用内阴影实现，保留 overflow 可见以显示时间线提示 */
.board-card {
  position: relative;
  display: flex;
  flex-direction: column;
  padding: 18px;
  border: 1px solid #e2e8f0;
  border-radius: 16px;
  background: #fff;
  box-shadow: inset 0 3px var(--c);
  cursor: pointer;
  transition: transform 0.2s, box-shadow 0.2s;
}
.board-card:hover,
.board-card:focus-visible {
  transform: translateY(-3px);
  box-shadow: inset 0 3px var(--c), 0 16px 40px rgba(15, 23, 42, 0.1);
  outline: none;
}
.board-card.tone-red {
  border-color: color-mix(in srgb, var(--c) 45%, #e2e8f0);
}
:global(.dark) .board-card {
  border-color: #1e293b;
  background: #0f172a;
}
:global(.dark) .board-card:hover,
:global(.dark) .board-card:focus-visible {
  box-shadow: inset 0 3px var(--c), 0 16px 40px rgba(0, 0, 0, 0.35);
}
:global(.dark) .board-card.tone-red {
  border-color: color-mix(in srgb, var(--c) 45%, #1e293b);
}

.board-pill {
  margin-left: auto;
  display: inline-flex;
  flex-shrink: 0;
  align-items: center;
  gap: 5px;
  padding: 3px 9px;
  border-radius: 999px;
  background: color-mix(in srgb, var(--c) 14%, transparent);
  color: var(--c);
  font-size: 12px;
  font-weight: 600;
  white-space: nowrap;
}
.board-pill i,
.board-dot {
  width: 6px;
  height: 6px;
  flex-shrink: 0;
  border-radius: 50%;
  background: var(--c);
}
.board-dot {
  width: 5px;
  height: 5px;
}
.tone-red .board-pill i {
  animation: board-blink 1s infinite;
}

.board-spark {
  display: block;
  width: 100%;
  height: 56px;
  color: var(--c);
}
.board-spark__line {
  fill: none;
  stroke: var(--c);
  stroke-width: 1.6;
  stroke-linejoin: round;
}
.board-spark__fail {
  fill: rgba(239, 68, 68, 0.14);
}

.board-strip {
  display: grid;
  grid-template-columns: repeat(60, minmax(0, 1fr));
  gap: 1.5px;
  margin-top: 10px;
}
.board-cell {
  position: relative;
  height: 10px;
  border-radius: 2px;
  background: #e2e8f0;
}
.board-cell.cell-green { background: rgba(16, 185, 129, 0.75); }
.board-cell.cell-yellow { background: #f59e0b; }
.board-cell.cell-red { background: #ef4444; }
:global(.dark) .board-cell { background: #1e293b; }
:global(.dark) .board-cell.cell-green { background: rgba(52, 211, 153, 0.75); }
:global(.dark) .board-cell.cell-yellow { background: #fbbf24; }
:global(.dark) .board-cell.cell-red { background: #f87171; }
.board-cell:hover {
  z-index: 20;
  outline: 1px solid #0f172a;
}
:global(.dark) .board-cell:hover {
  outline-color: #f1f5f9;
}
.board-tip {
  position: absolute;
  bottom: calc(100% + 8px);
  left: 50%;
  z-index: 30;
  display: none;
  min-width: 140px;
  padding: 8px 10px;
  border-radius: 8px;
  background: #0f172a;
  color: #fff;
  box-shadow: 0 10px 28px rgba(15, 23, 42, 0.25);
  font-size: 11px;
  line-height: 1.55;
  transform: translateX(-50%);
  white-space: nowrap;
  pointer-events: none;
}
.board-tip strong,
.board-tip span {
  display: block;
}
.board-tip strong {
  margin-bottom: 2px;
  color: #5eead4;
}
.board-cell.tip-start .board-tip {
  left: 0;
  transform: none;
}
.board-cell.tip-end .board-tip {
  right: 0;
  left: auto;
  transform: none;
}
.board-cell:hover .board-tip {
  display: block;
}

.board-empty {
  margin-top: 20px;
  padding: 48px 20px;
  border: 1px dashed #e2e8f0;
  border-radius: 16px;
  color: #64748b;
  font-size: 13px;
  text-align: center;
}
:global(.dark) .board-empty {
  border-color: #1e293b;
  color: #94a3b8;
}

.board-probe {
  display: grid;
  grid-template-columns: minmax(130px, 1.1fr) minmax(180px, 1.8fr) 80px 72px 56px;
  gap: 14px;
  align-items: center;
  min-height: 58px;
  border-bottom: 1px solid #f1f5f9;
}
.board-probe:last-child {
  border-bottom: 0;
}
:global(.dark) .board-probe {
  border-color: #1e293b;
}

@keyframes board-ring {
  from {
    transform: scale(0.5);
    opacity: 0.4;
  }
  to {
    transform: scale(1.1);
    opacity: 0;
  }
}
@keyframes board-blink {
  50% {
    opacity: 0.2;
  }
}
@media (max-width: 700px) {
  .board-probe {
    grid-template-columns: minmax(0, 1fr) 72px;
    gap: 8px;
    padding: 10px 0;
  }
  .board-probe > :nth-child(2) {
    grid-column: 1 / -1;
    grid-row: 2;
  }
  .board-probe > :nth-child(3),
  .board-probe > :nth-child(5) {
    display: none;
  }
}
@media (prefers-reduced-motion: reduce) {
  .board-pulse::after,
  .tone-red .board-pill i {
    animation: none;
  }
  .board-card {
    transition: none;
  }
}
</style>
