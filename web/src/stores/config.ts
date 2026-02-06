import { defineStore } from 'pinia'
import { ref } from 'vue'
import { configApi } from '../api'

export const useConfigStore = defineStore('config', () => {
  const websiteTitle = ref('SubDock')
  const loaded = ref(false)

  const fetchConfig = async () => {
    if (loaded.value) return
    try {
      const res = await configApi.get()
      websiteTitle.value = res.data.website_title || 'SubDock'
      document.title = websiteTitle.value
      loaded.value = true
    } catch {
    }
  }

  return { websiteTitle, loaded, fetchConfig }
})
