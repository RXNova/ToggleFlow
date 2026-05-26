<template>
  <div class="p-6 space-y-4">
    <div class="space-y-0.5">
      <h1 class="text-base font-semibold">{{ $t('settings.title') }}</h1>
      <p class="text-sm text-muted-foreground">{{ $t('settings.subtitle') }}</p>
    </div>

    <Separator />

    <div class="max-w-md space-y-6">
      <!-- Appearance -->
      <div class="space-y-3">
        <div class="space-y-0.5">
          <p class="text-sm font-medium">{{ $t('settings.appearance') }}</p>
          <p class="text-xs text-muted-foreground">{{ $t('settings.appearanceDescription') }}</p>
        </div>
        <div class="grid grid-cols-3 gap-2">
          <button
            v-for="opt in themeOptions"
            :key="opt.value"
            class="flex flex-col items-center gap-2 rounded-lg border p-3 text-xs font-medium transition-colors"
            :class="
              themeStore.theme === opt.value
                ? 'border-primary bg-primary/5 text-primary'
                : 'border-border text-muted-foreground hover:border-foreground/30 hover:text-foreground'
            "
            @click="themeStore.setTheme(opt.value)"
          >
            <component :is="opt.icon" class="size-4" />
            {{ $t(`settings.theme.${opt.value}`) }}
          </button>
        </div>
      </div>

      <Separator />

      <!-- Language -->
      <div class="space-y-3">
        <div class="space-y-0.5">
          <p class="text-sm font-medium">{{ $t('settings.language') }}</p>
          <p class="text-xs text-muted-foreground">{{ $t('settings.languageDescription') }}</p>
        </div>
        <div class="flex gap-2">
          <button
            v-for="loc in SUPPORTED_LOCALES"
            :key="loc"
            class="flex-1 rounded-md border px-3 py-2 text-sm font-medium transition-colors"
            :class="
              locale === loc
                ? 'border-primary bg-primary/5 text-primary'
                : 'border-border text-muted-foreground hover:border-foreground/30 hover:text-foreground'
            "
            @click="setLocale(loc)"
          >
            {{ $t(`lang.${loc}`) }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { Sun, Moon, Monitor } from '@lucide/vue'
import { useI18n } from 'vue-i18n'
import { Separator } from '@/components/ui/separator'
import { useThemeStore, type Theme } from '@/stores/theme'
import { setLocale, SUPPORTED_LOCALES } from '@/plugins/i18n'

const { locale } = useI18n()
const themeStore = useThemeStore()

const themeOptions: { value: Theme; icon: typeof Sun }[] = [
  { value: 'light', icon: Sun },
  { value: 'dark', icon: Moon },
  { value: 'system', icon: Monitor },
]
</script>
