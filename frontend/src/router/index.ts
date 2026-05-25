import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { authApi } from '@/api/auth'

let initialized: boolean | null = null

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/setup', name: 'setup', component: () => import('@/views/SetupView.vue') },
    { path: '/login', name: 'login', component: () => import('@/views/LoginView.vue') },
    {
      path: '/',
      component: () => import('@/layouts/AppLayout.vue'),
      meta: { requiresAuth: true },
      children: [
        { path: '', redirect: { name: 'flags' } },
        { path: 'flags',        name: 'flags',        component: () => import('@/views/FlagsView.vue') },
        { path: 'environments', name: 'environments', component: () => import('@/views/EnvironmentsView.vue') },
        { path: 'audit',        name: 'audit',        component: () => import('@/views/AuditView.vue') },
        { path: 'users',        name: 'users',        component: () => import('@/views/UsersView.vue'), meta: { requiresAdmin: true } },
      ],
    },
  ],
})

router.beforeEach(async (to) => {
  if (initialized === null) {
    try {
      const status = await authApi.status()
      initialized = status.initialized
    } catch {
      initialized = false
    }
  }

  if (!initialized && to.name !== 'setup') return { name: 'setup' }
  if (initialized && to.name === 'setup') return { name: 'login' }

  const auth = useAuthStore()

  if (to.meta.requiresAuth && !auth.isAuthenticated) {
    if (auth.token) {
      try {
        const user = await authApi.me()
        auth.setAuth(auth.token, user)
      } catch {
        auth.logout()
        return { name: 'login' }
      }
    } else {
      return { name: 'login' }
    }
  }

  if (to.meta.requiresAdmin && !auth.isAdmin) return { name: 'flags' }

  if (to.name === 'login' && auth.isAuthenticated) return { name: 'flags' }
})

export default router
