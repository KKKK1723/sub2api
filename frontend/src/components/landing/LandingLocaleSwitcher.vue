<template>
  <div ref="switcherRef" class="relative">
    <button
      type="button"
      class="landing-locale-control"
      :disabled="switching"
      :title="currentLocale?.name"
      :aria-label="currentLocale?.name"
      aria-haspopup="menu"
      :aria-expanded="isOpen"
      @click="isOpen = !isOpen"
    >
      <Languages :size="18" :stroke-width="1.8" aria-hidden="true" />
      <span class="hidden text-xs font-semibold sm:inline">{{ currentLocaleCode.toUpperCase() }}</span>
      <ChevronDown
        class="hidden transition-transform sm:block"
        :class="{ 'rotate-180': isOpen }"
        :size="14"
        :stroke-width="2"
        aria-hidden="true"
      />
    </button>

    <transition name="landing-menu">
      <div
        v-if="isOpen"
        class="absolute right-0 z-50 mt-2 w-40 overflow-hidden rounded-lg border border-[#dfe3e8] bg-white p-2 shadow-[0_16px_40px_rgba(15,23,42,0.12)]"
        role="menu"
      >
        <button
          v-for="option in availableLocales"
          :key="option.code"
          type="button"
          class="flex min-h-10 w-full items-center gap-2 rounded-md px-2 text-left text-sm font-medium text-[#344054] transition-colors hover:bg-[#f3f5f7]"
          :class="{ 'bg-[#f3f5f7] text-[#101828]': option.code === currentLocaleCode }"
          :disabled="switching"
          role="menuitemradio"
          :aria-checked="option.code === currentLocaleCode"
          @click="selectLocale(option.code)"
        >
          <span>{{ option.name }}</span>
          <Check
            v-if="option.code === currentLocaleCode"
            class="ml-auto"
            :size="16"
            :stroke-width="2"
            aria-hidden="true"
          />
        </button>
      </div>
    </transition>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { Check, ChevronDown, Languages } from 'lucide-vue-next'
import { availableLocales, setLocale } from '@/i18n'

const { locale } = useI18n()
const isOpen = ref(false)
const switching = ref(false)
const switcherRef = ref<HTMLElement | null>(null)

const currentLocaleCode = computed(() => locale.value)
const currentLocale = computed(() =>
  availableLocales.find((option) => option.code === currentLocaleCode.value)
)

async function selectLocale(code: string) {
  if (switching.value || code === currentLocaleCode.value) {
    isOpen.value = false
    return
  }

  switching.value = true
  try {
    await setLocale(code)
    isOpen.value = false
  } finally {
    switching.value = false
  }
}

function handleClickOutside(event: MouseEvent) {
  if (switcherRef.value && !switcherRef.value.contains(event.target as Node)) {
    isOpen.value = false
  }
}

onMounted(() => document.addEventListener('click', handleClickOutside))
onBeforeUnmount(() => document.removeEventListener('click', handleClickOutside))
</script>

<style scoped>
.landing-locale-control {
  display: inline-flex;
  height: 40px;
  min-width: 40px;
  align-items: center;
  justify-content: center;
  gap: 6px;
  border: 1px solid #dfe3e8;
  border-radius: 8px;
  background: #fff;
  padding: 0 10px;
  color: #344054;
  transition: border-color 160ms ease, background-color 160ms ease, color 160ms ease;
}

.landing-locale-control:hover {
  border-color: #b9c0ca;
  background: #f8fafb;
  color: #101828;
}

.landing-menu-enter-active,
.landing-menu-leave-active {
  transition: opacity 140ms ease, transform 140ms ease;
}

.landing-menu-enter-from,
.landing-menu-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}
</style>
