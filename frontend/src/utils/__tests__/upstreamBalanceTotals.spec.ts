import { describe, expect, it } from 'vitest'
import { summarizeUpstreamBalances } from '../upstreamBalanceTotals'
import type { UpstreamBalanceOverviewItem } from '@/api/admin/upstreamBalances'

const item = (balance?: number, unit = 'USD', status: UpstreamBalanceOverviewItem['status'] = 'ok'): UpstreamBalanceOverviewItem => ({ base_url: 'https://upstream.example', account_id: 1, balance, unit, status })

describe('summarizeUpstreamBalances', () => {
  it('汇总全部有效余额，保留零和欠费，按币种分别计算', () => {
    expect(summarizeUpstreamBalances([item(12.5), item(0), item(-2.5, ' usd '), item(80, 'CNY')])).toEqual({
      totals: [{ unit: 'CNY', balance: 80 }, { unit: 'USD', balance: 10 }], included: 4, excluded: 0
    })
  })
  it('排除失败、未知币种和非有限余额', () => {
    expect(summarizeUpstreamBalances([item(99, 'USD', 'failed'), item(1, 'USD', 'unsupported'), item(), item(NaN), item(Infinity), item(3, '')])).toEqual({ totals: [], included: 0, excluded: 6 })
  })
  it('刷新后的统计不累加旧数据，空列表不伪造余额', () => {
    summarizeUpstreamBalances([item(99)])
    expect(summarizeUpstreamBalances([item(7)]).totals).toEqual([{ unit: 'USD', balance: 7 }])
    expect(summarizeUpstreamBalances([])).toEqual({ totals: [], included: 0, excluded: 0 })
  })
})
