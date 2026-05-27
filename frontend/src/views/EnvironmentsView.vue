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
            <h1 class="text-base font-semibold">{{ $t('environments.title') }}</h1>
            <p class="text-sm text-muted-foreground">{{ $t('environments.subtitle') }}</p>
          </div>
          <Button size="sm" @click="createDialogOpen = true">
            <Plus class="size-3.5" />
            {{ $t('environments.create') }}
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
          v-else-if="environments.length === 0"
          class="flex h-full flex-col items-center justify-center text-center"
        >
          <Globe class="size-8 text-muted-foreground/30 mb-3" />
          <p class="text-sm font-medium">{{ $t('environments.emptyTitle') }}</p>
          <p class="mt-1 text-xs text-muted-foreground max-w-xs">
            {{ $t('environments.emptyDescription') }}
          </p>
        </div>

        <div v-else class="section-envs space-y-2">
          <div
            v-for="env in environments"
            :key="env.id"
            class="rounded-lg border bg-card p-4 space-y-3"
          >
            <div class="flex items-start justify-between gap-4">
              <div class="min-w-0 flex-1">
                <div class="flex flex-wrap items-center gap-2">
                  <p class="text-sm font-medium leading-none">{{ env.name }}</p>
                  <CopyKey :value="env.key" />
                  <span
                    v-if="env.protected"
                    class="inline-flex items-center gap-1 rounded px-1.5 py-0.5 text-[10px] font-medium bg-orange-500/10 text-orange-600 dark:text-orange-400"
                  >
                    <Lock class="size-2.5" />
                    {{ $t('environments.protected') }}
                  </span>
                </div>
                <p v-if="env.description" class="mt-1 text-xs text-muted-foreground">
                  {{ env.description }}
                </p>
              </div>
              <div class="flex items-center gap-1 shrink-0">
                <Tooltip :text="$t('environments.history')">
                  <button
                    class="rounded-md p-1.5 text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
                    @click="openHistory(env)"
                  >
                    <History class="size-3.5" />
                  </button>
                </Tooltip>
                <Tooltip :text="$t('common.edit')">
                  <button
                    class="rounded-md p-1.5 text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
                    @click="openEdit(env)"
                  >
                    <Pencil class="size-3.5" />
                  </button>
                </Tooltip>
                <Tooltip :text="$t('common.delete')">
                  <button
                    class="rounded-md p-1.5 text-muted-foreground transition-colors hover:bg-destructive/10 hover:text-destructive"
                    @click="openDelete(env)"
                  >
                    <Trash2 class="size-3.5" />
                  </button>
                </Tooltip>
              </div>
            </div>

            <!-- Stats -->
            <div class="flex items-center gap-3 pt-1 border-t">
              <div class="flex items-center gap-1.5 text-xs text-muted-foreground">
                <KeyRound class="size-3.5 shrink-0" />
                <span>{{ sdkKeys[env.id]?.length ?? '—' }} {{ $t('environments.statKeys') }}</span>
              </div>
              <div class="flex items-center gap-1.5 text-xs text-muted-foreground">
                <ShieldCheck class="size-3.5 shrink-0" />
                <span>{{ activeKeyCount(env.id) }} {{ $t('environments.statActive') }}</span>
              </div>
              <div class="flex items-center gap-1.5 text-xs text-muted-foreground">
                <Flag class="size-3.5 shrink-0" />
                <span>{{ flagTotal }} {{ $t('environments.statFlags') }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </template>
  </div>

  <CreateEnvironmentDialog
    v-if="projectStore.current"
    v-model:open="createDialogOpen"
    :project-id="projectStore.current.id"
    @created="onCreated"
  />
  <EditEnvironmentDialog
    v-if="projectStore.current"
    v-model:open="editDialogOpen"
    :environment="editTarget"
    :project-id="projectStore.current.id"
    @updated="onUpdated"
  />
  <DeleteEnvironmentDialog
    v-if="projectStore.current"
    v-model:open="deleteDialogOpen"
    :environment="deleteTarget"
    :project-id="projectStore.current.id"
    @deleted="onDeleted"
  />
  <AuditHistorySheet
    v-if="projectStore.current && historyTarget"
    v-model:open="historyOpen"
    :project-id="projectStore.current.id"
    :resource="historyTarget.key"
    :title="$t('environments.historyTitle')"
    :label="historyTarget.key"
  />
</template>

<script setup lang="ts">
import { ref, watch, reactive } from 'vue'
import {
  Globe,
  Plus,
  FolderOpen,
  Loader2,
  Pencil,
  Trash2,
  Lock,
  History,
  KeyRound,
  ShieldCheck,
  Flag,
} from '@lucide/vue'
import { Button } from '@/components/ui/button'
import { Separator } from '@/components/ui/separator'
import { useProjectStore } from '@/stores/project'
import { environmentsApi, type Environment, type SDKKeyRecord } from '@/api/environments'
import { flagsApi } from '@/api/flags'
import CreateEnvironmentDialog from '@/components/CreateEnvironmentDialog.vue'
import EditEnvironmentDialog from '@/components/EditEnvironmentDialog.vue'
import DeleteEnvironmentDialog from '@/components/DeleteEnvironmentDialog.vue'
import AuditHistorySheet from '@/components/AuditHistorySheet.vue'
import { Tooltip } from '@/components/ui/tooltip'
import CopyKey from '@/components/CopyKey.vue'

const projectStore = useProjectStore()
const environments = ref<Environment[]>([])
const loading = ref(false)
const createDialogOpen = ref(false)
const editDialogOpen = ref(false)
const editTarget = ref<Environment | null>(null)
const deleteDialogOpen = ref(false)
const deleteTarget = ref<Environment | null>(null)
const historyOpen = ref(false)
const historyTarget = ref<Environment | null>(null)
const sdkKeys = reactive<Record<number, SDKKeyRecord[]>>({})
const flagTotal = ref<number | string>('—')

async function load() {
  if (!projectStore.current) return
  loading.value = true
  try {
    const pid = projectStore.current.id
    const [envsResult, flagsResult] = await Promise.all([
      environmentsApi.list(pid),
      flagsApi.list(pid, { page: 1, page_size: 1 }),
    ])
    environments.value = envsResult.data ?? []
    flagTotal.value = flagsResult.total
    for (const env of environments.value) {
      loadSDKKeys(pid, env.id)
    }
  } finally {
    loading.value = false
  }
}

async function loadSDKKeys(pid: number, eid: number) {
  sdkKeys[eid] = ((await environmentsApi.sdkKeys.list(pid, eid)) ?? []) as SDKKeyRecord[]
}

function activeKeyCount(envId: number): number | string {
  const keys = sdkKeys[envId]
  if (!keys) return '—'
  const now = Date.now()
  return keys.filter((k) => !k.expires_at || new Date(k.expires_at).getTime() > now).length
}

watch(() => projectStore.current, load, { immediate: true })

function onCreated(env: Environment) {
  environments.value.push(env)
  if (projectStore.current) loadSDKKeys(projectStore.current.id, env.id)
}

function openEdit(env: Environment) {
  editTarget.value = env
  editDialogOpen.value = true
}

function openDelete(env: Environment) {
  deleteTarget.value = env
  deleteDialogOpen.value = true
}

function openHistory(env: Environment) {
  historyTarget.value = env
  historyOpen.value = true
}

function onUpdated(updated: Environment) {
  const i = environments.value.findIndex((e) => e.id === updated.id)
  if (i !== -1) environments.value[i] = updated
}

function onDeleted(env: Environment) {
  environments.value = environments.value.filter((e) => e.id !== env.id)
}
</script>
