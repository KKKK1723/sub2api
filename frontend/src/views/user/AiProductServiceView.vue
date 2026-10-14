<template>
  <AppLayout>
    <div class="ai-center relative mx-auto max-w-6xl pb-12">
      <!-- 标题区 -->
      <span
        class="inline-flex items-center gap-2 rounded-full border border-gray-200 bg-white/70 px-3 py-1 text-xs text-gray-600 dark:border-dark-700 dark:bg-dark-900/70 dark:text-dark-300"
      >
        <span class="h-1.5 w-1.5 rounded-full bg-primary-500 ring-4 ring-primary-500/15"></span>
        {{ t('aiProducts.eyebrow') }}
      </span>
      <div class="mb-10 mt-5 flex flex-col gap-6 md:flex-row md:items-end md:justify-between">
        <div>
          <h1 class="ai-center-title text-4xl font-semibold tracking-tight md:text-5xl">
            {{ t('aiProducts.title') }}
          </h1>
          <p class="mt-4 max-w-md text-base text-gray-500 dark:text-dark-400">
            {{ t('aiProducts.description') }}
          </p>
        </div>
        <button
          type="button"
          class="inline-flex items-center gap-2.5 self-start whitespace-nowrap rounded-xl bg-gray-900 px-5 py-2.5 text-sm font-semibold text-white shadow-glow transition-transform hover:-translate-y-0.5 dark:bg-white dark:text-dark-950 md:self-auto"
          @click="openContact(null)"
        >
          {{ t('aiProducts.contact') }}
          <span class="rounded-md bg-white/15 px-1.5 py-0.5 text-[11px] font-medium dark:bg-dark-950/10">{{ t('aiProducts.wechat') }}</span>
        </button>
      </div>

      <!-- 分类：官方订阅 / 接码服务（各自独立路由） -->
      <div class="mb-7 grid grid-cols-1 gap-3 md:grid-cols-2">
        <router-link
          v-for="category in categories"
          :key="category.key"
          :to="{ name: category.routeName }"
          class="category-tab group flex items-center gap-4 rounded-2xl border bg-white px-5 py-4 transition-colors dark:bg-dark-900"
          :class="[
            category.key,
            activeCategory === category.key
              ? 'active border-transparent'
              : 'border-gray-200 hover:border-gray-300 dark:border-dark-700 dark:hover:border-dark-600'
          ]"
        >
          <span
            class="flex h-11 w-11 flex-shrink-0 items-center justify-center rounded-xl border transition-colors"
            :class="activeCategory === category.key ? category.iconActiveClass : 'border-gray-200 bg-gray-50 text-gray-400 dark:border-dark-700 dark:bg-dark-800 dark:text-dark-400'"
          >
            <Icon :name="category.icon" size="md" />
          </span>
          <span class="min-w-0">
            <span class="flex items-center gap-2 text-base font-semibold text-gray-900 dark:text-white">
              {{ t(`aiProducts.categories.${category.key}.title`) }}
              <span
                v-if="category.key === 'sms'"
                class="rounded-full border border-amber-500/40 px-1.5 text-[11px] font-semibold text-amber-600 dark:text-amber-400"
              >NEW</span>
            </span>
            <span class="block truncate text-sm text-gray-500 dark:text-dark-400">
              {{ t(`aiProducts.categories.${category.key}.description`, { price: minPriceText(category.key) }) }}
            </span>
          </span>
          <span class="ml-auto font-mono text-xs text-gray-400 dark:text-dark-500">{{ padCount(countOf(category.key)) }}</span>
        </router-link>
      </div>

      <!-- 厂商筛选 -->
      <div class="mb-4 flex items-center justify-between gap-3">
        <div class="inline-flex rounded-xl border border-gray-200 bg-white p-1 dark:border-dark-700 dark:bg-dark-900">
          <button
            v-for="vendor in vendorFilters"
            :key="vendor"
            type="button"
            class="rounded-lg px-3.5 py-1.5 text-sm font-medium transition-colors"
            :class="selectedVendor === vendor
              ? 'bg-gray-100 text-gray-900 shadow-sm ring-1 ring-gray-200 dark:bg-dark-800 dark:text-white dark:ring-dark-700'
              : 'text-gray-500 hover:text-gray-800 dark:text-dark-400 dark:hover:text-dark-200'"
            @click="selectedVendor = vendor"
          >
            {{ vendor === 'all' ? t('aiProducts.all') : vendor }}
          </button>
        </div>
        <span class="font-mono text-xs text-gray-400 dark:text-dark-500">{{ t('aiProducts.itemCount', { count: padCount(visibleProducts.length) }) }}</span>
      </div>

      <!-- 商品卡片 -->
      <div class="grid grid-cols-1 gap-3.5 sm:grid-cols-2 lg:grid-cols-4">
        <article
          v-for="(product, index) in visibleProducts"
          :key="product.id"
          class="product-card group relative flex flex-col overflow-hidden rounded-2xl border bg-white p-5 transition-transform duration-200 hover:-translate-y-0.5 dark:bg-dark-900"
          :class="[product.tone, product.featured ? 'featured border-transparent' : 'border-gray-200 hover:border-gray-300 dark:border-dark-700 dark:hover:border-dark-600']"
          :style="{ animationDelay: `${index * 60}ms` }"
          @pointermove="trackPointer"
        >
          <div class="relative flex items-center justify-between">
            <span class="flex h-10 w-10 items-center justify-center rounded-xl border border-gray-200 bg-gray-50 dark:border-dark-700 dark:bg-dark-800">
              <img v-if="product.logo" :src="product.logo" alt="" class="h-5 w-5" :class="product.logoClass" />
              <Icon v-else name="chatBubble" size="md" class="text-amber-600 dark:text-amber-400" />
            </span>
            <span
              v-if="product.isNew"
              class="product-chip-new rounded-full border px-2 py-0.5 text-[11px] font-medium"
            >{{ t('aiProducts.new') }}</span>
            <span
              v-else
              class="rounded-full border border-gray-200 px-2 py-0.5 text-[11px] text-gray-500 dark:border-dark-700 dark:text-dark-400"
            >{{ t(`aiProducts.chip.${product.category}`) }}</span>
          </div>

          <h3 class="relative mt-4 text-lg font-semibold tracking-tight text-gray-900 dark:text-white">{{ productName(product) }}</h3>
          <p class="relative text-xs text-gray-500 dark:text-dark-400">{{ t(`aiProducts.products.${product.id}.subtitle`) }}</p>

          <ul class="relative my-5 space-y-2">
            <li
              v-for="feature in featureKeys"
              :key="feature"
              class="flex items-center gap-2 text-sm text-gray-600 dark:text-dark-300"
            >
              <Icon name="check" size="sm" class="product-check flex-shrink-0" :stroke-width="2.4" />
              {{ t(`aiProducts.products.${product.id}.${feature}`) }}
            </li>
          </ul>

          <div class="relative mt-auto border-t border-gray-100 pt-4 dark:border-dark-800">
            <div class="mb-3.5 flex items-baseline gap-1.5">
              <span class="text-base text-gray-400 dark:text-dark-500">¥</span>
              <span class="font-mono text-3xl font-semibold tracking-tight text-gray-900 dark:text-white">{{ product.price }}</span>
              <span class="text-sm text-gray-400 dark:text-dark-500">{{ t(`aiProducts.units.${product.unit}`) }}</span>
            </div>
            <button
              type="button"
              class="product-buy flex w-full items-center justify-center gap-2 rounded-xl border py-2.5 text-sm font-semibold transition-colors"
              @click="openContact(product)"
            >
              {{ t('aiProducts.consult') }}
              <Icon name="arrowRight" size="sm" />
            </button>
          </div>
        </article>
      </div>

      <div
        class="mt-6 flex gap-3 rounded-xl border border-dashed border-gray-300 px-4 py-3.5 text-xs leading-relaxed text-gray-500 dark:border-dark-700 dark:text-dark-400"
      >
        <Icon name="infoCircle" size="sm" class="mt-0.5 flex-shrink-0" />
        <span>{{ t(`aiProducts.terms.${activeCategory}`) }}</span>
      </div>
    </div>

    <!-- 联系站长 -->
    <BaseDialog
      :show="contactOpen"
      :title="contactProduct ? productName(contactProduct) : t('aiProducts.dialog.title')"
      width="narrow"
      close-on-click-outside
      @close="contactOpen = false"
    >
      <dl class="divide-y divide-dashed divide-gray-200 text-sm dark:divide-dark-700">
        <template v-if="contactProduct">
          <div class="flex items-center justify-between py-2.5">
            <dt class="text-gray-500 dark:text-dark-400">{{ t('aiProducts.dialog.product') }}</dt>
            <dd class="font-semibold text-gray-900 dark:text-white">{{ productName(contactProduct) }}</dd>
          </div>
          <div class="flex items-center justify-between py-2.5">
            <dt class="text-gray-500 dark:text-dark-400">{{ t('aiProducts.dialog.price') }}</dt>
            <dd class="font-mono font-semibold text-gray-900 dark:text-white">¥{{ contactProduct.price }} {{ t(`aiProducts.units.${contactProduct.unit}`) }}</dd>
          </div>
        </template>
        <div class="flex items-center justify-between py-2.5">
          <dt class="text-gray-500 dark:text-dark-400">{{ t('aiProducts.dialog.method') }}</dt>
          <dd class="font-semibold text-gray-900 dark:text-white">{{ t('aiProducts.dialog.methodValue') }}</dd>
        </div>
      </dl>

      <div class="mt-4 flex items-center gap-4 rounded-xl border border-gray-200 bg-gray-50 p-3 dark:border-dark-700 dark:bg-dark-800">
        <img :src="wechatImage" :alt="t('aiProducts.dialog.scan')" class="h-24 w-24 flex-shrink-0 rounded-lg bg-white p-1" />
        <div class="min-w-0 flex-1">
          <p class="text-xs text-gray-500 dark:text-dark-400">{{ t('aiProducts.dialog.scan') }}</p>
          <p class="mt-1 truncate font-mono text-sm font-semibold text-gray-900 dark:text-white">{{ site.wechat }}</p>
          <button
            type="button"
            class="mt-2 inline-flex items-center gap-1.5 rounded-lg bg-gray-900 px-3 py-1.5 text-xs font-semibold text-white dark:bg-white dark:text-dark-950"
            @click="copyToClipboard(site.wechat, t('aiProducts.dialog.wechatCopied'))"
          >
            <Icon name="copy" size="xs" />{{ t('aiProducts.dialog.copyWechat') }}
          </button>
        </div>
      </div>
      <button
        type="button"
        class="mt-2.5 flex w-full items-center justify-center gap-2 rounded-xl border border-gray-200 py-2.5 text-sm text-gray-700 transition-colors hover:border-gray-300 dark:border-dark-700 dark:text-dark-200 dark:hover:border-dark-600"
        @click="copyToClipboard(inquiryText, t('aiProducts.dialog.inquiryCopied'))"
      >
        <Icon name="clipboard" size="sm" />{{ t('aiProducts.dialog.copyInquiry') }}
      </button>
    </BaseDialog>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Icon from '@/components/icons/Icon.vue'
