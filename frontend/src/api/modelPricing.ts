/**
 * User Model Pricing API (non-admin)
 * 「模型价格」页：按用户可用分组返回模型列表与基础定价（倍率前，USD / token）。
 */

import { apiClient } from './client'

export interface ModelPricingInterval {
  min_tokens: number
  max_tokens: number | null
  tier_label?: string
  input_price: number | null
  output_price: number | null
  cache_write_price: number | null
  cache_read_price: number | null
  per_request_price: number | null
}

export interface ModelPricingItem {
  name: string
  /** token / per_request / image / video */
  billing_mode: string
  /** group / channel / litellm / fallback */
  pricing_source: string
  priced: boolean
  input_price: number
  output_price: number
  cache_write_price: number
  cache_read_price: number
  per_request_price: number
  intervals?: ModelPricingInterval[]
  long_context_threshold?: number
  long_context_input_multiplier?: number
  long_context_output_multiplier?: number
  context_window?: number
  reasoning?: boolean
  vision?: boolean
}

export interface ModelPricingGroup {
  id: number
  name: string
  description: string
  platform: string
  subscription_type: string
  rate_multiplier: number
  /** 用户专属倍率，覆盖 rate_multiplier；null 表示无专属倍率。 */
  user_rate_multiplier: number | null
  peak_rate_enabled: boolean
  peak_start: string
  peak_end: string
  peak_rate_multiplier: number
  /** 服务端时区下当前时刻的高峰因子，非高峰为 1。 */
  peak_factor_now: number
  is_exclusive: boolean
  /** group / accounts / default */
  model_source: string
  models: ModelPricingItem[]
}

export async function list(): Promise<ModelPricingGroup[]> {
  const { data } = await apiClient.get<ModelPricingGroup[] | null>('/model-pricing')
  return data || []
}

export const modelPricingAPI = { list }

export default modelPricingAPI
