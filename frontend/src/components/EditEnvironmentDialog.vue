<template>
  <Dialog :open="open" @update:open="$emit('update:open', $event)">
    <DialogContent class="sm:max-w-sm">
      <DialogHeader>
        <DialogTitle>Edit environment</DialogTitle>
        <DialogDescription>Update the environment name or description.</DialogDescription>
      </DialogHeader>

      <form class="space-y-3" @submit.prevent="submit">
        <div class="space-y-2">
          <Label for="edit-env-name">{{ $t('environments.name') }}</Label>
          <Input
            id="edit-env-name"
            v-model="name"
            :placeholder="$t('environments.namePlaceholder')"
            class="mt-2"
            required
            autofocus
          />
        </div>

        <div class="space-y-2">
          <Label for="edit-env-description">{{ $t('common.description') }}</Label>
          <Input
            id="edit-env-description"
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
import { environmentsApi, type Environment } from '@/api/environments'

const props = defineProps<{ open: boolean; environment: Environment | null; projectId: number }>()
const emit = defineEmits<{
  'update:open': [value: boolean]
  updated: [environment: Environment]
}>()

const name = ref('')
const description = ref('')
const loading = ref(false)
const error = ref('')

watch(
  () => props.open,
  (v) => {
    if (v && props.environment) {
      name.value = props.environment.name
      description.value = props.environment.description
      error.value = ''
    }
  }
)

async function submit() {
  if (!props.environment) return
  error.value = ''
  loading.value = true
  try {
    const updated = await environmentsApi.update(
      props.projectId,
      props.environment.id,
      name.value,
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
