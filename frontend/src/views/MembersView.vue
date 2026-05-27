<template>
  <div class="flex h-full flex-col overflow-hidden">
    <!-- No project selected -->
    <div
      v-if="!projectStore.current"
      class="flex flex-1 flex-col items-center justify-center text-center"
    >
      <FolderOpen class="size-8 text-muted-foreground/30 mb-3" />
      <p class="text-sm font-medium">{{ $t('projects.noProject') }}</p>
      <p class="mt-1 text-xs text-muted-foreground max-w-xs">
        {{ $t('projects.noProjectDescription') }}
      </p>
    </div>

    <template v-else>
      <!-- Fixed header -->
      <div class="shrink-0 space-y-4 px-6 pt-6 pb-4">
        <div class="space-y-0.5">
          <h1 class="text-base font-semibold">{{ $t('members.title') }}</h1>
          <p class="text-sm text-muted-foreground">{{ $t('members.subtitle') }}</p>
        </div>
        <Separator />
        <div class="relative">
          <Search
            class="absolute left-2.5 top-1/2 -translate-y-1/2 size-3.5 text-muted-foreground pointer-events-none"
          />
          <Input v-model="search" :placeholder="$t('members.search')" class="pl-8" />
        </div>
      </div>

      <!-- Body -->
      <div class="flex min-h-0 flex-1 flex-col overflow-hidden px-6 pb-6">
        <div v-if="loading" class="flex flex-1 items-center justify-center">
          <Loader2 class="size-6 animate-spin text-muted-foreground/40" />
        </div>

        <div v-else class="min-h-0 flex-1 overflow-y-auto space-y-1">
          <!-- Member count summary -->
          <p class="pb-2 text-xs text-muted-foreground">
            {{ memberCount }} {{ $t('members.ofTotal', { total: rows.length }) }}
          </p>

          <div
            v-for="row in filteredRows"
            :key="row.id"
            class="flex items-center gap-3 rounded-lg border bg-card px-4 py-3"
            :class="row.isMember ? '' : 'opacity-60'"
          >
            <!-- Avatar -->
            <div
              class="flex size-8 shrink-0 items-center justify-center rounded-full bg-muted text-xs font-semibold uppercase"
            >
              {{ row.name[0] }}
            </div>

            <!-- Name + email -->
            <div class="flex-1 min-w-0">
              <div class="flex flex-wrap items-center gap-1.5">
                <span class="text-sm font-medium leading-none">{{ row.name }}</span>
                <span v-if="row.id === authStore.user?.id" class="text-[10px] text-muted-foreground"
                  >({{ $t('users.you') }})</span
                >
                <span
                  v-if="!row.activated_at"
                  class="inline-flex items-center rounded px-1.5 py-0.5 text-[10px] font-medium bg-yellow-500/10 text-yellow-600 dark:text-yellow-400"
                >
                  {{ $t('users.pending') }}
                </span>
              </div>
              <p class="mt-0.5 text-xs text-muted-foreground truncate">{{ row.email }}</p>
            </div>

            <!-- Role — editable select for superusers on non-self rows with lower rank -->
            <select
              v-if="canChangeRole(row)"
              :value="row.role"
              :disabled="row.roleSaving"
              class="rounded border border-input bg-background px-1.5 py-0.5 text-[11px] text-muted-foreground focus:outline-none focus:ring-1 focus:ring-ring disabled:opacity-50 disabled:cursor-not-allowed"
              @change="changeRole(row, ($event.target as HTMLSelectElement).value as Role)"
            >
              <option v-for="r in assignableRoles" :key="r" :value="r">
                {{ $t(`roles.${r}`) }}
              </option>
            </select>
            <span
              v-else
              :class="roleBadgeClass(row.role)"
              class="shrink-0 inline-flex items-center rounded px-1.5 py-0.5 text-[10px] font-medium"
            >
              {{ $t(`roles.${row.role}`) }}
            </span>

            <!-- Membership toggle -->
            <Tooltip v-if="isAdminPlus(row)" :text="$t('members.adminAccess')">
              <div
                class="relative inline-flex h-4.5 w-8 shrink-0 cursor-not-allowed items-center rounded-full border-2 border-transparent bg-primary/30"
              >
                <span
                  class="block h-3.5 w-3.5 translate-x-3 rounded-full bg-background shadow-sm"
                />
              </div>
            </Tooltip>
            <button
              v-else-if="authStore.isOwner"
              class="relative inline-flex h-4.5 w-8 shrink-0 rounded-full border-2 border-transparent transition-colors disabled:pointer-events-none disabled:opacity-50"
              :class="row.isMember ? 'bg-primary' : 'bg-input'"
              :disabled="row.toggling"
              @click="toggleMembership(row)"
            >
              <span
                class="pointer-events-none block h-3.5 w-3.5 rounded-full bg-background shadow-sm ring-0 transition-transform"
                :class="row.isMember ? 'translate-x-3' : 'translate-x-0'"
              />
            </button>
            <div
              v-else
              class="relative inline-flex h-4.5 w-8 shrink-0 rounded-full border-2 border-transparent"
              :class="row.isMember ? 'bg-primary' : 'bg-input'"
            >
              <span
                class="pointer-events-none block h-3.5 w-3.5 rounded-full bg-background shadow-sm"
                :class="row.isMember ? 'translate-x-3' : 'translate-x-0'"
              />
            </div>
          </div>

          <div
            v-if="filteredRows.length === 0 && search"
            class="flex flex-1 flex-col items-center justify-center py-12 text-center"
          >
            <p class="text-sm text-muted-foreground">{{ $t('members.noResults') }}</p>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { FolderOpen, Loader2, Search } from '@lucide/vue'
