package sqlite

import (
	"database/sql"
	"fmt"
	"time"

	"iot-admin/internal/model"
)

type GroupStore struct {
	db *sql.DB
}

func NewGroupStore(db *sql.DB) *GroupStore {
	return &GroupStore{db: db}
}

func (s *GroupStore) Create(g *model.Group) error {
	_, err := s.db.Exec(
		"INSERT INTO groups (id, name, description, parent_id) VALUES (?, ?, ?, ?)",
		g.ID, g.Name, g.Description, g.ParentID,
	)
	return err
}

func (s *GroupStore) GetByID(id string) (*model.Group, error) {
	g := &model.Group{}
	err := s.db.QueryRow(
		"SELECT id, name, description, parent_id, created_at FROM groups WHERE id=?", id,
	).Scan(&g.ID, &g.Name, &g.Description, &g.ParentID, &g.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// Count devices
	s.db.QueryRow("SELECT COUNT(*) FROM device_groups WHERE group_id=?", id).Scan(&g.DeviceCount)
	return g, nil
}

func (s *GroupStore) List() ([]model.Group, error) {
	rows, err := s.db.Query(
		"SELECT g.id, g.name, g.description, g.parent_id, g.created_at, (SELECT COUNT(*) FROM device_groups dg WHERE dg.group_id = g.id) as device_count FROM groups g ORDER BY g.name",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	groups := []model.Group{}
	for rows.Next() {
		var g model.Group
		if err := rows.Scan(&g.ID, &g.Name, &g.Description, &g.ParentID, &g.CreatedAt, &g.DeviceCount); err != nil {
			return nil, err
		}
		groups = append(groups, g)
	}
	return groups, nil
}

func (s *GroupStore) Update(g *model.Group) error {
	_, err := s.db.Exec(
		"UPDATE groups SET name=?, description=?, parent_id=? WHERE id=?",
		g.Name, g.Description, g.ParentID, g.ID,
	)
	return err
}

func (s *GroupStore) Delete(id string) error {
	_, err := s.db.Exec("DELETE FROM groups WHERE id=?", id)
	return err
}

func (s *GroupStore) AddDevices(groupID string, deviceIDs []string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, did := range deviceIDs {
		if _, err := tx.Exec("INSERT OR IGNORE INTO device_groups (device_id, group_id) VALUES (?, ?)", did, groupID); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *GroupStore) RemoveDevices(groupID string, deviceIDs []string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, did := range deviceIDs {
		if _, err := tx.Exec("DELETE FROM device_groups WHERE device_id=? AND group_id=?", did, groupID); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *GroupStore) GetDevices(groupID string) ([]model.Device, error) {
	rows, err := s.db.Query(
		"SELECT d.id, d.name, d.device_key, d.status, d.metadata, d.last_seen_at, d.created_at, d.updated_at FROM devices d INNER JOIN device_groups dg ON d.id = dg.device_id WHERE dg.group_id = ? ORDER BY d.name",
		groupID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	devices := []model.Device{}
	for rows.Next() {
		var d model.Device
		if err := rows.Scan(&d.ID, &d.Name, &d.DeviceKey, &d.Status, &d.Metadata, &d.LastSeenAt, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, err
		}
		devices = append(devices, d)
	}
	return devices, nil
}

func (s *GroupStore) EnsureDefault(name string) (string, error) {
	var id string
	err := s.db.QueryRow(`
		INSERT INTO groups (id, name, description)
		VALUES (?, ?, ?)
		ON CONFLICT(name) DO UPDATE SET description = description
		RETURNING id
	`, fmt.Sprintf("group-%d", time.Now().UnixNano()), name, "Default group").Scan(&id)
	return id, err
}
