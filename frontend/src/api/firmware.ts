import client from './client'
import type {
  Firmware,
  OTAUpgrade,
  OTACreateRequest,
  OTABatchRequest,
  PaginatedResponse,
} from '@/types'

export function getFirmwareList(): Promise<Firmware[]> {
  return client.get('/firmware')
}

export function uploadFirmware(formData: FormData): Promise<Firmware> {
  return client.post('/firmware', formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
  })
}

export function deleteFirmware(id: number): Promise<void> {
  return client.delete(`/firmware/${id}`)
}

export function downloadFirmware(id: number): string {
  return `/api/v1/firmware/${id}/download`
}

export function getOTAList(params?: { page?: number; page_size?: number }): Promise<PaginatedResponse<OTAUpgrade>> {
  return client.get('/ota', { params })
}

export function createOTAUpgrade(data: OTACreateRequest): Promise<OTAUpgrade> {
  return client.post('/ota', data)
}

export function createOTABatchUpgrade(data: OTABatchRequest): Promise<OTAUpgrade[]> {
  return client.post('/ota/batch', data)
}

export function updateOTAStatus(id: number, status: string): Promise<OTAUpgrade> {
  return client.put(`/ota/${id}/status`, { status })
}

export function deleteOTA(id: number): Promise<void> {
  return client.delete(`/ota/${id}`)
}
