import axios from 'axios'
import { ElMessage } from 'element-plus'

// 从响应数据中提取错误信息
function getErrorMessage(responseData: any): string {
  if (!responseData || typeof responseData !== 'object') {
    return '请求失败'
  }

  return (
    responseData.message ||
    responseData.err_info?.message ||
    responseData.error?.message ||
    responseData.err_info?.description ||
    responseData.data?.message ||
    responseData.validations?.[0]?.message ||
    '请求失败'
  )
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
    const responseData = response.data
    
    // 处理业务错误
    if (responseData.success !== undefined && !responseData.success) {
      const errorMessage = getErrorMessage(responseData)
      ElMessage.error(errorMessage)
      return Promise.reject(new Error(errorMessage))
    }
    
    return responseData
  },
  (error) => {
    // 处理网络和服务器错误
    if (error.response?.status === 401) {
      // 登录过期，跳转到登录页
      localStorage.removeItem('token')
      localStorage.removeItem('user')
      localStorage.removeItem('role')
      window.location.href = '/login'
      ElMessage.error('登录已过期，请重新登录')
    } else if (error.code === 'ERR_NETWORK') {
      ElMessage.error('网络错误，无法连接服务器')
    } else {
      const errorMessage = getErrorMessage(error.response?.data) || '请求失败，请检查后端服务'
      ElMessage.error(errorMessage)
    }
    return Promise.reject(error)
  }
)

export default request

