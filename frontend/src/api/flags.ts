import { api, type PageResult, type PageParams } from './client'
import { useToastStore } from '@/stores/toast'

export type FlagType = 'boolean' | 'string' | 'number' | 'json'

export type ConditionOperator =
  | 'in'
  | 'notIn'
  | 'contains'
  | 'startsWith'
  | 'endsWith'
  | 'equals'
  | 'gt'
  | 'gte'
  | 'lt'
  | 'lte'

export interface Condition {
  attribute: string
  operator: ConditionOperator
  values: (string | number)[]
  segment?: string
}

export interface RolloutStep {
  variation: number
  weight: number
}

export interface Rule {
  conditions: Condition[]
  serve: number | null
  rollout: RolloutStep[] | null
}

export interface Variation {
  name: string
  value: boolean | string | number | Record<string, unknown>
}

export interface FlagEnvState {
  environment_id: number
  environment_name: string
  environment_key: string
  protected: boolean
  enabled: boolean
  default_variation: number
  rules: Rule[]
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
  list: (projectId: number, params?: PageParams) =>
    api.get<PageResult<Flag>>(`/projects/${projectId}/flags`, params),
  create: (
    projectId: number,
    data: {
      name: string
      key: string
      description?: string
      flag_type: FlagType
      variations: Variation[]
    }
  ) =>
    api.post<Flag>(`/projects/${projectId}/flags`, data).then((r) => {
      useToastStore().show('created flag')
      return r
    }),
  update: (
    projectId: number,
    flagKey: string,
    data: { name: string; description: string; variations: Variation[] }
  ) =>
    api.patch<Flag>(`/projects/${projectId}/flags/${flagKey}`, data).then((r) => {
      useToastStore().show('updated flag')
      return r
    }),
  toggle: (
    projectId: number,
    flagKey: string,
    environmentId: number,
    enabled: boolean,
    defaultVariation: number
  ) =>
    api
      .patch<{ ok: boolean }>(`/projects/${projectId}/flags/${flagKey}/env`, {
        environment_id: environmentId,
        enabled,
        default_variation: defaultVariation,
      })
      .then((r) => {
        useToastStore().show(enabled ? 'enabled flag' : 'disabled flag')
        return r
      }),
  saveRules: (projectId: number, flagKey: string, environmentId: number, rules: Rule[]) =>
    api
      .put<{ ok: boolean }>(`/projects/${projectId}/flags/${flagKey}/rules`, {
        environment_id: environmentId,
        rules,
      })
      .then((r) => {
        useToastStore().show('saved rules')
        return r
      }),
  delete: (projectId: number, flagKey: string) =>
    api.delete<void>(`/projects/${projectId}/flags/${flagKey}`).then((r) => {
      useToastStore().show('deleted flag')
      return r
    }),
}
