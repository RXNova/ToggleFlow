import { api, type PageResult, type PageParams } from './client'

export interface Environment {
  id: number
  project_id: number
  name: string
  slug: string
  description: string
  sdk_key: string
  created_at: string
  updated_at: string
}

export const environmentsApi = {
  list: (projectId: number, params?: PageParams) =>
    api.get<PageResult<Environment>>(`/projects/${projectId}/environments`, params),
  create: (projectId: number, name: string, description: string) =>
    api.post<Environment>(`/projects/${projectId}/environments`, { name, description }),
  update: (projectId: number, envId: number, name: string, description: string) =>
    api.patch<Environment>(`/projects/${projectId}/environments/${envId}`, { name, description }),
  delete: (projectId: number, envId: number) =>
    api.delete<void>(`/projects/${projectId}/environments/${envId}`),
}
