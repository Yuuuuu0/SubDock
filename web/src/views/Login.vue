<template>
  <div class="login-container">
    <div class="login-box">
      <div class="login-header">
        <h1 class="app-title">{{ websiteTitle }}</h1>
        <p class="app-subtitle">订阅管理系统</p>
      </div>
      <n-card class="login-card" :bordered="false" size="large">
        <n-form ref="formRef" :model="formValue" :rules="rules">
          <n-form-item path="username" label="用户名">
            <n-input v-model:value="formValue.username" placeholder="请输入用户名" @keydown.enter="handleLogin">
              <template #prefix>
                <n-icon :component="PersonOutline" />
              </template>
            </n-input>
          </n-form-item>
          <n-form-item path="password" label="密码">
            <n-input
              v-model:value="formValue.password"
              type="password"
              show-password-on="click"
              placeholder="请输入密码"
              @keydown.enter="handleLogin"
            >
              <template #prefix>
                <n-icon :component="LockClosedOutline" />
              </template>
            </n-input>
          </n-form-item>
          <n-button type="primary" block size="large" :loading="loading" @click="handleLogin">
            登录
          </n-button>
        </n-form>
      </n-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useMessage } from 'naive-ui'
import type { FormInst } from 'naive-ui'
import { PersonOutline, LockClosedOutline } from '@vicons/ionicons5'
import { useAuthStore } from '../stores/auth'
import { useConfigStore } from '../stores/config'
import { authApi } from '../api'

const router = useRouter()
const message = useMessage()
const authStore = useAuthStore()
const configStore = useConfigStore()

const websiteTitle = computed(() => configStore.websiteTitle)

const formRef = ref<FormInst | null>(null)
const loading = ref(false)

const formValue = ref({
  username: '',
  password: ''
})

const rules = {
  username: {
    required: true,
    message: '请输入用户名',
    trigger: 'blur'
  },
  password: {
    required: true,
    message: '请输入密码',
    trigger: 'blur'
  }
}

const handleLogin = (e: Event) => {
  e.preventDefault()
  formRef.value?.validate(async (errors) => {
    if (!errors) {
      loading.value = true
      try {
        const res = await authApi.login(formValue.value)
        authStore.setToken(res.data.token)
        message.success('登录成功')
        router.push('/')
      } catch (error: any) {
        message.error(error.response?.data?.error || '登录失败，请检查用户名和密码')
      } finally {
        loading.value = false
      }
    }
  })
}
</script>

<style scoped>
.login-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  background-color: #F3F4F6;
}

.login-box {
  width: 100%;
  max-width: 400px;
  padding: 20px;
}

.login-header {
  text-align: center;
  margin-bottom: 32px;
}

.app-title {
  font-size: 2.5rem;
  font-weight: 800;
  margin: 0;
  color: #1A1A1A;
  letter-spacing: -0.025em;
}

.app-subtitle {
  font-size: 1.1rem;
  color: #6B7280;
  margin-top: 8px;
}

.login-card {
  box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.1), 0 4px 6px -2px rgba(0, 0, 0, 0.05);
  background-color: white;
  border-radius: 12px;
}
</style>
