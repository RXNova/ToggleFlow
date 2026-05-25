<template>
  <div class="flex h-screen bg-background">
    <!-- Sidebar -->
    <aside class="w-56 border-r flex flex-col gap-1 p-3 shrink-0">
      <div class="px-2 py-3 mb-2">
        <h1 class="text-lg font-semibold tracking-tight">ToggleFlow</h1>
      </div>
      <RouterLink
        v-for="item in nav"
        :key="item.to"
        :to="item.to"
        class="flex items-center gap-2 px-2 py-1.5 rounded-md text-sm text-muted-foreground hover:text-foreground hover:bg-accent transition-colors"
        activeClass="bg-accent text-foreground font-medium"
      >
        {{ item.label }}
      </RouterLink>
    </aside>

    <!-- Main content -->
    <main class="flex-1 overflow-auto p-6">
      <!-- RouterView here renders the child route — like a nested <router-outlet> -->
      <RouterView />
    </main>
  </div>
</template>

<script setup lang="ts">
import { RouterLink, RouterView } from 'vue-router'

const nav = [
  { label: 'Feature Flags', to: '/flags' },
  { label: 'Environments',  to: '/environments' },
  { label: 'Audit Log',     to: '/audit' },
]
</script>
