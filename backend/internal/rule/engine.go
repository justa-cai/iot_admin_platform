package rule

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"iot-admin/internal/model"
	"iot-admin/internal/store/sqlite"
	"iot-admin/internal/ws"
)

type Engine struct {
	ruleStore    *sqlite.RuleStore
	hub          *ws.Hub
	broker       BrokerPublisher
	rules        []model.Rule
}

type BrokerPublisher interface {
	Publish(topic string, payload []byte, retain bool, qos byte) error
}

func NewEngine(ruleStore *sqlite.RuleStore, hub *ws.Hub, broker BrokerPublisher) *Engine {
	e := &Engine{
		ruleStore: ruleStore,
		hub:       hub,
		broker:    broker,
	}
	e.loadRules()
	return e
}

func (e *Engine) loadRules() {
	rules, err := e.ruleStore.ListEnabled()
	if err != nil {
		log.Printf("[Rule] Failed to load rules: %v", err)
		return
	}
	e.rules = rules
	log.Printf("[Rule] Loaded %d enabled rules", len(rules))
}

func (e *Engine) ReloadRules() {
	e.loadRules()
}

func (e *Engine) Evaluate(topic string, payload string, deviceID *string) {
	for i := range e.rules {
		rule := &e.rules[i]
		if !topicMatchesPattern(rule.TopicPattern, topic) {
			continue
		}

		// Parse condition
		var cond map[string]interface{}
		if err := json.Unmarshal([]byte(rule.Condition), &cond); err != nil {
			continue
		}

		// Parse payload
		var payloadMap map[string]interface{}
		if err := json.Unmarshal([]byte(payload), &payloadMap); err != nil {
			continue
		}

		if evaluateCondition(cond, payloadMap) {
			// Check cooldown
			if rule.LastTriggered != nil {
				elapsed := time.Since(*rule.LastTriggered)
				if elapsed < time.Duration(rule.CooldownSecs)*time.Second {
					continue
				}
			}

			// Update last triggered
			e.ruleStore.UpdateLastTriggered(rule.ID)

			// Execute action
			e.executeAction(rule, topic, payload, deviceID)
		}
	}
}

func topicMatchesPattern(pattern, topic string) bool {
	patternParts := strings.Split(pattern, "/")
	topicParts := strings.Split(topic, "/")

	if len(patternParts) != len(topicParts) {
		// Check for # wildcard at end
		if len(patternParts) > 0 && patternParts[len(patternParts)-1] == "#" {
			return strings.HasPrefix(topic, strings.Join(patternParts[:len(patternParts)-1], "/"))
		}
		return false
	}

	for i := range patternParts {
		if patternParts[i] == "+" || patternParts[i] == "#" {
			continue
		}
		if patternParts[i] != topicParts[i] {
			return false
		}
	}
	return true
}

func evaluateCondition(cond map[string]interface{}, payload map[string]interface{}) bool {
	field, _ := cond["field"].(string)
	operator, _ := cond["operator"].(string)
	value := cond["value"]

	if field == "" || operator == "" {
		return false
	}

	payloadVal, exists := payload[field]
	if !exists {
		return false
	}

	return compareValues(payloadVal, value, operator)
}

func compareValues(actual, expected interface{}, operator string) bool {
	actualFloat, ok1 := toFloat64(actual)
	expectedFloat, ok2 := toFloat64(expected)

	switch operator {
	case "gt":
		if ok1 && ok2 {
			return actualFloat > expectedFloat
		}
	case "gte":
		if ok1 && ok2 {
			return actualFloat >= expectedFloat
		}
	case "lt":
		if ok1 && ok2 {
			return actualFloat < expectedFloat
		}
	case "lte":
		if ok1 && ok2 {
			return actualFloat <= expectedFloat
		}
	case "eq":
		return actual == expected
	case "neq":
		return actual != expected
	case "contains":
		actualStr, _ := actual.(string)
		expectedStr, _ := expected.(string)
		return strings.Contains(actualStr, expectedStr)
	}
	return false
}

func toFloat64(v interface{}) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case float32:
		return float64(val), true
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	case json.Number:
		f, err := val.Float64()
		return f, err == nil
	}
	return 0, false
}

func (e *Engine) executeAction(r *model.Rule, topic string, payload string, deviceID *string) {
	var config map[string]interface{}
	if err := json.Unmarshal([]byte(r.ActionConfig), &config); err != nil {
		log.Printf("[Rule] Invalid action config for rule %s: %v", r.ID, err)
		return
	}

	var result string
	switch r.ActionType {
	case "alert":
		result = e.executeAlert(r, topic, payload, deviceID)
	case "publish":
		result = e.executePublish(config)
	case "forward":
		result = e.executeForward(config, payload)
	default:
		log.Printf("[Rule] Unknown action type: %s", r.ActionType)
		return
	}

	// Log the execution
	logEntry := &model.RuleLog{
		RuleID:       r.ID,
		DeviceID:     deviceID,
		Topic:        topic,
		Payload:      payload,
		ActionResult: result,
	}
	go e.ruleStore.InsertLog(logEntry)

	log.Printf("[Rule] Rule '%s' triggered for topic '%s': %s", r.Name, topic, result)
}

func (e *Engine) executeAlert(r *model.Rule, topic string, payload string, deviceID *string) string {
	// Push alert via WebSocket
	if e.hub != nil {
		e.hub.BroadcastEvent("rule.triggered", map[string]interface{}{
			"rule_id":   r.ID,
			"rule_name": r.Name,
			"topic":     topic,
			"payload":   payload,
			"device_id": deviceID,
			"timestamp": time.Now().Format(time.RFC3339),
		})
	}
	return "alert_sent"
}

func (e *Engine) executePublish(config map[string]interface{}) string {
	topic, _ := config["topic"].(string)
	message, _ := config["message"].(string)
	if topic == "" {
		return "error: no topic"
	}

	if e.broker != nil {
		if err := e.broker.Publish(topic, []byte(message), false, 0); err != nil {
			return "error: " + err.Error()
		}
	}
	return "published to " + topic
}

func (e *Engine) executeForward(config map[string]interface{}, payload string) string {
	url, _ := config["url"].(string)
	if url == "" {
		return "error: no url"
	}

	resp, err := http.Post(url, "application/json", strings.NewReader(payload))
	if err != nil {
		return "error: " + err.Error()
	}
	defer resp.Body.Close()
	return "forwarded to " + url + " (status: " + resp.Status + ")"
}
