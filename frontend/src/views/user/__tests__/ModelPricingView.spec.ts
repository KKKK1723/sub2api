import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount, type VueWrapper } from '@vue/test-utils'
import type { ModelPricingGroup, ModelPricingItem } from '@/api/modelPricing'
import ModelPricingView from '../ModelPricingView.vue'

const { list, showError, replace, copyToClipboard, query } = vi.hoisted(() => ({
  list: vi.fn(), showError: vi.fn(), replace: vi.fn(), copyToClipboard: vi.fn(), query: {} as Record<string, string>,
}))
vi.mock('@/api/modelPricing', () => ({ default: { list } }))
vi.mock('vue-router', () => ({ useRoute: () => ({ query }), useRouter: () => ({ replace }) }))
vi.mock('@/composables/useClipboard', () => ({ useClipboard: () => ({ copyToClipboard }) }))
vi.mock('@/stores/app', () => ({
  useAppStore: () => ({ showError, cachedPublicSettings: { server_utc_offset: '+08:00' } }),
}))
vi.mock('vue-i18n', async importOriginal => ({
  ...await importOriginal<typeof import('vue-i18n')>(),
  useI18n: () => ({ t: (key: string, params?: Record<string, unknown>) => params ? `${key} ${Object.values(params).join(' ')}` : key }),
}))

const tokenModel: ModelPricingItem = {
  name: 'test-token', billing_mode: 'token', pricing_source: 'litellm', priced: true,
  input_price: 2e-6, output_price: 8e-6, cache_write_price: 3e-6, cache_read_price: 0.5e-6, per_request_price: 0,
  context_window: 200000, reasoning: true, vision: true,
}
const group: ModelPricingGroup = {
  id: 1, name: '分组一', description: '', platform: 'openai', subscription_type: 'subscription',
  rate_multiplier: 2, user_rate_multiplier: 0.5, peak_rate_enabled: true, peak_start: '14:00', peak_end: '18:00',
  peak_rate_multiplier: 2, peak_factor_now: 2, is_exclusive: false, model_source: 'group',
  models: [tokenModel, { ...tokenModel, name: 'unknown-model', priced: false }],
}
const imageModel: ModelPricingItem = {
  ...tokenModel, name: 'test-image', billing_mode: 'image', per_request_price: 0.04,
  intervals: [{ min_tokens: 0, max_tokens: null, tier_label: '4K', input_price: null, output_price: null,
    cache_write_price: null, cache_read_price: null, per_request_price: 0.2 }],
}
const secondGroup = { ...group, id: 2, name: '分组二', platform: 'gemini', models: [imageModel] }
let wrapper: VueWrapper
function render() {
  wrapper = mount(ModelPricingView, {
    global: { stubs: { AppLayout: { template: '<main><slot /></main>' }, Icon: true,
      LoadingSpinner: { template: '<div data-testid="loading" />' }, teleport: true } },
  })
  return wrapper
}
function estimateButton(name = 'test-token') {
  return wrapper.get(`button[aria-label="modelPricing.calculateModel ${name}"]`)
}
function peakCheckbox() { return wrapper.findAll('input[type=checkbox]')[1] }
beforeEach(() => {
  vi.clearAllMocks()
  for (const key of Object.keys(query)) delete query[key]
  replace.mockResolvedValue(undefined)
  copyToClipboard.mockResolvedValue(true)
  list.mockResolvedValue([group, secondGroup])
})
afterEach(() => wrapper?.unmount())

