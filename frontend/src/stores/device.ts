import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { Device, PaginatedResponse } from '@/types'
import { getDevices, deleteDevice } from '@/api/device'
import type { DeviceQueryParams } from '@/api/device'

export const useDeviceStore = defineStore('device', () => {
  const devices = ref<Device[]>([])
  const total = ref(0)
  const loading = ref(false)
  const currentQuery = ref<DeviceQueryParams>({})

  async function fetchDevices(params?: DeviceQueryParams) {
    loading.value = true
    try {
      currentQuery.value = params || {}
      const res: PaginatedResponse<Device> = await getDevices(params)
      devices.value = res.data || []
      total.value = res.total || 0
    } finally {
      loading.value = false
    }
  }

  async function removeDevice(id: number) {
    await deleteDevice(id)
    await fetchDevices(currentQuery.value)
  }

  return { devices, total, loading, currentQuery, fetchDevices, removeDevice }
})
