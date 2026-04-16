<template>
  <div class="page-container">
    <div class="page-header">
      <h2>数据查询</h2>
    </div>

    <!-- Query Panel -->
    <div class="card-box">
      <el-row :gutter="16" align="middle">
        <el-col :xs="24" :sm="6" :md="5">
          <el-select
            v-model="selectedDeviceId"
            placeholder="选择设备"
            filterable
            style="width: 100%"
            @change="handleDeviceChange"
          >
            <el-option
              v-for="d in devices"
              :key="d.id"
              :label="d.name"
              :value="d.id"
            />
          </el-select>
        </el-col>
        <el-col :xs="24" :sm="12" :md="10">
          <el-date-picker
            v-model="timeRange"
            type="datetimerange"
            range-separator="至"
            start-placeholder="开始时间"
            end-placeholder="结束时间"
            format="YYYY-MM-DD HH:mm"
            style="width: 100%"
          />
        </el-col>
        <el-col :xs="24" :sm="6" :md="4">
          <el-button type="primary" :loading="loading" @click="fetchTelemetry" style="width: 100%">
            查询
          </el-button>
        </el-col>
      </el-row>

      <!-- Quick time range buttons -->
      <div class="quick-range" style="margin-top: 12px;">
        <el-button-group>
          <el-button size="small" :type="activeQuickRange === '1h' ? 'primary' : ''" @click="setQuickRange('1h')">1小时</el-button>
          <el-button size="small" :type="activeQuickRange === '6h' ? 'primary' : ''" @click="setQuickRange('6h')">6小时</el-button>
          <el-button size="small" :type="activeQuickRange === '24h' ? 'primary' : ''" @click="setQuickRange('24h')">24小时</el-button>
          <el-button size="small" :type="activeQuickRange === '7d' ? 'primary' : ''" @click="setQuickRange('7d')">7天</el-button>
          <el-button size="small" :type="activeQuickRange === '30d' ? 'primary' : ''" @click="setQuickRange('30d')">30天</el-button>
        </el-button-group>
      </div>
    </div>

    <!-- Chart -->
    <div class="card-box" style="margin-top: 16px;">
      <div class="card-header">
        <h3>遥测数据趋势</h3>
      </div>
      <LineChart :option="chartOption" height="450px" />
      <el-empty v-if="!selectedDeviceId" description="请选择一个设备" :image-size="100" />
    </div>

    <!-- Latest Data Table -->
    <div class="card-box" style="margin-top: 16px;">
      <div class="card-header">
        <h3>最新数据</h3>
      </div>
      <el-table :data="latestData" stripe v-loading="latestLoading" style="width: 100%">
        <el-table-column prop="field" label="字段" width="200" />
        <el-table-column prop="value" label="值" min-width="200" />
        <el-table-column prop="timestamp" label="时间" width="200">
          <template #default="{ row }">
            {{ dayjs(row.timestamp).format('YYYY-MM-DD HH:mm:ss') }}
          </template>
        </el-table-column>
      </el-table>
      <el-empty v-if="latestData.length === 0 && !latestLoading" description="暂无数据" :image-size="60" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import dayjs from 'dayjs'
import type { EChartsOption } from 'echarts'
import { getDevices } from '@/api/device'
import { queryTelemetry, getLatestTelemetry } from '@/api/telemetry'
import type { Device, TelemetryData, TelemetryLatest } from '@/types'
import LineChart from '@/components/charts/LineChart.vue'

const devices = ref<Device[]>([])
const selectedDeviceId = ref<number | string>('')
const loading = ref(false)
const latestLoading = ref(false)
const timeRange = ref<[Date, Date]>([
  dayjs().subtract(24, 'hour').toDate(),
  dayjs().toDate(),
])
const activeQuickRange = ref('24h')
const telemetryData = ref<TelemetryData[]>([])
const latestData = ref<TelemetryLatest[]>([])

