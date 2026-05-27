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
        <div class="flex items-center justify-between">
          <div class="space-y-0.5">
            <h1 class="text-base font-semibold">{{ $t('flags.title') }}</h1>
            <p class="text-sm text-muted-foreground">{{ $t('flags.subtitle') }}</p>
          </div>
          <Button size="sm" @click="createDialogOpen = true">
            <Plus class="size-3.5" />
            {{ $t('flags.create') }}
          </Button>
        </div>

        <Separator />

        <div class="relative">
          <Search
            class="absolute left-2.5 top-1/2 -translate-y-1/2 size-3.5 text-muted-foreground pointer-events-none"
          />
          <Input v-model="search" :placeholder="$t('flags.search')" class="pl-8" />
        </div>
      </div>

      <!-- Scrollable body + pinned pagination -->
      <div class="flex min-h-0 flex-1 flex-col overflow-hidden px-6 pb-6">
        <div v-if="loading" class="flex flex-1 items-center justify-center">
          <Loader2 class="size-6 animate-spin text-muted-foreground/40" />
        </div>

        <div
          v-else-if="flags.length === 0 && search"
          class="flex flex-1 flex-col items-center justify-center text-center"
        >
          <SearchX class="size-8 text-muted-foreground/30 mb-3" />
          <p class="text-sm font-medium">No flags match your search</p>
        </div>

        <div
          v-else-if="total === 0"
          class="flex flex-1 flex-col items-center justify-center text-center"
        >
          <FlagIcon class="size-8 text-muted-foreground/30 mb-3" />
          <p class="text-sm font-medium">{{ $t('flags.emptyTitle') }}</p>
          <p class="mt-1 text-xs text-muted-foreground max-w-xs">
            {{ $t('flags.emptyDescription') }}
          </p>
        </div>

        <template v-else>
          <div class="min-h-0 flex-1 overflow-y-auto space-y-2">
            <div v-for="flag in flags" :key="flag.id" class="rounded-lg border bg-card p-4">
              <div class="flex items-start justify-between gap-4">
                <div class="min-w-0 flex-1">
                  <!-- Name + type badge + key -->
                  <div class="flex flex-wrap items-center gap-2">
                    <p class="text-sm font-medium leading-none">{{ flag.name }}</p>
                    <span
                      class="inline-flex items-center rounded border px-1.5 py-0.5 text-[10px] font-medium text-muted-foreground"
                    >
                      {{ flag.flag_type }}
                    </span>
                    <span class="font-mono text-[11px] text-muted-foreground">{{ flag.key }}</span>
                  </div>

                  <p v-if="flag.description" class="mt-1 text-xs text-muted-foreground">
                    {{ flag.description }}
                  </p>

                  <!-- Variation chips -->
                  <div v-if="flag.variations.length" class="mt-2 flex flex-wrap gap-1">
                    <span
                      v-for="(v, i) in flag.variations"
                      :key="i"
                      class="rounded border bg-muted/40 px-1.5 py-0.5 text-[10px] font-mono"
                    >
                      {{ v.name }}:
                      <span class="text-muted-foreground">{{ formatValue(v.value) }}</span>
                    </span>
                  </div>
                </div>

                <!-- Actions -->
                <div class="flex items-center gap-1 shrink-0">
                  <Tooltip :text="$t('flags.history')">
                    <button
                      class="rounded-md p-1.5 text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
                      @click="openHistory(flag)"
                    >
                      <History class="size-3.5" />
                    </button>
                  </Tooltip>
                  <Tooltip :text="$t('common.edit')">
                    <button
                      class="rounded-md p-1.5 text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
                      @click="openEdit(flag)"
                    >
                      <Pencil class="size-3.5" />
                    </button>
                  </Tooltip>
                  <Tooltip :text="$t('common.delete')">
                    <button
                      class="rounded-md p-1.5 text-muted-foreground transition-colors hover:bg-destructive/10 hover:text-destructive"
                      @click="openDelete(flag)"
                    >
                      <Trash2 class="size-3.5" />
                    </button>
                  </Tooltip>
                </div>
              </div>

              <!-- Environment toggles -->
              <div
                v-if="flag.environments.length > 0"
                class="mt-3 flex flex-wrap gap-x-6 gap-y-2 border-t pt-3"
              >
                <div
                  v-for="env in flag.environments"
                  :key="env.environment_id"
                  class="flex items-center gap-2"
                >
                  <button
                    class="relative inline-flex h-4.5 w-8 shrink-0 rounded-full border-2 border-transparent transition-colors focus-visible:outline-none disabled:pointer-events-none disabled:opacity-50"
                    :class="[
                      env.enabled ? 'bg-primary' : 'bg-input',
                      env.protected && !authStore.isAdmin ? 'cursor-not-allowed' : 'cursor-pointer',
                    ]"
                    :disabled="
                      toggling[toggleKey(flag.id, env.environment_id)] ||
                      (env.protected && !authStore.isAdmin)
                    "
                    @click="toggle(flag, env)"
                  >
                    <span
                      class="pointer-events-none block h-3.5 w-3.5 rounded-full bg-background shadow-sm ring-0 transition-transform"
                      :class="env.enabled ? 'translate-x-3' : 'translate-x-0'"
                    />
                  </button>
                  <span class="text-xs text-muted-foreground flex items-center gap-1">
                    {{ env.environment_name }}
                    <Tooltip
                      v-if="env.protected && !authStore.isAdmin"
                      :text="$t('environments.protectedNote')"
                    >
                      <Lock class="size-3 text-orange-500/70" />
                    </Tooltip>
                  </span>
                  <!-- Default variation selector -->
                  <select
                    v-if="flag.variations.length > 1"
                    :value="env.default_variation"
                    :disabled="env.protected && !authStore.isAdmin"
                    class="rounded border border-input bg-background px-1.5 py-0.5 text-[11px] text-muted-foreground focus:outline-none focus:ring-1 focus:ring-ring disabled:opacity-50 disabled:cursor-not-allowed"
                    @change="changeDefaultVariation(flag, env, $event)"
                  >
                    <option v-for="(v, i) in flag.variations" :key="i" :value="i">
                      {{ v.name }}
                    </option>
                  </select>
                </div>
              </div>
              <p v-else class="mt-3 text-xs text-muted-foreground border-t pt-3">
                {{ $t('flags.noEnvironments') }}
              </p>

              <!-- Metadata -->
              <div
                class="mt-3 flex flex-wrap items-center gap-x-4 gap-y-1 border-t pt-3 text-xs text-muted-foreground"
              >
                <span class="flex items-center gap-1">
                  <CalendarDays class="size-3 shrink-0" />
                  {{ timeAgo(flag.created_at) }}
                </span>
                <span class="flex items-center gap-1">
                  <Clock class="size-3 shrink-0" />
                  {{ timeAgo(flag.updated_at) }}
                </span>
              </div>
            </div>
          </div>

          <div class="shrink-0 pt-4">
            <Pagination :page="page" :total="total" :limit="limit" @change="goTo" />
          </div>
        </template>
      </div>
    </template>
  </div>

  <CreateFlagDialog
    v-if="projectStore.current"
    v-model:open="createDialogOpen"
    :project-id="projectStore.current.id"
    @created="onCreated"
  />
  <EditFlagDialog
    v-if="projectStore.current"
    v-model:open="editDialogOpen"
    :flag="editTarget"
    :project-id="projectStore.current.id"
    @updated="onUpdated"
  />
  <DeleteFlagDialog
    v-if="projectStore.current"
    v-model:open="deleteDialogOpen"
    :flag="deleteTarget"
    :project-id="projectStore.current.id"
    @deleted="onDeleted"
  />
  <FlagHistorySheet
    v-if="projectStore.current && historyTarget"
    v-model:open="historyOpen"
    :project-id="projectStore.current.id"
    :flag-key="historyTarget.key"
  />
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { watchDebounced } from '@vueuse/core'
import {
  Flag as FlagIcon,
  Plus,
  FolderOpen,
  Loader2,
  Pencil,
  Trash2,
  Search,
  SearchX,
  CalendarDays,
  Clock,
  Lock,
  History,
} from '@lucide/vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Separator } from '@/components/ui/separator'
import Pagination from '@/components/ui/pagination/Pagination.vue'
import { useProjectStore } from '@/stores/project'
import { flagsApi, type Flag, type FlagEnvState } from '@/api/flags'
import { useAuthStore } from '@/stores/auth'
import { timeAgo } from '@/lib/utils'
import CreateFlagDialog from '@/components/CreateFlagDialog.vue'
import EditFlagDialog from '@/components/EditFlagDialog.vue'
import DeleteFlagDialog from '@/components/DeleteFlagDialog.vue'
import FlagHistorySheet from '@/components/FlagHistorySheet.vue'
import { Tooltip } from '@/components/ui/tooltip'

