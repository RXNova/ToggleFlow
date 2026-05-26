<template>
  <div class="p-6 space-y-4">
    <div class="flex items-center justify-between">
      <div class="space-y-0.5">
        <h1 class="text-base font-semibold">{{ $t('nav.projects') }}</h1>
        <p class="text-sm text-muted-foreground">{{ $t('projects.subtitle') }}</p>
      </div>
      <Button size="sm" @click="createDialogOpen = true">
        <Plus class="size-3.5" />
        {{ $t('projects.new') }}
      </Button>
    </div>

    <Separator />

    <div class="relative">
      <Search
        class="absolute left-2.5 top-1/2 -translate-y-1/2 size-3.5 text-muted-foreground pointer-events-none"
      />
      <Input v-model="search" :placeholder="$t('projects.search')" class="pl-8" />
    </div>

    <div v-if="loading" class="flex flex-col items-center justify-center py-24">
      <Loader2 class="size-6 animate-spin text-muted-foreground/40" />
    </div>

    <div
      v-else-if="projects.length === 0 && search"
      class="flex flex-col items-center justify-center py-24 text-center"
    >
      <SearchX class="size-8 text-muted-foreground/30 mb-3" />
      <p class="text-sm font-medium">{{ $t('projects.noResults') }}</p>
    </div>

    <div
      v-else-if="total === 0"
      class="flex flex-col items-center justify-center py-24 text-center"
    >
      <FolderOpen class="size-8 text-muted-foreground/30 mb-3" />
      <p class="text-sm font-medium">{{ $t('projects.emptyTitle') }}</p>
      <p class="mt-1 text-xs text-muted-foreground max-w-xs">
        {{ $t('projects.emptyDescription') }}
      </p>
    </div>

    <template v-else>
      <div class="space-y-2">
        <div v-for="project in projects" :key="project.id" class="rounded-lg border bg-card p-4">
          <div class="flex items-start justify-between gap-4">
            <div class="min-w-0 flex-1">
              <div class="flex flex-wrap items-center gap-2">
                <p class="text-sm font-medium leading-none">{{ project.name }}</p>
                <span
                  class="inline-flex items-center rounded bg-muted px-1.5 py-0.5 font-mono text-[11px] text-muted-foreground"
                  >{{ project.key }}</span
                >
              </div>
              <div
                class="mt-2 flex flex-wrap items-center gap-x-4 gap-y-1 text-xs text-muted-foreground"
              >
                <span class="flex items-center gap-1.5">
                  <span
                    class="flex size-4 shrink-0 items-center justify-center rounded-full bg-muted text-[10px] font-semibold uppercase"
                    >{{ project.created_by_name ? project.created_by_name[0] : '?' }}</span
                  >
                  {{ project.created_by_name || '—' }}
                </span>
                <span class="flex items-center gap-1">
                  <CalendarDays class="size-3 shrink-0" />
                  {{ timeAgo(project.created_at) }}
                </span>
                <span class="flex items-center gap-1">
                  <Clock class="size-3 shrink-0" />
                  {{ timeAgo(project.updated_at) }}
                </span>
              </div>
            </div>
            <div class="flex items-center gap-1 shrink-0">
              <button
                class="rounded-md p-1.5 text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
                @click="openEdit(project)"
              >
                <Pencil class="size-3.5" />
              </button>
              <button
                class="rounded-md p-1.5 text-muted-foreground transition-colors hover:bg-destructive/10 hover:text-destructive"
                @click="openDelete(project)"
              >
                <Trash2 class="size-3.5" />
              </button>
            </div>
          </div>
        </div>
      </div>

      <Pagination :page="page" :total="total" :limit="limit" @change="goTo" />
    </template>
  </div>

  <CreateProjectDialog v-model:open="createDialogOpen" @created="onCreated" />
  <EditProjectDialog v-model:open="editDialogOpen" :project="editTarget" @updated="onUpdated" />
  <DeleteProjectDialog
    v-model:open="deleteDialogOpen"
    :project="deleteTarget"
    @deleted="onDeleted"
  />
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { watchDebounced } from '@vueuse/core'
import {
  Plus,
  FolderOpen,
  Loader2,
  Pencil,
  Trash2,
  Search,
  SearchX,
  CalendarDays,
  Clock,
} from '@lucide/vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Separator } from '@/components/ui/separator'
import Pagination from '@/components/ui/pagination/Pagination.vue'
import { projectsApi, type Project } from '@/api/projects'
import { useProjectStore } from '@/stores/project'
import { timeAgo } from '@/lib/utils'
import CreateProjectDialog from '@/components/CreateProjectDialog.vue'
import EditProjectDialog from '@/components/EditProjectDialog.vue'
import DeleteProjectDialog from '@/components/DeleteProjectDialog.vue'

const LIMIT = 20

const projectStore = useProjectStore()
const projects = ref<Project[]>([])
const total = ref(0)
const page = ref(1)
const limit = ref(LIMIT)
const search = ref('')
const loading = ref(false)
const createDialogOpen = ref(false)
const editDialogOpen = ref(false)
const editTarget = ref<Project | null>(null)
const deleteDialogOpen = ref(false)
const deleteTarget = ref<Project | null>(null)

async function load() {
  loading.value = true
  try {
    const res = await projectsApi.list({
      limit: limit.value,
      offset: (page.value - 1) * limit.value,
      search: search.value || undefined,
    })
    projects.value = res.data
    total.value = res.total
  } finally {
    loading.value = false
  }
}

// Debounce search — like debounceTime(300) in RxJS
watchDebounced(
  search,
  () => {
    page.value = 1
    load()
  },
  { debounce: 300 }
)

// Reload when sidebar creates a project
watch(() => projectStore.projects.length, load)
load()

function goTo(p: number) {
  page.value = p
  load()
}

function openEdit(project: Project) {
  editTarget.value = project
  editDialogOpen.value = true
}

function openDelete(project: Project) {
  deleteTarget.value = project
  deleteDialogOpen.value = true
}

function onCreated(project: Project) {
  projectStore.projects.push(project)
  if (!projectStore.current) projectStore.setCurrent(project)
  load()
}

function onUpdated(updated: Project) {
  const si = projectStore.projects.findIndex((p) => p.id === updated.id)
  if (si !== -1) projectStore.projects[si] = updated
  if (projectStore.current?.id === updated.id) projectStore.setCurrent(updated)
  load()
}

function onDeleted(project: Project) {
  const remaining = projectStore.projects.filter((p) => p.id !== project.id)
  projectStore.setProjects(remaining)
  if (page.value > 1 && projects.value.length === 1) page.value--
  load()
}
</script>
