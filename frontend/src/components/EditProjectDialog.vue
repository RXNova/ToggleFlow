<template>
  <Dialog :open="open" @update:open="$emit('update:open', $event)">
    <DialogContent class="sm:max-w-sm">
      <DialogHeader>
        <DialogTitle>{{ $t('projects.editTitle') }}</DialogTitle>
        <DialogDescription>{{ $t('projects.editDescription') }}</DialogDescription>
      </DialogHeader>

      <form class="space-y-3" @submit.prevent="submit">
        <div class="space-y-2">
          <Label for="edit-project-name">{{ $t('projects.name') }}</Label>
          <Input
            id="edit-project-name"
            v-model="name"
            :placeholder="$t('projects.namePlaceholder')"
            class="mt-2"
            required
            autofocus
          />
        </div>

        <div class="space-y-2">
          <Label for="edit-project-slug">{{ $t('projects.slug') }}</Label>
          <Input
            id="edit-project-slug"
            v-model="slug"
            placeholder="my-project"
            class="mt-2 font-mono"
            required
          />
        </div>

        <Alert v-if="error" variant="destructive">
          <AlertCircle class="size-4" />
          <AlertDescription>{{ error }}</AlertDescription>
        </Alert>

        <DialogFooter>
          <Button type="button" variant="outline" @click="$emit('update:open', false)">
            {{ $t('common.cancel') }}
          </Button>
          <Button type="submit" :disabled="loading">
            <Loader2 v-if="loading" class="size-4 animate-spin" />
            {{ loading ? $t('projects.saving') : $t('projects.save') }}
          </Button>
        </DialogFooter>
      </form>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { AlertCircle, Loader2 } from '@lucide/vue'
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
  updated: [project: Project]
}>()

const name = ref('')
const slug = ref('')
const loading = ref(false)
const error = ref('')

watch(
  () => props.project,
  (p) => {
    if (p) {
      name.value = p.name
      slug.value = p.slug
      error.value = ''
    }
  },
  { immediate: true }
)

watch(
  () => props.open,
  (v) => {
    if (v && props.project) {
      name.value = props.project.name
      slug.value = props.project.slug
      error.value = ''
    }
  }
)

async function submit() {
  if (!props.project) return
  error.value = ''
  loading.value = true
  try {
    const updated = await projectsApi.update(props.project.id, name.value, slug.value)
    emit('updated', updated)
    emit('update:open', false)
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : 'Something went wrong'
  } finally {
    loading.value = false
  }
}
</script>
