<template>
  <BaseDialog :show="show" title="调整分组顺序" :close-on-escape="!saving" :show-close-button="!saving" @close="close">
    <div v-if="loading" class="py-8 text-center text-sm text-gray-500">正在加载分组...</div>
    <div v-else-if="loadError" role="alert" class="text-sm text-red-600">{{ loadError }}</div>
    <div v-else-if="items.length === 0" class="py-8 text-center text-sm text-gray-500">暂无分组</div>
    <VueDraggable v-else v-model="items" :animation="150" handle=".monitor-order-handle" :disabled="saving" class="divide-y divide-gray-200 dark:divide-dark-600">
      <div v-for="(item, index) in items" :key="item.id" class="flex min-w-0 items-center gap-2 py-3" :data-monitor-id="item.id">
        <button type="button" class="monitor-order-handle order-icon cursor-grab" :disabled="saving" :title="`拖动 ${item.group_name || item.name}`" :aria-label="`拖动 ${item.group_name || item.name}`">
          <Icon name="arrowsUpDown" size="sm" />
        </button>
        <span class="w-6 shrink-0 text-center text-xs tabular-nums text-gray-500">{{ index + 1 }}</span>
        <span class="min-w-0 flex-1 break-words text-sm text-gray-900 dark:text-gray-100">{{ item.group_name || item.name }}</span>
        <span v-if="!item.enabled" class="shrink-0 text-xs text-gray-500">已停用</span>
        <button type="button" class="order-icon" :disabled="saving || index === 0" :title="`上移 ${item.group_name || item.name}`" :aria-label="`上移 ${item.group_name || item.name}`" @click="move(index, -1)">
          <Icon name="arrowUp" size="sm" />
        </button>
        <button type="button" class="order-icon" :disabled="saving || index === items.length - 1" :title="`下移 ${item.group_name || item.name}`" :aria-label="`下移 ${item.group_name || item.name}`" @click="move(index, 1)">
          <Icon name="arrowDown" size="sm" />
        </button>
      </div>
    </VueDraggable>
    <template #footer>
      <button type="button" class="btn btn-secondary" :disabled="saving" @click="close">取消</button>
      <button type="button" class="btn btn-primary" :disabled="loading || saving || !!loadError || !changed" @click="save">{{ saving ? '保存中...' : '保存' }}</button>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import { VueDraggable } from 'vue-draggable-plus'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Icon from '@/components/icons/Icon.vue'
import { getOrder, updateOrder, type MonitorOrderItem } from '@/api/admin/channelMonitor'
import { useAppStore } from '@/stores/app'
import { extractApiErrorMessage } from '@/utils/apiError'

const props = defineProps<{ show: boolean }>()
const emit = defineEmits<{ (event: 'close'): void; (event: 'saved'): void }>()
const appStore = useAppStore()
const items = ref<MonitorOrderItem[]>([])
const expectedIds = ref<number[]>([])
const loading = ref(false)
const saving = ref(false)
const loadError = ref('')
let controller: AbortController | null = null
const changed = computed(() => items.value.some((item, index) => item.id !== expectedIds.value[index]))

watch(() => props.show, async show => {
  controller?.abort()
  if (!show) return
  const request = new AbortController()
  controller = request
  loading.value = true
  loadError.value = ''
  items.value = []
  expectedIds.value = []
  try {
    const result = await getOrder({ signal: request.signal })
    if (request.signal.aborted) return
    items.value = result
    expectedIds.value = result.map(item => item.id)
  } catch (error) {
    if (!request.signal.aborted) loadError.value = extractApiErrorMessage(error, '加载分组顺序失败')
  } finally {
    if (controller === request) loading.value = false
  }
}, { immediate: true })

function move(index: number, direction: number) {
  const next = index + direction
  if (saving.value || next < 0 || next >= items.value.length) return
  const [item] = items.value.splice(index, 1)
  items.value.splice(next, 0, item)
}

function close() {
  if (!saving.value) emit('close')
}

async function save() {
  if (saving.value || loading.value || loadError.value || !changed.value) return
  saving.value = true
  try {
    await updateOrder(items.value.map(item => item.id), expectedIds.value)
    appStore.showSuccess('分组顺序已保存')
    emit('saved')
    emit('close')
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, '保存分组顺序失败'))
  } finally {
    saving.value = false
  }
}

onBeforeUnmount(() => controller?.abort())
</script>

<style scoped>
.order-icon { display: inline-flex; flex: 0 0 32px; width: 32px; height: 32px; align-items: center; justify-content: center; border-radius: 6px; color: #64748b; }
.order-icon:hover:not(:disabled) { background: #e5e7eb; color: #111827; }
.order-icon:disabled { opacity: .3; cursor: not-allowed; }
</style>
