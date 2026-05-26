import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { Project } from '@/api/projects'

export const useProjectStore = defineStore('project', () => {
  const projects = ref<Project[]>([])
  const current = ref<Project | null>(JSON.parse(localStorage.getItem('currentProject') ?? 'null'))

  function setCurrent(project: Project | null) {
    current.value = project
    if (project) localStorage.setItem('currentProject', JSON.stringify(project))
    else localStorage.removeItem('currentProject')
  }

  function setProjects(list: Project[]) {
    projects.value = list
    // If the stored project was deleted, clear it
    if (current.value && !list.find((p) => p.id === current.value!.id)) {
      current.value = null
      localStorage.removeItem('currentProject')
    }
    // Auto-select when there's exactly one project
    if (!current.value && list.length === 1) {
      setCurrent(list[0])
    }
  }

  return { projects, current, setCurrent, setProjects }
})