import { Separator } from '@/components/ui/separator'
import { Input } from '@/components/ui/input'
import { Tooltip } from '@/components/ui/tooltip'
import { useProjectStore } from '@/stores/project'
import { useAuthStore } from '@/stores/auth'
import { membersApi } from '@/api/members'
import { usersApi, type Role } from '@/api/users'

type UserRow = {
  id: number
  name: string
  email: string
  role: Role
  activated_at?: string
  isMember: boolean
  toggling: boolean
  roleSaving: boolean
}

const projectStore = useProjectStore()
const authStore = useAuthStore()
const rows = ref<UserRow[]>([])
const loading = ref(false)
const search = ref('')

const assignableRoles: Role[] = ['owner', 'editor', 'viewer']

const memberCount = computed(() => rows.value.filter((r) => r.isMember).length)

const filteredRows = computed(() => {
  const q = search.value.toLowerCase()
  if (!q) return rows.value
  return rows.value.filter(
    (r) => r.name.toLowerCase().includes(q) || r.email.toLowerCase().includes(q)
  )
})

async function load() {
  if (!projectStore.current) return
  loading.value = true
  try {
    const [users, members] = await Promise.all([
      usersApi.list(),
      membersApi.list(projectStore.current.id),
    ])
    const memberIds = new Set(members.map((m) => m.user_id))
    rows.value = users.map((u) => ({
      id: u.id,
      name: u.name,
      email: u.email,
      role: u.role,
      activated_at: u.activated_at,
      isMember: memberIds.has(u.id),
      toggling: false,
      roleSaving: false,
    }))
  } finally {
    loading.value = false
  }
}

watch(() => projectStore.current, load, { immediate: true })

function isAdminPlus(row: UserRow) {
  return ['superuser', 'admin'].includes(row.role)
}

function canChangeRole(row: UserRow): boolean {
  if (!authStore.isSuperuser) return false
  if (row.id === authStore.user?.id) return false
  if (row.role === 'superuser') return false
  return true
}

function roleBadgeClass(role: string) {
  switch (role) {
    case 'superuser':
      return 'bg-destructive/10 text-destructive'
    case 'admin':
      return 'bg-orange-500/10 text-orange-600 dark:text-orange-400'
    case 'owner':
      return 'bg-blue-500/10 text-blue-600 dark:text-blue-400'
    case 'editor':
      return 'bg-green-500/10 text-green-600 dark:text-green-400'
    default:
      return 'bg-muted text-muted-foreground'
  }
}

async function toggleMembership(row: UserRow) {
  if (!projectStore.current) return
  row.toggling = true
  const prev = row.isMember
  row.isMember = !prev
  try {
    if (!prev) {
      await membersApi.add(projectStore.current.id, row.id)
    } else {
      await membersApi.remove(projectStore.current.id, row.id)
    }
  } catch {
    row.isMember = prev
  } finally {
    row.toggling = false
  }
}

async function changeRole(row: UserRow, newRole: Role) {
  row.roleSaving = true
  const prev = row.role
  row.role = newRole
  try {
    await usersApi.update(row.id, { role: newRole })
    if (row.id === authStore.user?.id) {
      authStore.setAuth(authStore.token!, { ...authStore.user!, role: newRole })
    }
  } catch {
    row.role = prev
  } finally {
    row.roleSaving = false
  }
}
</script>
