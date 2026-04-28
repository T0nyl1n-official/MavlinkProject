import axios from 'axios'
import { ElMessage } from 'element-plus'

function resolveErrorMessage(payload: any): string {
  if (!payload || typeof payload !== 'object') {
    return '请求失败'
  }

  return (
    payload.message ||
    payload.err_info?.message ||
    payload.error?.message ||
    payload.err_info?.description ||
    payload.data?.message ||
    payload.validations?.[0]?.message ||
    '请求失败'
  )
}

function shouldBypass401Redirect(): boolean {
  if (typeof window === 'undefined') return false

  return ['/test-api', '/mavlink'].some(path => window.location.pathname.startsWith(path))
}

function clearSession() {
  localStorage.removeItem('token')
  localStorage.removeItem('user')
  localStorage.removeItem('role')
}

function normalizeResponse(responseBody: any) {
  if (
    responseBody &&
    typeof responseBody === 'object' &&
    responseBody.success === undefined &&
    typeof responseBody.code === 'number'
  ) {
    responseBody.success = responseBody.code === 0
  }

  return responseBody
}

// 创建 axios 实例
const request = axios.create({
  baseURL: '',
  timeout: 15000,
  headers: {
    'Content-Type': 'application/json'
  }
})

// 请求拦截器 - 添加认证 token
request.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => Promise.reject(error)
)

// 响应拦截器 - 统一处理响应
request.interceptors.response.use(
  (response) => {
    const responseBody = normalizeResponse(response.data)
    
    if (responseBody?.success !== undefined && !responseBody.success) {
      const message = resolveErrorMessage(responseBody)
      ElMessage.error(message)
      return Promise.reject(new Error(message))
    }
    
    return responseBody
  },
  (error) => {
    if (error.response?.status === 401) {
      if (shouldBypass401Redirect()) {
        const authError = new Error('401 Unauthorized')
        ;(authError as any).response = error.response
        return Promise.reject(authError)
      }

      clearSession()
      ElMessage.error('登录已过期，请重新登录')
      window.location.href = '/login'
    } else if (error.code === 'ERR_NETWORK') {
      ElMessage.error('网络连不上，请稍后再试')
    } else {
      ElMessage.error(resolveErrorMessage(error.response?.data) || '请求没有成功，请稍后再试')
    }
    return Promise.reject(error)
  }
)

export default request

