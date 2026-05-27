<template>
  <!-- Logo -->
  <div class="flex h-12 items-center gap-2 border-b px-4 shrink-0">
    <ToggleRight class="size-4 text-sidebar-foreground" />
    <span class="text-sm font-semibold tracking-tight text-sidebar-foreground">ToggleFlow</span>
  </div>

  <!-- Project selector -->
  <div class="border-b px-2 py-2.5 shrink-0">
    <p
      class="px-2 pb-1.5 text-[10px] font-medium uppercase tracking-wider text-sidebar-foreground/40"
    >
      {{ $t('nav.project') }}
    </p>
    <DropdownMenu>
      <DropdownMenuTrigger as-child>
        <button
          class="flex w-full items-center gap-2.5 rounded-md px-2 py-2 text-left transition-colors hover:bg-sidebar-accent"
          :class="projectStore.current ? '' : 'opacity-60'"
        >
          <!-- Project avatar -->
          <div
            v-if="projectStore.current"
            class="flex size-6 shrink-0 items-center justify-center rounded-md text-[11px] font-bold uppercase text-white"
            :style="{ background: projectColor(projectStore.current.name) }"
          >
            {{ projectStore.current.name[0] }}
          </div>
          <div
            v-else
            class="flex size-6 shrink-0 items-center justify-center rounded-md bg-sidebar-accent"
          >
            <FolderOpen class="size-3.5 text-sidebar-foreground/50" />
          </div>

          <div class="flex-1 min-w-0">
            <p class="truncate text-xs font-semibold text-sidebar-foreground leading-none">
              {{ projectStore.current?.name ?? $t('projects.select') }}
            </p>
            <p
              v-if="projectStore.current"
              class="mt-0.5 truncate text-[10px] text-sidebar-foreground/50 leading-none"
            >
              {{ projectStore.current.key }}
            </p>
          </div>
          <ChevronsUpDown class="size-3.5 shrink-0 text-sidebar-foreground/40" />
        </button>
      </DropdownMenuTrigger>
      <DropdownMenuContent class="w-52" align="start">
        <DropdownMenuItem
          v-for="p in projectStore.projects"
          :key="p.id"
          class="gap-2"
          @click="projectStore.setCurrent(p)"
        >
          <Check v-if="p.id === projectStore.current?.id" class="size-3.5" />
          <span v-else class="size-3.5 inline-block" />
          {{ p.name }}
        </DropdownMenuItem>
        <DropdownMenuSeparator v-if="projectStore.projects.length" />
        <DropdownMenuItem class="gap-2" @click="createDialogOpen = true">
          <Plus class="size-3.5" />
          {{ $t('projects.new') }}
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  </div>

  <!-- Nav -->
  <nav class="flex-1 overflow-y-auto px-2 py-2.5 flex flex-col gap-4">
    <!-- Project nav -->
    <div class="space-y-0.5">
      <button
        v-for="item in projectNavItems"
        :key="item.to"
        class="flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-sm transition-colors border-l-[3px]"
        :class="
          isActive(item.to)
            ? 'bg-sidebar-accent font-medium text-sidebar-accent-foreground'
            : 'border-l-transparent text-sidebar-foreground hover:bg-sidebar-accent hover:text-sidebar-accent-foreground'
        "
        :style="isActive(item.to) ? { borderLeftColor: item.color } : {}"
        @click="navigate(item.to)"
      >
        <component
          :is="item.icon"
          class="size-4 shrink-0"
          :style="isActive(item.to) ? { color: item.color } : {}"
        />
        {{ item.label }}
      </button>
    </div>

    <!-- Administration section — admin+ only -->
    <div v-if="authStore.isAdmin" class="space-y-0.5">
      <p
        class="px-2 pb-1 text-[10px] font-medium uppercase tracking-wider text-sidebar-foreground/40"
      >
        {{ $t('nav.administration') }}
      </p>
      <button
        v-for="item in adminNavItems"
        :key="item.to"
        class="flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-sm transition-colors border-l-[3px]"
        :class="
          isActive(item.to)
            ? 'bg-sidebar-accent font-medium text-sidebar-accent-foreground'
            : 'border-l-transparent text-sidebar-foreground hover:bg-sidebar-accent hover:text-sidebar-accent-foreground'
        "
        :style="isActive(item.to) ? { borderLeftColor: item.color } : {}"
        @click="navigate(item.to)"
      >
        <component
          :is="item.icon"
          class="size-4 shrink-0"
          :style="isActive(item.to) ? { color: item.color } : {}"
        />
        {{ item.label }}
      </button>
    </div>
  </nav>

  <!-- User footer -->
  <div class="border-t p-2 space-y-0.5 shrink-0">
    <button
      class="flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-left transition-colors"
      :class="
        isActive('/settings')
          ? 'bg-sidebar-accent text-sidebar-accent-foreground'
          : 'text-sidebar-foreground hover:bg-sidebar-accent hover:text-sidebar-accent-foreground'
      "
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

  <CreateProjectDialog v-model:open="createDialogOpen" @created="onProjectCreated" />
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import {
  ToggleRight,
  Flag,
  FolderOpen,
  Globe,
  ClipboardList,
  Users,
  UserCheck,
  KeyRound,
  LogOut,
  Settings,
  ChevronsUpDown,
  Check,
  Plus,
} from '@lucide/vue'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import CreateProjectDialog from '@/components/CreateProjectDialog.vue'
import { useAuthStore } from '@/stores/auth'
import { useProjectStore } from '@/stores/project'
import type { Project } from '@/api/projects'

const emit = defineEmits<{ navigate: [] }>()

const { t } = useI18n()
const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()
const projectStore = useProjectStore()
const createDialogOpen = ref(false)

const projectNavItems = computed(() => {
  const items = [
    { to: '/projects', icon: FolderOpen, label: t('nav.projects'), color: 'oklch(0.56 0.12 22)' },
    { to: '/flags', icon: Flag, label: t('nav.flags'), color: 'oklch(0.60 0.11 245)' },
    {
      to: '/environments',
      icon: Globe,
      label: t('nav.environments'),
      color: 'oklch(0.60 0.14 55)',
    },
    { to: '/members', icon: UserCheck, label: t('nav.members'), color: 'oklch(0.52 0.12 290)' },
    { to: '/audit', icon: ClipboardList, label: t('nav.audit'), color: 'oklch(0.58 0.10 175)' },
  ]
  return items
})

const adminNavItems = computed(() => [
  { to: '/keys', icon: KeyRound, label: t('nav.apiKeys'), color: 'oklch(0.54 0.12 265)' },
  { to: '/users', icon: Users, label: t('nav.users'), color: 'oklch(0.54 0.12 265)' },
])

// Deterministic color from project name — cycles through a fixed palette
const PROJECT_COLORS = [
  '#6366f1',
  '#8b5cf6',
  '#ec4899',
  '#f59e0b',
  '#10b981',
  '#3b82f6',
  '#ef4444',
  '#14b8a6',
]
function projectColor(name: string): string {
  let hash = 0
  for (let i = 0; i < name.length; i++) hash = name.charCodeAt(i) + ((hash << 5) - hash)
  return PROJECT_COLORS[Math.abs(hash) % PROJECT_COLORS.length]
}

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

function onProjectCreated(project: Project) {
  projectStore.projects.push(project)
  projectStore.setCurrent(project)
}
</script>
