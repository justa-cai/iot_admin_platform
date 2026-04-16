// ====== Auth ======
export interface User {
  id: number
  username: string
  email: string
  role: string
  status: string
  created_at: string
  updated_at: string
}

export interface LoginRequest {
  username: string
  password: string
}

export interface LoginResponse {
  token: string
  user: User
}

export interface RegisterRequest {
  username: string
  password: string
  email: string
}

// ====== Device ======
export interface Device {
  id: number
  name: string
  device_key: string
  status: 'online' | 'offline' | 'inactive'
  metadata: Record<string, unknown>
  group_id?: number
  group?: Group
  tags?: Tag[]
  last_active_at?: string
  created_at: string
  updated_at: string
}

export interface DeviceCreateRequest {
  name: string
  device_key: string
  metadata?: Record<string, unknown>
  group_id?: number
  tag_ids?: number[]
}

export interface DeviceUpdateRequest {
  name?: string
  metadata?: Record<string, unknown>
  group_id?: number
  tag_ids?: number[]
}

export interface DeviceStatus {
  device_id: number
  status: string
  last_active_at: string
  ip_address?: string
  firmware_version?: string
}

// ====== Group ======
export interface Group {
  id: number
  name: string
  description?: string
  parent_id?: number
  children?: Group[]
  device_count?: number
  created_at: string
  updated_at: string
}

export interface GroupCreateRequest {
  name: string
  description?: string
  parent_id?: number
}

// ====== Tag ======
export interface Tag {
  id: number
  name: string
  color?: string
  created_at: string
}

export interface TagCreateRequest {
  name: string
  color?: string
}

// ====== Message ======
export interface Message {
  id: number
  device_id?: number
  device_name?: string
  topic: string
  payload: string
  qos: number
  direction: 'in' | 'out'
  created_at: string
}

export interface MessagePublishRequest {
  topic: string
  payload: string
  qos?: number
  device_id?: number
}

export interface TopicInfo {
  topic: string
  subscriber_count: number
}

// ====== Rule ======
export interface Rule {
  id: string
  name: string
  description?: string
  topic_pattern: string
  condition: string
  action_type: string
  action_config: string
  cooldown_secs: number
  enabled: boolean
  last_triggered?: string
  created_at: string
  updated_at: string
}

export interface RuleCondition {
  field: string
  operator: 'eq' | 'neq' | 'gt' | 'gte' | 'lt' | 'lte' | 'contains'
  value: string | number
}

export interface RuleCreateRequest {
  name: string
  description?: string
  topic_pattern: string
  condition: string
  action_type: string
  action_config: string
  cooldown_secs?: number
  enabled?: boolean
}

export interface RuleLog {
  id: number
  rule_id: number
  rule_name: string
  trigger_data: string
  action_result: string
  created_at: string
}

// ====== Dashboard ======
export interface DashboardStats {
  total_devices: number
  online_devices: number
  offline_devices: number
  inactive_devices: number
  messages_today: number
  total_rules: number
  active_rules: number
  alerts_today: number
}

export interface ThroughputData {
  timestamps: string[]
  values: number[]
}

// ====== Telemetry ======
export interface TelemetryData {
  id: number
  device_id: number
  field: string
  value: number
  created_at: string
}

export interface TelemetryQueryParams {
  device_id?: number
  start_time?: string
  end_time?: string
  page?: number
  page_size?: number
}

export interface TelemetryLatest {
  device_id: number
  field: string
  value: number
  timestamp: string
}

// ====== Firmware ======
export interface Firmware {
  id: number
  name: string
  version: string
  file_name: string
  file_size: number
  description?: string
  created_at: string
}

export interface OTAUpgrade {
  id: number
  device_id: number
  device_name?: string
  firmware_id: number
  firmware_name?: string
  firmware_version?: string
  status: 'pending' | 'downloading' | 'installing' | 'success' | 'failed'
  progress: number
  message?: string
  created_at: string
  updated_at: string
}

export interface OTACreateRequest {
  device_id: number
  firmware_id: number
}

export interface OTABatchRequest {
  device_ids: number[]
  firmware_id: number
}

// ====== WebSocket ======
export interface WSEvent {
  event: 'device.status_changed' | 'message.received' | 'rule.triggered'
  data: Record<string, unknown>
  timestamp: string
}

// ====== Pagination ======
export interface PaginatedResponse<T> {
  items: T[]
  total: number
  page: number
  page_size: number
  total_pages: number
}

// ====== API Response ======
export interface ApiResponse<T = unknown> {
  code: number
  message: string
  data: T
}
