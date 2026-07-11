import apiClient from './client'
import type {
  MessageResponse,
  SubscriptionRecord,
  SubscriptionRenewal,
  SubscriptionWriteRequest,
  UpdateSubscriptionRequest
} from './contracts'

/** 订阅管理接口。 */
export const subscriptionApi = {
  /** 获取全部订阅。 */
  list() {
    return apiClient.get<SubscriptionRecord[]>('/subscriptions')
  },

  /** 创建订阅。 */
  create(data: SubscriptionWriteRequest) {
    return apiClient.post<SubscriptionRecord>('/subscriptions', data)
  },

  /** 更新指定订阅。 */
  update(id: number, data: UpdateSubscriptionRequest) {
    return apiClient.put<SubscriptionRecord>(`/subscriptions/${id}`, data)
  },

  /** 删除指定订阅。 */
  delete(id: number) {
    return apiClient.delete<MessageResponse>(`/subscriptions/${id}`)
  },

  /** 发送指定订阅的测试通知。 */
  testNotify(id: number) {
    return apiClient.post<MessageResponse>(`/subscriptions/${id}/test-notify`)
  },

  /** 手动续订一次并返回更新后的订阅。 */
  renew(id: number) {
    return apiClient.post<SubscriptionRecord>(`/subscriptions/${id}/renew`)
  },

  /** 获取指定订阅的全部续订历史。 */
  listRenewals(id: number) {
    return apiClient.get<SubscriptionRenewal[]>(`/subscriptions/${id}/renewals`)
  }
}
