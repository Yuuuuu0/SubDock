import { createApp } from 'vue'
import { createPinia } from 'pinia'
import naive from 'naive-ui'

import App from './App.vue'
import router from './router'
import './style.css'
import { useConfigStore } from './stores/config'

const app = createApp(App)
const pinia = createPinia()

app.use(pinia)
app.use(router)
app.use(naive)

// 加载网站配置
const configStore = useConfigStore(pinia)
configStore.fetchConfig().then(() => {
  document.title = configStore.websiteTitle
})

app.mount('#app')
