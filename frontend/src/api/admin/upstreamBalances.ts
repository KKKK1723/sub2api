import { apiClient } from '../client'

export interface UpstreamBalanceOverviewItem {
  base_url: string
  account_id: number
  group_ids?: number[]
  group_names?: string[]
  balance?: number
  unit?: string
  last_synced_at?: string
  status: 'ok' | 'unsupported' | 'failed'
}

export async function refresh(): Promise<UpstreamBalanceOverviewItem[]> {
  const { data } = await apiClient.post<UpstreamBalanceOverviewItem[]>('/admin/accounts/upstream-balances/refresh')
  return data
}

export interface UpstreamBalanceWriteRequest {
  base_url: string
  api_key?: string
  name?: string
  platform?: string
  group_ids?: number[]
}

export async function create(payload: UpstreamBalanceWriteRequest) {
  const { data } = await apiClient.post('/admin/accounts/upstream-balances', payload)
  return data
}

export async function update(id: number, payload: UpstreamBalanceWriteRequest) {
  const { data } = await apiClient.put(`/admin/accounts/upstream-balances/${id}`, payload)
  return data
}

export async function remove(id: number) {
  await apiClient.delete(`/admin/accounts/upstream-balances/${id}`)
}

export default { refresh, create, update, remove }
