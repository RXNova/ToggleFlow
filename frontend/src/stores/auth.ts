import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { User } from '@/api/auth'

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(localStorage.getItem('token'))
  const user = ref<User | null>(JSON.parse(localStorage.getItem('user') ?? 'null'))

  const isAuthenticated = computed(() => !!token.value)
  const isSuperuser = computed(() => user.value?.role === 'superuser')
  const isAdmin = computed(() => ['superuser', 'admin'].includes(user.value?.role ?? ''))
  const canEdit = computed(() =>
    ['superuser', 'admin', 'owner', 'editor'].includes(user.value?.role ?? '')
  )

  function setAuth(newToken: string, newUser: User) {
    token.value = newToken
    user.value = newUser
    localStorage.setItem('token', newToken)
    localStorage.setItem('user', JSON.stringify(newUser))
  }

  function logout() {
    token.value = null
    user.value = null
    localStorage.removeItem('token')
    localStorage.removeItem('user')
  }

  return { token, user, isAuthenticated, isSuperuser, isAdmin, canEdit, setAuth, logout }
})