import { site } from '@/components/public/content'
import { useClipboard } from '@/composables/useClipboard'

type CategoryKey = 'subscription' | 'sms'
type UnitKey = 'month' | 'once' | 'days25'

interface Product {
  id: string
  category: CategoryKey
  vendor: string
  // 订阅商品名不翻译；接码商品名走 i18n
  name?: string
  price: number
  unit: UnitKey
  logo?: string
  logoClass?: string
  tone: 'openai' | 'claude' | 'sms'
  isNew?: boolean
  featured?: boolean
}

const { t } = useI18n()
const route = useRoute()
const { copyToClipboard } = useClipboard()

const wechatImage = '/contact/wechat.png'
const featureKeys = ['feature1', 'feature2', 'feature3'] as const

// 价格与开通方式由站长人工确认，页面仅做展示；顺序即展示顺序
const products: Product[] = [
  { id: 'gptPlus', category: 'subscription', vendor: 'OpenAI', name: 'GPT Plus', price: 128, unit: 'month', logo: '/brands/openai.svg', logoClass: 'dark:invert', tone: 'openai' },
  { id: 'gptPro5x', category: 'subscription', vendor: 'OpenAI', name: 'GPT Pro 5x', price: 650, unit: 'month', logo: '/brands/openai.svg', logoClass: 'dark:invert', tone: 'openai' },
  { id: 'gptPro20x', category: 'subscription', vendor: 'OpenAI', name: 'GPT Pro 20x', price: 1099, unit: 'month', logo: '/brands/openai.svg', logoClass: 'dark:invert', tone: 'openai', isNew: true, featured: true },
  { id: 'claudePro', category: 'subscription', vendor: 'Anthropic', name: 'Claude Pro', price: 178, unit: 'month', logo: '/brands/claude.svg', logoClass: 'claude-logo', tone: 'claude' },
  { id: 'codexUsSimOnce', category: 'sms', vendor: 'Codex', price: 5.2, unit: 'once', tone: 'sms', isNew: true },
  { id: 'codexUs25d', category: 'sms', vendor: 'Codex', price: 28.8, unit: 'days25', tone: 'sms', isNew: true },
]

