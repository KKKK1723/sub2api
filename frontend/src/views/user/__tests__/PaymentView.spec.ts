import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, shallowMount } from '@vue/test-utils'
import PaymentView from '../PaymentView.vue'

vi.mock('vue-i18n', async importOriginal => ({
  ...await importOriginal<typeof import('vue-i18n')>(),
  useI18n: () => ({ t: (key: string, params?: Record<string, unknown>) => params ? `${key} ${Object.values(params).join(' ')}` : key }),
}))

describe('PaymentView contact recharge page', () => {
  beforeEach(() => {
    Object.defineProperty(window, 'isSecureContext', {
      configurable: true,
      value: true,
    })
    Object.defineProperty(navigator, 'clipboard', {
      configurable: true,
      value: { writeText: vi.fn().mockResolvedValue(undefined) },
    })
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  function mountPage() {
    return shallowMount(PaymentView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
        },
      },
    })
  }

  it('renders the wechat-only recharge message and QR image', () => {
    const wrapper = mountPage()

    expect(wrapper.get('h1').text()).toContain('payment.recharge.headlineAny')
    expect(wrapper.get('h1').text()).toContain('payment.recharge.headlineRefund')
    expect(wrapper.findAll('.promise')).toHaveLength(2)
    expect(wrapper.text()).toContain('Taokkkkkkkboy')
    expect(wrapper.text()).not.toContain('2052436429')
    expect(wrapper.findAll('img')).toHaveLength(1)
    expect(wrapper.get('img').attributes('src')).toBe('/contact/wechat.png')
  })

  it('copies the wechat account and shows a confirmation toast', async () => {
    const wrapper = mountPage()
    const clipboard = navigator.clipboard.writeText as ReturnType<typeof vi.fn>

    await wrapper.get('.copy-button').trigger('click')
    await flushPromises()

    expect(clipboard).toHaveBeenCalledWith('Taokkkkkkkboy')
    expect(wrapper.get('.copy-toast').text()).toBe('payment.recharge.copied Taokkkkkkkboy')
  })
})
