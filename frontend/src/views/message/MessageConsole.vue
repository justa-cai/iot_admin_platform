<template>
  <div class="page-container">
    <div class="page-header">
      <h2>消息控制台</h2>
      <el-button type="primary" @click="showPublishDialog = true">
        <el-icon><Promotion /></el-icon>
        发布消息
      </el-button>
    </div>

    <el-row :gutter="20">
      <!-- Message Stream -->
      <el-col :xs="24" :md="18">
        <div class="card-box">
          <div class="card-header">
            <h3>
              实时消息流
              <el-tag :type="wsConnected ? 'success' : 'danger'" size="small" style="margin-left: 8px">
                {{ wsConnected ? '已连接' : '未连接' }}
              </el-tag>
            </h3>
            <el-button text type="danger" @click="clearEvents" size="small">清空</el-button>
          </div>

          <div class="message-stream" ref="streamRef">
            <div
              v-for="(msg, idx) in displayEvents"
              :key="idx"
              class="message-item"
              :class="{ 'msg-in': isDirectionIn(msg), 'msg-rule': msg.event === 'rule.triggered' }"
            >
              <div class="msg-header">
                <el-tag :type="getEventTagType(msg.event)" size="small">
                  {{ getEventLabel(msg.event) }}
                </el-tag>
                <span class="msg-time">{{ formatTime(msg.timestamp) }}</span>
              </div>
              <div class="msg-body">
                <template v-if="msg.event === 'message.received'">
                  <div class="msg-field">
                    <strong>Topic:</strong>
                    <code>{{ msg.data.topic || '-' }}</code>
                  </div>
                  <div class="msg-field">
                    <strong>Payload:</strong>
                    <code>{{ msg.data.payload || '-' }}</code>
                  </div>
                </template>
                <template v-else-if="msg.event === 'device.status_changed'">
                  <div class="msg-field">
                    设备 <strong>{{ msg.data.device_name || msg.data.device_id }}</strong>
                    状态变更为
                    <el-tag :type="msg.data.status === 'online' ? 'success' : 'danger'" size="small">
                      {{ msg.data.status }}
                    </el-tag>
                  </div>
                </template>
                <template v-else-if="msg.event === 'rule.triggered'">
                  <div class="msg-field">
                    规则 <strong>{{ msg.data.rule_name || msg.data.rule_id }}</strong> 已触发
                  </div>
                  <div class="msg-field" v-if="msg.data.action_result">
                    <strong>结果:</strong> {{ msg.data.action_result }}
                  </div>
                </template>
                <template v-else>
                  <code>{{ JSON.stringify(msg.data) }}</code>
                </template>
              </div>
            </div>
            <el-empty v-if="displayEvents.length === 0" description="等待消息..." :image-size="80" />
          </div>
        </div>
      </el-col>

      <!-- Topics Panel -->
      <el-col :xs="24" :md="6">
        <div class="card-box">
          <div class="card-header">
            <h3>主题列表</h3>
            <el-button text type="primary" size="small" @click="fetchTopics">
              <el-icon><Refresh /></el-icon>
            </el-button>
          </div>
          <div class="topic-list">
            <div v-for="topic in topics" :key="topic.topic" class="topic-item" @click="selectTopic(topic.topic)">
              <code class="topic-name">{{ topic.topic }}</code>
              <el-tag size="small" type="info">{{ topic.subscriber_count }}</el-tag>
            </div>
            <el-empty v-if="topics.length === 0" description="暂无主题" :image-size="60" />
          </div>
        </div>
      </el-col>
    </el-row>

    <!-- Publish Dialog -->
    <el-dialog v-model="showPublishDialog" title="发布消息" width="560px">
      <el-form :model="publishForm" label-width="80px">
        <el-form-item label="Topic">
          <el-input v-model="publishForm.topic" placeholder="请输入或选择Topic" />
        </el-form-item>
        <el-form-item label="Payload">
          <el-input
            v-model="publishForm.payload"
            type="textarea"
            :rows="6"
            placeholder='例如: {"temperature": 25.6}'
          />
        </el-form-item>
        <el-form-item label="QoS">
          <el-radio-group v-model="publishForm.qos">
            <el-radio :value="0">QoS 0</el-radio>
            <el-radio :value="1">QoS 1</el-radio>
            <el-radio :value="2">QoS 2</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showPublishDialog = false">取消</el-button>
        <el-button type="primary" :loading="publishing" @click="handlePublish">发布</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, nextTick, watch } from 'vue'
