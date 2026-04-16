<template>
  <div class="page-container">
    <div class="page-header">
      <h2>消息历史</h2>
    </div>

    <!-- Filters -->
    <div class="card-box filter-bar">
      <el-row :gutter="16" align="middle">
        <el-col :xs="24" :sm="8" :md="6">
          <el-input v-model="filters.topic" placeholder="搜索Topic" clearable :prefix-icon="Search" @clear="fetchData" @keyup.enter="fetchData" />
        </el-col>
        <el-col :xs="12" :sm="6" :md="4">
          <el-select v-model="filters.direction" placeholder="方向" clearable style="width: 100%" @change="fetchData">
            <el-option label="接收" value="in" />
            <el-option label="发送" value="out" />
          </el-select>
        </el-col>
        <el-col :xs="24" :sm="10" :md="8">
          <el-date-picker
            v-model="timeRange"
            type="datetimerange"
            range-separator="至"
            start-placeholder="开始时间"
            end-placeholder="结束时间"
            format="YYYY-MM-DD HH:mm"
            style="width: 100%"
            @change="fetchData"
          />
        </el-col>
        <el-col :xs="12" :sm="4" :md="3">
          <el-button type="primary" @click="fetchData">搜索</el-button>
        </el-col>
      </el-row>
    </div>

    <!-- Table -->
    <div class="card-box" style="margin-top: 16px;">
      <el-table v-loading="loading" :data="messages" stripe style="width: 100%">
        <el-table-column prop="id" label="#" width="60" />
        <el-table-column prop="topic" label="Topic" min-width="220">
          <template #default="{ row }">
            <el-text family="monospace" size="small" class="topic-text">{{ row.topic }}</el-text>
          </template>
        </el-table-column>
        <el-table-column prop="payload" label="Payload" min-width="280">
          <template #default="{ row }">
            <el-popover trigger="click" :width="400" v-if="row.payload && row.payload.length > 60">
              <template #reference>
                <el-text size="small" class="payload-text">{{ row.payload.substring(0, 60) }}...</el-text>
              </template>
              <div style="max-height: 300px; overflow: auto;">
                <pre style="white-space: pre-wrap; word-break: break-all; margin: 0; font-size: 12px;">{{ formatPayload(row.payload) }}</pre>
              </div>
            </el-popover>
            <el-text v-else size="small" class="payload-text">{{ row.payload }}</el-text>
          </template>
        </el-table-column>
        <el-table-column prop="qos" label="QoS" width="70" />
        <el-table-column prop="direction" label="方向" width="80">
          <template #default="{ row }">
            <el-tag :type="row.direction === 'in' ? 'success' : 'primary'" size="small">
              {{ row.direction === 'in' ? '接收' : '发送' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="device_name" label="设备" width="120">
          <template #default="{ row }">
            {{ row.device_name || '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="时间" width="170">
          <template #default="{ row }">
            {{ dayjs(row.created_at).format('YYYY-MM-DD HH:mm:ss') }}
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-wrapper">
        <el-pagination
          v-model:current-page="currentPage"
          v-model:page-size="pageSize"
          :total="total"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="fetchData"
          @current-change="fetchData"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { Search } from '@element-plus/icons-vue'
import dayjs from 'dayjs'
import { getMessageHistory } from '@/api/message'
import type { Message } from '@/types'

const loading = ref(false)
const messages = ref<Message[]>([])
const total = ref(0)
const currentPage = ref(1)
const pageSize = ref(20)
const timeRange = ref<[Date, Date] | null>(null)

const filters = reactive({
  topic: '',
  direction: '',
})

function formatPayload(payload: string): string {
  try {
    return JSON.stringify(JSON.parse(payload), null, 2)
  } catch {
    return payload
  }
}

async function fetchData() {
  loading.value = true
  try {
    const params: Record<string, unknown> = {
      page: currentPage.value,
      page_size: pageSize.value,
    }
    if (filters.topic) params.topic = filters.topic
    if (filters.direction) params.direction = filters.direction
    if (timeRange.value) {
      params.start_time = dayjs(timeRange.value[0]).toISOString()
      params.end_time = dayjs(timeRange.value[1]).toISOString()
    }
    const res = await getMessageHistory(params)
    messages.value = res.data || []
    total.value = res.total || 0
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchData()
})
</script>

<style lang="scss" scoped>
.filter-bar { margin-bottom: 0; }

.topic-text, .payload-text {
  font-family: 'Courier New', monospace;
  font-size: 12px;
  word-break: break-all;
}

.pagination-wrapper {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
  padding-top: 16px;
}
</style>
