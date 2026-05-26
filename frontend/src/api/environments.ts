import { api, type PageResult, type PageParams } from './client'

export interface Environment {
  id: number
  project_id: number
  name: string
  key: string
  description: string
  protected: boolean
  sdk_key: string
  created_at: string
  updated_at: string
}

export const environmentsApi = {
  list: (projectId: number, params?: PageParams) =>
    api.get<PageResult<Environment>>(`/projects/${projectId}/environments`, params),
  create: (
    projectId: number,
    name: string,
    key: string,
    description: string,
    isProtected: boolean
  ) =>
    api.post<Environment>(`/projects/${projectId}/environments`, {
      name,
      key,
      description,
      protected: isProtected,
    }),
  update: (
    projectId: number,
    envId: number,
    name: string,
    key: string,
    description: string,
    isProtected: boolean
  ) =>
    api.patch<Environment>(`/projects/${projectId}/environments/${envId}`, {
      name,
      key,
      description,
      protected: isProtected,
    }),
  delete: (projectId: number, envId: number) =>
    api.delete<void>(`/projects/${projectId}/environments/${envId}`),
}
