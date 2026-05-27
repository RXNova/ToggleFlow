<template>
  <div class="flex h-full flex-col overflow-hidden">
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
      <div class="shrink-0 space-y-4 px-6 pt-6 pb-4">
        <div class="flex items-center justify-between">
          <div class="space-y-0.5">
            <h1 class="text-base font-semibold">{{ $t('keys.apiKeysTitle') }}</h1>
            <p class="text-sm text-muted-foreground">{{ $t('keys.apiKeysSubtitle') }}</p>
          </div>
          <Button size="sm" @click="createOpen = true">
            <Plus class="size-3.5" />
            {{ $t('keys.add') }}
          </Button>
        </div>
        <Separator />
      </div>

      <div class="min-h-0 flex-1 overflow-y-auto px-6 pb-6">
        <div v-if="loading" class="flex h-full items-center justify-center">
          <Loader2 class="size-6 animate-spin text-muted-foreground/40" />
        </div>

        <div
          v-else-if="keys.length === 0"
          class="flex h-full flex-col items-center justify-center text-center"
        >
          <KeyRound class="size-8 text-muted-foreground/30 mb-3" />
          <p class="text-sm font-medium">{{ $t('keys.noKeys') }}</p>
        </div>

        <div v-else class="space-y-2">
          <div
            v-for="k in keys"
            :key="k.id"
            class="rounded-lg border bg-card px-4 py-3 flex items-center gap-3"
          >
            <div class="flex-1 min-w-0 space-y-0.5">
              <p class="text-sm font-medium">{{ k.label }}</p>
              <p class="text-xs font-mono text-muted-foreground">{{ k.key_prefix }}...</p>
            </div>
            <div class="text-right text-xs text-muted-foreground space-y-0.5 shrink-0">
              <p>
                {{ $t('keys.expiry') }}:
                {{ k.expires_at ? formatDate(k.expires_at) : $t('keys.expiryNever') }}
              </p>
              <p>{{ formatDate(k.created_at) }}</p>
            </div>
            <button
              class="rounded-md p-1.5 text-muted-foreground transition-colors hover:bg-destructive/10 hover:text-destructive"
              @click="deleteKey(k)"
            >
              <Trash2 class="size-3.5" />
            </button>
          </div>
        </div>
      </div>
    </template>
  </div>

  <CreateKeyDialog
    v-if="projectStore.current"
    v-model:open="createOpen"
    :title="$t('keys.createApiTitle')"
    :description="$t('keys.createApiDescription')"
    :on-create="
      (label, expiresAt) =>
        environmentsApi.apiKeys
          .create(projectStore.current!.id, label, expiresAt)
          .then((r) => r.key)
    "
    @created="load"
  />
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { FolderOpen, Plus, Loader2, Trash2, KeyRound } from '@lucide/vue'
import { Button } from '@/components/ui/button'
import { Separator } from '@/components/ui/separator'
import { useProjectStore } from '@/stores/project'
import { environmentsApi, type APIKeyRecord } from '@/api/environments'
import CreateKeyDialog from '@/components/CreateKeyDialog.vue'

const projectStore = useProjectStore()
const keys = ref<APIKeyRecord[]>([])
const loading = ref(false)
const createOpen = ref(false)

async function load() {
  if (!projectStore.current) return
  loading.value = true
  try {
    keys.value = ((await environmentsApi.apiKeys.list(projectStore.current.id)) ??
      []) as APIKeyRecord[]
  } finally {
    loading.value = false
  }
}

watch(() => projectStore.current, load, { immediate: true })

async function deleteKey(k: APIKeyRecord) {
  if (!projectStore.current) return
  await environmentsApi.apiKeys.delete(projectStore.current.id, k.id)
  keys.value = keys.value.filter((x) => x.id !== k.id)
}

function formatDate(iso: string) {
  return new Date(iso).toLocaleDateString(undefined, {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  })
}
</script>
