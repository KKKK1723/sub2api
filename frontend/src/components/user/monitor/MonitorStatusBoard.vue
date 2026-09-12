<template>
  <section class="status-board">
    <div class="status-board__hero">
      <div class="status-board__hero-orbit" aria-hidden="true"><span></span></div>
      <div class="status-board__hero-content">
        <span class="status-board__eyebrow"><i></i>实时监控</span>
        <h1>监控服务运行正常</h1>
        <p>最近一次采样 · 下一轮将在 {{ countdownSeconds }} 秒后开始</p>
        <div class="status-board__hero-stat">
          <span>采样周期</span>
          <strong>{{ Math.round(intervalSeconds / 60) }} min</strong>
          <small>自动</small>
        </div>
      </div>
    </div>

    <div class="status-board__toolbar">
      <div class="status-board__views" role="tablist" aria-label="监控视图">
        <button
          type="button"
          role="tab"
          :aria-selected="activeView === 'groups'"
          :class="{ active: activeView === 'groups' }"
          @click="selectView('groups')"
        >
          分组状态
        </button>
        <button
          v-if="isAdmin"
          type="button"
          role="tab"
          :aria-selected="activeView === 'probes'"
          :class="{ active: activeView === 'probes' }"
          @click="selectView('probes')"
        >
          探针状态
        </button>
      </div>
      <div class="status-board__actions">
        <div class="status-board__windows" role="tablist" aria-label="可用性时间范围">
          <button
            v-for="option in windowOptions"
            :key="option.value"
            type="button"
            :class="{ active: window === option.value }"
            @click="emit('update:window', option.value)"
          >
            {{ option.label }}
          </button>
        </div>
      </div>
    </div>

    <div v-if="activeView === 'groups'" class="status-board__panel">
      <div class="status-board__panel-head">
        <div>
          <h2>分组状态</h2>
          <p>每个分组显示最近 5 小时探测结果</p>
        </div>
        <span class="status-board__panel-note">{{ items.length }} 个分组</span>
      </div>

      <div v-if="loading && items.length === 0" class="status-board__empty">正在加载监控状态…</div>
      <div v-else-if="items.length === 0" class="status-board__empty">暂无可展示的分组</div>
      <div v-else class="status-board__rows">
        <button
          v-for="item in items"
          :key="item.id"
          type="button"
          class="status-row"
          @click="emit('cardClick', item)"
          @mouseenter="emit('cardHover', item)"
        >
          <span class="status-row__name">
            <span class="status-row__provider" :class="providerClass(item.provider)">
              <ProviderIcon :provider="item.provider" :size="18" />
            </span>
            <span>
              <strong>{{ item.group_name || item.name }}</strong>
            </span>
          </span>

          <span class="status-row__timeline" aria-label="最近 5 小时状态">
            <span
              v-for="(point, index) in timelinePoints(item)"
              :key="`${item.id}-${index}`"
              class="status-cell"
              :class="statusCellClass(point.status)"
            >
              <span v-if="point.checked_at" class="status-cell__tooltip">
                <strong>{{ formatCheckedAt(point.checked_at) }}</strong>
                <span>状态：{{ statusLabel(point.status) }}</span>
                <span>延迟：{{ formatLatency(point.latency_ms) }} ms</span>
                <span>可用性：{{ formatAvailability(item) }}</span>
              </span>
            </span>
          </span>

          <span class="status-row__metrics">
            <span class="status-row__metric status-row__metric--status">
              <i :class="statusDotClass(item.primary_status)"></i>
              <strong>{{ statusLabel(item.primary_status) }}</strong>
            </span>
            <span class="status-row__metric">
              <strong>{{ formatAvailability(item) }}</strong>
              <small>可用性</small>
            </span>
            <span class="status-row__metric">
              <strong>{{ formatLatency(item.primary_latency_ms) }} ms</strong>
              <small>延迟</small>
            </span>
          </span>
        </button>
      </div>

      <div class="status-board__legend">
        <span><i class="green"></i>正常</span>
        <span><i class="yellow"></i>降级</span>
        <span><i class="red"></i>失败</span>
        <span class="status-board__legend-range">← 过去 5 小时　现在 →</span>
      </div>
    </div>

    <div v-else class="probe-board">
      <div class="probe-board__summary">
        <div>
          <h2>探针状态</h2>
          <p>管理员可查看每个分组的独立探针、接口和最近一次探测结果</p>
        </div>
        <span>{{ probeHealthyCount }} / {{ probeTotalCount }} 正常</span>
      </div>

      <div v-if="items.length === 0" class="status-board__empty">暂无可展示的分组</div>
      <div v-else class="probe-board__groups">
        <article v-for="item in items" :key="item.id" class="probe-group">
          <header class="probe-group__head">
            <div class="probe-group__title">
              <span class="status-row__provider" :class="providerClass(item.provider)">
                <ProviderIcon :provider="item.provider" :size="18" />
              </span>
              <div>
                <h3>{{ item.group_name || item.name }}</h3>
              </div>
            </div>
            <span class="probe-group__count">{{ probesFor(item.id).length }} 个探针</span>
          </header>

          <div v-if="probeLoading[item.id]" class="probe-group__empty">正在加载探针状态…</div>
          <div v-else-if="probesFor(item.id).length === 0" class="probe-group__empty">暂无探针数据</div>
          <div v-else class="probe-list">
            <div v-for="(probe, index) in probesFor(item.id)" :key="`${probe.endpoint}-${index}`" class="probe-item">
              <div class="probe-item__identity">
                <span class="probe-item__index">{{ String(index + 1).padStart(2, '0') }}</span>
                <strong>{{ probe.name || `探针 ${index + 1}` }}</strong>
              </div>
              <div class="probe-item__endpoint" :title="probe.endpoint">
                <span>{{ probe.endpoint || '未配置接口' }}</span>
                <small>API Key：{{ probe.api_key || probe.api_key_masked || '未返回' }}</small>
              </div>
              <div class="probe-item__latency">
                <strong>{{ formatLatency(probe.latency_ms) }} ms</strong>
                <small>延迟</small>
              </div>
              <span class="probe-item__status" :class="probeStatusClass(probe.status || '')">{{ statusLabel(probe.status || '') }}</span>
              <span class="probe-item__enabled" :class="{ off: !probe.enabled }">{{ probe.enabled ? '已启用' : '已停用' }}</span>
            </div>
          </div>

          <footer v-if="!probeLoading[item.id] && probesFor(item.id).length" class="probe-group__foot">
            <span>探针总体状态</span>
            <strong :class="probeStatusClass(probeOverallStatus(item.id))">{{ statusLabel(probeOverallStatus(item.id)) }}</strong>
          </footer>
        </article>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import type { MonitorProbe, MonitorStatus, Provider } from '@/api/admin/channelMonitor'
