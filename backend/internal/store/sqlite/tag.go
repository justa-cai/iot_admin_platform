package sqlite

import (
	"database/sql"

	"iot-admin/internal/model"
)

type TagStore struct {
	db *sql.DB
}

func NewTagStore(db *sql.DB) *TagStore {
	return &TagStore{db: db}
}

func (s *TagStore) Create(t *model.Tag) error {
	_, err := s.db.Exec("INSERT INTO tags (id, name, color) VALUES (?, ?, ?)", t.ID, t.Name, t.Color)
	return err
}

func (s *TagStore) GetByID(id string) (*model.Tag, error) {
	t := &model.Tag{}
	err := s.db.QueryRow("SELECT id, name, color, created_at FROM tags WHERE id=?", id).Scan(&t.ID, &t.Name, &t.Color, &t.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return t, err
}

func (s *TagStore) List() ([]model.Tag, error) {
	rows, err := s.db.Query("SELECT id, name, color, created_at FROM tags ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tags := []model.Tag{}
	for rows.Next() {
		var t model.Tag
		if err := rows.Scan(&t.ID, &t.Name, &t.Color, &t.CreatedAt); err != nil {
			return nil, err
		}
		tags = append(tags, t)
	}
	return tags, nil
}

func (s *TagStore) Update(t *model.Tag) error {
	_, err := s.db.Exec("UPDATE tags SET name=?, color=? WHERE id=?", t.Name, t.Color, t.ID)
	return err
}

func (s *TagStore) Delete(id string) error {
	_, err := s.db.Exec("DELETE FROM tags WHERE id=?", id)
	return err
}
