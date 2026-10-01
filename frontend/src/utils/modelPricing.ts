import type { ModelPricingGroup, ModelPricingInterval, ModelPricingItem } from '@/api/modelPricing'

export interface PricingUsage {
  input: number
  output: number
  cacheWrite: number
  cacheRead: number
  count: number
  tier: string
}

export function effectiveGroupRate(group: ModelPricingGroup): number {
  return Math.max(0, group.user_rate_multiplier ?? group.rate_multiplier)
}

export function modelRate(group: ModelPricingGroup, model: ModelPricingItem): number {
  return effectiveGroupRate(group) * (model.billing_mode === 'token' ? group.peak_factor_now : 1)
}

export function tokenPrices(model: ModelPricingItem, interval?: ModelPricingInterval) {
  // 命中区间后，空价格按后端规则归零，不继承基础单价。
  return {
    input: interval ? interval.input_price ?? 0 : model.input_price,
    output: interval ? interval.output_price ?? 0 : model.output_price,
    cacheWrite: interval ? interval.cache_write_price ?? 0 : model.cache_write_price,
    cacheRead: interval ? interval.cache_read_price ?? 0 : model.cache_read_price,
  }
}

export function estimateModelCost(model: ModelPricingItem, group: ModelPricingGroup, usage: PricingUsage) {
  const isToken = model.billing_mode === 'token'
  const quantities = isToken
    ? [usage.input, usage.output, usage.cacheWrite, usage.cacheRead]
    : [usage.input, usage.cacheWrite, usage.cacheRead, usage.count]
  if (!model.priced || quantities.some(value => !Number.isSafeInteger(value) || value < 0) ||
    (!isToken && usage.count < 1)) return null

  const context = usage.input + usage.cacheWrite + usage.cacheRead
  if (!Number.isSafeInteger(context)) return null
  const intervals = model.intervals ?? []
  const matched = intervals.find(interval => context > interval.min_tokens &&
    (interval.max_tokens === null || context <= interval.max_tokens))
  const rate = modelRate(group, model)

  if (!isToken) {
    const tier = usage.tier ? intervals.find(interval => interval.tier_label === usage.tier &&
      interval.per_request_price !== null) : undefined
    // 后端先匹配规格，再匹配上下文；零价表示继续回退到默认单价。
    const interval = tier?.per_request_price ? tier : matched?.per_request_price ? matched : undefined
    const unitPrice = interval?.per_request_price ?? model.per_request_price
    const base = unitPrice * usage.count
    return { base, actual: base * rate, rate, context, interval, longContext: false }
  }

  const prices = tokenPrices(model, matched)
  const longContext = intervals.length === 0 && (model.long_context_threshold ?? 0) > 0 &&
    context > model.long_context_threshold! &&
    ((model.long_context_input_multiplier ?? 1) > 1 || (model.long_context_output_multiplier ?? 1) > 1)
  const inputMultiplier = longContext ? model.long_context_input_multiplier ?? 1 : 1
  const outputMultiplier = longContext ? model.long_context_output_multiplier ?? 1 : 1
  const base = (usage.input * prices.input + usage.cacheWrite * prices.cacheWrite +
    usage.cacheRead * prices.cacheRead) * inputMultiplier + usage.output * prices.output * outputMultiplier
  return { base, actual: base * rate, rate, context, interval: matched, longContext }
}
