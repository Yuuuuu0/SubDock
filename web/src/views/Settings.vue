<template>
  <div class="page-container settings-page">
    <header class="page-header">
      <div class="page-header__copy">
        <p class="page-eyebrow">Preferences</p>
        <h1 class="page-title">设置</h1>
        <p class="page-subtitle">管理提醒时间、通知渠道与账户安全。敏感配置仅保存在你的 SubDock 数据库中。</p>
      </div>
      <n-button
        type="primary"
        :loading="savingSettings"
        :disabled="loadingSettings || !hasSettingsChanges"
        @click="saveSettings()"
      >
        <template #icon><n-icon :component="SaveOutline" /></template>
        保存更改
      </n-button>
    </header>

    <div v-if="loadingSettings" class="settings-loading surface-card" aria-label="正在加载设置">
      <n-skeleton text :repeat="2" />
      <n-skeleton height="120px" />
      <n-skeleton text :repeat="3" />
    </div>

    <template v-else>
      <n-alert v-if="settingsError" type="error" :bordered="false" class="settings-alert">
        <div class="settings-alert__content">
          <span>{{ settingsError }}</span>
          <n-button size="small" @click="fetchSettings">重新加载</n-button>
        </div>
      </n-alert>

      <div class="settings-layout">
        <section class="settings-main">
          <article class="settings-section surface-card">
            <div class="section-heading">
              <div>
                <span class="section-icon section-icon--blue" aria-hidden="true">
                  <n-icon :component="TimeOutline" />
                </span>
                <h2>提醒时间</h2>
                <p>系统每小时检查一次，并在所选整点发送符合条件的到期提醒。</p>
              </div>
              <span class="selection-count">已选 {{ notifyForm.notify_hours.length }} 个时段</span>
            </div>

            <n-form label-placement="top" class="settings-form">
              <n-form-item label="每天通知时段">
                <n-select
                  v-model:value="notifyForm.notify_hours"
                  multiple
                  filterable
                  clearable
                  :options="hourOptions"
                  placeholder="选择一个或多个整点"
                  :max-tag-count="'responsive'"
                />
              </n-form-item>
            </n-form>

            <p class="field-hint">至少选择一个时段；同一天选择多个时段会发送多次提醒。</p>
          </article>

          <article class="settings-section surface-card">
            <div class="section-heading">
              <div>
                <span class="section-icon section-icon--telegram" aria-hidden="true">
                  <n-icon :component="PaperPlaneOutline" />
                </span>
                <h2>Telegram</h2>
                <p>通过 Telegram Bot 向指定会话推送订阅到期消息。</p>
              </div>
              <span class="channel-status" :class="{ 'channel-status--ready': telegramConfigured }">
                <span />{{ telegramConfigured ? '已配置' : '未配置' }}
              </span>
            </div>

            <n-form label-placement="top" class="settings-form settings-form--grid">
              <n-form-item label="Bot Token">
                <n-input
                  v-model:value="notifyForm.telegram_bot_token"
                  type="password"
                  show-password-on="click"
                  autocomplete="off"
                  placeholder="例如 123456:ABC..."
                />
              </n-form-item>
              <n-form-item label="Chat ID">
                <n-input
                  v-model:value="notifyForm.telegram_chat_id"
                  autocomplete="off"
                  placeholder="接收通知的会话 ID"
                />
              </n-form-item>
            </n-form>

            <div class="section-actions">
              <p>测试会先保存当前页面中的全部通知配置。</p>
              <n-button :loading="testingTelegram" @click="testNotify('telegram')">
                <template #icon><n-icon :component="SendOutline" /></template>
                保存并测试
              </n-button>
            </div>
          </article>

          <article class="settings-section surface-card">
            <div class="section-heading">
              <div>
                <span class="section-icon section-icon--bark" aria-hidden="true">
                  <n-icon :component="NotificationsOutline" />
                </span>
                <h2>Bark</h2>
                <p>将提醒发送到安装了 Bark 的 iPhone，也支持兼容的自建服务。</p>
              </div>
              <span class="channel-status" :class="{ 'channel-status--ready': barkConfigured }">
                <span />{{ barkConfigured ? '已配置' : '未配置' }}
              </span>
            </div>

            <n-form label-placement="top" class="settings-form">
              <n-form-item label="Bark URL">
                <n-input
                  v-model:value="notifyForm.bark_url"
                  type="password"
                  show-password-on="click"
                  autocomplete="off"
                  placeholder="https://api.day.app/your-key"
                />
              </n-form-item>
            </n-form>

            <div class="section-actions">
              <p>清空地址并保存即可停用 Bark 通知。</p>
              <n-button :loading="testingBark" @click="testNotify('bark')">
                <template #icon><n-icon :component="SendOutline" /></template>
                保存并测试
              </n-button>
            </div>
          </article>
        </section>

        <aside class="settings-aside">
          <article class="security-card surface-card">
            <div class="security-card__top">
              <span class="security-shield" aria-hidden="true">
                <n-icon :component="ShieldCheckmarkOutline" />
              </span>
              <div>
                <h2>账户安全</h2>
                <p>定期更新管理员密码，避免与其他服务重复使用。</p>
              </div>
            </div>

            <n-form
              ref="passwordFormRef"
              :model="passwordForm"
              :rules="passwordRules"
              label-placement="top"
              @submit.prevent="changePassword"
            >
              <n-form-item label="当前密码" path="old_password">
                <n-input
                  v-model:value="passwordForm.old_password"
                  type="password"
                  show-password-on="click"
                  autocomplete="current-password"
                />
              </n-form-item>
              <n-form-item label="新密码" path="new_password">
                <n-input
                  v-model:value="passwordForm.new_password"
                  type="password"
                  show-password-on="click"
                  autocomplete="new-password"
                  placeholder="至少 8 位"
                />
              </n-form-item>
              <n-form-item label="确认新密码" path="confirm_password">
                <n-input
                  v-model:value="passwordForm.confirm_password"
                  type="password"
                  show-password-on="click"
                  autocomplete="new-password"
                />
              </n-form-item>
              <n-button attr-type="submit" block :loading="changingPassword">
                更新密码并重新登录
              </n-button>
            </n-form>
          </article>

          <article class="privacy-note">
            <n-icon :component="InformationCircleOutline" />
            <div>
              <strong>关于通知凭据</strong>
              <p>Bot Token、Chat ID 与 Bark URL 会保存在本地 SQLite 数据库。请同时保护好数据目录和备份文件。</p>
            </div>
          </article>
        </aside>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import {
  NAlert,
  NButton,
  NForm,
  NFormItem,
  NIcon,
  NInput,
  NSelect,
  NSkeleton,
  useMessage
} from 'naive-ui'
import type { FormInst, FormRules } from 'naive-ui'
import {
  InformationCircleOutline,
  NotificationsOutline,
  PaperPlaneOutline,
  SaveOutline,
  SendOutline,
  ShieldCheckmarkOutline,
  TimeOutline
} from '@vicons/ionicons5'

