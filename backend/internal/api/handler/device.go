package handler

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"iot-admin/internal/model"
	"iot-admin/internal/store/sqlite"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type DeviceHandler struct {
	deviceStore *sqlite.DeviceStore
}

func NewDeviceHandler(deviceStore *sqlite.DeviceStore) *DeviceHandler {
	return &DeviceHandler{deviceStore: deviceStore}
}

func (h *DeviceHandler) List(c *gin.Context) {
	var q model.ListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	devices, total, err := h.deviceStore.ListWithMetadata(&q)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      devices,
		"total":     total,
		"page":      q.Page,
		"page_size": q.PageSize,
	})
}

func (h *DeviceHandler) Get(c *gin.Context) {
	id := c.Param("id")
	device, err := h.deviceStore.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if device == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Device not found"})
		return
	}

	c.JSON(http.StatusOK, device)
}

type CreateDeviceRequest struct {
	Name     string `json:"name" binding:"required"`
	GroupIDs []string `json:"group_ids"`
	TagIDs   []string `json:"tag_ids"`
	Metadata string `json:"metadata"`
}

func (h *DeviceHandler) Create(c *gin.Context) {
	var req CreateDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	key := generateKey(16)
	secret := generateKey(32)
	metadata := req.Metadata
	if metadata == "" {
		metadata = "{}"
	}

	device := &model.Device{
		ID:           uuid.New().String(),
		Name:         req.Name,
		DeviceKey:    key,
		DeviceSecret: secret,
		Status:       "offline",
		Metadata:     metadata,
	}

	if err := h.deviceStore.Create(device); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Set groups and tags
	if len(req.GroupIDs) > 0 {
		h.deviceStore.SetGroups(device.ID, req.GroupIDs)
	}
	if len(req.TagIDs) > 0 {
		h.deviceStore.SetTags(device.ID, req.TagIDs)
	}

	// Return device with secret (only time it's visible)
	c.JSON(http.StatusCreated, gin.H{
		"id":            device.ID,
		"name":          device.Name,
		"device_key":    device.DeviceKey,
		"device_secret": device.DeviceSecret,
		"status":        device.Status,
		"metadata":      device.Metadata,
	})
}

type UpdateDeviceRequest struct {
	Name     string   `json:"name"`
	Metadata string   `json:"metadata"`
	GroupIDs []string `json:"group_ids"`
	TagIDs   []string `json:"tag_ids"`
}

func (h *DeviceHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req UpdateDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	device, err := h.deviceStore.GetByID(id)
	if err != nil || device == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Device not found"})
		return
	}

	if req.Name != "" {
		device.Name = req.Name
	}
	if req.Metadata != "" {
		device.Metadata = req.Metadata
	}

	if err := h.deviceStore.Update(device); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if req.GroupIDs != nil {
		h.deviceStore.SetGroups(id, req.GroupIDs)
	}
	if req.TagIDs != nil {
		h.deviceStore.SetTags(id, req.TagIDs)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Device updated"})
}

func (h *DeviceHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.deviceStore.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Device deleted"})
}

func (h *DeviceHandler) GetStatus(c *gin.Context) {
	id := c.Param("id")
	device, err := h.deviceStore.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if device == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Device not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":           device.ID,
		"status":       device.Status,
		"last_seen_at": device.LastSeenAt,
	})
}

func generateKey(length int) string {
	b := make([]byte, length)
	rand.Read(b)
	return hex.EncodeToString(b)
}
