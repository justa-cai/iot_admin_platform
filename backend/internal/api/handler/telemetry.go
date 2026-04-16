package handler

import (
	"net/http"

	"iot-admin/internal/store/sqlite"

	"github.com/gin-gonic/gin"
)

type TelemetryHandler struct {
	telemetryStore *sqlite.TelemetryStore
}

func NewTelemetryHandler(telemetryStore *sqlite.TelemetryStore) *TelemetryHandler {
	return &TelemetryHandler{telemetryStore: telemetryStore}
}

func (h *TelemetryHandler) Query(c *gin.Context) {
	deviceID := c.Query("device_id")
	startTime := c.Query("start_time")
	endTime := c.Query("end_time")
	page := getIntQuery(c, "page", 1)
	pageSize := getIntQuery(c, "page_size", 100)

	records, total, err := h.telemetryStore.Query(deviceID, startTime, endTime, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      records,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (h *TelemetryHandler) Latest(c *gin.Context) {
	deviceID := c.Query("device_id")
	limit := getIntQuery(c, "limit", 10)

	records, err := h.telemetryStore.Latest(deviceID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": records})
}
