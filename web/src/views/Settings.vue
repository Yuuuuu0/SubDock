<template>
  <div class="page-container">
    <div class="page-header">
      <div>
        <h1 class="page-title">系统设置</h1>
        <p class="page-subtitle">管理通知配置和账户安全</p>
      </div>
    </div>

    <n-space vertical size="large">
      <!-- Notification Settings -->
      <n-card title="通知设置" :bordered="false" class="settings-card">
        <template #header-extra>
          <n-icon size="20" :component="NotificationsOutline" />
        </template>
        
        <n-form
          :model="notifyForm"
          label-placement="left"
          label-width="120"
          require-mark-placement="right-hanging"
        >
          <n-form-item label="通知时段">
            <n-checkbox-group v-model:value="notifyForm.notify_hours">
              <n-space item-style="display: flex;">
                <n-checkbox v-for="h in 24" :key="h-1" :value="h-1" :label="`${h-1}点`" />
              </n-space>
            </n-checkbox-group>
          </n-form-item>

          <n-divider title-placement="left">Telegram 通知</n-divider>
          
          <n-form-item label="Bot Token">
            <n-input v-model:value="notifyForm.telegram_bot_token" type="password" show-password-on="click" placeholder="请输入 Bot Token" />
          </n-form-item>
          <n-form-item label="Chat ID">
            <n-input-group>
              <n-input v-model:value="notifyForm.telegram_chat_id" placeholder="请输入 Chat ID" />
              <n-button @click="testNotify('telegram')" :loading="testingTelegram">
                测试
              </n-button>
            </n-input-group>
          </n-form-item>

          <n-divider title-placement="left">Bark 通知 (iOS)</n-divider>

          <n-form-item label="Bark URL">
            <n-input-group>
              <n-input v-model:value="notifyForm.bark_url" placeholder="https://api.day.app/..." />
              <n-button @click="testNotify('bark')" :loading="testingBark">
                测试
              </n-button>
            </n-input-group>
          </n-form-item>

          <n-row>
            <n-col :span="24">
              <div style="display: flex; justify-content: flex-end">
                <n-button type="primary" @click="saveSettings" :loading="savingSettings">
                  保存配置
                </n-button>
              </div>
            </n-col>
          </n-row>
        </n-form>
      </n-card>

      <!-- Password Change -->
      <n-card title="修改密码" :bordered="false" class="settings-card">
        <template #header-extra>
          <n-icon size="20" :component="LockClosedOutline" />
        </template>

        <n-form
          ref="pwdFormRef"
          :model="pwdForm"
          :rules="pwdRules"
          label-placement="left"
          label-width="120"
          require-mark-placement="right-hanging"
        >
          <n-form-item label="当前密码" path="old_password">
            <n-input v-model:value="pwdForm.old_password" type="password" show-password-on="click" />
          </n-form-item>
          <n-form-item label="新密码" path="new_password">
            <n-input v-model:value="pwdForm.new_password" type="password" show-password-on="click" />
          </n-form-item>
          <n-form-item label="确认新密码" path="confirm_password">
            <n-input v-model:value="pwdForm.confirm_password" type="password" show-password-on="click" />
          </n-form-item>

          <n-row>
            <n-col :span="24">
              <div style="display: flex; justify-content: flex-end">
                <n-button type="primary" @click="changePassword" :loading="changingPwd">
                  修改密码
                </n-button>
              </div>
            </n-col>
          </n-row>
        </n-form>
      </n-card>
    </n-space>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useMessage } from 'naive-ui'
import type { FormInst, FormRules } from 'naive-ui'
import { NotificationsOutline, LockClosedOutline } from '@vicons/ionicons5'
import { settingsApi, authApi } from '../api'
import type { Settings } from '../api'

const message = useMessage()

// Settings State
const notifyForm = ref<Settings>({
  notify_hours: [],
  telegram_bot_token: '',
  telegram_chat_id: '',
  bark_url: ''
})
const savingSettings = ref(false)
const testingTelegram = ref(false)
const testingBark = ref(false)

// Password State
const pwdFormRef = ref<FormInst | null>(null)
const pwdForm = ref({
  old_password: '',
  new_password: '',
  confirm_password: ''
})
const changingPwd = ref(false)

const pwdRules: FormRules = {
  old_password: { required: true, message: '请输入当前密码', trigger: 'blur' },
  new_password: { required: true, message: '请输入新密码', trigger: 'blur' },
  confirm_password: {
    required: true,
    message: '请确认新密码',
    trigger: 'blur',
    validator: (_rule, value) => {
      if (value !== pwdForm.value.new_password) {
        return new Error('两次输入的密码不一致')
      }
      return true
    }
  }
}

// Methods
const fetchSettings = async () => {
  try {
    const res = await settingsApi.get()
    const data = res.data
    notifyForm.value = {
      notify_hours: data.notify_hours ? data.notify_hours.split(',').map(h => parseInt(h, 10)) : [],
      telegram_bot_token: data.telegram_bot_token,
      telegram_chat_id: data.telegram_chat_id,
      bark_url: data.bark_url
    }
  } catch (error) {
    message.error('获取设置失败')
  }
}

const saveSettings = async () => {
  savingSettings.value = true
  try {
    const payload = {
      notify_hours: notifyForm.value.notify_hours.join(','),
      telegram_bot_token: notifyForm.value.telegram_bot_token,
      telegram_chat_id: notifyForm.value.telegram_chat_id,
      bark_url: notifyForm.value.bark_url
    }
    await settingsApi.update(payload)
    message.success('设置已保存')
  } catch (error) {
    message.error('保存失败')
  } finally {
    savingSettings.value = false
  }
}

const testNotify = async (type: 'telegram' | 'bark') => {
  if (type === 'telegram') testingTelegram.value = true
  else testingBark.value = true

  try {
    const payload = {
      notify_hours: notifyForm.value.notify_hours.join(','),
      telegram_bot_token: notifyForm.value.telegram_bot_token,
      telegram_chat_id: notifyForm.value.telegram_chat_id,
      bark_url: notifyForm.value.bark_url
    }
    await settingsApi.update(payload)
    await settingsApi.testNotify(type)
    message.success('测试消息已发送')
  } catch (error: any) {
    message.error(error.response?.data?.error || '测试发送失败')
  } finally {
    if (type === 'telegram') testingTelegram.value = false
    else testingBark.value = false
  }
}

const changePassword = async (e: Event) => {
  e.preventDefault()
  pwdFormRef.value?.validate(async (errors) => {
    if (!errors) {
      changingPwd.value = true
      try {
        await authApi.changePassword({
          old_password: pwdForm.value.old_password,
          new_password: pwdForm.value.new_password
        })
        message.success('密码修改成功，请重新登录')
      } catch (error: any) {
        message.error(error.response?.data?.error || '密码修改失败')
      } finally {
        changingPwd.value = false
      }
    }
  })
}

onMounted(() => {
  fetchSettings()
})
</script>

<style scoped>
.page-container {
  max-width: 800px;
  margin: 0 auto;
}

.page-header {
  margin-bottom: 24px;
}

.page-title {
  font-size: 1.8rem;
  font-weight: 700;
  margin: 0;
  color: #1A1A1A;
}

.page-subtitle {
  color: #6B7280;
  margin-top: 4px;
}

.settings-card {
  box-shadow: 0 1px 3px 0 rgba(0, 0, 0, 0.1), 0 1px 2px 0 rgba(0, 0, 0, 0.06);
}
</style>
