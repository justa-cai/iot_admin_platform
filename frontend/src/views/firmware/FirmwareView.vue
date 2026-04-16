<template>
  <div class="page-container">
    <div class="page-header">
      <h2>固件管理</h2>
      <el-button type="primary" @click="showUploadDialog = true">
        <el-icon><Upload /></el-icon>
        上传固件
      </el-button>
    </div>

    <el-tabs v-model="activeTab" type="border-card">
      <!-- Firmware List -->
      <el-tab-pane label="固件列表" name="firmware">
        <el-table v-loading="firmwareLoading" :data="firmwareList" stripe style="width: 100%">
          <el-table-column prop="id" label="#" width="60" />
          <el-table-column prop="name" label="固件名称" min-width="160" />
          <el-table-column prop="version" label="版本" width="120">
            <template #default="{ row }">
              <el-tag size="small" effect="dark">v{{ row.version }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="file_name" label="文件名" min-width="180" />
          <el-table-column prop="file_size" label="文件大小" width="120">
            <template #default="{ row }">
              {{ formatFileSize(row.file_size) }}
            </template>
          </el-table-column>
          <el-table-column prop="description" label="描述" min-width="200">
            <template #default="{ row }">
              {{ row.description || '-' }}
            </template>
          </el-table-column>
          <el-table-column prop="created_at" label="上传时间" width="170">
            <template #default="{ row }">
              {{ dayjs(row.created_at).format('YYYY-MM-DD HH:mm') }}
            </template>
          </el-table-column>
          <el-table-column label="操作" width="200" fixed="right">
            <template #default="{ row }">
              <el-button text type="primary" size="small" @click="handleDownload(row)">下载</el-button>
              <el-button text type="warning" size="small" @click="openOTADialog(row)">OTA升级</el-button>
              <el-popconfirm title="确定删除该固件吗?" @confirm="handleDeleteFirmware(row.id)">
                <template #reference>
                  <el-button text type="danger" size="small">删除</el-button>
                </template>
              </el-popconfirm>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>

      <!-- OTA Tasks -->
      <el-tab-pane label="OTA任务" name="ota">
        <div style="margin-bottom: 16px; display: flex; justify-content: flex-end;">
          <el-button type="primary" @click="showBatchOTADialog = true">
            <el-icon><Position /></el-icon>
            批量升级
          </el-button>
        </div>
        <el-table v-loading="otaLoading" :data="otaList" stripe style="width: 100%">
          <el-table-column prop="id" label="#" width="60" />
          <el-table-column prop="device_name" label="设备" min-width="140" />
          <el-table-column prop="firmware_name" label="固件" min-width="140" />
          <el-table-column prop="firmware_version" label="版本" width="100">
            <template #default="{ row }">
              <el-tag size="small">v{{ row.firmware_version }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="status" label="状态" width="120">
            <template #default="{ row }">
              <el-tag :type="otaStatusType(row.status)" size="small" effect="dark">
                {{ otaStatusLabel(row.status) }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="progress" label="进度" width="160">
            <template #default="{ row }">
              <el-progress
                :percentage="row.progress || 0"
                :status="row.status === 'success' ? 'success' : row.status === 'failed' ? 'exception' : undefined"
                :stroke-width="10"
              />
            </template>
          </el-table-column>
          <el-table-column prop="message" label="消息" min-width="160">
            <template #default="{ row }">
              {{ row.message || '-' }}
            </template>
          </el-table-column>
          <el-table-column prop="created_at" label="创建时间" width="170">
            <template #default="{ row }">
              {{ dayjs(row.created_at).format('YYYY-MM-DD HH:mm') }}
            </template>
          </el-table-column>
          <el-table-column label="操作" width="100" fixed="right">
            <template #default="{ row }">
              <el-popconfirm title="确定删除该OTA任务?" @confirm="handleDeleteOTA(row.id)">
                <template #reference>
                  <el-button text type="danger" size="small">删除</el-button>
                </template>
              </el-popconfirm>
            </template>
          </el-table-column>
        </el-table>

        <div class="pagination-wrapper">
          <el-pagination
            v-model:current-page="otaPage"
            v-model:page-size="otaPageSize"
            :total="otaTotal"
            :page-sizes="[10, 20, 50]"
            layout="total, sizes, prev, pager, next"
            @size-change="fetchOTAList"
            @current-change="fetchOTAList"
          />
        </div>
      </el-tab-pane>
    </el-tabs>

    <!-- Upload Dialog -->
    <el-dialog v-model="showUploadDialog" title="上传固件" width="500px">
      <el-form :model="uploadForm" label-width="80px">
        <el-form-item label="固件名称">
          <el-input v-model="uploadForm.name" placeholder="请输入固件名称" />
        </el-form-item>
        <el-form-item label="版本号">
          <el-input v-model="uploadForm.version" placeholder="例如: 1.0.0" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="uploadForm.description" type="textarea" :rows="3" placeholder="固件描述（可选）" />
        </el-form-item>
        <el-form-item label="固件文件">
          <el-upload
            ref="uploadRef"
            :auto-upload="false"
            :limit="1"
            :on-change="handleFileChange"
            :on-remove="() => selectedFile = null"
            accept=".bin,.hex,.fw,.zip"
          >
            <template #trigger>
              <el-button type="primary" size="small">选择文件</el-button>
            </template>
            <template #tip>
              <div class="el-upload__tip">支持 .bin, .hex, .fw, .zip 文件</div>
            </template>
          </el-upload>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showUploadDialog = false">取消</el-button>
        <el-button type="primary" :loading="uploading" @click="handleUpload">上传</el-button>
      </template>
    </el-dialog>

    <!-- OTA Dialog -->
    <el-dialog v-model="showOTADialog" title="创建OTA升级任务" width="500px">
      <el-form :model="otaForm" label-width="80px">
        <el-form-item label="固件">
          <el-tag effect="dark">{{ selectedFirmware?.name }} v{{ selectedFirmware?.version }}</el-tag>
        </el-form-item>
        <el-form-item label="设备">
          <el-select v-model="otaForm.device_id" placeholder="请选择设备" filterable style="width: 100%">
            <el-option
              v-for="d in devices"
              :key="d.id"
              :label="d.name"
              :value="d.id"
            />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showOTADialog = false">取消</el-button>
        <el-button type="primary" :loading="creatingOTA" @click="handleCreateOTA">创建任务</el-button>
      </template>
    </el-dialog>

    <!-- Batch OTA Dialog -->
    <el-dialog v-model="showBatchOTADialog" title="批量OTA升级" width="600px">
      <el-form :model="batchOTAForm" label-width="80px">
        <el-form-item label="固件">
          <el-select v-model="batchOTAForm.firmware_id" placeholder="选择固件" style="width: 100%">
            <el-option
              v-for="f in firmwareList"
              :key="f.id"
              :label="`${f.name} v${f.version}`"
              :value="f.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="设备">
          <el-select
            v-model="batchOTAForm.device_ids"
            multiple
            filterable
            placeholder="选择设备"
            style="width: 100%"
          >
            <el-option
              v-for="d in devices"
              :key="d.id"
              :label="d.name"
              :value="d.id"
            />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showBatchOTADialog = false">取消</el-button>
        <el-button type="primary" :loading="creatingBatchOTA" @click="handleBatchOTA">批量创建</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import dayjs from 'dayjs'
import {
  getFirmwareList,
  uploadFirmware,
  deleteFirmware,
  downloadFirmware,
  getOTAList,
  createOTAUpgrade,
  createOTABatchUpgrade,
  deleteOTA,
} from '@/api/firmware'
import { getDevices } from '@/api/device'
import type { Firmware, OTAUpgrade, Device } from '@/types'
import type { UploadFile } from 'element-plus'

const activeTab = ref('firmware')

// Firmware
const firmwareLoading = ref(false)
const firmwareList = ref<Firmware[]>([])
const showUploadDialog = ref(false)
const uploading = ref(false)
const selectedFile = ref<File | null>(null)
const uploadRef = ref()

const uploadForm = reactive({
  name: '',
  version: '',
  description: '',
})

// OTA
const otaLoading = ref(false)
const otaList = ref<OTAUpgrade[]>([])
const otaTotal = ref(0)
const otaPage = ref(1)
const otaPageSize = ref(20)

const showOTADialog = ref(false)
const showBatchOTADialog = ref(false)
const creatingOTA = ref(false)
const creatingBatchOTA = ref(false)
const selectedFirmware = ref<Firmware | null>(null)
const devices = ref<Device[]>([])

const otaForm = reactive({
  device_id: undefined as number | undefined,
})

const batchOTAForm = reactive({
  firmware_id: undefined as number | undefined,
  device_ids: [] as number[],
})

function formatFileSize(bytes: number): string {
  if (!bytes) return '-'
  const units = ['B', 'KB', 'MB', 'GB']
  let size = bytes
  let unitIdx = 0
  while (size >= 1024 && unitIdx < units.length - 1) {
    size /= 1024
    unitIdx++
  }
  return `${size.toFixed(1)} ${units[unitIdx]}`
}

function otaStatusType(status: string): 'success' | 'warning' | 'danger' | 'info' | 'primary' {
  const map: Record<string, 'success' | 'warning' | 'danger' | 'info' | 'primary'> = {
    pending: 'info',
    downloading: 'primary',
    installing: 'warning',
    success: 'success',
    failed: 'danger',
  }
  return map[status] || 'info'
}

function otaStatusLabel(status: string): string {
  const map: Record<string, string> = {
    pending: '等待中',
    downloading: '下载中',
    installing: '安装中',
    success: '成功',
    failed: '失败',
  }
  return map[status] || status
}

function handleFileChange(file: UploadFile) {
  selectedFile.value = file.raw || null
}

async function fetchFirmware() {
  firmwareLoading.value = true
  try {
    firmwareList.value = await getFirmwareList()
  } finally {
    firmwareLoading.value = false
  }
}

async function fetchOTAList() {
  otaLoading.value = true
  try {
    const res = await getOTAList({
      page: otaPage.value,
      page_size: otaPageSize.value,
    })
    otaList.value = res.data || []
    otaTotal.value = res.total || 0
  } finally {
    otaLoading.value = false
  }
}

async function handleUpload() {
  if (!uploadForm.name || !uploadForm.version) {
    ElMessage.warning('请填写固件名称和版本')
    return
  }
  if (!selectedFile.value) {
    ElMessage.warning('请选择固件文件')
    return
  }
  uploading.value = true
  try {
    const formData = new FormData()
    formData.append('file', selectedFile.value)
    formData.append('name', uploadForm.name)
    formData.append('version', uploadForm.version)
    formData.append('description', uploadForm.description)
    await uploadFirmware(formData)
    ElMessage.success('固件已上传')
    showUploadDialog.value = false
    uploadForm.name = ''
    uploadForm.version = ''
    uploadForm.description = ''
    selectedFile.value = null
    fetchFirmware()
  } finally {
    uploading.value = false
  }
}

function handleDownload(row: Firmware) {
  const url = downloadFirmware(row.id)
  window.open(url, '_blank')
}

async function handleDeleteFirmware(id: number) {
  await deleteFirmware(id)
  ElMessage.success('固件已删除')
  fetchFirmware()
}

function openOTADialog(firmware: Firmware) {
  selectedFirmware.value = firmware
  otaForm.device_id = undefined
  showOTADialog.value = true
}

async function handleCreateOTA() {
  if (!otaForm.device_id || !selectedFirmware.value) {
    ElMessage.warning('请选择设备')
    return
  }
  creatingOTA.value = true
  try {
    await createOTAUpgrade({
      device_id: otaForm.device_id,
      firmware_id: selectedFirmware.value.id,
    })
    ElMessage.success('OTA任务已创建')
    showOTADialog.value = false
    if (activeTab.value === 'ota') fetchOTAList()
  } finally {
    creatingOTA.value = false
  }
}

async function handleBatchOTA() {
  if (!batchOTAForm.firmware_id || batchOTAForm.device_ids.length === 0) {
    ElMessage.warning('请选择固件和设备')
    return
  }
  creatingBatchOTA.value = true
  try {
    await createOTABatchUpgrade({
      firmware_id: batchOTAForm.firmware_id,
      device_ids: batchOTAForm.device_ids,
    })
    ElMessage.success(`已创建 ${batchOTAForm.device_ids.length} 个OTA任务`)
    showBatchOTADialog.value = false
    batchOTAForm.device_ids = []
    batchOTAForm.firmware_id = undefined
    fetchOTAList()
  } finally {
    creatingBatchOTA.value = false
  }
}

async function handleDeleteOTA(id: number) {
  await deleteOTA(id)
  ElMessage.success('OTA任务已删除')
  fetchOTAList()
}

onMounted(async () => {
  const devRes = await getDevices({ page: 1, page_size: 500 })
  devices.value = devRes.data || []
  fetchFirmware()
  fetchOTAList()
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