const chartOption = computed<EChartsOption>(() => {
  const fieldMap = new Map<string, { times: string[]; values: number[] }>()
  for (const item of telemetryData.value) {
    if (!fieldMap.has(item.field)) {
      fieldMap.set(item.field, { times: [], values: [] })
    }
    const entry = fieldMap.get(item.field)!
    entry.times.push(dayjs(item.created_at).format('MM-DD HH:mm'))
    entry.values.push(item.value)
  }

  const colors = ['#6366f1', '#10b981', '#f59e0b', '#ef4444', '#3b82f6', '#8b5cf6', '#ec4899', '#14b8a6']
  const series: EChartsOption['series'] = []
  let idx = 0
  for (const [field, data] of fieldMap) {
    series.push({
      name: field,
      type: 'line',
      smooth: true,
      symbol: 'circle',
      symbolSize: 4,
      data: data.values,
      lineStyle: { color: colors[idx % colors.length], width: 2 },
      itemStyle: { color: colors[idx % colors.length] },
      areaStyle: {
        color: {
          type: 'linear',
          x: 0, y: 0, x2: 0, y2: 1,
          colorStops: [
            { offset: 0, color: colors[idx % colors.length] + '30' },
            { offset: 1, color: colors[idx % colors.length] + '05' },
          ],
        },
      },
    })
    idx++
  }

  const firstEntry = fieldMap.values().next().value
  return {
    tooltip: {
      trigger: 'axis',
      backgroundColor: 'rgba(30, 27, 75, 0.9)',
      borderColor: '#4338ca',
      textStyle: { color: '#fff' },
    },
    legend: {
      top: 0,
      textStyle: { color: '#6b7280' },
    },
    grid: {
      left: '3%',
      right: '4%',
      bottom: '3%',
      top: '40px',
      containLabel: true,
    },
    xAxis: {
      type: 'category',
      data: firstEntry?.times || [],
      boundaryGap: false,
      axisLine: { lineStyle: { color: '#e5e7eb' } },
      axisLabel: { color: '#6b7280', fontSize: 11 },
    },
    yAxis: {
      type: 'value',
      axisLine: { show: false },
      axisLabel: { color: '#6b7280' },
      splitLine: { lineStyle: { color: '#f3f4f6' } },
    },
    dataZoom: [
      { type: 'inside', start: 0, end: 100 },
    ],
    series,
  }
})

function setQuickRange(range: string) {
  activeQuickRange.value = range
  const now = dayjs()
  const map: Record<string, [dayjs.Dayjs, dayjs.Dayjs]> = {
    '1h': [now.subtract(1, 'hour'), now],
    '6h': [now.subtract(6, 'hour'), now],
    '24h': [now.subtract(24, 'hour'), now],
    '7d': [now.subtract(7, 'day'), now],
    '30d': [now.subtract(30, 'day'), now],
  }
  const [start, end] = map[range] || map['24h']
  timeRange.value = [start.toDate(), end.toDate()]
  fetchTelemetry()
}

async function handleDeviceChange() {
  if (selectedDeviceId.value) {
    await fetchLatest()
    fetchTelemetry()
  }
}

async function fetchTelemetry() {
  if (!selectedDeviceId.value || !timeRange.value) return
  loading.value = true
  try {
    const res = await queryTelemetry({
      device_id: Number(selectedDeviceId.value),
      start_time: dayjs(timeRange.value[0]).toISOString(),
      end_time: dayjs(timeRange.value[1]).toISOString(),
      page: 1,
      page_size: 500,
    })
    telemetryData.value = res.data || []
  } finally {
    loading.value = false
  }
}

async function fetchLatest() {
  if (!selectedDeviceId.value) return
  latestLoading.value = true
  try {
    latestData.value = await getLatestTelemetry(Number(selectedDeviceId.value), 20)
  } finally {
    latestLoading.value = false
  }
}

onMounted(async () => {
  const res = await getDevices({ page: 1, page_size: 200 })
  devices.value = res.data || []
})
</script>

<style lang="scss" scoped>
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
  h3 { font-size: 16px; font-weight: 600; color: #1f2937; }
}
</style>