import { authApi, getApiError, settingsApi } from '../api'
import type { NotificationType, Settings } from '../api'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const message = useMessage()
const authStore = useAuthStore()

const notifyForm = ref<Settings>({
  notify_hours: [9],
  telegram_bot_token: '',
  telegram_chat_id: '',
  bark_url: ''
})
const savedSettingsSignature = ref('')
const loadingSettings = ref(true)
const savingSettings = ref(false)
const testingTelegram = ref(false)
const testingBark = ref(false)
const settingsError = ref<string | null>(null)

const passwordFormRef = ref<FormInst | null>(null)
const changingPassword = ref(false)
const passwordForm = ref({
  old_password: '',
  new_password: '',
  confirm_password: ''
})

const hourOptions = Array.from({ length: 24 }, (_, hour) => ({
  value: hour,
  label: `${String(hour).padStart(2, '0')}:00`
}))

const settingsSignature = computed(() => JSON.stringify({
  ...notifyForm.value,
  notify_hours: [...notifyForm.value.notify_hours].sort((left, right) => left - right)
}))
const hasSettingsChanges = computed(() => settingsSignature.value !== savedSettingsSignature.value)
const telegramConfigured = computed(() => (
  notifyForm.value.telegram_bot_token.trim() !== '' && notifyForm.value.telegram_chat_id.trim() !== ''
))
const barkConfigured = computed(() => notifyForm.value.bark_url.trim() !== '')

