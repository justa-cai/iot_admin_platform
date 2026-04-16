<template>
  <div class="page-container">
    <div class="page-header">
      <h2>
        <el-button text @click="$router.push('/devices')">
          <el-icon><ArrowLeft /></el-icon>
        </el-button>
        {{ isEdit ? '编辑设备' : '创建设备' }}
      </h2>
    </div>

    <div class="card-box form-wrapper">
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="120px"
        label-position="right"
        style="max-width: 640px;"
      >
        <el-form-item label="设备名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入设备名称" />
        </el-form-item>

        <el-form-item v-if="!isEdit" label="设备Key" prop="device_key">
          <el-input v-model="form.device_key" placeholder="请输入设备唯一标识">
            <template #append>
              <el-button @click="generateKey">自动生成</el-button>
            </template>
          </el-input>
        </el-form-item>

        <el-form-item label="设备分组" prop="group_id">
          <el-select v-model="form.group_id" placeholder="请选择分组" clearable style="width: 100%">
            <el-option
              v-for="g in groups"
              :key="g.id"
              :label="g.name"
              :value="g.id"
            />
          </el-select>
        </el-form-item>

        <el-form-item label="标签">
          <el-select
            v-model="form.tag_ids"
            multiple
            placeholder="请选择标签"
            style="width: 100%"
          >
            <el-option
              v-for="t in tags"
              :key="t.id"
              :label="t.name"
              :value="t.id"
            />
          </el-select>
        </el-form-item>

        <el-form-item label="元数据">
          <div style="width: 100%;">
            <div v-for="(item, index) in metadataEntries" :key="index" style="display: flex; gap: 8px; margin-bottom: 8px;">
              <el-input v-model="item.key" placeholder="Key" style="flex: 1;" />
              <el-input v-model="item.value" placeholder="Value" style="flex: 1;" />
              <el-button text type="danger" @click="metadataEntries.splice(index, 1)">
                <el-icon><Delete /></el-icon>
              </el-button>
            </div>
            <el-button text type="primary" @click="metadataEntries.push({ key: '', value: '' })">
              <el-icon><Plus /></el-icon> 添加属性
            </el-button>
          </div>
        </el-form-item>

        <el-form-item>
          <el-button type="primary" :loading="saving" @click="handleSubmit">
            {{ isEdit ? '保存修改' : '创建设备' }}
          </el-button>
          <el-button @click="$router.push('/devices')">取消</el-button>
        </el-form-item>
      </el-form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { getDevice, createDevice, updateDevice } from '@/api/device'
import { getGroups } from '@/api/group'
import { getTags } from '@/api/tag'
import type { Group, Tag } from '@/types'

const route = useRoute()
const router = useRouter()

const deviceId = computed(() => route.params.id as string)
const isEdit = computed(() => !!deviceId.value && route.name === 'DeviceEdit')

const formRef = ref<FormInstance>()
const saving = ref(false)
const groups = ref<Group[]>([])
const tags = ref<Tag[]>([])

const metadataEntries = ref<{ key: string; value: string }[]>([])

const form = reactive({
  name: '',
  device_key: '',
  group_id: undefined as number | undefined,
  tag_ids: [] as number[],
})

const rules: FormRules = {
  name: [{ required: true, message: '请输入设备名称', trigger: 'blur' }],
  device_key: [{ required: true, message: '请输入设备Key', trigger: 'blur' }],
}

function generateKey() {
  const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789'
  let result = 'DEV-'
  for (let i = 0; i < 16; i++) {
    result += chars.charAt(Math.floor(Math.random() * chars.length))
  }
  form.device_key = result
}

async function loadDevice() {
  if (!isEdit.value) return
  const data = await getDevice(Number(deviceId.value))
  form.name = data.name
  form.device_key = data.device_key
  form.group_id = data.group_id
  form.tag_ids = (data.tags || []).map(t => t.id)
  if (data.metadata && typeof data.metadata === 'object') {
    metadataEntries.value = Object.entries(data.metadata).map(([key, value]) => ({
      key,
      value: String(value),
    }))
  }
}

function buildMetadata(): Record<string, unknown> {
  const meta: Record<string, unknown> = {}
  for (const entry of metadataEntries.value) {
    if (entry.key.trim()) {
      meta[entry.key.trim()] = entry.value
    }
  }
  return meta
}

async function handleSubmit() {
  const formEl = formRef.value
  if (!formEl) return
  await formEl.validate(async (valid) => {
    if (!valid) return
    saving.value = true
    try {
      const payload = {
        name: form.name,
        device_key: form.device_key,
        metadata: JSON.stringify(buildMetadata()),
        group_id: form.group_id,
        tag_ids: form.tag_ids,
      }
      if (isEdit.value) {
        await updateDevice(Number(deviceId.value), payload)
        ElMessage.success('设备已更新')
      } else {
        await createDevice(payload)
        ElMessage.success('设备已创建')
      }
      router.push('/devices')
    } finally {
      saving.value = false
    }
  })
}

onMounted(async () => {
  const [groupRes, tagRes] = await Promise.all([getGroups(), getTags()])
  groups.value = groupRes || []
  tags.value = tagRes || []
  loadDevice()
})
</script>

<style lang="scss" scoped>
.form-wrapper {
  max-width: 800px;
}
</style>
