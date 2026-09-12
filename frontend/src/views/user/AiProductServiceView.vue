<template>
  <AppLayout>
    <div class="ai-products-page">
      <div class="content-head">
        <div>
          <h1>AI 商品服务</h1>
          <p class="intro">精选主流 AI 订阅服务，透明展示套餐与价格。</p>
        </div>
        <div class="contact-chip"><span class="contact-dot"></span>需要购买？联系站长</div>
      </div>

      <section class="hero" aria-label="AI 服务宣传">
        <div class="hero-copy">
          <p class="hero-kicker">AI SUBSCRIPTION SERVICE</p>
          <h2>把更多时间，留给真正重要的创造。</h2>
          <p>GPT、Claude 等专业 AI 服务持续上新。页面仅作商品介绍，如有需要，请联系站长购买。</p>
        </div>
        <div class="hero-orbit" aria-hidden="true"></div>
      </section>

      <div class="category-tabs" role="tablist" aria-label="AI 服务商分类">
        <button
          v-for="category in categories"
          :key="category.key"
          class="category-tab"
          :class="{ active: selectedCategory === category.key }"
          type="button"
          role="tab"
          :aria-selected="selectedCategory === category.key"
          @click="selectedCategory = category.key"
        >
          <span class="category-label">
            <span class="provider-dot" :class="category.dotClass"></span>
            {{ category.label }}
          </span>
        </button>
      </div>

      <section class="product-grid" aria-live="polite">
        <article v-for="product in selectedProducts" :key="product.name" class="product-card">
          <div class="product-top">
            <div class="product-logo" :class="product.logoClass">{{ product.logo }}</div>
            <span class="tag">服务介绍</span>
          </div>
          <h2>{{ product.name }}</h2>
          <p class="product-desc">{{ product.description }}</p>
          <div class="meta">
            <span v-for="tag in product.tags" :key="tag">{{ tag }}</span>
          </div>
          <div class="product-bottom">
            <div class="price">{{ product.price }} <small>/ 月</small></div>
            <div class="contact-note">如有需要<br>请联系站长购买</div>
          </div>
        </article>

        <div v-if="selectedProducts.length === 0" class="empty-state">
          <strong>{{ selectedCategoryLabel }} 商品即将上架</strong>
          <span>我们正在准备稳定可靠的订阅服务，敬请关注。</span>
        </div>
      </section>

      <div class="notice">
        <span>ⓘ</span>
        <span>商品页面仅用于服务介绍，不提供在线下单。商品含质保订阅：如账号订阅掉订，按剩余天数补差价；账号封号风险不在质保范围内。如有需要，请联系站长购买。</span>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import AppLayout from '@/components/layout/AppLayout.vue'

type CategoryKey = 'openai' | 'anthropic' | 'grok'

interface Product {
  name: string
  description: string
  price: string
  tags: string[]
  logo: string
  logoClass?: string
}

const selectedCategory = ref<CategoryKey>('openai')

const categories: Array<{ key: CategoryKey; label: string; dotClass?: string }> = [
  { key: 'openai', label: 'OpenAI' },
  { key: 'anthropic', label: 'Anthropic', dotClass: 'anthropic' },
  { key: 'grok', label: 'Grok', dotClass: 'grok' },
]

const products: Record<CategoryKey, Product[]> = {
  openai: [
    {
      name: 'GPT Plus',
      description: '适用于 ChatGPT Plus，体验更强模型、语音与图像能力。含质保订阅，掉订按剩余天数补差价，不质保封号。',
      price: '¥ 128',
      tags: ['GPT Plus', '月度订阅', '含质保'],
      logo: 'GPT',
    },
    {
      name: 'GPT Pro 5x',
      description: '适用于高频专业用户，提供更高额度与更稳定的使用体验。含质保订阅，掉订按剩余天数补差价，不质保封号。',
      price: '¥ 688',
      tags: ['Pro 5x', '高额度', '含质保'],
      logo: '5x',
      logoClass: 'pro',
    },
  ],
  anthropic: [
    {
      name: 'Claude Pro',
      description: '适用于 Claude Pro，支持更长上下文与更高使用额度。含质保订阅，掉订按剩余天数补差价，不质保封号。',
      price: '¥ 138',
      tags: ['Claude Pro', '月度订阅', '含质保'],
      logo: 'C',
      logoClass: 'claude',
    },
  ],
  grok: [],
}