import type { MonitorTimelinePoint, UserMonitorDetail, UserMonitorView } from '@/api/channelMonitor'
import ProviderIcon from './ProviderIcon.vue'
import type { MonitorWindow } from './MonitorHero.vue'
import { useChannelMonitorFormat } from '@/composables/useChannelMonitorFormat'

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
}>()

const { statusLabel, formatLatency, formatPercent } = useChannelMonitorFormat()
const activeView = ref<'groups' | 'probes'>('groups')
const windowOptions: { value: MonitorWindow; label: string }[] = [
  { value: '7d', label: '7 天' },
  { value: '15d', label: '15 天' },
  { value: '30d', label: '30 天' },
]

const probeTotalCount = computed(() => props.items.reduce((sum, item) => sum + probesFor(item.id).length, 0))
const probeHealthyCount = computed(() => props.items.reduce((sum, item) => {
  return sum + probesFor(item.id).filter(probe => probe.enabled && (!probe.status || probe.status === 'operational')).length
}, 0))

function selectView(view: 'groups' | 'probes') {
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
  const value = availabilityValue(item)
  return value == null ? '-' : formatPercent(value)
}

function timelinePoints(item: UserMonitorView): MonitorTimelinePoint[] {
  const points = [...(item.timeline ?? [])].slice(0, 60).reverse()
  while (points.length < 60) points.unshift({ status: '' as MonitorStatus, latency_ms: null, ping_latency_ms: null, checked_at: '' })
  return points
}

function statusCellClass(status: MonitorStatus | ''): string {
  if (status === 'failed' || status === 'error') return 'red'
  if (status === 'degraded') return 'yellow'
  if (status === 'operational') return 'green'
  return 'empty'
}

function statusDotClass(status: MonitorStatus | ''): string {
  return `status-dot status-dot--${statusCellClass(status)}`
}

