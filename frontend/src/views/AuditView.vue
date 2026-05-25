<template>
  <div class="p-6 space-y-4">
    <div class="space-y-0.5">
      <h1 class="text-base font-semibold">{{ $t('audit.title') }}</h1>
      <p class="text-sm text-muted-foreground">{{ $t('audit.subtitle') }}</p>
    </div>

    <Separator />

    <div class="flex flex-col items-center justify-center py-24 text-center">
      <ClipboardList class="size-8 text-muted-foreground/30 mb-3" />
      <p class="text-sm font-medium">{{ $t('audit.emptyTitle') }}</p>
      <p class="mt-1 text-xs text-muted-foreground max-w-xs">{{ $t('audit.emptyDescription') }}</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ClipboardList } from '@lucide/vue'
import { Separator } from '@/components/ui/separator'
</script>
