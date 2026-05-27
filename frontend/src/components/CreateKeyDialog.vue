<template>
  <Dialog :open="open" @update:open="$emit('update:open', $event)">
    <DialogContent class="sm:max-w-md">
      <DialogHeader>
        <DialogTitle>{{ title }}</DialogTitle>
        <DialogDescription>{{ description }}</DialogDescription>
      </DialogHeader>

      <!-- Phase 1: form -->
      <template v-if="phase === 'form'">
        <div class="space-y-4">
          <div class="space-y-1.5">
            <Label>{{ $t('keys.label') }}</Label>
            <Input v-model="label" :placeholder="$t('keys.labelPlaceholder')" />
          </div>

          <div class="space-y-1.5">
            <Label>{{ $t('keys.expiry') }}</Label>
            <select
              v-model="expiryPreset"
              class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
            >
              <option value="never">{{ $t('keys.expiryNever') }}</option>
              <option value="30d">{{ $t('keys.expiry30d') }}</option>
              <option value="90d">{{ $t('keys.expiry90d') }}</option>
              <option value="1y">{{ $t('keys.expiry1y') }}</option>
              <option value="custom">{{ $t('keys.expiryCustom') }}</option>
            </select>
            <Input
              v-if="expiryPreset === 'custom'"
              v-model="customDate"
              type="date"
              :min="minDate"
              class="mt-1.5"
            />
          </div>
        </div>

        <Alert v-if="error" variant="destructive" class="mt-2">
          <AlertCircle class="size-4" />
          <AlertDescription>{{ error }}</AlertDescription>
        </Alert>

        <DialogFooter>
          <Button variant="outline" @click="$emit('update:open', false)">
            {{ $t('common.cancel') }}
          </Button>
          <Button :disabled="!label.trim() || loading" @click="submit">
            <Loader2 v-if="loading" class="size-4 animate-spin" />
            {{ loading ? $t('keys.creating') : $t('keys.create') }}
          </Button>
        </DialogFooter>
      </template>

      <!-- Phase 2: show key once -->
      <template v-else>
        <Alert class="border-amber-500/30 bg-amber-500/10">
          <AlertCircle class="size-4 text-amber-600" />
          <AlertDescription class="text-amber-700 dark:text-amber-400">
            {{ $t('keys.copyNow') }}
          </AlertDescription>
        </Alert>

        <div class="space-y-1.5">
          <Label>{{ $t('keys.keyValue') }}</Label>
          <div class="flex items-center gap-2">
            <div class="flex-1 min-w-0 rounded-md border bg-muted/40 px-3 py-2">
              <p class="text-xs font-mono break-all">{{ createdKey }}</p>
            </div>
            <Button size="sm" variant="outline" class="shrink-0" @click="copy">
              <Check v-if="copied" class="size-3.5 text-green-500" />
              <Copy v-else class="size-3.5" />
            </Button>
          </div>
        </div>

        <DialogFooter>
          <Button @click="done">{{ $t('keys.done') }}</Button>
        </DialogFooter>
      </template>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { AlertCircle, Loader2, Copy, Check } from '@lucide/vue'
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
import { useAsyncAction } from '@/composables/useAsyncAction'

const props = defineProps<{
  open: boolean
  title: string
  description: string
  onCreate: (label: string, expiresAt: string | null) => Promise<string>
}>()

const emit = defineEmits<{
  'update:open': [value: boolean]
  created: []
}>()

const phase = ref<'form' | 'success'>('form')
const label = ref('')
const expiryPreset = ref('never')
const customDate = ref('')
const { loading, error, run } = useAsyncAction()
const createdKey = ref('')
const copied = ref(false)

const minDate = computed(() => new Date().toISOString().slice(0, 10))

const expiresAt = computed<string | null>(() => {
  if (expiryPreset.value === 'never') return null
  if (expiryPreset.value === 'custom') {
    if (!customDate.value) return null
    return new Date(customDate.value + 'T23:59:59Z').toISOString()
  }
  const days = expiryPreset.value === '30d' ? 30 : expiryPreset.value === '90d' ? 90 : 365
  const d = new Date()
  d.setDate(d.getDate() + days)
  return d.toISOString()
})

watch(
  () => props.open,
  (v) => {
    if (v) {
      phase.value = 'form'
      label.value = ''
      expiryPreset.value = 'never'
      customDate.value = ''
      error.value = ''
      createdKey.value = ''
      copied.value = false
    }
  },
  { immediate: true }
)

async function submit() {
  await run(async () => {
    createdKey.value = await props.onCreate(label.value.trim(), expiresAt.value)
    phase.value = 'success'
  })
}

function copy() {
  navigator.clipboard.writeText(createdKey.value)
  copied.value = true
  setTimeout(() => (copied.value = false), 2000)
}

function done() {
  emit('created')
  emit('update:open', false)
}
</script>
