import client from './client'
import type {
  Message,
  MessagePublishRequest,
  TopicInfo,
  PaginatedResponse,
} from '@/types'

export interface MessageHistoryParams {
  page?: number
  page_size?: number
  topic?: string
  device_id?: number
  direction?: string
  start_time?: string
  end_time?: string
}

export function publishMessage(data: MessagePublishRequest): Promise<Message> {
  return client.post('/messages/publish', data)
}

export function getMessageHistory(params?: MessageHistoryParams): Promise<PaginatedResponse<Message>> {
  return client.get('/messages/history', { params })
}

export function getTopics(): Promise<TopicInfo[]> {
  return client.get('/messages/topics')
}
