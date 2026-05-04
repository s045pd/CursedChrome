import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { auth as authApi } from '@/api/endpoints'
import { ApiError } from '@/api/client'
import type { MeResult } from '@/types/api'

export const useAuthStore = defineStore('auth', () => {
  const me = ref<MeResult | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  const isAuthenticated = computed(() => me.value !== null)
  const mustChangePassword = computed(() => me.value?.password_should_be_changed === true)

  async function refresh(): Promise<boolean> {
    try {
      me.value = await authApi.me()
      return true
    } catch (e) {
      if (e instanceof ApiError && e.status === 401) {
        me.value = null
        return false
      }
      throw e
    }
  }

  async function login(username: string, password: string): Promise<void> {
    loading.value = true
    error.value = null
    try {
      me.value = await authApi.login(username, password)
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'login failed'
      throw e
    } finally {
      loading.value = false
    }
  }

  async function logout(): Promise<void> {
    try {
      await authApi.logout()
    } finally {
      me.value = null
    }
  }

  async function changePassword(newPassword: string): Promise<void> {
    await authApi.changePassword(newPassword)
    if (me.value) me.value.password_should_be_changed = false
  }

  return { me, loading, error, isAuthenticated, mustChangePassword, refresh, login, logout, changePassword }
})
