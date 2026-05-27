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
            <DialogTitle>{{ $t('segments.deleteTitle') }}</DialogTitle>
            <DialogDescription class="mt-0.5">{{ $t('common.cannotUndo') }}</DialogDescription>
          </div>
        </div>
      </DialogHeader>

      <p class="text-sm text-muted-foreground">
        {{ $t('segments.deleteWarning') }}
        <span class="font-medium text-foreground">{{ segment?.name }}</span>
        {{ $t('segments.deleteWarningEnd') }}
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
              ? $t('segments.deleting')
              : countdown > 0
                ? $t('segments.deleteCountdown', { countdown })
                : $t('segments.deleteConfirm')
          }}
        </Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import { ref, watch, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
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
import { segmentsApi, type Segment } from '@/api/segments'

const { t } = useI18n()
const props = defineProps<{ open: boolean; segment: Segment | null; projectId: number }>()
const emit = defineEmits<{
  'update:open': [value: boolean]
  deleted: [segment: Segment]
}>()

const loading = ref(false)
const error = ref('')
const countdown = ref(0)
let timer: ReturnType<typeof setInterval> | null = null

watch(
  () => props.open,
  (v) => {
    if (v) {
      error.value = ''
      countdown.value = 5
      timer = setInterval(() => {
        countdown.value--
        if (countdown.value <= 0) {
          clearInterval(timer!)
          timer = null
        }
      }, 1000)
    } else {
      if (timer) {
        clearInterval(timer)
        timer = null
      }
    }
  }
)

onUnmounted(() => {
  if (timer) clearInterval(timer)
})

async function submit() {
  if (!props.segment) return
  error.value = ''
  loading.value = true
  try {
    await segmentsApi.delete(props.projectId, props.segment.id)
    emit('deleted', props.segment)
    emit('update:open', false)
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : t('common.error')
  } finally {
    loading.value = false
  }
}
</script>
