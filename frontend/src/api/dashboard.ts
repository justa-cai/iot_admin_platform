import client from './client'
import type { DashboardStats, ThroughputData } from '@/types'

export function getDashboardStats(): Promise<DashboardStats> {
  return client.get('/dashboard/stats')
}

export function getThroughput(hours: number = 24): Promise<ThroughputData> {
  return client.get('/dashboard/throughput', { params: { hours } })
}
