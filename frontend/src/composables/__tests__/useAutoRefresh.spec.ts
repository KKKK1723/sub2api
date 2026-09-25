import { defineComponent } from 'vue'
import { mount } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { useAutoRefresh } from '../useAutoRefresh'

describe('useAutoRefresh', () => {
  const storageKey = 'test-auto-refresh'

  beforeEach(() => {
    vi.useFakeTimers()
    vi.setSystemTime(new Date('2026-01-01T00:00:00.000Z'))
    localStorage.clear()
  })

  afterEach(() => {
    vi.useRealTimers()
    localStorage.clear()
  })

  function mountRefresh() {
    let state: ReturnType<typeof useAutoRefresh> | undefined
    const wrapper = mount(defineComponent({
      setup() {
        state = useAutoRefresh({
          storageKey,
          intervals: [180],
          defaultInterval: 180,
          onRefresh: vi.fn(),
        })
        return () => null
      },
    }))
    return { wrapper, state: state as ReturnType<typeof useAutoRefresh> }
  }

  it('重新进入页面时按已保存的刷新时间恢复倒计时', () => {
    const first = mountRefresh()
    first.state.setEnabled(true)
    vi.advanceTimersByTime(45_000)
    expect(first.state.countdown.value).toBe(135)
    first.wrapper.unmount()

    vi.advanceTimersByTime(30_000)
    const second = mountRefresh()
    second.state.setEnabled(true)

    expect(second.state.countdown.value).toBe(105)
    second.wrapper.unmount()
  })

  it('旧的本地存储没有时间点时会初始化新的刷新周期', () => {
    localStorage.setItem(storageKey, JSON.stringify({ enabled: true, interval_seconds: 180 }))

    const { wrapper, state } = mountRefresh()
    state.setEnabled(true)

    expect(state.countdown.value).toBe(180)
    wrapper.unmount()
  })
})
