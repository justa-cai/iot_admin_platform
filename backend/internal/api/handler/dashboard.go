package handler

import (
	"net/http"
	"time"

	"iot-admin/internal/store/sqlite"

	"github.com/gin-gonic/gin"
)

type DashboardHandler struct {
	deviceStore    *sqlite.DeviceStore
	messageStore   *sqlite.MessageStore
	ruleStore      *sqlite.RuleStore
	userStore      *sqlite.UserStore
	telemetryStore *sqlite.TelemetryStore
}

func NewDashboardHandler(
	deviceStore *sqlite.DeviceStore,
	messageStore *sqlite.MessageStore,
	ruleStore *sqlite.RuleStore,
	userStore *sqlite.UserStore,
	telemetryStore *sqlite.TelemetryStore,
) *DashboardHandler {
	return &DashboardHandler{
		deviceStore:    deviceStore,
		messageStore:   messageStore,
		ruleStore:      ruleStore,
		userStore:      userStore,
		telemetryStore: telemetryStore,
	}
}

func (h *DashboardHandler) Stats(c *gin.Context) {
	totalDevices, _ := h.deviceStore.Count()
	onlineDevices, _ := h.deviceStore.CountByStatus("online")
	totalMessages, _ := h.messageStore.Count()
	activeRules, _ := h.ruleStore.CountEnabled()
	totalUsers, _ := h.userStore.Count()

	today := time.Now().Format("2006-01-02") + " 00:00:00"
	messagesToday, _ := h.messageStore.CountSince(today)

	c.JSON(http.StatusOK, gin.H{
		"total_devices":   totalDevices,
		"online_devices":  onlineDevices,
		"offline_devices": totalDevices - onlineDevices,
		"total_messages":  totalMessages,
		"messages_today":  messagesToday,
		"active_rules":    activeRules,
		"total_users":     totalUsers,
	})
}

func (h *DashboardHandler) Throughput(c *gin.Context) {
	hours := getIntQuery(c, "hours", 24)
	since := time.Now().Add(-time.Duration(hours) * time.Hour).Format("2006-01-02 15:04:05")

	data, err := h.messageStore.GetThroughput(5, since)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": data})
}
