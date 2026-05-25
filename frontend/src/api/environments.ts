import { api } from './client'

export interface Environment {
  id: number
  project_id: number
  name: string
  slug: string
  sdk_key: string
  created_at: string
}

export const environmentsApi = {
  list: (projectId: number) => api.get<Environment[]>(`/projects/${projectId}/environments`),
  create: (projectId: number, name: string) =>
    api.post<Environment>(`/projects/${projectId}/environments`, { name }),
}
