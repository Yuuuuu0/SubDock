import axios from 'axios'
import type { AxiosInstance } from 'axios'

// --- Types ---

export interface LoginResponse {
  token: string
}

export interface Subscription {
  id?: number
  name: string
  amount: number
  currency: string
  start_date: string // YYYY-MM-DD
  cycle_value: number
  cycle_unit: 'day' | 'month' | 'quarter' | 'half_year' | 'year'
  expire_date: string | null // YYYY-MM-DD or null
  remind_days: number
  remark?: string
}

export interface Settings {
  notify_hours: number[]
  telegram_bot_token: string
  telegram_chat_id: string
  bark_url: string
}

// 后端返回的设置格式（notify_hours 为逗号分隔的字符串）
export interface SettingsResponse {
  notify_hours: string
  telegram_bot_token: string
  telegram_chat_id: string
  bark_url: string
}

export interface PasswordChange {
  old_password?: string
  new_password?: string
}

// --- API Client ---

const apiClient: AxiosInstance = axios.create({
  baseURL: '/api',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json'
  }
})

// Request interceptor to add token
apiClient.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// Response interceptor to handle 401
apiClient.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response && error.response.status === 401) {
      localStorage.removeItem('token')
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

export default apiClient

// --- API Functions ---

export const authApi = {
  login(data: { username?: string; password?: string }) {
    return apiClient.post<LoginResponse>('/login', data)
  },
  changePassword(data: PasswordChange) {
    return apiClient.post('/change-password', data)
  }
}

export const subscriptionApi = {
  list() {
    return apiClient.get<Subscription[]>('/subscriptions')
  },
  create(data: Subscription) {
    return apiClient.post('/subscriptions', data)
  },
  update(id: number, data: Subscription) {
    return apiClient.put(`/subscriptions/${id}`, data)
  },
  delete(id: number) {
    return apiClient.delete(`/subscriptions/${id}`)
  },
  testNotify(id: number) {
    return apiClient.post(`/subscriptions/${id}/test-notify`)
  }
}

export interface PublicConfig {
  website_title: string
}

export const configApi = {
  get() {
    return apiClient.get<PublicConfig>('/config')
  }
}

export const settingsApi = {
  get() {
    return apiClient.get<SettingsResponse>('/settings')
  },
  update(data: SettingsResponse) {
    return apiClient.put('/settings', data)
  },
  testNotify(type: 'telegram' | 'bark') {
    return apiClient.post('/settings/test-notify', { type })
  }
}
