import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { User } from '@/types'
import { login as apiLogin, getProfile } from '@/api/auth'

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string>(localStorage.getItem('token') || '')
  const user = ref<User | null>(null)

  const isLoggedIn = computed(() => !!token.value)
  const isAdmin = computed(() => user.value?.role === 'admin')

  async function login(username: string, password: string) {
    const res = await apiLogin({ username, password })
    token.value = res.token
    user.value = res.user
    localStorage.setItem('token', res.token)
    localStorage.setItem('user', JSON.stringify(res.user))
  }

  async function fetchProfile() {
    const profile = await getProfile()
    user.value = profile
    localStorage.setItem('user', JSON.stringify(profile))
  }

  function logout() {
    token.value = ''
    user.value = null
    localStorage.removeItem('token')
    localStorage.removeItem('user')
  }

  function init() {
    if (token.value) {
      const savedUser = localStorage.getItem('user')
      if (savedUser) {
        try {
          user.value = JSON.parse(savedUser)
        } catch {
          logout()
        }
      }
    }
  }

  return { token, user, isLoggedIn, isAdmin, login, fetchProfile, logout, init }
})
