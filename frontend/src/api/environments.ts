import { api, type PageResult, type PageParams } from './client'

export interface Environment {
  id: number
  project_id: number
  name: string
  slug: string
  sdk_key: string
  created_at: string
}

export const environmentsApi = {
  list: (projectId: number, params?: PageParams) =>
    api.get<PageResult<Environment>>(`/projects/${projectId}/environments`, params),
  create: (projectId: number, name: string) =>
    api.post<Environment>(`/projects/${projectId}/environments`, { name }),
}
