<template>
  <div class="min-h-screen grid lg:grid-cols-2">
    <!-- Left: branding panel -->
    <div class="hidden lg:flex flex-col justify-between bg-primary text-primary-foreground p-8">
      <div class="flex items-center gap-2">
        <ToggleRight class="size-5" />
        <span class="font-semibold tracking-tight">ToggleFlow</span>
      </div>

      <div class="space-y-8">
        <div class="space-y-2">
          <h1 class="text-3xl font-bold leading-tight">{{ $t('brand.tagline') }}</h1>
          <p class="text-sm leading-relaxed max-w-xs opacity-60">
            {{ $t('brand.description') }}
          </p>
        </div>

        <ul class="space-y-3">
          <li
            v-for="feat in features"
            :key="feat.label"
            class="flex items-center gap-3 text-sm opacity-80"
          >
            <div class="flex size-7 shrink-0 items-center justify-center rounded-md bg-white/10">
              <component :is="feat.icon" class="size-3.5" />
            </div>
            {{ feat.label }}
          </li>
        </ul>
      </div>

      <p class="text-xs opacity-40">{{ $t('brand.footer') }}</p>
    </div>

    <!-- Right: setup form -->
    <div class="flex flex-col items-center justify-center bg-background px-8 py-10">
      <div class="w-full max-w-sm space-y-5">
        <!-- Mobile logo -->
        <div class="flex items-center gap-2 lg:hidden">
          <ToggleRight class="size-5 text-primary" />
          <span class="font-semibold tracking-tight">ToggleFlow</span>
        </div>

        <div class="space-y-1">
          <h2 class="text-xl font-semibold tracking-tight">{{ $t('setup.title') }}</h2>
          <p class="text-sm text-muted-foreground">{{ $t('setup.subtitle') }}</p>
        </div>

        <form class="space-y-3" @submit.prevent="submit">
          <div class="space-y-2">
            <Label for="name">{{ $t('setup.name') }}</Label>
            <Input
              id="name"
              v-model="form.name"
              type="text"
              :placeholder="$t('setup.namePlaceholder')"
              class="mt-2"
              required
            />
          </div>

          <div class="space-y-2">
            <Label for="email">{{ $t('setup.email') }}</Label>
            <Input
              id="email"
              v-model="form.email"
              type="email"
              :placeholder="$t('setup.emailPlaceholder')"
              class="mt-2"
              required
            />
          </div>

          <div class="space-y-2">
            <Label for="password">{{ $t('setup.password') }}</Label>
            <Input
              id="password"
              v-model="form.password"
              type="password"
              :placeholder="$t('setup.passwordPlaceholder')"
              class="mt-2"
              required
              minlength="8"
            />
          </div>

          <Alert v-if="error" variant="destructive">
            <AlertCircle class="size-4" />
            <AlertDescription>{{ error }}</AlertDescription>
          </Alert>

          <Button type="submit" class="w-full" :disabled="loading">
            <Loader2 v-if="loading" class="size-4 animate-spin" />
            {{ loading ? $t('setup.submitting') : $t('setup.submit') }}
          </Button>
        </form>

        <div class="flex justify-center">
          <LangSwitcher />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ToggleRight, Loader2, AlertCircle, Flag, Zap, Shield } from '@lucide/vue'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Button } from '@/components/ui/button'
import { Alert, AlertDescription } from '@/components/ui/alert'
import LangSwitcher from '@/components/LangSwitcher.vue'
import { authApi } from '@/api/auth'
import { useAuthStore } from '@/stores/auth'
import { getLocale, setLocale, type Locale } from '@/plugins/i18n'

const { t } = useI18n()
const router = useRouter()
const authStore = useAuthStore()

const form = ref({ name: '', email: '', password: '' })
const loading = ref(false)
const error = ref('')

const features = computed(() => [
  { icon: Flag, label: t('brand.features.targeting') },
  { icon: Zap, label: t('brand.features.realtime') },
  { icon: Shield, label: t('brand.features.rbac') },
])

async function submit() {
  error.value = ''
  loading.value = true
  try {
    const { token, user } = await authApi.setup(
      form.value.name,
      form.value.email,
      form.value.password,
      getLocale()
    )
    authStore.setAuth(token, user)
    if (user.locale) setLocale(user.locale as Locale)
    router.push('/')
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : 'Something went wrong'
  } finally {
    loading.value = false
  }
}
</script>
