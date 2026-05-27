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
        <div class="space-y-0.5">
          <h1 class="text-base font-semibold">{{ $t('keys.pageTitle') }}</h1>
          <p class="text-sm text-muted-foreground">{{ $t('keys.pageSubtitle') }}</p>
        </div>
        <Separator />
      </div>

      <div class="min-h-0 flex-1 overflow-y-auto px-6 pb-6">
        <div v-if="loading" class="flex h-32 items-center justify-center">
          <Loader2 class="size-6 animate-spin text-muted-foreground/40" />
        </div>

        <Tabs v-else default-value="sdk" class="space-y-4">
          <TabsList>
            <TabsTrigger value="sdk">{{ $t('keys.sdkSection') }}</TabsTrigger>
            <TabsTrigger value="api">{{ $t('keys.apiSection') }}</TabsTrigger>
            <TabsTrigger value="usage">{{ $t('keys.usageSection') }}</TabsTrigger>
          </TabsList>

          <!-- SDK Keys tab -->
          <TabsContent value="sdk" class="space-y-3">
            <p class="text-xs text-muted-foreground">{{ $t('keys.sdkSectionDescription') }}</p>

            <div v-if="environments.length === 0" class="text-sm text-muted-foreground italic">
              {{ $t('environments.emptyTitle') }}
            </div>

            <div v-else class="space-y-3">
              <div
                v-for="env in environments"
                :key="env.id"
                class="rounded-lg border bg-card p-4 space-y-3"
              >
                <div class="flex items-center justify-between">
                  <div class="flex items-center gap-2">
                    <p class="text-sm font-medium">{{ env.name }}</p>
                    <CopyKey :value="env.key" />
                  </div>
                  <Button
                    size="sm"
                    variant="ghost"
                    class="h-7 px-2 text-xs"
                    @click="openCreateSDKKey(env)"
                  >
                    <Plus class="size-3 mr-1" />{{ $t('keys.add') }}
                  </Button>
                </div>

                <div v-if="!sdkKeys[env.id]" class="text-xs text-muted-foreground italic">
                  {{ $t('keys.loading') }}
                </div>
                <div
                  v-else-if="sdkKeys[env.id].length === 0"
                  class="text-xs text-muted-foreground italic"
                >
                  {{ $t('keys.noKeys') }}
                </div>
                <div v-else class="space-y-1.5">
                  <div v-for="k in sdkKeys[env.id]" :key="k.id" class="flex items-center gap-2">
                    <div
                      class="flex-1 min-w-0 rounded-md border bg-muted/40 px-3 py-1.5 flex items-center gap-3"
                    >
                      <span class="text-xs font-medium truncate">{{ k.label }}</span>
                      <span class="font-mono text-[11px] text-muted-foreground"
                        >{{ k.key_prefix }}...</span
                      >
                      <span class="ml-auto text-[10px] text-muted-foreground shrink-0">
                        {{ k.expires_at ? formatDate(k.expires_at) : $t('keys.expiryNever') }}
                      </span>
                    </div>
                    <button
                      class="rounded-md p-1.5 text-muted-foreground transition-colors hover:bg-destructive/10 hover:text-destructive"
                      @click="deleteSDKKey(env, k)"
                    >
                      <Trash2 class="size-3.5" />
                    </button>
                  </div>
                </div>
              </div>
            </div>
          </TabsContent>

          <!-- Usage guide tab -->
          <TabsContent value="usage" class="space-y-6 text-sm">
            <!-- SDK Keys -->
            <div class="space-y-3">
              <div class="space-y-0.5">
                <h2 class="text-sm font-semibold">{{ $t('keys.usageSdkTitle') }}</h2>
                <p class="text-xs text-muted-foreground">{{ $t('keys.usageSdkDescription') }}</p>
              </div>

              <div class="space-y-2">
                <p class="text-xs font-medium text-muted-foreground uppercase tracking-wide">
                  {{ $t('keys.usagePollFlags') }}
                </p>
                <pre
                  class="rounded-lg bg-muted px-4 py-3 text-[12px] font-mono overflow-x-auto leading-relaxed"
                >
