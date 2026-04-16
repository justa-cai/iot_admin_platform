package sqlite

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"iot-admin/internal/model"
)

type DeviceStore struct {
	db *sql.DB
}

func NewDeviceStore(db *sql.DB) *DeviceStore {
	return &DeviceStore{db: db}
}

func (s *DeviceStore) Create(d *model.Device) error {
	_, err := s.db.Exec(
		"INSERT INTO devices (id, name, device_key, device_secret, metadata) VALUES (?, ?, ?, ?, ?)",
		d.ID, d.Name, d.DeviceKey, d.DeviceSecret, d.Metadata,
	)
	return err
}

func (s *DeviceStore) GetByID(id string) (*model.Device, error) {
	d := &model.Device{}
	err := s.db.QueryRow(
		"SELECT id, name, device_key, device_secret, status, metadata, last_seen_at, created_at, updated_at FROM devices WHERE id=?",
		id,
	).Scan(&d.ID, &d.Name, &d.DeviceKey, &d.DeviceSecret, &d.Status, &d.Metadata, &d.LastSeenAt, &d.CreatedAt, &d.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	s.loadRelations(d)
	return d, nil
}

func (s *DeviceStore) GetByKey(deviceKey string) (*model.Device, error) {
	d := &model.Device{}
	err := s.db.QueryRow(
		"SELECT id, name, device_key, device_secret, status, metadata, last_seen_at, created_at, updated_at FROM devices WHERE device_key=?",
		deviceKey,
	).Scan(&d.ID, &d.Name, &d.DeviceKey, &d.DeviceSecret, &d.Status, &d.Metadata, &d.LastSeenAt, &d.CreatedAt, &d.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return d, nil
}

func (s *DeviceStore) List(q *model.ListQuery) ([]model.Device, int64, error) {
	conditions := []string{"1=1"}
	args := []interface{}{}

	if q.Search != "" {
		conditions = append(conditions, "(d.name LIKE ? OR d.device_key LIKE ?)")
		args = append(args, "%"+q.Search+"%", "%"+q.Search+"%")
	}
	if q.Status != "" {
		conditions = append(conditions, "d.status = ?")
		args = append(args, q.Status)
	}
	if q.GroupID != "" {
		conditions = append(conditions, "EXISTS (SELECT 1 FROM device_groups dg WHERE dg.device_id = d.id AND dg.group_id = ?)")
		args = append(args, q.GroupID)
	}
	if q.TagID != "" {
		conditions = append(conditions, "EXISTS (SELECT 1 FROM device_tags dt WHERE dt.device_id = d.id AND dt.tag_id = ?)")
		args = append(args, q.TagID)
	}

	where := strings.Join(conditions, " AND ")

	var total int64
	countSQL := fmt.Sprintf("SELECT COUNT(*) FROM devices d WHERE %s", where)
	if err := s.db.QueryRow(countSQL, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	pageSize := q.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	page := q.Page
	if page <= 0 {
		page = 1
	}

	sortBy := "d.created_at"
	sortDir := "DESC"
	if q.SortBy != "" {
		sortBy = "d." + q.SortBy
		if q.SortDir != "" {
			sortDir = strings.ToUpper(q.SortDir)
		}
	}

	querySQL := fmt.Sprintf(
		"SELECT d.id, d.name, d.device_key, d.status, d.metadata, d.last_seen_at, d.created_at, d.updated_at FROM devices d WHERE %s ORDER BY %s %s LIMIT ? OFFSET ?",
		where, sortBy, sortDir,
	)
	pagArgs := append(args, pageSize, (page-1)*pageSize)

	rows, err := s.db.Query(querySQL, pagArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	devices := []model.Device{}
	for rows.Next() {
		var d model.Device
		if err := rows.Scan(&d.ID, &d.Name, &d.DeviceKey, &d.Status, &d.Metadata, &d.LastSeenAt, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, 0, err
		}
		devices = append(devices, d)
	}

	// Load relations for each device
	for i := range devices {
		s.loadRelations(&devices[i])
	}

	return devices, total, nil
}

func (s *DeviceStore) Update(d *model.Device) error {
	_, err := s.db.Exec(
		"UPDATE devices SET name=?, metadata=?, updated_at=? WHERE id=?",
		d.Name, d.Metadata, time.Now(), d.ID,
	)
	return err
}

func (s *DeviceStore) Delete(id string) error {
	_, err := s.db.Exec("DELETE FROM devices WHERE id=?", id)
	return err
}

func (s *DeviceStore) UpdateStatus(id, status string) error {
	_, err := s.db.Exec(
		"UPDATE devices SET status=?, last_seen_at=?, updated_at=? WHERE id=?",
		status, time.Now(), time.Now(), id,
	)
	return err
}

func (s *DeviceStore) Count() (int, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM devices").Scan(&count)
	return count, err
}

func (s *DeviceStore) CountByStatus(status string) (int, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM devices WHERE status=?", status).Scan(&count)
	return count, err
}

func (s *DeviceStore) SetGroups(deviceID string, groupIDs []string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec("DELETE FROM device_groups WHERE device_id=?", deviceID); err != nil {
		return err
	}
	for _, gid := range groupIDs {
		if _, err := tx.Exec("INSERT INTO device_groups (device_id, group_id) VALUES (?, ?)", deviceID, gid); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *DeviceStore) SetTags(deviceID string, tagIDs []string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec("DELETE FROM device_tags WHERE device_id=?", deviceID); err != nil {
		return err
	}
	for _, tid := range tagIDs {
		if _, err := tx.Exec("INSERT INTO device_tags (device_id, tag_id) VALUES (?, ?)", deviceID, tid); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *DeviceStore) loadRelations(d *model.Device) {
	// Load groups
	rows, err := s.db.Query(
		"SELECT g.id, g.name, g.description, g.parent_id, g.created_at FROM groups g INNER JOIN device_groups dg ON g.id = dg.group_id WHERE dg.device_id = ?",
		d.ID,
	)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var g model.Group
			if rows.Scan(&g.ID, &g.Name, &g.Description, &g.ParentID, &g.CreatedAt) == nil {
				d.Groups = append(d.Groups, g)
			}
		}
	}

	// Load tags
	rows2, err := s.db.Query(
		"SELECT t.id, t.name, t.color, t.created_at FROM tags t INNER JOIN device_tags dt ON t.id = dt.tag_id WHERE dt.device_id = ?",
		d.ID,
	)
	if err == nil {
		defer rows2.Close()
		for rows2.Next() {
			var t model.Tag
			if rows2.Scan(&t.ID, &t.Name, &t.Color, &t.CreatedAt) == nil {
				d.Tags = append(d.Tags, t)
			}
		}
	}
}

// ListWithMetadata returns device list with metadata parsed
func (s *DeviceStore) ListWithMetadata(q *model.ListQuery) ([]model.Device, int64, error) {
	devices, total, err := s.List(q)
	if err != nil {
		return nil, 0, err
	}
	// Ensure metadata is valid JSON
	for i := range devices {
		if devices[i].Metadata == "" {
			devices[i].Metadata = "{}"
		}
		var raw map[string]interface{}
		if json.Unmarshal([]byte(devices[i].Metadata), &raw) != nil {
			devices[i].Metadata = "{}"
		}
	}
	return devices, total, nil
}
