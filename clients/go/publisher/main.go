package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func main() {
	broker := getEnv("MQTT_BROKER", "tcp://localhost:1883")
	deviceKey := getEnv("DEVICE_KEY", "test-device-key")
	deviceSecret := getEnv("DEVICE_SECRET", "test-device-secret")
	deviceName := getEnv("DEVICE_NAME", "sensor-001")
	interval := getEnvInt("PUBLISH_INTERVAL", 5)

	opts := mqtt.NewClientOptions().
		AddBroker(broker).
		SetClientID(deviceName).
		SetUsername(deviceKey).
		SetPassword(deviceSecret).
		SetAutoReconnect(true).
		SetConnectRetry(true).
		SetConnectRetryInterval(5 * time.Second).
		SetOnConnectHandler(func(c mqtt.Client) {
			log.Printf("Connected to %s as %s", broker, deviceName)
		}).
		SetConnectionLostHandler(func(c mqtt.Client, err error) {
			log.Printf("Connection lost: %v", err)
		})

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("Failed to connect: %v", token.Error())
	}

	topic := fmt.Sprintf("telemetry/%s/data", deviceName)

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	log.Printf("Publishing telemetry to %s every %d seconds", topic, interval)

	for {
		select {
		case <-ticker.C:
			payload := map[string]interface{}{
				"temperature":     20 + rand.Float64()*30,
				"humidity":        30 + rand.Float64()*60,
				"pressure":        980 + rand.Float64()*60,
				"battery_level":   50 + rand.Float64()*50,
				"signal_strength": int(-30 - rand.Intn(70)),
				"timestamp":       time.Now().Unix(),
			}

			data, _ := json.Marshal(payload)
			token := client.Publish(topic, 1, false, data)
			token.Wait()
			if token.Error() != nil {
				log.Printf("Publish error: %v", token.Error())
			} else {
				log.Printf("Published: %s", string(data))
			}

		case <-sigChan:
			log.Println("Shutting down...")
			client.Disconnect(1000)
			return
		}
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	var n int
	fmt.Sscanf(val, "%d", &n)
	if n <= 0 {
		return defaultVal
	}
	return n
}
