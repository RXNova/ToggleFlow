<template>
  <div class="p-6 space-y-4">
    <template v-if="!projectStore.current">
      <div class="flex flex-col items-center justify-center py-24 text-center">
        <FolderOpen class="size-8 text-muted-foreground/30 mb-3" />
        <p class="text-sm font-medium">{{ $t('projects.noProject') }}</p>
        <p class="mt-1 text-xs text-muted-foreground max-w-xs">{{ $t('projects.noProjectDescription') }}</p>
      </div>
    </template>

    <template v-else>
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

      <div v-if="loading" class="flex flex-col items-center justify-center py-24">
        <Loader2 class="size-6 animate-spin text-muted-foreground/40" />
      </div>

      <div v-else-if="flags.length === 0" class="flex flex-col items-center justify-center py-24 text-center">
        <FlagIcon class="size-8 text-muted-foreground/30 mb-3" />
        <p class="text-sm font-medium">{{ $t('flags.emptyTitle') }}</p>
        <p class="mt-1 text-xs text-muted-foreground max-w-xs">{{ $t('flags.emptyDescription') }}</p>
      </div>

      <div v-else class="space-y-2">
        <div
          v-for="flag in flags"
          :key="flag.id"
          class="rounded-lg border bg-card p-4"
        >
          <div class="flex items-start justify-between gap-4">
            <div class="min-w-0 flex-1">
              <div class="flex items-center gap-2">
                <p class="text-sm font-medium">{{ flag.name }}</p>
                <span class="rounded border px-1.5 py-0.5 font-mono text-[10px] text-muted-foreground">
                  {{ flag.flag_type }}
                </span>
              </div>
              <p class="text-xs font-mono text-muted-foreground mt-0.5">{{ flag.key }}</p>
              <p v-if="flag.description" class="text-xs text-muted-foreground mt-1">{{ flag.description }}</p>

              <!-- Variation chips -->
              <div v-if="flag.variations.length" class="mt-2 flex flex-wrap gap-1">
                <span
                  v-for="(v, i) in flag.variations"
                  :key="i"
                  class="rounded border bg-muted/40 px-1.5 py-0.5 text-[10px] font-mono"
                >
                  {{ v.name }}: <span class="text-muted-foreground">{{ formatValue(v.value) }}</span>
                </span>
              </div>
            </div>

            <button
              class="shrink-0 rounded-md p-1 text-muted-foreground transition-colors hover:bg-destructive/10 hover:text-destructive"
              @click="deleteFlag(flag)"
            >
              <Trash2 class="size-3.5" />
            </button>
          </div>

          <!-- Environment toggles -->
          <div v-if="flag.environments.length > 0" class="mt-3 flex flex-wrap gap-3 border-t pt-3">
            <div
              v-for="env in flag.environments"
              :key="env.environment_id"
              class="flex items-center gap-2"
            >
              <button
                class="relative inline-flex h-4.5 w-8 shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors focus-visible:outline-none disabled:pointer-events-none"
                :class="env.enabled ? 'bg-primary' : 'bg-input'"
                :disabled="toggling[toggleKey(flag.id, env.environment_id)]"
                @click="toggle(flag, env)"
              >
                <span
                  class="pointer-events-none block h-3.5 w-3.5 rounded-full bg-background shadow-sm ring-0 transition-transform"
                  :class="env.enabled ? 'translate-x-3' : 'translate-x-0'"
                />
              </button>
              <span class="text-xs text-muted-foreground">{{ env.environment_name }}</span>
            </div>
          </div>
          <p v-else class="mt-3 text-xs text-muted-foreground border-t pt-3">
            {{ $t('flags.noEnvironments') }}
          </p>
        </div>
      </div>
    </template>
  </div>

  <CreateFlagDialog
    v-if="projectStore.current"
    v-model:open="createDialogOpen"
    :project-id="projectStore.current.id"
    @created="onCreated"
  />
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { Flag as FlagIcon, Plus, FolderOpen, Loader2, Trash2 } from '@lucide/vue'
import { Button } from '@/components/ui/button'
import { Separator } from '@/components/ui/separator'
import { useProjectStore } from '@/stores/project'
import { flagsApi, type Flag, type FlagEnvState } from '@/api/flags'
import CreateFlagDialog from '@/components/CreateFlagDialog.vue'

const projectStore = useProjectStore()
const flags = ref<Flag[]>([])
const loading = ref(false)
const createDialogOpen = ref(false)
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
    flags.value = (await flagsApi.list(projectStore.current.id)).data ?? []
  } finally {
    loading.value = false
  }
}

watch(() => projectStore.current, load, { immediate: true })

function onCreated() {
  load()
}

async function toggle(flag: Flag, env: FlagEnvState) {
  if (!projectStore.current) return
  const k = toggleKey(flag.id, env.environment_id)
  toggling.value[k] = true
  const prev = env.enabled
  env.enabled = !prev
  try {
    await flagsApi.toggle(projectStore.current.id, flag.key, env.environment_id, env.enabled, env.default_variation)
  } catch {
    env.enabled = prev
  } finally {
    delete toggling.value[k]
  }
}

async function deleteFlag(flag: Flag) {
  if (!projectStore.current) return
  if (!confirm(`Delete flag "${flag.name}"?`)) return
  try {
    await flagsApi.delete(projectStore.current.id, flag.key)
    flags.value = flags.value.filter(f => f.id !== flag.id)
  } catch {
    // ignore
  }
}
</script>
