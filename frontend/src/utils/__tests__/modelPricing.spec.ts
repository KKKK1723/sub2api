import { describe, expect, it } from 'vitest'
import type { ModelPricingGroup, ModelPricingInterval, ModelPricingItem } from '@/api/modelPricing'
import { effectiveGroupRate, estimateModelCost, type PricingUsage } from '../modelPricing'

const group: ModelPricingGroup = {
  id: 1, name: '测试分组', description: '', platform: 'openai', subscription_type: 'subscription',
  rate_multiplier: 2, user_rate_multiplier: 0.5, peak_rate_enabled: true,
  peak_start: '14:00', peak_end: '18:00', peak_rate_multiplier: 2, peak_factor_now: 2,
  is_exclusive: false, model_source: 'group', models: [],
}
const model: ModelPricingItem = {
  name: '测试模型', billing_mode: 'token', pricing_source: 'channel', priced: true,
  input_price: 2e-6, output_price: 8e-6, cache_write_price: 3e-6, cache_read_price: 0.5e-6,
  per_request_price: 0.04,
}
const usage: PricingUsage = { input: 1000, output: 100, cacheWrite: 0, cacheRead: 0, count: 1, tier: '' }
const interval: ModelPricingInterval = {
  min_tokens: 0, max_tokens: 2000, input_price: 3e-6, output_price: 6e-6,
  cache_write_price: null, cache_read_price: null, per_request_price: null,
}

describe('模型价格费用估算', () => {
  it('优先使用专属倍率，并保留零倍率', () => {
    expect(effectiveGroupRate(group)).toBe(0.5)
    expect(effectiveGroupRate({ ...group, user_rate_multiplier: null })).toBe(2)
    expect(effectiveGroupRate({ ...group, user_rate_multiplier: 0 })).toBe(0)
    expect(estimateModelCost(model, { ...group, user_rate_multiplier: 0 }, usage)?.actual).toBe(0)
  })

  it('输入、输出和缓存都乘适用的高峰及分组倍率', () => {
    const result = estimateModelCost(model, { ...group, user_rate_multiplier: null }, { ...usage, cacheWrite: 100, cacheRead: 100 })!
    expect(result.base).toBeCloseTo(0.00315, 12)
    expect(result.rate).toBe(4)
    expect(result.actual).toBeCloseTo(0.0126, 12)
  })

  it.each(['per_request', 'image'])('%s 按数量计费，不使用高峰倍率', mode => {
    const result = estimateModelCost({ ...model, billing_mode: mode }, group, { ...usage, count: 3 })!
    expect(result.base).toBeCloseTo(0.12, 12)
    expect(result.actual).toBeCloseTo(0.06, 12)
    expect(result.rate).toBe(0.5)
  })

  it('按总输入上下文匹配区间，包含缓存但不包含输出', () => {
    const higher = { ...interval, min_tokens: 2000, max_tokens: null, input_price: 4e-6 }
    const tiered = { ...model, intervals: [interval, higher] }
    expect(estimateModelCost(tiered, group, { ...usage, output: 100000, cacheRead: 1000 })?.interval).toBe(interval)
    expect(estimateModelCost(tiered, group, { ...usage, cacheWrite: 500, cacheRead: 501 })?.interval).toBe(higher)
    expect(estimateModelCost(tiered, group, { ...usage, input: 0 })?.interval).toBeUndefined()
  })

  it('区间空价格按零计算，区间未命中时回退基础价格', () => {
    const tiered = { ...model, intervals: [interval] }
    expect(estimateModelCost(tiered, group, { ...usage, cacheWrite: 100, cacheRead: 100 })?.base).toBeCloseTo(0.0036, 12)
    expect(estimateModelCost(tiered, group, { ...usage, input: 3000 })?.base).toBeCloseTo(0.0068, 12)
  })

  it('仅在严格超过长上下文阈值时，对所有输入侧费用和输出应用相应倍率', () => {
    const longModel = { ...model, long_context_threshold: 2000, long_context_input_multiplier: 2, long_context_output_multiplier: 1.5 }
    const atBoundary = estimateModelCost(longModel, group, { ...usage, cacheWrite: 500, cacheRead: 500 })!
    expect(atBoundary.longContext).toBe(false)
    expect(atBoundary.base).toBeCloseTo(0.00455, 12)
    const aboveBoundary = estimateModelCost(longModel, group, { ...usage, cacheWrite: 500, cacheRead: 501 })!
    expect(aboveBoundary.longContext).toBe(true)
    expect(aboveBoundary.base).toBeCloseTo(0.008701, 12)
  })

  it('有区间时，即使未命中区间也不叠加长上下文倍率', () => {
    const result = estimateModelCost({ ...model, intervals: [interval], long_context_threshold: 100, long_context_input_multiplier: 2, long_context_output_multiplier: 2 }, group, { ...usage, input: 3000 })!
    expect(result.longContext).toBe(false)
    expect(result.base).toBeCloseTo(0.0068, 12)
  })

  it('规格价格优先；缺失或零价按上下文、默认价格依次回退', () => {
    const contextTier = { ...interval, per_request_price: 0.08 }
    const sizeTier = { ...interval, tier_label: '4K', min_tokens: 5000, per_request_price: 0.2 }
    const requestModel = { ...model, billing_mode: 'image', intervals: [contextTier, sizeTier] }
    expect(estimateModelCost(requestModel, group, { ...usage, tier: '4K' })?.base).toBe(0.2)
    expect(estimateModelCost(requestModel, group, usage)?.base).toBe(0.08)
    expect(estimateModelCost(requestModel, group, { ...usage, input: 0 })?.base).toBe(0.04)
    expect(estimateModelCost({ ...requestModel, intervals: [contextTier, { ...sizeTier, per_request_price: 0 }] }, group, { ...usage, tier: '4K' })?.base).toBe(0.08)
  })

  it.each([-1, 0.5, NaN, Infinity, Number.MAX_SAFE_INTEGER + 1])('拒绝非法 Token 数量 %s', input => {
    expect(estimateModelCost(model, group, { ...usage, input })).toBeNull()
  })

  it('拒绝空输入、总上下文溢出、无定价及零次请求', () => {
    expect(estimateModelCost(model, group, { ...usage, input: '' as unknown as number })).toBeNull()
    expect(estimateModelCost(model, group, { ...usage, input: Number.MAX_SAFE_INTEGER, cacheRead: 1 })).toBeNull()
    expect(estimateModelCost({ ...model, priced: false }, group, usage)).toBeNull()
    expect(estimateModelCost({ ...model, billing_mode: 'per_request' }, group, { ...usage, count: 0 })).toBeNull()
  })
})
