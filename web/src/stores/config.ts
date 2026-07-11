import { defineStore } from 'pinia'
import { ref } from 'vue'
import { configApi, getApiError } from '../api'

export const useConfigStore = defineStore('config', () => {
  const websiteTitle = ref('SubDock')
  const loaded = ref(false)
  const error = ref<string | null>(null)

  /** 获取公开站点配置，失败时保留默认标题并暴露错误状态。 */
  const fetchConfig = async (): Promise<void> => {
    if (loaded.value) return
    error.value = null

    try {
      const res = await configApi.get()
      websiteTitle.value = res.data.website_title || 'SubDock'
      document.title = websiteTitle.value
      loaded.value = true
    } catch (cause: unknown) {
      error.value = getApiError(cause, '获取站点配置失败')
    }
  }

  return { websiteTitle, loaded, error, fetchConfig }
})
