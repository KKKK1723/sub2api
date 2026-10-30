/**
 * Channel Status Store
 * 缓存渠道状态列表：再次进入状态页时先展示上一次结果，再在后台静默刷新，
 * 避免每次进入都要等待接口返回才出内容。
 */

import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { UserMonitorView } from '@/api/channelMonitor'

export const useChannelStatusStore = defineStore('channelStatus', () => {
  const items = ref<UserMonitorView[]>([])
  const loaded = ref(false)

  function setItems(next: UserMonitorView[]) {
    items.value = next
    loaded.value = true
  }

  function clear() {
    items.value = []
    loaded.value = false
  }

  return {
    items,
    loaded,
    setItems,
    clear
  }
})
