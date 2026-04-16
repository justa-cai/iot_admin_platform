package handler

import (
	"net/http"

	"iot-admin/internal/model"
	"iot-admin/internal/store/sqlite"

	"github.com/gin-gonic/gin"
)

type MessageHandler struct {
	messageStore *sqlite.MessageStore
}

func NewMessageHandler(messageStore *sqlite.MessageStore) *MessageHandler {
	return &MessageHandler{messageStore: messageStore}
}

type PublishRequest struct {
	Topic   string `json:"topic" binding:"required"`
	Payload string `json:"payload"`
	QoS    int    `json:"qos"`
}

func (h *MessageHandler) Publish(c *gin.Context) {
	var req PublishRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	msg := &model.Message{
		Topic:     req.Topic,
		Payload:   req.Payload,
		QoS:      req.QoS,
		Direction: "outbound",
	}

	if err := h.messageStore.Insert(msg); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Published", "topic": req.Topic})
}

func (h *MessageHandler) History(c *gin.Context) {
	var q model.ListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	q.DeviceID = c.Query("device_id")
	startTime := c.Query("start_time")
	endTime := c.Query("end_time")

	messages, total, err := h.messageStore.List(&q, startTime, endTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      messages,
		"total":     total,
		"page":      q.Page,
		"page_size": q.PageSize,
	})
}

func (h *MessageHandler) Topics(c *gin.Context) {
	topics, err := h.messageStore.GetTopics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": topics})
}
