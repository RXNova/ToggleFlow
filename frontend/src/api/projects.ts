import { api, type PageResult, type PageParams } from './client'

export interface Project {
  id: number
  name: string
  slug: string
  created_at: string
  updated_at: string
}

export const projectsApi = {
  list: (params?: PageParams) =>
    api.get<PageResult<Project>>('/projects', params),
  create: (name: string, slug: string) =>
    api.post<Project>('/projects', { name, slug }),
  update: (id: number, name: string, slug: string) =>
    api.patch<Project>(`/projects/${id}`, { name, slug }),
  delete: (id: number) =>
    api.delete<void>(`/projects/${id}`),
}
