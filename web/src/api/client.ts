import axios from 'axios'
import type { AxiosInstance } from 'axios'

/** 判断未知值是否为可读取字段的普通对象。 */
const isRecord = (value: unknown): value is Record<string, unknown> => {
  return typeof value === 'object' && value !== null
}

/** 从后端错误响应中提取可展示的中文或业务错误信息。 */
const extractErrorMessage = (payload: unknown): string | null => {
  if (!isRecord(payload)) return null

  const candidates = [payload.error, payload.message]
  for (const candidate of candidates) {
    if (typeof candidate === 'string' && candidate.trim()) {
      return candidate.trim()
    }
  }

  return null
}

/**
 * 将未知请求异常转换为可展示信息。
 * @param error 捕获到的未知异常。
 * @param fallback 后端未返回有效信息时使用的兜底文案。
 * @returns 后端业务错误或兜底文案。
 */
export const getApiError = (error: unknown, fallback: string): string => {
  if (!axios.isAxiosError(error)) return fallback
  return extractErrorMessage(error.response?.data) ?? fallback
}

const apiClient: AxiosInstance = axios.create({
  baseURL: '/api',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json'
  }
})

apiClient.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error: unknown) => Promise.reject(error)
)

apiClient.interceptors.response.use(
  (response) => response,
  (error: unknown) => {
    if (axios.isAxiosError(error) && error.response?.status === 401) {
      const isLoginRequest = error.config?.url?.includes('/login') ?? false
      if (!isLoginRequest) {
        localStorage.removeItem('token')
        if (window.location.pathname !== '/login') {
          window.location.assign('/login')
        }
      }
    }
    return Promise.reject(error)
  }
)

export { apiClient }
export default apiClient
