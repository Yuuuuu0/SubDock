import { createApp } from 'vue'
import { createPinia } from 'pinia'

import App from './App.vue'
import router from './router'
import './style.css'
import { useConfigStore } from './stores/config'

const app = createApp(App)
const pinia = createPinia()

app.use(pinia)
app.use(router)

// 加载公开站点配置；失败时 Store 会保留默认标题。
const configStore = useConfigStore(pinia)
void configStore.fetchConfig()

app.mount('#app')
