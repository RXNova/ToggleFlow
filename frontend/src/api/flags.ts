import { api } from './client'

export type FlagType = 'boolean' | 'string' | 'number' | 'json'

export interface Variation {
  name: string
  value: boolean | string | number | Record<string, unknown>
}

export interface FlagEnvState {
  environment_id: number
  environment_name: string
  environment_slug: string
  enabled: boolean
  default_variation: number
}

export interface Flag {
  id: number
  project_id: number
  key: string
  name: string
  description: string
  flag_type: FlagType
  variations: Variation[]
  created_at: string
  updated_at: string
  environments: FlagEnvState[]
}

export const flagsApi = {
  list: (projectId: number) => api.get<Flag[]>(`/projects/${projectId}/flags`),
  create: (projectId: number, data: { name: string; key: string; description?: string; flag_type: FlagType; variations: Variation[] }) =>
    api.post<Flag>(`/projects/${projectId}/flags`, data),
  toggle: (projectId: number, flagKey: string, environmentId: number, enabled: boolean, defaultVariation: number) =>
    api.patch<{ ok: boolean }>(`/projects/${projectId}/flags/${flagKey}`, { environment_id: environmentId, enabled, default_variation: defaultVariation }),
  delete: (projectId: number, flagKey: string) =>
    api.delete<void>(`/projects/${projectId}/flags/${flagKey}`),
}
