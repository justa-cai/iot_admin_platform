import client from './client'
import type { Group, GroupCreateRequest } from '@/types'

export function getGroups(): Promise<Group[]> {
  return client.get('/groups')
}

export function getGroup(id: number): Promise<Group> {
  return client.get(`/groups/${id}`)
}

export function createGroup(data: GroupCreateRequest): Promise<Group> {
  return client.post('/groups', data)
}

export function updateGroup(id: number, data: Partial<GroupCreateRequest>): Promise<Group> {
  return client.put(`/groups/${id}`, data)
}

export function deleteGroup(id: number): Promise<void> {
  return client.delete(`/groups/${id}`)
}

export function addDevicesToGroup(groupId: number, deviceIds: number[]): Promise<void> {
  return client.post(`/groups/${groupId}/devices`, { device_ids: deviceIds })
}

export function removeDevicesFromGroup(groupId: number, deviceIds: number[]): Promise<void> {
  return client.delete(`/groups/${groupId}/devices`, { data: { device_ids: deviceIds } })
}