const LIMIT = 20

const projectStore = useProjectStore()
const authStore = useAuthStore()
const flags = ref<Flag[]>([])
const total = ref(0)
const page = ref(1)
const limit = ref(LIMIT)
const search = ref('')
const loading = ref(false)
const createDialogOpen = ref(false)
const editDialogOpen = ref(false)
const editTarget = ref<Flag | null>(null)
const deleteDialogOpen = ref(false)
const deleteTarget = ref<Flag | null>(null)
const historyOpen = ref(false)
const historyTarget = ref<Flag | null>(null)
const toggling = ref<Record<string, boolean>>({})

function toggleKey(flagId: number, envId: number) {
  return `${flagId}:${envId}`
}

function formatValue(value: unknown): string {
  if (typeof value === 'boolean') return String(value)
  if (typeof value === 'string') return value === '' ? '""' : `"${value}"`
  if (typeof value === 'object') return JSON.stringify(value)
  return String(value)
}

async function load() {
  if (!projectStore.current) return
  loading.value = true
  try {
    const res = await flagsApi.list(projectStore.current.id, {
      limit: limit.value,
      offset: (page.value - 1) * limit.value,
      search: search.value || undefined,
    })
    flags.value = res.data ?? []
    total.value = res.total
  } finally {
    loading.value = false
  }
}

