<template>
  <div class="page-container">
    <div class="page-header">
      <h2>规则引擎</h2>
      <el-button type="primary" @click="$router.push('/rules/create')">
        <el-icon><Plus /></el-icon>
        创建规则
      </el-button>
    </div>

    <div class="card-box">
      <el-table v-loading="loading" :data="rules" stripe style="width: 100%">
        <el-table-column type="index" label="#" width="50" />
        <el-table-column prop="name" label="规则名称" min-width="160">
          <template #default="{ row }">
            <el-link type="primary" @click="$router.push(`/rules/${row.id}/edit`)">{{ row.name }}</el-link>
          </template>
        </el-table-column>
        <el-table-column prop="topic_pattern" label="Topic匹配" min-width="160">
          <template #default="{ row }">
            <el-text size="small" type="info">{{ row.topic_pattern }}</el-text>
          </template>
        </el-table-column>
        <el-table-column prop="action_type" label="动作类型" width="100">
          <template #default="{ row }">
            <el-tag size="small" :type="row.action_type === 'alert' ? 'danger' : row.action_type === 'publish' ? 'warning' : 'info'">
              {{ row.action_type === 'alert' ? '告警' : row.action_type === 'publish' ? '发布' : '转发' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="description" label="描述" min-width="160">
          <template #default="{ row }">
            {{ row.description || '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="enabled" label="状态" width="100">
          <template #default="{ row }">
            <el-switch
              v-model="row.enabled"
              :loading="row._toggling"
              active-color="#6366f1"
              @change="handleToggle(row)"
            />
          </template>
        </el-table-column>
        <el-table-column prop="last_triggered_at" label="最后触发" width="170">
          <template #default="{ row }">
            {{ row.last_triggered_at ? dayjs(row.last_triggered_at).format('YYYY-MM-DD HH:mm:ss') : '-' }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button text type="primary" size="small" @click="showLogs(row)">日志</el-button>
            <el-button text type="warning" size="small" @click="$router.push(`/rules/${row.id}/edit`)">编辑</el-button>
            <el-popconfirm title="确定删除该规则吗?" @confirm="handleDelete(row.id)">
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
          :page-sizes="[10, 20, 50]"
          layout="total, sizes, prev, pager, next"
          @size-change="fetchData"
          @current-change="fetchData"
        />
      </div>
    </div>

    <!-- Log Dialog -->
    <el-dialog v-model="showLogDialog" :title="`规则日志 - ${selectedRuleName}`" width="700px">
      <el-table :data="ruleLogs" stripe v-loading="logsLoading" style="width: 100%">
        <el-table-column prop="trigger_data" label="触发数据" min-width="200">
          <template #default="{ row }">
            <el-text size="small" style="word-break: break-all;">{{ row.trigger_data }}</el-text>
          </template>
        </el-table-column>
        <el-table-column prop="action_result" label="执行结果" min-width="200">
          <template #default="{ row }">
            <el-text size="small" style="word-break: break-all;">{{ row.action_result }}</el-text>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="时间" width="170">
          <template #default="{ row }">
            {{ dayjs(row.created_at).format('MM-DD HH:mm:ss') }}
          </template>
        </el-table-column>
      </el-table>
      <el-empty v-if="ruleLogs.length === 0 && !logsLoading" description="暂无触发记录" :image-size="60" />
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import dayjs from 'dayjs'
import { getRules, deleteRule, enableRule, disableRule, getRuleLogs } from '@/api/rule'
import type { Rule, RuleLog } from '@/types'

const loading = ref(false)
const rules = ref<(Rule & { _toggling?: boolean })[]>([])
const total = ref(0)
const currentPage = ref(1)
const pageSize = ref(20)

const showLogDialog = ref(false)
const selectedRuleName = ref('')
const selectedRuleId = ref('')
const ruleLogs = ref<RuleLog[]>([])
const logsLoading = ref(false)

async function fetchData() {
  loading.value = true
  try {
    const res = await getRules({
      page: currentPage.value,
      page_size: pageSize.value,
    })
    rules.value = res.data || []
    total.value = res.total || 0
  } finally {
    loading.value = false
  }
}

async function handleToggle(row: Rule & { _toggling?: boolean }) {
  row._toggling = true
  try {
    if (row.enabled) {
      await enableRule(row.id)
      ElMessage.success('规则已启用')
    } else {
      await disableRule(row.id)
      ElMessage.success('规则已禁用')
    }
  } catch {
    row.enabled = !row.enabled
  } finally {
    row._toggling = false
  }
}

async function handleDelete(id: string) {
  await deleteRule(id)
  ElMessage.success('规则已删除')
  fetchData()
}

async function showLogs(row: Rule) {
  selectedRuleName.value = row.name
  selectedRuleId.value = row.id
  showLogDialog.value = true
  logsLoading.value = true
  try {
    const res = await getRuleLogs(row.id, { page: 1, page_size: 50 })
    ruleLogs.value = res.data || []
  } finally {
    logsLoading.value = false
  }
}

onMounted(() => {
  fetchData()
})
</script>

<style lang="scss" scoped>
.pagination-wrapper {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
  padding-top: 16px;
}
</style>
