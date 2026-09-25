export const site = {
  name: 'KeepCoding API',
  origin: 'https://taokc.xyz',
  wechat: 'Taokkkkkkkboy',
}

// 套餐及价格沿用现有商品页，开通仍由站长确认。
export const products = [
  {
    id: 'gpt-plus',
    provider: 'OpenAI',
    name: 'GPT Plus',
    price: 128,
    logo: '/brands/openai.svg',
    tone: 'mint',
    tag: { zh: '日常创作', en: 'Everyday work' },
    description: {
      zh: '写作、学习、图像与日常灵感。',
      en: 'Writing, learning, images, and everyday ideas.',
    },
    features: {
      zh: ['ChatGPT Plus 月度订阅', '语音与图像能力', '含掉订质保'],
      en: [
        'Monthly ChatGPT Plus subscription',
        'Voice and image capabilities',
        'Subscription lapse coverage',
      ],
    },
  },
  {
    id: 'claude-pro',
    provider: 'Anthropic',
    name: 'Claude Pro',
    price: 138,
    logo: '/brands/claude.svg',
    tone: 'apricot',
    tag: { zh: '代码与长文', en: 'Code & long context' },
    description: {
      zh: '从一段代码，到一份完整的方案。',
      en: 'From a piece of code to a complete proposal.',
    },
    features: {
      zh: ['Claude Pro 月度订阅', '长上下文与更高使用额度', '含掉订质保'],
      en: [
        'Monthly Claude Pro subscription',
        'Long context and higher usage limits',
        'Subscription lapse coverage',
      ],
    },
  },
  {
    id: 'gpt-pro',
    provider: 'OpenAI',
    name: 'GPT Pro 5x',
    price: 688,
    logo: '/brands/openai.svg',
    tone: 'blue',
    tag: { zh: '高频使用', en: 'Frequent use' },
    description: {
      zh: '为更密集的专业 AI 工作准备。',
      en: 'For more demanding professional AI work.',
    },
    features: {
      zh: ['GPT Pro 5x 月度订阅', '更高使用额度', '含掉订质保'],
      en: [
        'Monthly GPT Pro 5x subscription',
        'Higher usage allowance',
        'Subscription lapse coverage',
      ],
    },
  },
]

export const providers = [
  { name: 'OpenAI', logo: '/brands/openai.svg' },
  { name: 'Claude', logo: '/brands/claude.svg', color: true },
  { name: 'Gemini', logo: '/brands/gemini.svg', color: true },
  { name: 'Grok', logo: '/brands/grok.svg' },
]

