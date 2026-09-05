import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, shallowMount } from '@vue/test-utils'
import PaymentView from '../PaymentView.vue'

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

  it('renders the recharge message, contact accounts, notes, and QR images', () => {
    const wrapper = mountPage()

    expect(wrapper.get('h1').text()).toBe('充值与订阅')
    expect(wrapper.get('.intro').text()).toBe('站点小范围运营，充值请联系站长。')
    expect(wrapper.text()).toContain('2052436429')
    expect(wrapper.text()).toContain('Taokkkkkkkboy')
    expect(wrapper.findAll('.contact-note')).toHaveLength(2)
    expect(wrapper.findAll('.contact-note').every(note => note.text() === '备注：站点充值')).toBe(true)
    expect(wrapper.get('img[alt="QQ 联系二维码"]').attributes('src')).toBe('/contact/qq.png')
    expect(wrapper.get('img[alt="微信联系二维码"]').attributes('src')).toBe('/contact/wechat.png')
  })

  it('copies a contact account and shows a confirmation toast', async () => {
    const wrapper = mountPage()
    const clipboard = navigator.clipboard.writeText as ReturnType<typeof vi.fn>

    await wrapper.get('button').trigger('click')
    await flushPromises()

    expect(clipboard).toHaveBeenCalledWith('2052436429')
    expect(wrapper.get('.copy-toast').text()).toContain('已复制：2052436429')
  })
})
