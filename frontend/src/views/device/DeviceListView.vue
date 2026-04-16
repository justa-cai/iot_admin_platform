<template>
  <div class="page-container">
    <div class="page-header">
      <h2>设备管理</h2>
      <el-button type="primary" @click="$router.push('/devices/create')">
        <el-icon><Plus /></el-icon>
        添加设备
      </el-button>
    </div>

    <!-- Filters -->
    <div class="card-box filter-bar">
      <el-row :gutter="16" align="middle">
        <el-col :xs="24" :sm="8" :md="6">
          <el-input
            v-model="searchQuery"
            placeholder="搜索设备名称/Key"
            :prefix-icon="Search"
            clearable
            @clear="handleSearch"
            @keyup.enter="handleSearch"
          />
        </el-col>
        <el-col :xs="12" :sm="6" :md="4">
          <el-select v-model="statusFilter" placeholder="状态筛选" clearable style="width: 100%" @change="handleSearch">
            <el-option label="在线" value="online" />
            <el-option label="离线" value="offline" />
            <el-option label="未激活" value="inactive" />
          </el-select>
        </el-col>
        <el-col :xs="12" :sm="6" :md="4">
          <el-select v-model="groupFilter" placeholder="分组筛选" clearable style="width: 100%" @change="handleSearch">
            <el-option
              v-for="g in groups"
              :key="g.id"
              :label="g.name"
              :value="g.id"
            />
          </el-select>
        </el-col>
        <el-col :xs="12" :sm="4" :md="3">
          <el-button type="primary" @click="handleSearch">搜索</el-button>
        </el-col>
      </el-row>
    </div>

    <!-- Table -->
    <div class="card-box" style="margin-top: 16px;">
      <el-table
        v-loading="loading"
        :data="devices"
        stripe
        style="width: 100%"
        @sort-change="handleSortChange"
      >
        <el-table-column type="index" label="#" width="50" :index="indexMethod" />
        <el-table-column prop="name" label="设备名称" min-width="160">
          <template #default="{ row }">
            <el-link type="primary" @click="$router.push(`/devices/${row.id}`)">{{ row.name }}</el-link>
          </template>
        </el-table-column>
        <el-table-column prop="device_key" label="设备Key" min-width="140">
          <template #default="{ row }">
            <el-tooltip :content="row.device_key" placement="top">
              <el-text size="small" class="device-key">{{ row.device_key?.substring(0, 12) }}...</el-text>
            </el-tooltip>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)" size="small" effect="dark">
              {{ statusLabel(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="group" label="分组" width="120">
          <template #default="{ row }">
            {{ row.group?.name || '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="tags" label="标签" min-width="140">
          <template #default="{ row }">
            <el-tag
              v-for="tag in (row.tags || []).slice(0, 3)"
              :key="tag.id"
              size="small"
              style="margin-right: 4px"
            >
              {{ tag.name }}
            </el-tag>
            <el-tag
              v-if="(row.tags || []).length > 3"
              size="small"
              type="info"
            >
              +{{ row.tags.length - 3 }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="last_active_at" label="最后活跃" width="170">
          <template #default="{ row }">
            {{ row.last_active_at ? dayjs(row.last_active_at).format('YYYY-MM-DD HH:mm') : '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="170">
          <template #default="{ row }">
            {{ dayjs(row.created_at).format('YYYY-MM-DD HH:mm') }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button text type="primary" size="small" @click="$router.push(`/devices/${row.id}`)">
              详情
            </el-button>
            <el-button text type="warning" size="small" @click="$router.push(`/devices/${row.id}/edit`)">
              编辑
            </el-button>
            <el-popconfirm title="确定删除该设备吗?" @confirm="handleDelete(row.id)">
              <template #reference>
                <el-button text type="danger" size="small">删除</el-button>
              </template>
            </el-popconfirm>
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
          @size-change="handleSearch"
          @current-change="handleSearch"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { Search } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import dayjs from 'dayjs'
import { getDevices, deleteDevice } from '@/api/device'
import { getGroups } from '@/api/group'
import type { Device, Group } from '@/types'

const loading = ref(false)
const devices = ref<Device[]>([])
const groups = ref<Group[]>([])
const total = ref(0)
const currentPage = ref(1)
const pageSize = ref(20)
const searchQuery = ref('')
const statusFilter = ref('')
const groupFilter = ref<number | string>('')

function statusTagType(status: string): 'success' | 'danger' | 'info' {
  const map: Record<string, 'success' | 'danger' | 'info'> = {
    online: 'success',
    offline: 'danger',
    inactive: 'info',
  }
  return map[status] || 'info'
}

function statusLabel(status: string): string {
  const map: Record<string, string> = {
    online: '在线',
    offline: '离线',
    inactive: '未激活',
  }
  return map[status] || status
}

async function fetchData() {
  loading.value = true
  try {
    const res = await getDevices({
      page: currentPage.value,
      page_size: pageSize.value,
      search: searchQuery.value || undefined,
      status: statusFilter.value || undefined,
      group_id: groupFilter.value ? Number(groupFilter.value) : undefined,
    })
    devices.value = res.data || []
    total.value = res.total || 0
  } finally {
    loading.value = false
  }
}

function indexMethod(index: number) {
  return (currentPage.value - 1) * pageSize.value + index + 1
}

function handleSearch() {
  currentPage.value = 1
  fetchData()
}

function handleSortChange() {
  fetchData()
}

async function handleDelete(id: number) {
  await deleteDevice(id)
  ElMessage.success('设备已删除')
  fetchData()
}

onMounted(async () => {
  const groupRes = await getGroups()
  groups.value = groupRes || []
  fetchData()
})
</script>

<style lang="scss" scoped>
.filter-bar {
  margin-bottom: 0;
}

.device-key {
  font-family: 'Courier New', monospace;
  color: #6b7280;
  font-size: 12px;
}

.pagination-wrapper {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
  padding-top: 16px;
}
</style>
