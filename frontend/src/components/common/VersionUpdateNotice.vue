<template>
  <Teleport to="body">
    <Transition
      enter-active-class="transition ease-out duration-300"
      enter-from-class="opacity-0 translate-y-4"
      enter-to-class="opacity-100 translate-y-0"
      leave-active-class="transition ease-in duration-200"
      leave-from-class="opacity-100 translate-y-0"
      leave-to-class="opacity-0 translate-y-4"
    >
      <div
        v-if="visible"
        class="fixed bottom-4 right-4 z-[9998] flex w-[340px] max-w-[calc(100vw-2rem)] items-start gap-3 rounded-[10px] border border-l-4 border-gray-200 border-l-primary-600 bg-white py-3.5 pl-4 pr-3.5 shadow-xl dark:border-dark-700 dark:border-l-primary-500 dark:bg-dark-800"
        role="status"
        aria-live="polite"
      >
        <div class="flex h-8 w-8 flex-shrink-0 items-center justify-center rounded-lg bg-primary-50 text-primary-600 dark:bg-primary-500/10 dark:text-primary-400">
          <Icon name="refresh" size="sm" :stroke-width="2" aria-hidden="true" />
        </div>

        <div class="min-w-0 flex-1">
          <p class="mt-px text-sm font-semibold text-gray-900 dark:text-white">
            {{ t('versionUpdate.title') }}
          </p>
          <p class="mb-2.5 mt-0.5 text-[13px] text-gray-500 dark:text-gray-400">
            {{ t('versionUpdate.description') }}
          </p>
          <div class="flex gap-2">
            <button
              type="button"
              class="h-[30px] rounded-[7px] bg-primary-600 px-3.5 text-[13px] font-semibold text-white transition-colors hover:bg-primary-700 dark:bg-primary-500 dark:hover:bg-primary-400"
              @click="reloadNow"
            >
              {{ t('versionUpdate.refresh') }}
            </button>
            <button
              type="button"
              class="h-[30px] rounded-[7px] px-2.5 text-[13px] text-gray-500 transition-colors hover:bg-gray-100 hover:text-gray-900 dark:text-gray-400 dark:hover:bg-dark-700 dark:hover:text-white"
              @click="dismissNotice"
            >
              {{ t('versionUpdate.later') }}
            </button>
          </div>
        </div>

        <button
          type="button"
          class="-mr-1 -mt-1 flex h-[26px] w-[26px] flex-shrink-0 items-center justify-center rounded-md text-gray-400 transition-colors hover:bg-gray-100 hover:text-gray-600 dark:text-gray-500 dark:hover:bg-dark-700 dark:hover:text-gray-300"
          :aria-label="t('versionUpdate.close')"
          @click="dismissNotice"
        >
          <Icon name="x" size="sm" />
        </button>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import { useVersionUpdate } from '@/utils/versionUpdate'

const { t } = useI18n()
const { updateAvailable, noticeDismissed, dismissNotice, reloadNow } = useVersionUpdate()

// 检测到新版本且用户未关闭时展示；切换页面会直接整页加载新版本，无需等待用户点击
const visible = computed(() => updateAvailable.value && !noticeDismissed.value)
</script>
