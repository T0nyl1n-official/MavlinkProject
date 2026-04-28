import { defineStore } from 'pinia'
import { getProfileApi, loginApi, logoutApi } from '@/api/auth'
import type { LoginParams, LoginResponse, UserProfileResponse, LoginResponseData, UserProfileData } from '@/types/auth'

const TOKEN_KEY = 'token'
const USER_KEY = 'user'
const ROLE_KEY = 'role'

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

function resolveUserRole(profile: LoginResponseData) {
  return (
    profile.Role ||
    (profile.is_admin === true ? 'admin' : profile.is_admin === false ? 'user' : null)
  )
}

function resetSessionState(store: AuthState) {
  store.token = null
  store.userInfo = null
  store.role = null
  localStorage.removeItem(TOKEN_KEY)
  localStorage.removeItem(USER_KEY)
  localStorage.removeItem(ROLE_KEY)
}

export const useAuthStore = defineStore('auth', {
  state: (): AuthState => ({
    token: localStorage.getItem(TOKEN_KEY),
    userInfo: getStoredUser(),
    role: localStorage.getItem(ROLE_KEY)
  }),
  actions: {
    async login(credentials: LoginParams) {
      const response: LoginResponse = await loginApi(credentials)
      const token = response.data?.token || response.token
      const profile = response.data || {}
      const role = resolveUserRole(profile)

      if (response.success && token) {
        this.token = token
        this.userInfo = profile
        this.role = role

        localStorage.setItem(TOKEN_KEY, token)
        localStorage.setItem(USER_KEY, JSON.stringify(profile))
        if (role) {
          localStorage.setItem(ROLE_KEY, role)
        }

        return response
      }

      throw new Error(response.message || response.data?.message || '账号或密码不对')
    },
    async fetchProfile() {
      const response: UserProfileResponse = await getProfileApi()

      if (response.success && response.data) {
        this.userInfo = response.data.user || response.data
      }
    },
    async logout() {
      try {
        await logoutApi()
      } finally {
        resetSessionState(this)
      }
    }
  }
})