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
      <!-- Header -->
      <div class="shrink-0 space-y-4 px-6 pt-6 pb-4">
        <div class="flex items-center justify-between">
          <div class="space-y-0.5">
            <h1 class="text-base font-semibold">{{ $t('segments.title') }}</h1>
            <p class="text-sm text-muted-foreground">{{ $t('segments.subtitle') }}</p>
          </div>
          <Button size="sm" @click="createOpen = true">
            <Plus class="size-3.5" />
            {{ $t('segments.create') }}
          </Button>
        </div>
        <Separator />
      </div>

      <!-- Body -->
      <div class="flex min-h-0 flex-1 flex-col overflow-hidden px-6 pb-6">
        <div v-if="loading" class="flex flex-1 items-center justify-center">
          <Loader2 class="size-6 animate-spin text-muted-foreground/40" />
        </div>

        <div
          v-else-if="segments.length === 0"
          class="flex flex-1 flex-col items-center justify-center text-center"
        >
          <Users class="size-8 text-muted-foreground/30 mb-3" />
          <p class="text-sm font-medium">{{ $t('segments.emptyTitle') }}</p>
          <p class="mt-1 text-xs text-muted-foreground max-w-xs">
            {{ $t('segments.emptyDescription') }}
          </p>
        </div>

        <div v-else class="section-segments min-h-0 flex-1 overflow-y-auto space-y-2">
          <div v-for="seg in segments" :key="seg.id" class="rounded-lg border bg-card p-4">
            <div class="flex items-start justify-between gap-4">
              <div class="min-w-0 flex-1">
                <div class="flex items-center gap-2">
                  <p class="text-sm font-medium leading-none">{{ seg.name }}</p>
                  <CopyKey :value="seg.key" />
                </div>
                <div class="mt-2 flex flex-wrap gap-1">
                  <span v-if="seg.values.length === 0" class="text-xs text-muted-foreground">{{
                    $t('segments.noValues')
                  }}</span>
                  <span
                    v-for="(val, vi) in seg.values.slice(0, 12)"
                    :key="vi"
                    class="rounded bg-muted px-1.5 py-0.5 text-[11px] font-mono"
                    >{{ val }}</span
                  >
                  <span
                    v-if="seg.values.length > 12"
                    class="rounded bg-muted px-1.5 py-0.5 text-[11px] text-muted-foreground"
                    >+{{ seg.values.length - 12 }} more</span
                  >
                </div>
              </div>
              <div class="flex items-center gap-1 shrink-0">
                <Tooltip :text="$t('common.edit')">
                  <button
                    class="rounded-md p-1.5 text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
                    @click="openEdit(seg)"
                  >
                    <Pencil class="size-3.5" />
                  </button>
                </Tooltip>
                <Tooltip :text="$t('common.delete')">
                  <button
                    class="rounded-md p-1.5 text-muted-foreground transition-colors hover:bg-destructive/10 hover:text-destructive"
                    @click="openDelete(seg)"
                  >
                    <Trash2 class="size-3.5" />
                  </button>
                </Tooltip>
              </div>
            </div>
            <div class="mt-3 border-t pt-3 flex items-center gap-4 text-xs text-muted-foreground">
              <span class="flex items-center gap-1">
                <Hash class="size-3 shrink-0" />
                {{ $t('segments.valueCount', { n: seg.values.length }) }}
              </span>
              <span class="flex items-center gap-1">
                <Clock class="size-3 shrink-0" />
                {{ timeAgo(seg.updated_at) }}
              </span>
            </div>
          </div>
        </div>
      </div>
    </template>
  </div>

  <CreateSegmentDialog
    v-if="projectStore.current"
    v-model:open="createOpen"
    :project-id="projectStore.current.id"
    @created="onCreated"
  />
  <EditSegmentDialog
    v-if="projectStore.current"
    v-model:open="editOpen"
    :project-id="projectStore.current.id"
    :segment="editTarget"
    @updated="onUpdated"
  />
  <DeleteSegmentDialog
    v-if="projectStore.current"
    v-model:open="deleteOpen"
    :project-id="projectStore.current.id"
    :segment="deleteTarget"
    @deleted="onDeleted"
  />
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { Plus, FolderOpen, Loader2, Pencil, Trash2, Users, Hash, Clock } from '@lucide/vue'
import { Button } from '@/components/ui/button'
import { Separator } from '@/components/ui/separator'
import { Tooltip } from '@/components/ui/tooltip'
import { useProjectStore } from '@/stores/project'
import { segmentsApi, type Segment } from '@/api/segments'
import { timeAgo } from '@/lib/utils'
import CopyKey from '@/components/CopyKey.vue'
import CreateSegmentDialog from '@/components/CreateSegmentDialog.vue'
import EditSegmentDialog from '@/components/EditSegmentDialog.vue'
import DeleteSegmentDialog from '@/components/DeleteSegmentDialog.vue'

const projectStore = useProjectStore()
const segments = ref<Segment[]>([])
const loading = ref(false)
const createOpen = ref(false)
const editOpen = ref(false)
const editTarget = ref<Segment | null>(null)
const deleteOpen = ref(false)
const deleteTarget = ref<Segment | null>(null)

async function load() {
  if (!projectStore.current) return
  loading.value = true
  try {
    segments.value = await segmentsApi.list(projectStore.current.id)
  } finally {
    loading.value = false
  }
}

watch(() => projectStore.current, load, { immediate: true })

function openEdit(seg: Segment) {
  editTarget.value = seg
  editOpen.value = true
}

function openDelete(seg: Segment) {
  deleteTarget.value = seg
  deleteOpen.value = true
}

function onCreated(seg: Segment) {
  segments.value.push(seg)
  segments.value.sort((a, b) => a.name.localeCompare(b.name))
}

function onUpdated(updated: Segment) {
  const i = segments.value.findIndex((s) => s.id === updated.id)
  if (i !== -1) segments.value[i] = updated
}

function onDeleted(seg: Segment) {
  segments.value = segments.value.filter((s) => s.id !== seg.id)
}
</script>
