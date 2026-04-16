import client from './client'
import type { Tag, TagCreateRequest } from '@/types'

export function getTags(): Promise<Tag[]> {
  return client.get('/tags')
}

export function createTag(data: TagCreateRequest): Promise<Tag> {
  return client.post('/tags', data)
}

export function updateTag(id: number, data: Partial<TagCreateRequest>): Promise<Tag> {
  return client.put(`/tags/${id}`, data)
}

export function deleteTag(id: number): Promise<void> {
  return client.delete(`/tags/${id}`)
}
