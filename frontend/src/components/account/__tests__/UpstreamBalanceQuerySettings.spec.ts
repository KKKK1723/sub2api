import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import UpstreamBalanceQuerySettings from '../UpstreamBalanceQuerySettings.vue'

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({ t: (key: string) => key })
  }
})

describe('UpstreamBalanceQuerySettings', () => {
  it('enables probing and applies the constrained custom defaults', async () => {
    const wrapper = mount(UpstreamBalanceQuerySettings, {
      props: {
        enabled: false,
        query: { preset: 'ccswitch' }
      }
    })

    await wrapper.get('[data-testid="upstream-balance-enabled"]').trigger('click')
    expect(wrapper.emitted('update:enabled')?.at(-1)).toEqual([true])

    await wrapper.setProps({ enabled: true })
    await wrapper.get('[data-testid="upstream-balance-preset"]').setValue('custom')

    expect(wrapper.emitted('update:query')?.at(-1)).toEqual([{
      preset: 'custom',
      method: 'GET',
      url: '{{baseUrl}}/v1/usage',
      auth_type: 'bearer',
      balance_path: 'remaining',
      unit_path: 'unit',
      plan_path: 'planName',
      error_path: 'error',
      unit: 'USD'
    }])
  })

  it('reduces built-in presets to their canonical value', async () => {
    const wrapper = mount(UpstreamBalanceQuerySettings, {
      props: {
        enabled: true,
        query: { preset: 'custom', method: 'POST', url: '{{origin}}/api/usage', balance_path: 'balance' }
      }
    })

    await wrapper.get('[data-testid="upstream-balance-preset"]').setValue('sub2api')
    expect(wrapper.emitted('update:query')?.at(-1)).toEqual([{ preset: 'sub2api' }])
  })

  it('shows legacy ccswitch settings as the compatible built-in preset', () => {
    const wrapper = mount(UpstreamBalanceQuerySettings, {
      props: {
        enabled: true,
        query: { preset: 'ccswitch' }
      }
    })

    const preset = wrapper.get<HTMLSelectElement>('[data-testid="upstream-balance-preset"]')
    expect(preset.element.value).toBe('sub2api')
    expect(preset.findAll('option')).toHaveLength(2)
  })
})
