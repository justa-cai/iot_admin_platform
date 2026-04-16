#!/usr/bin/env python3
"""IoT Device Simulator - Publishes telemetry data to MQTT broker"""

import json
import os
import random
import signal
import sys
import time

import paho.mqtt.client as mqtt


def main():
    broker = os.environ.get("MQTT_BROKER", "localhost")
    port = int(os.environ.get("MQTT_PORT", "1883"))
    device_key = os.environ.get("DEVICE_KEY", "test-device-key")
    device_secret = os.environ.get("DEVICE_SECRET", "test-device-secret")
    device_name = os.environ.get("DEVICE_NAME", "py-sensor-001")
    interval = int(os.environ.get("PUBLISH_INTERVAL", "5"))

    client = mqtt.Client(client_id=device_name)
    client.username_pw_set(device_key, device_secret)

    def on_connect(c, userdata, flags, rc):
        if rc == 0:
            print(f"Connected to {broker}:{port} as {device_name}")
        else:
            print(f"Connection failed with code {rc}")

    def on_disconnect(c, userdata, rc):
        print(f"Disconnected (rc={rc})")

    client.on_connect = on_connect
    client.on_disconnect = on_disconnect

    try:
        client.connect(broker, port, 60)
    except Exception as e:
        print(f"Failed to connect: {e}")
        sys.exit(1)

    client.loop_start()
    topic = f"telemetry/{device_name}/data"

    print(f"Publishing to {topic} every {interval}s")

    running = True

    def signal_handler(sig, frame):
        nonlocal running
        running = False

    signal.signal(signal.SIGINT, signal_handler)

    while running:
        payload = {
            "temperature": round(20 + random.random() * 30, 2),
            "humidity": round(30 + random.random() * 60, 2),
            "pressure": round(980 + random.random() * 60, 2),
            "battery_level": round(50 + random.random() * 50, 2),
            "signal_strength": -30 - random.randint(0, 70),
            "timestamp": int(time.time()),
        }

        result = client.publish(topic, json.dumps(payload), qos=1)
        if result.rc == mqtt.MQTT_ERR_SUCCESS:
            print(f"Published: {json.dumps(payload)}")
        else:
            print(f"Publish error: {result.rc}")

        time.sleep(interval)

    client.loop_stop()
    client.disconnect()
    print("Shutting down...")


if __name__ == "__main__":
    main()
