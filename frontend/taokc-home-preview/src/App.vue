<template>
  <div
    class="kc-site"
    :data-theme="isDark ? 'dark' : 'light'"
    :lang="locale === 'zh' ? 'zh-CN' : 'en'"
  >
    <header class="site-header" :class="{ 'is-scrolled': scrolled }">
      <div class="container header-inner">
        <a
          class="brand"
          href="#top"
          @click="menuOpen = false"
          aria-label="KeepCoding API"
        >
          <span class="brand-symbol" aria-hidden="true"
            ><Icon name="terminal" size="md"
          /></span>
          <span class="brand-name">KeepCoding <span>API</span></span>
        </a>
        <nav class="desktop-nav" :aria-label="tr('主导航', 'Main navigation')">
          <a
            v-for="link in navLinks"
            :key="link.id"
            :href="`#${link.id}`"
            :class="{ active: activeSection === link.id }"
            >{{ link.label }}</a
          >
        </nav>
        <div class="header-actions">
          <div ref="localeRoot" class="locale-control">
            <button
              class="icon-button language-toggle"
              :title="tr('选择语言', 'Choose language')"
              :aria-label="tr('选择语言', 'Choose language')"
              :aria-expanded="localeOpen"
              aria-controls="language-menu"
              @click="localeOpen = !localeOpen"
            >
              <Icon name="globe" size="sm" />
            </button>
            <div
              v-if="localeOpen"
              id="language-menu"
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
            class="icon-button theme-toggle"
            :title="themeLabel"
            :aria-label="themeLabel"
            :disabled="themeChanging"
            @click="toggleTheme"
          >
            <Icon :name="isDark ? 'sun' : 'moon'" size="sm" />
          </button>
          <a
            class="button button-small button-ink console-link"
            href="/login"
            :aria-label="c.console"
            ><span>{{ c.console }}</span
            ><Icon name="arrowUp" size="sm" class="diagonal-arrow"
          /></a>
          <button
            class="icon-button menu-toggle"
            :aria-label="
              tr(
                menuOpen ? '关闭菜单' : '打开菜单',
                menuOpen ? 'Close menu' : 'Open menu',
              )
            "
            :aria-expanded="menuOpen"
            aria-controls="mobile-menu"
            @click="menuOpen = !menuOpen"
          >
            <Icon :name="menuOpen ? 'x' : 'menu'" size="md" />
          </button>
        </div>
      </div>
      <nav
        v-if="menuOpen"
        id="mobile-menu"
        class="mobile-menu"
        :aria-label="tr('移动端导航', 'Mobile navigation')"
      >
        <a
          v-for="link in navLinks"
          :key="link.id"
          :href="`#${link.id}`"
          @click="menuOpen = false"
          >{{ link.label }}<Icon name="arrowRight" size="sm"
        /></a>
        <div class="mobile-languages">
          <button :class="{ active: locale === 'zh' }" @click="setLocale('zh')">
            中文</button
          ><button
            :class="{ active: locale === 'en' }"
            @click="setLocale('en')"
          >
            English
          </button>
        </div>
      </nav>
    </header>

    <main id="top">
      <section class="hero" :aria-label="tr('AI 服务', 'AI services')">
        <HeroFlow :dark="isDark" />
        <div class="container hero-content">
          <p class="hero-eyebrow" data-reveal>
            <span class="eyebrow-mark"></span>{{ c.eyebrow }}
          </p>
          <h1>
            <span class="hero-headline-line" data-reveal :style="delay(1)">{{
              c.headline
            }}</span>
            <span
              class="hero-headline-line hero-headline-accent"
              data-reveal
              :style="delay(2)"
              >{{ c.headlineAccent }}</span
            >
          </h1>
          <p class="hero-description" data-reveal :style="delay(3)">
            {{ c.heroDescription }}
          </p>
          <div class="hero-actions" data-reveal :style="delay(4)">
            <a href="#subscriptions" class="button button-ink"
              >{{ c.browse }}<Icon name="arrowRight" size="sm" /></a
            ><a :href="`${site.origin}/keys`" class="button button-glass"
              ><Icon name="terminal" size="sm" />{{ c.getKey }}</a
            >
          </div>
        </div>
        <div class="container service-rail" data-reveal :style="delay(5)">
          <a href="#subscriptions" class="service-entry mint"
            ><span class="service-icon"><Icon name="sparkles" size="md" /></span
            ><span class="service-entry-copy"
              ><strong>{{ c.subscriptions }}</strong
              ><span>{{ c.serviceSub }}</span></span
            ><span class="service-number">01</span
            ><Icon name="arrowRight" size="sm" class="service-arrow"
          /></a>
          <a href="#api" class="service-entry blue"
            ><span class="service-icon"><Icon name="terminal" size="md" /></span
            ><span class="service-entry-copy"
              ><strong>{{ c.api }}</strong
              ><span>{{ c.serviceApi }}</span></span
            ><span class="service-number">02</span
            ><Icon name="arrowRight" size="sm" class="service-arrow"
          /></a>
          <a href="#verification" class="service-entry apricot"
            ><span class="service-icon"
              ><Icon name="chatBubble" size="md" /></span
            ><span class="service-entry-copy"
              ><strong>{{ c.sms }}</strong
              ><span>{{ c.serviceSms }}</span></span
            ><span class="service-number">03</span
            ><Icon name="arrowRight" size="sm" class="service-arrow"
          /></a>
        </div>
      </section>

      <div class="ecosystem-band">
        <div class="container ecosystem-inner" data-reveal>
          <span class="ecosystem-label">{{ c.ecosystem }}</span>
          <div class="provider-list">
            <span
              v-for="provider in providers"
              :key="provider.name"
              class="provider-item"
              ><img
                :src="provider.logo"
                alt=""
                :class="{
                  'mono-logo': !provider.color,
                  'claude-logo': provider.name === 'Claude',
                }"
              /><span>{{ provider.name }}</span></span
            >
          </div>
          <span class="ecosystem-tail"
            >& more <Icon name="arrowUp" size="xs" class="diagonal-arrow"
          /></span>
        </div>
      </div>

      <ToolsMarquee
        :label="c.compatible"
        :pause-label="tr('暂停工具流动', 'Pause tool animation')"
        :resume-label="tr('继续工具流动', 'Resume tool animation')"
      />

      <section id="subscriptions" class="subscriptions-section section-space">
        <div class="container">
          <div class="section-heading" data-reveal>
            <div>
              <p class="section-label"><span></span>{{ c.catalogLabel }}</p>
              <h2>{{ c.catalogTitle }}</h2>
              <p class="section-description">{{ c.catalogDescription }}</p>
            </div>
            <div
              class="filter-tabs"
              role="tablist"
              @keydown="navigateTabs"
              :aria-label="tr('订阅平台', 'Subscription provider')"
            >
              <button
                v-for="filter in filters"
                :key="filter.value"
                role="tab"
                :aria-selected="selectedProvider === filter.value"
                :class="{ active: selectedProvider === filter.value }"
                @click="selectedProvider = filter.value"
              >
                {{ filter.label }}<span>{{ filter.count }}</span>
              </button>
            </div>
          </div>
          <div class="product-grid" aria-live="polite">
            <div
              v-for="(product, index) in filteredProducts"
              :key="product.id"
              class="product-reveal"
              data-reveal
              :style="delay(index)"
            >
              <article class="product-card" :class="product.tone">
                <div class="product-visual">
                  <span class="product-logo"
                    ><img
                      :src="product.logo"
                      :alt="product.provider"
                      :class="{
                        'claude-logo': product.provider === 'Anthropic',
                      }" /></span
                  ><span class="product-tag">{{ product.tag[locale] }}</span
                  ><img
                    class="product-watermark"
                    :src="product.logo"
                    alt=""
                    aria-hidden="true"
                  />
                </div>
                <div class="product-body">
                  <p class="product-provider">
                    {{ product.provider }}<span>{{ c.monthly }}</span>
                  </p>
                  <h3>{{ product.name }}</h3>
                  <p class="product-description">
                    {{ product.description[locale] }}
                  </p>
                  <div class="product-price">
                    <span class="currency">¥</span
                    ><strong>{{ product.price }}</strong
                    ><span>{{ c.perMonth }}</span>
                  </div>
                  <ul class="product-features">
                    <li
                      v-for="feature in product.features[locale]"
                      :key="feature"
                    >
                      <Icon name="check" size="sm" />{{ feature }}
                    </li>
                  </ul>
                  <button
                    class="button product-action"
                    @click="openContact('product', product)"
                  >
                    {{ c.consult }}<Icon name="arrowRight" size="sm" />
                  </button>
                </div>
              </article>
            </div>
          </div>
          <div class="catalog-footer" data-reveal>
            <p><Icon name="infoCircle" size="sm" />{{ c.productNotice }}</p>
            <button class="text-button" @click="openContact('general')">
              {{ c.otherProducts }} <span>{{ c.askUs }}</span
              ><Icon name="arrowRight" size="sm" />
            </button>
          </div>
        </div>
      </section>

      <section id="api" class="api-section section-space">
        <div class="container api-layout">
          <div class="api-copy">
            <p class="section-label" data-reveal>
              <span></span>{{ c.apiLabel }}
            </p>
            <h2 data-reveal :style="delay(1)">{{ c.apiTitle }}</h2>
            <p class="section-description" data-reveal :style="delay(2)">
              {{ c.apiDescription }}
            </p>
            <ul class="api-benefits" data-reveal :style="delay(3)">
              <li v-for="item in c.apiFeatures" :key="item">
                <Icon name="checkCircle" size="sm" />{{ item }}
              </li>
            </ul>
            <div class="api-actions" data-reveal :style="delay(4)">
              <a :href="`${site.origin}/keys`" class="button button-mint"
                >{{ c.getKey }}<Icon name="arrowRight" size="sm" /></a
              ><a
                :href="`${site.origin}/available-channels`"
                class="text-button"
                >{{ c.channels }}<Icon name="externalLink" size="sm"
              /></a>
            </div>
          </div>
          <div class="code-workbench" data-reveal :style="delay(2)">
            <div class="workbench-top">
              <span><Icon name="terminal" size="sm" />{{ c.example }}</span
              ><a v-if="site.docs" :href="site.docs" target="_blank" rel="noopener noreferrer"
                >{{ c.docs }}<Icon name="externalLink" size="xs"
              /></a>
            </div>
            <div
              class="code-tabs"
              role="tablist"
              @keydown="navigateTabs"
              :aria-label="tr('代码语言', 'Code language')"
            >
              <button
                v-for="tab in codeTabs"
                :key="tab"
                role="tab"
                :aria-selected="codeTab === tab"
                :class="{ active: codeTab === tab }"
                @click="codeTab = tab"
              >
                {{ tab }}</button
              ><button
                class="icon-button"
                :title="c.copyCode"
                :aria-label="c.copyCode"
                @click="copyText(codeSample, 'code')"
              >
                <Icon :name="copied === 'code' ? 'check' : 'copy'" size="sm" />
              </button>
            </div>
            <pre class="code-sample"><code>{{ codeSample }}</code></pre>
            <div class="endpoint-bar">
              <span>BASE URL</span><code>{{ site.origin }}/v1</code
              ><button
                class="icon-button"
                :title="c.copyEndpoint"
                :aria-label="c.copyEndpoint"
                @click="copyText(`${site.origin}/v1`, 'endpoint')"
              >
                <Icon
                  :name="copied === 'endpoint' ? 'check' : 'copy'"
                  size="sm"
                />
              </button>
            </div>
          </div>
        </div>
      </section>

      <section id="verification" class="verification-section section-space">
        <div class="container verification-layout">
          <div>
            <p class="section-label amber" data-reveal>
              <span></span>{{ c.smsLabel }}
            </p>
            <h2 data-reveal :style="delay(1)">{{ c.smsTitle }}</h2>
            <p class="section-description" data-reveal :style="delay(2)">
              {{ c.smsDescription }}
            </p>
            <ol class="verification-steps">
              <li
                v-for="(step, index) in c.steps"
                :key="index"
                data-reveal
                :style="delay(index)"
              >
                <span>{{ String(index + 1).padStart(2, '0') }}</span>
                <div>
                  <h3>{{ step.title }}</h3>
                  <p>{{ step.description }}</p>
                </div>
              </li>
            </ol>
            <button
              class="text-button sms-contact"
              data-reveal
              @click="openContact('sms')"
            >
              {{ c.contact }}<Icon name="arrowRight" size="sm" />
            </button>
          </div>
          <SmsFlow :copy="c.smsFlow" data-reveal :style="delay(2)" />
        </div>
      </section>

      <section id="faq" class="faq-section section-space">
        <div class="container faq-layout">
          <div data-reveal>
            <p class="section-label"><span></span>{{ c.faqLabel }}</p>
            <h2>{{ c.faqTitle }}</h2>
            <button class="text-button" @click="openContact('general')">
              {{ c.contact }}<Icon name="arrowRight" size="sm" />
            </button>
          </div>
          <div class="faq-list">
            <article
              v-for="(item, index) in c.faqs"
              :key="index"
              class="faq-item"
              data-reveal
              :style="delay(index)"
            >
              <h3>
                <button
                  :aria-expanded="openFaq === index"
                  :aria-controls="`faq-${index}`"
                  @click="openFaq = openFaq === index ? -1 : index"
                >
                  {{ item.q
                  }}<Icon
                    :name="openFaq === index ? 'chevronUp' : 'plus'"
                    size="sm"
                  />
                </button>
              </h3>
              <div
                v-if="openFaq === index"
                :id="`faq-${index}`"
                class="faq-answer"
              >
                {{ item.a }}
              </div>
            </article>
          </div>
        </div>
      </section>
    </main>

    <footer class="site-footer">
      <div class="container footer-main" data-reveal>
        <div>
          <a class="brand" href="#top"
            ><span class="brand-symbol" aria-hidden="true"
              ><Icon name="terminal" size="md" /></span
            ><span class="brand-name">KeepCoding <span>API</span></span></a
          >
          <p>{{ c.footerDescription }}</p>
        </div>
        <div class="footer-links">
          <h3>{{ c.services }}</h3>
          <a href="#subscriptions">{{ c.subscriptions }}</a
          ><a href="#api">{{ c.api }}</a
          ><a href="#verification">{{ c.sms }}</a>
        </div>
        <div class="footer-links">
          <h3>{{ c.support }}</h3>
          <a v-if="site.docs" :href="site.docs" target="_blank" rel="noopener noreferrer"
            >{{ c.docs }}<Icon name="externalLink" size="xs" /></a
          ><button @click="openContact('general')">
            {{ c.contact }}<Icon name="arrowRight" size="xs" /></button
          ><span class="wechat-footer">WeChat / {{ site.wechat }}</span>
        </div>
      </div>
      <div class="container footer-bottom">
        <span>© {{ year }} KeepCoding API</span
        ><span>Keep creating. Keep coding.</span
        ><a
          href="#top"
          class="icon-button"
          :title="tr('回到顶部', 'Back to top')"
          :aria-label="tr('回到顶部', 'Back to top')"
          ><Icon name="arrowUp" size="sm"
        /></a>
      </div>
    </footer>

    <button
      class="support-fab"
      :title="c.contact"
      :aria-label="c.contact"
      @click="openContact('general')"
    >
      <Icon name="chatBubble" size="md" /><span>{{ c.contact }}</span>
    </button>
    <Transition name="toast"
      ><div v-if="toastMessage" class="toast-message" role="status">
        <Icon name="checkCircle" size="sm" />{{ toastMessage }}
      </div></Transition
    >

    <dialog
      ref="contactDialog"
      class="contact-dialog"
      aria-labelledby="contact-title"
      @click="closeOnBackdrop"
      @close="contactOpen = false"
    >
      <div v-if="contactOpen" class="dialog-content">
        <button
          class="icon-button dialog-close"
          :aria-label="c.close"
          @click="contactDialog.close()"
        >
          <Icon name="x" size="md" />
        </button>
        <p class="section-label"><span></span>{{ c.contactTitle }}</p>
        <h2 id="contact-title">
          {{
            contactProduct?.name ||
            (contactType === 'sms' ? c.sms : c.contactTitle)
          }}
        </h2>
        <p class="dialog-subtitle">
          {{ contactType === 'sms' ? c.smsContactSubtitle : c.contactSubtitle }}
        </p>
        <div v-if="contactProduct" class="dialog-product">
          <img
            :src="contactProduct.logo"
            alt=""
            :class="{ 'claude-logo': contactProduct.provider === 'Anthropic' }"
          />
          <div>
            <strong
              >¥ {{ contactProduct.price }}
              <small>{{ c.perMonth }}</small></strong
            ><span>{{ c.monthly }}</span>
          </div>
        </div>
        <p v-if="contactProduct" class="dialog-terms">{{ c.productNotice }}</p>
        <div class="contact-details">
          <div class="contact-qr">
            <img
              :src="wechatImage"
              :alt="tr('站长微信二维码', 'Support WeChat QR code')"
            /><span>{{ c.scan }}</span>
          </div>
          <div class="contact-copy">
            <label for="wechat-id">WeChat</label
            ><input
              id="wechat-id"
              :value="site.wechat"
              readonly
              @focus="$event.target.select()"
            /><button
              class="button button-ink"
              @click="copyText(site.wechat, 'wechat')"
            >
              <Icon
                :name="copied === 'wechat' ? 'check' : 'copy'"
                size="sm"
              />{{ copied === 'wechat' ? c.copied : c.copyWechat }}</button
            ><button
              class="text-button"
              @click="copyText(inquiryText, 'inquiry')"
            >
              <Icon
                :name="copied === 'inquiry' ? 'check' : 'clipboard'"
                size="sm"
              />{{
                copied === 'inquiry'
                  ? c.copied
                  : contactType === 'sms'
                    ? c.copySmsRequest
                    : c.copyInquiry
              }}
            </button>
          </div>
        </div>
      </div>
    </dialog>
  </div>
