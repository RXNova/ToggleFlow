import { api } from './client'

export type Role = 'superuser' | 'admin' | 'owner' | 'editor' | 'viewer'

export interface User {
  id: number
  uuid: string
  name: string
  email: string
  role: Role
  locale: string
  activated_at?: string
  created_by?: number
  created_at: string
  updated_at: string
}

export interface InviteResult {
  user: User
  welcome_token: string
}

export interface ResetResult {
  user: User
  reset_token: string
}

export const usersApi = {
  list: () => api.get<User[]>('/users'),
  create: (name: string, email: string, role: Role, expiryDays: number) =>
    api.post<InviteResult>('/users', { name, email, role, expiry_days: expiryDays }),
  update: (id: number, data: Partial<{ name: string; email: string; role: Role }>) =>
    api.patch<User>(`/users/${id}`, data),
  reinvite: (id: number, expiryDays: number) =>
    api.post<InviteResult>(`/users/${id}/reinvite`, { expiry_days: expiryDays }),
  delete: (id: number) => api.delete<void>(`/users/${id}`),
  generateResetLink: (id: number, expiryDays = 1) =>
    api.post<ResetResult>(`/users/${id}/reset-link`, { expiry_days: expiryDays }),
}
