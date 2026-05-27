<template>
  <DialogRoot :open="open" @update:open="$emit('update:open', $event)">
    <DialogPortal>
      <DialogOverlay
        class="fixed inset-0 z-50 bg-black/40 data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0"
      />
      <DialogContent
        class="fixed right-0 top-0 z-50 flex h-full w-full max-w-md flex-col border-l bg-background shadow-xl duration-300 data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:slide-out-to-right data-[state=open]:slide-in-from-right focus:outline-none"
      >
        <!-- Header -->
        <div class="flex items-start justify-between border-b px-5 py-4">
          <div class="space-y-0.5">
            <DialogTitle class="text-sm font-semibold">{{ title }}</DialogTitle>
            <DialogDescription class="font-mono text-[11px] text-muted-foreground">
              {{ label }}
            </DialogDescription>
          </div>
          <DialogClose
            class="rounded-sm p-1 opacity-70 ring-offset-background transition-opacity hover:opacity-100 focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2"
          >
            <X class="size-4" />
            <span class="sr-only">Close</span>
          </DialogClose>
        </div>

        <!-- Body -->
        <div class="flex min-h-0 flex-1 flex-col overflow-hidden px-5 py-4">
          <div v-if="loading" class="flex flex-1 items-center justify-center">
            <Loader2 class="size-5 animate-spin text-muted-foreground/40" />
          </div>

          <div
            v-else-if="total === 0"
            class="flex flex-1 flex-col items-center justify-center text-center"
          >
            <ClipboardList class="size-8 text-muted-foreground/30 mb-3" />
            <p class="text-sm font-medium">{{ $t('audit.emptyTitle') }}</p>
            <p class="mt-1 text-xs text-muted-foreground max-w-xs">
              {{ $t('audit.emptyDescription') }}
            </p>
          </div>

          <template v-else>
            <div class="min-h-0 flex-1 overflow-y-auto space-y-2">
              <div
                v-for="entry in entries"
                :key="entry.id"
                class="rounded-lg border bg-card px-3 py-2.5"
              >
                <div class="flex items-start justify-between gap-3">
                  <div class="flex items-start gap-2.5 min-w-0">
                    <div
                      class="flex size-6 shrink-0 items-center justify-center rounded-full bg-muted text-[10px] font-semibold uppercase mt-0.5"
                    >
                      {{ entry.actor[0] }}
                    </div>

                    <div class="min-w-0">
                      <div class="flex flex-wrap items-center gap-1.5">
                        <span class="text-xs font-medium">{{ entry.actor }}</span>
                        <span
                          :class="actionBadgeClass(entry.action)"
                          class="inline-flex items-center rounded px-1.5 py-0.5 text-[10px] font-medium"
                        >
                          {{ actionLabel(entry.action) }}
                        </span>
                      </div>

                      <div v-if="diffLines(entry).length" class="mt-1 space-y-0.5">
                        <div
                          v-for="(line, i) in diffLines(entry)"
                          :key="i"
                          class="text-[11px] text-muted-foreground font-mono"
                        >
                          {{ line }}
                        </div>
                      </div>
                    </div>
                  </div>

                  <span class="shrink-0 text-[11px] text-muted-foreground whitespace-nowrap">
                    {{ timeAgo(entry.created_at) }}
                  </span>
                </div>
              </div>
            </div>

            <div class="shrink-0 pt-3">
              <Pagination :page="page" :total="total" :limit="limit" @change="goTo" />
            </div>
          </template>
        </div>
      </DialogContent>
    </DialogPortal>
  </DialogRoot>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { ClipboardList, Loader2, X } from '@lucide/vue'
import {
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogOverlay,
  DialogPortal,
  DialogRoot,
  DialogTitle,
} from 'reka-ui'
import Pagination from '@/components/ui/pagination/Pagination.vue'
import { auditApi, userAuditApi, type AuditEntry } from '@/api/audit'
import { timeAgo } from '@/lib/utils'

const props = defineProps<{
  open: boolean
  // projectId + resource/actor for flag and environment history
  projectId?: number
  resource?: string
  actor?: string
  // userId for user account history (uses a separate endpoint)
  userId?: number
  title: string
  label: string
}>()

defineEmits<{ 'update:open': [value: boolean] }>()

const LIMIT = 20

const entries = ref<AuditEntry[]>([])
const total = ref(0)
const page = ref(1)
const limit = ref(LIMIT)
const loading = ref(false)

async function load() {
  loading.value = true
  try {
    const offset = (page.value - 1) * limit.value
    const res =
      props.userId != null
        ? await userAuditApi.list(props.userId, { limit: limit.value, offset })
        : await auditApi.list(props.projectId!, {
            limit: limit.value,
            offset,
            resource: props.resource,
            actor: props.actor,
          })
    entries.value = res.data ?? []
    total.value = res.total
  } finally {
    loading.value = false
  }
}

watch(
  () => props.open,
  (isOpen) => {
    if (isOpen) {
      page.value = 1
      load()
    }
  }
)

function goTo(p: number) {
  page.value = p
  load()
}

function actionLabel(action: string): string {
  const labels: Record<string, string> = {
    'flag.created': 'created flag',
    'flag.updated': 'updated flag',
    'flag.toggled': 'toggled flag',
    'flag.deleted': 'deleted flag',
    'env.created': 'created env',
    'env.updated': 'updated env',
    'env.deleted': 'deleted env',
    'user.created': 'created user',
    'user.updated': 'updated user',
    'user.deleted': 'deleted user',
  }
  return labels[action] ?? action
}

function actionBadgeClass(action: string): string {
  if (action.endsWith('.created')) return 'bg-green-500/10 text-green-600 dark:text-green-400'
  if (action.endsWith('.deleted')) return 'bg-destructive/10 text-destructive'
  if (action === 'flag.toggled') return 'bg-purple-500/10 text-purple-600 dark:text-purple-400'
  return 'bg-blue-500/10 text-blue-600 dark:text-blue-400'
}

function diffLines(entry: AuditEntry): string[] {
  const lines: string[] = []
  try {
    const oldV = entry.old_value ? JSON.parse(entry.old_value) : null
    const newV = entry.new_value ? JSON.parse(entry.new_value) : null

    if (entry.action === 'flag.toggled' && newV) {
      lines.push(`${newV.env}: ${oldV?.enabled ?? false} → ${newV.enabled}`)
      return lines
    }

    if (!oldV && newV) {
      for (const [k, v] of Object.entries(newV)) {
        if (v !== '' && v !== null) lines.push(`${k}: ${v}`)
      }
      return lines
    }

    if (oldV && newV) {
      for (const k of Object.keys(newV)) {
        if (String(oldV[k]) !== String(newV[k])) {
          lines.push(`${k}: ${oldV[k]} → ${newV[k]}`)
        }
      }
    }
  } catch {
    // malformed JSON — show nothing
  }
  return lines
}
</script>