export const copy = {
  zh: {
    subscriptions: 'AI 订阅',
    api: '模型 API',
    sms: '接码服务',
    help: '常见问题',
    console: '控制台',
    eyebrow: '让 AI 的每一种用法，都更简单',
    headline: '模型 API，AI 订阅，',
    headlineAccent: '还有接码服务。',
    heroDescription:
      '开发接入、日常创作、手机验证。按你的需求选择，剩下的交给我们。',
    browse: '挑选 AI 订阅',
    getKey: '获取 API Key',
    serviceApi: '接入你的应用与工具',
    serviceSub: '开通你常用的 AI',
    serviceSms: '购买短信验证码，完成手机验证',
    ecosystem: '连接你熟悉的 AI',
    catalogLabel: '01 / AI 订阅',
    catalogTitle: 'AI 订阅与套餐',
    catalogDescription: '常用的 AI 订阅，套餐与价格一目了然。',
    all: '全部',
    monthly: '月度订阅',
    perMonth: '/ 月',
    consult: '咨询开通',
    details: '套餐说明',
    otherProducts: '有其他订阅需求？',
    askUs: '和站长聊聊',
    productNotice:
      '订阅开通前请与站长确认。掉订按剩余天数补差价，账号封号不在质保范围内。',
    apiLabel: '02 / 模型 API',
    apiTitle: '你专注构建，\n模型接入交给我们。',
    apiDescription:
      '把模型接入你的应用、编辑器和自动化工作流。在一个控制台管理密钥，查看用量与调用记录。',
    apiFeatures: ['独立 API 密钥', '用量与账单可查', '按需选择可用渠道'],
    channels: '查看可用渠道',
    docs: '接入文档',
    compatible: '常用工具接入',
    example: '接入示例',
    copyCode: '复制代码',
    copyEndpoint: '复制接口地址',
    smsLabel: '03 / 接码服务',
    smsTitle: '需要手机号验证？\n购买短信验证码。',
    smsDescription:
      'Claude、ChatGPT 等平台要求手机号验证，却没有可用的海外手机号？按平台购买短信验证码，收到后填入验证页面，完成手机验证。',
    steps: [
      {
        title: '选择平台',
        description: '选择需要手机验证的 Claude、ChatGPT 等平台',
      },
      {
        title: '下单购买',
        description: '确认价格与可用号码，购买对应平台的短信验证码',
      },
      {
        title: '接收验证码',
        description: '将收到的短信验证码填入平台，完成手机验证',
      },
    ],
    smsFlow: {
      title: '手机验证流程',
      number: '获取接码号码',
      numberNote: '获取可用于目标平台验证的手机号',
      code: '接收短信验证码',
      complete: '返回平台完成验证',
      completeNote: '填入收到的验证码，继续使用你的 AI 账号',
    },
    smsContactSubtitle: '添加站长微信，确认价格后购买对应平台的短信验证码。',
    copySmsRequest: '复制购买需求',
    faqLabel: '了解更多',
    faqTitle: '开始之前，你可能还想知道。',
    faqs: [
      {
        q: 'API 和 AI 订阅有什么区别？',
        a: 'API 用于在应用、编辑器或自动化工具中调用模型；AI 订阅用于使用对应平台的会员功能。两类服务分别开通，订阅不等于 API 额度。',
      },
      {
        q: '如何购买和开通订阅？',
        a: '选择套餐后点击“咨询开通”，通过微信联系站长，确认套餐、账号条件、价格和开通方式后办理。当前商品页面用于服务展示。',
      },
      {
        q: '订阅质保包含哪些内容？',
        a: '现有商品包含掉订质保：如账号订阅掉订，按剩余天数补差价。账号封号不在质保范围内，具体办理前请确认商品说明。',
      },
      {
        q: '接码服务是什么，怎么购买？',
        a: '当 Claude、ChatGPT 等平台要求手机号验证，而你没有可用的海外手机号时，可以购买短信验证码。通过微信向站长说明需要验证的平台，确认价格并下单，收到验证码后填入对应平台完成手机验证。具体支持情况以当前可用号码为准。',
      },
      {
        q: '在哪里查看可用模型和调用记录？',
        a: '登录控制台后，可以查看可用渠道、创建 API 密钥，并查询用量与调用记录。可调用的模型和额度以你的账号配置为准。',
      },
    ],
    footerDescription: '模型 API、AI 订阅与接码服务。\n让想做的事，更快开始。',
    services: '服务',
    support: '联系与支持',
    contact: '联系站长',
    contactTitle: '联系站长',
    contactSubtitle: '微信沟通，确认后办理。',
    scan: '微信扫码联系',
    copyWechat: '复制微信号',
    copyInquiry: '复制咨询内容',
    close: '关闭',
    copied: '已复制',
    copyFailed: '复制失败，请选中文字手动复制',
    contactGeneric: '你好，我想了解 AI 服务。',
  },
  en: {
    subscriptions: 'Subscriptions',
    api: 'Model API',
    sms: 'SMS codes',
    help: 'FAQ',
    console: 'Console',
    eyebrow: 'More ways to work with AI. Less friction.',
    headline: 'Model APIs. AI subscriptions.',
    headlineAccent: 'SMS codes, too.',
    heroDescription:
      'Build applications, create every day, and get SMS codes for phone verification. Start with what you need.',
    browse: 'Explore subscriptions',
    getKey: 'Get an API key',
    serviceApi: 'Connect your apps and tools',
    serviceSub: 'Access the AI you use',
    serviceSms: 'Buy codes for phone verification',
    ecosystem: 'AI you already know',
    catalogLabel: '01 / SUBSCRIPTIONS',
    catalogTitle: 'AI subscriptions & plans',
    catalogDescription:
      'Popular AI subscriptions, with clear plans and prices.',
    all: 'All',
    monthly: 'Monthly subscription',
    perMonth: '/ month',
    consult: 'Ask about this plan',
    details: 'Plan details',
    otherProducts: 'Looking for another subscription?',
    askUs: 'Talk to us',
    productNotice:
      'Confirm details before purchase. Subscription lapses are compensated for remaining days; account bans are not covered.',
    apiLabel: '02 / MODEL API',
    apiTitle: 'Focus on building.\nConnect to the models you need.',
    apiDescription:
      'Bring AI into your applications, editors, and workflows. Manage API keys, usage, and request history in one console.',
    apiFeatures: [
      'Individual API keys',
      'Visible usage and billing',
      'Choose available channels',
    ],
    channels: 'Available channels',
    docs: 'Documentation',
    compatible: 'Tools you work with',
    example: 'Quick start',
    copyCode: 'Copy code',
    copyEndpoint: 'Copy base URL',
    smsLabel: '03 / SMS CODES',
    smsTitle: 'Phone verification?\nGet an SMS code.',
    smsDescription:
      'Claude or ChatGPT asking for phone verification, but you do not have an eligible number? Purchase an SMS code for your platform and enter it on the verification page.',
    steps: [
      {
        title: 'Choose a platform',
        description: 'Select Claude, ChatGPT, or another platform',
      },
      {
        title: 'Place your order',
        description: 'Confirm the price and number availability, then purchase',
      },
      {
        title: 'Receive your code',
        description: 'Enter the SMS code to complete phone verification',
      },
    ],
    smsFlow: {
      title: 'Phone verification flow',
      number: 'Get a phone number',
      numberNote: 'Use a number available for your platform',
      code: 'Receive the SMS code',
      complete: 'Complete verification',
      completeNote: 'Enter the code on your platform to continue',
    },
    smsContactSubtitle:
      'Contact us on WeChat to confirm the price and purchase an SMS code for your platform.',
    copySmsRequest: 'Copy purchase request',
    faqLabel: 'GOOD TO KNOW',
    faqTitle: 'A few things before you start.',
    faqs: [
      {
        q: 'How do APIs and subscriptions differ?',
        a: 'APIs let applications, editors, and tools call models. Subscriptions provide membership features on the respective AI platform. They are separate services; a subscription does not include API credit.',
      },
      {
        q: 'How do I purchase a subscription?',
        a: 'Choose a plan and contact support on WeChat. Confirm the plan, account requirements, price, and activation process before proceeding. This page is a service catalog.',
      },
      {
        q: 'What does subscription coverage include?',
        a: 'If a subscription lapses, compensation is calculated for the remaining days. Account bans are not covered. Confirm all plan details before activation.',
      },
      {
        q: 'What are SMS codes and how do I buy one?',
        a: 'When Claude, ChatGPT, or another platform requests phone verification and you do not have an eligible number, you can purchase an SMS code. Tell us your platform on WeChat, confirm the price and place an order, then enter the received code on that platform. Supported platforms depend on current number availability.',
      },
      {
        q: 'Where can I see models and request history?',
        a: 'Sign in to the console to see available channels, create API keys, and view usage and request history. Models and quotas depend on your account configuration.',
      },
    ],
    footerDescription:
      'Model APIs, AI subscriptions, and SMS codes.\nGet to the work you want to do.',
    services: 'Services',
    support: 'Support',
    contact: 'Contact support',
    contactTitle: 'Talk to us',
    contactSubtitle: 'Confirm the details with us on WeChat.',
    scan: 'Scan with WeChat',
    copyWechat: 'Copy WeChat ID',
    copyInquiry: 'Copy your inquiry',
    close: 'Close',
    copied: 'Copied',
    copyFailed: 'Please select the text and copy it manually',
    contactGeneric: 'Hi, I would like to learn about your AI services.',
  },
}
