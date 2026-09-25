<template>
  <div
    class="kc-site auth-page"
    :data-theme="isDark ? 'dark' : 'light'"
    :lang="locale === 'zh' ? 'zh-CN' : 'en'"
  >
    <header class="auth-header">
      <div class="container auth-header-inner">
        <RouterLink class="brand" to="/" aria-label="KeepCoding API">
          <span class="brand-symbol" aria-hidden="true"
            ><Icon name="terminal" size="md"
          /></span>
          <span class="brand-name">KeepCoding <span>API</span></span>
        </RouterLink>
        <div class="auth-header-actions">
          <RouterLink
            class="auth-home-link"
            to="/"
            :title="c.home"
            :aria-label="c.home"
          >
            <Icon name="arrowLeft" size="sm" /><span>{{ c.home }}</span>
          </RouterLink>
          <span class="auth-header-divider" aria-hidden="true"></span>
          <div ref="localeRoot" class="locale-control auth-locale-control">
            <button
              class="icon-button"
              :title="tr('选择语言', 'Choose language')"
              :aria-label="tr('选择语言', 'Choose language')"
              :aria-expanded="localeOpen"
              aria-controls="auth-language-menu"
              @click="localeOpen = !localeOpen"
            >
              <Icon name="globe" size="sm" />
            </button>
            <div
              v-if="localeOpen"
              id="auth-language-menu"
              class="language-menu"
              role="menu"
            >
              <button
                v-for="option in languages"
                :key="option.code"
                role="menuitem"
                @click="setLocale(option.code)"
              >
                {{ option.label
                }}<Icon v-if="locale === option.code" name="check" size="sm" />
              </button>
            </div>
          </div>
          <button
            class="icon-button"
            :title="themeLabel"
            :aria-label="themeLabel"
            :disabled="themeChanging"
            @click="toggleTheme"
          >
            <Icon :name="isDark ? 'sun' : 'moon'" size="sm" />
          </button>
        </div>
      </div>
    </header>

    <main class="auth-main">
      <HeroFlow :dark="isDark" class="auth-flow" />
      <div class="auth-center">
        <section class="auth-panel" :aria-labelledby="'auth-title'">
          <div
            v-if="isRegister && showVerificationStep"
            class="auth-progress"
            :aria-label="tr('注册进度', 'Registration progress')"
          >
            <span
              :class="{ active: !verifying, complete: verifying }"
              :aria-current="!verifying ? 'step' : undefined"
            >
              <b
                ><Icon v-if="verifying" name="check" size="xs" /><template
                  v-else
                  >1</template
                ></b
              >{{ c.accountStep }}
            </span>
            <i aria-hidden="true"></i>
            <span
              :class="{ active: verifying }"
              :aria-current="verifying ? 'step' : undefined"
              ><b>2</b>{{ c.verifyStep }}</span
            >
          </div>

          <div class="auth-heading">
            <div class="auth-heading-top">
              <p class="section-label">
                <span></span
                >{{
                  verifying
                    ? c.verifyLabel
                    : isRegister
                      ? c.registerLabel
                      : c.loginLabel
                }}
              </p>
              <span
                class="auth-heading-icon"
                :class="{ amber: isRegister }"
                aria-hidden="true"
                ><Icon
                  :name="verifying ? 'mail' : isRegister ? 'userPlus' : 'login'"
                  size="md"
              /></span>
            </div>
            <h1 id="auth-title" ref="heading" tabindex="-1">
              {{
                verifying
                  ? c.verifyTitle
                  : isRegister
                    ? c.registerTitle
                    : c.loginTitle
              }}
            </h1>
            <p class="auth-description">
              {{
                verifying
                  ? c.verifyDescription
                  : isRegister
                    ? c.registerDescription
                    : c.loginDescription
              }}
            </p>
          </div>

          <slot :c="c" :open-support="openSupport" />

          <div v-if="!hideAlternative" class="auth-alternative">
            <button
              v-if="verifying"
              class="auth-inline-link auth-back-step"
              type="button"
              @click="$emit('back')"
            >
              <Icon name="arrowLeft" size="sm" />{{ c.backToRegister }}
            </button>
            <template v-else
              ><span>{{ isRegister ? c.hasAccount : c.noAccount }}</span
              ><RouterLink :to="isRegister ? '/login' : '/register'"
                >{{ isRegister ? c.signIn : c.createAccount
                }}<Icon name="arrowRight" size="xs" /></RouterLink
            ></template>
          </div>
        </section>

        <div class="auth-ecosystem" aria-label="AI platforms">
          <div class="auth-provider-row">
            <span v-for="provider in providers" :key="provider.name"
              ><img
                :src="provider.logo"
                alt=""
                :class="{
                  'mono-logo': !provider.color,
                  'claude-logo': provider.name === 'Claude',
                }"
              />{{ provider.name }}</span
            >
          </div>
          <div class="auth-service-row">
            <RouterLink
              v-for="(service, index) in c.services"
              :key="index"
              :to="`/home#${['api', 'subscriptions', 'verification'][index]}`"
              >{{ service }}</RouterLink
            >
          </div>
        </div>
      </div>
    </main>

    <footer class="auth-footer container">
      <span>© {{ year }} KeepCoding API</span
      ><button class="auth-inline-link" @click="openSupport('support')">
        <Icon name="chatBubble" size="sm" />{{ c.support }}
      </button>
    </footer>

    <dialog
      ref="supportDialog"
      class="contact-dialog auth-support-dialog"
      aria-labelledby="auth-support-title"
      @click="closeOnBackdrop"
    >
      <div class="dialog-content">
        <button
          class="icon-button dialog-close"
          :aria-label="c.close"
          @click="supportDialog.close()"
        >
          <Icon name="x" size="md" />
        </button>
        <p class="section-label"><span></span>{{ c.support }}</p>
        <h2 id="auth-support-title">
          {{ supportMode === 'recover' ? c.recoverTitle : c.support }}
        </h2>
        <p class="dialog-subtitle">
          {{
            supportMode === 'recover'
              ? c.recoverDescription
              : c.supportDescription
          }}
        </p>
        <div class="contact-details">
          <div class="contact-qr">
            <img :src="wechatImage" :alt="c.scan" /><span>{{ c.scan }}</span>
          </div>
          <div class="contact-copy">
            <label for="auth-wechat">WeChat</label
            ><input
              id="auth-wechat"
              :value="site.wechat"
              readonly
              @focus="$event.target.select()"
            /><button class="button button-ink" @click="copyWechat">
              <Icon :name="copied ? 'check' : 'copy'" size="sm" />{{
                copied ? c.copied : c.copyWechat
              }}
            </button>
            <p v-if="copyFailed" class="auth-field-error" role="status">
              {{ c.copyFailed }}
            </p>
          </div>
        </div>
      </div>
    </dialog>
  </div>
