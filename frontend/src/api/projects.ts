import { api, type PageResult, type PageParams } from './client'

export interface Project {
  id: number
  name: string
  key: string
  description: string
  created_by: number | null
  created_by_name: string
  created_at: string
  updated_at: string
}

export const projectsApi = {
  list: (params?: PageParams) => api.get<PageResult<Project>>('/projects', params),
  create: (name: string, key: string, description: string) =>
    api.post<Project>('/projects', { name, key, description }),
  update: (id: number, name: string, key: string, description: string) =>
    api.patch<Project>(`/projects/${id}`, { name, key, description }),
  delete: (id: number) => api.delete<void>(`/projects/${id}`),
}