GET /sdk/flags?sdk_key=YOUR_SDK_KEY</pre
                >
              </div>

              <div class="space-y-2">
                <p class="text-xs font-medium text-muted-foreground uppercase tracking-wide">
                  {{ $t('keys.usageEvaluate') }}
                </p>
                <pre
                  class="rounded-lg bg-muted px-4 py-3 text-[12px] font-mono overflow-x-auto leading-relaxed"
                >
POST /sdk/evaluate?sdk_key=YOUR_SDK_KEY
Content-Type: application/json

{
  "flag_key": "my-feature",
  "user_key": "user-123",
  "attributes": {
    "plan": "pro",
    "email": "alice@example.com"
  }
}</pre
                >
              </div>

              <div class="space-y-2">
                <p class="text-xs font-medium text-muted-foreground uppercase tracking-wide">
                  {{ $t('keys.usageStream') }}
                </p>
                <pre
                  class="rounded-lg bg-muted px-4 py-3 text-[12px] font-mono overflow-x-auto leading-relaxed"
                >
GET /sdk/stream?sdk_key=YOUR_SDK_KEY
Accept: text/event-stream</pre
                >
                <p class="text-xs text-muted-foreground">{{ $t('keys.usageStreamNote') }}</p>
              </div>
            </div>

            <Separator />

            <!-- API Keys -->
            <div class="space-y-3">
              <div class="space-y-0.5">
                <h2 class="text-sm font-semibold">{{ $t('keys.usageApiTitle') }}</h2>
                <p class="text-xs text-muted-foreground">{{ $t('keys.usageApiDescription') }}</p>
              </div>

              <div class="space-y-2">
                <p class="text-xs font-medium text-muted-foreground uppercase tracking-wide">
                  {{ $t('keys.usageListFlags') }}
                </p>
                <pre
                  class="rounded-lg bg-muted px-4 py-3 text-[12px] font-mono overflow-x-auto leading-relaxed"
                >
GET /api/projects/{{ projectStore.current!.id }}/flags
Authorization: Bearer YOUR_API_KEY</pre
                >
              </div>

              <div class="space-y-2">
                <p class="text-xs font-medium text-muted-foreground uppercase tracking-wide">
                  {{ $t('keys.usageToggleFlag') }}
                </p>
                <pre
                  class="rounded-lg bg-muted px-4 py-3 text-[12px] font-mono overflow-x-auto leading-relaxed"
                >
PATCH /api/projects/{{ projectStore.current!.id }}/flags/FLAG_KEY/env
Authorization: Bearer YOUR_API_KEY
Content-Type: application/json

{
  "environment_id": 1,
  "enabled": true,
  "default_variation": 0
}</pre
                >
              </div>

              <div class="space-y-2">
                <p class="text-xs font-medium text-muted-foreground uppercase tracking-wide">
                  {{ $t('keys.usageCreateFlag') }}
                </p>
                <pre
                  class="rounded-lg bg-muted px-4 py-3 text-[12px] font-mono overflow-x-auto leading-relaxed"
                >
POST /api/projects/{{ projectStore.current!.id }}/flags
Authorization: Bearer YOUR_API_KEY
Content-Type: application/json

