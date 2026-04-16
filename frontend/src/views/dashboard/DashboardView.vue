<template>
  <div class="dashboard-page">
    <!-- Stat Cards -->
    <el-row :gutter="20" class="stat-row">
      <el-col :xs="12" :sm="12" :md="6">
        <StatCard
          :value="stats.total_devices"
          label="设备总数"
          icon="Monitor"
          gradient="indigo"
        />
      </el-col>
      <el-col :xs="12" :sm="12" :md="6">
        <StatCard
          :value="stats.online_devices"
          label="在线设备"
          icon="Connection"
          gradient="green"
        />
      </el-col>
      <el-col :xs="12" :sm="12" :md="6">
        <StatCard
          :value="stats.messages_today"
          label="今日消息"
          icon="ChatDotRound"
          gradient="cyan"
        />
      </el-col>
      <el-col :xs="12" :sm="12" :md="6">
        <StatCard
          :value="stats.active_rules"
          label="活跃规则"
          icon="SetUp"
          gradient="orange"
        />
      </el-col>
    </el-row>

    <!-- Charts Row -->
    <el-row :gutter="20" class="chart-row">
      <el-col :xs="24" :md="16">
        <div class="card-box">
          <div class="card-header">
            <h3>消息吞吐量趋势</h3>
            <el-radio-group v-model="throughputHours" size="small" @change="handleHoursChange">
              <el-radio-button :value="6">6小时</el-radio-button>
              <el-radio-button :value="12">12小时</el-radio-button>
              <el-radio-button :value="24">24小时</el-radio-button>
              <el-radio-button :value="72">3天</el-radio-button>
            </el-radio-group>
          </div>
          <LineChart :option="throughputOption" height="350px" />
        </div>
      </el-col>
      <el-col :xs="24" :md="8">
        <div class="card-box">
          <div class="card-header">
            <h3>设备状态分布</h3>
          </div>
          <LineChart :option="deviceStatusOption" height="350px" />
        </div>
      </el-col>
    </el-row>

    <!-- Bottom Row: Alerts + Online Rate -->
    <el-row :gutter="20" class="bottom-row">
      <el-col :xs="24" :md="16">
        <div class="card-box">
          <div class="card-header">
            <h3>最近告警</h3>
            <el-button text type="primary" @click="$router.push('/rules')">查看全部</el-button>
          </div>
          <el-table :data="recentAlerts" stripe style="width: 100%" max-height="320">
            <el-table-column prop="event" label="事件类型" width="160">
              <template #default="{ row }">
                <el-tag :type="getAlertType(row.event)" size="small">
                  {{ getAlertLabel(row.event) }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="data" label="详情" min-width="200">
              <template #default="{ row }">
                {{ formatAlertData(row) }}
              </template>
            </el-table-column>
            <el-table-column prop="timestamp" label="时间" width="180">
              <template #default="{ row }">
                {{ formatTime(row.timestamp) }}
              </template>
            </el-table-column>
          </el-table>
          <el-empty v-if="recentAlerts.length === 0" description="暂无告警" :image-size="80" />
        </div>
      </el-col>
      <el-col :xs="24" :md="8">
        <div class="card-box" style="display: flex; flex-direction: column; align-items: center;">
          <div class="card-header" style="width: 100%;">
            <h3>设备在线率</h3>
          </div>
          <GaugeChart
            :value="onlineRate"
            name="在线率"
            height="280px"
          />
          <div class="online-stats">
            <div class="online-stat-item">
              <span class="dot online"></span>
              <span>在线: {{ stats.online_devices }}</span>
            </div>
            <div class="online-stat-item">
              <span class="dot offline"></span>
              <span>离线: {{ stats.offline_devices }}</span>
            </div>
            <div class="online-stat-item">
              <span class="dot inactive"></span>
              <span>未激活: {{ stats.inactive_devices }}</span>
            </div>
          </div>
        </div>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import dayjs from 'dayjs'
import type { EChartsOption } from 'echarts'
import { useDashboardStore } from '@/stores/dashboard'
import { useWebSocket } from '@/composables/useWebSocket'
import StatCard from '@/components/common/StatCard.vue'
import LineChart from '@/components/charts/LineChart.vue'
import GaugeChart from '@/components/charts/GaugeChart.vue'
import type { WSEvent } from '@/types'

const dashboardStore = useDashboardStore()
const { events, latestEvent, connect, clearEvents } = useWebSocket()

const stats = computed(() => dashboardStore.stats)
const throughputHours = ref(24)
const recentAlerts = ref<WSEvent[]>([])

const onlineRate = computed(() => {
  const total = stats.value.total_devices
  if (total === 0) return 0
  return Math.round((stats.value.online_devices / total) * 100)
})

const throughputOption = computed<EChartsOption>(() => {
  const data = dashboardStore.throughput
  return {
    tooltip: {
      trigger: 'axis',
      backgroundColor: 'rgba(30, 27, 75, 0.9)',
      borderColor: '#4338ca',
      textStyle: { color: '#fff' },
    },
    grid: {
      left: '3%',
      right: '4%',
      bottom: '3%',
      top: '10%',
      containLabel: true,
    },
    xAxis: {
      type: 'category',
      data: data.timestamps.map((t: string) => dayjs(t).format('HH:mm')),
      boundaryGap: false,
      axisLine: { lineStyle: { color: '#e5e7eb' } },
      axisLabel: { color: '#6b7280' },
    },
    yAxis: {
      type: 'value',
      axisLine: { show: false },
      axisLabel: { color: '#6b7280' },
      splitLine: { lineStyle: { color: '#f3f4f6' } },
    },
    series: [
      {
        name: '消息数',
        type: 'line',
        smooth: true,
        data: data.values,
        lineStyle: { color: '#6366f1', width: 3 },
        itemStyle: { color: '#6366f1' },
        areaStyle: {
          color: {
            type: 'linear',
            x: 0, y: 0, x2: 0, y2: 1,
            colorStops: [
              { offset: 0, color: 'rgba(99, 102, 241, 0.35)' },
              { offset: 1, color: 'rgba(99, 102, 241, 0.02)' },
            ],
          },
        },
      },
    ],
  }
})

const deviceStatusOption = computed<EChartsOption>(() => ({
  tooltip: {
    trigger: 'item',
    backgroundColor: 'rgba(30, 27, 75, 0.9)',
    borderColor: '#4338ca',
    textStyle: { color: '#fff' },
  },
  legend: {
    bottom: '5%',
    left: 'center',
    textStyle: { color: '#6b7280' },
  },
  series: [
    {
      name: '设备状态',
      type: 'pie',
      radius: ['45%', '70%'],
      center: ['50%', '45%'],
      avoidLabelOverlap: false,
      itemStyle: {
        borderRadius: 8,
        borderColor: '#fff',
        borderWidth: 2,
      },
      label: {
        show: true,
        position: 'center',
        formatter: `{total|${stats.value.total_devices}}\n{label|设备总数}`,
        rich: {
          total: {
            fontSize: 28,
            fontWeight: 'bold',
            color: '#1f2937',
          },
          label: {
            fontSize: 12,
            color: '#9ca3af',
            padding: [4, 0, 0, 0],
          },
        },
      },
      emphasis: {
        label: { show: true },
      },
      data: [
        {
          value: stats.value.online_devices,
          name: '在线',
          itemStyle: { color: '#10b981' },
        },
        {
          value: stats.value.offline_devices,
          name: '离线',
          itemStyle: { color: '#ef4444' },
        },
        {
          value: stats.value.inactive_devices,
          name: '未激活',
          itemStyle: { color: '#9ca3af' },
        },
      ],
    },
  ],
}))

async function handleHoursChange(val: string | number | boolean | undefined) {
  await dashboardStore.fetchThroughput(Number(val))
}

function getAlertType(event: string): 'success' | 'warning' | 'danger' | 'info' {
  if (event === 'device.status_changed') return 'warning'
  if (event === 'rule.triggered') return 'danger'
  return 'info'
}

function getAlertLabel(event: string): string {
  const map: Record<string, string> = {
    'device.status_changed': '设备状态变更',
    'message.received': '消息接收',
    'rule.triggered': '规则触发',
  }
  return map[event] || event
}

function formatAlertData(row: WSEvent): string {
  const d = row.data
  if (row.event === 'device.status_changed') {
    return `设备 ${d.device_name || d.device_id} 状态变为 ${d.status}`
  }
  if (row.event === 'rule.triggered') {
    return `规则 "${d.rule_name || d.rule_id}" 已触发`
  }
  if (row.event === 'message.received') {
    return `收到来自 ${d.topic || 'unknown'} 的消息`
  }
  return JSON.stringify(d)
}

function formatTime(ts: string): string {
  return dayjs(ts).format('YYYY-MM-DD HH:mm:ss')
}

let refreshTimer: ReturnType<typeof setInterval> | null = null

onMounted(async () => {
  await dashboardStore.refresh()
  connect()
  refreshTimer = setInterval(() => {
    dashboardStore.fetchStats()
  }, 30000)
})

onUnmounted(() => {
  if (refreshTimer) {
    clearInterval(refreshTimer)
    refreshTimer = null
  }
  clearEvents()
})

// Watch for WS events and add to alerts
import { watch } from 'vue'
watch(latestEvent, (event) => {
  if (event) {
    recentAlerts.value.unshift(event)
    if (recentAlerts.value.length > 20) {
      recentAlerts.value = recentAlerts.value.slice(0, 20)
    }
    // Update stats on device status change
    if (event.event === 'device.status_changed') {
      dashboardStore.fetchStats()
    }
  }
})
</script>

<style lang="scss" scoped>
.dashboard-page {
  .stat-row {
    margin-bottom: 20px;
  }

  .chart-row {
    margin-bottom: 20px;
  }

  .card-box {
    height: 100%;
  }

  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 16px;

    h3 {
      font-size: 16px;
      font-weight: 600;
      color: #1f2937;
    }
  }

  .online-stats {
    display: flex;
    gap: 24px;
    margin-top: 12px;

    .online-stat-item {
      display: flex;
      align-items: center;
      gap: 6px;
      font-size: 13px;
      color: #6b7280;

      .dot {
        width: 8px;
        height: 8px;
        border-radius: 50%;
        display: inline-block;

        &.online { background-color: #10b981; }
        &.offline { background-color: #ef4444; }
        &.inactive { background-color: #9ca3af; }
      }
    }
  }
}
</style>
