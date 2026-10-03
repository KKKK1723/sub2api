import { describe, expect, it } from 'vitest'

import type { GroupPlatform } from '@/types'
import { getKeyGroupProvider, KEY_GROUP_PROVIDER_ICONS } from '../keyGroupProviders'

describe('API Key 分组平台分类', () => {
  it('按分组平台归类厂商', () => {
    const expected: Record<GroupPlatform, string> = {
      anthropic: 'anthropic',
      openai: 'openai',
      kimi: 'domestic',
      zhipu: 'domestic',
      deepseek: 'domestic',
      minimax: 'domestic',
      gemini: 'other',
      grok: 'other',
      antigravity: 'other',
      composite: 'other',
      opencode_go: 'other'
    }

    for (const [platform, provider] of Object.entries(expected)) {
      expect(getKeyGroupProvider(platform as GroupPlatform)).toBe(provider)
    }
  })

  it('为每个厂商提供代表性平台图标', () => {
    expect(KEY_GROUP_PROVIDER_ICONS).toEqual({
      anthropic: ['anthropic'],
      openai: ['openai'],
      domestic: ['deepseek', 'kimi'],
      other: ['gemini', 'grok']
    })
  })
})
