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
          <Label for="edit-project-key">{{ $t('projects.key') }}</Label>
          <Input
            id="edit-project-key"
            v-model="key"
            placeholder="my-project"
            class="mt-2 font-mono"
            required
            @focus="keyTouched = true"
          />
        </div>

        <div class="space-y-2">
          <Label for="edit-project-description">{{ $t('common.description') }}</Label>
          <Input
            id="edit-project-description"
            v-model="description"
            :placeholder="$t('common.descriptionPlaceholder')"
            class="mt-2"
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
import { ref, computed, watch } from 'vue'
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
const keyRaw = ref('')
const keyTouched = ref(false)
const description = ref('')
const loading = ref(false)
const error = ref('')

const key = computed({
  get: () => keyRaw.value,
  set: (v: string) => {
    keyRaw.value = v.toLowerCase().replace(/[^a-z0-9-]/g, '')
  },
})

function slugify(s: string) {
  return s
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-|-$/g, '')
}

watch(name, (v: string) => {
  if (!keyTouched.value) keyRaw.value = slugify(v)
})

watch(
  () => props.project,
  (p) => {
    if (p) {
      name.value = p.name
      keyRaw.value = p.key
      description.value = p.description
      keyTouched.value = false
      error.value = ''
    }
  },
  { immediate: true }
)

watch(
  () => props.open,
  (v: boolean) => {
    if (v && props.project) {
      name.value = props.project.name
      keyRaw.value = props.project.key
      description.value = props.project.description
      keyTouched.value = false
      error.value = ''
    }
  }
)

async function submit() {
  if (!props.project) return
  error.value = ''
  loading.value = true
  try {
    const updated = await projectsApi.update(
      props.project.id,
      name.value,
      key.value,
      description.value
    )
    emit('updated', updated)
    emit('update:open', false)
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : 'Something went wrong'
  } finally {
    loading.value = false
  }
}
</script>
