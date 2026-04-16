#!/usr/bin/env python3
"""MQTT Subscriber - Subscribes to topics and prints messages"""

import os
import signal
import sys
import time
from datetime import datetime

import paho.mqtt.client as mqtt


def main():
    broker = os.environ.get("MQTT_BROKER", "localhost")
    port = int(os.environ.get("MQTT_PORT", "1883"))
    topic = os.environ.get("SUBSCRIBE_TOPIC", "#")

    client = mqtt.Client(client_id="py-subscriber")

    def on_connect(c, userdata, flags, rc):
        if rc == 0:
            print(f"Connected to {broker}:{port}")
            c.subscribe(topic, qos=1)
            print(f"Subscribed to: {topic}")
        else:
            print(f"Connection failed with code {rc}")

    def on_message(c, userdata, msg):
        timestamp = datetime.now().strftime("%H:%M:%S")
        print(f"[{msg.topic}] {timestamp}: {msg.payload.decode()}")

    client.on_connect = on_connect
    client.on_message = on_message

    try:
        client.connect(broker, port, 60)
    except Exception as e:
        print(f"Failed to connect: {e}")
        sys.exit(1)

    client.loop_start()

    running = True

    def signal_handler(sig, frame):
        nonlocal running
        running = False

    signal.signal(signal.SIGINT, signal_handler)

    while running:
        time.sleep(1)

    client.loop_stop()
    client.disconnect()
    print("Shutting down...")


if __name__ == "__main__":
    main()
