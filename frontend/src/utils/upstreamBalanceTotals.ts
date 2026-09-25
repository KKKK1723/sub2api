import type { UpstreamBalanceOverviewItem } from '@/api/admin/upstreamBalances'

export function summarizeUpstreamBalances(items: UpstreamBalanceOverviewItem[]) {
  const totals = new Map<string, number>()
  let included = 0
  for (const item of items) {
    const unit = item.unit?.trim().toUpperCase()
    if (item.status !== 'ok' || typeof item.balance !== 'number' || !Number.isFinite(item.balance) || !unit) continue
    totals.set(unit, (totals.get(unit) ?? 0) + item.balance)
    included++
  }
  return {
    totals: Array.from(totals, ([unit, balance]) => ({ unit, balance })).sort((a, b) => a.unit.localeCompare(b.unit)),
    included,
    excluded: items.length - included
  }
}
