<template>
  <div class="page-container">
    <div class="page-header">
      <h2>
        <el-button text @click="$router.push('/rules')">
          <el-icon><ArrowLeft /></el-icon>
        </el-button>
        {{ isEdit ? '编辑规则' : '创建规则' }}
      </h2>
    </div>

    <div class="card-box form-wrapper">
      <el-form ref="formRef" :model="form" :rules="formRules" label-width="100px" style="max-width: 800px;">
        <el-form-item label="规则名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入规则名称" />
        </el-form-item>

        <el-form-item label="描述">
          <el-input v-model="form.description" type="textarea" :rows="2" placeholder="规则描述（可选）" />
        </el-form-item>

        <el-form-item label="Topic匹配" prop="topic_pattern">
          <el-input v-model="form.topic_pattern" placeholder="如: telemetry/+/temperature" />
        </el-form-item>

        <el-divider content-position="left">触发条件</el-divider>

        <el-row :gutter="8" align="middle">
          <el-col :span="6">
            <el-input v-model="condition.field" placeholder="字段名 (如 temperature)" size="small" />
          </el-col>
          <el-col :span="5">
            <el-select v-model="condition.operator" placeholder="操作符" size="small" style="width: 100%">
              <el-option label="等于" value="eq" />
              <el-option label="不等于" value="neq" />
              <el-option label="大于" value="gt" />
              <el-option label="大于等于" value="gte" />
              <el-option label="小于" value="lt" />
              <el-option label="小于等于" value="lte" />
              <el-option label="包含" value="contains" />
            </el-select>
          </el-col>
          <el-col :span="8">
            <el-input v-model="condition.value" placeholder="阈值" size="small" />
          </el-col>
        </el-row>

        <el-divider content-position="left">执行动作</el-divider>

        <el-row :gutter="12" align="middle">
          <el-col :span="6">
            <el-select v-model="form.action_type" placeholder="动作类型" size="small" style="width: 100%">
              <el-option label="告警通知" value="alert" />
              <el-option label="发布消息" value="publish" />
              <el-option label="HTTP转发" value="forward" />
            </el-select>
          </el-col>
          <el-col :span="18">
            <template v-if="form.action_type === 'alert'">
              <el-input v-model="actionConfig.message" placeholder="告警消息内容" size="small" />
            </template>
            <template v-else-if="form.action_type === 'publish'">
              <div style="display: flex; gap: 8px;">
                <el-input v-model="actionConfig.topic" placeholder="目标Topic" size="small" />
                <el-input v-model="actionConfig.payload" placeholder="消息内容" size="small" />
              </div>
            </template>
            <template v-else-if="form.action_type === 'forward'">
              <div style="display: flex; gap: 8px;">
                <el-select v-model="actionConfig.method" size="small" style="width: 100px;">
                  <el-option label="POST" value="POST" />
                  <el-option label="GET" value="GET" />
                </el-select>
                <el-input v-model="actionConfig.url" placeholder="Webhook URL" size="small" />
              </div>
            </template>
          </el-col>
        </el-row>

        <el-form-item label="冷却时间" style="margin-top: 16px;">
          <el-input-number v-model="form.cooldown_secs" :min="0" :max="3600" :step="10" />
          <span style="margin-left: 8px; color: #999;">秒</span>
        </el-form-item>

        <el-form-item label="启用">
          <el-switch v-model="form.enabled" active-color="#6366f1" />
        </el-form-item>

        <el-divider />
        <el-form-item>
          <el-button type="primary" :loading="saving" @click="handleSubmit">
            {{ isEdit ? '保存修改' : '创建规则' }}
          </el-button>
          <el-button @click="$router.push('/rules')">取消</el-button>
        </el-form-item>
      </el-form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { getRule, createRule, updateRule } from '@/api/rule'

const route = useRoute()
const router = useRouter()

const ruleId = computed(() => route.params.id as string)
const isEdit = computed(() => !!ruleId.value && route.name === 'RuleEdit')

const formRef = ref<FormInstance>()
const saving = ref(false)

const form = reactive({
  name: '',
  description: '',
  topic_pattern: '',
  action_type: 'alert',
  cooldown_secs: 60,
  enabled: true,
})

const condition = reactive({
  field: '',
  operator: 'gt',
  value: '',
})

const actionConfig = reactive<Record<string, string>>({
  message: '',
  topic: '',
  payload: '',
  method: 'POST',
  url: '',
})

const formRules: FormRules = {
  name: [{ required: true, message: '请输入规则名称', trigger: 'blur' }],
  topic_pattern: [{ required: true, message: '请输入Topic匹配模式', trigger: 'blur' }],
}

async function loadRule() {
  if (!isEdit.value) return
  try {
    const data = await getRule(ruleId.value)
    form.name = data.name
    form.description = data.description || ''
    form.topic_pattern = data.topic_pattern || ''
    form.action_type = data.action_type || 'alert'
    form.cooldown_secs = data.cooldown_secs || 60
    form.enabled = data.enabled

    // Parse condition JSON
    if (data.condition) {
      const cond = typeof data.condition === 'string' ? JSON.parse(data.condition) : data.condition
      condition.field = cond.field || ''
      condition.operator = cond.operator || 'gt'
      condition.value = String(cond.value ?? '')
    }

    // Parse action_config JSON
    if (data.action_config) {
      const ac = typeof data.action_config === 'string' ? JSON.parse(data.action_config) : data.action_config
      Object.assign(actionConfig, ac)
    }
  } catch (e) {
    console.error('Failed to load rule', e)
  }
}

async function handleSubmit() {
  const formEl = formRef.value
  if (!formEl) return
  await formEl.validate(async (valid) => {
    if (!valid) return
    if (!condition.field || !condition.value) {
      ElMessage.warning('请填写完整的触发条件')
      return
    }

    saving.value = true
    try {
      const payload = {
        name: form.name,
        description: form.description,
        topic_pattern: form.topic_pattern,
        condition: JSON.stringify({
          field: condition.field,
          operator: condition.operator,
          value: isNaN(Number(condition.value)) ? condition.value : Number(condition.value),
        }),
        action_type: form.action_type,
        action_config: JSON.stringify(buildActionConfig()),
        cooldown_secs: form.cooldown_secs,
        enabled: form.enabled,
      }
      if (isEdit.value) {
        await updateRule(ruleId.value, payload)
        ElMessage.success('规则已更新')
      } else {
        await createRule(payload)
        ElMessage.success('规则已创建')
      }
      router.push('/rules')
    } finally {
      saving.value = false
    }
  })
}

function buildActionConfig(): Record<string, string> {
  if (form.action_type === 'alert') {
    return { message: actionConfig.message || '触发告警' }
  } else if (form.action_type === 'publish') {
    return { topic: actionConfig.topic, payload: actionConfig.payload }
  } else {
    return { method: actionConfig.method, url: actionConfig.url }
  }
}

onMounted(() => {
  loadRule()
})
</script>

<style lang="scss" scoped>
.form-wrapper {
  max-width: 900px;
}
</style>
