<template>
  <!-- Logo -->
  <div class="flex h-12 items-center gap-2 border-b px-4 shrink-0">
    <ToggleRight class="size-4 text-sidebar-foreground" />
    <span class="text-sm font-semibold tracking-tight text-sidebar-foreground">ToggleFlow</span>
  </div>

  <!-- Nav -->
  <nav class="flex-1 space-y-0.5 overflow-y-auto px-2 py-2">
    <button
      v-for="item in navItems"
      :key="item.to"
      class="flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-sm transition-colors"
      :class="isActive(item.to)
        ? 'bg-sidebar-accent font-medium text-sidebar-accent-foreground'
        : 'text-sidebar-foreground hover:bg-sidebar-accent hover:text-sidebar-accent-foreground'"
      @click="navigate(item.to)"
    >
      <component :is="item.icon" class="size-4 shrink-0" />
      {{ item.label }}
    </button>
  </nav>

  <!-- User footer -->
  <div class="border-t p-2 space-y-0.5 shrink-0">
    <button
      class="flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-left transition-colors"
      :class="isActive('/settings')
        ? 'bg-sidebar-accent text-sidebar-accent-foreground'
        : 'text-sidebar-foreground hover:bg-sidebar-accent hover:text-sidebar-accent-foreground'"
      @click="navigate('/settings')"
    >
      <div class="flex-1 min-w-0">
        <p class="text-xs font-medium truncate">{{ authStore.user?.name }}</p>
        <p class="text-xs opacity-50">{{ $t(`roles.${authStore.user?.role}`) }}</p>
      </div>
      <Settings class="size-3.5 shrink-0 opacity-50" />
    </button>
    <button
      class="flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-sm text-sidebar-foreground transition-colors hover:bg-sidebar-accent hover:text-sidebar-accent-foreground"
      @click="signOut"
    >
      <LogOut class="size-4 shrink-0" />
      {{ $t('nav.signOut') }}
    </button>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ToggleRight, Flag, Globe, ClipboardList, Users, LogOut, Settings } from '@lucide/vue'
import { useAuthStore } from '@/stores/auth'

const emit = defineEmits<{ navigate: [] }>()

const { t } = useI18n()
const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()

const navItems = computed(() => {
  const items = [
    { to: '/flags',        icon: Flag,         label: t('nav.flags') },
    { to: '/environments', icon: Globe,         label: t('nav.environments') },
    { to: '/audit',        icon: ClipboardList, label: t('nav.audit') },
  ]
  if (authStore.isAdmin) {
    items.push({ to: '/users', icon: Users, label: t('nav.users') })
  }
  return items
})

function isActive(path: string) {
  return route.path.startsWith(path)
}

function navigate(to: string) {
  emit('navigate')
  router.push(to)
}

function signOut() {
  emit('navigate')
  authStore.logout()
  router.push('/login')
}
</script>
