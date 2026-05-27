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
            <DialogTitle>{{ $t('environments.deleteTitle') }}</DialogTitle>
            <DialogDescription class="mt-0.5">{{ $t('common.cannotUndo') }}</DialogDescription>
          </div>
        </div>
      </DialogHeader>

      <p class="text-sm text-muted-foreground">
        {{ $t('environments.deleteWarning') }}
        <span class="font-medium text-foreground">{{ environment?.name }}</span>
        {{ $t('environments.deleteWarningEnd') }}
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
              ? $t('environments.deleting')
              : countdown > 0
                ? $t('environments.deleteConfirmCountdown', { countdown })
                : $t('environments.deleteConfirmButton')
          }}
        </Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import { watch } from 'vue'
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
import { useCountdown } from '@/composables/useCountdown'
import { useAsyncAction } from '@/composables/useAsyncAction'

const props = defineProps<{ open: boolean; environment: Environment | null; projectId: number }>()
const emit = defineEmits<{
  'update:open': [value: boolean]
  deleted: [environment: Environment]
}>()

const { countdown } = useCountdown(() => props.open, 10)
const { loading, error, run } = useAsyncAction()

watch(
  () => props.open,
  (v) => {
    if (v) error.value = ''
  }
)

async function submit() {
  if (!props.environment) return
  await run(async () => {
    await environmentsApi.delete(props.projectId, props.environment!.id)
    emit('deleted', props.environment!)
    emit('update:open', false)
  })
}
</script>
