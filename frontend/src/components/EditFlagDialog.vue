<template>
  <Dialog :open="open" @update:open="$emit('update:open', $event)">
    <DialogContent class="sm:max-w-md">
      <DialogHeader>
        <DialogTitle>Edit flag</DialogTitle>
        <DialogDescription>Update the flag name, description, or variations.</DialogDescription>
      </DialogHeader>

      <form class="space-y-4" @submit.prevent="submit">
        <div class="space-y-2">
          <Label for="edit-flag-name">{{ $t('flags.name') }}</Label>
          <Input
            id="edit-flag-name"
            v-model="name"
            :placeholder="$t('flags.namePlaceholder')"
            class="mt-2"
            required
            autofocus
          />
        </div>

        <div class="space-y-2">
          <Label for="edit-flag-key">{{ $t('flags.key') }}</Label>
          <Input id="edit-flag-key" :model-value="flag?.key" class="mt-2 font-mono" disabled />
        </div>

        <div class="space-y-2">
          <Label for="edit-flag-description">{{ $t('flags.description') }}</Label>
          <Input
            id="edit-flag-description"
            v-model="description"
            :placeholder="$t('flags.descriptionPlaceholder')"
            class="mt-2"
          />
        </div>

        <div class="space-y-2">
          <Label>{{ $t('flags.variations') }}</Label>
          <div class="space-y-2">
            <div v-for="(v, i) in variations" :key="i" class="flex items-center gap-2">
              <Input v-model="v.name" :placeholder="$t('flags.variationName')" class="flex-1" />
              <template v-if="flag?.flag_type === 'boolean'">
                <span
                  class="shrink-0 rounded border px-2 py-1 font-mono text-xs"
                  :class="
                    i === 0
                      ? 'border-emerald-500/40 bg-emerald-500/10 text-emerald-600 dark:text-emerald-400'
                      : 'border-rose-500/40 bg-rose-500/10 text-rose-600 dark:text-rose-400'
                  "
                >
                  {{ i === 0 ? 'true' : 'false' }}
                </span>
              </template>
              <template v-else>
                <Input
                  v-model="v.value"
                  :placeholder="$t('flags.variationValue')"
                  :type="flag?.flag_type === 'number' ? 'number' : 'text'"
                  class="flex-1 font-mono text-xs"
                />
                <button
                  v-if="variations.length > 2"
                  type="button"
                  class="shrink-0 rounded p-1 text-muted-foreground hover:text-destructive"
                  @click="removeVariation(i)"
                >
                  <X class="size-3.5" />
                </button>
                <div v-else class="size-6 shrink-0" />
              </template>
            </div>
          </div>
          <button
            v-if="flag?.flag_type !== 'boolean'"
            type="button"
            class="flex items-center gap-1 text-xs text-muted-foreground hover:text-foreground"
            @click="addVariation"
          >
            <Plus class="size-3.5" />
            {{ $t('flags.addVariation') }}
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
import { ref, watch } from 'vue'
import { AlertCircle, Loader2, Plus, X } from '@lucide/vue'
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
import { flagsApi, type Flag, type Variation } from '@/api/flags'

const props = defineProps<{ open: boolean; flag: Flag | null; projectId: number }>()
const emit = defineEmits<{
  'update:open': [value: boolean]
  updated: [flag: Flag]
}>()

const name = ref('')
const description = ref('')
const loading = ref(false)
const error = ref('')

interface VariationForm {
  name: string
  value: string
}
const variations = ref<VariationForm[]>([])

watch(
  () => props.open,
  (v) => {
    if (v && props.flag) {
      name.value = props.flag.name
      description.value = props.flag.description ?? ''
      variations.value = props.flag.variations.map((vr) => ({
        name: vr.name,
        value: typeof vr.value === 'object' ? JSON.stringify(vr.value) : String(vr.value),
      }))
      error.value = ''
    }
  }
)

function addVariation() {
  variations.value.push({
    name: `Variation ${String.fromCharCode(65 + variations.value.length)}`,
    value: '',
  })
}

function removeVariation(i: number) {
  variations.value.splice(i, 1)
}

function coerceVariations(): Variation[] {
  return variations.value.map((v, i) => {
    let value: Variation['value']
    if (props.flag?.flag_type === 'boolean') {
      value = i === 0
    } else if (props.flag?.flag_type === 'number') {
      value = Number(v.value)
    } else if (props.flag?.flag_type === 'json') {
      try {
        value = JSON.parse(v.value)
      } catch {
        value = v.value
      }
    } else {
      value = v.value
    }
    return { name: v.name, value }
  })
}

async function submit() {
  if (!props.flag) return
  error.value = ''
  loading.value = true
  try {
    const updated = await flagsApi.update(props.projectId, props.flag.key, {
      name: name.value,
      description: description.value,
      variations: coerceVariations(),
    })
    emit('updated', updated)
    emit('update:open', false)
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : 'Something went wrong'
  } finally {
    loading.value = false
  }
}
</script>
