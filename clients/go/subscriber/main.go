package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func main() {
	broker := getEnv("MQTT_BROKER", "tcp://localhost:1883")
	topic := getEnv("SUBSCRIBE_TOPIC", "#")

	opts := mqtt.NewClientOptions().
		AddBroker(broker).
		SetClientID("subscriber-cli").
		SetAutoReconnect(true).
		SetOnConnectHandler(func(c mqtt.Client) {
			log.Printf("Connected to %s", broker)
			token := c.Subscribe(topic, 1, func(c mqtt.Client, m mqtt.Message) {
				fmt.Printf("[%s] %s: %s\n", m.Topic(), time.Now().Format("15:04:05"), string(m.Payload()))
			})
			token.Wait()
			if token.Error() != nil {
				log.Printf("Subscribe error: %v", token.Error())
			} else {
				log.Printf("Subscribed to: %s", topic)
			}
		}).
		SetConnectionLostHandler(func(c mqtt.Client, err error) {
			log.Printf("Connection lost: %v", err)
		})

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("Failed to connect: %v", token.Error())
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down...")
	client.Disconnect(1000)
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
