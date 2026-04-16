package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"iot-admin/internal/api"
	"iot-admin/internal/config"
	"iot-admin/internal/mqtt"
	"iot-admin/internal/rule"
	"iot-admin/internal/store/sqlite"
	"iot-admin/internal/ws"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database
	store, err := sqlite.NewStore(cfg.Database.Path)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer store.Close()
	db := store.DB()

	// Initialize stores
	userStore := sqlite.NewUserStore(db)
	deviceStore := sqlite.NewDeviceStore(db)
	groupStore := sqlite.NewGroupStore(db)
	tagStore := sqlite.NewTagStore(db)
	messageStore := sqlite.NewMessageStore(db)
	telemetryStore := sqlite.NewTelemetryStore(db)
	ruleStore := sqlite.NewRuleStore(db)
	firmwareStore := sqlite.NewFirmwareStore(db)
	otaStore := sqlite.NewOTAStore(db)

	// Ensure admin user exists
	adminHash, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if adminID, err := userStore.EnsureAdmin("admin", string(adminHash)); err == nil {
		log.Printf("Admin user ready (id: %s, password: admin123)", adminID)
	}

	// Initialize WebSocket hub
	hub := ws.NewHub()
	go hub.Run()

	// Initialize MQTT broker (create hook first, then broker)
	// Rule engine needs broker for publish actions - we'll use a wrapper
	brokerWrapper := &mqttBrokerWrapper{}

	ruleEngine := rule.NewEngine(ruleStore, hub, brokerWrapper)

	hook := mqtt.NewHook(deviceStore, messageStore, telemetryStore, ruleEngine, hub)
	broker := mqtt.NewBroker(cfg.MQTT.TCPPort, cfg.MQTT.WSPort, deviceStore)

	// Wire broker to rule engine
	brokerWrapper.broker = broker

	if err := broker.Start(cfg.MQTT.TCPPort, cfg.MQTT.WSPort, hook); err != nil {
		log.Fatalf("Failed to start MQTT broker: %v", err)
	}
	defer broker.Stop()

	// Initialize API router
	uploadDir := "./data/firmware"
	router := api.NewRouter(
		struct {
			JWTSecret string
			JWTExpire time.Duration
		}{
			JWTSecret: cfg.JWT.Secret,
			JWTExpire: cfg.JWT.ExpireTime,
		},
		struct {
			UserStore      *sqlite.UserStore
			DeviceStore    *sqlite.DeviceStore
			GroupStore     *sqlite.GroupStore
			TagStore       *sqlite.TagStore
			MessageStore   *sqlite.MessageStore
			TelemetryStore *sqlite.TelemetryStore
			RuleStore      *sqlite.RuleStore
			FirmwareStore  *sqlite.FirmwareStore
			OTAStore       *sqlite.OTAStore
		}{
			UserStore:      userStore,
			DeviceStore:    deviceStore,
			GroupStore:     groupStore,
			TagStore:       tagStore,
			MessageStore:   messageStore,
			TelemetryStore: telemetryStore,
			RuleStore:      ruleStore,
			FirmwareStore:  firmwareStore,
			OTAStore:       otaStore,
		},
		struct {
			Hub        *ws.Hub
			RuleEngine *rule.Engine
		}{
			Hub:        hub,
			RuleEngine: ruleEngine,
		},
		uploadDir,
	)

	// Start HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      router.Engine(),
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	go func() {
		log.Printf("HTTP Server starting on :%d", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	log.Printf("IoT Admin Platform started successfully!")
	log.Printf("  API: http://localhost:%d/api/v1", cfg.Server.Port)
	log.Printf("  MQTT TCP: :%d", cfg.MQTT.TCPPort)
	log.Printf("  MQTT WS: :%d", cfg.MQTT.WSPort)
	log.Printf("  Login: admin / admin123")

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced shutdown: %v", err)
	}

	log.Println("Server exited")
}

// Wrapper to break circular dependency between broker and rule engine
type mqttBrokerWrapper struct {
	broker *mqtt.Broker
}

func (w *mqttBrokerWrapper) Publish(topic string, payload []byte, retain bool, qos byte) error {
	if w.broker == nil {
		return fmt.Errorf("broker not initialized")
	}
	return w.broker.Publish(topic, payload, retain, qos)
}