describe('模型价格参考布局与交互', () => {
  it('展示平台标签、分组卡片、上下文和能力标记，默认使用普通倍率', async () => {
    render()
    expect(wrapper.find('[data-testid="loading"]').exists()).toBe(true)
    await flushPromises()
    expect(wrapper.findAll('[role=tab]')).toHaveLength(2)
    expect(wrapper.get('[role=radio][aria-checked=true]').text()).toContain('分组一')
    expect(wrapper.find('select#mp-group').exists()).toBe(false)
    expect(wrapper.text()).toContain('14:00-18:00 ×2 (UTC+08:00)')
    expect(wrapper.get('tbody').text()).toContain('200K')
    expect(wrapper.get('tbody').text()).toContain('modelPricing.reasoning')
    expect(wrapper.get('tbody').text()).toContain('$1')
    expect(wrapper.get('tbody').text()).toContain('modelPricing.unpricedShort')
  })

  it('支持 g 参数直接定位平台及分组，同平台分组按专属倍率升序', async () => {
    query.g = '2'
    list.mockResolvedValueOnce([group, secondGroup, { ...group, id: 3, name: '低倍率', user_rate_multiplier: 0 }])
    render()
    await flushPromises()
    expect(wrapper.get('[role=tab][aria-selected=true]').text()).toContain('Gemini')
    expect(wrapper.get('[role=radio][aria-checked=true]').attributes('data-gid')).toBe('2')
    await wrapper.findAll('[role=tab]')[0].trigger('click')
    await flushPromises()
    expect(wrapper.findAll('[role=radio]').map(item => item.attributes('data-gid'))).toEqual(['3', '1'])
    expect(wrapper.get('[role=radio][aria-checked=true]').attributes('data-gid')).toBe('3')
    expect(replace).toHaveBeenCalledWith({ query: { g: '3' } })
  })

  it('搜索和切换平台更新模型列表，保留其他查询参数', async () => {
    query.from = 'test'
    render()
    await flushPromises()
    await wrapper.get('input[type=search]').setValue('absent')
    expect(wrapper.text()).toContain('modelPricing.noMatches')
    await wrapper.findAll('[role=tab]')[1].trigger('click')
    await flushPromises()
    expect(wrapper.get('tbody').text()).toContain('test-image')
    expect(wrapper.get('tbody').text()).not.toContain('test-token')
    expect(replace).toHaveBeenCalledWith({ query: { from: 'test', g: '2' } })
  })

  it('原价、高峰和 1M/1K 开关改变价格展示，不改变后端费率', async () => {
    render()
    await flushPromises()
    const row = () => wrapper.get('tbody tr')
    expect(row().text()).toContain('modelPricing.official')
    await wrapper.findAll('input[type=checkbox]')[0].setValue(false)
    expect(row().text()).not.toContain('modelPricing.official')
    await peakCheckbox().setValue(true)
    expect(row().text()).toContain('$2')
    await wrapper.findAll('button').find(button => button.text() === '1K')!.trigger('click')
    expect(row().text()).toContain('$0.002')
    expect(group.user_rate_multiplier).toBe(0.5)
  })

  it('单价排序切换方向，未定价模型保持在末尾', async () => {
    list.mockResolvedValueOnce([{ ...group, models: [...group.models, { ...tokenModel, name: 'cheaper', input_price: 1e-6 }] }])
    render()
    await flushPromises()
    const sortButton = wrapper.get('thead').findAll('button').find(button => button.text() === 'modelPricing.input')!
    await sortButton.trigger('click')
    expect(wrapper.findAll('tbody tr')[0].text()).toContain('cheaper')
    await sortButton.trigger('click')
    expect(wrapper.findAll('tbody tr')[0].text()).toContain('test-token')
    expect(wrapper.findAll('tbody tr').at(-1)!.text()).toContain('unknown-model')
  })

  it('右侧试算支持请求数，按单次上下文判断长上下文，并可用 Escape 关闭', async () => {
    list.mockResolvedValueOnce([{ ...group, models: [{ ...tokenModel, long_context_threshold: 2000, long_context_input_multiplier: 2, long_context_output_multiplier: 1.5 }] }])
    render()
    await flushPromises()
    await estimateButton().trigger('click')
    expect(wrapper.get('[role=dialog]').element.tagName).toBe('ASIDE')
    await wrapper.get('#mp-input').setValue(1000)
    await wrapper.get('#mp-output').setValue(100)
    await wrapper.get('#mp-count').setValue(3)
    expect(wrapper.get('[data-testid=actual-cost]').text()).toBe('$0.0042')
    expect(wrapper.get('[role=dialog]').text()).not.toContain('modelPricing.longContextApplied')
    await wrapper.get('#mp-cacheRead').setValue(1001)
    expect(wrapper.get('[data-testid=actual-cost]').text()).toBe('$0.0093015')
    expect(wrapper.get('[role=dialog]').text()).toContain('modelPricing.longContextApplied')
    await wrapper.get('[role=dialog]').trigger('keydown', { key: 'Escape' })
    expect(wrapper.find('[role=dialog]').exists()).toBe(false)
    expect(document.body.classList.contains('modal-open')).toBe(false)
  })

  it('试算拒绝空值、负数和零次请求，未定价模型不提供试算', async () => {
    render()
    await flushPromises()
    expect(wrapper.find('button[aria-label="modelPricing.calculateModel unknown-model"]').exists()).toBe(false)
    await estimateButton().trigger('click')
    for (const value of ['', -1]) {
      await wrapper.get('#mp-input').setValue(value)
      expect(wrapper.find('[data-testid=actual-cost]').exists()).toBe(false)
    }
    await wrapper.get('#mp-input').setValue(1000)
    await wrapper.get('#mp-count').setValue(0)
    expect(wrapper.text()).toContain('modelPricing.invalidInput')
    expect(wrapper.find('[data-testid=actual-cost]').exists()).toBe(false)
  })

  it('图片规格和数量不受高峰开关及计价单位影响', async () => {
    query.g = '2'
    render()
    await flushPromises()
    await peakCheckbox().setValue(true)
    await wrapper.findAll('button').find(button => button.text() === '1K')!.trigger('click')
    await estimateButton('test-image').trigger('click')
    expect(wrapper.find('#mp-output').exists()).toBe(false)
    await wrapper.get('#mp-tier').setValue('4K')
    await wrapper.get('#mp-count').setValue(2)
    expect(wrapper.get('[data-testid=actual-cost]').text()).toBe('$0.2')
  })

  it('复制模型名称，切换到无高峰分组时清除高峰试算状态', async () => {
    list.mockResolvedValueOnce([group, { ...secondGroup, peak_rate_enabled: false }])
    render()
    await flushPromises()
    await wrapper.get('button[aria-label="modelPricing.copyModel test-token"]').trigger('click')
    expect(copyToClipboard).toHaveBeenCalledWith('test-token')
    await peakCheckbox().setValue(true)
    await wrapper.findAll('[role=tab]')[1].trigger('click')
    await flushPromises()
    expect((peakCheckbox().element as HTMLInputElement).checked).toBe(false)
    expect(peakCheckbox().attributes('disabled')).toBeDefined()
  })

  it('支持平台键盘切换及分组选择', async () => {
    render()
    await flushPromises()
    await wrapper.get('[role=tab][aria-selected=true]').trigger('keydown', { key: 'ArrowRight' })
    await flushPromises()
    expect(wrapper.get('[role=tab][aria-selected=true]').text()).toContain('Gemini')
  })

  it('区分无分组和分组无模型两种空态', async () => {
    list.mockResolvedValueOnce([])
    render()
    await flushPromises()
    expect(wrapper.text()).toContain('modelPricing.emptyHint')
    wrapper.unmount()
    list.mockResolvedValueOnce([{ ...group, models: [] }])
    render()
    await flushPromises()
    expect(wrapper.text()).toContain('modelPricing.noModels')
    expect(wrapper.find('tbody').exists()).toBe(false)
  })

  it('请求失败可重试，不会被错误显示成无分组', async () => {
    list.mockRejectedValueOnce(new Error('network'))
    render()
    await flushPromises()
    expect(showError).toHaveBeenCalled()
    expect(wrapper.text()).not.toContain('modelPricing.emptyHint')
    await wrapper.findAll('button').find(button => button.text() === 'modelPricing.retry')!.trigger('click')
    await flushPromises()
    expect(list).toHaveBeenCalledTimes(2)
    expect(wrapper.text()).toContain('test-token')
  })
})
