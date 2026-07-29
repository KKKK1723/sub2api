<template>
  <div ref="pickerRef" class="appearance-picker">
    <button
      type="button"
      class="appearance-trigger"
      :title="t('home.appearance.title')"
      :aria-label="t('home.appearance.title')"
      aria-haspopup="menu"
      :aria-expanded="isOpen"
      @click="isOpen = !isOpen"
    >
      <Palette :size="17" :stroke-width="1.8" aria-hidden="true" />
      <span class="appearance-swatch" :style="{ backgroundColor: activeOption.color }" aria-hidden="true"></span>
    </button>

    <transition name="appearance-menu">
      <div v-if="isOpen" class="appearance-menu" role="menu">
        <p>{{ t('home.appearance.title') }}</p>
        <button
          v-for="option in options"
          :key="option.value"
          type="button"
          class="appearance-option"
          :class="{ 'is-active': option.value === modelValue }"
          role="menuitemradio"
          :aria-checked="option.value === modelValue"
          @click="selectOption(option.value)"
        >
          <span class="option-swatch" :style="{ backgroundColor: option.color }" aria-hidden="true"></span>
          <span>{{ option.label }}</span>
          <Check v-if="option.value === modelValue" class="option-check" :size="16" :stroke-width="2" aria-hidden="true" />
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
  { value: 'cobalt' as const, label: t('home.appearance.cobalt'), color: '#2856d8' },
  { value: 'graphite' as const, label: t('home.appearance.graphite'), color: '#252932' },
  { value: 'vermilion' as const, label: t('home.appearance.vermilion'), color: '#c9422e' }
])

const activeOption = computed(
  () => options.value.find((option) => option.value === props.modelValue) ?? options.value[0]
)

function selectOption(value: LandingPalette) {
  emit('update:modelValue', value)
  isOpen.value = false
}

function handleClickOutside(event: MouseEvent) {
  if (pickerRef.value && !pickerRef.value.contains(event.target as Node)) isOpen.value = false
}

onMounted(() => document.addEventListener('click', handleClickOutside))
onBeforeUnmount(() => document.removeEventListener('click', handleClickOutside))
</script>

<style scoped>
.appearance-picker { position: relative; }

.appearance-trigger {
  display: inline-flex;
  width: 38px;
  height: 38px;
  align-items: center;
  justify-content: center;
  gap: 5px;
  border: 1px solid #e3e7ed;
  border-radius: 6px;
  background: #fff;
  color: #596174;
  transition: border-color 160ms ease, background-color 160ms ease, color 160ms ease;
}

.appearance-trigger:hover {
  border-color: #b8c0cd;
  background: #f8f9fb;
  color: #1d2532;
}

.appearance-swatch {
  width: 5px;
  height: 16px;
  border-radius: 999px;
}

.appearance-menu {
  position: absolute;
  right: 0;
  z-index: 50;
  width: 188px;
  margin-top: 8px;
  padding: 7px;
  border: 1px solid #e1e6ed;
  border-radius: 8px;
  background: #fff;
  box-shadow: 0 18px 44px rgba(25, 34, 52, 0.13);
}

.appearance-menu > p {
  padding: 6px 8px 8px;
  color: #7a8493;
  font-size: 11px;
  font-weight: 750;
}

.appearance-option {
  display: flex;
  width: 100%;
  min-height: 38px;
  align-items: center;
  gap: 10px;
  padding: 0 8px;
  border-radius: 5px;
  color: #445064;
  font-size: 13px;
  font-weight: 700;
  text-align: left;
  transition: background-color 140ms ease, color 140ms ease;
}

.appearance-option:hover,
.appearance-option.is-active {
  background: #f4f6f9;
  color: #1d2532;
}

.option-swatch {
  width: 18px;
  height: 18px;
  flex: none;
  border: 1px solid rgba(17, 24, 39, 0.12);
  border-radius: 4px;
}

.option-check { margin-left: auto; }

.appearance-menu-enter-active,
.appearance-menu-leave-active { transition: opacity 140ms ease, transform 140ms ease; }
.appearance-menu-enter-from,
.appearance-menu-leave-to { opacity: 0; transform: translateY(-4px); }
</style>