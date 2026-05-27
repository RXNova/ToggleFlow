<template>
  <div class="flex h-full flex-col overflow-hidden">
    <!-- Fixed header -->
    <div class="shrink-0 space-y-4 px-6 pt-6 pb-4">
      <div class="flex items-center justify-between">
        <div class="space-y-0.5">
          <h1 class="text-base font-semibold">{{ $t('users.title') }}</h1>
          <p class="text-sm text-muted-foreground">{{ $t('users.subtitle') }}</p>
        </div>
        <Button v-if="authStore.isAdmin" size="sm" @click="createDialogOpen = true">
          <UserPlus class="size-3.5" />
          {{ $t('users.invite') }}
        </Button>
      </div>

      <Separator />
    </div>

    <!-- Scrollable body -->
    <div class="min-h-0 flex-1 overflow-y-auto px-6 pb-6">
      <div v-if="loading" class="flex h-full items-center justify-center">
        <Loader2 class="size-6 animate-spin text-muted-foreground/40" />
      </div>

      <div
        v-else-if="users.length === 0"
        class="flex h-full flex-col items-center justify-center text-center"
      >
        <Users class="size-8 text-muted-foreground/30 mb-3" />
        <p class="text-sm font-medium">{{ $t('users.emptyTitle') }}</p>
        <p class="mt-1 text-xs text-muted-foreground max-w-xs">
          {{ $t('users.emptyDescription') }}
        </p>
      </div>

      <div v-else class="space-y-2">
        <div v-for="user in users" :key="user.id" class="rounded-lg border bg-card p-4">
          <div class="flex items-center justify-between gap-4">
            <div class="flex items-center gap-3 min-w-0">
              <!-- Avatar -->
              <div
                class="flex size-8 shrink-0 items-center justify-center rounded-full bg-muted text-xs font-semibold uppercase"
              >
                {{ user.name[0] }}
              </div>
              <!-- Name + email -->
              <div class="min-w-0">
                <div class="flex flex-wrap items-center gap-2">
                  <p class="text-sm font-medium leading-none">{{ user.name }}</p>
                  <span
                    v-if="user.id === authStore.user?.id"
                    class="text-[10px] text-muted-foreground"
                    >({{ $t('users.you') }})</span
                  >
                  <span
                    :class="roleBadgeClass(user.role)"
                    class="inline-flex items-center rounded px-1.5 py-0.5 text-[10px] font-medium"
                  >
                    {{ $t(`roles.${user.role}`) }}
                  </span>
                  <span
                    v-if="!user.activated_at"
                    class="inline-flex items-center rounded px-1.5 py-0.5 text-[10px] font-medium bg-yellow-500/10 text-yellow-600 dark:text-yellow-400"
                  >
                    {{ $t('users.pending') }}
                  </span>
                </div>
                <p class="mt-0.5 text-xs text-muted-foreground">{{ user.email }}</p>
              </div>
            </div>

            <!-- Actions -->
            <div class="flex items-center gap-1 shrink-0">
              <span class="text-xs text-muted-foreground mr-2">
                <CalendarDays class="inline size-3 mr-0.5" />{{ timeAgo(user.created_at) }}
              </span>
              <Tooltip v-if="authStore.isAdmin" :text="$t('users.history')">
                <button
                  class="rounded-md p-1.5 text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
                  @click="openHistory(user)"
                >
                  <History class="size-3.5" />
                </button>
              </Tooltip>
              <Tooltip v-if="canEdit(user)" :text="$t('common.edit')">
                <button
                  class="rounded-md p-1.5 text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
                  @click="openEdit(user)"
                >
                  <Pencil class="size-3.5" />
                </button>
              </Tooltip>
              <Tooltip v-if="canReinvite(user)" :text="$t('users.reinviteTooltip')">
                <button
                  class="rounded-md p-1.5 text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
                  @click="openReinvite(user)"
                >
                  <RotateCcw class="size-3.5" />
                </button>
              </Tooltip>
              <Tooltip v-if="canReset(user)" :text="$t('users.resetTooltip')">
                <button
                  class="rounded-md p-1.5 text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
                  @click="openReset(user)"
                >
                  <KeyRound class="size-3.5" />
                </button>
              </Tooltip>
              <Tooltip v-if="canDelete(user)" :text="$t('common.delete')">
                <button
                  class="rounded-md p-1.5 text-muted-foreground transition-colors hover:bg-destructive/10 hover:text-destructive"
                  @click="openDelete(user)"
                >
                  <Trash2 class="size-3.5" />
                </button>
              </Tooltip>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>

  <CreateUserDialog v-model:open="createDialogOpen" @created="onCreated" />
  <EditUserDialog v-model:open="editDialogOpen" :user="editTarget" @updated="onUpdated" />
  <DeleteUserDialog v-model:open="deleteDialogOpen" :user="deleteTarget" @deleted="onDeleted" />
  <ResetLinkDialog v-model:open="resetDialogOpen" :user="resetTarget" />
  <ReinviteDialog v-model:open="reinviteDialogOpen" :user="reinviteTarget" />
  <AuditHistorySheet
    v-if="historyTarget"
    v-model:open="historyOpen"
    :user-id="historyTarget.id"
    :title="$t('users.historyTitle')"
    :label="historyTarget.email"
  />
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import {
  Users,
  UserPlus,
  Loader2,
  Pencil,
  Trash2,
  CalendarDays,
  KeyRound,
  RotateCcw,
  History,
} from '@lucide/vue'
import { Tooltip } from '@/components/ui/tooltip'
import { Button } from '@/components/ui/button'
import { Separator } from '@/components/ui/separator'
import { usersApi, type User, type Role } from '@/api/users'
import { useAuthStore } from '@/stores/auth'
import { timeAgo } from '@/lib/utils'
import CreateUserDialog from '@/components/CreateUserDialog.vue'
import EditUserDialog from '@/components/EditUserDialog.vue'
import DeleteUserDialog from '@/components/DeleteUserDialog.vue'
import ResetLinkDialog from '@/components/ResetLinkDialog.vue'
import ReinviteDialog from '@/components/ReinviteDialog.vue'
import AuditHistorySheet from '@/components/AuditHistorySheet.vue'