</template>

<script setup>
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import Icon from '../../src/components/icons/Icon.vue'
import HeroFlow from './HeroFlow.vue'
import ToolsMarquee from './ToolsMarquee.vue'
import SmsFlow from './SmsFlow.vue'
import wechatImage from '../../public/contact/wechat.png'
import { site, products, providers, copy } from './content.js'

const locale = ref(
  localStorage.getItem('keepcoding-locale') === 'en' ? 'en' : 'zh',
)
const isDark = ref(localStorage.getItem('keepcoding-theme') === 'dark')
const c = computed(() => copy[locale.value])
const tr = (zh, en) => (locale.value === 'zh' ? zh : en)
const languages = [
  { code: 'zh', label: '简体中文' },
  { code: 'en', label: 'English' },
]
const navLinks = computed(() => [
  { id: 'subscriptions', label: c.value.subscriptions },
  { id: 'api', label: c.value.api },
  { id: 'verification', label: c.value.sms },
  { id: 'faq', label: c.value.help },
])
const filters = computed(() => [
  { value: 'all', label: c.value.all, count: products.length },
  ...['OpenAI', 'Anthropic'].map((value) => ({
    value,
    label: value,
    count: products.filter((product) => product.provider === value).length,
  })),
])
const selectedProvider = ref('all')
const filteredProducts = computed(() =>
  selectedProvider.value === 'all'
    ? products
    : products.filter((product) => product.provider === selectedProvider.value),
)
const localeRoot = ref(null),
  localeOpen = ref(false),
  menuOpen = ref(false),
  scrolled = ref(false),
  activeSection = ref('')