</template>
<script setup>
import { computed, nextTick, onMounted, onUnmounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
import Icon from '@/components/icons/Icon.vue'
import HeroFlow from './HeroFlow.vue'
import { site, providers } from './content'
import { authCopy } from './auth-content'
import { usePublicPage, syncPublicLocale, syncPublicTheme } from './usePublicPage'
const props = defineProps({ isRegister: Boolean, verifying: Boolean, hideAlternative: Boolean, showVerificationStep: { type: Boolean, default: true } })
defineEmits(['back'])
const { locale, isDark } = usePublicPage()
const c = computed(() => authCopy[locale.value])
const tr = (zh, en) => locale.value === 'zh' ? zh : en
const wechatImage = '/contact/wechat.png'
const heading = ref(null)
const themeChanging = ref(false), localeOpen = ref(false), localeRoot = ref(null)
const supportDialog = ref(null), supportMode = ref('support'), copied = ref(false), copyFailed = ref(false)
const year = new Date().getFullYear()
const languages = [{ code: 'zh', label: '简体中文' }, { code: 'en', label: 'English' }]
const themeLabel = computed(() => tr(isDark.value ? '切换浅色模式' : '切换深色模式', isDark.value ? 'Switch to light mode' : 'Switch to dark mode'))
let copyTimer
function setLocale(value) {
  locale.value = value
  localeOpen.value = false
  syncPublicLocale(value)
}

function applyTheme(value) {
  isDark.value = value
  document.documentElement.dataset.theme = value ? 'dark' : 'light'
  document
    .querySelector('meta[name="theme-color"]')
    ?.setAttribute('content', value ? '#071510' : '#f6faf8')
  syncPublicTheme(value)
}

async function toggleTheme(event) {
  if (themeChanging.value) return
  const next = !isDark.value
  if (
    !document.startViewTransition ||
    matchMedia('(prefers-reduced-motion: reduce)').matches
  ) {
    applyTheme(next)
    return
  }
  const rect = event.currentTarget.getBoundingClientRect()
  document.documentElement.style.setProperty(
    '--theme-origin',
    `${rect.left + rect.width / 2}px ${rect.top + rect.height / 2}px`,
  )
  themeChanging.value = true
  try {
    await document.startViewTransition(async () => {
      applyTheme(next)
      await nextTick()
    }).finished
  } finally {
    themeChanging.value = false
    document.documentElement.style.removeProperty('--theme-origin')
  }
}

function openSupport(mode) {
  supportMode.value = mode
  copied.value = false
  copyFailed.value = false
  supportDialog.value.showModal()
}

function closeOnBackdrop(event) {
  if (event.target !== supportDialog.value) return
  const bounds = supportDialog.value.getBoundingClientRect()
  if (
    event.clientX < bounds.left ||
    event.clientX > bounds.right ||
    event.clientY < bounds.top ||
    event.clientY > bounds.bottom
  )
    supportDialog.value.close()
}

async function copyWechat() {
  try {
    await navigator.clipboard.writeText(site.wechat)
    copied.value = true
    copyFailed.value = false
    clearTimeout(copyTimer)
    copyTimer = setTimeout(() => {
      copied.value = false
    }, 2200)
  } catch {
    copyFailed.value = true
  }
}

function onDocumentClick(event) {
  if (!localeRoot.value?.contains(event.target)) localeOpen.value = false
}
function onKeydown(event) {
  if (event.key === 'Escape') localeOpen.value = false
}


onMounted(() => {
  applyTheme(isDark.value)
  document.addEventListener('click', onDocumentClick)
  document.addEventListener('keydown', onKeydown)
})
onUnmounted(() => {
  clearTimeout(copyTimer)
  document.removeEventListener('click', onDocumentClick)
  document.removeEventListener('keydown', onKeydown)
})
</script>
