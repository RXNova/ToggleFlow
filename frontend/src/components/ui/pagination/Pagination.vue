<template>
  <div v-if="totalPages > 1" class="flex items-center justify-between text-sm">
    <p class="text-xs text-muted-foreground">
      {{ rangeStart }}–{{ rangeEnd }} of {{ total }}
    </p>
    <div class="flex items-center gap-1">
      <button
        class="rounded-md p-1.5 text-muted-foreground transition-colors hover:bg-accent hover:text-foreground disabled:pointer-events-none disabled:opacity-40"
        :disabled="page === 1"
        @click="$emit('change', page - 1)"
      >
        <ChevronLeft class="size-4" />
      </button>

      <template v-for="p in pages" :key="p">
        <span v-if="p === '...'" class="px-1 text-muted-foreground">…</span>
        <button
          v-else
          class="min-w-[2rem] rounded-md px-2 py-1 text-xs transition-colors"
          :class="p === page
            ? 'bg-primary text-primary-foreground font-medium'
            : 'text-muted-foreground hover:bg-accent hover:text-foreground'"
          @click="$emit('change', p as number)"
        >
          {{ p }}
        </button>
      </template>

      <button
        class="rounded-md p-1.5 text-muted-foreground transition-colors hover:bg-accent hover:text-foreground disabled:pointer-events-none disabled:opacity-40"
        :disabled="page === totalPages"
        @click="$emit('change', page + 1)"
      >
        <ChevronRight class="size-4" />
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { ChevronLeft, ChevronRight } from '@lucide/vue'

const props = defineProps<{ page: number; total: number; limit: number }>()
defineEmits<{ change: [page: number] }>()

const totalPages = computed(() => Math.ceil(props.total / props.limit))
const rangeStart = computed(() => (props.page - 1) * props.limit + 1)
const rangeEnd = computed(() => Math.min(props.page * props.limit, props.total))

// Show at most 7 page buttons; collapse middle pages with ellipsis
const pages = computed(() => {
  const total = totalPages.value
  const cur = props.page
  if (total <= 7) return Array.from({ length: total }, (_, i) => i + 1)

  const result: (number | '...')[] = [1]
  if (cur > 3) result.push('...')
  for (let i = Math.max(2, cur - 1); i <= Math.min(total - 1, cur + 1); i++) result.push(i)
  if (cur < total - 2) result.push('...')
  result.push(total)
  return result
})
</script>