const authStore = useAuthStore()
const users = ref<User[]>([])
const loading = ref(false)
const createDialogOpen = ref(false)
const editDialogOpen = ref(false)
const editTarget = ref<User | null>(null)
const deleteDialogOpen = ref(false)
const deleteTarget = ref<User | null>(null)
const resetDialogOpen = ref(false)
const resetTarget = ref<User | null>(null)
const reinviteDialogOpen = ref(false)
const reinviteTarget = ref<User | null>(null)
const historyOpen = ref(false)
const historyTarget = ref<User | null>(null)

const roleRank: Record<Role, number> = {
  superuser: 5,
  admin: 4,
  owner: 3,
  editor: 2,
  viewer: 1,
}

function myRank() {
  return roleRank[authStore.user?.role as Role] ?? 0
}

function canEdit(user: User) {
  if (!authStore.isAdmin) return false
  if (user.id === authStore.user?.id) return true
  return roleRank[user.role] < myRank()
}

function canDelete(user: User) {
  if (!authStore.isSuperuser) return false
  return user.id !== authStore.user?.id
}

function canReinvite(user: User) {
  return authStore.isAdmin && !user.activated_at
}

function canReset(user: User) {
  if (!authStore.isSuperuser) return false
  if (user.id === authStore.user?.id) return false
  return !!user.activated_at
}

function roleBadgeClass(role: Role) {
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

async function load() {
  loading.value = true
  try {
    users.value = await usersApi.list()
  } finally {
    loading.value = false
  }
}

onMounted(load)

function openEdit(user: User) {
  editTarget.value = user
  editDialogOpen.value = true
}

function openDelete(user: User) {
  deleteTarget.value = user
  deleteDialogOpen.value = true
}

function openReset(user: User) {
  resetTarget.value = user
  resetDialogOpen.value = true
}

function openReinvite(user: User) {
  reinviteTarget.value = user
  reinviteDialogOpen.value = true
}

function openHistory(user: User) {
  historyTarget.value = user
  historyOpen.value = true
}

function onCreated(user: User) {
  users.value.push(user)
}

function onUpdated(updated: User) {
  const i = users.value.findIndex((u) => u.id === updated.id)
  if (i !== -1) users.value[i] = updated
  if (updated.id === authStore.user?.id) {
    authStore.setAuth(authStore.token!, { ...authStore.user, ...updated })
  }
}

function onDeleted(user: User) {
  users.value = users.value.filter((u) => u.id !== user.id)
}
</script>
