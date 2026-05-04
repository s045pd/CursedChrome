import { createRouter, createWebHistory, type RouteLocationNormalized } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const routes = [
  {
    path: '/login',
    name: 'login',
    component: () => import('@/pages/Login.vue'),
    meta: { layout: 'blank', public: true },
  },
  {
    path: '/',
    component: () => import('@/layouts/AppShell.vue'),
    children: [
      {
        path: '',
        name: 'dashboard',
        component: () => import('@/pages/Dashboard.vue'),
      },
      {
        path: 'bots/:id',
        name: 'bot-detail',
        component: () => import('@/pages/BotDetail.vue'),
        props: true,
      },
      {
        path: 'settings',
        name: 'settings',
        component: () => import('@/pages/Settings.vue'),
      },
    ],
  },
  { path: '/:pathMatch(.*)*', redirect: '/' },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

router.beforeEach(async (to: RouteLocationNormalized) => {
  const auth = useAuthStore()
  if (to.meta.public) return true

  if (!auth.isAuthenticated) {
    const ok = await auth.refresh()
    if (!ok) return { name: 'login', query: { redirect: to.fullPath } }
  }
  return true
})

export default router
