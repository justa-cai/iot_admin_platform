package sqlite

import (
	"database/sql"
	"fmt"
	"time"

	"iot-admin/internal/model"
)

type UserStore struct {
	db *sql.DB
}

func NewUserStore(db *sql.DB) *UserStore {
	return &UserStore{db: db}
}

func (s *UserStore) Create(user *model.User) error {
	_, err := s.db.Exec(
		"INSERT INTO users (id, username, password_hash, role, status) VALUES (?, ?, ?, ?, ?)",
		user.ID, user.Username, user.PasswordHash, user.Role, user.Status,
	)
	return err
}

func (s *UserStore) GetByID(id string) (*model.User, error) {
	user := &model.User{}
	err := s.db.QueryRow(
		"SELECT id, username, password_hash, role, status, last_login_at, created_at, updated_at FROM users WHERE id = ?",
		id,
	).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.Role, &user.Status, &user.LastLoginAt, &user.CreatedAt, &user.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return user, err
}

func (s *UserStore) GetByUsername(username string) (*model.User, error) {
	user := &model.User{}
	err := s.db.QueryRow(
		"SELECT id, username, password_hash, role, status, last_login_at, created_at, updated_at FROM users WHERE username = ?",
		username,
	).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.Role, &user.Status, &user.LastLoginAt, &user.CreatedAt, &user.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return user, err
}

func (s *UserStore) List(q *model.ListQuery) ([]model.User, int64, error) {
	var total int64
	countSQL := "SELECT COUNT(*) FROM users WHERE 1=1"
	args := []interface{}{}

	if q.Search != "" {
		countSQL += " AND username LIKE ?"
		args = append(args, "%"+q.Search+"%")
	}
	if q.Status != "" {
		countSQL += " AND status = ?"
		args = append(args, q.Status)
	}

	if err := s.db.QueryRow(countSQL, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	querySQL := "SELECT id, username, role, status, last_login_at, created_at, updated_at FROM users WHERE 1=1"
	queryArgs := make([]interface{}, len(args))
	copy(queryArgs, args)

	if q.Search != "" {
		querySQL += " AND username LIKE ?"
		queryArgs = append(queryArgs, "%"+q.Search+"%")
	}
	if q.Status != "" {
		querySQL += " AND status = ?"
		queryArgs = append(queryArgs, q.Status)
	}

	querySQL += " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	pageSize := q.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	page := q.Page
	if page <= 0 {
		page = 1
	}
	queryArgs = append(queryArgs, pageSize, (page-1)*pageSize)

	rows, err := s.db.Query(querySQL, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	users := []model.User{}
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.Username, &u.Role, &u.Status, &u.LastLoginAt, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, 0, err
		}
		users = append(users, u)
	}
	return users, total, nil
}

func (s *UserStore) Update(user *model.User) error {
	_, err := s.db.Exec(
		"UPDATE users SET username=?, role=?, status=?, updated_at=? WHERE id=?",
		user.Username, user.Role, user.Status, time.Now(), user.ID,
	)
	return err
}

func (s *UserStore) UpdateLastLogin(id string) error {
	_, err := s.db.Exec("UPDATE users SET last_login_at=?, updated_at=? WHERE id=?", time.Now(), time.Now(), id)
	return err
}

func (s *UserStore) Delete(id string) error {
	_, err := s.db.Exec("DELETE FROM users WHERE id=?", id)
	return err
}

func (s *UserStore) UpdatePassword(id, passwordHash string) error {
	_, err := s.db.Exec("UPDATE users SET password_hash=?, updated_at=? WHERE id=?", passwordHash, time.Now(), id)
	return err
}

func (s *UserStore) UpdateRole(id, role string) error {
	_, err := s.db.Exec("UPDATE users SET role=?, updated_at=? WHERE id=?", role, time.Now(), id)
	return err
}

func (s *UserStore) Count() (int, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	return count, err
}

func (s *UserStore) EnsureAdmin(username, passwordHash string) (string, error) {
	user, err := s.GetByUsername(username)
	if err != nil {
		return "", err
	}
	if user != nil {
		return user.ID, nil
	}

	var id string
	err = s.db.QueryRow(`
		INSERT INTO users (id, username, password_hash, role, status)
		VALUES (?, ?, ?, 'admin', 'active')
		ON CONFLICT(username) DO UPDATE SET updated_at=updated_at
		RETURNING id
	`, fmt.Sprintf("admin-%d", time.Now().UnixNano()), username, passwordHash).Scan(&id)
	return id, err
}
