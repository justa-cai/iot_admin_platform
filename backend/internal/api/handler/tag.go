package handler

import (
	"net/http"

	"iot-admin/internal/model"
	"iot-admin/internal/store/sqlite"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type TagHandler struct {
	tagStore *sqlite.TagStore
}

func NewTagHandler(tagStore *sqlite.TagStore) *TagHandler {
	return &TagHandler{tagStore: tagStore}
}

func (h *TagHandler) List(c *gin.Context) {
	tags, err := h.tagStore.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": tags})
}

type CreateTagRequest struct {
	Name  string `json:"name" binding:"required"`
	Color string `json:"color"`
}

func (h *TagHandler) Create(c *gin.Context) {
	var req CreateTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	color := req.Color
	if color == "" {
		color = "#409EFF"
	}

	tag := &model.Tag{
		ID:    uuid.New().String(),
		Name:  req.Name,
		Color: color,
	}

	if err := h.tagStore.Create(tag); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Tag name already exists"})
		return
	}

	c.JSON(http.StatusCreated, tag)
}

type UpdateTagRequest struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

func (h *TagHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req UpdateTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tag, err := h.tagStore.GetByID(id)
	if err != nil || tag == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tag not found"})
		return
	}

	if req.Name != "" {
		tag.Name = req.Name
	}
	if req.Color != "" {
		tag.Color = req.Color
	}

	if err := h.tagStore.Update(tag); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tag updated"})
}

func (h *TagHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.tagStore.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Tag deleted"})
}
