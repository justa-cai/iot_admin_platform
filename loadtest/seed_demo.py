#!/usr/bin/env python3
"""Populate IoT Admin with demo data."""

import requests
import json
import time
import random
import uuid
import sys

BASE = "http://localhost:8080/api/v1"

def login():
    r = requests.post(f"{BASE}/auth/login", json={"username": "admin", "password": "admin123"})
    return r.json()["token"]

def api(method, path, token, data=None):
    headers = {"Authorization": f"Bearer {token}", "Content-Type": "application/json"}
    url = f"{BASE}{path}"
    if method == "GET":
        r = requests.get(url, headers=headers)
    elif method == "POST":
        r = requests.post(url, headers=headers, json=data)
    elif method == "PUT":
        r = requests.put(url, headers=headers, json=data)
    elif method == "DELETE":
        r = requests.delete(url, headers=headers)
    return r.json()

def main():
    token = login()
    print(f"Logged in, token: {token[:20]}...")

    # ========== Groups ==========
    print("\n=== Groups ===")
    groups_data = api("GET", "/groups", token)
    groups = {g["name"]: g["id"] for g in groups_data.get("data", [])}

    if "车间设备" not in groups:
        for name, desc in [("车间设备", "生产车间相关设备"), ("环境监测", "温湿度环境监测设备"),
                          ("安防设备", "门禁、摄像头等安防设备"), ("网关设备", "边缘网关和中继设备")]:
            r = api("POST", "/groups", token, {"name": name, "description": desc})
            groups[name] = r["id"]
            print(f"  Created group: {name}")

    groups_data = api("GET", "/groups", token)
    groups = {g["name"]: g["id"] for g in groups_data.get("data", [])}
    print(f"  Total groups: {len(groups)}")

    # ========== Tags ==========
    print("\n=== Tags ===")
    tags_data = api("GET", "/tags", token)
    tags = {t["name"]: t["id"] for t in tags_data.get("data", [])}

    needed_tags = [("重要", "#F56C6C"), ("生产线A", "#409EFF"), ("生产线B", "#67C23A"),
                   ("测试中", "#E6A23C"), ("已部署", "#909399"), ("待维修", "#F56C6C")]
    for name, color in needed_tags:
        if name not in tags:
            r = api("POST", "/tags", token, {"name": name, "color": color})
            tags[name] = r["id"]
            print(f"  Created tag: {name}")

    tags_data = api("GET", "/tags", token)
    tags = {t["name"]: t["id"] for t in tags_data.get("data", [])}
    print(f"  Total tags: {len(tags)}")

    # ========== Devices ==========
    print("\n=== Devices ===")
    devices_data = api("GET", "/devices", token)
    existing_devices = {d["name"]: d for d in devices_data.get("data", [])}

    new_devices = [
        {"name": "温度传感器-A01", "group": "车间设备", "tags": ["生产线A", "已部署"],
         "metadata": {"location": "车间A-1号线", "model": "DHT22", "firmware": "v2.1.0"}},
        {"name": "温度传感器-A02", "group": "车间设备", "tags": ["生产线A", "已部署"],
         "metadata": {"location": "车间A-2号线", "model": "DHT22", "firmware": "v2.1.0"}},
        {"name": "温度传感器-B01", "group": "车间设备", "tags": ["生产线B", "已部署"],
         "metadata": {"location": "车间B-1号线", "model": "DHT22", "firmware": "v2.1.0"}},
        {"name": "湿度传感器-A01", "group": "环境监测", "tags": ["生产线A", "重要"],
         "metadata": {"location": "车间A", "model": "SHT30", "firmware": "v1.8.3"}},
        {"name": "湿度传感器-B01", "group": "环境监测", "tags": ["生产线B"],
         "metadata": {"location": "车间B", "model": "SHT30", "firmware": "v1.8.3"}},
        {"name": "压力传感器-01", "group": "车间设备", "tags": ["生产线A", "重要", "已部署"],
         "metadata": {"location": "管道入口", "model": "BMP280", "firmware": "v1.5.0"}},
        {"name": "烟雾报警器-01", "group": "安防设备", "tags": ["重要", "已部署"],
         "metadata": {"location": "仓库入口", "model": "MQ-2", "firmware": "v3.0.1"}},
        {"name": "烟雾报警器-02", "group": "安防设备", "tags": ["已部署"],
         "metadata": {"location": "配电间", "model": "MQ-2", "firmware": "v3.0.1"}},
        {"name": "门禁控制器-01", "group": "安防设备", "tags": ["重要", "已部署"],
         "metadata": {"location": "正门", "model": "RC522", "firmware": "v1.2.0"}},
        {"name": "智能电表-01", "group": "车间设备", "tags": ["生产线A", "已部署"],
         "metadata": {"location": "配电柜A", "model": "PZEM-004T", "firmware": "v2.0.0"}},
        {"name": "智能电表-02", "group": "车间设备", "tags": ["生产线B", "已部署"],
         "metadata": {"location": "配电柜B", "model": "PZEM-004T", "firmware": "v2.0.0"}},
        {"name": "光照传感器-01", "group": "环境监测", "tags": ["已部署"],
         "metadata": {"location": "大棚区域", "model": "BH1750", "firmware": "v1.0.2"}},
        {"name": "GPS定位器-01", "group": "网关设备", "tags": ["重要"],
         "metadata": {"location": "运输车辆", "model": "NEO-6M", "firmware": "v1.1.0"}},
        {"name": "边缘网关-01", "group": "网关设备", "tags": ["重要", "已部署"],
         "metadata": {"location": "机房", "model": "ESP32-GW", "firmware": "v3.2.1"}},
        {"name": "边缘网关-02", "group": "网关设备", "tags": ["待维修"],
         "metadata": {"location": "仓库", "model": "ESP32-GW", "firmware": "v3.1.0"}},
        {"name": "振动传感器-01", "group": "车间设备", "tags": ["生产线A", "测试中"],
         "metadata": {"location": "冲压机旁", "model": "MPU6050", "firmware": "v0.9.0"}},
        {"name": "噪声传感器-01", "group": "环境监测", "tags": ["测试中"],
         "metadata": {"location": "厂区边界", "model": "MAX4466", "firmware": "v0.5.0"}},
        {"name": "水位传感器-01", "group": "环境监测", "tags": ["生产线B", "已部署"],
         "metadata": {"location": "蓄水池", "model": "HC-SR04", "firmware": "v1.3.0"}},
    ]

    created_devices = {}
    for dev in new_devices:
        if dev["name"] in existing_devices:
            created_devices[dev["name"]] = existing_devices[dev["name"]]
            continue
        key = f"DEV-{uuid.uuid4().hex[:16]}"
        secret = f"SEC-{uuid.uuid4().hex[:16]}"
        payload = {
            "name": dev["name"],
            "device_key": key,
            "device_secret": secret,
            "metadata": json.dumps(dev.get("metadata", {})),
            "group_id": groups.get(dev.get("group"), ""),
            "tag_ids": [tags[t] for t in dev.get("tags", []) if t in tags],
        }
        r = api("POST", "/devices", token, payload)
        created_devices[dev["name"]] = r
        print(f"  Created: {dev['name']}")

    print(f"  Total devices created: {len(created_devices)}")

    # ========== Users ==========
    print("\n=== Users ===")
    users = [
        {"username": "operator1", "password": "oper123", "role": "operator"},
        {"username": "viewer1", "password": "view123", "role": "viewer"},
        {"username": "operator2", "password": "oper456", "role": "operator"},
    ]
    for u in users:
        r = api("POST", "/auth/register", token, u)
        status = r.get("username", r.get("error", "?"))
        print(f"  User: {u['username']} ({u['role']}) -> {status}")

    # ========== Rules ==========
    print("\n=== Rules ===")
    rules = [
        {"name": "高温告警", "description": "温度超过80度触发告警", "topic_pattern": "telemetry/+/data",
         "condition": json.dumps({"field": "temperature", "operator": "gt", "value": 80}),
         "action_type": "alert", "action_config": json.dumps({"message": "温度超过80度！请立即检查设备！"}),
         "cooldown_secs": 60},
        {"name": "低温预警", "description": "温度低于5度触发预警", "topic_pattern": "telemetry/+/data",
         "condition": json.dumps({"field": "temperature", "operator": "lt", "value": 5}),
         "action_type": "alert", "action_config": json.dumps({"message": "温度过低，可能结冰！"}),
         "cooldown_secs": 120},
        {"name": "湿度异常告警", "description": "湿度超过90%触发告警", "topic_pattern": "telemetry/+/data",
         "condition": json.dumps({"field": "humidity", "operator": "gt", "value": 90}),
         "action_type": "alert", "action_config": json.dumps({"message": "湿度过高，注意防潮！"}),
         "cooldown_secs": 300},
        {"name": "烟雾报警", "description": "烟雾浓度超过阈值", "topic_pattern": "telemetry/+/smoke",
         "condition": json.dumps({"field": "smoke_level", "operator": "gt", "value": 500}),
         "action_type": "alert", "action_config": json.dumps({"message": "烟雾浓度异常！可能发生火灾！"}),
         "cooldown_secs": 30},
        {"name": "电量不足提醒", "description": "电池电量低于20%", "topic_pattern": "telemetry/+/battery",
         "condition": json.dumps({"field": "battery", "operator": "lt", "value": 20}),
         "action_type": "alert", "action_config": json.dumps({"message": "设备电量不足，请及时充电或更换电池"}),
         "cooldown_secs": 3600},
    ]
    for rule in rules:
        r = api("POST", "/rules", token, rule)
        name = r.get("name", r.get("error", "?"))
        print(f"  Rule: {rule['name']} -> {name}")

    # ========== Generate Telemetry via MQTT publish API ==========
    print("\n=== Generating Telemetry Data ===")

    # Get all devices with their keys
    devices_data = api("GET", "/devices", token)
    all_devices = devices_data.get("data", [])

    device_names = [d["name"] for d in all_devices]
    print(f"  Found {len(all_devices)} devices")

    # Generate historical telemetry data via direct message publishing
    # Use the message publish API to create message history
    topics_data = [
        ("telemetry/temperature", "temperature"),
        ("telemetry/humidity", "humidity"),
        ("telemetry/pressure", "pressure"),
        ("telemetry/smoke", "smoke_level"),
        ("telemetry/battery", "battery"),
        ("telemetry/light", "light"),
    ]

    total_published = 0
    for i in range(200):
        dev = random.choice(all_devices)
        topic_type, field = random.choice(topics_data)

        # Generate realistic values
        if field == "temperature":
            val = round(random.gauss(25, 15), 1)
        elif field == "humidity":
            val = round(random.gauss(55, 20), 1)
        elif field == "pressure":
            val = round(random.gauss(1013, 10), 1)
        elif field == "smoke_level":
            val = round(random.gauss(50, 80), 1)
        elif field == "battery":
            val = round(random.gauss(70, 25), 1)
        elif field == "light":
            val = round(random.gauss(500, 300), 1)
        else:
            val = round(random.random() * 100, 1)

        payload_data = {
            field: val,
            "device_name": dev["name"],
            "timestamp": int(time.time() * 1000) - random.randint(0, 3600000),
        }

        r = api("POST", "/messages/publish", token, {
            "topic": f"{topic_type}/{dev['id'][:8]}",
            "payload": json.dumps(payload_data),
            "qos": 0,
        })
        total_published += 1

    print(f"  Published {total_published} messages")

    # ========== Summary ==========
    print("\n" + "=" * 50)
    print("Demo data population complete!")
    print(f"  Groups: {len(groups)}")
    print(f"  Tags: {len(tags)}")
    print(f"  Devices: {len(all_devices)}")
    print(f"  Rules: {len(rules)}")
    print(f"  Messages: {total_published}")
    print("=" * 50)

if __name__ == "__main__":
    main()
