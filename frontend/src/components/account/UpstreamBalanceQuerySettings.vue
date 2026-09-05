<template>
  <div class="border-t border-gray-200 pt-4 dark:border-dark-600">
    <div class="flex items-center justify-between gap-4">
      <div>
        <label class="input-label mb-0">{{ t('admin.accounts.upstreamBalance.enabled') }}</label>
        <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
          {{ t('admin.accounts.upstreamBalance.enabledHint') }}
        </p>
      </div>
      <Toggle
        :model-value="enabled"
        data-testid="upstream-balance-enabled"
        :aria-label="t('admin.accounts.upstreamBalance.enabled')"
        @update:model-value="emit('update:enabled', $event)"
      />
    </div>

    <div v-if="enabled" class="mt-4 space-y-4">
      <div>
        <label class="input-label">{{ t('admin.accounts.upstreamBalance.preset') }}</label>
        <select
          :value="query.preset === 'custom' ? 'custom' : 'sub2api'"
          class="input"
          data-testid="upstream-balance-preset"
          @change="handlePresetChange"
        >
          <option value="sub2api">{{ t('admin.accounts.upstreamBalance.presets.sub2api') }}</option>
          <option value="custom">{{ t('admin.accounts.upstreamBalance.presets.custom') }}</option>
        </select>
      </div>

      <template v-if="query.preset === 'custom'">
        <div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
          <div>
            <label class="input-label">{{ t('admin.accounts.upstreamBalance.method') }}</label>
            <select
              :value="query.method || 'GET'"
              class="input"
              @change="updateField('method', ($event.target as HTMLSelectElement).value as 'GET' | 'POST')"
            >
              <option value="GET">GET</option>
              <option value="POST">POST</option>
            </select>
          </div>
          <div>
            <label class="input-label">{{ t('admin.accounts.upstreamBalance.authType') }}</label>
            <select
              :value="query.auth_type || 'bearer'"
              class="input"
              @change="updateField('auth_type', ($event.target as HTMLSelectElement).value as 'bearer' | 'x-api-key')"
            >
              <option value="bearer">Bearer</option>
              <option value="x-api-key">X-API-Key</option>
            </select>
          </div>
        </div>

        <div>
          <label class="input-label">{{ t('admin.accounts.upstreamBalance.url') }}</label>
          <input
            :value="query.url || ''"
            class="input font-mono text-sm"
            type="text"
            placeholder="{{baseUrl}}/v1/usage"
            @input="updateField('url', ($event.target as HTMLInputElement).value)"
          />
          <p class="input-hint">{{ t('admin.accounts.upstreamBalance.urlHint') }}</p>
        </div>

        <div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
          <div>
            <label class="input-label">{{ t('admin.accounts.upstreamBalance.balancePath') }}</label>
            <input
              :value="query.balance_path || ''"
              class="input font-mono text-sm"
              type="text"
              placeholder="remaining"
              @input="updateField('balance_path', ($event.target as HTMLInputElement).value)"
            />
          </div>
          <div>
            <label class="input-label">{{ t('admin.accounts.upstreamBalance.unitPath') }}</label>
            <input
              :value="query.unit_path || ''"
              class="input font-mono text-sm"
              type="text"
              placeholder="unit"
              @input="updateField('unit_path', ($event.target as HTMLInputElement).value)"
            />
          </div>
          <div>
            <label class="input-label">{{ t('admin.accounts.upstreamBalance.planPath') }}</label>
            <input
              :value="query.plan_path || ''"
              class="input font-mono text-sm"
              type="text"
              placeholder="planName"
              @input="updateField('plan_path', ($event.target as HTMLInputElement).value)"
            />
          </div>
          <div>
            <label class="input-label">{{ t('admin.accounts.upstreamBalance.errorPath') }}</label>
            <input
              :value="query.error_path || ''"
              class="input font-mono text-sm"
              type="text"
              placeholder="error"
              @input="updateField('error_path', ($event.target as HTMLInputElement).value)"
            />
          </div>
          <div>
            <label class="input-label">{{ t('admin.accounts.upstreamBalance.totalPath') }}</label>
            <input
              :value="query.total_path || ''"
              class="input font-mono text-sm"
              type="text"
              placeholder="quota.limit"
              @input="updateField('total_path', ($event.target as HTMLInputElement).value)"
            />
          </div>
          <div>
            <label class="input-label">{{ t('admin.accounts.upstreamBalance.usedPath') }}</label>
            <input
              :value="query.used_path || ''"
              class="input font-mono text-sm"
              type="text"
              placeholder="quota.used"
              @input="updateField('used_path', ($event.target as HTMLInputElement).value)"
            />
          </div>
          <div>
            <label class="input-label">{{ t('admin.accounts.upstreamBalance.validPath') }}</label>
            <input
              :value="query.valid_path || ''"
              class="input font-mono text-sm"
              type="text"
              placeholder="is_active"
              @input="updateField('valid_path', ($event.target as HTMLInputElement).value)"
            />
          </div>
          <div>
            <label class="input-label">{{ t('admin.accounts.upstreamBalance.defaultUnit') }}</label>
            <input
              :value="query.unit || ''"
              class="input font-mono text-sm"
              type="text"
              placeholder="USD"
              maxlength="16"
              @input="updateField('unit', ($event.target as HTMLInputElement).value)"
            />
          </div>
        </div>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Toggle from '@/components/common/Toggle.vue'
import type { UpstreamBalancePreset, UpstreamBalanceQuery } from '@/types'

const props = defineProps<{
  enabled: boolean
  query: UpstreamBalanceQuery
}>()

const emit = defineEmits<{
  'update:enabled': [value: boolean]
  'update:query': [value: UpstreamBalanceQuery]
}>()

const { t } = useI18n()

const customDefaults = (): UpstreamBalanceQuery => ({
  preset: 'custom',
  method: 'GET',
  url: '{{baseUrl}}/v1/usage',
  auth_type: 'bearer',
  balance_path: 'remaining',
  unit_path: 'unit',
  plan_path: 'planName',
  error_path: 'error',
  unit: 'USD'
})

const handlePresetChange = (event: Event) => {
  const preset = (event.target as HTMLSelectElement).value as UpstreamBalancePreset
  emit('update:query', preset === 'custom' ? customDefaults() : { preset })
}

const updateField = <K extends keyof UpstreamBalanceQuery>(key: K, value: UpstreamBalanceQuery[K]) => {
  emit('update:query', { ...props.query, [key]: value })
}
</script>
