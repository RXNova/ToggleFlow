<template>
  <div class="p-6 space-y-4">
    <div class="flex items-center justify-between">
      <div class="space-y-0.5">
        <h1 class="text-base font-semibold">{{ $t('users.title') }}</h1>
        <p class="text-sm text-muted-foreground">{{ $t('users.subtitle') }}</p>
      </div>
      <Button size="sm" disabled>
        <UserPlus class="size-3.5" />
        {{ $t('users.invite') }}
      </Button>
    </div>

    <Separator />

    <div class="flex flex-col items-center justify-center py-24 text-center">
      <Users class="size-8 text-muted-foreground/30 mb-3" />
      <p class="text-sm font-medium">{{ $t('users.emptyTitle') }}</p>
      <p class="mt-1 text-xs text-muted-foreground max-w-xs">{{ $t('users.emptyDescription') }}</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { Users, UserPlus } from '@lucide/vue'
import { Button } from '@/components/ui/button'
import { Separator } from '@/components/ui/separator'
</script>
