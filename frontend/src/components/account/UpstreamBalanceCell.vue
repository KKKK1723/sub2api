<template>
  <div v-if="eligible && configured" class="flex h-6 min-w-[8rem] items-center gap-1">
    <HelpTooltip width-class="w-max max-w-[calc(100vw-2rem)]">
      <template #trigger>
        <span
          class="cursor-help border-b border-dotted border-gray-300 font-mono text-sm font-medium dark:border-dark-600"
          :class="valueClass"
          data-testid="upstream-balance-value"
        >
          {{ primaryValue }}
        </span>
      </template>
      <div class="space-y-1 text-xs">
        <p>{{ t('admin.accounts.upstreamBalance.statusLabel', { value: statusLabel }) }}</p>
        <p v-if="data?.plan_name">{{ t('admin.accounts.upstreamBalance.planLabel', { value: data.plan_name }) }}</p>
        <p v-if="data?.total != null">{{ t('admin.accounts.upstreamBalance.totalLabel', { value: formatAmount(data.total, data.unit) }) }}</p>
        <p v-if="data?.used != null">{{ t('admin.accounts.upstreamBalance.usedLabel', { value: formatAmount(data.used, data.unit) }) }}</p>
        <p v-if="snapshot?.received_at">{{ t('admin.accounts.upstreamBalance.updatedAt', { value: formatDate(snapshot.received_at) }) }}</p>
        <p v-if="snapshot?.last_error">{{ t('admin.accounts.upstreamBalance.errorLabel', { value: snapshot.last_error }) }}</p>
      </div>
    </HelpTooltip>
    <button
      type="button"
      class="inline-flex h-6 w-6 flex-shrink-0 items-center justify-center rounded text-primary-600 transition-colors hover:bg-primary-50 disabled:cursor-not-allowed disabled:opacity-50 dark:text-primary-300 dark:hover:bg-primary-900/30"
      :disabled="probing"
      :aria-label="t('admin.accounts.upstreamBalance.refresh')"
      :title="t('admin.accounts.upstreamBalance.refresh')"
      data-testid="upstream-balance-probe"
      @click="$emit('probe')"
    >
      <Icon name="refresh" size="xs" :class="{ 'animate-spin': probing }" />
    </button>
  </div>
  <span v-else class="text-sm text-gray-400 dark:text-dark-500">-</span>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import HelpTooltip from '@/components/common/HelpTooltip.vue'
import Icon from '@/components/icons/Icon.vue'
import type { Account, UpstreamBalanceProbeSnapshot } from '@/types'

const props = withDefaults(defineProps<{
  account: Account
  now: number
  probing?: boolean
}>(), {
  probing: false
})

defineEmits<{
  probe: []
}>()

const { t, locale } = useI18n()
const eligible = computed(() => props.account.type === 'apikey')
const configured = computed(() => !!props.account.extra?.upstream_balance_query)
const snapshot = computed<UpstreamBalanceProbeSnapshot | undefined>(() => props.account.extra?.upstream_balance_probe)
const data = computed(() => snapshot.value?.data)
const stale = computed(() => {
  if (!snapshot.value || snapshot.value.status !== 'ok') return true
  const freshUntil = snapshot.value.fresh_until
  return !freshUntil || !Number.isFinite(Date.parse(freshUntil)) || props.now > Date.parse(freshUntil)
})

const formatAmount = (value: number, unit?: string) => {
  const formatted = new Intl.NumberFormat(locale.value, { maximumFractionDigits: 4 }).format(value)
  return unit ? `${formatted} ${unit}` : formatted
}

const primaryValue = computed(() => {
  if (data.value) return formatAmount(data.value.balance, data.value.unit)
  if (!snapshot.value) return t('admin.accounts.upstreamBalance.notQueried')
  return statusLabel.value
})

const statusLabel = computed(() => {
  if (data.value?.is_valid === false) return t('admin.accounts.upstreamBalance.invalid')
  if (snapshot.value?.status === 'unsupported') return t('admin.accounts.upstreamBalance.unsupported')
  if (snapshot.value?.status === 'failed') return t('admin.accounts.upstreamBalance.failed')
  if (stale.value && data.value) return t('admin.accounts.upstreamBalance.stale')
  if (snapshot.value?.status === 'ok') return t('admin.accounts.upstreamBalance.ok')
  return t('admin.accounts.upstreamBalance.notQueried')
})

const valueClass = computed(() => {
  if (data.value?.is_valid === false || (snapshot.value?.status === 'failed' && !data.value)) {
    return 'text-red-600 dark:text-red-400'
  }
  if (stale.value) return 'text-amber-600 dark:text-amber-400'
  return 'text-gray-800 dark:text-gray-200'
})

const formatDate = (value?: string) => {
  if (!value) return '-'
  const parsed = new Date(value)
  return Number.isNaN(parsed.getTime()) ? '-' : parsed.toLocaleString(locale.value)
}
</script>
