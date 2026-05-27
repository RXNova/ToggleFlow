<template>
  <Dialog :open="open" @update:open="$emit('update:open', $event)">
    <DialogContent class="sm:max-w-sm">
      <DialogHeader>
        <DialogTitle>{{ $t('users.editTitle') }}</DialogTitle>
        <DialogDescription>{{ $t('users.editDescription') }}</DialogDescription>
      </DialogHeader>

      <form class="space-y-3" @submit.prevent="submit">
        <div class="space-y-2">
          <Label for="edit-user-name">{{ $t('users.name') }}</Label>
          <Input
            id="edit-user-name"
            v-model="name"
            :placeholder="$t('users.namePlaceholder')"
            class="mt-2"
            required
            autofocus
          />
        </div>

        <div class="space-y-2">
          <Label for="edit-user-email">{{ $t('users.email') }}</Label>
          <Input
            id="edit-user-email"
            v-model="email"
            type="email"
            :placeholder="$t('users.emailPlaceholder')"
            class="mt-2"
            required
          />
        </div>

        <div class="space-y-2">
          <Label for="edit-user-role">{{ $t('users.role') }}</Label>
          <select
            id="edit-user-role"
            v-model="role"
            :disabled="roleDisabled"
            class="mt-2 w-full rounded-md border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring disabled:cursor-not-allowed disabled:opacity-50"
          >
            <option v-for="r in availableRoles" :key="r" :value="r">{{ $t(`roles.${r}`) }}</option>
          </select>
          <p v-if="isSelfSuperuser" class="text-xs text-muted-foreground">
            {{ $t('users.selfRoleNote') }}
          </p>
          <p v-else-if="!authStore.isSuperuser" class="text-xs text-muted-foreground">
            {{ $t('users.roleNote') }}
          </p>
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
import { usersApi, type User, type Role } from '@/api/users'
import { useAuthStore } from '@/stores/auth'
import { useAsyncAction } from '@/composables/useAsyncAction'

const props = defineProps<{ open: boolean; user: User | null }>()
const emit = defineEmits<{
  'update:open': [value: boolean]
  updated: [user: User]
}>()

const authStore = useAuthStore()

const name = ref('')
const email = ref('')
const role = ref<Role>('viewer')
const { loading, error, run } = useAsyncAction()

const isSelfSuperuser = computed(
  () => authStore.isSuperuser && props.user?.id === authStore.user?.id
)

const roleDisabled = computed(() => !authStore.isSuperuser || isSelfSuperuser.value)

const availableRoles = computed<Role[]>(() => {
  if (authStore.isSuperuser) return ['superuser', 'admin', 'owner', 'editor', 'viewer']
  return ['owner', 'editor', 'viewer']
})

watch(
  () => props.open,
  (v) => {
    if (v && props.user) {
      name.value = props.user.name
      email.value = props.user.email
      role.value = props.user.role
      error.value = ''
    }
  }
)

async function submit() {
  if (!props.user) return
  await run(async () => {
    const updated = await usersApi.update(props.user!.id, {
      name: name.value,
      email: email.value,
      role: role.value,
    })
    emit('updated', updated)
    emit('update:open', false)
  })
}
</script>