const passwordRules: FormRules = {
  old_password: { required: true, message: '请输入当前密码', trigger: ['blur', 'input'] },
  new_password: [
    { required: true, message: '请输入新密码', trigger: ['blur', 'input'] },
    { min: 8, message: '新密码至少需要 8 位', trigger: ['blur', 'input'] }
  ],
  confirm_password: {
    required: true,
    trigger: ['blur', 'input'],
    validator: (_rule, value: string) => {
      if (!value) return new Error('请再次输入新密码')
      if (value !== passwordForm.value.new_password) return new Error('两次输入的新密码不一致')
      return true
    }
  }
}

/** 从服务端读取设置，并把逗号分隔时段转换为数字数组。 */
const fetchSettings = async () => {
  loadingSettings.value = true
  settingsError.value = null
  try {
    const response = await settingsApi.get()
    const hours = response.data.notify_hours
      .split(',')
      .map((hour) => Number.parseInt(hour.trim(), 10))
      .filter((hour) => Number.isInteger(hour) && hour >= 0 && hour <= 23)

    notifyForm.value = {
      notify_hours: hours.length > 0 ? [...new Set(hours)].sort((left, right) => left - right) : [9],
      telegram_bot_token: response.data.telegram_bot_token,
      telegram_chat_id: response.data.telegram_chat_id,
      bark_url: response.data.bark_url
    }
    savedSettingsSignature.value = settingsSignature.value
  } catch (error: unknown) {
    settingsError.value = getApiError(error, '获取设置失败，请稍后重试')
  } finally {
    loadingSettings.value = false
  }
}

/** 保存当前通知设置；测试通知时可关闭独立成功提示。 */
const saveSettings = async (silent = false): Promise<boolean> => {
  if (notifyForm.value.notify_hours.length === 0) {
    message.warning('请至少选择一个通知时段')
    return false
  }

  savingSettings.value = true
  try {
    const hours = [...new Set(notifyForm.value.notify_hours)].sort((left, right) => left - right)
    await settingsApi.update({
      notify_hours: hours.join(','),
      telegram_bot_token: notifyForm.value.telegram_bot_token.trim(),
      telegram_chat_id: notifyForm.value.telegram_chat_id.trim(),
      bark_url: notifyForm.value.bark_url.trim()
    })
    notifyForm.value.notify_hours = hours
    savedSettingsSignature.value = settingsSignature.value
    if (!silent) message.success('通知设置已保存')
    return true
  } catch (error: unknown) {
    message.error(getApiError(error, '保存设置失败'))
    return false
  } finally {
    savingSettings.value = false
  }
}

/** 保存设置后向指定渠道发送一次测试消息。 */
const testNotify = async (type: NotificationType) => {
  if (type === 'telegram' && !telegramConfigured.value) {
    message.warning('请先填写完整的 Bot Token 和 Chat ID')
    return
  }
  if (type === 'bark' && !barkConfigured.value) {
    message.warning('请先填写 Bark URL')
    return
  }

  if (type === 'telegram') testingTelegram.value = true
  else testingBark.value = true

  try {
    if (!(await saveSettings(true))) return
    await settingsApi.testNotify(type)
    message.success(`${type === 'telegram' ? 'Telegram' : 'Bark'} 测试消息已发送`)
  } catch (error: unknown) {
    message.error(getApiError(error, '测试通知发送失败'))
  } finally {
    if (type === 'telegram') testingTelegram.value = false
    else testingBark.value = false
  }
}

/** 修改管理员密码，成功后清理当前令牌并要求重新登录。 */
const changePassword = async () => {
  try {
    await passwordFormRef.value?.validate()
  } catch {
    return
  }

  changingPassword.value = true
  try {
    await authApi.changePassword({
      old_password: passwordForm.value.old_password,
      new_password: passwordForm.value.new_password
    })
    authStore.logout()
    message.success('密码已更新，请使用新密码重新登录')
    await router.replace('/login')
  } catch (error: unknown) {
    message.error(getApiError(error, '密码修改失败'))
  } finally {
    changingPassword.value = false
  }
}

