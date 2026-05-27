import { api, type PageResult, type PageParams } from './client'

export interface AuditEntry {
  id: number
  project_id: number
  actor: string
  action: string
  resource: string
  old_value: string
  new_value: string
  created_at: string
}

export const auditApi = {
  list: (projectId: number, params?: PageParams & { resource?: string; actor?: string }) =>
    api.get<PageResult<AuditEntry>>(`/projects/${projectId}/audit`, params),
}

export const userAuditApi = {
  list: (userId: number, params?: PageParams) =>
    api.get<PageResult<AuditEntry>>(`/users/${userId}/audit`, params),
}
