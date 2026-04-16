<template>
  <div class="page-container">
    <div class="page-header">
      <h2>用户管理</h2>
      <el-button type="primary" @click="showCreateDialog = true">
        <el-icon><Plus /></el-icon>
        添加用户
      </el-button>
    </div>

    <div class="card-box">
      <el-table v-loading="loading" :data="users" stripe style="width: 100%">
        <el-table-column type="index" label="#" width="50" />
        <el-table-column prop="username" label="用户名" min-width="140" />
        <el-table-column prop="email" label="邮箱" min-width="200" />
        <el-table-column prop="role" label="角色" width="120">
          <template #default="{ row }">
            <el-tag :type="row.role === 'admin' ? 'danger' : row.role === 'operator' ? 'warning' : 'primary'" size="small" effect="dark">
              {{ row.role === 'admin' ? '管理员' : row.role === 'operator' ? '操作员' : '观察者' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 'active' ? 'success' : 'info'" size="small">
              {{ row.status === 'active' ? '正常' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="170">
          <template #default="{ row }">
            {{ dayjs(row.created_at).format('YYYY-MM-DD HH:mm:ss') }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="220" fixed="right">
          <template #default="{ row }">
            <el-button text type="primary" size="small" @click="editUser(row)">编辑</el-button>
            <el-button text type="warning" size="small" @click="resetPassword(row)">重置密码</el-button>
            <el-popconfirm title="确定删除该用户吗?" @confirm="handleDelete(row.id)">
              <template #reference>
                <el-button text type="danger" size="small" :disabled="row.role === 'admin'">删除</el-button>
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

    <!-- Create/Edit Dialog -->
    <el-dialog
      v-model="showCreateDialog"
      :title="editingUser ? '编辑用户' : '创建用户'"
      width="480px"
      @close="resetUserForm"
    >
      <el-form :model="userForm" :rules="formRules" ref="formRef" label-width="80px">
        <el-form-item label="用户名" prop="username">
          <el-input v-model="userForm.username" :disabled="!!editingUser" placeholder="请输入用户名" />
        </el-form-item>
        <el-form-item v-if="!editingUser" label="密码" prop="password">
          <el-input v-model="userForm.password" type="password" show-password placeholder="请输入密码" />
        </el-form-item>
        <el-form-item label="邮箱" prop="email">
          <el-input v-model="userForm.email" placeholder="请输入邮箱" />
        </el-form-item>
        <el-form-item label="角色" prop="role">
          <el-select v-model="userForm.role" style="width: 100%">
            <el-option label="管理员" value="admin" />
            <el-option label="操作员" value="operator" />
            <el-option label="观察者" value="viewer" />
            <el-option label="普通用户" value="user" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="userForm.status" style="width: 100%">
            <el-option label="正常" value="active" />
            <el-option label="禁用" value="disabled" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreateDialog = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleSave">保存</el-button>
      </template>
    </el-dialog>

    <!-- Reset Password Dialog -->
    <el-dialog v-model="showResetDialog" title="重置密码" width="400px">
      <el-form :model="resetFormData" label-width="80px">
        <el-form-item label="用户">
          <el-input :model-value="resetFormData.username" disabled />
        </el-form-item>
        <el-form-item label="新密码">
          <el-input v-model="resetFormData.new_password" type="password" show-password placeholder="请输入新密码" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showResetDialog = false">取消</el-button>
        <el-button type="primary" :loading="resetting" @click="handleResetPassword">确认重置</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import dayjs from 'dayjs'
import client from '@/api/client'
import type { User } from '@/types'

const loading = ref(false)
const saving = ref(false)
const resetting = ref(false)
const users = ref<User[]>([])
const total = ref(0)
const currentPage = ref(1)
const pageSize = ref(20)

const showCreateDialog = ref(false)
const showResetDialog = ref(false)
const editingUser = ref<User | null>(null)
const formRef = ref<FormInstance>()

const userForm = reactive({
  username: '',
  password: '',
  email: '',
  role: 'viewer',
  status: 'active',
})

const resetFormData = reactive({
  userId: 0,
  username: '',
  new_password: '',
})

const formRules: FormRules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }, { min: 6, message: '密码至少6位', trigger: 'blur' }],
  email: [{ type: 'email', message: '请输入正确的邮箱', trigger: 'blur' }],
  role: [{ required: true, message: '请选择角色', trigger: 'change' }],
}

async function fetchData() {
  loading.value = true
  try {
    const res: any = await client.get('/users', {
      params: { page: currentPage.value, page_size: pageSize.value },
    })
    users.value = res.data || []
    total.value = res.total || 0
  } finally {
    loading.value = false
  }
}

function editUser(user: User) {
  editingUser.value = user
  userForm.username = user.username
  userForm.email = user.email
  userForm.role = user.role
  userForm.status = user.status
  showCreateDialog.value = true
}

function resetUserForm() {
  editingUser.value = null
  userForm.username = ''
  userForm.password = ''
  userForm.email = ''
  userForm.role = 'user'
  userForm.status = 'active'
}

async function handleSave() {
  const form = formRef.value
  if (!form) return
  await form.validate(async (valid) => {
    if (!valid) return
    saving.value = true
    try {
      if (editingUser.value) {
        await client.put(`/users/${editingUser.value.id}`, {
          email: userForm.email,
          role: userForm.role,
          status: userForm.status,
        })
        ElMessage.success('用户已更新')
      } else {
        await client.post('/users', userForm)
        ElMessage.success('用户已创建')
      }
      showCreateDialog.value = false
      resetUserForm()
      fetchData()
    } finally {
      saving.value = false
    }
  })
}

function resetPassword(user: User) {
  resetFormData.userId = user.id
  resetFormData.username = user.username
  resetFormData.new_password = ''
  showResetDialog.value = true
}

async function handleResetPassword() {
  if (!resetFormData.new_password) {
    ElMessage.warning('请输入新密码')
    return
  }
  resetting.value = true
  try {
    await client.put(`/users/${resetFormData.userId}`, {
      password: resetFormData.new_password,
    })
    ElMessage.success('密码已重置')
    showResetDialog.value = false
  } finally {
    resetting.value = false
  }
}

async function handleDelete(id: number) {
  await client.delete(`/users/${id}`)
  ElMessage.success('用户已删除')
  fetchData()
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