{
  "name": "My Feature",
  "key": "my-feature",
  "flag_type": "boolean",
  "variations": [
    { "name": "on",  "value": true },
    { "name": "off", "value": false }
  ]
}</pre
                >
              </div>
            </div>
          </TabsContent>

          <!-- API Keys tab -->
          <TabsContent value="api" class="space-y-3">
            <div class="flex items-center justify-between">
              <p class="text-xs text-muted-foreground">{{ $t('keys.apiSectionDescription') }}</p>
              <Button
                size="sm"
                variant="ghost"
                class="h-7 px-2 text-xs"
                @click="createApiOpen = true"
              >
                <Plus class="size-3 mr-1" />{{ $t('keys.add') }}
              </Button>
            </div>

            <div v-if="apiKeys.length === 0" class="text-sm text-muted-foreground italic">
              {{ $t('keys.noKeys') }}
            </div>
            <div v-else class="space-y-1.5">
              <div v-for="k in apiKeys" :key="k.id" class="flex items-center gap-2">
                <div
                  class="flex-1 min-w-0 rounded-md border bg-muted/40 px-3 py-1.5 flex items-center gap-3"
                >
                  <span class="text-xs font-medium truncate">{{ k.label }}</span>
                  <span class="font-mono text-[11px] text-muted-foreground"
                    >{{ k.key_prefix }}...</span
                  >
                  <span class="ml-auto text-[10px] text-muted-foreground shrink-0">
                    {{ k.expires_at ? formatDate(k.expires_at) : $t('keys.expiryNever') }}
                  </span>
                </div>
                <button
                  class="rounded-md p-1.5 text-muted-foreground transition-colors hover:bg-destructive/10 hover:text-destructive"
                  @click="deleteAPIKey(k)"
                >
                  <Trash2 class="size-3.5" />
                </button>
              </div>
            </div>
          </TabsContent>
        </Tabs>
      </div>
    </template>
  </div>

  <CreateKeyDialog
    v-if="projectStore.current && createSdkEnv"
    v-model:open="createSdkOpen"
    :title="$t('keys.createSdkTitle')"
    :description="$t('keys.createSdkDescription')"
    :on-create="
      (label, expiresAt) =>
        environmentsApi.sdkKeys
          .create(projectStore.current!.id, createSdkEnv!.id, label, expiresAt)
          .then((r) => r.key)
    "
    @created="onSDKKeyCreated"
  />
  <CreateKeyDialog
    v-if="projectStore.current"
    v-model:open="createApiOpen"
    :title="$t('keys.createApiTitle')"
    :description="$t('keys.createApiDescription')"
    :on-create="
      (label, expiresAt) =>
        environmentsApi.apiKeys
          .create(projectStore.current!.id, label, expiresAt)
          .then((r) => r.key)
    "
    @created="loadAPIKeys"
  />
</template>

<script setup lang="ts">
import { ref, watch, reactive } from 'vue'
import { FolderOpen, Plus, Loader2, Trash2 } from '@lucide/vue'
import { Button } from '@/components/ui/button'
import { Separator } from '@/components/ui/separator'
import { Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui/tabs'
import { useProjectStore } from '@/stores/project'
import {
  environmentsApi,
  type Environment,
  type SDKKeyRecord,
  type APIKeyRecord,
} from '@/api/environments'
import CreateKeyDialog from '@/components/CreateKeyDialog.vue'
import CopyKey from '@/components/CopyKey.vue'

const projectStore = useProjectStore()
const environments = ref<Environment[]>([])
const sdkKeys = reactive<Record<number, SDKKeyRecord[]>>({})
const apiKeys = ref<APIKeyRecord[]>([])
const loading = ref(false)
const createSdkOpen = ref(false)
const createSdkEnv = ref<Environment | null>(null)
const createApiOpen = ref(false)

async function load() {
  if (!projectStore.current) return
  loading.value = true
  try {
    const pid = projectStore.current.id
    const [envsResult, keys] = await Promise.all([
      environmentsApi.list(pid),
      environmentsApi.apiKeys.list(pid),
    ])
    environments.value = envsResult.data ?? []
    apiKeys.value = (keys ?? []) as APIKeyRecord[]
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

async function loadAPIKeys() {
  if (!projectStore.current) return
  apiKeys.value = ((await environmentsApi.apiKeys.list(projectStore.current.id)) ??
    []) as APIKeyRecord[]
}

watch(() => projectStore.current, load, { immediate: true })

function openCreateSDKKey(env: Environment) {
  createSdkEnv.value = env
  createSdkOpen.value = true
}

function onSDKKeyCreated() {
  if (!projectStore.current || !createSdkEnv.value) return
  loadSDKKeys(projectStore.current.id, createSdkEnv.value.id)
}

async function deleteSDKKey(env: Environment, key: SDKKeyRecord) {
  if (!projectStore.current) return
  await environmentsApi.sdkKeys.delete(projectStore.current.id, env.id, key.id)
  sdkKeys[env.id] = sdkKeys[env.id].filter((k) => k.id !== key.id)
}

async function deleteAPIKey(key: APIKeyRecord) {
  if (!projectStore.current) return
  await environmentsApi.apiKeys.delete(projectStore.current.id, key.id)
  apiKeys.value = apiKeys.value.filter((k) => k.id !== key.id)
}

function formatDate(iso: string) {
  return new Date(iso).toLocaleDateString(undefined, {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  })
}
</script>
