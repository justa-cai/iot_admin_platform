<template>
  <div class="page-container">
    <div class="page-header">
      <h2>
        <el-button text @click="$router.push('/devices')">
          <el-icon><ArrowLeft /></el-icon>
        </el-button>
        设备详情
      </h2>
      <div>
        <el-button @click="$router.push(`/devices/${deviceId}/edit`)">
          <el-icon><Edit /></el-icon> 编辑
        </el-button>
      </div>
    </div>

    <!-- Device Info Card -->
    <div class="card-box device-info-card" v-loading="loading">
      <el-descriptions :column="4" border>
        <el-descriptions-item label="设备名称">{{ device.name }}</el-descriptions-item>
        <el-descriptions-item label="设备Key">
          <el-text family="monospace" size="small">{{ device.device_key }}</el-text>
        </el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="statusTagType(device.status || '')" effect="dark" size="small">
            {{ statusLabel(device.status || '') }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="分组">{{ device.group?.name || '-' }}</el-descriptions-item>
        <el-descriptions-item label="标签">
          <el-tag v-for="tag in (device.tags || [])" :key="tag.id" size="small" style="margin-right: 4px">
            {{ tag.name }}
          </el-tag>
          <span v-if="!device.tags?.length">-</span>
        </el-descriptions-item>
        <el-descriptions-item label="最后活跃">{{ device.last_active_at ? dayjs(device.last_active_at).format('YYYY-MM-DD HH:mm:ss') : '-' }}</el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ dayjs(device.created_at).format('YYYY-MM-DD HH:mm:ss') }}</el-descriptions-item>
        <el-descriptions-item label="更新时间">{{ dayjs(device.updated_at).format('YYYY-MM-DD HH:mm:ss') }}</el-descriptions-item>
      </el-descriptions>
    </div>

    <!-- Tabs -->
    <el-tabs v-model="activeTab" type="border-card" class="detail-tabs" @tab-change="handleTabChange">
      <!-- Telemetry Tab -->
      <el-tab-pane label="遥测数据" name="telemetry">
        <div class="tab-content">
          <div class="card-box">
            <div class="card-header">
              <h3>遥测趋势</h3>
              <div style="display: flex; gap: 12px; align-items: center;">
                <el-date-picker
                  v-model="telemetryTimeRange"
                  type="datetimerange"
                  range-separator="至"
                  start-placeholder="开始时间"
                  end-placeholder="结束时间"
                  size="small"
                  format="YYYY-MM-DD HH:mm"
                  @change="fetchTelemetry"
                />
              </div>
            </div>
            <LineChart :option="telemetryChartOption" height="400px" />
          </div>
        </div>
      </el-tab-pane>

      <!-- Messages Tab -->
      <el-tab-pane label="消息记录" name="messages">
        <div class="tab-content">
          <div class="card-box">
            <el-table :data="messages" stripe v-loading="messagesLoading" style="width: 100%">
              <el-table-column prop="topic" label="Topic" min-width="200">
                <template #default="{ row }">
                  <el-text family="monospace" size="small">{{ row.topic }}</el-text>
                </template>
              </el-table-column>
              <el-table-column prop="payload" label="Payload" min-width="250">
                <template #default="{ row }">
                  <el-text size="small" class="payload-text">{{ row.payload }}</el-text>
                </template>
              </el-table-column>
              <el-table-column prop="qos" label="QoS" width="80" />
              <el-table-column prop="direction" label="方向" width="80">
                <template #default="{ row }">
                  <el-tag :type="row.direction === 'in' ? 'success' : 'primary'" size="small">
                    {{ row.direction === 'in' ? '接收' : '发送' }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="created_at" label="时间" width="170">
                <template #default="{ row }">
                  {{ dayjs(row.created_at).format('HH:mm:ss') }}
                </template>
              </el-table-column>
            </el-table>
          </div>
        </div>
      </el-tab-pane>

      <!-- Send Command Tab -->
      <el-tab-pane label="发送指令" name="command">
        <div class="tab-content">
          <div class="card-box command-form">
            <el-form :model="commandForm" label-width="100px" label-position="top">
              <el-form-item label="Topic">
                <el-input v-model="commandForm.topic" placeholder="例如: devices/{device_key}/cmd" />
              </el-form-item>
              <el-form-item label="Payload">
                <el-input
                  v-model="commandForm.payload"
                  type="textarea"
                  :rows="8"
                  placeholder='例如: {"action": "reboot"}'
                />
              </el-form-item>
              <el-form-item label="QoS">
                <el-radio-group v-model="commandForm.qos">
                  <el-radio :value="0">QoS 0</el-radio>
                  <el-radio :value="1">QoS 1</el-radio>
                  <el-radio :value="2">QoS 2</el-radio>
                </el-radio-group>
              </el-form-item>
              <el-form-item>
                <el-button type="primary" :loading="sendingCommand" @click="sendCommand">
                  发送指令
                </el-button>
              </el-form-item>
            </el-form>
          </div>
        </div>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import dayjs from 'dayjs'
import type { EChartsOption } from 'echarts'
import { getDevice } from '@/api/device'
import { getMessageHistory, publishMessage } from '@/api/message'
import { queryTelemetry } from '@/api/telemetry'
import type { Device, Message, TelemetryData } from '@/types'
import LineChart from '@/components/charts/LineChart.vue'

const route = useRoute()
const deviceId = computed(() => Number(route.params.id))

const loading = ref(false)
const device = ref<Partial<Device>>({})
const activeTab = ref('telemetry')

// Messages
const messages = ref<Message[]>([])
const messagesLoading = ref(false)

// Telemetry
const telemetryTimeRange = ref<[Date, Date]>([
  dayjs().subtract(24, 'hour').toDate(),
  dayjs().toDate(),
])
const telemetryData = ref<TelemetryData[]>([])

// Command
const commandForm = ref({
  topic: '',
  payload: '',
  qos: 0,
})
const sendingCommand = ref(false)

const telemetryChartOption = computed<EChartsOption>(() => {
  const fieldMap = new Map<string, { times: string[]; values: number[] }>()
  for (const item of telemetryData.value) {
    if (!fieldMap.has(item.field)) {
      fieldMap.set(item.field, { times: [], values: [] })
    }
    const entry = fieldMap.get(item.field)!
    entry.times.push(dayjs(item.created_at).format('HH:mm'))
    entry.values.push(item.value)
  }

  const colors = ['#6366f1', '#10b981', '#f59e0b', '#ef4444', '#3b82f6', '#8b5cf6']
  const series: EChartsOption['series'] = []
  let index = 0
  for (const [field, data] of fieldMap) {
    series.push({
      name: field,
      type: 'line',
      smooth: true,
      data: data.values,
      lineStyle: { color: colors[index % colors.length] },
      itemStyle: { color: colors[index % colors.length] },
    })
    index++
  }

  const firstEntry = fieldMap.values().next().value
  return {
    tooltip: { trigger: 'axis' },
    legend: { top: 0 },
    grid: { left: '3%', right: '4%', bottom: '3%', top: '40px', containLabel: true },
    xAxis: {
      type: 'category',
      data: firstEntry?.times || [],
      boundaryGap: false,
    },
    yAxis: { type: 'value' },
    series,
  }
})

function statusTagType(status: string): 'success' | 'danger' | 'info' {
  const map: Record<string, 'success' | 'danger' | 'info'> = {
    online: 'success', offline: 'danger', inactive: 'info',
  }
  return map[status] || 'info'
}

function statusLabel(status: string): string {
  const map: Record<string, string> = { online: '在线', offline: '离线', inactive: '未激活' }
  return map[status] || status
}

async function fetchDevice() {
  loading.value = true
  try {
    device.value = await getDevice(deviceId.value)
    commandForm.value.topic = `devices/${(device.value as Device).device_key}/cmd`
  } finally {
    loading.value = false
  }
}

async function fetchTelemetry() {
  if (!telemetryTimeRange.value) return
  try {
    const res = await queryTelemetry({
      device_id: deviceId.value,
      start_time: dayjs(telemetryTimeRange.value[0]).toISOString(),
      end_time: dayjs(telemetryTimeRange.value[1]).toISOString(),
      page: 1,
      page_size: 500,
    })
    telemetryData.value = res.data || []
  } catch {
    // ignore
  }
}

async function fetchMessages() {
  messagesLoading.value = true
  try {
    const res = await getMessageHistory({
      device_id: deviceId.value,
      page: 1,
      page_size: 50,
    })
    messages.value = res.data || []
  } finally {
    messagesLoading.value = false
  }
}

async function sendCommand() {
  if (!commandForm.value.topic || !commandForm.value.payload) {
    ElMessage.warning('请填写Topic和Payload')
    return
  }
  sendingCommand.value = true
  try {
    await publishMessage({
      topic: commandForm.value.topic,
      payload: commandForm.value.payload,
      qos: commandForm.value.qos,
      device_id: deviceId.value,
    })
    ElMessage.success('指令已发送')
    commandForm.value.payload = ''
  } finally {
    sendingCommand.value = false
  }
}

function handleTabChange(tab: string | number) {
  if (tab === 'messages') fetchMessages()
  if (tab === 'telemetry') fetchTelemetry()
}

onMounted(async () => {
  await fetchDevice()
  fetchTelemetry()
})
</script>

<style lang="scss" scoped>
.device-info-card {
  margin-bottom: 20px;
}

.detail-tabs {
  margin-top: 0;
}

.tab-content {
  padding: 0;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
  h3 { font-size: 16px; font-weight: 600; color: #1f2937; }
}

.command-form {
  max-width: 700px;
}

.payload-text {
  font-family: 'Courier New', monospace;
  word-break: break-all;
}
</style>
