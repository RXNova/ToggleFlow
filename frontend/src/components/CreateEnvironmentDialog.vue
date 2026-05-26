<template>
  <Dialog :open="open" @update:open="$emit('update:open', $event)">
    <DialogContent class="sm:max-w-sm">
      <DialogHeader>
        <DialogTitle>{{ $t('environments.createTitle') }}</DialogTitle>
        <DialogDescription>{{ $t('environments.createDescription') }}</DialogDescription>
      </DialogHeader>

      <form class="space-y-3" @submit.prevent="submit">
        <div class="space-y-2">
          <Label for="env-name">{{ $t('environments.name') }}</Label>
          <Input
            id="env-name"
            v-model="name"
            :placeholder="$t('environments.namePlaceholder')"
            class="mt-2"
            required
            autofocus
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
            {{ loading ? $t('environments.creating') : $t('environments.create') }}
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

const props = defineProps<{ open: boolean; projectId: number }>()
const emit = defineEmits<{
  'update:open': [value: boolean]
  created: [environment: Environment]
}>()

const name = ref('')
const loading = ref(false)
const error = ref('')

watch(
  () => props.open,
  (v) => {
    if (v) {
      name.value = ''
      error.value = ''
    }
  }
)

async function submit() {
  error.value = ''
  loading.value = true
  try {
    const env = await environmentsApi.create(props.projectId, name.value)
    emit('created', env)
    emit('update:open', false)
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : 'Something went wrong'
  } finally {
    loading.value = false
  }
}
</script>
