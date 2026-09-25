export type OpenAICompatibleProvider = 'openai' | 'deepseek' | 'zhipu' | 'custom'

export const OPENAI_COMPATIBLE_PROVIDER_EXTRA_KEY = 'openai_compatible_provider'

export const OPENAI_COMPATIBLE_PROVIDER_PRESETS: Record<
  Exclude<OpenAICompatibleProvider, 'custom'>,
  { baseUrl: string; models: string[] }
> = {
  openai: { baseUrl: 'https://api.openai.com', models: [] },
  deepseek: {
    baseUrl: 'https://api.deepseek.com',
    models: ['deepseek-chat', 'deepseek-reasoner']
  },
  zhipu: {
    baseUrl: 'https://open.bigmodel.cn/api/paas/v4',
    models: ['glm-4.5', 'glm-4.5-air', 'glm-4.5-flash']
  }
}

export function normalizeOpenAICompatibleProvider(value: unknown): OpenAICompatibleProvider {
  return value === 'openai' || value === 'deepseek' || value === 'zhipu' || value === 'custom'
    ? value
    : 'openai'
}

export function resolveOpenAICompatibleProvider(value: unknown, baseUrl: unknown): OpenAICompatibleProvider {
  if (value === 'openai' || value === 'deepseek' || value === 'zhipu' || value === 'custom') {
    return value
  }
  if (typeof baseUrl !== 'string') return 'openai'
  const normalizedUrl = baseUrl.toLowerCase()
  if (normalizedUrl.includes('api.deepseek.com')) return 'deepseek'
  if (normalizedUrl.includes('open.bigmodel.cn')) return 'zhipu'
  return normalizedUrl.includes('api.openai.com') ? 'openai' : 'custom'
}
