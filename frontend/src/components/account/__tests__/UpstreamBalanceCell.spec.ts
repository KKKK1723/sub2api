import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import UpstreamBalanceCell from '../UpstreamBalanceCell.vue'
import type { Account } from '@/types'

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key,
      locale: { value: 'en-US' }
    })
  }
})

const makeAccount = (overrides: Partial<Account> = {}): Account => ({
  id: 1,
  name: 'upstream',
  platform: 'openai',
  type: 'apikey',
  proxy_id: null,
  concurrency: 1,
  priority: 1,
  status: 'active',
  error_message: null,
  last_used_at: null,
  expires_at: null,
  auto_pause_on_expired: false,
  created_at: '2026-07-28T00:00:00Z',
  updated_at: '2026-07-28T00:00:00Z',
  schedulable: true,
  rate_limited_at: null,
  rate_limit_reset_at: null,
  overload_until: null,
  temp_unschedulable_until: null,
  temp_unschedulable_reason: null,
  session_window_start: null,
  session_window_end: null,
  session_window_status: null,
  ...overrides
})

describe('UpstreamBalanceCell', () => {
  it('shows the latest balance and emits an icon-only refresh command', async () => {
    const wrapper = mount(UpstreamBalanceCell, {
      props: {
        account: makeAccount({
          extra: {
            upstream_balance_query: { preset: 'ccswitch' },
            upstream_balance_probe: {
              status: 'ok',
              data: { balance: 12.34, unit: 'USD', plan_name: 'Pro' },
              received_at: '2026-07-28T00:00:00Z',
              fresh_until: '2026-07-28T02:00:00Z',
              last_attempt_at: '2026-07-28T00:00:00Z',
              next_probe_at: '2026-07-28T01:00:00Z'
            }
          }
        }),
        now: Date.parse('2026-07-28T00:30:00Z')
      }
    })

    expect(wrapper.get('[data-testid="upstream-balance-value"]').text()).toBe('12.34 USD')
    expect(wrapper.get('[data-testid="upstream-balance-probe"]').text()).toBe('')
    await wrapper.get('[data-testid="upstream-balance-probe"]').trigger('click')
    expect(wrapper.emitted('probe')).toHaveLength(1)
  })

  it('stays neutral for accounts without a configured balance query', () => {
    const wrapper = mount(UpstreamBalanceCell, {
      props: { account: makeAccount(), now: Date.now() }
    })

    expect(wrapper.text()).toBe('-')
    expect(wrapper.find('button').exists()).toBe(false)
  })
})
