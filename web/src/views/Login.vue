<template>
  <main class="login-page">
    <section class="login-showcase" aria-labelledby="login-intro-title">
      <app-brand :title="websiteTitle" />

      <div class="showcase-copy">
        <p class="showcase-eyebrow">Subscription clarity</p>
        <h1 id="login-intro-title">把下一次续费，<br>提前变成确定的事。</h1>
        <p>集中查看到期日、周期费用与提醒状态，让每一项长期服务都处在掌控之中。</p>
      </div>

      <div class="preview-card" aria-hidden="true">
        <div class="preview-card__header">
          <span>本月到期</span>
          <span class="preview-pill">3 项</span>
        </div>
        <div class="preview-item">
          <span class="preview-icon preview-icon--blue"><n-icon :component="CloudOutline" /></span>
          <span><strong>云存储服务</strong><small>还有 4 天</small></span>
          <b>¥ 38</b>
        </div>
        <div class="preview-item">
          <span class="preview-icon preview-icon--amber"><n-icon :component="MusicalNotesOutline" /></span>
          <span><strong>音乐会员</strong><small>还有 12 天</small></span>
          <b>¥ 15</b>
        </div>
        <div class="preview-progress"><span /></div>
      </div>

      <ul class="showcase-points">
        <li><n-icon :component="CheckmarkCircleOutline" /> 到期状态一目了然</li>
        <li><n-icon :component="CheckmarkCircleOutline" /> Telegram 与 Bark 主动提醒</li>
        <li><n-icon :component="CheckmarkCircleOutline" /> 数据保留在自己的设备中</li>
      </ul>
    </section>

    <section class="login-panel" aria-labelledby="login-form-title">
      <div class="login-panel__inner">
        <div class="mobile-brand">
          <app-brand :title="websiteTitle" />
        </div>

        <div class="login-heading">
          <span class="login-heading__icon" aria-hidden="true">
            <n-icon :component="LockClosedOutline" />
          </span>
          <div>
            <p>欢迎回来</p>
            <h2 id="login-form-title">登录管理控制台</h2>
          </div>
        </div>

        <n-form
          ref="formRef"
          :model="formValue"
          :rules="rules"
          label-placement="top"
          size="large"
          @submit.prevent="handleLogin"
        >
          <n-form-item path="username" label="用户名">
            <n-input
              v-model:value="formValue.username"
              autocomplete="username"
              placeholder="请输入管理员用户名"
            >
              <template #prefix>
                <n-icon :component="PersonOutline" />
              </template>
            </n-input>
          </n-form-item>

          <n-form-item path="password" label="密码">
            <n-input
              v-model:value="formValue.password"
              type="password"
              autocomplete="current-password"
              show-password-on="click"
              placeholder="请输入密码"
            >
              <template #prefix>
                <n-icon :component="KeyOutline" />
              </template>
            </n-input>
          </n-form-item>

          <n-button attr-type="submit" type="primary" block size="large" :loading="loading">
            登录 SubDock
            <template #icon>
              <n-icon :component="ArrowForwardOutline" />
            </template>
          </n-button>
        </n-form>

        <p class="login-hint">
          首次启动时，请从服务日志中获取随机生成的管理员密码。
        </p>
      </div>
    </section>
  </main>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  NButton,
  NForm,
  NFormItem,
  NIcon,
  NInput,
  useMessage
} from 'naive-ui'
import type { FormInst, FormRules } from 'naive-ui'
import {
  ArrowForwardOutline,
  CheckmarkCircleOutline,
  CloudOutline,
  KeyOutline,
  LockClosedOutline,
  MusicalNotesOutline,
  PersonOutline
} from '@vicons/ionicons5'

import AppBrand from '../components/app/AppBrand.vue'
import { authApi, getApiError } from '../api'
import { useAuthStore } from '../stores/auth'
import { useConfigStore } from '../stores/config'