onMounted(fetchSettings)
</script>

<style scoped>
.settings-page {
  max-width: 1180px;
}

.settings-alert {
  margin-bottom: 20px;
}

.settings-alert__content {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

.settings-loading {
  display: grid;
  padding: 28px;
  gap: 22px;
}

.settings-layout {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 330px;
  gap: 22px;
  align-items: start;
}

.settings-main {
  display: grid;
  gap: 18px;
}

.settings-section,
.security-card {
  padding: 24px;
}

.section-heading > div:first-child {
  display: grid;
  grid-template-columns: 42px minmax(0, 1fr);
  column-gap: 13px;
}

.section-heading h2,
.section-heading p {
  grid-column: 2;
}

.section-icon {
  display: grid;
  width: 42px;
  height: 42px;
  grid-row: 1 / span 2;
  border-radius: 13px;
  place-items: center;
  font-size: 20px;
}

.section-icon--blue {
  color: var(--color-primary);
  background: var(--color-primary-soft);
}

.section-icon--telegram {
  color: #147fc2;
  background: #eaf6fd;
}

.section-icon--bark {
  color: #9b5b00;
  background: #fff4df;
}

.selection-count,
.channel-status {
  flex: 0 0 auto;
  padding: 6px 9px;
  border-radius: 999px;
  color: var(--color-text-muted);
  background: #f2f4f7;
  font-size: 11px;
  font-weight: 650;
}

.channel-status {
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.channel-status span {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #9aa3b2;
}

.channel-status--ready {
  color: var(--color-success);
  background: var(--color-success-soft);
}

.channel-status--ready span {
  background: var(--color-success);
}

.settings-form {
  margin-top: 22px;
}

.settings-form--grid {
  display: grid;
  grid-template-columns: minmax(0, 1.2fr) minmax(180px, 0.8fr);
  gap: 14px;
}

.field-hint,
.section-actions p {
  margin: 0;
  color: var(--color-text-muted);
  font-size: 11px;
  line-height: 1.6;
}

.section-actions {
  display: flex;
  padding-top: 18px;
  border-top: 1px solid var(--color-border);
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

.settings-aside {
  position: sticky;
  top: 30px;
  display: grid;
  gap: 16px;
}

.security-card__top {
  display: flex;
  margin-bottom: 24px;
  align-items: flex-start;
  gap: 13px;
}

.security-shield {
  display: grid;
  width: 44px;
  height: 44px;
  flex: 0 0 auto;
  border-radius: 14px;
  place-items: center;
  color: var(--color-success);
  background: var(--color-success-soft);
  font-size: 22px;
}

.security-card h2 {
  margin: 1px 0 5px;
  font-size: 17px;
}

.security-card p {
  margin: 0;
  color: var(--color-text-muted);
  font-size: 12px;
  line-height: 1.6;
}

.privacy-note {
  display: flex;
  padding: 18px;
  border: 1px solid #e1e6f0;
  border-radius: 16px;
  align-items: flex-start;
  gap: 11px;
  color: #5a6579;
  background: rgba(247, 249, 253, 0.9);
}

.privacy-note > .n-icon {
  flex: 0 0 auto;
  margin-top: 2px;
  color: var(--color-primary);
  font-size: 19px;
}

.privacy-note strong {
  color: var(--color-text-secondary);
  font-size: 12px;
}

.privacy-note p {
  margin: 5px 0 0;
  font-size: 11px;
  line-height: 1.65;
}

@media (max-width: 1100px) {
  .settings-layout {
    grid-template-columns: 1fr;
  }

  .settings-aside {
    position: static;
    grid-template-columns: minmax(0, 1fr) minmax(260px, 0.65fr);
  }
}

@media (max-width: 720px) {
  .page-header .n-button {
    width: 100%;
  }

  .settings-section,
  .security-card {
    padding: 20px;
  }

  .settings-form--grid,
  .settings-aside {
    grid-template-columns: 1fr;
  }

  .section-actions {
    align-items: stretch;
    flex-direction: column;
  }

  .section-actions .n-button {
    width: 100%;
  }
}

@media (max-width: 480px) {
  .section-heading {
    flex-direction: column;
  }
}
</style>
