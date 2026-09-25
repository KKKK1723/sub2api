<template>
  <canvas ref="canvas" class="hero-flow" aria-hidden="true"></canvas>
</template>

<script setup>
import { onMounted, onUnmounted, ref, watch } from 'vue'

const props = defineProps({ dark: Boolean })
const canvas = ref(null)
let context,
  pointerTarget,
  resizeObserver,
  motionQuery,
  frame,
  width = 0,
  height = 0,
  pointer = 0

function render(time = 0) {
  if (!context) return
  context.clearRect(0, 0, width, height)
  const dark = props.dark
  context.strokeStyle = dark
    ? 'rgba(172,213,194,0.065)'
    : 'rgba(39,105,80,0.08)'
  context.lineWidth = 0.7
  for (let x = 32; x < width; x += 64) {
    for (let y = 35; y < height; y += 64) {
      context.beginPath()
      context.moveTo(x - 2, y)
      context.lineTo(x + 2, y)
      context.moveTo(x, y - 2)
      context.lineTo(x, y + 2)
      context.stroke()
    }
  }
  const t = motionQuery?.matches ? 0 : time / 16000
  // 三组流线共用一条曲面，亮线沿着路径移动，避免动画改变页面布局。
  for (let line = 0; line < 42; line++) {
    const group = Math.floor(line / 14)
    const color =
      group === 0 ? '29,157,132' : group === 1 ? '73,154,183' : '196,168,86'
    const position = (s) => ({
      x: s * width,
      y:
        height * 0.74 +
        Math.sin(s * Math.PI * 2.1 - 0.5 + group * 0.19) * height * 0.105 +
        Math.cos(s * Math.PI) * height * 0.055 +
        (line - 21) * 3.7 +
        pointer * Math.sin(s * Math.PI) * 5,
    })
    context.beginPath()
    for (let step = 0; step <= 100; step++) {
      const p = position(step / 100)
      if (!step) context.moveTo(p.x, p.y)
      else context.lineTo(p.x, p.y)
    }
    context.lineWidth = 0.85
    context.strokeStyle = `rgba(${color},${dark ? 0.12 : 0.105})`
    context.stroke()
    if (line % 6 !== 0) continue
    const progress = ((t + line * 0.037) % 1.24) - 0.12
    context.beginPath()
    for (let step = 0; step <= 18; step++) {
      const p = position(progress + step / 130)
      if (!step) context.moveTo(p.x, p.y)
      else context.lineTo(p.x, p.y)
    }
    context.lineWidth = 1.25
    context.strokeStyle = `rgba(${color},${dark ? 0.65 : 0.4})`
    context.stroke()
  }
}

function animate(time) {
  render(time)
  if (!document.hidden && !motionQuery.matches)
    frame = requestAnimationFrame(animate)
}

function restart() {
  cancelAnimationFrame(frame)
  render()
  if (!document.hidden && !motionQuery.matches)
    frame = requestAnimationFrame(animate)
}

function resize() {
  const box = canvas.value.getBoundingClientRect()
  width = box.width
  height = box.height
  const ratio = Math.min(devicePixelRatio || 1, 2)
  canvas.value.width = Math.round(width * ratio)
  canvas.value.height = Math.round(height * ratio)
  context.setTransform(ratio, 0, 0, ratio, 0, 0)
  render()
}

function onPointer(event) {
  pointer = (event.clientX / innerWidth - 0.5) * 2
}
watch(
  () => props.dark,
  () => render(),
)
onMounted(() => {
  context = canvas.value.getContext('2d')
  if (!context) return
  motionQuery = matchMedia('(prefers-reduced-motion: reduce)')
  resizeObserver = new ResizeObserver(resize)
  resizeObserver.observe(canvas.value)
  motionQuery.addEventListener('change', restart)
  document.addEventListener('visibilitychange', restart)
  pointerTarget = canvas.value.parentElement
  pointerTarget.addEventListener('pointermove', onPointer)
  resize()
  restart()
})
onUnmounted(() => {
  cancelAnimationFrame(frame)
  resizeObserver?.disconnect()
  motionQuery?.removeEventListener('change', restart)
  document.removeEventListener('visibilitychange', restart)
  pointerTarget?.removeEventListener('pointermove', onPointer)
})
</script>
