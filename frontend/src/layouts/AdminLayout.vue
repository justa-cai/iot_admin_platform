<template>
  <el-container class="admin-layout">
    <!-- Sidebar -->
    <el-aside :width="isCollapsed ? '64px' : '240px'" class="sidebar">
      <div class="sidebar-logo">
        <el-icon :size="28" color="#fff"><Monitor /></el-icon>
        <span v-show="!isCollapsed" class="logo-text">IoT 管理平台</span>
      </div>
      <el-menu
        :default-active="currentRoute"
        :collapse="isCollapsed"
        :collapse-transition="true"
        background-color="#1e1b4b"
        text-color="rgba(255,255,255,0.7)"
        active-text-color="#ffffff"
        router
        class="sidebar-menu"
      >
        <el-menu-item index="/dashboard">
          <el-icon><Odometer /></el-icon>
          <template #title>仪表盘</template>
        </el-menu-item>

        <el-sub-menu index="device-menu">
          <template #title>
            <el-icon><Monitor /></el-icon>
            <span>设备管理</span>
          </template>
          <el-menu-item index="/devices">设备列表</el-menu-item>
          <el-menu-item index="/groups">设备分组</el-menu-item>
        </el-sub-menu>

        <el-sub-menu index="message-menu">
          <template #title>
            <el-icon><ChatDotRound /></el-icon>
            <span>消息中心</span>
          </template>
          <el-menu-item index="/messages/console">消息控制台</el-menu-item>
          <el-menu-item index="/messages/history">消息历史</el-menu-item>
        </el-sub-menu>

        <el-menu-item index="/rules">
          <el-icon><SetUp /></el-icon>
          <template #title>规则引擎</template>
        </el-menu-item>

        <el-menu-item index="/telemetry">
          <el-icon><TrendCharts /></el-icon>
          <template #title>数据查询</template>
        </el-menu-item>

        <el-menu-item index="/firmware">
          <el-icon><Upload /></el-icon>
          <template #title>固件管理</template>
        </el-menu-item>

        <el-menu-item index="/system/users">
          <el-icon><User /></el-icon>
          <template #title>系统管理</template>
        </el-menu-item>
      </el-menu>
    </el-aside>

    <!-- Main Area -->
    <el-container class="main-container">
      <!-- Header -->
      <el-header class="header">
        <div class="header-left">
          <el-icon
            class="collapse-btn"
            :size="20"
            @click="isCollapsed = !isCollapsed"
          >
            <Fold v-if="!isCollapsed" />
            <Expand v-else />
          </el-icon>
          <el-breadcrumb separator="/">
            <el-breadcrumb-item :to="{ path: '/dashboard' }">首页</el-breadcrumb-item>
            <el-breadcrumb-item v-if="route.meta.title">{{ route.meta.title }}</el-breadcrumb-item>
          </el-breadcrumb>
        </div>
        <div class="header-right">
          <el-badge :is-dot="wsConnected" class="ws-badge">
            <el-tag :type="wsConnected ? 'success' : 'info'" size="small" effect="dark">
              {{ wsConnected ? '已连接' : '未连接' }}
            </el-tag>
          </el-badge>
          <el-dropdown @command="handleCommand">
            <span class="user-dropdown">
              <el-avatar :size="32" style="background-color: #4338ca">
                {{ username.charAt(0).toUpperCase() }}
              </el-avatar>
              <span class="username">{{ username }}</span>
              <el-icon><ArrowDown /></el-icon>
            </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="profile">个人设置</el-dropdown-item>
                <el-dropdown-item command="logout" divided>退出登录</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </el-header>

      <!-- Content -->
      <el-main class="main-content">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useWebSocket } from '@/composables/useWebSocket'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const { connected: wsConnected, connect } = useWebSocket()

const isCollapsed = ref(false)

const currentRoute = computed(() => route.path)
const username = computed(() => authStore.user?.username || '用户')

function handleCommand(command: string) {
  if (command === 'logout') {
    authStore.logout()
    router.push('/login')
  }
}

onMounted(() => {
  connect()
})
</script>

<style lang="scss" scoped>
.admin-layout {
  height: 100vh;
}

.sidebar {
  background-color: #1e1b4b;
  transition: width 0.3s ease;
  overflow: hidden;
  box-shadow: 2px 0 8px rgba(0, 0, 0, 0.15);

  .sidebar-logo {
    height: 64px;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 10px;
    border-bottom: 1px solid rgba(255, 255, 255, 0.1);
    padding: 0 16px;
    white-space: nowrap;
    overflow: hidden;

    .logo-text {
      font-size: 18px;
      font-weight: 700;
      color: #fff;
      letter-spacing: 1px;
    }
  }

  .sidebar-menu {
    border-right: none;

    &::-webkit-scrollbar {
      width: 0;
    }

    :deep(.el-menu-item),
    :deep(.el-sub-menu__title) {
      &:hover {
        background-color: rgba(255, 255, 255, 0.08) !important;
      }
    }

    :deep(.el-menu-item.is-active) {
      background-color: #3730a3 !important;
    }

    :deep(.el-sub-menu.is-active > .el-sub-menu__title) {
      color: #fff !important;
    }
  }
}

.main-container {
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.header {
  background: #fff;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
  height: 64px;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.08);
  z-index: 10;

  .header-left {
    display: flex;
    align-items: center;
    gap: 16px;

    .collapse-btn {
      cursor: pointer;
      color: #6b7280;
      transition: color 0.2s;

      &:hover {
        color: #1e1b4b;
      }
    }
  }

  .header-right {
    display: flex;
    align-items: center;
    gap: 20px;

    .ws-badge {
      :deep(.el-badge__content.is-dot) {
        background-color: #10b981;
      }
    }

    .user-dropdown {
      display: flex;
      align-items: center;
      gap: 8px;
      cursor: pointer;
      color: #374151;

      .username {
        font-size: 14px;
        font-weight: 500;
      }
    }
  }
}

.main-content {
  background: #f0f2f5;
  overflow-y: auto;
  padding: 20px;
}
</style>
