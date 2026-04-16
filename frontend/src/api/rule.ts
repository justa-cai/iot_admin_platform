import client from './client'
import type {
  Rule,
  RuleCreateRequest,
  RuleLog,
  PaginatedResponse,
} from '@/types'

export interface RuleQueryParams {
  page?: number
  page_size?: number
  search?: string
}

export function getRules(params?: RuleQueryParams): Promise<PaginatedResponse<Rule>> {
  return client.get('/rules', { params })
}

export function getRule(id: string): Promise<Rule> {
  return client.get(`/rules/${id}`)
}

export function createRule(data: RuleCreateRequest): Promise<Rule> {
  return client.post('/rules', data)
}

export function updateRule(id: string, data: Partial<RuleCreateRequest>): Promise<Rule> {
  return client.put(`/rules/${id}`, data)
}

export function deleteRule(id: string): Promise<void> {
  return client.delete(`/rules/${id}`)
}

export function enableRule(id: string): Promise<Rule> {
  return client.put(`/rules/${id}/enable`, { enabled: true })
}

export function disableRule(id: string): Promise<Rule> {
  return client.put(`/rules/${id}/enable`, { enabled: false })
}

export function getRuleLogs(id: string, params?: { page?: number; page_size?: number }): Promise<PaginatedResponse<RuleLog>> {
  return client.get(`/rules/${id}/logs`, { params })
}
