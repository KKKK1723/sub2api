<template>
  <section ref="root" class="tools-band" :aria-label="label">
    <div class="container tools-inner" data-reveal>
      <div class="tools-label">
        <span class="eyebrow-mark" aria-hidden="true"></span>{{ label }}
      </div>
      <div class="tools-viewport">
        <div
          class="tools-track"
          :class="{ 'is-paused': paused || !inView || pageHidden }"
        >
          <div
            v-for="group in 2"
            :key="group"
            class="tools-group"
            :aria-hidden="group === 2 ? 'true' : undefined"
            :role="group === 1 ? 'list' : undefined"
          >
            <span
              v-for="tool in tools"
              :key="tool.name"
              class="tool-item"
              :class="tool.tone"
              :role="group === 1 ? 'listitem' : undefined"
            >
              <img
                v-if="tool.logo"
                :src="tool.logo"
                alt=""
                :class="
                  tool.name === 'Claude Code' ? 'claude-logo' : 'mono-logo'
                "
              />
              <Icon v-else :name="tool.icon" size="md" />
              <span>{{ tool.name }}</span
              ><Icon name="arrowUp" size="xs" class="diagonal-arrow" />
            </span>
          </div>
        </div>
      </div>
      <button
        class="icon-button tools-pause"
        :title="paused ? resumeLabel : pauseLabel"
        :aria-label="paused ? resumeLabel : pauseLabel"
        :aria-pressed="paused"
        @click="paused = !paused"
      >
        <Icon v-if="paused" name="play" size="sm" /><span
          v-else
          class="pause-symbol"
          aria-hidden="true"
        ></span>
      </button>
    </div>
  </section>
</template>

<script setup>
import { onMounted, onUnmounted, ref } from 'vue'
import Icon from '@/components/icons/Icon.vue'

defineProps({ label: String, pauseLabel: String, resumeLabel: String })
const tools = [
  { name: 'Codex', logo: '/brands/openai.svg', tone: 'mint' },
  { name: 'Claude Code', logo: '/brands/claude.svg', tone: 'apricot' },
  { name: 'Cherry Studio', icon: 'sparkles', tone: 'blue' },
  { name: 'OpenAI SDK', logo: '/brands/openai.svg', tone: 'mint' },
  { name: 'cURL', icon: 'terminal', tone: 'apricot' },
  { name: 'Python', icon: 'terminal', tone: 'blue' },
]
const root = ref(null),
  paused = ref(false),
  inView = ref(false),
  pageHidden = ref(document.hidden)
let observer
const onVisibility = () => {
  pageHidden.value = document.hidden
}
onMounted(() => {
  observer = new IntersectionObserver(([entry]) => {
    inView.value = entry.isIntersecting
  })
  observer.observe(root.value)
  document.addEventListener('visibilitychange', onVisibility)
})
onUnmounted(() => {
  observer?.disconnect()
  document.removeEventListener('visibilitychange', onVisibility)
})
</script>

