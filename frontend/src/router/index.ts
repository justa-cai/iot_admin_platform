import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/auth/LoginView.vue'),
    meta: { layout: 'auth', requiresAuth: false },
  },
  {
    path: '/',
    redirect: '/dashboard',
  },
  {
    path: '/dashboard',
    name: 'Dashboard',
    component: () => import('@/views/dashboard/DashboardView.vue'),
    meta: { title: '仪表盘', icon: 'Odometer' },
  },
  {
    path: '/devices',
    name: 'DeviceList',
    component: () => import('@/views/device/DeviceListView.vue'),
    meta: { title: '设备管理', icon: 'Monitor' },
  },
  {
    path: '/devices/create',
    name: 'DeviceCreate',
    component: () => import('@/views/device/DeviceForm.vue'),
    meta: { title: '创建设备', hidden: true },
  },
  {
    path: '/devices/:id',
    name: 'DeviceDetail',
    component: () => import('@/views/device/DeviceDetailView.vue'),
    meta: { title: '设备详情', hidden: true },
  },
  {
    path: '/devices/:id/edit',
    name: 'DeviceEdit',
    component: () => import('@/views/device/DeviceForm.vue'),
    meta: { title: '编辑设备', hidden: true },
  },
  {
    path: '/groups',
    name: 'Groups',
    component: () => import('@/views/group/GroupListView.vue'),
    meta: { title: '设备分组', icon: 'Grid' },
  },
  {
    path: '/messages/console',
    name: 'MessageConsole',
    component: () => import('@/views/message/MessageConsole.vue'),
    meta: { title: '消息控制台', icon: 'ChatDotRound' },
  },
  {
    path: '/messages/history',
    name: 'MessageHistory',
    component: () => import('@/views/message/MessageHistory.vue'),
    meta: { title: '消息历史', icon: 'Document' },
  },
  {
    path: '/rules',
    name: 'RuleList',
    component: () => import('@/views/rule/RuleListView.vue'),
    meta: { title: '规则引擎', icon: 'SetUp' },
  },
  {
    path: '/rules/create',
    name: 'RuleCreate',
    component: () => import('@/views/rule/RuleForm.vue'),
    meta: { title: '创建规则', hidden: true },
  },
  {
    path: '/rules/:id/edit',
    name: 'RuleEdit',
    component: () => import('@/views/rule/RuleForm.vue'),
    meta: { title: '编辑规则', hidden: true },
  },
  {
    path: '/telemetry',
    name: 'Telemetry',
    component: () => import('@/views/telemetry/TelemetryView.vue'),
    meta: { title: '数据查询', icon: 'TrendCharts' },
  },
  {
    path: '/firmware',
    name: 'Firmware',
    component: () => import('@/views/firmware/FirmwareView.vue'),
    meta: { title: '固件管理', icon: 'Upload' },
  },
  {
    path: '/system/users',
    name: 'Users',
    component: () => import('@/views/system/UserView.vue'),
    meta: { title: '用户管理', icon: 'User' },
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

router.beforeEach((to, _from, next) => {
  const authStore = useAuthStore()
  if (to.meta.requiresAuth === false) {
    if (authStore.isLoggedIn && to.name === 'Login') {
      next('/dashboard')
      return
    }
    next()
    return
  }
  if (!authStore.isLoggedIn) {
    next({ name: 'Login', query: { redirect: to.fullPath } })
    return
  }
  next()
})

export default router
