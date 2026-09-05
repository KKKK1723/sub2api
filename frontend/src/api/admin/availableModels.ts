import { apiClient } from '../client'

export interface AvailableModelGroup {
  group_id: number
  group_name: string
  models: string[]
  enabled: boolean
  last_synced_at?: string
  sync_error?: string
}

export async function list(): Promise<AvailableModelGroup[]> {
  const { data } = await apiClient.get<AvailableModelGroup[]>('/available-models')
  return data
}

export async function sync(groupId: number): Promise<AvailableModelGroup> {
  const { data } = await apiClient.post<AvailableModelGroup>(`/admin/available-models/${groupId}/sync`)
  return data
}

export async function update(groupId: number, payload: { enabled?: boolean; models: string[] }) {
  const { data } = await apiClient.put(`/admin/available-models/${groupId}`, payload)
  return data
}

export async function remove(groupId: number) {
  const { data } = await apiClient.delete(`/admin/available-models/${groupId}`)
  return data
}

export default { list, sync, update, remove }
