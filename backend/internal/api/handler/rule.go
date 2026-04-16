package handler

import (
	"encoding/json"
	"net/http"

	"iot-admin/internal/model"
	"iot-admin/internal/store/sqlite"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type RuleHandler struct {
	ruleStore *sqlite.RuleStore
}

func NewRuleHandler(ruleStore *sqlite.RuleStore) *RuleHandler {
	return &RuleHandler{ruleStore: ruleStore}
}

func (h *RuleHandler) List(c *gin.Context) {
	rules, err := h.ruleStore.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data":  rules,
		"total": len(rules),
	})
}

func (h *RuleHandler) Get(c *gin.Context) {
	id := c.Param("id")
	rule, err := h.ruleStore.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if rule == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Rule not found"})
		return
	}

	c.JSON(http.StatusOK, rule)
}

type CreateRuleRequest struct {
	Name         string `json:"name" binding:"required"`
	Description  string `json:"description"`
	TopicPattern string `json:"topic_pattern" binding:"required"`
	Condition    string `json:"condition" binding:"required"`
	ActionType   string `json:"action_type" binding:"required"`
	ActionConfig string `json:"action_config" binding:"required"`
	CooldownSecs int    `json:"cooldown_secs"`
	Enabled      *bool  `json:"enabled"`
}

func (h *RuleHandler) Create(c *gin.Context) {
	var req CreateRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate condition JSON
	var condMap map[string]interface{}
	if err := json.Unmarshal([]byte(req.Condition), &condMap); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid condition JSON"})
		return
	}

	// Validate action_config JSON
	var actionMap map[string]interface{}
	if err := json.Unmarshal([]byte(req.ActionConfig), &actionMap); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid action_config JSON"})
		return
	}

	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}
	cooldown := req.CooldownSecs
	if cooldown <= 0 {
		cooldown = 60
	}

	rule := &model.Rule{
		ID:           uuid.New().String(),
		Name:         req.Name,
		Description:  req.Description,
		Enabled:      enabled,
		TopicPattern: req.TopicPattern,
		Condition:    req.Condition,
		ActionType:   req.ActionType,
		ActionConfig: req.ActionConfig,
		CooldownSecs: cooldown,
	}

	if err := h.ruleStore.Create(rule); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, rule)
}

type UpdateRuleRequest struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	TopicPattern string `json:"topic_pattern"`
	Condition    string `json:"condition"`
	ActionType   string `json:"action_type"`
	ActionConfig string `json:"action_config"`
	CooldownSecs *int   `json:"cooldown_secs"`
	Enabled      *bool  `json:"enabled"`
}

func (h *RuleHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req UpdateRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	rule, err := h.ruleStore.GetByID(id)
	if err != nil || rule == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Rule not found"})
		return
	}

	if req.Name != "" {
		rule.Name = req.Name
	}
	rule.Description = req.Description
	if req.TopicPattern != "" {
		rule.TopicPattern = req.TopicPattern
	}
	if req.Condition != "" {
		rule.Condition = req.Condition
	}
	if req.ActionType != "" {
		rule.ActionType = req.ActionType
	}
	if req.ActionConfig != "" {
		rule.ActionConfig = req.ActionConfig
	}
	if req.CooldownSecs != nil {
		rule.CooldownSecs = *req.CooldownSecs
	}
	if req.Enabled != nil {
		rule.Enabled = *req.Enabled
	}

	if err := h.ruleStore.Update(rule); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Rule updated"})
}

func (h *RuleHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.ruleStore.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Rule deleted"})
}

type EnableRuleRequest struct {
	Enabled bool `json:"enabled"`
}

func (h *RuleHandler) SetEnabled(c *gin.Context) {
	id := c.Param("id")
	var req EnableRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.ruleStore.SetEnabled(id, req.Enabled); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Rule status updated"})
}

func (h *RuleHandler) Logs(c *gin.Context) {
	id := c.Param("id")
	page := getIntQuery(c, "page", 1)
	pageSize := getIntQuery(c, "page_size", 20)

	logs, total, err := h.ruleStore.ListLogs(id, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      logs,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func getIntQuery(c *gin.Context, key string, defaultVal int) int {
	val := c.Query(key)
	if val == "" {
		return defaultVal
	}
	var result int
	if _, err := parseInt(val); err == nil {
		result, _ = parseInt(val)
	} else {
		result = defaultVal
	}
	return result
}

func parseInt(s string) (int, error) {
	var result int
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, nil
		}
		result = result*10 + int(c-'0')
	}
	return result, nil
}
