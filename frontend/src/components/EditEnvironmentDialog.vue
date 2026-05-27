<template>
  <Dialog :open="open" @update:open="$emit('update:open', $event)">
    <DialogContent class="sm:max-w-sm">
      <DialogHeader>
        <DialogTitle>{{ $t('environments.editTitle') }}</DialogTitle>
        <DialogDescription>{{ $t('environments.editDescription') }}</DialogDescription>
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
          <Label for="edit-env-key">{{ $t('environments.key') }}</Label>
          <Input
            id="edit-env-key"
            v-model="key"
            placeholder="production"
            class="mt-2 font-mono"
            required
            @focus="keyTouched = true"
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

        <div class="flex items-center justify-between rounded-md border px-3 py-2.5">
          <div>
            <p class="text-sm font-medium">{{ $t('environments.protectedLabel') }}</p>
            <p class="text-xs text-muted-foreground">
              {{ $t('environments.protectedDescription') }}
            </p>
          </div>
          <button
            type="button"
            class="relative inline-flex h-4.5 w-8 shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors focus-visible:outline-none"
            :class="isProtected ? 'bg-primary' : 'bg-input'"
            @click="isProtected = !isProtected"
          >
            <span
              class="pointer-events-none block h-3.5 w-3.5 rounded-full bg-background shadow-sm ring-0 transition-transform"
              :class="isProtected ? 'translate-x-3' : 'translate-x-0'"
            />
          </button>
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
import { environmentsApi, type Environment } from '@/api/environments'
import { useAsyncAction } from '@/composables/useAsyncAction'

const props = defineProps<{ open: boolean; environment: Environment | null; projectId: number }>()
const emit = defineEmits<{
  'update:open': [value: boolean]
  updated: [environment: Environment]
}>()

const name = ref('')
const keyRaw = ref('')
const keyTouched = ref(false)
const description = ref('')
const isProtected = ref(false)
const { loading, error, run } = useAsyncAction()

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
  () => props.open,
  (v) => {
    if (v && props.environment) {
      name.value = props.environment.name
      keyRaw.value = props.environment.key
      keyTouched.value = false
      description.value = props.environment.description
      isProtected.value = props.environment.protected
      error.value = ''
    }
  }
)

async function submit() {
  if (!props.environment) return
  await run(async () => {
    const updated = await environmentsApi.update(
      props.projectId,
      props.environment!.id,
      name.value,
      key.value,
      description.value,
      isProtected.value
    )
    emit('updated', updated)
    emit('update:open', false)
  })
}
</script>
