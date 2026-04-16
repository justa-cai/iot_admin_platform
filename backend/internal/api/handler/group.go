package handler

import (
	"net/http"

	"iot-admin/internal/model"
	"iot-admin/internal/store/sqlite"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type GroupHandler struct {
	groupStore *sqlite.GroupStore
}

func NewGroupHandler(groupStore *sqlite.GroupStore) *GroupHandler {
	return &GroupHandler{groupStore: groupStore}
}

func (h *GroupHandler) List(c *gin.Context) {
	groups, err := h.groupStore.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": groups})
}

func (h *GroupHandler) Get(c *gin.Context) {
	id := c.Param("id")
	group, err := h.groupStore.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if group == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Group not found"})
		return
	}

	// Include devices
	devices, err := h.groupStore.GetDevices(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"group":   group,
		"devices": devices,
	})
}

type CreateGroupRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description string  `json:"description"`
	ParentID    *string `json:"parent_id"`
}

func (h *GroupHandler) Create(c *gin.Context) {
	var req CreateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	group := &model.Group{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Description: req.Description,
		ParentID:    req.ParentID,
	}

	if err := h.groupStore.Create(group); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Group name already exists"})
		return
	}

	c.JSON(http.StatusCreated, group)
}

type UpdateGroupRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	ParentID    *string `json:"parent_id"`
}

func (h *GroupHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req UpdateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	group, err := h.groupStore.GetByID(id)
	if err != nil || group == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Group not found"})
		return
	}

	if req.Name != "" {
		group.Name = req.Name
	}
	group.Description = req.Description
	group.ParentID = req.ParentID

	if err := h.groupStore.Update(group); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Group updated"})
}

func (h *GroupHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.groupStore.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Group deleted"})
}

type GroupDevicesRequest struct {
	DeviceIDs []string `json:"device_ids" binding:"required"`
}

func (h *GroupHandler) AddDevices(c *gin.Context) {
	id := c.Param("id")
	var req GroupDevicesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.groupStore.AddDevices(id, req.DeviceIDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Devices added to group"})
}

func (h *GroupHandler) RemoveDevices(c *gin.Context) {
	id := c.Param("id")
	var req GroupDevicesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.groupStore.RemoveDevices(id, req.DeviceIDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Devices removed from group"})
}
