<template>
  <div class="p-6 space-y-4">
    <div class="flex items-center justify-between">
      <div class="space-y-0.5">
        <h1 class="text-base font-semibold">{{ $t('environments.title') }}</h1>
        <p class="text-sm text-muted-foreground">{{ $t('environments.subtitle') }}</p>
      </div>
      <Button size="sm" disabled>
        <Plus class="size-3.5" />
        {{ $t('environments.create') }}
      </Button>
    </div>

    <Separator />

    <div class="flex flex-col items-center justify-center py-24 text-center">
      <Globe class="size-8 text-muted-foreground/30 mb-3" />
      <p class="text-sm font-medium">{{ $t('environments.emptyTitle') }}</p>
      <p class="mt-1 text-xs text-muted-foreground max-w-xs">{{ $t('environments.emptyDescription') }}</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { Globe, Plus } from '@lucide/vue'
import { Button } from '@/components/ui/button'
import { Separator } from '@/components/ui/separator'
</script>
