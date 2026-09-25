import { ref, onBeforeUnmount, type Ref } from 'vue'

export interface UseAutoRefreshOptions {
  storageKey: string
  intervals?: readonly number[]
  defaultInterval?: number
  onRefresh: () => Promise<void> | void
  /** Skip tick when this returns true (e.g. modal open, document hidden). */
  shouldPause?: () => boolean
}

export function useAutoRefresh(options: UseAutoRefreshOptions) {
  const {
    storageKey,
    intervals = [5, 10, 15, 30] as const,
    defaultInterval,
    onRefresh,
    shouldPause,
  } = options

  const enabled = ref(false)
  const intervalSeconds = ref(defaultInterval ?? intervals[intervals.length - 1])
  const countdown = ref(0)
  const fetching = ref(false)

  let timerId: number | undefined
  let nextRefreshAt: number | null = null

  function syncCountdown() {
    countdown.value = nextRefreshAt === null
      ? 0
      : Math.max(0, Math.ceil((nextRefreshAt - Date.now()) / 1000))
  }

  function loadFromStorage() {
    try {
      const saved = localStorage.getItem(storageKey)
      if (!saved) return
      const parsed = JSON.parse(saved) as {
        enabled?: boolean
        interval_seconds?: number
        next_refresh_at?: number
      }
      enabled.value = parsed.enabled === true
      const iv = Number(parsed.interval_seconds)
      if (intervals.includes(iv as any)) intervalSeconds.value = iv
      if (Number.isFinite(parsed.next_refresh_at)) nextRefreshAt = parsed.next_refresh_at as number
      syncCountdown()
    } catch { /* ignore */ }
  }

  function saveToStorage() {
    try {
      localStorage.setItem(storageKey, JSON.stringify({
        enabled: enabled.value,
        interval_seconds: intervalSeconds.value,
        next_refresh_at: nextRefreshAt,
      }))
    } catch { /* ignore */ }
  }

  async function tick() {
    if (!enabled.value) return
    if (shouldPause?.()) return
    if (fetching.value) return

    syncCountdown()
    if (countdown.value <= 0) {
      fetching.value = true
      try { await onRefresh() } finally {
        fetching.value = false
        resetCountdown()
      }
      return
    }
  }

  function start() {
    if (timerId !== undefined) return
    timerId = setInterval(tick, 1000) as unknown as number
  }

  function stop() {
    if (timerId !== undefined) {
      clearInterval(timerId)
      timerId = undefined
    }
  }

  function setEnabled(value: boolean) {
    enabled.value = value
    if (value) {
      if (nextRefreshAt === null) resetCountdown()
      else syncCountdown()
      start()
    } else {
      stop()
      nextRefreshAt = null
      countdown.value = 0
    }
    saveToStorage()
  }

  function setInterval_(seconds: number) {
    intervalSeconds.value = seconds
    if (enabled.value) resetCountdown()
    saveToStorage()
  }

  function resetCountdown() {
    if (!enabled.value) {
      nextRefreshAt = null
      countdown.value = 0
      saveToStorage()
      return
    }
    nextRefreshAt = Date.now() + intervalSeconds.value * 1000
    syncCountdown()
    saveToStorage()
  }

  loadFromStorage()

  onBeforeUnmount(stop)

  return {
    enabled: enabled as Ref<boolean>,
    intervalSeconds: intervalSeconds as Ref<number>,
    countdown: countdown as Ref<number>,
    fetching: fetching as Ref<boolean>,
    intervals,
    setEnabled,
    setInterval: setInterval_,
    resetCountdown,
    start,
    stop,
  }
}
