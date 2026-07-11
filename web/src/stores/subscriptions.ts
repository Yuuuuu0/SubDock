import { defineStore } from 'pinia'
import { ref } from 'vue'

import { getApiError, subscriptionApi } from '../api'
import type { SubscriptionRecord } from '../api'

/** 集中维护订阅列表、加载状态与刷新行为。 */
export const useSubscriptionsStore = defineStore('subscriptions', () => {
  const items = ref<SubscriptionRecord[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)
  const loaded = ref(false)

  let activeRequest: Promise<void> | null = null

  /** 请求订阅数据；刷新失败时保留已有数据，便于页面继续展示。 */
  const requestItems = (force: boolean): Promise<void> => {
    if (!force && loaded.value) return Promise.resolve()
    if (activeRequest) return activeRequest

    activeRequest = (async () => {
      loading.value = true
      error.value = null

      try {
        const response = await subscriptionApi.list()
        items.value = response.data.map((item) => ({
          ...item,
          expire_date: item.expire_date || null,
          auto_renew: item.auto_renew ?? false,
          renew_count: item.renew_count ?? 0,
          remark: item.remark ?? ''
        }))
        loaded.value = true
      } catch (cause: unknown) {
        error.value = getApiError(cause, '获取订阅列表失败')
      } finally {
        loading.value = false
      }
    })().finally(() => {
      activeRequest = null
    })

    return activeRequest
  }

  /** 首次读取订阅；已有成功缓存时不会重复请求。 */
  const fetch = (): Promise<void> => requestItems(false)

  /** 强制刷新订阅列表。 */
  const refresh = (): Promise<void> => requestItems(true)

  return { items, loading, error, loaded, fetch, refresh }
})
