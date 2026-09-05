<template>
  <AppLayout>
    <main class="recharge-page">
      <header class="recharge-hero">
        <p class="eyebrow">Sub2API · Account</p>
        <h1>充值与订阅</h1>
        <p class="intro">站点小范围运营，充值请联系站长。</p>
      </header>

      <section class="contact-grid" aria-label="站长联系方式">
        <article class="contact-card">
          <div class="contact-copy">
            <div>
              <p class="contact-label">联系渠道</p>
              <h2>QQ 联系</h2>
              <p class="account">2052436429</p>
              <button class="copy-button" type="button" @click="copyContact('2052436429')">
                复制 QQ <span aria-hidden="true">↗</span>
              </button>
            </div>
            <p class="contact-note">备注：站点充值</p>
          </div>
          <div class="qr-panel">
            <img class="qr-image" src="/contact/qq.png" alt="QQ 联系二维码">
          </div>
        </article>

        <article class="contact-card">
          <div class="contact-copy">
            <div>
              <p class="contact-label">联系渠道</p>
              <h2>微信联系</h2>
              <p class="account">Taokkkkkkkboy</p>
              <button class="copy-button" type="button" @click="copyContact('Taokkkkkkkboy')">
                复制微信号 <span aria-hidden="true">↗</span>
              </button>
            </div>
            <p class="contact-note">备注：站点充值</p>
          </div>
          <div class="qr-panel">
            <img class="qr-image" src="/contact/wechat.png" alt="微信联系二维码">
          </div>
        </article>
      </section>

      <p class="page-footer">充值到账后可在账户余额中查看变更记录。</p>
    </main>

    <div v-if="toastMessage" class="copy-toast" role="status" aria-live="polite">
      {{ toastMessage }}
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { onUnmounted, ref } from 'vue'
import AppLayout from '@/components/layout/AppLayout.vue'

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
    toastMessage.value = `已复制：${value}`
  } catch {
    toastMessage.value = `请手动复制：${value}`
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
  --qr-bg: #edf3f7;
  min-height: calc(100vh - 9rem);
  margin: -0.25rem;
  padding: clamp(2rem, 5vw, 4.5rem) clamp(1rem, 4vw, 3rem) 3.5rem;
  border-radius: 1.5rem;
  color: var(--ink);
  background: var(--page-bg);
}

.recharge-hero {
  max-width: 42rem;
  margin: 0 auto clamp(2rem, 5vw, 3.5rem);
  text-align: center;
}

.eyebrow {
  margin: 0 0 0.75rem;
  color: var(--accent);
  font-size: 0.75rem;
  font-weight: 700;
  letter-spacing: 0.16em;
  text-transform: uppercase;
}

h1 {
  margin: 0 0 0.85rem;
  font-size: clamp(2.25rem, 5vw, 3.75rem);
  line-height: 1.1;
  letter-spacing: 0;
}

.intro {
  margin: 0;
  color: var(--muted);
  font-size: 1.05rem;
  line-height: 1.75;
}

.contact-grid {
  display: grid;
  max-width: 70rem;
  margin: 0 auto;
  gap: 1.5rem;
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.contact-card {
  display: grid;
  min-width: 0;
  min-height: 31rem;
  overflow: hidden;
  border: 1px solid var(--line);
  background: var(--card-bg);
  box-shadow: 0 1rem 2.5rem rgba(43, 58, 68, 0.08);
  grid-template-columns: minmax(0, 0.92fr) minmax(15rem, 1.08fr);
}

.contact-copy {
  display: flex;
  min-width: 0;
  flex-direction: column;
  justify-content: space-between;
  padding: clamp(1.75rem, 4vw, 3rem) clamp(1.5rem, 4vw, 2.75rem) 2.5rem;
}

.contact-label {
  margin: 0 0 1.5rem;
  color: var(--accent);
  font-size: 0.75rem;
  font-weight: 700;
  letter-spacing: 0.12em;
  text-transform: uppercase;
}

h2 {
  margin: 0 0 0.7rem;
  font-size: clamp(1.8rem, 3vw, 2.45rem);
  line-height: 1.2;
}

.account {
  margin: 0;
  overflow-wrap: anywhere;
  color: var(--ink);
  font-family: "SFMono-Regular", Consolas, monospace;
  font-size: 1rem;
}

.copy-button {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  margin-top: 2rem;
  padding: 0.6rem 0;
  border: 0;
  color: var(--accent);
  background: transparent;
  font: inherit;
  font-size: 0.9rem;
  font-weight: 700;
  cursor: pointer;
}

.copy-button:hover { color: var(--ink); }
.copy-button:focus-visible { outline: 2px solid var(--accent); outline-offset: 4px; }

.contact-note {
  margin: 2rem 0 0;
  padding-top: 1rem;
  border-top: 1px solid var(--line);
  color: var(--ink);
  font-size: 0.9rem;
  font-weight: 600;
}

.qr-panel {
  display: grid;
  min-width: 0;
  place-items: center;
  padding: 2rem;
  background: var(--qr-bg);
}

.qr-image {
  display: block;
  width: min(100%, 22rem);
  height: auto;
  max-height: 25rem;
  object-fit: contain;
}

.page-footer {
  max-width: 70rem;
  margin: 2rem auto 0;
  color: var(--muted);
  font-size: 0.82rem;
  text-align: center;
}

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
  .contact-grid { grid-template-columns: 1fr; }
  .contact-card { min-height: 0; }
}

@media (max-width: 560px) {
  .recharge-page {
    margin: -0.5rem;
    padding: 2rem 0.75rem 2.5rem;
    border-radius: 1rem;
  }

  .contact-card { display: flex; flex-direction: column; }
  .contact-copy { padding: 2rem 1.5rem 1.5rem; }
  .qr-panel { order: 2; padding: 1.5rem; }
  .contact-note { order: 3; margin-top: 1.5rem; }
  .page-footer { margin-top: 1.5rem; }
}
</style>
