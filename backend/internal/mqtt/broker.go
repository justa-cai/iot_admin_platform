package mqtt

import (
	"fmt"
	"log"
	"time"

	"iot-admin/internal/store/sqlite"

	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/hooks/auth"
	"github.com/mochi-mqtt/server/v2/listeners"
)

type Broker struct {
	server     *mqtt.Server
	deviceStore *sqlite.DeviceStore
}

func NewBroker(tcpPort, wsPort int, deviceStore *sqlite.DeviceStore) *Broker {
	return &Broker{
		server:      mqtt.New(nil),
		deviceStore: deviceStore,
	}
}

func (b *Broker) Start(tcpPort, wsPort int, hook *Hook) error {
	// Add auth hook - allow all for device auth handled by our custom hook
	b.server.AddHook(new(auth.AllowHook), nil)

	// Add our custom hook
	if hook != nil {
		b.server.AddHook(hook, nil)
	}

	// TCP listener
	tcp := listeners.NewTCP(listeners.Config{
		ID:      "mqtt-tcp",
		Address: fmt.Sprintf(":%d", tcpPort),
	})
	if err := b.server.AddListener(tcp); err != nil {
		return fmt.Errorf("add TCP listener: %w", err)
	}

	// WebSocket listener
	ws := listeners.NewWebsocket(listeners.Config{
		ID:      "mqtt-ws",
		Address: fmt.Sprintf(":%d", wsPort),
	})
	if err := b.server.AddListener(ws); err != nil {
		return fmt.Errorf("add WS listener: %w", err)
	}

	go func() {
		if err := b.server.Serve(); err != nil {
			log.Printf("MQTT server error: %v", err)
		}
	}()

	log.Printf("MQTT Broker started - TCP:%d WS:%d", tcpPort, wsPort)
	return nil
}

func (b *Broker) Publish(topic string, payload []byte, retain bool, qos byte) error {
	return b.server.Publish(topic, payload, retain, qos)
}

func (b *Broker) Stop() {
	b.server.Close()
	time.Sleep(100 * time.Millisecond)
}
