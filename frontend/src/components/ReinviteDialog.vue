<template>
  <Dialog :open="open" @update:open="onOpenChange">
    <DialogContent class="sm:max-w-sm">
      <template v-if="phase === 'form'">
        <DialogHeader>
          <DialogTitle>{{ $t('users.reinviteTitle') }}</DialogTitle>
          <DialogDescription>
            {{ $t('users.reinviteDescription', { name: user?.name }) }}
          </DialogDescription>
        </DialogHeader>

        <div class="space-y-2">
          <Label for="reinvite-expiry">{{ $t('users.expiresIn') }}</Label>
          <select
            id="reinvite-expiry"
            v-model="expiryDays"
            class="mt-2 w-full rounded-md border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
          >
            <option :value="1">{{ $t('users.expiry1d') }}</option>
            <option :value="3">{{ $t('users.expiry3d') }}</option>
            <option :value="7">{{ $t('users.expiry7d') }}</option>
            <option :value="14">{{ $t('users.expiry14d') }}</option>
            <option :value="30">{{ $t('users.expiry30d') }}</option>
          </select>
        </div>

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
            {{ loading ? $t('users.reinviting') : $t('users.reinviteGenerate') }}
          </Button>
        </DialogFooter>
      </template>

      <template v-else>
        <DialogHeader>
          <DialogTitle>{{ $t('users.inviteCreated') }}</DialogTitle>
          <DialogDescription>{{ $t('users.inviteSecurityNote') }}</DialogDescription>
        </DialogHeader>

        <div class="space-y-3">
          <div class="space-y-1.5">
            <Label>{{ $t('users.welcomeLink') }}</Label>
            <div class="mt-2 flex items-center gap-2">
              <Input :model-value="welcomeLink" readonly class="font-mono text-xs" />
              <Button
                size="sm"
                variant="outline"
                :title="$t('common.copyLink')"
                @click="copy(welcomeLink, 'link')"
              >
                <Check v-if="copied === 'link'" class="size-4" />
                <Copy v-else class="size-4" />
              </Button>
            </div>
          </div>

          <div class="space-y-1.5">
            <Label>{{ $t('users.welcomeSecret') }}</Label>
            <div class="mt-2 flex items-center gap-2">
              <Input
                :model-value="welcomeToken"
                readonly
                class="font-mono tracking-widest text-sm"
              />
              <Button
                size="sm"
                variant="outline"
                :title="$t('common.copySecret')"
                @click="copy(welcomeToken, 'secret')"
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
import { useAsyncAction } from '@/composables/useAsyncAction'

const props = defineProps<{ open: boolean; user: User | null }>()
const emit = defineEmits<{ 'update:open': [value: boolean] }>()

const phase = ref<'form' | 'success'>('form')
const expiryDays = ref(7)
const { loading, error, run } = useAsyncAction()
const welcomeToken = ref('')
const welcomeLink = ref('')
const copied = ref<'link' | 'secret' | null>(null)

watch(
  () => props.open,
  (v) => {
    if (v) {
      phase.value = 'form'
      expiryDays.value = 7
      error.value = ''
      welcomeToken.value = ''
      welcomeLink.value = ''
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
    const result = await usersApi.reinvite(props.user!.id, expiryDays.value)
    welcomeToken.value = result.welcome_token
    welcomeLink.value = `${window.location.origin}/activate?id=${result.user.uuid}`
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