watchDebounced(
  search,
  () => {
    page.value = 1
    load()
  },
  { debounce: 300 }
)

watch(() => projectStore.current, load, { immediate: true })

function goTo(p: number) {
  page.value = p
  load()
}

function openEdit(flag: Flag) {
  editTarget.value = flag
  editDialogOpen.value = true
}

function openDelete(flag: Flag) {
  deleteTarget.value = flag
  deleteDialogOpen.value = true
}

function openHistory(flag: Flag) {
  historyTarget.value = flag
  historyOpen.value = true
}

function onCreated() {
  load()
}

function onUpdated(updated: Flag) {
  const i = flags.value.findIndex((f) => f.id === updated.id)
  if (i !== -1) {
    // preserve live environment state from the list — backend update returns nil envs
    flags.value[i] = { ...updated, environments: flags.value[i].environments }
  }
}

function onDeleted(flag: Flag) {
  flags.value = flags.value.filter((f) => f.id !== flag.id)
  total.value--
  if (page.value > 1 && flags.value.length === 0) {
    page.value--
    load()
  }
}

async function toggle(flag: Flag, env: FlagEnvState) {
  if (!projectStore.current) return
  const k = toggleKey(flag.id, env.environment_id)
  toggling.value[k] = true
  const prev = env.enabled
  env.enabled = !prev
  try {
    await flagsApi.toggle(
      projectStore.current.id,
      flag.key,
      env.environment_id,
      env.enabled,
      env.default_variation
    )
  } catch {
    env.enabled = prev
  } finally {
    delete toggling.value[k]
  }
}

async function changeDefaultVariation(flag: Flag, env: FlagEnvState, event: Event) {
  if (!projectStore.current) return
  const newVariation = Number((event.target as HTMLSelectElement).value)
  const prev = env.default_variation
  env.default_variation = newVariation
  try {
    await flagsApi.toggle(
      projectStore.current.id,
      flag.key,
      env.environment_id,
      env.enabled,
      newVariation
    )
  } catch {
    env.default_variation = prev
  }
}
</script>
