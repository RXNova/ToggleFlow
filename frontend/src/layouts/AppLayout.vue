<template>
  <div class="flex h-screen overflow-hidden bg-background">

    <!-- Desktop sidebar: always visible on lg+ -->
    <aside class="hidden lg:flex w-56 flex-none flex-col border-r bg-sidebar">
      <SidebarNav />
    </aside>

    <!-- Right side: mobile top bar + scrollable content -->
    <div class="flex flex-1 flex-col overflow-hidden">

      <!-- Mobile top bar -->
      <div class="flex h-12 shrink-0 items-center gap-3 border-b bg-background px-4 lg:hidden">
        <button
          class="rounded-md p-1 text-foreground transition-colors hover:bg-accent"
          @click="drawerOpen = true"
        >
          <Menu class="size-5" />
        </button>
        <div class="flex items-center gap-2">
          <ToggleRight class="size-4" />
          <span class="text-sm font-semibold tracking-tight">ToggleFlow</span>
        </div>
      </div>

      <!-- Page content -->
      <main class="flex-1 overflow-y-auto">
        <RouterView />
      </main>

    </div>

    <!-- Mobile drawer backdrop -->
    <Transition
      enter-from-class="opacity-0"
      enter-active-class="transition-opacity duration-200"
      leave-active-class="transition-opacity duration-200"
      leave-to-class="opacity-0"
    >
      <div
        v-if="drawerOpen"
        class="fixed inset-0 z-40 bg-black/40 lg:hidden"
        @click="drawerOpen = false"
      />
    </Transition>

    <!-- Mobile drawer panel -->
    <Transition
      enter-from-class="-translate-x-full"
      enter-active-class="transition-transform duration-200 ease-out"
      leave-active-class="transition-transform duration-200 ease-in"
      leave-to-class="-translate-x-full"
    >
      <aside
        v-if="drawerOpen"
        class="fixed left-0 top-0 z-50 flex h-full w-64 flex-col bg-sidebar shadow-xl lg:hidden"
      >
        <SidebarNav @navigate="drawerOpen = false" />
      </aside>
    </Transition>

  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { Menu, ToggleRight } from '@lucide/vue'
import SidebarNav from './SidebarNav.vue'

const drawerOpen = ref(false)
</script>
