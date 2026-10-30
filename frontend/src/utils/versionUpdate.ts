/**
 * 前端新版本检测
 *
 * 单页应用打开后不会再请求 index.html，发版后已打开的页面会一直运行旧代码，
 * 只有用户手动刷新才会更新。这里定期以协商缓存方式请求入口页，比对入口脚本地址：
 * Vite 产物按内容哈希命名，代码有变化时入口脚本地址必然变化，与版本号是否更新无关。
 * 不比对 index.html 的 ETag，因为后台修改公开设置也会改变注入后的页面内容。
 */
import { readonly, ref } from 'vue'

/** 定时检测间隔：5 分钟 */
const CHECK_INTERVAL_MS = 5 * 60 * 1000
/** 两次检测的最小间隔，避免切换标签页、频繁跳转时重复请求 */
const MIN_CHECK_GAP_MS = 60 * 1000
/** 从入口页 HTML 中提取入口模块脚本地址（注入的设置脚本没有 type="module"，不会误匹配） */
const ENTRY_SCRIPT_PATTERN = /<script\b[^>]*\btype=["']module["'][^>]*\bsrc=["']([^"']+)["']/i

const updateAvailable = ref(false)
const noticeDismissed = ref(false)

let currentEntry: string | null = null
let lastCheckAt = 0
let checking = false
let started = false

function readCurrentEntry(): string | null {
  const script = document.querySelector<HTMLScriptElement>('script[type="module"][src]')
  return script?.getAttribute('src') ?? null
}

/**
 * @brief 请求入口页并比对入口脚本地址，不一致时标记为有新版本
 * @param force 为 true 时忽略最小间隔限制
 * @note 网络异常、非 2xx 响应或解析不到入口脚本时静默跳过，等下一次检测
 */
export async function checkForUpdate(force = false): Promise<void> {
  if (!started || updateAvailable.value || checking) return
  const now = Date.now()
  if (!force && now - lastCheckAt < MIN_CHECK_GAP_MS) return

  checking = true
  lastCheckAt = now
  try {
    // no-cache 走 ETag 协商，入口页未变化时服务端返回 304，开销很小
    const res = await fetch('/', {
      cache: 'no-cache',
      credentials: 'same-origin',
      headers: { Accept: 'text/html' }
    })
    if (!res.ok) return
    const match = ENTRY_SCRIPT_PATTERN.exec(await res.text())
    if (match && match[1] !== currentEntry) {
      updateAvailable.value = true
    }
  } catch {
    // 网络波动时忽略，下次再检测
  } finally {
    checking = false
  }
}

/**
 * @brief 启动新版本检测：定时检测，并在页面重新可见时补测一次
 * @note 仅生产构建生效；开发环境入口是 /src/main.ts，没有哈希，不做检测。由 App.vue 挂载时调用，重复调用无副作用。
 */
export function startVersionUpdateCheck(): void {
  if (started || !import.meta.env.PROD) return
  currentEntry = readCurrentEntry()
  if (!currentEntry) return
  started = true

  window.setInterval(() => {
    if (!document.hidden) void checkForUpdate()
  }, CHECK_INTERVAL_MS)
  document.addEventListener('visibilitychange', () => {
    if (!document.hidden) void checkForUpdate()
  })
}

/**
 * @brief 判断本次站内跳转是否应改为整页加载
 * @param fromPath 跳转前路径；首次导航时没有匹配的路由，传 null
 * @param toPath 跳转目标路径
 * @return 已检测到新版本且跳转到不同页面时返回 true
 * @note 仅比较 path，同一页面内只改 query（如切换标签）不触发整页加载，避免打断正在操作的内容
 */
export function shouldReloadOnNavigation(fromPath: string | null, toPath: string): boolean {
  return updateAvailable.value && fromPath !== null && fromPath !== toPath
}

/**
 * @brief 新版本提示条使用的状态与操作
 */
export function useVersionUpdate() {
  return {
    updateAvailable: readonly(updateAvailable),
    noticeDismissed: readonly(noticeDismissed),
    /** 关闭提示，本次页面生命周期内不再弹出；切换页面时仍会自动加载新版本 */
    dismissNotice: () => {
      noticeDismissed.value = true
    },
    reloadNow: () => {
      window.location.reload()
    }
  }
}
