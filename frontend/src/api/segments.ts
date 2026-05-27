import { api } from './client'
import { useToastStore } from '@/stores/toast'

export interface Segment {
  id: number
  project_id: number
  name: string
  key: string
  values: (string | number)[]
  created_at: string
  updated_at: string
}

export const segmentsApi = {
  list: (projectId: number) => api.get<Segment[]>(`/projects/${projectId}/segments`),

  create: (projectId: number, data: { name: string; key: string; values: (string | number)[] }) =>
    api.post<Segment>(`/projects/${projectId}/segments`, data).then((r) => {
      useToastStore().show('created segment')
      return r
    }),

  update: (
    projectId: number,
    segmentId: number,
    data: { name: string; values: (string | number)[] }
  ) =>
    api.patch<Segment>(`/projects/${projectId}/segments/${segmentId}`, data).then((r) => {
      useToastStore().show('updated segment')
      return r
    }),

  delete: (projectId: number, segmentId: number) =>
    api.delete<void>(`/projects/${projectId}/segments/${segmentId}`).then((r) => {
      useToastStore().show('deleted segment')
      return r
    }),
}