const themeChanging = ref(false),
  codeTab = ref('cURL'),
  codeTabs = ['cURL', 'Python'],
  openFaq = ref(0)
const contactDialog = ref(null),
  contactOpen = ref(false),
  contactProduct = ref(null),
  contactType = ref('general')
const copied = ref(''),
  toastMessage = ref(''),
  year = new Date().getFullYear()
const themeLabel = computed(() =>
  tr(
    isDark.value ? '切换浅色模式' : '切换深色模式',
    isDark.value ? 'Switch to light mode' : 'Switch to dark mode',
  ),
)
let revealObserver,
  revealFrame,
  sectionObserver,
  scrollFrame,
  toastTimer,
  copyTimer
const delay = (index) => ({ '--reveal-delay': `${index * 130}ms` })

const codeSample = computed(() =>
  codeTab.value === 'cURL'
    ? [
        `curl ${site.origin}/v1/chat/completions \\`,
        '  -H "Authorization: Bearer YOUR_API_KEY" \\',
        '  -H "Content-Type: application/json" \\',
        "  -d '{",
        '    "model": "YOUR_MODEL",',
        '    "messages": [',
        '      {"role": "user", "content": "Hello!"}',
        '    ]',
        "  }'",
      ].join('\n')
    : `from openai import OpenAI\n\nclient = OpenAI(\n    api_key="YOUR_API_KEY",\n    base_url="${site.origin}/v1",\n)\n\nresponse = client.chat.completions.create(\n    model="YOUR_MODEL",\n    messages=[{"role": "user", "content": "Hello!"}],\n)\nprint(response.choices[0].message.content)`,
)

