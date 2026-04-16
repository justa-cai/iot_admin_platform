package model

import "time"

// User represents a system user
type User struct {
	ID           string     `json:"id"`
	Username     string     `json:"username"`
	PasswordHash string     `json:"-"`
	Role         string     `json:"role"` // admin, operator, viewer
	Status       string     `json:"status"`
	LastLoginAt  *time.Time `json:"last_login_at"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// Device represents an IoT device
type Device struct {
	ID           string     `json:"id"`
	Name         string     `json:"name"`
	DeviceKey    string     `json:"device_key"`
	DeviceSecret string     `json:"-" `
	Status       string     `json:"status"` // online, offline
	Metadata     string     `json:"metadata"`
	LastSeenAt   *time.Time `json:"last_seen_at"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	Groups       []Group    `json:"groups,omitempty"`
	Tags         []Tag      `json:"tags,omitempty"`
}

// Group represents a device group
type Group struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	ParentID    *string `json:"parent_id"`
	CreatedAt   time.Time `json:"created_at"`
	DeviceCount int     `json:"device_count,omitempty"`
}

// Tag represents a device tag
type Tag struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Color     string    `json:"color"`
	CreatedAt time.Time `json:"created_at"`
}

// Message represents an MQTT message
type Message struct {
	ID        int64     `json:"id"`
	DeviceID  *string   `json:"device_id"`
	Topic     string    `json:"topic"`
	Payload   string    `json:"payload"`
	QoS      int       `json:"qos"`
	Direction string    `json:"direction"` // inbound, outbound
	CreatedAt time.Time `json:"created_at"`
}

// Telemetry represents a telemetry data point
type Telemetry struct {
	ID        int64     `json:"id"`
	DeviceID  string    `json:"device_id"`
	Topic     string    `json:"topic"`
	Fields    string    `json:"fields"` // JSON
	CreatedAt time.Time `json:"created_at"`
}

// Rule represents a data processing rule
type Rule struct {
	ID            string     `json:"id"`
	Name          string     `json:"name"`
	Description   string     `json:"description"`
	Enabled       bool       `json:"enabled"`
	TopicPattern  string     `json:"topic_pattern"`
	Condition     string     `json:"condition"`     // JSON
	ActionType    string     `json:"action_type"`   // alert, publish, forward
	ActionConfig  string     `json:"action_config"` // JSON
	CooldownSecs  int        `json:"cooldown_secs"`
	LastTriggered *time.Time `json:"last_triggered"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// RuleLog represents a rule execution log
type RuleLog struct {
	ID           int64     `json:"id"`
	RuleID       string    `json:"rule_id"`
	DeviceID     *string   `json:"device_id"`
	Topic        string    `json:"topic"`
	Payload      string    `json:"payload"`
ActionResult  string    `json:"action_result"`
	CreatedAt    time.Time `json:"created_at"`
	RuleName     string    `json:"rule_name,omitempty"`
}

// Firmware represents a firmware version
type Firmware struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Version     string    `json:"version"`
	DeviceModel string    `json:"device_model"` // applicable device model
	FilePath    string    `json:"file_path"`
	FileSize    int64     `json:"file_size"`
	Checksum    string    `json:"checksum"` // SHA256
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

// OTAUpgrade represents an OTA upgrade task
type OTAUpgrade struct {
	ID           string     `json:"id"`
	DeviceID     string     `json:"device_id"`
	FirmwareID   string     `json:"firmware_id"`
	Status       string     `json:"status"` // pending, downloading, installing, success, failed
	Progress     int        `json:"progress"`
	ErrorMsg     string     `json:"error_msg,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	CompletedAt  *time.Time `json:"completed_at"`
	DeviceName   string     `json:"device_name,omitempty"`
	FirmwareName string     `json:"firmware_name,omitempty"`
	Version      string     `json:"version,omitempty"`
}

// DashboardStats represents dashboard statistics
type DashboardStats struct {
	TotalDevices  int `json:"total_devices"`
	OnlineDevices int `json:"online_devices"`
	TotalMessages int64 `json:"total_messages"`
	ActiveRules   int `json:"active_rules"`
	TotalUsers    int `json:"total_users"`
}

// Pagination represents pagination parameters
type Pagination struct {
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
	Total    int64  `json:"total"`
}

// ListQuery represents common list query parameters
type ListQuery struct {
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
	Search   string `form:"search"`
	Status   string `form:"status"`
	GroupID  string `form:"group_id"`
	TagID    string `form:"tag_id"`
	DeviceID string `form:"device_id"`
	SortBy   string `form:"sort_by"`
	SortDir  string `form:"sort_dir"`
}
