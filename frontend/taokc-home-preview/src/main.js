import { createApp, h } from 'vue'
import { createRouter, createWebHistory, RouterView } from 'vue-router'
import App from './App.vue'
import AuthPreview from './AuthPreview.vue'
import './style.css'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', component: App },
    { path: '/login', component: AuthPreview },
    { path: '/register', component: AuthPreview },
    { path: '/:pathMatch(.*)*', redirect: '/' },
  ],
  scrollBehavior: (to) => (to.hash ? { el: to.hash, top: 98 } : { top: 0 }),
})

router.afterEach((to) => {
  if (to.path === '/') {
    document.title = 'AI 订阅、模型 API 与接码服务 | KeepCoding API'
  }
})

createApp({ render: () => h(RouterView) })
  .use(router)
  .mount('#app')
