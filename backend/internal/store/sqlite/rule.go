package sqlite

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"iot-admin/internal/model"
)

type RuleStore struct {
	db *sql.DB
}

func NewRuleStore(db *sql.DB) *RuleStore {
	return &RuleStore{db: db}
}

func (s *RuleStore) Create(r *model.Rule) error {
	enabledInt := 0
	if r.Enabled {
		enabledInt = 1
	}
	_, err := s.db.Exec(
		"INSERT INTO rules (id, name, description, enabled, topic_pattern, condition, action_type, action_config, cooldown_secs) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		r.ID, r.Name, r.Description, enabledInt, r.TopicPattern, r.Condition, r.ActionType, r.ActionConfig, r.CooldownSecs,
	)
	return err
}

func (s *RuleStore) GetByID(id string) (*model.Rule, error) {
	r := &model.Rule{}
	var enabledInt int
	err := s.db.QueryRow(
		"SELECT id, name, description, enabled, topic_pattern, condition, action_type, action_config, cooldown_secs, last_triggered, created_at, updated_at FROM rules WHERE id=?",
		id,
	).Scan(&r.ID, &r.Name, &r.Description, &enabledInt, &r.TopicPattern, &r.Condition, &r.ActionType, &r.ActionConfig, &r.CooldownSecs, &r.LastTriggered, &r.CreatedAt, &r.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	r.Enabled = enabledInt == 1
	return r, err
}

func (s *RuleStore) List() ([]model.Rule, error) {
	rows, err := s.db.Query(
		"SELECT id, name, description, enabled, topic_pattern, condition, action_type, action_config, cooldown_secs, last_triggered, created_at, updated_at FROM rules ORDER BY created_at DESC",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rules := []model.Rule{}
	for rows.Next() {
		var r model.Rule
		var enabledInt int
		if err := rows.Scan(&r.ID, &r.Name, &r.Description, &enabledInt, &r.TopicPattern, &r.Condition, &r.ActionType, &r.ActionConfig, &r.CooldownSecs, &r.LastTriggered, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, err
		}
		r.Enabled = enabledInt == 1
		rules = append(rules, r)
	}
	return rules, nil
}

func (s *RuleStore) ListEnabled() ([]model.Rule, error) {
	rows, err := s.db.Query(
		"SELECT id, name, description, enabled, topic_pattern, condition, action_type, action_config, cooldown_secs, last_triggered, created_at, updated_at FROM rules WHERE enabled = 1",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rules := []model.Rule{}
	for rows.Next() {
		var r model.Rule
		var enabledInt int
		if err := rows.Scan(&r.ID, &r.Name, &r.Description, &enabledInt, &r.TopicPattern, &r.Condition, &r.ActionType, &r.ActionConfig, &r.CooldownSecs, &r.LastTriggered, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, err
		}
		r.Enabled = true
		rules = append(rules, r)
	}
	return rules, nil
}

func (s *RuleStore) Update(r *model.Rule) error {
	enabledInt := 0
	if r.Enabled {
		enabledInt = 1
	}
	_, err := s.db.Exec(
		"UPDATE rules SET name=?, description=?, enabled=?, topic_pattern=?, condition=?, action_type=?, action_config=?, cooldown_secs=?, updated_at=? WHERE id=?",
		r.Name, r.Description, enabledInt, r.TopicPattern, r.Condition, r.ActionType, r.ActionConfig, r.CooldownSecs, time.Now(), r.ID,
	)
	return err
}

func (s *RuleStore) SetEnabled(id string, enabled bool) error {
	enabledInt := 0
	if enabled {
		enabledInt = 1
	}
	_, err := s.db.Exec("UPDATE rules SET enabled=?, updated_at=? WHERE id=?", enabledInt, time.Now(), id)
	return err
}

func (s *RuleStore) Delete(id string) error {
	_, err := s.db.Exec("DELETE FROM rules WHERE id=?", id)
	return err
}

func (s *RuleStore) UpdateLastTriggered(id string) error {
	_, err := s.db.Exec("UPDATE rules SET last_triggered=?, updated_at=? WHERE id=?", time.Now(), time.Now(), id)
	return err
}

func (s *RuleStore) CountEnabled() (int, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM rules WHERE enabled=1").Scan(&count)
	return count, err
}

// RuleLog operations

func (s *RuleStore) InsertLog(log *model.RuleLog) error {
	_, err := s.db.Exec(
		"INSERT INTO rule_logs (rule_id, device_id, topic, payload, action_result) VALUES (?, ?, ?, ?, ?)",
		log.RuleID, log.DeviceID, log.Topic, log.Payload, log.ActionResult,
	)
	return err
}

func (s *RuleStore) ListLogs(ruleID string, page, pageSize int) ([]model.RuleLog, int64, error) {
	conditions := []string{"1=1"}
	args := []interface{}{}

	if ruleID != "" {
		conditions = append(conditions, "rl.rule_id = ?")
		args = append(args, ruleID)
	}

	where := strings.Join(conditions, " AND ")

	var total int64
	if err := s.db.QueryRow(
		fmt.Sprintf("SELECT COUNT(*) FROM rule_logs rl WHERE %s", where), args...,
	).Scan(&total); err != nil {
		return nil, 0, err
	}

	if pageSize <= 0 {
		pageSize = 20
	}
	if page <= 0 {
		page = 1
	}

	querySQL := fmt.Sprintf(
		"SELECT rl.id, rl.rule_id, rl.device_id, rl.topic, rl.payload, rl.action_result, rl.created_at, COALESCE(r.name, '') as rule_name FROM rule_logs rl LEFT JOIN rules r ON rl.rule_id = r.id WHERE %s ORDER BY rl.created_at DESC LIMIT ? OFFSET ?",
		where,
	)
	pagArgs := append(args, pageSize, (page-1)*pageSize)

	rows, err := s.db.Query(querySQL, pagArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	logs := []model.RuleLog{}
	for rows.Next() {
		var l model.RuleLog
		if err := rows.Scan(&l.ID, &l.RuleID, &l.DeviceID, &l.Topic, &l.Payload, &l.ActionResult, &l.CreatedAt, &l.RuleName); err != nil {
			return nil, 0, err
		}
		logs = append(logs, l)
	}
	return logs, total, nil
}
