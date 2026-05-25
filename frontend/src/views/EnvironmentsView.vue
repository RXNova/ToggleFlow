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
          <h1 class="text-base font-semibold">{{ $t('environments.title') }}</h1>
          <p class="text-sm text-muted-foreground">{{ $t('environments.subtitle') }}</p>
        </div>
        <Button size="sm" @click="createDialogOpen = true">
          <Plus class="size-3.5" />
          {{ $t('environments.create') }}
        </Button>
      </div>

      <Separator />

      <!-- Loading -->
      <div v-if="loading" class="flex flex-col items-center justify-center py-24 text-center">
        <Loader2 class="size-6 animate-spin text-muted-foreground/40 mb-3" />
      </div>

      <!-- Empty -->
      <div v-else-if="environments.length === 0" class="flex flex-col items-center justify-center py-24 text-center">
        <Globe class="size-8 text-muted-foreground/30 mb-3" />
        <p class="text-sm font-medium">{{ $t('environments.emptyTitle') }}</p>
        <p class="mt-1 text-xs text-muted-foreground max-w-xs">{{ $t('environments.emptyDescription') }}</p>
      </div>

      <!-- List -->
      <div v-else class="space-y-2">
        <div
          v-for="env in environments"
          :key="env.id"
          class="rounded-lg border bg-card p-4 space-y-3"
        >
          <div class="flex items-start justify-between gap-4">
            <div class="min-w-0">
              <p class="text-sm font-medium">{{ env.name }}</p>
              <p class="text-xs text-muted-foreground font-mono mt-0.5">{{ env.slug }}</p>
            </div>
          </div>

          <div class="flex items-center gap-2">
            <div class="flex-1 min-w-0 rounded-md border bg-muted/40 px-3 py-1.5">
              <p class="text-xs font-mono truncate text-muted-foreground">{{ env.sdk_key }}</p>
            </div>
            <Button
              size="sm"
              variant="outline"
              class="shrink-0"
              @click="copy(env.sdk_key, env.id)"
            >
              <Check v-if="copiedId === env.id" class="size-3.5 text-green-500" />
              <Copy v-else class="size-3.5" />
              {{ copiedId === env.id ? $t('environments.copied') : $t('environments.sdkKey') }}
            </Button>
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
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { Globe, Plus, FolderOpen, Loader2, Copy, Check } from '@lucide/vue'
import { Button } from '@/components/ui/button'
import { Separator } from '@/components/ui/separator'
import { useProjectStore } from '@/stores/project'
import { environmentsApi, type Environment } from '@/api/environments'
import CreateEnvironmentDialog from '@/components/CreateEnvironmentDialog.vue'

const projectStore = useProjectStore()
const environments = ref<Environment[]>([])
const loading = ref(false)
const createDialogOpen = ref(false)
const copiedId = ref<number | null>(null)

async function load() {
  if (!projectStore.current) return
  loading.value = true
  try {
    environments.value = await environmentsApi.list(projectStore.current.id)
  } finally {
    loading.value = false
  }
}

watch(() => projectStore.current, load, { immediate: true })

function onCreated(env: Environment) {
  environments.value.push(env)
}

function copy(key: string, id: number) {
  navigator.clipboard.writeText(key)
  copiedId.value = id
  setTimeout(() => { copiedId.value = null }, 2000)
}
</script>