import { ElMessage } from 'element-plus'
import dayjs from 'dayjs'
import { useWebSocket } from '@/composables/useWebSocket'
import { publishMessage, getTopics } from '@/api/message'
import type { TopicInfo, WSEvent } from '@/types'

const { connected: wsConnected, events, latestEvent, connect, clearEvents } = useWebSocket()

const streamRef = ref<HTMLElement | null>(null)
const showPublishDialog = ref(false)
const publishing = ref(false)
const topics = ref<TopicInfo[]>([])

const displayEvents = computed(() => {
  return [...events.value].reverse().slice(0, 100)
})

const publishForm = reactive({
  topic: '',
  payload: '',
  qos: 0,
})

function isDirectionIn(msg: WSEvent): boolean {
  return msg.event === 'message.received'
}

function getEventTagType(event: string): 'success' | 'warning' | 'danger' | 'info' {
  const map: Record<string, 'success' | 'warning' | 'danger' | 'info'> = {
    'message.received': 'success',
    'device.status_changed': 'warning',
    'rule.triggered': 'danger',
  }
  return map[event] || 'info'
}

function getEventLabel(event: string): string {
  const map: Record<string, string> = {
    'message.received': '消息接收',
    'device.status_changed': '状态变更',
    'rule.triggered': '规则触发',
  }
  return map[event] || event
}

function formatTime(ts: string): string {
  return dayjs(ts).format('HH:mm:ss.SSS')
}

function selectTopic(topic: string) {
  publishForm.topic = topic
}

async function fetchTopics() {
  try {
    topics.value = await getTopics()
  } catch {
    topics.value = []
  }
}

async function handlePublish() {
  if (!publishForm.topic || !publishForm.payload) {
    ElMessage.warning('请填写Topic和Payload')
    return
  }
  publishing.value = true
  try {
    await publishMessage({
      topic: publishForm.topic,
      payload: publishForm.payload,
      qos: publishForm.qos,
    })
    ElMessage.success('消息已发布')
    showPublishDialog.value = false
    publishForm.topic = ''
    publishForm.payload = ''
  } finally {
    publishing.value = false
  }
}

function scrollToBottom() {
  nextTick(() => {
    if (streamRef.value) {
      streamRef.value.scrollTop = 0
    }
  })
}

watch(latestEvent, () => {
  scrollToBottom()
})

onMounted(() => {
  connect()
  fetchTopics()
})
</script>

<style lang="scss" scoped>
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
  h3 { font-size: 16px; font-weight: 600; color: #1f2937; display: flex; align-items: center; }
}

.message-stream {
  max-height: 600px;
  overflow-y: auto;
  padding: 4px;

  .message-item {
    padding: 12px;
    margin-bottom: 8px;
    border-radius: 8px;
    background: #f9fafb;
    border-left: 4px solid #d1d5db;

    &.msg-in {
      border-left-color: #10b981;
      background: #f0fdf4;
    }

    &.msg-rule {
      border-left-color: #f59e0b;
      background: #fffbeb;
    }

    .msg-header {
      display: flex;
      align-items: center;
      gap: 8px;
      margin-bottom: 8px;

      .msg-time {
        font-size: 12px;
        color: #9ca3af;
        font-family: monospace;
      }
    }

    .msg-body {
      .msg-field {
        font-size: 13px;
        color: #374151;
        margin-bottom: 4px;

        code {
          background: #e5e7eb;
          padding: 2px 6px;
          border-radius: 4px;
          font-size: 12px;
          word-break: break-all;
        }
      }
    }
  }
}

.topic-list {
  max-height: 600px;
  overflow-y: auto;

  .topic-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 8px 12px;
    border-radius: 6px;
    cursor: pointer;
    transition: background 0.2s;

    &:hover {
      background: #f3f4f6;
    }

    .topic-name {
      font-size: 12px;
      color: #4338ca;
      word-break: break-all;
      margin-right: 8px;
    }
  }
}
</style>
