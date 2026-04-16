package sqlite

import (
	"database/sql"
	"fmt"
	"strings"

	"iot-admin/internal/model"
)

type MessageStore struct {
	db *sql.DB
}

func NewMessageStore(db *sql.DB) *MessageStore {
	return &MessageStore{db: db}
}

func (s *MessageStore) Insert(m *model.Message) error {
	_, err := s.db.Exec(
		"INSERT INTO message_history (device_id, topic, payload, qos, direction) VALUES (?, ?, ?, ?, ?)",
		m.DeviceID, m.Topic, m.Payload, m.QoS, m.Direction,
	)
	return err
}

func (s *MessageStore) List(q *model.ListQuery, startTime, endTime string) ([]model.Message, int64, error) {
	conditions := []string{"1=1"}
	args := []interface{}{}

	if q.Search != "" {
		conditions = append(conditions, "topic LIKE ?")
		args = append(args, "%"+q.Search+"%")
	}
	if q.DeviceID != "" {
		conditions = append(conditions, "(device_id = ? OR topic LIKE ?)")
		args = append(args, q.DeviceID, "%"+q.DeviceID+"%")
	}
	if startTime != "" {
		conditions = append(conditions, "created_at >= ?")
		args = append(args, startTime)
	}
	if endTime != "" {
		conditions = append(conditions, "created_at <= ?")
		args = append(args, endTime)
	}

	where := strings.Join(conditions, " AND ")

	var total int64
	if err := s.db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM message_history WHERE %s", where), args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	pageSize := q.PageSize
	if pageSize <= 0 {
		pageSize = 50
	}
	page := q.Page
	if page <= 0 {
		page = 1
	}

	querySQL := fmt.Sprintf(
		"SELECT id, device_id, topic, payload, qos, direction, created_at FROM message_history WHERE %s ORDER BY created_at DESC LIMIT ? OFFSET ?",
		where,
	)
	pagArgs := append(args, pageSize, (page-1)*pageSize)

	rows, err := s.db.Query(querySQL, pagArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	messages := []model.Message{}
	for rows.Next() {
		var m model.Message
		if err := rows.Scan(&m.ID, &m.DeviceID, &m.Topic, &m.Payload, &m.QoS, &m.Direction, &m.CreatedAt); err != nil {
			return nil, 0, err
		}
		messages = append(messages, m)
	}
	return messages, total, nil
}

func (s *MessageStore) GetTopics() ([]string, error) {
	rows, err := s.db.Query("SELECT DISTINCT topic FROM message_history ORDER BY topic LIMIT 500")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	topics := []string{}
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err != nil {
			return nil, err
		}
		topics = append(topics, t)
	}
	return topics, nil
}

func (s *MessageStore) Count() (int64, error) {
	var count int64
	err := s.db.QueryRow("SELECT COUNT(*) FROM message_history").Scan(&count)
	return count, err
}

func (s *MessageStore) CountSince(since string) (int64, error) {
	var count int64
	err := s.db.QueryRow("SELECT COUNT(*) FROM message_history WHERE created_at >= ?", since).Scan(&count)
	return count, err
}

// GetThroughput returns message counts per time bucket
func (s *MessageStore) GetThroughput(bucketMinutes int, since string) ([]map[string]interface{}, error) {
	query := fmt.Sprintf(
		"SELECT strftime('%%Y-%%m-%%dT%%H:%%M:00', created_at) as bucket, COUNT(*) as count FROM message_history WHERE created_at >= ? GROUP BY bucket ORDER BY bucket",
	)
	if bucketMinutes >= 60 {
		query = fmt.Sprintf(
			"SELECT strftime('%%Y-%%m-%%dT%%H:00:00', created_at) as bucket, COUNT(*) as count FROM message_history WHERE created_at >= ? GROUP BY bucket ORDER BY bucket",
		)
	}

	rows, err := s.db.Query(query, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := []map[string]interface{}{}
	for rows.Next() {
		var bucket string
		var count int64
		if err := rows.Scan(&bucket, &count); err != nil {
			return nil, err
		}
		result = append(result, map[string]interface{}{
			"time":  bucket,
			"count": count,
		})
	}
	return result, nil
}
