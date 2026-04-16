package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var (
	totalPublished  int64
	totalReceived   int64
	totalErrors     int64
	startTime       time.Time
)

func main() {
	broker := getEnv("MQTT_BROKER", "tcp://localhost:1883")
	numDevices := getEnvInt("NUM_DEVICES", 100)
	msgInterval := getEnvInt("MSG_INTERVAL_MS", 1000)
	durationSecs := getEnvInt("DURATION_SECS", 60)
	apiURL := getEnv("API_URL", "http://localhost:8080")
	adminToken := getEnv("ADMIN_TOKEN", "")

	log.Printf("=== IoT Admin Load Test ===")
	log.Printf("Broker: %s", broker)
	log.Printf("Devices: %d", numDevices)
	log.Printf("Message interval: %dms", msgInterval)
	log.Printf("Duration: %ds", durationSecs)
	log.Printf("API URL: %s", apiURL)

	startTime = time.Now()

	// First, create devices via API if token provided
	deviceCreds := make([]DeviceCred, numDevices)
	if adminToken != "" {
		log.Printf("Creating %d devices via API...", numDevices)
		for i := 0; i < numDevices; i++ {
			cred, err := createDevice(apiURL, adminToken, fmt.Sprintf("load-device-%03d", i))
			if err != nil {
				log.Printf("Failed to create device %d: %v", i, err)
				continue
			}
			deviceCreds[i] = cred
		}
		log.Printf("Created %d devices", numDevices)
	} else {
		// Use pre-generated keys
		for i := 0; i < numDevices; i++ {
			deviceCreds[i] = DeviceCred{
				Key:    fmt.Sprintf("load-key-%03d", i),
				Secret: fmt.Sprintf("load-secret-%03d", i),
			}
		}
	}

	// Connect all devices
	clients := make([]mqtt.Client, numDevices)
	var connectedCount int64

	log.Printf("Connecting %d devices...", numDevices)

	var wg sync.WaitGroup

	for i := 0; i < numDevices; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			cred := deviceCreds[idx]
			clientID := fmt.Sprintf("loadtest-device-%03d", idx)

			opts := mqtt.NewClientOptions().
				AddBroker(broker).
				SetClientID(clientID).
				SetUsername(cred.Key).
				SetPassword(cred.Secret).
				SetAutoReconnect(false).
				SetConnectTimeout(10 * time.Second).
				SetOnConnectHandler(func(c mqtt.Client) {
					atomic.AddInt64(&connectedCount, 1)
				}).
				SetConnectionLostHandler(func(c mqtt.Client, err error) {
					atomic.AddInt64(&connectedCount, -1)
					atomic.AddInt64(&totalErrors, 1)
				})

			client := mqtt.NewClient(opts)
			token := client.Connect()
			if token.Wait() && token.Error() != nil {
				atomic.AddInt64(&totalErrors, 1)
				return
			}

			clients[idx] = client
		}(i)

		// Stagger connections to avoid overwhelming broker
		if i%10 == 0 && i > 0 {
			time.Sleep(100 * time.Millisecond)
		}
	}

	wg.Wait()
	log.Printf("Connected: %d / %d devices", connectedCount, numDevices)

	// Start stats reporter
	stopStats := make(chan struct{})
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				elapsed := time.Since(startTime).Seconds()
				pubRate := float64(atomic.LoadInt64(&totalPublished)) / elapsed
				log.Printf("[Stats] Connected: %d | Published: %d (%.0f msg/s) | Errors: %d | Elapsed: %.1fs",
					atomic.LoadInt64(&connectedCount),
					atomic.LoadInt64(&totalPublished),
					pubRate,
					atomic.LoadInt64(&totalErrors),
					elapsed,
				)
			case <-stopStats:
				return
			}
		}
	}()

	// Start publishing
	stopPub := make(chan struct{})
	for i := 0; i < numDevices; i++ {
		if clients[i] == nil {
			continue
		}
		go func(idx int) {
			ticker := time.NewTicker(time.Duration(msgInterval) * time.Millisecond)
			defer ticker.Stop()

			topic := fmt.Sprintf("telemetry/loadtest-%03d/data", idx)
			for {
				select {
				case <-ticker.C:
					payload := map[string]interface{}{
						"temperature": 20 + rand.Float64()*30,
						"humidity":    30 + rand.Float64()*60,
						"timestamp":   time.Now().UnixMilli(),
					}
					data, _ := json.Marshal(payload)
					token := clients[idx].Publish(topic, 0, false, data)
					if token.Wait() && token.Error() != nil {
						atomic.AddInt64(&totalErrors, 1)
					} else {
						atomic.AddInt64(&totalPublished, 1)
					}
				case <-stopPub:
					return
				}
			}
		}(i)
	}

	// Wait for duration or signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	timeout := time.After(time.Duration(durationSecs) * time.Second)

	select {
	case <-timeout:
		log.Println("Duration reached")
	case <-sigChan:
		log.Println("Signal received")
	}

	close(stopPub)
	close(stopStats)

	// Final stats
	elapsed := time.Since(startTime).Seconds()
	log.Printf("\n=== Final Results ===")
	log.Printf("Duration:        %.1fs", elapsed)
	log.Printf("Total Published: %d", atomic.LoadInt64(&totalPublished))
	log.Printf("Total Errors:    %d", atomic.LoadInt64(&totalErrors))
	log.Printf("Avg Publish Rate: %.0f msg/s", float64(atomic.LoadInt64(&totalPublished))/elapsed)
	log.Printf("Connected:       %d / %d", atomic.LoadInt64(&connectedCount), numDevices)

	// Disconnect all
	for i := 0; i < numDevices; i++ {
		if clients[i] != nil {
			clients[i].Disconnect(100)
		}
	}
}

type DeviceCred struct {
	Key    string `json:"device_key"`
	Secret string `json:"device_secret"`
}

func createDevice(apiURL, token, name string) (DeviceCred, error) {
	// This is a simplified version - in real test, use net/http to call the API
	return DeviceCred{
		Key:    name + "-key",
		Secret: name + "-secret",
	}, nil
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
