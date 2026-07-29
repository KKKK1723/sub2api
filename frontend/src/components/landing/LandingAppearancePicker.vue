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
      <Palette :size="19" :stroke-width="1.8" aria-hidden="true" />
      <span class="appearance-swatch" :style="{ backgroundColor: activeOption.color }" aria-hidden="true"></span>
      <ChevronDown class="appearance-chevron" :size="14" :stroke-width="2" aria-hidden="true" />
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
          <Check v-if="option.value === modelValue" class="option-check" :size="17" :stroke-width="2.2" aria-hidden="true" />
        </button>
      </div>
    </transition>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { Check, ChevronDown, Palette } from 'lucide-vue-next'

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
  width: 44px;
  height: 42px;
  align-items: center;
  justify-content: center;
  gap: 6px;
  border: 1px solid #dbe1e9;
  border-radius: 7px;
  background: #fff;
  color: #445065;
  box-shadow: 0 1px 1px rgba(17, 24, 39, 0.02);
  transition: border-color 160ms ease, background-color 160ms ease, color 160ms ease, box-shadow 160ms ease;
}

.appearance-trigger:hover {
  border-color: #aeb9c9;
  background: #f8f9fb;
  color: #1b2432;
  box-shadow: 0 4px 12px rgba(28, 39, 59, 0.08);
}

.appearance-swatch {
  width: 15px;
  height: 15px;
  flex: none;
  border: 1px solid rgba(17, 24, 39, 0.14);
  border-radius: 4px;
}

.appearance-chevron { margin-left: -2px; color: #7b8798; }

.appearance-menu {
  position: absolute;
  right: 0;
  z-index: 50;
  width: 220px;
  margin-top: 9px;
  padding: 9px;
  border: 1px solid #dce2eb;
  border-radius: 10px;
  background: #fff;
  box-shadow: 0 18px 48px rgba(25, 34, 52, 0.15);
}

.appearance-menu > p {
  padding: 6px 9px 9px;
  color: #687589;
  font-family: "PingFang SC", "Microsoft YaHei UI", "Microsoft YaHei", system-ui, sans-serif;
  font-size: 12px;
  font-weight: 800;
}

.appearance-option {
  display: flex;
  width: 100%;
  min-height: 46px;
  align-items: center;
  gap: 12px;
  padding: 0 9px;
  border-radius: 6px;
  color: #3d4a5e;
  font-family: "PingFang SC", "Microsoft YaHei UI", "Microsoft YaHei", system-ui, sans-serif;
  font-size: 14px;
  font-weight: 750;
  text-align: left;
  transition: background-color 140ms ease, color 140ms ease;
}

.appearance-option:hover,
.appearance-option.is-active {
  background: #f3f6fa;
  color: #1c2738;
}

.option-swatch {
  width: 22px;
  height: 22px;
  flex: none;
  border: 1px solid rgba(17, 24, 39, 0.13);
  border-radius: 5px;
}

.option-check { margin-left: auto; }

.appearance-menu-enter-active,
.appearance-menu-leave-active { transition: opacity 140ms ease, transform 140ms ease; }
.appearance-menu-enter-from,
.appearance-menu-leave-to { opacity: 0; transform: translateY(-4px); }
</style>