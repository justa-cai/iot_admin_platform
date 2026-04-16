import client from './client'
import type {
  TelemetryData,
  TelemetryQueryParams,
  TelemetryLatest,
  PaginatedResponse,
} from '@/types'

export function queryTelemetry(params: TelemetryQueryParams): Promise<PaginatedResponse<TelemetryData>> {
  return client.get('/telemetry/query', { params })
}

export function getLatestTelemetry(deviceId: number, limit: number = 50): Promise<TelemetryLatest[]> {
  return client.get('/telemetry/latest', { params: { device_id: deviceId, limit } })
}
