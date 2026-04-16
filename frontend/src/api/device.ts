import client from './client'
import type {
  Device,
  DeviceCreateRequest,
  DeviceUpdateRequest,
  DeviceStatus,
  PaginatedResponse,
} from '@/types'

export interface DeviceQueryParams {
  page?: number
  page_size?: number
  search?: string
  status?: string
  group_id?: number
  tag_id?: number
}

export function getDevices(params?: DeviceQueryParams): Promise<PaginatedResponse<Device>> {
  return client.get('/devices', { params })
}

export function getDevice(id: number): Promise<Device> {
  return client.get(`/devices/${id}`)
}

export function createDevice(data: DeviceCreateRequest): Promise<Device> {
  return client.post('/devices', data)
}

export function updateDevice(id: number, data: DeviceUpdateRequest): Promise<Device> {
  return client.put(`/devices/${id}`, data)
}

export function deleteDevice(id: number): Promise<void> {
  return client.delete(`/devices/${id}`)
}

export function getDeviceStatus(id: number): Promise<DeviceStatus> {
  return client.get(`/devices/${id}/status`)
}
