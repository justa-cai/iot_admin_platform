package sqlite

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"iot-admin/internal/model"
)

type FirmwareStore struct {
	db *sql.DB
}

func NewFirmwareStore(db *sql.DB) *FirmwareStore {
	return &FirmwareStore{db: db}
}

func (s *FirmwareStore) Create(f *model.Firmware) error {
	_, err := s.db.Exec(
		"INSERT INTO firmwares (id, name, version, device_model, file_path, file_size, checksum, description) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		f.ID, f.Name, f.Version, f.DeviceModel, f.FilePath, f.FileSize, f.Checksum, f.Description,
	)
	return err
}

func (s *FirmwareStore) GetByID(id string) (*model.Firmware, error) {
	f := &model.Firmware{}
	err := s.db.QueryRow(
		"SELECT id, name, version, device_model, file_path, file_size, checksum, description, created_at FROM firmwares WHERE id=?", id,
	).Scan(&f.ID, &f.Name, &f.Version, &f.DeviceModel, &f.FilePath, &f.FileSize, &f.Checksum, &f.Description, &f.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return f, err
}

func (s *FirmwareStore) List(q *model.ListQuery) ([]model.Firmware, int64, error) {
	conditions := []string{"1=1"}
	args := []interface{}{}

	if q.Search != "" {
		conditions = append(conditions, "(name LIKE ? OR device_model LIKE ?)")
		args = append(args, "%"+q.Search+"%", "%"+q.Search+"%")
	}

	where := strings.Join(conditions, " AND ")

	var total int64
	if err := s.db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM firmwares WHERE %s", where), args...).Scan(&total); err != nil {
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

	querySQL := fmt.Sprintf(
		"SELECT id, name, version, device_model, file_path, file_size, checksum, description, created_at FROM firmwares WHERE %s ORDER BY created_at DESC LIMIT ? OFFSET ?",
		where,
	)
	pagArgs := append(args, pageSize, (page-1)*pageSize)

	rows, err := s.db.Query(querySQL, pagArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	list := []model.Firmware{}
	for rows.Next() {
		var f model.Firmware
		if err := rows.Scan(&f.ID, &f.Name, &f.Version, &f.DeviceModel, &f.FilePath, &f.FileSize, &f.Checksum, &f.Description, &f.CreatedAt); err != nil {
			return nil, 0, err
		}
		list = append(list, f)
	}
	return list, total, nil
}

func (s *FirmwareStore) Delete(id string) error {
	_, err := s.db.Exec("DELETE FROM firmwares WHERE id=?", id)
	return err
}

// OTA Upgrade operations

type OTAStore struct {
	db *sql.DB
}

func NewOTAStore(db *sql.DB) *OTAStore {
	return &OTAStore{db: db}
}

func (s *OTAStore) Create(o *model.OTAUpgrade) error {
	_, err := s.db.Exec(
		"INSERT INTO ota_upgrades (id, device_id, firmware_id, status) VALUES (?, ?, ?, ?)",
		o.ID, o.DeviceID, o.FirmwareID, o.Status,
	)
	return err
}

func (s *OTAStore) GetByID(id string) (*model.OTAUpgrade, error) {
	o := &model.OTAUpgrade{}
	err := s.db.QueryRow(
		"SELECT id, device_id, firmware_id, status, progress, error_msg, created_at, completed_at FROM ota_upgrades WHERE id=?", id,
	).Scan(&o.ID, &o.DeviceID, &o.FirmwareID, &o.Status, &o.Progress, &o.ErrorMsg, &o.CreatedAt, &o.CompletedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return o, err
}

func (s *OTAStore) List(q *model.ListQuery) ([]model.OTAUpgrade, int64, error) {
	conditions := []string{"1=1"}
	args := []interface{}{}

	if q.Status != "" {
		conditions = append(conditions, "o.status = ?")
		args = append(args, q.Status)
	}
	if q.DeviceID != "" {
		conditions = append(conditions, "o.device_id = ?")
		args = append(args, q.DeviceID)
	}

	where := strings.Join(conditions, " AND ")

	var total int64
	if err := s.db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM ota_upgrades o WHERE %s", where), args...).Scan(&total); err != nil {
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

	querySQL := fmt.Sprintf(
		"SELECT o.id, o.device_id, o.firmware_id, o.status, o.progress, o.error_msg, o.created_at, o.completed_at, COALESCE(d.name,''), COALESCE(f.name,''), COALESCE(f.version,'') FROM ota_upgrades o LEFT JOIN devices d ON o.device_id = d.id LEFT JOIN firmwares f ON o.firmware_id = f.id WHERE %s ORDER BY o.created_at DESC LIMIT ? OFFSET ?",
		where,
	)
	pagArgs := append(args, pageSize, (page-1)*pageSize)

	rows, err := s.db.Query(querySQL, pagArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	list := []model.OTAUpgrade{}
	for rows.Next() {
		var o model.OTAUpgrade
		if err := rows.Scan(&o.ID, &o.DeviceID, &o.FirmwareID, &o.Status, &o.Progress, &o.ErrorMsg, &o.CreatedAt, &o.CompletedAt, &o.DeviceName, &o.FirmwareName, &o.Version); err != nil {
			return nil, 0, err
		}
		list = append(list, o)
	}
	return list, total, nil
}

func (s *OTAStore) UpdateStatus(id, status string, progress int, errMsg string) error {
	if status == "success" || status == "failed" {
		_, err := s.db.Exec(
			"UPDATE ota_upgrades SET status=?, progress=?, error_msg=?, completed_at=? WHERE id=?",
			status, progress, errMsg, time.Now(), id,
		)
		return err
	}
	_, err := s.db.Exec(
		"UPDATE ota_upgrades SET status=?, progress=?, error_msg=? WHERE id=?",
		status, progress, errMsg, id,
	)
	return err
}

func (s *OTAStore) GetPendingByDevice(deviceID string) (*model.OTAUpgrade, error) {
	o := &model.OTAUpgrade{}
	err := s.db.QueryRow(
		"SELECT id, device_id, firmware_id, status, progress, error_msg, created_at, completed_at FROM ota_upgrades WHERE device_id=? AND status IN ('pending', 'downloading', 'installing') ORDER BY created_at DESC LIMIT 1",
		deviceID,
	).Scan(&o.ID, &o.DeviceID, &o.FirmwareID, &o.Status, &o.Progress, &o.ErrorMsg, &o.CreatedAt, &o.CompletedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return o, err
}

func (s *OTAStore) Delete(id string) error {
	_, err := s.db.Exec("DELETE FROM ota_upgrades WHERE id=?", id)
	return err
}
