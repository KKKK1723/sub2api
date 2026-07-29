<template>
  <div ref="pickerRef" class="relative">
    <button
      type="button"
      class="landing-control"
      :title="t('home.appearance.title')"
      :aria-label="t('home.appearance.title')"
      aria-haspopup="menu"
      :aria-expanded="isOpen"
      @click="isOpen = !isOpen"
    >
      <Palette :size="18" :stroke-width="1.8" aria-hidden="true" />
      <span
        class="h-2.5 w-2.5 rounded-full border border-black/10"
        :style="{ backgroundColor: activeOption.color }"
        aria-hidden="true"
      ></span>
    </button>

    <transition name="landing-menu">
      <div
        v-if="isOpen"
        class="absolute right-0 z-50 mt-2 w-52 overflow-hidden rounded-lg border border-[#dfe3e8] bg-white p-2 shadow-[0_16px_40px_rgba(15,23,42,0.12)]"
        role="menu"
      >
        <p class="px-2 pb-2 pt-1 text-xs font-semibold text-[#667085]">
          {{ t('home.appearance.title') }}
        </p>
        <button
          v-for="option in options"
          :key="option.value"
          type="button"
          class="flex min-h-10 w-full items-center gap-3 rounded-md px-2 text-left text-sm font-medium text-[#344054] transition-colors hover:bg-[#f3f5f7]"
          :class="{ 'bg-[#f3f5f7] text-[#101828]': option.value === modelValue }"
          role="menuitemradio"
          :aria-checked="option.value === modelValue"
          @click="selectOption(option.value)"
        >
          <span
            class="h-5 w-5 rounded-md border border-black/10"
            :style="{ backgroundColor: option.color }"
            aria-hidden="true"
          ></span>
          <span>{{ option.label }}</span>
          <Check
            v-if="option.value === modelValue"
            class="ml-auto text-[#101828]"
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
import { Check, Palette } from 'lucide-vue-next'

export type LandingPalette = 'cobalt' | 'graphite' | 'vermilion'

const props = defineProps<{
  modelValue: LandingPalette
}>()

const emit = defineEmits<{
  'update:modelValue': [value: LandingPalette]
}>()

const { t } = useI18n()
const isOpen = ref(false)
const pickerRef = ref<HTMLElement | null>(null)

const options = computed(() => [
  { value: 'cobalt' as const, label: t('home.appearance.cobalt'), color: '#2454d6' },
  { value: 'graphite' as const, label: t('home.appearance.graphite'), color: '#24262b' },
  { value: 'vermilion' as const, label: t('home.appearance.vermilion'), color: '#d9472b' }
])

const activeOption = computed(
  () => options.value.find((option) => option.value === props.modelValue) ?? options.value[0]
)

function selectOption(value: LandingPalette) {
  emit('update:modelValue', value)
  isOpen.value = false
}

function handleClickOutside(event: MouseEvent) {
  if (pickerRef.value && !pickerRef.value.contains(event.target as Node)) {
    isOpen.value = false
  }
}

onMounted(() => document.addEventListener('click', handleClickOutside))
onBeforeUnmount(() => document.removeEventListener('click', handleClickOutside))
</script>

<style scoped>
.landing-control {
  display: inline-flex;
  width: 40px;
  height: 40px;
  align-items: center;
  justify-content: center;
  gap: 5px;
  border: 1px solid #d8dadd;
  border-radius: 6px;
  background: #fff;
  color: #344054;
  transition: border-color 160ms ease, background-color 160ms ease, color 160ms ease;
}

.landing-control:hover {
  border-color: #a8abb0;
  background: #f5f5f4;
  color: #15171a;
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
