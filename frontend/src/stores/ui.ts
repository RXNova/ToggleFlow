import { defineStore } from 'pinia'
import { ref } from 'vue'

// UI store — only client-side state lives here.
// Server data (flags, projects, environments) lives in Tanstack Query, not here.
// Angular equivalent: a service with BehaviorSubjects for UI state.
export const useUIStore = defineStore('ui', () => {
  const activeProjectId = ref<number | null>(null)
  const activeEnvironmentId = ref<number | null>(null)
  const sidebarOpen = ref(true)

  function setActiveProject(id: number) {
    activeProjectId.value = id
  }

  function setActiveEnvironment(id: number) {
    activeEnvironmentId.value = id
  }

  function toggleSidebar() {
    sidebarOpen.value = !sidebarOpen.value
  }

  return { activeProjectId, activeEnvironmentId, sidebarOpen, setActiveProject, setActiveEnvironment, toggleSidebar }
})
