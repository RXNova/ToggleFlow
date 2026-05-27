<template>
  <Dialog :open="open" @update:open="$emit('update:open', $event)">
    <DialogContent class="sm:max-w-sm">
      <DialogHeader>
        <DialogTitle>{{ $t('segments.createTitle') }}</DialogTitle>
        <DialogDescription>{{ $t('segments.createDescription') }}</DialogDescription>
      </DialogHeader>

      <form class="space-y-3" @submit.prevent="submit">
        <div class="space-y-2">
          <Label for="seg-name">{{ $t('segments.name') }}</Label>
          <Input
            id="seg-name"
            v-model="name"
            :placeholder="$t('segments.namePlaceholder')"
            autofocus
          />
        </div>

        <div class="space-y-2">
          <Label for="seg-key">{{ $t('segments.key') }}</Label>
          <Input
            id="seg-key"
            v-model="key"
            placeholder="test-users"
            class="font-mono"
            @focus="keyTouched = true"
          />
        </div>

        <div class="space-y-2">
          <Label>{{ $t('segments.values') }}</Label>
          <div
            class="flex flex-wrap gap-1 rounded-md border border-input bg-background px-2 py-1 min-h-9 focus-within:ring-2 focus-within:ring-ring cursor-text"
            @click="tagInputEl?.focus()"
          >
            <span
              v-for="(val, vi) in values"
              :key="vi"
              class="inline-flex items-center gap-0.5 rounded bg-muted px-1.5 py-0.5 text-xs"
            >
              {{ val }}
              <button type="button" class="hover:text-destructive" @click.stop="removeTag(vi)">
                <X class="size-2.5" />
              </button>
            </span>
            <input
              ref="tagInputEl"
              class="flex-1 min-w-20 text-sm bg-transparent outline-none py-0.5"
              :placeholder="values.length === 0 ? $t('segments.valuesPlaceholder') : ''"
              @keydown.enter.prevent="addTag"
              @keydown.backspace="onBackspace"
              @keydown="(e: KeyboardEvent) => e.key === ',' && (e.preventDefault(), addTag(e))"
              @paste="onPaste"
            />
          </div>
          <p class="text-xs text-muted-foreground">{{ $t('segments.valuesHint') }}</p>
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
            {{ loading ? $t('segments.creating') : $t('segments.create') }}
          </Button>
        </DialogFooter>
      </form>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { X, AlertCircle, Loader2 } from '@lucide/vue'
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
import { segmentsApi, type Segment } from '@/api/segments'
import { useTagInput } from '@/composables/useTagInput'
import { useAsyncAction } from '@/composables/useAsyncAction'

const { t } = useI18n()
const props = defineProps<{ open: boolean; projectId: number }>()
const emit = defineEmits<{
  'update:open': [value: boolean]
  created: [segment: Segment]
}>()

const name = ref('')
const keyRaw = ref('')
const keyTouched = ref(false)
const { values, tagInputEl, addTag, removeTag, onBackspace, onPaste } = useTagInput()
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

watch(name, (v) => {
  if (!keyTouched.value) keyRaw.value = slugify(v)
})

watch(
  () => props.open,
  (v) => {
    if (v) {
      name.value = ''
      keyRaw.value = ''
      keyTouched.value = false
      values.value = []
      error.value = ''
    }
  }
)

async function submit() {
  if (!name.value.trim()) {
    error.value = t('segments.errorName')
    return
  }
  await run(async () => {
    const seg = await segmentsApi.create(props.projectId, {
      name: name.value.trim(),
      key: key.value,
      values: values.value,
    })
    emit('created', seg)
    emit('update:open', false)
  })
}
</script>
