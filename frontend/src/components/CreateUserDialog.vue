<template>
  <Dialog :open="open" @update:open="onOpenChange">
    <DialogContent class="sm:max-w-sm">
      <!-- Phase 1: invite form -->
      <template v-if="phase === 'form'">
        <DialogHeader>
          <DialogTitle>{{ $t('users.createTitle') }}</DialogTitle>
          <DialogDescription>{{ $t('users.createDescription') }}</DialogDescription>
        </DialogHeader>

        <form class="space-y-3" @submit.prevent="submit">
          <div class="space-y-2">
            <Label for="user-name">{{ $t('users.name') }}</Label>
            <Input
              id="user-name"
              v-model="name"
              :placeholder="$t('users.namePlaceholder')"
              class="mt-2"
              required
              autofocus
            />
          </div>

          <div class="space-y-2">
            <Label for="user-email">{{ $t('users.email') }}</Label>
            <Input
              id="user-email"
              v-model="email"
              type="email"
              :placeholder="$t('users.emailPlaceholder')"
              class="mt-2"
              required
            />
          </div>

          <div class="space-y-2">
            <Label for="user-role">{{ $t('users.role') }}</Label>
            <select
              id="user-role"
              v-model="role"
              class="mt-2 w-full rounded-md border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
              required
            >
              <option v-for="r in availableRoles" :key="r" :value="r">
                {{ $t(`roles.${r}`) }}
              </option>
            </select>
          </div>

          <div class="space-y-2">
            <Label for="user-expiry">{{ $t('users.expiresIn') }}</Label>
            <select
              id="user-expiry"
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
            <Button type="submit" :disabled="loading">
              <Loader2 v-if="loading" class="size-4 animate-spin" />
              {{ loading ? $t('users.creating') : $t('users.invite') }}
            </Button>
          </DialogFooter>
        </form>
      </template>

      <!-- Phase 2: show welcome link + secret -->
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
              <Tooltip :text="$t('common.copyLink')">
                <Button size="sm" variant="outline" @click="copy(welcomeLink, 'link')">
                  <Check v-if="copied === 'link'" class="size-4" />
                  <Copy v-else class="size-4" />
                </Button>
              </Tooltip>
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
              <Tooltip :text="$t('common.copySecret')">
                <Button size="sm" variant="outline" @click="copy(welcomeToken, 'secret')">
                  <Check v-if="copied === 'secret'" class="size-4" />
                  <Copy v-else class="size-4" />
                </Button>
              </Tooltip>
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
import { usersApi, type User, type Role } from '@/api/users'
import { useAuthStore } from '@/stores/auth'
import { Tooltip } from '@/components/ui/tooltip'
import { useAsyncAction } from '@/composables/useAsyncAction'

const props = defineProps<{ open: boolean }>()
const emit = defineEmits<{
  'update:open': [value: boolean]
  created: [user: User]
}>()

const authStore = useAuthStore()

const phase = ref<'form' | 'success'>('form')
const name = ref('')
const email = ref('')
const role = ref<Role>('viewer')
const expiryDays = ref(7)
const { loading, error, run } = useAsyncAction()
const welcomeToken = ref('')
const copied = ref<'link' | 'secret' | null>(null)

const inviteUUID = ref('')
const welcomeLink = computed(() =>
  inviteUUID.value ? `${window.location.origin}/activate?id=${inviteUUID.value}` : ''
)

const availableRoles = computed<Role[]>(() => {
  if (authStore.isSuperuser) return ['superuser', 'admin', 'owner', 'editor', 'viewer']
  return ['owner', 'editor', 'viewer']
})

watch(
  () => props.open,
  (v) => {
    if (v) {
      phase.value = 'form'
      name.value = ''
      email.value = ''
      role.value = 'viewer'
      expiryDays.value = 7
      error.value = ''
      welcomeToken.value = ''
      inviteUUID.value = ''
      copied.value = null
    }
  }
)

function onOpenChange(v: boolean) {
  emit('update:open', v)
}

async function submit() {
  await run(async () => {
    const result = await usersApi.create(name.value, email.value, role.value, expiryDays.value)
    emit('created', result.user)
    welcomeToken.value = result.welcome_token
    inviteUUID.value = result.user.uuid
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
