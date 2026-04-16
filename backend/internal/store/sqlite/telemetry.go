package sqlite

import (
	"database/sql"
	"fmt"
	"strings"

	"iot-admin/internal/model"
)

type TelemetryStore struct {
	db *sql.DB
}

func NewTelemetryStore(db *sql.DB) *TelemetryStore {
	return &TelemetryStore{db: db}
}

func (s *TelemetryStore) Insert(t *model.Telemetry) error {
	_, err := s.db.Exec(
		"INSERT INTO telemetry (device_id, topic, fields) VALUES (?, ?, ?)",
		t.DeviceID, t.Topic, t.Fields,
	)
	return err
}

func (s *TelemetryStore) Query(deviceID, startTime, endTime string, page, pageSize int) ([]model.Telemetry, int64, error) {
	conditions := []string{"1=1"}
	args := []interface{}{}

	if deviceID != "" {
		conditions = append(conditions, "device_id = ?")
		args = append(args, deviceID)
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
	if err := s.db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM telemetry WHERE %s", where), args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	if pageSize <= 0 {
		pageSize = 100
	}
	if page <= 0 {
		page = 1
	}

	querySQL := fmt.Sprintf(
		"SELECT id, device_id, topic, fields, created_at FROM telemetry WHERE %s ORDER BY created_at DESC LIMIT ? OFFSET ?",
		where,
	)
	pagArgs := append(args, pageSize, (page-1)*pageSize)

	rows, err := s.db.Query(querySQL, pagArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	records := []model.Telemetry{}
	for rows.Next() {
		var r model.Telemetry
		if err := rows.Scan(&r.ID, &r.DeviceID, &r.Topic, &r.Fields, &r.CreatedAt); err != nil {
			return nil, 0, err
		}
		records = append(records, r)
	}
	return records, total, nil
}

func (s *TelemetryStore) Latest(deviceID string, limit int) ([]model.Telemetry, error) {
	if limit <= 0 {
		limit = 10
	}
	query := "SELECT id, device_id, topic, fields, created_at FROM telemetry"
	args := []interface{}{limit}

	if deviceID != "" {
		query += " WHERE device_id = ?"
		args = []interface{}{deviceID, limit}
	}
	query += " ORDER BY created_at DESC LIMIT ?"

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	records := []model.Telemetry{}
	for rows.Next() {
		var r model.Telemetry
		if err := rows.Scan(&r.ID, &r.DeviceID, &r.Topic, &r.Fields, &r.CreatedAt); err != nil {
			return nil, err
		}
		records = append(records, r)
	}
	return records, nil
}

// CleanupOlderThan deletes telemetry older than the given number of days
func (s *TelemetryStore) CleanupOlderThan(days int) (int64, error) {
	result, err := s.db.Exec(
		"DELETE FROM telemetry WHERE created_at < datetime('now', ?)",
		fmt.Sprintf("-%d days", days),
	)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
