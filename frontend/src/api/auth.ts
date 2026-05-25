import { api } from './client'

export interface User {
  id: number
  name: string
  email: string
  role: string
  locale: string
}

interface AuthResponse {
  token: string
  user: User
}

export const authApi = {
  status: () => api.get<{ initialized: boolean }>('/setup/status'),
  setup: (name: string, email: string, password: string, locale: string) =>
    api.post<AuthResponse>('/setup', { name, email, password, locale }),
  login: (email: string, password: string) =>
    api.post<AuthResponse>('/auth/login', { email, password }),
  me: () => api.get<User>('/auth/me'),
}
