<template>
  <Dialog :open="open" @update:open="$emit('update:open', $event)">
    <DialogContent class="sm:max-w-sm">
      <DialogHeader>
        <div class="flex items-center gap-3">
          <div
            class="flex size-9 shrink-0 items-center justify-center rounded-full bg-destructive/10"
          >
            <TriangleAlert class="size-4 text-destructive" />
          </div>
          <div>
            <DialogTitle>Delete environment</DialogTitle>
            <DialogDescription class="mt-0.5">This action cannot be undone</DialogDescription>
          </div>
        </div>
      </DialogHeader>

      <p class="text-sm text-muted-foreground">
        This will permanently delete
        <span class="font-medium text-foreground">{{ environment?.name }}</span> and remove all flag
        states associated with it.
      </p>

      <Alert v-if="error" variant="destructive">
        <AlertCircle class="size-4" />
        <AlertDescription>{{ error }}</AlertDescription>
      </Alert>

      <DialogFooter>
        <Button type="button" variant="outline" @click="$emit('update:open', false)">
          {{ $t('common.cancel') }}
        </Button>
        <Button variant="destructive" :disabled="countdown > 0 || loading" @click="submit">
          <Loader2 v-if="loading" class="size-4 animate-spin" />
          {{
            loading
              ? 'Deleting...'
              : countdown > 0
                ? `Delete environment (${countdown}s)`
                : 'Delete environment'
          }}
        </Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import { ref, watch, onUnmounted } from 'vue'
import { TriangleAlert, AlertCircle, Loader2 } from '@lucide/vue'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { environmentsApi, type Environment } from '@/api/environments'

const props = defineProps<{ open: boolean; environment: Environment | null; projectId: number }>()
const emit = defineEmits<{
  'update:open': [value: boolean]
  deleted: [environment: Environment]
}>()

const loading = ref(false)
const error = ref('')
const countdown = ref(0)
let countdownTimer: ReturnType<typeof setInterval> | null = null

watch(
  () => props.open,
  (v) => {
    if (v) {
      error.value = ''
      countdown.value = 10
      countdownTimer = setInterval(() => {
        countdown.value--
        if (countdown.value <= 0) {
          clearInterval(countdownTimer!)
          countdownTimer = null
        }
      }, 1000)
    } else {
      if (countdownTimer) {
        clearInterval(countdownTimer)
        countdownTimer = null
      }
    }
  }
)

onUnmounted(() => {
  if (countdownTimer) clearInterval(countdownTimer)
})

async function submit() {
  if (!props.environment) return
  error.value = ''
  loading.value = true
  try {
    await environmentsApi.delete(props.projectId, props.environment.id)
    emit('deleted', props.environment)
    emit('update:open', false)
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : 'Something went wrong'
  } finally {
    loading.value = false
  }
}
</script>
