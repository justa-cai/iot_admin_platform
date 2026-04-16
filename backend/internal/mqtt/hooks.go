package mqtt

import (
	"encoding/json"
	"log"
	"strings"
	"sync"
	"time"

	"iot-admin/internal/model"
	"iot-admin/internal/rule"
	"iot-admin/internal/store/sqlite"
	"iot-admin/internal/ws"

	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/packets"
)

// deviceMap stores clientID -> deviceID mapping
var deviceMap sync.Map

type Hook struct {
	mqtt.HookBase
	deviceStore    *sqlite.DeviceStore
	messageStore   *sqlite.MessageStore
	telemetryStore *sqlite.TelemetryStore
	ruleEngine     *rule.Engine
	hub            *ws.Hub
}

func NewHook(
	deviceStore *sqlite.DeviceStore,
	messageStore *sqlite.MessageStore,
	telemetryStore *sqlite.TelemetryStore,
	ruleEngine *rule.Engine,
	hub *ws.Hub,
) *Hook {
	return &Hook{
		deviceStore:    deviceStore,
		messageStore:   messageStore,
		telemetryStore: telemetryStore,
		ruleEngine:     ruleEngine,
		hub:            hub,
	}
}

func (h *Hook) ID() string {
	return "iot-admin-hook"
}

func (h *Hook) Provides(b byte) bool {
	return b == mqtt.OnConnect || b == mqtt.OnDisconnect || b == mqtt.OnPublish
}

// OnConnect handles device connection - authenticate and set online
func (h *Hook) OnConnect(cl *mqtt.Client, pk packets.Packet) error {
	username := string(pk.Connect.Username)
	password := string(pk.Connect.Password)

	if username == "" {
		log.Printf("[MQTT] Connect rejected: no credentials from %s", cl.ID)
		return nil // return nil to let default auth handle it
	}

	device, err := h.deviceStore.GetByKey(username)
	if err != nil || device == nil {
		log.Printf("[MQTT] Connect rejected: unknown device key %s", username)
		return nil
	}

	// Simple password check
	if password != device.DeviceSecret {
		log.Printf("[MQTT] Connect rejected: invalid password for %s", username)
		return nil
	}

	// Store device mapping
	deviceMap.Store(cl.ID, map[string]string{
		"device_id":   device.ID,
		"device_name": device.Name,
	})

	// Set device online
	h.deviceStore.UpdateStatus(device.ID, "online")
	log.Printf("[MQTT] Device connected: %s (%s)", device.Name, device.ID)

	// Notify WebSocket clients
	if h.hub != nil {
		h.hub.BroadcastEvent("device.status_changed", map[string]interface{}{
			"device_id":   device.ID,
			"device_name": device.Name,
			"status":      "online",
			"timestamp":   time.Now().Format(time.RFC3339),
		})
	}

	return nil
}

// OnDisconnect handles device disconnection
func (h *Hook) OnDisconnect(cl *mqtt.Client, err error, expire bool) {
	val, ok := deviceMap.LoadAndDelete(cl.ID)
	if !ok {
		return
	}

	info := val.(map[string]string)
	deviceID := info["device_id"]
	deviceName := info["device_name"]

	// Set device offline
	h.deviceStore.UpdateStatus(deviceID, "offline")
	log.Printf("[MQTT] Device disconnected: %s (%s)", deviceName, deviceID)

	// Notify WebSocket clients
	if h.hub != nil {
		h.hub.BroadcastEvent("device.status_changed", map[string]interface{}{
			"device_id":   deviceID,
			"device_name": deviceName,
			"status":      "offline",
			"timestamp":   time.Now().Format(time.RFC3339),
		})
	}
}

// OnPublish handles incoming messages
func (h *Hook) OnPublish(cl *mqtt.Client, pk packets.Packet) (packets.Packet, error) {
	topic := pk.TopicName
	payload := string(pk.Payload)

	// Try to get device ID from mapping
	var deviceID *string
	if val, ok := deviceMap.Load(cl.ID); ok {
		info := val.(map[string]string)
		id := info["device_id"]
		deviceID = &id
	}

	// Store message in history
	msg := &model.Message{
		DeviceID:  deviceID,
		Topic:     topic,
		Payload:   payload,
		QoS:      int(pk.FixedHeader.Qos),
		Direction: "inbound",
	}
	go h.messageStore.Insert(msg)

	// If topic starts with "telemetry/", store as telemetry data
	if strings.HasPrefix(topic, "telemetry/") {
		var fields map[string]interface{}
		if err := json.Unmarshal(pk.Payload, &fields); err == nil {
			if deviceID != nil {
				telemetry := &model.Telemetry{
					DeviceID: *deviceID,
					Topic:    topic,
					Fields:   string(pk.Payload),
				}
				go h.telemetryStore.Insert(telemetry)
			}
		}
	}

	// Notify WebSocket clients
	if h.hub != nil {
		h.hub.BroadcastEvent("message.received", map[string]interface{}{
			"device_id": deviceID,
			"topic":     topic,
			"payload":   payload,
			"qos":       int(pk.FixedHeader.Qos),
			"timestamp": time.Now().Format(time.RFC3339),
		})
	}

	// Send to rule engine
	if h.ruleEngine != nil {
		h.ruleEngine.Evaluate(topic, payload, deviceID)
	}

	return pk, nil
}