const selectedProducts = computed(() => products[selectedCategory.value])
const selectedCategoryLabel = computed(() => categories.find((category) => category.key === selectedCategory.value)?.label || '')
</script>

<style scoped>
.ai-products-page {
  --ink: #172033;
  --muted: #667085;
  --line: #e5e9f0;
  --surface: #ffffff;
  --canvas: #f5f7fb;
  --navy: #172a46;
  --cyan: #1ca6a8;
  --cyan-soft: #e5f7f6;
  --lime: #c7e86d;
  --amber: #f0b35b;
  --shadow: 0 18px 48px rgba(23, 42, 70, .08);
  min-height: calc(100vh - 96px);
  margin: -32px;
  padding: clamp(28px, 5vw, 56px) clamp(20px, 5vw, 64px) 70px;
  background: var(--canvas);
  color: var(--ink);
  font: 15px/1.5 Inter, "PingFang SC", "Microsoft YaHei", sans-serif;
}

.content-head { display: flex; align-items: flex-end; justify-content: space-between; gap: 24px; margin-bottom: 22px; }
h1 { margin: 0 0 7px; font-size: clamp(26px, 3vw, 36px); line-height: 1.15; letter-spacing: -.025em; }
.intro { max-width: 580px; margin: 0; color: var(--muted); }
.contact-chip { display: inline-flex; align-items: center; gap: 9px; padding: 10px 14px; border: 1px solid #cfe6e3; border-radius: 9px; background: var(--cyan-soft); color: #087b7e; font-size: 13px; font-weight: 700; }
.contact-dot { width: 8px; height: 8px; border-radius: 50%; background: var(--cyan); box-shadow: 0 0 0 4px rgba(28,166,168,.12); }

.hero { position: relative; overflow: hidden; min-height: 208px; margin-bottom: 30px; padding: 28px 32px; border-radius: 16px; background: linear-gradient(115deg, #142842 0%, #1b4760 52%, #168f92 100%); color: #fff; box-shadow: 0 20px 42px rgba(18, 58, 81, .2); }
.hero::before, .hero::after { content: ""; position: absolute; border: 1px solid rgba(255,255,255,.15); border-radius: 50%; pointer-events: none; }
.hero::before { width: 300px; height: 300px; right: -90px; top: -170px; }
.hero::after { width: 380px; height: 380px; right: -120px; bottom: -290px; }
.hero-copy { position: relative; z-index: 1; max-width: 530px; }
.hero-kicker { margin: 0 0 12px; color: var(--lime); font-size: 11px; font-weight: 800; letter-spacing: .15em; text-transform: uppercase; }
.hero h2 { margin: 0 0 9px; font-size: clamp(23px, 3vw, 32px); line-height: 1.18; letter-spacing: -.02em; }
.hero p { max-width: 490px; margin: 0; color: rgba(255,255,255,.74); font-size: 14px; }
.hero-orbit { position: absolute; z-index: 0; right: 13%; top: 50%; width: 128px; height: 128px; border: 1px solid rgba(199,232,109,.45); border-radius: 50%; transform: translateY(-50%); }
.hero-orbit::before, .hero-orbit::after { content: ""; position: absolute; border-radius: 50%; background: var(--lime); box-shadow: 0 0 24px rgba(199,232,109,.65); }
.hero-orbit::before { width: 10px; height: 10px; left: 8px; top: 19px; }
.hero-orbit::after { width: 7px; height: 7px; right: 13px; bottom: 22px; opacity: .7; }

.category-tabs { display: flex; gap: 8px; margin-bottom: 22px; border-bottom: 1px solid var(--line); }
.category-tab { position: relative; padding: 0 15px 13px; border: 0; background: transparent; color: var(--muted); cursor: pointer; }
.category-tab::after { content: ""; position: absolute; right: 14px; bottom: -1px; left: 14px; height: 2px; background: transparent; }
.category-tab:hover, .category-tab.active { color: var(--navy); font-weight: 700; }
.category-tab.active::after { background: var(--cyan); }
.category-label { display: inline-flex; align-items: center; gap: 8px; }
.provider-dot { width: 8px; height: 8px; border-radius: 50%; background: var(--cyan); }
.provider-dot.anthropic { background: #d39c75; }
.provider-dot.grok { background: #6b7280; }

.product-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(255px, 1fr)); gap: 18px; }
.product-card { display: flex; flex-direction: column; min-height: 300px; padding: 22px; border: 1px solid var(--line); border-radius: 12px; background: var(--surface); box-shadow: var(--shadow); }
.product-card:hover { border-color: #b7dedd; transform: translateY(-2px); transition: transform .2s ease, border-color .2s ease; }
.product-top { display: flex; align-items: flex-start; justify-content: space-between; gap: 12px; }
.product-logo { width: 42px; height: 42px; display: grid; place-items: center; border-radius: 11px; background: var(--cyan-soft); color: #087b7e; font-weight: 800; }
.product-logo.pro { background: #edf0ff; color: #5966b3; }
.product-logo.claude { background: #fff0e6; color: #b96d38; }
.tag { padding: 4px 8px; border-radius: 5px; background: #f0f9d8; color: #55730d; font-size: 11px; font-weight: 700; }
.product-card h2 { margin: 21px 0 8px; font-size: 19px; line-height: 1.25; }
.product-desc { min-height: 66px; margin: 0; color: var(--muted); font-size: 13px; }
.meta { display: flex; gap: 7px; flex-wrap: wrap; margin: 17px 0 23px; }
.meta span { padding: 5px 8px; border-radius: 5px; background: #f6f8fb; color: #647084; font-size: 11px; }
.product-bottom { display: flex; align-items: flex-end; justify-content: space-between; gap: 14px; margin-top: auto; }
.price { color: var(--navy); font-size: 22px; font-weight: 800; letter-spacing: -.02em; }
.price small { color: var(--muted); font-size: 11px; font-weight: 400; }
.contact-note { color: var(--cyan); font-size: 12px; font-weight: 700; text-align: right; }
.empty-state { grid-column: 1 / -1; padding: 58px 20px; border: 1px dashed #ccd4df; border-radius: 12px; background: rgba(255,255,255,.6); text-align: center; color: var(--muted); }
.empty-state strong { display: block; margin-bottom: 6px; color: var(--ink); font-size: 17px; }
.notice { display: flex; align-items: flex-start; gap: 10px; max-width: 900px; margin-top: 28px; padding: 14px 16px; border-left: 3px solid var(--amber); background: #fffaf1; color: #76562d; font-size: 13px; }

@media (max-width: 760px) {
  .ai-products-page { margin: -16px; padding: 30px 18px 48px; }
  .content-head { align-items: flex-start; flex-direction: column; margin-bottom: 23px; }
  .contact-chip { align-self: stretch; justify-content: center; }
  .hero { min-height: 220px; padding: 24px; }
  .hero-orbit { right: -24px; top: auto; bottom: -43px; transform: none; }
  .category-tabs { overflow-x: auto; }
  .category-tab { flex: 0 0 auto; }
}

@media (min-width: 768px) and (max-width: 1023px) {
  .ai-products-page { margin: -24px; }
}
</style>
