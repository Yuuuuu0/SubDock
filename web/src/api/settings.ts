import apiClient from './client'
import type { MessageResponse, NotificationType, SettingsResponse } from './contracts'

/** 系统设置接口。 */
export const settingsApi = {
  /** 获取通知渠道与通知时段设置。 */
  get() {
    return apiClient.get<SettingsResponse>('/settings')
  },

  /** 保存通知渠道与通知时段设置。 */
  update(data: SettingsResponse) {
    return apiClient.put<MessageResponse>('/settings', data)
  },

  /** 向指定渠道发送测试通知。 */
  testNotify(type: NotificationType) {
    return apiClient.post<MessageResponse>('/settings/test-notify', { type })
  }
}
