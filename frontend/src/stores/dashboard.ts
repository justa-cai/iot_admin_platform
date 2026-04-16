import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { DashboardStats, ThroughputData } from '@/types'
import { getDashboardStats, getThroughput } from '@/api/dashboard'

export const useDashboardStore = defineStore('dashboard', () => {
  const stats = ref<DashboardStats>({
    total_devices: 0,
    online_devices: 0,
    offline_devices: 0,
    inactive_devices: 0,
    messages_today: 0,
    total_rules: 0,
    active_rules: 0,
    alerts_today: 0,
  })
  const throughput = ref<ThroughputData>({ timestamps: [], values: [] })
  const loading = ref(false)

  async function fetchStats() {
    loading.value = true
    try {
      stats.value = await getDashboardStats()
    } finally {
      loading.value = false
    }
  }

  async function fetchThroughput(hours: number = 24) {
    throughput.value = await getThroughput(hours)
  }

  async function refresh() {
    await Promise.all([fetchStats(), fetchThroughput()])
  }

  return { stats, throughput, loading, fetchStats, fetchThroughput, refresh }
})
