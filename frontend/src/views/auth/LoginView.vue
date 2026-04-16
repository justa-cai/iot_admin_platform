<template>
  <el-form
    ref="formRef"
    :model="loginForm"
    :rules="rules"
    label-position="top"
    size="large"
    @submit.prevent="handleLogin"
  >
    <el-form-item label="用户名" prop="username">
      <el-input
        v-model="loginForm.username"
        placeholder="请输入用户名"
        :prefix-icon="User"
        autocomplete="username"
      />
    </el-form-item>

    <el-form-item label="密码" prop="password">
      <el-input
        v-model="loginForm.password"
        type="password"
        placeholder="请输入密码"
        :prefix-icon="Lock"
        show-password
        autocomplete="current-password"
        @keyup.enter="handleLogin"
      />
    </el-form-item>

    <el-form-item>
      <el-button
        type="primary"
        :loading="loading"
        style="width: 100%; height: 44px; font-size: 16px; background-color: #4338ca; border-color: #4338ca"
        @click="handleLogin"
      >
        {{ loading ? '登录中...' : '登 录' }}
      </el-button>
    </el-form-item>
  </el-form>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { User, Lock } from '@element-plus/icons-vue'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()

const formRef = ref<FormInstance>()
const loading = ref(false)

const loginForm = reactive({
  username: '',
  password: '',
})

const rules: FormRules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, message: '密码至少6个字符', trigger: 'blur' },
  ],
}

async function handleLogin() {
  const form = formRef.value
  if (!form) return
  await form.validate(async (valid) => {
    if (!valid) return
    loading.value = true
    try {
      await authStore.login(loginForm.username, loginForm.password)
      ElMessage.success('登录成功')
      const redirect = (route.query.redirect as string) || '/dashboard'
      router.push(redirect)
    } catch {
      // error handled in interceptor
    } finally {
      loading.value = false
    }
  })
}
</script>