// 官方订阅与接码各有独立路由；侧边栏按前缀匹配，两个子页面都高亮同一入口
const categories: Array<{ key: CategoryKey; routeName: string; icon: 'sparkles' | 'chatBubble'; iconActiveClass: string }> = [
  {
    key: 'subscription',
    routeName: 'AiProductService',
    icon: 'sparkles',
    iconActiveClass: 'border-primary-500/40 bg-primary-50 text-primary-600 dark:bg-primary-500/10 dark:text-primary-400',
  },
  {
    key: 'sms',
    routeName: 'AiProductSms',
    icon: 'chatBubble',
    iconActiveClass: 'border-amber-500/40 bg-amber-50 text-amber-600 dark:bg-amber-500/10 dark:text-amber-400',
  },
]

const activeCategory = computed<CategoryKey>(() => (route.name === 'AiProductSms' ? 'sms' : 'subscription'))
const selectedVendor = ref('all')
watch(activeCategory, () => {
  selectedVendor.value = 'all'
})

const categoryProducts = computed(() => products.filter((product) => product.category === activeCategory.value))
const vendorFilters = computed(() => ['all', ...new Set(categoryProducts.value.map((product) => product.vendor))])
const visibleProducts = computed(() =>
  selectedVendor.value === 'all'
    ? categoryProducts.value
    : categoryProducts.value.filter((product) => product.vendor === selectedVendor.value)
)

