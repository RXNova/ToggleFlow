import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { authApi } from '@/api/auth'

let initialized: boolean | null = null

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/setup', name: 'setup', component: () => import('@/views/SetupView.vue') },
    { path: '/login', name: 'login', component: () => import('@/views/LoginView.vue') },
    { path: '/activate', name: 'activate', component: () => import('@/views/ActivateView.vue') },
    { path: '/reset', name: 'reset', component: () => import('@/views/ResetView.vue') },
    {
      path: '/',
      component: () => import('@/layouts/AppLayout.vue'),
      meta: { requiresAuth: true },
      children: [
        { path: '', redirect: { name: 'flags' } },
        { path: 'projects', name: 'projects', component: () => import('@/views/ProjectsView.vue') },
        { path: 'flags', name: 'flags', component: () => import('@/views/FlagsView.vue') },
        {
          path: 'environments',
          name: 'environments',
          component: () => import('@/views/EnvironmentsView.vue'),
        },
        { path: 'members', name: 'members', component: () => import('@/views/MembersView.vue') },
        { path: 'audit', name: 'audit', component: () => import('@/views/AuditView.vue') },
        {
          path: 'keys',
          name: 'api-keys',
          component: () => import('@/views/APIKeysView.vue'),
          meta: { requiresAdmin: true },
        },
        {
          path: 'users',
          name: 'users',
          component: () => import('@/views/UsersView.vue'),
          meta: { requiresAdmin: true },
        },
        { path: 'settings', name: 'settings', component: () => import('@/views/SettingsView.vue') },
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

  const publicRoutes = ['setup', 'login', 'activate', 'reset']
  if (!initialized && to.name !== 'setup') return { name: 'setup' }
  if (initialized && to.name === 'setup') return { name: 'login' }
  if (publicRoutes.includes(to.name as string)) return

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