const route = useRoute()
const router = useRouter()
const message = useMessage()
const authStore = useAuthStore()
const configStore = useConfigStore()

const websiteTitle = computed(() => configStore.websiteTitle)
const formRef = ref<FormInst | null>(null)
const loading = ref(false)
const formValue = ref({ username: '', password: '' })

const rules: FormRules = {
  username: { required: true, message: '请输入用户名', trigger: ['blur', 'input'] },
  password: { required: true, message: '请输入密码', trigger: ['blur', 'input'] }
}

/** 校验登录表单、保存令牌并跳转到原始目标页。 */
const handleLogin = async () => {
  try {
    await formRef.value?.validate()
  } catch {
    return
  }

  loading.value = true
  try {
    const response = await authApi.login(formValue.value)
    authStore.setToken(response.data.token)
    message.success('登录成功')

    const redirect = typeof route.query.redirect === 'string' ? route.query.redirect : '/dashboard'
    await router.replace(redirect)
  } catch (error: unknown) {
    message.error(getApiError(error, '登录失败，请检查用户名和密码'))
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-page {
  display: grid;
  min-height: 100vh;
  grid-template-columns: minmax(0, 1.08fr) minmax(420px, 0.92fr);
  background: #fff;
}

.login-showcase {
  position: relative;
  display: flex;
  min-height: 100vh;
  padding: clamp(34px, 5vw, 68px);
  overflow: hidden;
  flex-direction: column;
  color: #17203b;
  background:
    linear-gradient(rgba(255, 255, 255, 0.62), rgba(255, 255, 255, 0.62)),
    repeating-linear-gradient(0deg, transparent 0 39px, rgba(54, 89, 227, 0.055) 39px 40px),
    repeating-linear-gradient(90deg, transparent 0 39px, rgba(54, 89, 227, 0.055) 39px 40px),
    linear-gradient(145deg, #f4f7ff 0%, #e9eeff 54%, #f7f8fc 100%);
}

.login-showcase::before,
.login-showcase::after {
  position: absolute;
  content: '';
  border-radius: 50%;
  pointer-events: none;
}

.login-showcase::before {
  top: 12%;
  right: -110px;
  width: 360px;
  height: 360px;
  border: 70px solid rgba(54, 89, 227, 0.07);
}

.login-showcase::after {
  bottom: -160px;
  left: -120px;
  width: 360px;
  height: 360px;
  background: rgba(109, 137, 239, 0.08);
  filter: blur(2px);
}

.showcase-copy {
  position: relative;
  z-index: 1;
  max-width: 670px;
  margin-top: auto;
}

.showcase-eyebrow {
  margin: 0 0 16px;
  color: var(--color-primary);
  font-size: 12px;
  font-weight: 750;
  letter-spacing: 0.16em;
  text-transform: uppercase;
}

.showcase-copy h1 {
  margin: 0;
  font-size: clamp(38px, 5vw, 64px);
  font-weight: 780;
  letter-spacing: -0.055em;
  line-height: 1.12;
}

.showcase-copy > p:last-child {
  max-width: 570px;
  margin: 22px 0 0;
  color: #57647b;
  font-size: clamp(15px, 1.5vw, 18px);
  line-height: 1.8;
}

.preview-card {
  position: relative;
  z-index: 1;
  width: min(100%, 490px);
  margin: 42px 0 0 auto;
  padding: 20px;
  border: 1px solid rgba(255, 255, 255, 0.9);
  border-radius: 20px;
  background: rgba(255, 255, 255, 0.8);
  box-shadow: 0 26px 60px rgba(61, 78, 131, 0.17);
  backdrop-filter: blur(18px);
  transform: rotate(-1.2deg);
}

.preview-card__header,
.preview-item {
  display: flex;
  align-items: center;
}

.preview-card__header {
  justify-content: space-between;
  color: #3e4960;
  font-size: 12px;
  font-weight: 700;
}

.preview-pill {
  padding: 5px 9px;
  border-radius: 999px;
  color: #3450b6;
  background: #edf1ff;
}

.preview-item {
  gap: 12px;
  margin-top: 16px;
}

.preview-icon {
  display: grid;
  width: 38px;
  height: 38px;
  flex: 0 0 auto;
  border-radius: 12px;
  place-items: center;
  font-size: 18px;
}

.preview-icon--blue {
  color: #3659e3;
  background: #eef2ff;
}

.preview-icon--amber {
  color: #b85f00;
  background: #fff2df;
}

.preview-item > span:nth-child(2) {
  display: grid;
  min-width: 0;
  flex: 1;
}

.preview-item strong {
  color: #253049;
  font-size: 13px;
}

.preview-item small {
  margin-top: 3px;
  color: #798399;
  font-size: 10px;
}

.preview-item b {
  font-size: 13px;
  font-variant-numeric: tabular-nums;
}

.preview-progress {
  height: 5px;
  margin-top: 18px;
  overflow: hidden;
  border-radius: 999px;
  background: #e9edf5;
}

.preview-progress span {
  display: block;
  width: 68%;
  height: 100%;
  border-radius: inherit;
  background: linear-gradient(90deg, #3659e3, #7a95f4);
}

.showcase-points {
  position: relative;
  z-index: 1;
  display: flex;
  margin: auto 0 0;
  padding: 44px 0 0;
  flex-wrap: wrap;
  gap: 12px 24px;
  list-style: none;
}

.showcase-points li {
  display: flex;
  align-items: center;
  gap: 7px;
  color: #59667c;
  font-size: 12px;
  font-weight: 600;
}

.showcase-points .n-icon {
  color: var(--color-primary);
  font-size: 16px;
}

.login-panel {
  display: grid;
  padding: 40px clamp(32px, 6vw, 86px);
  place-items: center;
  background: #fff;
}

.login-panel__inner {
  width: min(100%, 420px);
}

.mobile-brand {
  display: none;
}

.login-heading {
  display: flex;
  margin-bottom: 34px;
  align-items: center;
  gap: 14px;
}

.login-heading__icon {
  display: grid;
  width: 48px;
  height: 48px;
  flex: 0 0 auto;
  border-radius: 14px;
  place-items: center;
  color: var(--color-primary);
  background: var(--color-primary-soft);
  font-size: 22px;
}

.login-heading p {
  margin: 0 0 4px;
  color: var(--color-text-muted);
  font-size: 12px;
  font-weight: 650;
}

.login-heading h2 {
  margin: 0;
  color: var(--color-text);
  font-size: 25px;
  font-weight: 760;
  letter-spacing: -0.03em;
}

.login-hint {
  margin: 22px 0 0;
  padding-top: 20px;
  border-top: 1px solid var(--color-border);
  color: var(--color-text-muted);
  font-size: 12px;
  line-height: 1.7;
  text-align: center;
}

@media (max-width: 980px) {
  .login-page {
    grid-template-columns: 1fr;
    background:
      radial-gradient(circle at 100% 0, rgba(89, 121, 237, 0.16), transparent 22rem),
      #f7f8fc;
  }

  .login-showcase {
    display: none;
  }

  .login-panel {
    min-height: 100vh;
    padding: 34px 22px;
    background: transparent;
  }

  .login-panel__inner {
    padding: 28px;
    border: 1px solid rgba(224, 229, 239, 0.92);
    border-radius: 22px;
    background: rgba(255, 255, 255, 0.94);
    box-shadow: var(--shadow-float);
  }

  .mobile-brand {
    display: block;
    margin-bottom: 36px;
  }
}

@media (max-width: 480px) {
  .login-panel {
    padding: 18px 14px;
  }

  .login-panel__inner {
    padding: 24px 20px;
  }

  .login-heading h2 {
    font-size: 22px;
  }
}
</style>
