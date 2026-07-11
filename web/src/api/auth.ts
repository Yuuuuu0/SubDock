import apiClient from './client'
import type { LoginRequest, LoginResponse, MessageResponse, PasswordChange } from './contracts'

/** 认证相关接口。 */
export const authApi = {
  /** 登录并获取访问令牌。 */
  login(data: LoginRequest) {
    return apiClient.post<LoginResponse>('/login', data)
  },

  /** 使用旧密码更新当前账户密码。 */
  changePassword(data: PasswordChange) {
    return apiClient.post<MessageResponse>('/change-password', data)
  }
}
