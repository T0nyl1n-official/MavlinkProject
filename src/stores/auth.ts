import { defineStore } from 'pinia'
import { getProfileApi, loginApi, logoutApi } from '@/api/auth'
import type { LoginParams, LoginResponse, UserProfileResponse, LoginResponseData, UserProfileData } from '@/types/auth'

function getStoredUser(): LoginResponseData | null {
  const rawUser = localStorage.getItem('user')
  if (!rawUser) return null

  try {
    return JSON.parse(rawUser) as LoginResponseData
  } catch {
    return null
  }
}

interface AuthState {
  token: string | null
  userInfo: UserProfileData['user'] | LoginResponseData | null
  role: string | null
}

export const useAuthStore = defineStore('auth', {
  state: (): AuthState => ({
    token: localStorage.getItem('token'),
    userInfo: getStoredUser(),
    role: localStorage.getItem('role')
  }),
  actions: {
    async login(params: LoginParams) {
      const res: LoginResponse = await loginApi(params)

      if (res.success && res.data?.token) {
        this.token = res.data.token
        this.userInfo = res.data
        this.role = res.data.Role || null

        localStorage.setItem('token', res.data.token)
        localStorage.setItem('user', JSON.stringify(res.data))
        if (res.data.Role) {
          localStorage.setItem('role', res.data.Role)
        }

        return res
      }

      throw new Error(res.message || res.data?.message || '登录失败')
    },
    async fetchProfile() {
      const res: UserProfileResponse = await getProfileApi()

      if (res.success && res.data) {
        this.userInfo = res.data.user || res.data
      }
    },
    async logout() {
      try {
        await logoutApi()
      } finally {
        this.token = null
        this.userInfo = null
        this.role = null
        localStorage.removeItem('token')
        localStorage.removeItem('user')
        localStorage.removeItem('role')
      }
    }
  }
})