const countOf = (category: CategoryKey) => products.filter((product) => product.category === category).length
const minPriceText = (category: CategoryKey) =>
  `¥${Math.min(...products.filter((product) => product.category === category).map((product) => product.price))}`
const padCount = (count: number) => String(count).padStart(2, '0')
const productName = (product: Product) => product.name || t(`aiProducts.products.${product.id}.name`)

const contactOpen = ref(false)
const contactProduct = ref<Product | null>(null)
const inquiryText = computed(() => {
  const product = contactProduct.value
  if (!product) return t('aiProducts.dialog.generalInquiry')
  return t('aiProducts.dialog.inquiry', {
    name: productName(product),
    price: product.price,
    unit: t(`aiProducts.units.${product.unit}`),
  })
})

function openContact(product: Product | null) {
  contactProduct.value = product
  contactOpen.value = true
}

// 卡片聚光效果跟随指针位置
function trackPointer(event: PointerEvent) {
  const card = event.currentTarget as HTMLElement
  const rect = card.getBoundingClientRect()
  card.style.setProperty('--mx', `${event.clientX - rect.left}px`)
  card.style.setProperty('--my', `${event.clientY - rect.top}px`)
}
</script>

<style scoped>
.ai-center::before {
  content: '';
  position: absolute;
  inset: -2rem -2rem auto;
  z-index: -1;
  height: 420px;
  background-image:
    linear-gradient(rgba(15, 23, 42, 0.045) 1px, transparent 1px),
    linear-gradient(90deg, rgba(15, 23, 42, 0.045) 1px, transparent 1px);
  background-size: 56px 56px;
  -webkit-mask-image: radial-gradient(ellipse 70% 60% at 50% 0%, #000 30%, transparent 75%);
  mask-image: radial-gradient(ellipse 70% 60% at 50% 0%, #000 30%, transparent 75%);
  pointer-events: none;
}
:global(.dark) .ai-center::before {
  background-image:
    linear-gradient(rgba(255, 255, 255, 0.035) 1px, transparent 1px),
    linear-gradient(90deg, rgba(255, 255, 255, 0.035) 1px, transparent 1px);
}

.ai-center-title {
  background: linear-gradient(180deg, #0f172a 30%, rgba(15, 23, 42, 0.5));
  -webkit-background-clip: text;
  background-clip: text;
  color: transparent;
}
:global(.dark) .ai-center-title {
  background-image: linear-gradient(180deg, #f8fafc 30%, rgba(248, 250, 252, 0.45));
}

/* 选中分类用渐变描边 */
.category-tab.active {
  background:
    linear-gradient(#fff, #fff) padding-box,
    linear-gradient(120deg, #14b8a6, transparent 60%) border-box;
}
.category-tab.sms.active {
  background:
    linear-gradient(#fff, #fff) padding-box,
    linear-gradient(120deg, #f59e0b, transparent 60%) border-box;
}
:global(.dark) .category-tab.active {
  background:
    linear-gradient(#0f172a, #0f172a) padding-box,
    linear-gradient(120deg, #2dd4bf, transparent 60%) border-box;
}
:global(.dark) .category-tab.sms.active {
  background:
    linear-gradient(#0f172a, #0f172a) padding-box,
    linear-gradient(120deg, #fbbf24, transparent 60%) border-box;
}

/* 商品卡片：按品牌色调整聚光、勾选与按钮 */
.product-card {
  --mx: 50%;
  --my: 0%;
  --tone: #0d9488;
  animation: card-rise 0.45s cubic-bezier(0.2, 0.8, 0.2, 1) both;
}
.product-card.claude {
  --tone: #c2410c;
}
.product-card.sms {
  --tone: #b45309;
}
:global(.dark) .product-card {
  --tone: #2dd4bf;
}
:global(.dark) .product-card.claude {
  --tone: #f0875f;
}
:global(.dark) .product-card.sms {
  --tone: #fbbf24;
}
.product-card::before {
  content: '';
  position: absolute;
  inset: 0;
  background: radial-gradient(380px circle at var(--mx) var(--my), color-mix(in srgb, var(--tone) 12%, transparent), transparent 45%);
  opacity: 0;
  transition: opacity 0.3s;
  pointer-events: none;
}
.product-card:hover::before {
  opacity: 1;
}
.product-check {
  color: var(--tone);
}
.product-chip-new {
  border-color: color-mix(in srgb, var(--tone) 50%, transparent);
  color: var(--tone);
}
.product-buy {
  border-color: rgba(15, 23, 42, 0.14);
  background: #f8fafc;
  color: #0f172a;
}
.product-buy:hover {
  border-color: var(--tone);
  background: color-mix(in srgb, var(--tone) 10%, #f8fafc);
  color: var(--tone);
}
:global(.dark) .product-buy {
  border-color: rgba(255, 255, 255, 0.14);
  background: #1e293b;
  color: #f1f5f9;
}
:global(.dark) .product-buy:hover {
  background: color-mix(in srgb, var(--tone) 14%, #1e293b);
  color: var(--tone);
}

/* 旗舰商品：旋转流光描边 */
@property --spin {
  syntax: '<angle>';
  inherits: false;
  initial-value: 0deg;
}
.product-card.featured {
  background:
    linear-gradient(#fff, #fff) padding-box,
    conic-gradient(from var(--spin), #14b8a6, transparent 25%, transparent 50%, #818cf8, transparent 75%, #14b8a6) border-box;
  animation:
    card-rise 0.45s cubic-bezier(0.2, 0.8, 0.2, 1) both,
    border-spin 6s linear infinite;
}
:global(.dark) .product-card.featured {
  background:
    linear-gradient(#0f172a, #0f172a) padding-box,
    conic-gradient(from var(--spin), #2dd4bf, transparent 25%, transparent 50%, #818cf8, transparent 75%, #2dd4bf) border-box;
}
.product-card.featured .product-buy {
  border-color: transparent;
  background: #0f172a;
  color: #fff;
}
.product-card.featured .product-buy:hover {
  background: #0d9488;
  color: #fff;
}
:global(.dark) .product-card.featured .product-buy {
  background: #f8fafc;
  color: #020617;
}
:global(.dark) .product-card.featured .product-buy:hover {
  background: #2dd4bf;
  color: #042f2e;
}

.claude-logo {
  filter: invert(47%) sepia(53%) saturate(519%) hue-rotate(333deg) brightness(92%) contrast(83%);
}

@keyframes card-rise {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
}
@keyframes border-spin {
  to {
    --spin: 360deg;
  }
}
@media (prefers-reduced-motion: reduce) {
  .product-card,
  .product-card.featured {
    animation: none;
  }
}
</style>
