package handler

import (
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"iot-admin/internal/model"
	"iot-admin/internal/store/sqlite"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type FirmwareHandler struct {
	firmwareStore *sqlite.FirmwareStore
	otaStore      *sqlite.OTAStore
	uploadDir     string
}

func NewFirmwareHandler(firmwareStore *sqlite.FirmwareStore, otaStore *sqlite.OTAStore, uploadDir string) *FirmwareHandler {
	os.MkdirAll(uploadDir, 0755)
	return &FirmwareHandler{
		firmwareStore: firmwareStore,
		otaStore:      otaStore,
		uploadDir:     uploadDir,
	}
}

func (h *FirmwareHandler) List(c *gin.Context) {
	var q model.ListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	list, total, err := h.firmwareStore.List(&q)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      list,
		"total":     total,
		"page":      q.Page,
		"page_size": q.PageSize,
	})
}

func (h *FirmwareHandler) Upload(c *gin.Context) {
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File required"})
		return
	}
	defer file.Close()

	name := c.PostForm("name")
	version := c.PostForm("version")
	deviceModel := c.PostForm("device_model")
	description := c.PostForm("description")

	if name == "" || version == "" || deviceModel == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name, version, device_model are required"})
		return
	}

	id := uuid.New().String()
	fileName := fmt.Sprintf("%s_%s_%s", deviceModel, version, id)
	filePath := filepath.Join(h.uploadDir, fileName)

	dst, err := os.Create(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}
	defer dst.Close()

	hasher := sha256.New()
	size, err := io.Copy(dst, io.TeeReader(file, hasher))
	if err != nil {
		os.Remove(filePath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	checksum := fmt.Sprintf("%x", hasher.Sum(nil))

	firmware := &model.Firmware{
		ID:          id,
		Name:        name,
		Version:     version,
		DeviceModel: deviceModel,
		FilePath:    filePath,
		FileSize:    size,
		Checksum:    checksum,
		Description: description,
	}

	if err := h.firmwareStore.Create(firmware); err != nil {
		os.Remove(filePath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, firmware)
}

func (h *FirmwareHandler) Download(c *gin.Context) {
	id := c.Param("id")
	firmware, err := h.firmwareStore.GetByID(id)
	if err != nil || firmware == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Firmware not found"})
		return
	}

	c.FileAttachment(firmware.FilePath, filepath.Base(firmware.FilePath))
}

func (h *FirmwareHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	firmware, err := h.firmwareStore.GetByID(id)
	if err != nil || firmware == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Firmware not found"})
		return
	}

	// Delete file
	os.Remove(firmware.FilePath)

	if err := h.firmwareStore.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Firmware deleted"})
}

// OTA Upgrade operations

type CreateOTARequest struct {
	DeviceID   string `json:"device_id" binding:"required"`
	FirmwareID string `json:"firmware_id" binding:"required"`
}

func (h *FirmwareHandler) CreateOTA(c *gin.Context) {
	var req CreateOTARequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check firmware exists
	firmware, err := h.firmwareStore.GetByID(req.FirmwareID)
	if err != nil || firmware == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Firmware not found"})
		return
	}

	ota := &model.OTAUpgrade{
		ID:         uuid.New().String(),
		DeviceID:   req.DeviceID,
		FirmwareID: req.FirmwareID,
		Status:     "pending",
	}

	if err := h.otaStore.Create(ota); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ota.DeviceName = ""
	ota.FirmwareName = firmware.Name
	ota.Version = firmware.Version

	c.JSON(http.StatusCreated, ota)
}

type BatchOTARequest struct {
	DeviceIDs  []string `json:"device_ids" binding:"required"`
	FirmwareID string   `json:"firmware_id" binding:"required"`
}

func (h *FirmwareHandler) BatchCreateOTA(c *gin.Context) {
	var req BatchOTARequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	firmware, err := h.firmwareStore.GetByID(req.FirmwareID)
	if err != nil || firmware == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Firmware not found"})
		return
	}

	otas := []model.OTAUpgrade{}
	for _, deviceID := range req.DeviceIDs {
		ota := &model.OTAUpgrade{
			ID:         uuid.New().String(),
			DeviceID:   deviceID,
			FirmwareID: req.FirmwareID,
			Status:     "pending",
		}
		if err := h.otaStore.Create(ota); err != nil {
			continue
		}
		ota.FirmwareName = firmware.Name
		ota.Version = firmware.Version
		otas = append(otas, *ota)
	}

	c.JSON(http.StatusCreated, gin.H{"data": otas, "count": len(otas)})
}

func (h *FirmwareHandler) ListOTA(c *gin.Context) {
	var q model.ListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	q.DeviceID = c.Query("device_id")

	list, total, err := h.otaStore.List(&q)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      list,
		"total":     total,
		"page":      q.Page,
		"page_size": q.PageSize,
	})
}

type OTAStatusRequest struct {
	Status   string `json:"status" binding:"required"`
	Progress int    `json:"progress"`
	ErrorMsg string `json:"error_msg"`
}

func (h *FirmwareHandler) UpdateOTAStatus(c *gin.Context) {
	id := c.Param("id")
	var req OTAStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.otaStore.UpdateStatus(id, req.Status, req.Progress, req.ErrorMsg); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "OTA status updated", "updated_at": time.Now()})
}

func (h *FirmwareHandler) DeleteOTA(c *gin.Context) {
	id := c.Param("id")
	if err := h.otaStore.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "OTA task deleted"})
}
