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
import Icon from '../../src/components/icons/Icon.vue'

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

<style>
.tools-band {
  border-bottom: 1px solid var(--line);
  background: var(--surface);
}
.tools-inner {
  display: flex;
  align-items: center;
  gap: 23px;
  min-height: 108px;
}
.tools-label {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-shrink: 0;
  font-size: 11px;
  font-weight: 500;
  color: var(--muted);
}
.tools-label .eyebrow-mark {
  width: 4px;
  height: 4px;
}
.tools-viewport {
  min-width: 0;
  flex: 1;
  overflow: hidden;
  mask-image: linear-gradient(
    90deg,
    transparent,
    #000 5%,
    #000 95%,
    transparent
  );
}
.tools-track {
  display: flex;
  width: max-content;
  animation: tools-stream 32s linear infinite;
}
.tools-group {
  display: flex;
  gap: 14px;
  padding-right: 14px;
  flex-shrink: 0;
}
.tool-item {
  width: 190px;
  min-height: 47px;
  padding: 0 15px;
  display: inline-flex;
  align-items: center;
  gap: 10px;
  background: var(--page);
  border: 1px solid var(--line);
  border-radius: 6px;
  white-space: nowrap;
  font-size: 12px;
  font-weight: 550;
  color: var(--ink);
}
.tool-item > img {
  width: 20px;
  height: 20px;
  object-fit: contain;
}
.tool-item > svg:first-child {
  color: var(--accent);
}
.tool-item.apricot > svg:first-child {
  color: var(--amber);
}
.tool-item.blue > svg:first-child {
  color: var(--blue);
}
.tool-item > svg:last-child {
  margin-left: auto;
  color: var(--subtle);
  width: 11px;
  height: 11px;
}
.tools-track.is-paused,
.tools-viewport:hover .tools-track {
  animation-play-state: paused;
}
.tools-pause {
  flex-shrink: 0;
  width: 30px;
  height: 30px;
}
.pause-symbol {
  width: 10px;
  height: 12px;
  border-inline: 3px solid currentColor;
}
@keyframes tools-stream {
  to {
    transform: translateX(-50%);
  }
}
@media (max-width: 600px) {
  .tools-inner {
    display: grid;
    grid-template-columns: minmax(0, 1fr) 28px;
    gap: 13px 8px;
    padding-block: 18px 20px;
  }
  .tools-label {
    grid-column: 1;
    grid-row: 1;
    font-size: 10px;
  }
  .tools-pause {
    grid-column: 2;
    grid-row: 1;
    width: 28px;
    height: 24px;
  }
  .tools-viewport {
    grid-column: 1 / -1;
    grid-row: 2;
  }
  .tool-item {
    width: 166px;
    min-height: 43px;
    font-size: 11px;
  }
  .tool-item > img {
    width: 18px;
    height: 18px;
  }
}
@media (prefers-reduced-motion: reduce) {
  .tools-track {
    animation: none;
  }
  .tools-group:nth-child(2),
  .tools-pause {
    display: none;
  }
  .tools-viewport {
    overflow-x: auto;
    mask-image: none;
  }
}
</style>
