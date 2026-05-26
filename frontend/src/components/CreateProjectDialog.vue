<template>
  <Dialog :open="open" @update:open="$emit('update:open', $event)">
    <DialogContent class="sm:max-w-sm">
      <DialogHeader>
        <DialogTitle>{{ $t('projects.createTitle') }}</DialogTitle>
        <DialogDescription>{{ $t('projects.createDescription') }}</DialogDescription>
      </DialogHeader>

      <form class="space-y-3" @submit.prevent="submit">
        <div class="space-y-2">
          <Label for="project-name">{{ $t('projects.name') }}</Label>
          <Input
            id="project-name"
            v-model="name"
            :placeholder="$t('projects.namePlaceholder')"
            class="mt-2"
            required
            autofocus
          />
        </div>

        <div class="space-y-2">
          <Label for="project-key">{{ $t('projects.key') }}</Label>
          <Input
            id="project-key"
            v-model="key"
            placeholder="my-project"
            class="mt-2 font-mono"
            required
            @focus="keyTouched = true"
          />
        </div>

        <div class="space-y-2">
          <Label for="project-description">{{ $t('common.description') }}</Label>
          <Input
            id="project-description"
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
            {{ loading ? $t('projects.creating') : $t('projects.create') }}
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

const props = defineProps<{ open: boolean }>()
const emit = defineEmits<{
  'update:open': [value: boolean]
  created: [project: Project]
}>()

const name = ref('')
const keyRaw = ref('')
const keyTouched = ref(false)
const description = ref('')
const loading = ref(false)
const error = ref('')

// Computed setter sanitizes every keystroke: lowercase, a-z 0-9 hyphens only
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

// Auto-populate slug from name until user touches the slug field
watch(name, (v: string) => {
  if (!keyTouched.value) keyRaw.value = slugify(v)
})

watch(
  () => props.open,
  (v: boolean) => {
    if (v) {
      name.value = ''
      keyRaw.value = ''
      keyTouched.value = false
      description.value = ''
      error.value = ''
    }
  }
)

async function submit() {
  error.value = ''
  loading.value = true
  try {
    const project = await projectsApi.create(name.value, key.value, description.value)
    emit('created', project)
    emit('update:open', false)
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : 'Something went wrong'
  } finally {
    loading.value = false
  }
}
</script>
