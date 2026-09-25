import { describe, expect, it } from 'vitest'
import {
  OPENAI_COMPATIBLE_PROVIDER_PRESETS,
  resolveOpenAICompatibleProvider
} from '../openAICompatibleProvider'

describe('OpenAI 兼容供应商预设', () => {
  it('提供 DeepSeek 和智谱的兼容地址及模型', () => {
    expect(OPENAI_COMPATIBLE_PROVIDER_PRESETS.deepseek.baseUrl).toBe('https://api.deepseek.com')
    expect(OPENAI_COMPATIBLE_PROVIDER_PRESETS.deepseek.models).toContain('deepseek-reasoner')
    expect(OPENAI_COMPATIBLE_PROVIDER_PRESETS.zhipu.baseUrl).toBe(
      'https://open.bigmodel.cn/api/paas/v4'
    )
    expect(OPENAI_COMPATIBLE_PROVIDER_PRESETS.zhipu.models).toContain('glm-4.5')
  })

  it('根据旧账号的 Base URL 回填供应商', () => {
    expect(resolveOpenAICompatibleProvider(undefined, 'https://api.deepseek.com/v1')).toBe('deepseek')
    expect(resolveOpenAICompatibleProvider(undefined, 'https://open.bigmodel.cn/api/paas/v4')).toBe('zhipu')
    expect(resolveOpenAICompatibleProvider(undefined, 'https://relay.example.com/v1')).toBe('custom')
  })

  it('优先使用已经保存的供应商元数据', () => {
    expect(resolveOpenAICompatibleProvider('deepseek', 'https://relay.example.com/v1')).toBe('deepseek')
  })
})
