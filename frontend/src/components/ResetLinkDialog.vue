<template>
  <Dialog :open="open" @update:open="onOpenChange">
    <DialogContent class="sm:max-w-sm">
      <template v-if="phase === 'confirm'">
        <DialogHeader>
          <DialogTitle>{{ $t('users.resetTitle') }}</DialogTitle>
          <DialogDescription>
            {{ $t('users.resetDescription', { name: user?.name }) }}
          </DialogDescription>
        </DialogHeader>

        <Alert v-if="error" variant="destructive">
          <AlertCircle class="size-4" />
          <AlertDescription>{{ error }}</AlertDescription>
        </Alert>

        <DialogFooter>
          <Button type="button" variant="outline" @click="$emit('update:open', false)">
            {{ $t('common.cancel') }}
          </Button>
          <Button :disabled="loading" @click="generate">
            <Loader2 v-if="loading" class="size-4 animate-spin" />
            {{ loading ? $t('users.resetting') : $t('users.resetGenerate') }}
          </Button>
        </DialogFooter>
      </template>

      <template v-else>
        <DialogHeader>
          <DialogTitle>{{ $t('users.resetCreated') }}</DialogTitle>
          <DialogDescription>{{ $t('users.inviteSecurityNote') }}</DialogDescription>
        </DialogHeader>

        <div class="space-y-3">
          <div class="space-y-1.5">
            <Label>{{ $t('users.welcomeLink') }}</Label>
            <div class="mt-2 flex items-center gap-2">
              <Input :model-value="resetLink" readonly class="font-mono text-xs" />
              <Tooltip :text="$t('common.copyLink')">
                <Button size="sm" variant="outline" @click="copy(resetLink, 'link')">
                  <Check v-if="copied === 'link'" class="size-4" />
                  <Copy v-else class="size-4" />
                </Button>
              </Tooltip>
            </div>
          </div>

          <div class="space-y-1.5">
            <Label>{{ $t('users.welcomeSecret') }}</Label>
            <div class="mt-2 flex items-center gap-2">
              <Input :model-value="resetToken" readonly class="font-mono tracking-widest text-sm" />
              <Button
                size="sm"
                variant="outline"
                :title="$t('common.copySecret')"
                @click="copy(resetToken, 'secret')"
              >
                <Check v-if="copied === 'secret'" class="size-4" />
                <Copy v-else class="size-4" />
              </Button>
            </div>
          </div>
        </div>

        <DialogFooter>
          <Button @click="$emit('update:open', false)">{{ $t('users.done') }}</Button>
        </DialogFooter>
      </template>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
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
import { usersApi, type User } from '@/api/users'
import { Tooltip } from '@/components/ui/tooltip'
import { useAsyncAction } from '@/composables/useAsyncAction'

const props = defineProps<{ open: boolean; user: User | null }>()
const emit = defineEmits<{ 'update:open': [value: boolean] }>()

const phase = ref<'confirm' | 'success'>('confirm')
const { loading, error, run } = useAsyncAction()
const resetToken = ref('')
const copied = ref<'link' | 'secret' | null>(null)

const resetLink = ref('')

watch(
  () => props.open,
  (v) => {
    if (v) {
      phase.value = 'confirm'
      error.value = ''
      resetToken.value = ''
      resetLink.value = ''
      copied.value = null
    }
  }
)

function onOpenChange(v: boolean) {
  emit('update:open', v)
}

async function generate() {
  if (!props.user) return
  await run(async () => {
    const result = await usersApi.generateResetLink(props.user!.id)
    resetToken.value = result.reset_token
    resetLink.value = `${window.location.origin}/reset?id=${result.user.uuid}`
    phase.value = 'success'
  })
}

async function copy(text: string, which: 'link' | 'secret') {
  await navigator.clipboard.writeText(text)
  copied.value = which
  setTimeout(() => {
    copied.value = null
  }, 2000)
}
</script>
