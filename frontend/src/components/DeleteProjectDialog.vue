<template>
  <Dialog :open="open" @update:open="$emit('update:open', $event)">
    <DialogContent class="sm:max-w-md">
      <DialogHeader>
        <div class="flex items-center gap-3">
          <div
            class="flex size-9 shrink-0 items-center justify-center rounded-full bg-destructive/10"
          >
            <TriangleAlert class="size-4 text-destructive" />
          </div>
          <div>
            <DialogTitle>{{ $t('projects.deleteTitle') }}</DialogTitle>
            <DialogDescription class="mt-0.5">{{
              $t('projects.deleteSubtitle')
            }}</DialogDescription>
          </div>
        </div>
      </DialogHeader>

      <div class="space-y-4">
        <p class="text-sm text-muted-foreground">
          {{ $t('projects.deleteWarning') }}
          <span class="font-medium text-foreground">{{ project?.name }}</span>
          {{ $t('projects.deleteWarningEnd') }}
        </p>

        <div
          class="rounded-md border border-destructive/30 bg-destructive/5 px-3 py-2.5 text-xs text-destructive"
        >
          {{ $t('projects.deleteConsequences') }}
        </div>

        <div class="space-y-2">
          <Label for="delete-confirm-input">
            {{ $t('projects.deleteTypePrompt') }}
            <span class="font-mono font-medium text-foreground">{{ project?.name }}</span>
          </Label>
          <Input
            id="delete-confirm-input"
            v-model="confirmation"
            :placeholder="project?.name"
            class="mt-2"
            autofocus
            @keydown.enter="submit"
          />
        </div>

        <Alert v-if="error" variant="destructive">
          <AlertCircle class="size-4" />
          <AlertDescription>{{ error }}</AlertDescription>
        </Alert>
      </div>

      <DialogFooter>
        <Button type="button" variant="outline" @click="$emit('update:open', false)">
          {{ $t('common.cancel') }}
        </Button>
        <Button
          variant="destructive"
          :disabled="confirmation !== project?.name || loading"
          @click="submit"
        >
          <Loader2 v-if="loading" class="size-4 animate-spin" />
          {{ loading ? $t('projects.deleting') : $t('projects.deleteConfirmButton') }}
        </Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { TriangleAlert, AlertCircle, Loader2 } from '@lucide/vue'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Button } from '@/components/ui/button'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { projectsApi, type Project } from '@/api/projects'

const props = defineProps<{ open: boolean; project: Project | null }>()
const emit = defineEmits<{
  'update:open': [value: boolean]
  deleted: [project: Project]
}>()

const confirmation = ref('')
const loading = ref(false)
const error = ref('')

watch(
  () => props.open,
  (v) => {
    if (v) {
      confirmation.value = ''
      error.value = ''
    }
  }
)

async function submit() {
  if (!props.project || confirmation.value !== props.project.name) return
  error.value = ''
  loading.value = true
  try {
    await projectsApi.delete(props.project.id)
    emit('deleted', props.project)
    emit('update:open', false)
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : 'Something went wrong'
  } finally {
    loading.value = false
  }
}
</script>
