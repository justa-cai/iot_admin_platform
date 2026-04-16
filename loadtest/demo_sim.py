#!/usr/bin/env python3
"""Simulate live MQTT devices for demo purposes."""

import json
import time
import random
import threading
import sys

try:
    import paho.mqtt.client as mqtt
except ImportError:
    print("Installing paho-mqtt...")
    import subprocess
    subprocess.check_call([sys.executable, "-m", "pip", "install", "paho-mqtt", "-q"])
    import paho.mqtt.client as mqtt

# Devices to simulate (name, key, secret)
DEVICES = [
    ("温度传感器-A01", "temp-a01", "sec-a01"),
    ("温度传感器-A02", "temp-a02", "sec-a02"),
    ("温度传感器-B01", "temp-b01", "sec-b01"),
    ("湿度传感器-A01", "hum-a01", "sec-a01"),
    ("智能电表-01", "elec-01", "sec-elec01"),
    ("边缘网关-01", "gw-01", "sec-gw01"),
    ("光照传感器-01", "light-01", "sec-light01"),
    ("压力传感器-01", "press-01", "sec-press01"),
]

BROKER = "tcp://localhost:1883"

class SimDevice:
    def __init__(self, name, key, secret):
        self.name = name
        self.key = key
        self.secret = secret
        self.client = None
        self.connected = False

        # Base values for realistic simulation
        if "温度" in name:
            self.base_temp = random.uniform(20, 30)
            self.fields = lambda: {
                "temperature": round(self.base_temp + random.gauss(0, 3), 1),
                "humidity": round(random.uniform(40, 70), 1),
            }
        elif "湿度" in name:
            self.base_hum = random.uniform(45, 65)
            self.fields = lambda: {
                "humidity": round(self.base_hum + random.gauss(0, 5), 1),
                "temperature": round(random.uniform(22, 28), 1),
            }
        elif "电表" in name:
            self.base_power = random.uniform(2, 8)
            self.fields = lambda: {
                "voltage": round(220 + random.gauss(0, 2), 1),
                "current": round(self.base_power + random.gauss(0, 1), 2),
                "power": round(self.base_power * 220 / 1000 + random.gauss(0, 0.1), 2),
                "energy_total": round(random.uniform(100, 500), 1),
            }
        elif "网关" in name:
            self.fields = lambda: {
                "cpu_usage": round(random.uniform(10, 60), 1),
                "mem_usage": round(random.uniform(30, 70), 1),
                "uptime": random.randint(3600, 864000),
                "connected_devices": random.randint(5, 30),
            }
        elif "光照" in name:
            self.fields = lambda: {
                "light": round(random.uniform(100, 1000), 1),
                "uv_index": round(random.uniform(0, 8), 1),
            }
        elif "压力" in name:
            self.fields = lambda: {
                "pressure": round(1013 + random.gauss(0, 5), 1),
                "altitude": round(random.uniform(-10, 50), 1),
            }
        else:
            self.fields = lambda: {
                "value": round(random.uniform(0, 100), 1),
            }

    def connect(self):
        self.client = mqtt.Client(client_id=f"sim-{self.key}")
        self.client.username_pw_set(self.key, self.secret)

        self.client.on_connect = lambda c, u, f, rc: self._on_connect(rc)
        self.client.on_disconnect = lambda c, u, rc: self._on_disconnect()

        try:
            self.client.connect("localhost", 1883, 60)
            self.client.loop_start()
            time.sleep(0.5)
        except Exception as e:
            print(f"  [{self.name}] Connect failed: {e}")

    def _on_connect(self, rc):
        self.connected = True
        print(f"  [{self.name}] Connected")

    def _on_disconnect(self):
        self.connected = False

    def publish(self):
        if not self.connected:
            return
        topic = f"telemetry/{self.key}/data"
        payload = {
            **self.fields(),
            "device_name": self.name,
            "timestamp": int(time.time() * 1000),
        }
        self.client.publish(topic, json.dumps(payload), qos=0)

    def disconnect(self):
        if self.client:
            self.client.loop_stop()
            self.client.disconnect()


def main():
    print("=== IoT Device Simulator ===")
    print(f"Simulating {len(DEVICES)} devices")

    devices = [SimDevice(name, key, secret) for name, key, secret in DEVICES]

    print("\nConnecting devices...")
    for d in devices:
        d.connect()

    time.sleep(2)
    connected = sum(1 for d in devices if d.connected)
    print(f"\nConnected: {connected}/{len(devices)}")

    if connected == 0:
        print("No devices connected, exiting")
        return

    print("\nPublishing telemetry data (press Ctrl+C to stop)...\n")

    try:
        count = 0
        while True:
            for d in devices:
                d.publish()
            count += 1
            if count % 10 == 0:
                print(f"  [{time.strftime('%H:%M:%S')}] Published {count * connected} messages total")
            time.sleep(2)
    except KeyboardInterrupt:
        print("\n\nStopping simulator...")
    finally:
        for d in devices:
            d.disconnect()
        print("All devices disconnected. Bye!")


if __name__ == "__main__":
    main()