const inquiryText = computed(() => {
  if (contactProduct.value)
    return tr(
      `你好，我想咨询 ${contactProduct.value.name}，页面价格为 ¥${contactProduct.value.price}/月，请确认开通方式和售后说明。`,
      `Hi, I am interested in ${contactProduct.value.name} at CNY ${contactProduct.value.price}/month. Please confirm activation details and coverage.`,
    )
  if (contactType.value === 'sms')
    return tr(
      '你好，我想购买 Claude / ChatGPT 的短信验证码，用于完成手机号验证。请确认支持的平台、价格、可用号码和下单方式。',
      'Hi, I would like to purchase an SMS code for Claude / ChatGPT to complete phone verification. Please confirm supported platforms, pricing, available numbers, and how to order.',
    )
  return c.value.contactGeneric
})

function setLocale(value) {
  locale.value = value
  localeOpen.value = false
  document.documentElement.lang = value === 'zh' ? 'zh-CN' : 'en'
  localStorage.setItem('keepcoding-locale', value)
}

function navigateTabs(event) {
  if (!['ArrowLeft', 'ArrowRight', 'Home', 'End'].includes(event.key)) return
  const tabs = [...event.currentTarget.querySelectorAll('[role="tab"]')]
  const current = tabs.indexOf(document.activeElement)
  if (current < 0) return
  event.preventDefault()
  const next =
    event.key === 'Home'
      ? 0
      : event.key === 'End'
        ? tabs.length - 1
        : (current + (event.key === 'ArrowRight' ? 1 : -1) + tabs.length) %
          tabs.length
  tabs[next].focus()
  tabs[next].click()
}

