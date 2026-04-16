<template>
  <div class="page-container">
    <div class="page-header">
      <h2>设备分组</h2>
      <el-button type="primary" @click="showCreateDialog = true">
        <el-icon><Plus /></el-icon>
        创建分组
      </el-button>
    </div>

    <el-row :gutter="20">
      <!-- Group Tree -->
      <el-col :xs="24" :md="8">
        <div class="card-box">
          <h3 class="section-title">分组列表</h3>
          <el-tree
            :data="groupTree"
            :props="treeProps"
            node-key="id"
            highlight-current
            default-expand-all
            @node-click="handleNodeClick"
          >
            <template #default="{ node, data }">
              <div class="tree-node">
                <span>{{ node.label }}</span>
                <span class="tree-node-actions">
                  <el-tag size="small" type="info">{{ data.device_count || 0 }}</el-tag>
                  <el-button text size="small" type="primary" @click.stop="editGroup(data)">
                    <el-icon><Edit /></el-icon>
                  </el-button>
                  <el-popconfirm title="确定删除该分组吗?" @confirm="handleDeleteGroup(data.id)">
                    <template #reference>
                      <el-button text size="small" type="danger" @click.stop>
                        <el-icon><Delete /></el-icon>
                      </el-button>
                    </template>
                  </el-popconfirm>
                </span>
              </div>
            </template>
          </el-tree>
          <el-empty v-if="groupTree.length === 0" description="暂无分组" :image-size="80" />
        </div>
      </el-col>

      <!-- Device List for Selected Group -->
      <el-col :xs="24" :md="16">
        <div class="card-box">
          <div class="card-header">
            <h3>{{ selectedGroup ? `"${selectedGroup.name}" 的设备` : '请选择一个分组' }}</h3>
            <el-button
              v-if="selectedGroup"
              type="primary"
              size="small"
              @click="showAddDeviceDialog = true"
            >
              添加设备
            </el-button>
          </div>

          <el-table v-if="selectedGroup" :data="groupDevices" stripe style="width: 100%">
            <el-table-column type="index" label="#" width="50" />
            <el-table-column prop="name" label="设备名称" min-width="160" />
            <el-table-column prop="device_key" label="设备Key" min-width="160">
              <template #default="{ row }">
                <el-text family="monospace" size="small">{{ row.device_key }}</el-text>
              </template>
            </el-table-column>
            <el-table-column prop="status" label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="row.status === 'online' ? 'success' : row.status === 'offline' ? 'danger' : 'info'" size="small" effect="dark">
                  {{ row.status === 'online' ? '在线' : row.status === 'offline' ? '离线' : '未激活' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="100">
              <template #default="{ row }">
                <el-popconfirm title="确定从分组中移除该设备吗?" @confirm="removeDeviceFromGroup(row.id)">
                  <template #reference>
                    <el-button text type="danger" size="small">移除</el-button>
                  </template>
                </el-popconfirm>
              </template>
            </el-table-column>
          </el-table>

          <el-empty v-else description="请在左侧选择一个分组" :image-size="100" />
        </div>
      </el-col>
    </el-row>

    <!-- Create/Edit Group Dialog -->
    <el-dialog
      v-model="showCreateDialog"
      :title="editingGroup ? '编辑分组' : '创建分组'"
      width="480px"
      @close="resetGroupForm"
    >
      <el-form :model="groupForm" label-width="80px">
        <el-form-item label="名称">
          <el-input v-model="groupForm.name" placeholder="请输入分组名称" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="groupForm.description" type="textarea" :rows="3" placeholder="分组描述（可选）" />
        </el-form-item>
        <el-form-item label="父分组">
          <el-select v-model="groupForm.parent_id" placeholder="无父分组" clearable style="width: 100%">
            <el-option
              v-for="g in flatGroups"
              :key="g.id"
              :label="g.name"
              :value="g.id"
              :disabled="!!editingGroup && editingGroup.id === g.id"
            />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreateDialog = false">取消</el-button>
        <el-button type="primary" :loading="savingGroup" @click="handleSaveGroup">保存</el-button>
      </template>
    </el-dialog>

    <!-- Add Device Dialog -->
    <el-dialog v-model="showAddDeviceDialog" title="添加设备到分组" width="600px">
      <el-table
        ref="addDeviceTableRef"
        :data="availableDevices"
        stripe
        @selection-change="handleDeviceSelection"
      >
        <el-table-column type="selection" width="55" />
        <el-table-column prop="name" label="设备名称" min-width="150" />
        <el-table-column prop="device_key" label="设备Key" min-width="150" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 'online' ? 'success' : row.status === 'offline' ? 'danger' : 'info'" size="small">
              {{ row.status === 'online' ? '在线' : row.status === 'offline' ? '离线' : '未激活' }}
            </el-tag>
          </template>
        </el-table-column>
      </el-table>
      <template #footer>
        <el-button @click="showAddDeviceDialog = false">取消</el-button>
        <el-button type="primary" :loading="addingDevices" @click="handleAddDevices">添加</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { getGroups, createGroup, updateGroup, deleteGroup, addDevicesToGroup, removeDevicesFromGroup } from '@/api/group'
import { getDevices } from '@/api/device'
import type { Group, Device } from '@/types'

const groups = ref<Group[]>([])
const groupTree = ref<Group[]>([])
const flatGroups = ref<Group[]>([])
const selectedGroup = ref<Group | null>(null)
const groupDevices = ref<Device[]>([])
const allDevices = ref<Device[]>([])
const savingGroup = ref(false)
const addingDevices = ref(false)
const showCreateDialog = ref(false)
const showAddDeviceDialog = ref(false)
const editingGroup = ref<Group | null>(null)
const selectedDevices = ref<Device[]>([])

const groupForm = reactive({
  name: '',
  description: '',
  parent_id: undefined as number | undefined,
})

const treeProps = {
  children: 'children',
  label: 'name',
}

function buildTree(list: Group[], parentId?: number): Group[] {
  return list
    .filter(g => g.parent_id === parentId)
    .map(g => ({ ...g, children: buildTree(list, g.id) }))
}

async function fetchGroups() {
  const res = await getGroups()
  groups.value = res || []
  flatGroups.value = res || []
  groupTree.value = buildTree(res || [])
}

async function fetchAllDevices() {
  const res = await getDevices({ page: 1, page_size: 200 })
  allDevices.value = res.data || []
}

async function handleNodeClick(data: Group) {
  selectedGroup.value = data
  const res = await getDevices({ group_id: data.id, page: 1, page_size: 100 })
  groupDevices.value = res.data || []
}

function editGroup(data: Group) {
  editingGroup.value = data
  groupForm.name = data.name
  groupForm.description = data.description || ''
  groupForm.parent_id = data.parent_id
  showCreateDialog.value = true
}

function resetGroupForm() {
  editingGroup.value = null
  groupForm.name = ''
  groupForm.description = ''
  groupForm.parent_id = undefined
}

async function handleSaveGroup() {
  if (!groupForm.name) {
    ElMessage.warning('请输入分组名称')
    return
  }
  savingGroup.value = true
  try {
    if (editingGroup.value) {
      await updateGroup(editingGroup.value.id, {
        name: groupForm.name,
        description: groupForm.description,
        parent_id: groupForm.parent_id,
      })
      ElMessage.success('分组已更新')
    } else {
      await createGroup({
        name: groupForm.name,
        description: groupForm.description,
        parent_id: groupForm.parent_id,
      })
      ElMessage.success('分组已创建')
    }
    showCreateDialog.value = false
    resetGroupForm()
    await fetchGroups()
  } finally {
    savingGroup.value = false
  }
}

async function handleDeleteGroup(id: number) {
  await deleteGroup(id)
  ElMessage.success('分组已删除')
  if (selectedGroup.value?.id === id) {
    selectedGroup.value = null
    groupDevices.value = []
  }
  await fetchGroups()
}

const availableDevices = computed(() => {
  const ids = new Set(groupDevices.value.map(d => d.id))
  return allDevices.value.filter(d => !ids.has(d.id))
})

function handleDeviceSelection(selection: Device[]) {
  selectedDevices.value = selection
}

async function handleAddDevices() {
  if (!selectedGroup.value || selectedDevices.value.length === 0) {
    ElMessage.warning('请选择要添加的设备')
    return
  }
  addingDevices.value = true
  try {
    await addDevicesToGroup(
      selectedGroup.value.id,
      selectedDevices.value.map(d => d.id),
    )
    ElMessage.success('设备已添加到分组')
    showAddDeviceDialog.value = false
    await handleNodeClick(selectedGroup.value)
  } finally {
    addingDevices.value = false
  }
}

async function removeDeviceFromGroup(deviceId: number) {
  if (!selectedGroup.value) return
  await removeDevicesFromGroup(selectedGroup.value.id, [deviceId])
  ElMessage.success('设备已从分组中移除')
  await handleNodeClick(selectedGroup.value)
}

onMounted(async () => {
  await Promise.all([fetchGroups(), fetchAllDevices()])
})
</script>

<style lang="scss" scoped>
.section-title {
  font-size: 16px;
  font-weight: 600;
  color: #1f2937;
  margin-bottom: 16px;
}

.tree-node {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  padding-right: 8px;

  .tree-node-actions {
    display: flex;
    align-items: center;
    gap: 2px;
  }
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
  h3 { font-size: 16px; font-weight: 600; color: #1f2937; }
}
</style>
