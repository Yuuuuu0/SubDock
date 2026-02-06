<template>
  <div class="layout-container">
    <nav class="navbar">
      <div class="navbar-content">
        <div class="navbar-left">
          <div class="logo">{{ websiteTitle }}</div>
          <div class="current-time">{{ currentTime }}</div>
        </div>
        <div class="navbar-right">
          <router-link to="/subscriptions" class="nav-item" active-class="active">
            <n-icon :component="ListOutline" />
            <span>订阅列表</span>
          </router-link>
          <router-link to="/settings" class="nav-item" active-class="active">
            <n-icon :component="SettingsOutline" />
            <span>系统设置</span>
          </router-link>
          <div class="nav-item logout" @click="handleLogout">
            <n-icon :component="LogOutOutline" />
            <span>退出登录</span>
          </div>
        </div>
      </div>
    </nav>

    <main class="main-content">
      <router-view />
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useMessage } from 'naive-ui'
import { ListOutline, SettingsOutline, LogOutOutline } from '@vicons/ionicons5'
import { useAuthStore } from '../stores/auth'
import { useConfigStore } from '../stores/config'

const router = useRouter()
const message = useMessage()
const authStore = useAuthStore()
const configStore = useConfigStore()

const websiteTitle = computed(() => configStore.websiteTitle)

const currentTime = ref('')
let timer: ReturnType<typeof setInterval>

const updateTime = () => {
  const now = new Date()
  currentTime.value = now.toLocaleString('zh-CN', {
    timeZoneName: 'short',
    hour12: false,
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  })
}

const handleLogout = () => {
  authStore.logout()
  message.success('已退出登录')
  router.push('/login')
}

onMounted(() => {
  updateTime()
  timer = setInterval(updateTime, 1000)
})

onUnmounted(() => {
  if (timer) clearInterval(timer)
})
</script>

<style scoped>
.layout-container {
  min-height: 100vh;
  background-color: #F3F4F6;
  padding-top: 64px; /* Space for fixed navbar */
}

.navbar {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  height: 64px;
  background-color: white;
  box-shadow: 0 1px 3px 0 rgba(0, 0, 0, 0.1), 0 1px 2px 0 rgba(0, 0, 0, 0.06);
  z-index: 1000;
}

.navbar-content {
  max-width: 1280px;
  margin: 0 auto;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
}

.navbar-left {
  display: flex;
  align-items: center;
  gap: 24px;
}

.logo {
  font-size: 1.25rem;
  font-weight: 800;
  color: #1A1A1A;
  letter-spacing: -0.025em;
}

.current-time {
  font-size: 0.875rem;
  color: #6B7280;
  font-family: monospace;
}

.navbar-right {
  display: flex;
  align-items: center;
  gap: 8px;
  height: 100%;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 0 16px;
  height: 100%;
  color: #6B7280;
  font-weight: 500;
  font-size: 0.95rem;
  cursor: pointer;
  text-decoration: none;
  transition: all 0.2s;
  position: relative;
}

.nav-item:hover {
  color: #1A1A1A;
  background-color: #F9FAFB;
}

.nav-item.active {
  color: #1A1A1A;
}

.nav-item.active::after {
  content: '';
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  height: 2px;
  background-color: #1A1A1A;
}

.nav-item.logout:hover {
  color: #EF4444;
  background-color: #FEF2F2;
}

.main-content {
  max-width: 1280px;
  margin: 0 auto;
  padding: 32px 24px;
}

@media (max-width: 640px) {
  .current-time {
    display: none;
  }
  
  .navbar-content {
    padding: 0 16px;
  }
  
  .nav-item span {
    display: none;
  }
  
  .nav-item {
    padding: 0 12px;
  }
}
</style>
