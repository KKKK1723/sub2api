import { CONCRETE_PLATFORM_OPTIONS } from '@/constants/platforms'
import type { UserMonitorView } from '@/api/channelMonitor'

// 监控页优先展示 OpenAI，其余平台沿用原顺序。
const MONITOR_PLATFORM_OPTIONS = [
  ...CONCRETE_PLATFORM_OPTIONS.filter(platform => platform.value === 'openai'),
  ...CONCRETE_PLATFORM_OPTIONS.filter(platform => platform.value !== 'openai')
]

// 按平台归类，保留平台内原排序，并保留未知平台的数据。
export function groupMonitorPlatforms(items: UserMonitorView[]) {
  const grouped = new Map<string, UserMonitorView[]>()
  for (const item of items) {
    const group = grouped.get(item.provider) ?? []
    group.push(item)
    grouped.set(item.provider, group)
  }
  const sections = MONITOR_PLATFORM_OPTIONS.flatMap(platform => {
    const members = grouped.get(platform.value)
    grouped.delete(platform.value)
    return members?.length ? [{ ...platform, items: members }] : []
  })
  return [...sections, ...Array.from(grouped, ([value, members]) => ({ value, label: value, items: members }))]
}
