<template>
  <AppLayout>
    <main class="recharge-page">
      <section class="recharge-hero">
        <p class="eyebrow">Sub2API · Recharge</p>
        <h1><em>{{ t('payment.recharge.headlineAny') }}</em><br>{{ t('payment.recharge.headlineRefund') }}</h1>
        <p class="intro">{{ t('payment.recharge.intro') }}</p>
        <div class="promises">
          <div class="promise">
            <div class="promise-icon" aria-hidden="true">¥</div>
            <div>
              <b>{{ t('payment.recharge.anyAmountTitle') }}</b>
              <span>{{ t('payment.recharge.anyAmountDesc') }}</span>
            </div>
          </div>
          <div class="promise">
            <div class="promise-icon" aria-hidden="true">↩</div>
            <div>
              <b>{{ t('payment.recharge.refundTitle') }}</b>
              <span>{{ t('payment.recharge.refundDesc') }}</span>
            </div>
          </div>
        </div>
      </section>

      <aside class="wechat-card">
        <div class="wechat-head">
          <svg viewBox="0 0 24 24" fill="currentColor" aria-hidden="true"><path d="M9 4C4.6 4 1 7 1 10.7c0 2.1 1.2 4 3 5.2L3.3 18l2.6-1.3c.9.3 2 .4 3 .4h.6a5.6 5.6 0 0 1-.2-1.5c0-3.5 3.4-6.3 7.6-6.3h.6C16.7 6.2 13.2 4 9 4Zm-3 4.4a1 1 0 1 1 0 2 1 1 0 0 1 0-2Zm6 0a1 1 0 1 1 0 2 1 1 0 0 1 0-2Zm4.9 2.2c-3.5 0-6.2 2.3-6.2 5.1s2.8 5.2 6.2 5.2c.7 0 1.5-.1 2.2-.4l2.1 1.1-.6-1.8c1.5-1 2.5-2.5 2.5-4.1 0-2.8-2.8-5.1-6.2-5.1Zm-2.2 2.9a.8.8 0 1 1 0 1.6.8.8 0 0 1 0-1.6Zm4.4 0a.8.8 0 1 1 0 1.6.8.8 0 0 1 0-1.6Z"/></svg>
          {{ t('payment.recharge.wechatTitle') }}
        </div>
        <div class="wechat-body">
          <img class="qr-image" src="/contact/wechat.png" :alt="t('payment.recharge.wechatQrAlt')">
          <p class="account">{{ WECHAT_ID }}</p>
          <button class="copy-button" type="button" @click="copyContact(WECHAT_ID)">
            {{ t('payment.recharge.copyWechat') }}
          </button>
        </div>
      </aside>
    </main>

    <div v-if="toastMessage" class="copy-toast" role="status" aria-live="polite">
      {{ toastMessage }}
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { onUnmounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'

const WECHAT_ID = 'Taokkkkkkkboy'

const { t } = useI18n()
const toastMessage = ref('')
let toastTimer: ReturnType<typeof setTimeout> | undefined

async function copyContact(value: string) {
  try {
    if (navigator.clipboard && window.isSecureContext) {
      await navigator.clipboard.writeText(value)
    } else {
      const input = document.createElement('textarea')
      input.value = value
      input.setAttribute('readonly', '')
      input.style.position = 'fixed'
      input.style.opacity = '0'
      document.body.appendChild(input)
      input.select()
      document.execCommand('copy')
      input.remove()
    }
    toastMessage.value = t('payment.recharge.copied', { value })
  } catch {
    toastMessage.value = t('payment.recharge.copyManually', { value })
  }

  if (toastTimer) clearTimeout(toastTimer)
  toastTimer = setTimeout(() => {
    toastMessage.value = ''
  }, 1800)
}

onUnmounted(() => {
  if (toastTimer) clearTimeout(toastTimer)
})
</script>

<style scoped>
.recharge-page {
  --page-bg: #f4f6f7;
  --card-bg: #ffffff;
  --ink: #18222d;
  --muted: #687582;
  --line: #dfe6eb;
  --accent: #1b8f9a;
  --wechat: #07c160;
  --wechat-soft: #e9f9f0;
  display: grid;
  min-height: calc(100vh - 9rem);
  margin: -0.25rem;
  padding: clamp(2rem, 5vw, 4rem) clamp(1rem, 4vw, 3rem);
  border-radius: 1.5rem;
  color: var(--ink);
  background: var(--page-bg);
  gap: clamp(1.5rem, 4vw, 3.5rem);
  grid-template-columns: minmax(0, 1.1fr) minmax(18rem, 0.9fr);
  align-items: center;
}

:global(.dark) .recharge-page {
  --page-bg: #111827;
  --card-bg: #1f2937;
  --ink: #f3f4f6;
  --muted: #9ca3af;
  --line: #374151;
  --accent: #2dd4bf;
  --wechat: #22c55e;
  --wechat-soft: rgba(34, 197, 94, 0.12);
}

.eyebrow {
  margin: 0 0 1rem;
  color: var(--accent);
  font-size: 0.75rem;
  font-weight: 700;
  letter-spacing: 0.16em;
  text-transform: uppercase;
}

h1 {
  margin: 0;
  font-size: clamp(2.4rem, 5.5vw, 4rem);
  line-height: 1.12;
  letter-spacing: 0;
}

h1 em {
  color: var(--wechat);
  font-style: normal;
}

.intro {
  max-width: 30rem;
  margin: 1.25rem 0 2rem;
  color: var(--muted);
  font-size: 1.05rem;
  line-height: 1.75;
}

.promises {
  display: grid;
  max-width: 30rem;
  gap: 0.9rem;
}

.promise {
  display: flex;
  align-items: flex-start;
  gap: 0.9rem;
  padding: 1rem 1.1rem;
  border: 1px solid var(--line);
  border-radius: 0.9rem;
  background: var(--card-bg);
}

.promise-icon {
  display: grid;
  flex: none;
  width: 2.25rem;
  height: 2.25rem;
  place-items: center;
  border-radius: 0.6rem;
  color: var(--wechat);
  background: var(--wechat-soft);
  font-size: 1.1rem;
  font-weight: 700;
}

.promise b {
  display: block;
  margin-bottom: 0.2rem;
}

.promise span {
  color: var(--muted);
  font-size: 0.88rem;
  line-height: 1.5;
}

.wechat-card {
  overflow: hidden;
  border: 1px solid var(--line);
  border-radius: 1.25rem;
  background: var(--card-bg);
  box-shadow: 0 1.25rem 3rem rgba(43, 58, 68, 0.12);
}

.wechat-head {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  padding: 1rem 1.25rem;
  color: #ffffff;
  background: var(--wechat);
  font-weight: 700;
}

.wechat-head svg {
  width: 1.4rem;
  height: 1.4rem;
}

.wechat-body {
  padding: 1.5rem;
  text-align: center;
}

.qr-image {
  display: block;
  width: min(100%, 17rem);
  height: auto;
  margin: 0 auto;
  border-radius: 0.5rem;
}

.account {
  margin: 1.25rem 0 0.25rem;
  overflow-wrap: anywhere;
  font-family: "SFMono-Regular", Consolas, monospace;
  font-size: 1.05rem;
}

.copy-button {
  width: 100%;
  margin-top: 0.5rem;
  padding: 0.75rem;
  border: 0;
  border-radius: 0.7rem;
  color: var(--wechat);
  background: var(--wechat-soft);
  font: inherit;
  font-weight: 700;
  cursor: pointer;
}

.copy-button:hover { filter: brightness(0.96); }
.copy-button:focus-visible { outline: 2px solid var(--wechat); outline-offset: 3px; }

.copy-toast {
  position: fixed;
  right: 1.25rem;
  bottom: 1.25rem;
  z-index: 50;
  padding: 0.7rem 1rem;
  border: 1px solid #b9ded0;
  border-radius: 999px;
  color: #267a52;
  background: #f0faf4;
  box-shadow: 0 0.5rem 1.5rem rgba(32, 48, 61, 0.12);
  font-size: 0.85rem;
}

@media (max-width: 820px) {
  .recharge-page { grid-template-columns: 1fr; }
  .wechat-card { width: 100%; max-width: 24rem; margin: 0 auto; }
}

@media (max-width: 560px) {
  .recharge-page {
    margin: -0.5rem;
    padding: 2rem 0.75rem 2.5rem;
    border-radius: 1rem;
  }
}
</style>