function probeStatusClass(status: MonitorStatus | ''): string {
  const tone = statusCellClass(status)
  return tone === 'empty' ? '' : `probe-status--${tone}`
}

function probeOverallStatus(id: number): MonitorStatus | '' {
  const statuses = probesFor(id).filter(probe => probe.enabled).map(probe => probe.status || '')
  if (statuses.includes('failed')) return 'failed'
  if (statuses.includes('error')) return 'error'
  if (statuses.includes('degraded')) return 'degraded'
  if (statuses.includes('operational')) return 'operational'
  return ''
}

function formatCheckedAt(value: string): string {
  if (!value) return '未采样'
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? value : date.toLocaleString('zh-CN', { month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' })
}

function providerClass(provider: Provider | string): string {
  return `provider-${provider}`
}
</script>

<style scoped>
.status-board { --canvas: #f5f7fb; --surface: #fff; --line: #e5e9f0; --ink: #172033; --muted: #667085; --navy: #172a46; --cyan: #1ca6a8; --green: #168653; --green-soft: #e9f9f1; --yellow: #d89a22; --yellow-soft: #fff3df; --red: #c74d58; --red-soft: #fdecee; color: var(--ink); }
.status-board__hero { position: relative; display: flex; min-height: 248px; margin: 0 0 22px; overflow: hidden; align-items: center; justify-content: center; padding: 34px 32px; border-radius: 16px; background: linear-gradient(115deg, #142842 0%, #1b4760 52%, #168f92 100%); color: #fff; text-align: center; box-shadow: 0 20px 42px rgba(18,58,81,.2); }
.status-board__hero-content { position: relative; z-index: 1; }
.status-board__hero h1 { font-size: 32px; line-height: 1.2; }
.status-board__hero p { margin-top: 9px; color: rgba(255,255,255,.72); font-size: 13px; }
.status-board__eyebrow { display: inline-flex; align-items: center; gap: 7px; margin-bottom: 12px; padding: 5px 10px; border: 1px solid rgba(199,232,109,.28); border-radius: 999px; background: rgba(255,255,255,.08); color: rgba(255,255,255,.82); font-size: 11px; font-weight: 750; }
.status-board__eyebrow i { width: 7px; height: 7px; border-radius: 50%; background: #c7e86d; box-shadow: 0 0 0 4px rgba(199,232,109,.14); }
.status-board__hero-stat { display: inline-flex; flex-direction: column; align-items: center; gap: 2px; margin-top: 22px; min-width: 140px; }
.status-board__hero-stat span { color: rgba(255,255,255,.58); font-size: 10px; }
.status-board__hero-stat strong { font: 700 19px ui-monospace, monospace; }
.status-board__hero-stat small { color: #c7e86d; font-size: 10px; }
.status-board__hero-orbit { position: absolute; left: 50%; top: 50%; width: 164px; height: 164px; border: 1px solid rgba(199,232,109,.48); border-radius: 50%; opacity: .2; transform: translate(-50%, -50%); }
.status-board__hero-orbit span { position: absolute; inset: 27px; border: 1px solid rgba(255,255,255,.19); border-radius: 50%; }
.status-board__toolbar { display: flex; align-items: center; justify-content: space-between; gap: 14px; margin-bottom: 18px; border-bottom: 1px solid var(--line); }
.status-board__views, .status-board__actions, .status-board__windows { display: flex; align-items: center; gap: 8px; }
.status-board__views button { padding: 0 14px 11px; border: 0; background: transparent; color: #778398; font-size: 12px; font-weight: 700; }
.status-board__views button.active { color: var(--navy); box-shadow: inset 0 -2px var(--cyan); }
.status-board__windows { padding: 3px; border: 1px solid var(--line); border-radius: 9px; background: #f8fafc; }
.status-board__windows button { padding: 4px 9px; border: 0; border-radius: 6px; background: transparent; color: #7b8796; font-size: 10px; }
.status-board__windows button.active { background: #fff; color: var(--navy); box-shadow: 0 1px 4px rgba(23,42,70,.1); font-weight: 750; }
.status-board__panel { overflow: hidden; border: 1px solid var(--line); border-radius: 13px; background: var(--surface); box-shadow: 0 18px 48px rgba(23,42,70,.08); }
.status-board__panel-head { display: flex; align-items: center; justify-content: space-between; gap: 12px; padding: 17px 19px; border-bottom: 1px solid var(--line); }
.status-board__panel-head h2, .probe-board__summary h2 { font-size: 14px; }
.status-board__panel-head p, .probe-board__summary p { margin-top: 3px; color: var(--muted); font-size: 11px; }
.status-board__panel-note { color: #9aa5b4; font-size: 11px; }
.status-board__rows { padding: 0 19px; }
.status-row { display: grid; grid-template-columns: 190px minmax(0, 1fr) 220px; gap: 14px; align-items: center; width: 100%; min-height: 92px; border: 0; border-bottom: 1px solid #eff2f6; background: transparent; text-align: left; }
.status-row:last-child { border-bottom: 0; }
.status-row__name, .probe-group__title { display: flex; min-width: 0; align-items: center; gap: 11px; }
.status-row__provider { display: grid; width: 36px; height: 36px; flex: 0 0 auto; place-items: center; border-radius: 11px; }
.provider-openai { background: #e7f7ef; color: #07885f; }.provider-anthropic { background: #fff0e4; color: #b36b3c; }.provider-gemini { background: #eaf1fb; color: #5677aa; }.provider-grok { background: #eef0f2; color: #59636c; }
.status-row__name strong { display: block; overflow: hidden; color: var(--navy); font-size: 13px; text-overflow: ellipsis; white-space: nowrap; }.status-row__name small { display: block; overflow: hidden; margin-top: 4px; color: var(--muted); font-size: 10px; text-overflow: ellipsis; white-space: nowrap; }
.status-row__timeline { display: grid; grid-template-columns: repeat(60, minmax(7px, 1fr)); gap: 3px; min-width: 0; align-items: center; }
.status-cell { position: relative; height: 34px; min-width: 7px; border-radius: 5px; }.status-cell.green { background: #35b67e; }.status-cell.yellow { background: #e0a327; }.status-cell.red { background: #d85a63; }.status-cell.empty { background: #e9eef0; }
.status-cell__tooltip { position: absolute; bottom: calc(100% + 10px); left: 50%; z-index: 30; display: none; min-width: 130px; padding: 8px 10px; border-radius: 7px; background: #172a46; color: #fff; box-shadow: 0 9px 22px rgba(23,42,70,.2); font-size: 10px; line-height: 1.45; transform: translateX(-50%); white-space: nowrap; }.status-cell__tooltip::after { content: ''; position: absolute; bottom: -5px; left: 50%; width: 8px; height: 8px; background: #172a46; transform: translateX(-50%) rotate(45deg); }.status-cell__tooltip strong, .status-cell__tooltip span { display: block; }.status-cell__tooltip strong { margin-bottom: 3px; color: #c7e86d; }.status-cell:hover { z-index: 20; }.status-cell:hover .status-cell__tooltip { display: block; }
.status-row__metrics { display: grid; grid-template-columns: repeat(3, 1fr); gap: 8px; align-items: center; text-align: center; }.status-row__metric { min-width: 0; }.status-row__metric strong { display: block; overflow: hidden; color: var(--navy); font: 700 13px ui-monospace, monospace; text-overflow: ellipsis; white-space: nowrap; }.status-row__metric small { display: block; margin-top: 3px; color: #9aa5b4; font-size: 9px; }.status-row__metric--status { display: flex; align-items: center; justify-content: center; gap: 5px; }.status-dot { width: 6px; height: 6px; flex: 0 0 auto; border-radius: 50%; }.status-dot--green { background: var(--green); }.status-dot--yellow { background: var(--yellow); }.status-dot--red { background: var(--red); }.status-dot--empty { background: #a9b4bf; }.status-row__metric--status strong { font-family: inherit; font-size: 10px; }
.status-board__legend { display: flex; align-items: center; gap: 14px; padding: 15px 19px; color: #9aa5b4; font-size: 10px; }.status-board__legend span { display: inline-flex; align-items: center; }.status-board__legend i { width: 7px; height: 7px; margin-right: 4px; border-radius: 2px; }.status-board__legend i.green { background: #35b67e; }.status-board__legend i.yellow { background: #e0a327; }.status-board__legend i.red { background: #d85a63; }.status-board__legend-range { margin-left: auto; }
.status-board__empty, .probe-group__empty { padding: 45px 20px; color: var(--muted); font-size: 12px; text-align: center; }
.probe-board__summary { display: flex; align-items: center; justify-content: space-between; gap: 12px; margin-bottom: 15px; }.probe-board__summary > span { padding: 5px 9px; border-radius: 999px; background: var(--green-soft); color: var(--green); font-size: 10px; font-weight: 800; }.probe-board__groups { display: grid; gap: 14px; }.probe-group { overflow: hidden; border: 1px solid var(--line); border-radius: 13px; background: var(--surface); box-shadow: 0 18px 48px rgba(23,42,70,.08); }.probe-group__head { display: flex; align-items: center; justify-content: space-between; gap: 12px; padding: 15px 18px; border-bottom: 1px solid var(--line); }.probe-group__title h3 { overflow: hidden; color: var(--navy); font-size: 13px; text-overflow: ellipsis; white-space: nowrap; }.probe-group__title small, .probe-group__count { color: #9aa5b4; font-size: 10px; }.probe-list { padding: 5px 18px 8px; }.probe-item { display: grid; grid-template-columns: minmax(150px,1.2fr) minmax(180px,1.6fr) 90px 76px 64px; gap: 14px; align-items: center; min-height: 59px; border-bottom: 1px solid #eff2f6; }.probe-item:last-child { border-bottom: 0; }.probe-item__identity { display: flex; min-width: 0; align-items: center; gap: 9px; }.probe-item__index { display: grid; width: 24px; height: 24px; flex: 0 0 auto; place-items: center; border-radius: 7px; background: #eef3f4; color: #62727d; font-size: 10px; font-weight: 800; }.probe-item__identity strong, .probe-item__endpoint span { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }.probe-item__identity strong { color: var(--ink); font-size: 12px; }.probe-item__endpoint span, .probe-item__endpoint small { display: block; }.probe-item__endpoint span { color: var(--muted); font: 10px ui-monospace, monospace; }.probe-item__endpoint small { overflow: hidden; margin-top: 3px; color: #8b98a7; font: 10px ui-monospace, monospace; text-overflow: ellipsis; white-space: nowrap; }.probe-item__latency { text-align: right; }.probe-item__latency strong { display: block; color: var(--navy); font: 700 12px ui-monospace, monospace; }.probe-item__latency small { display: block; margin-top: 2px; color: #9aa5b4; font-size: 9px; }.probe-item__status { justify-self: end; padding: 4px 8px; border-radius: 999px; background: var(--green-soft); color: var(--green); font-size: 10px; font-weight: 800; }.probe-item__status.probe-status--yellow { background: var(--yellow-soft); color: var(--yellow); }.probe-item__status.probe-status--red { background: var(--red-soft); color: var(--red); }.probe-item__enabled { justify-self: end; color: #78909c; font-size: 10px; }.probe-item__enabled.off { color: #b45f64; }.probe-group__foot { display: flex; align-items: center; justify-content: space-between; gap: 8px; padding: 10px 18px; border-top: 1px solid var(--line); color: #8995a3; font-size: 10px; }.probe-group__foot strong { padding: 3px 8px; border-radius: 999px; background: var(--green-soft); color: var(--green); font-size: 10px; }.probe-group__foot strong.probe-status--yellow { background: var(--yellow-soft); color: var(--yellow); }.probe-group__foot strong.probe-status--red { background: var(--red-soft); color: var(--red); }
.probe-item__endpoint { min-width: 0; }
.probe-item__endpoint span, .probe-item__endpoint small { overflow: visible; overflow-wrap: anywhere; word-break: break-word; white-space: normal; text-overflow: clip; }
@media (max-width: 900px) { .status-row { grid-template-columns: 165px minmax(0,1fr) 190px; }.probe-item { grid-template-columns: minmax(130px,1fr) minmax(160px,1.4fr) 80px 70px 60px; gap: 8px; } }
@media (max-width: 700px) { .status-board__toolbar { align-items: flex-start; flex-direction: column; }.status-board__actions { width: 100%; justify-content: space-between; }.status-row { grid-template-columns: 150px minmax(620px,1fr) 180px; min-width: 980px; }.status-board__panel { overflow-x: auto; }.status-board__panel-head, .status-board__legend { min-width: 980px; }.probe-board { overflow-x: auto; }.probe-board__groups { min-width: 760px; }.status-board__hero { min-height: 270px; }.status-board__hero h1 { font-size: 26px; } }
</style>
