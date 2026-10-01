import { describe, expect, it } from 'vitest'
import { formatScaled } from '../pricing'

describe('价格金额格式', () => {
  it.each([
    [null, 1, '-'],
    [0, 1, '$0'],
    [3e-6, 1_000_000, '$3'],
    [0.1 + 0.2, 1, '$0.3'],
    [1e-10, 1, '$1e-10'],
    [1e-20, 1, '$1e-20'],
    [1e10, 1, '$10000000000'],
  ])('正确展示 %s × %s，保留科学计数法的指数', (price, scale, expected) => {
    expect(formatScaled(price as number | null, scale as number)).toBe(expected)
  })
})