function applyTheme(value) {
  isDark.value = value
  document.documentElement.dataset.theme = value ? 'dark' : 'light'
  document
    .querySelector('meta[name="theme-color"]')
    ?.setAttribute('content', value ? '#071510' : '#f6faf8')
  localStorage.setItem('keepcoding-theme', value ? 'dark' : 'light')
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
  const rect = event.currentTarget.getBoundingClientRect(),
    x = rect.left + rect.width / 2,
    y = rect.top + rect.height / 2
  document.documentElement.style.setProperty('--theme-origin', `${x}px ${y}px`)
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

function openContact(type, product = null) {
  contactType.value = type
  contactProduct.value = product
  contactOpen.value = true
  nextTick(() => contactDialog.value.showModal())
}

function closeOnBackdrop(event) {
  if (event.target !== contactDialog.value) return
  const bounds = contactDialog.value.getBoundingClientRect()
  if (
    event.clientX < bounds.left ||
    event.clientX > bounds.right ||
    event.clientY < bounds.top ||
    event.clientY > bounds.bottom
  )
    contactDialog.value.close()
}

async function copyText(value, key) {
  try {
    await navigator.clipboard.writeText(value)
    copied.value = key
    toastMessage.value = c.value.copied
  } catch {
    toastMessage.value = c.value.copyFailed
  }
  clearTimeout(toastTimer)
  clearTimeout(copyTimer)
  toastTimer = setTimeout(() => {
    toastMessage.value = ''
  }, 2200)
  copyTimer = setTimeout(() => {
    copied.value = ''
  }, 2200)
}

function onScroll() {
  if (scrollFrame) return
  scrollFrame = requestAnimationFrame(() => {
    scrolled.value = scrollY > 12
    scrollFrame = null
  })
}

function onDocumentClick(event) {
  if (!localeRoot.value?.contains(event.target)) localeOpen.value = false
}
function onKeydown(event) {
  if (event.key === 'Escape') {
    localeOpen.value = false
    menuOpen.value = false
  }
}

function observeReveals() {
  revealObserver?.disconnect()
  const elements = document.querySelectorAll('[data-reveal]:not(.is-visible)')
  if (
    matchMedia('(prefers-reduced-motion: reduce)').matches ||
    !('IntersectionObserver' in window)
  ) {
    elements.forEach((element) => element.classList.add('is-visible'))
  } else {
    revealObserver = new IntersectionObserver(
      (entries) =>
        entries.forEach((entry) => {
          if (entry.isIntersecting) {
            entry.target.classList.add('is-visible')
            revealObserver.unobserve(entry.target)
          }
        }),
      { threshold: 0.08, rootMargin: '0px 0px -30px 0px' },
    )
    elements.forEach((element) => revealObserver.observe(element))
  }
}

function scheduleReveals() {
  cancelAnimationFrame(revealFrame)
  // 先绘制初始隐藏状态，再观察可见内容，保证首屏也能触发渐入过渡。
  revealFrame = requestAnimationFrame(() => {
    revealFrame = requestAnimationFrame(observeReveals)
  })
}

watch(selectedProvider, async () => {
  await nextTick()
  scheduleReveals()
})

onMounted(() => {
  applyTheme(isDark.value)
  setLocale(locale.value)
  document.documentElement.classList.add('reveal-ready')
  scheduleReveals()
  sectionObserver = new IntersectionObserver(
    (entries) =>
      entries.forEach((entry) => {
        if (entry.isIntersecting) activeSection.value = entry.target.id
      }),
    { rootMargin: '-15% 0px -60% 0px' },
  )
  document
    .querySelectorAll('main section[id]')
    .forEach((section) => sectionObserver.observe(section))
  window.addEventListener('scroll', onScroll, { passive: true })
  onScroll()
  document.addEventListener('click', onDocumentClick)
  document.addEventListener('keydown', onKeydown)
})

onUnmounted(() => {
  revealObserver?.disconnect()
  sectionObserver?.disconnect()
  cancelAnimationFrame(scrollFrame)
  cancelAnimationFrame(revealFrame)
  clearTimeout(toastTimer)
  clearTimeout(copyTimer)
  window.removeEventListener('scroll', onScroll)
  document.removeEventListener('click', onDocumentClick)
  document.removeEventListener('keydown', onKeydown)
  document.documentElement.classList.remove('reveal-ready')
})
</script>
