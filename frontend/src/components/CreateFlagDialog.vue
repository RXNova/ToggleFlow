<template>
  <Dialog :open="open" @update:open="$emit('update:open', $event)">
    <DialogContent class="sm:max-w-md">
      <DialogHeader>
        <DialogTitle>{{ $t('flags.createTitle') }}</DialogTitle>
        <DialogDescription>{{ $t('flags.createDescription') }}</DialogDescription>
      </DialogHeader>

      <form class="space-y-4" @submit.prevent="submit">
        <!-- Name + key -->
        <div class="space-y-2">
          <Label for="flag-name">{{ $t('flags.name') }}</Label>
          <Input
            id="flag-name"
            v-model="name"
            :placeholder="$t('flags.namePlaceholder')"
            class="mt-2"
            required
            autofocus
            @input="syncKey"
          />
        </div>

        <div class="space-y-2">
          <Label for="flag-key">{{ $t('flags.key') }}</Label>
          <Input
            id="flag-key"
            v-model="key"
            :placeholder="$t('flags.keyPlaceholder')"
            class="mt-2 font-mono"
            required
            @input="keyTouched = true"
          />
        </div>

        <div class="space-y-2">
          <Label for="flag-description">{{ $t('flags.description') }}</Label>
          <Input
            id="flag-description"
            v-model="description"
            :placeholder="$t('flags.descriptionPlaceholder')"
            class="mt-2"
          />
        </div>

        <!-- Flag type selector — like a radio group but styled as pills -->
        <div class="space-y-2">
          <Label>{{ $t('flags.flagType') }}</Label>
          <div class="flex gap-1.5">
            <button
              v-for="t in FLAG_TYPES"
              :key="t"
              type="button"
              class="flex-1 rounded-md border px-3 py-1.5 text-xs font-medium transition-colors"
              :class="
                flagType === t
                  ? 'border-primary bg-primary text-primary-foreground'
                  : 'border-input bg-background text-muted-foreground hover:text-foreground'
              "
              @click="setType(t)"
            >
              {{ $t(`flags.types.${t}`) }}
            </button>
          </div>
        </div>

        <!-- Variations -->
        <div class="space-y-2">
          <Label>{{ $t('flags.variations') }}</Label>

          <div class="space-y-2">
            <div v-for="(v, i) in variations" :key="i" class="flex items-center gap-2">
              <Input v-model="v.name" :placeholder="$t('flags.variationName')" class="flex-1" />

              <!-- Boolean: show locked true/false badge instead of input -->
              <template v-if="flagType === 'boolean'">
                <span
                  class="shrink-0 rounded border px-2 py-1 font-mono text-xs text-muted-foreground"
                  :class="
                    i === 0
                      ? 'border-emerald-500/40 bg-emerald-500/10 text-emerald-600 dark:text-emerald-400'
                      : 'border-rose-500/40 bg-rose-500/10 text-rose-600 dark:text-rose-400'
                  "
                >
                  {{ i === 0 ? 'true' : 'false' }}
                </span>
              </template>

              <!-- String / number / json: editable value field -->
              <template v-else>
                <Input
                  v-model="v.value"
                  :placeholder="$t('flags.variationValue')"
                  :type="flagType === 'number' ? 'number' : 'text'"
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
            v-if="flagType !== 'boolean'"
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
            {{ loading ? $t('flags.creating') : $t('flags.create') }}
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
import { flagsApi, type Flag, type FlagType, type Variation } from '@/api/flags'

const FLAG_TYPES: FlagType[] = ['boolean', 'string', 'number', 'json']

const props = defineProps<{ open: boolean; projectId: number }>()
const emit = defineEmits<{
  'update:open': [value: boolean]
  created: [flag: Flag]
}>()

const name = ref('')
const key = ref('')
const description = ref('')
const flagType = ref<FlagType>('boolean')
const loading = ref(false)
const error = ref('')
let keyTouched = false

// Each variation in the form is {name, value} where value is always a string
// so all inputs are text/number fields. We coerce on submit.
interface VariationForm {
  name: string
  value: string
}
const variations = ref<VariationForm[]>(defaultVariationsFor('boolean'))

function defaultVariationsFor(t: FlagType): VariationForm[] {
  if (t === 'boolean')
    return [
      { name: 'Enabled', value: 'true' },
      { name: 'Disabled', value: 'false' },
    ]
  if (t === 'number')
    return [
      { name: 'Variation A', value: '0' },
      { name: 'Variation B', value: '1' },
    ]
  return [
    { name: 'Variation A', value: '' },
    { name: 'Variation B', value: '' },
  ]
}

function setType(t: FlagType) {
  flagType.value = t
  variations.value = defaultVariationsFor(t)
}

function addVariation() {
  variations.value.push({
    name: `Variation ${String.fromCharCode(65 + variations.value.length)}`,
    value: '',
  })
}

function removeVariation(i: number) {
  variations.value.splice(i, 1)
}

watch(
  () => props.open,
  (v) => {
    if (v) {
      name.value = ''
      key.value = ''
      description.value = ''
      flagType.value = 'boolean'
      variations.value = defaultVariationsFor('boolean')
      error.value = ''
      keyTouched = false
    }
  }
)

function slugify(s: string) {
  return s
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-|-$/g, '')
}

function syncKey() {
  if (!keyTouched) key.value = slugify(name.value)
}

// Coerce form string values into the right JS type before sending to the API.
// Think of it like a DTO transform in NestJS — the form always works with strings,
// but the API needs the actual typed value.
function coerceVariations(): Variation[] {
  return variations.value.map((v, i) => {
    let value: Variation['value']
    if (flagType.value === 'boolean') {
      value = i === 0
    } else if (flagType.value === 'number') {
      value = Number(v.value)
    } else if (flagType.value === 'json') {
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
  error.value = ''
  loading.value = true
  try {
    const flag = await flagsApi.create(props.projectId, {
      name: name.value,
      key: key.value,
      description: description.value || undefined,
      flag_type: flagType.value,
      variations: coerceVariations(),
    })
    emit('created', flag)
    emit('update:open', false)
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : 'Something went wrong'
  } finally {
    loading.value = false
  }
}
</script>
