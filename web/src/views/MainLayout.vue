<template>
  <div class="app-shell">
    <a class="skip-link" href="#main-content">跳到主要内容</a>

    <aside class="sidebar" aria-label="主导航">
      <div class="sidebar__brand">
        <app-brand :title="websiteTitle" />
      </div>

      <div class="sidebar__date">
        <span>{{ currentDate.weekday }}</span>
        <strong>{{ currentDate.date }}</strong>
      </div>

      <nav class="sidebar__nav">
        <router-link v-for="item in navItems" :key="item.to" :to="item.to" class="nav-link">
          <span class="nav-link__icon" aria-hidden="true">
            <n-icon :component="item.icon" />
          </span>
          <span>{{ item.label }}</span>
        </router-link>
      </nav>

      <div class="sidebar__spacer" />

      <router-link class="quick-add" :to="{ path: '/subscriptions', query: { create: '1' } }">
        <n-icon :component="AddOutline" />
        <span>添加新订阅</span>
      </router-link>

      <div class="sidebar__account">
        <span class="account-avatar" aria-hidden="true">A</span>
        <span class="account-copy">
          <strong>管理员</strong>
          <small>本地账户</small>
        </span>
        <button class="logout-button" type="button" aria-label="退出登录" @click="handleLogout">
          <n-icon :component="LogOutOutline" />
        </button>
      </div>
    </aside>

    <header class="mobile-header">
      <app-brand :title="websiteTitle" compact />
      <router-link class="mobile-settings" to="/settings" aria-label="打开设置">
        <n-icon :component="SettingsOutline" />
      </router-link>
    </header>

    <div class="workspace">
      <main id="main-content" class="main-content" tabindex="-1">
        <router-view v-slot="{ Component, route: currentRoute }">
          <transition name="page-fade" mode="out-in">
            <component :is="Component" :key="currentRoute.path" />
          </transition>
        </router-view>
      </main>
    </div>

    <nav class="mobile-nav" aria-label="移动端主导航">
      <router-link v-for="item in navItems" :key="item.to" :to="item.to" class="mobile-nav__item">
        <n-icon :component="item.icon" />
        <span>{{ item.shortLabel }}</span>
      </router-link>
    </nav>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { NIcon, useMessage } from 'naive-ui'
import {
  AddOutline,
  AppsOutline,
  CardOutline,
  LogOutOutline,
  SettingsOutline
} from '@vicons/ionicons5'

import AppBrand from '../components/app/AppBrand.vue'
import { useAuthStore } from '../stores/auth'
import { useConfigStore } from '../stores/config'

const router = useRouter()
const message = useMessage()
const authStore = useAuthStore()
const configStore = useConfigStore()

const websiteTitle = computed(() => configStore.websiteTitle)

const navItems = [
  { to: '/dashboard', label: '概览', shortLabel: '概览', icon: AppsOutline },
  { to: '/subscriptions', label: '订阅管理', shortLabel: '订阅', icon: CardOutline },
  { to: '/settings', label: '系统设置', shortLabel: '设置', icon: SettingsOutline }
]

const currentDate = computed(() => {
  const now = new Date()
  return {
    weekday: new Intl.DateTimeFormat('zh-CN', { weekday: 'long' }).format(now),
    date: new Intl.DateTimeFormat('zh-CN', { month: 'long', day: 'numeric' }).format(now)
  }
})

/** 清理本地登录状态并返回登录页。 */
const handleLogout = async () => {
  authStore.logout()
  message.success('已安全退出')
  await router.push('/login')
}
</script>

<style scoped>
.app-shell {
  min-height: 100vh;
}

.sidebar {
  position: fixed;
  inset: 18px auto 18px 18px;
  z-index: 30;
  display: flex;
  width: 224px;
  padding: 22px 16px 16px;
  border: 1px solid rgba(224, 229, 239, 0.9);
  border-radius: 22px;
  flex-direction: column;
  background: rgba(255, 255, 255, 0.88);
  box-shadow: 0 18px 48px rgba(35, 47, 80, 0.09);
  backdrop-filter: blur(18px);
}

.sidebar__brand {
  padding: 0 6px;
}

.sidebar__date {
  display: grid;
  margin: 28px 6px 20px;
  padding: 14px 15px;
  border: 1px solid #e5e9f2;
  border-radius: 14px;
  background: linear-gradient(135deg, #fafbff 0%, #f2f5ff 100%);
}

.sidebar__date span {
  color: var(--color-text-muted);
  font-size: 11px;
  font-weight: 650;
}

.sidebar__date strong {
  margin-top: 3px;
  color: var(--color-text);
  font-size: 15px;
  font-weight: 720;
}

.sidebar__nav {
  display: grid;
  gap: 6px;
}

.nav-link {
  display: flex;
  min-height: 46px;
  padding: 0 12px;
  border-radius: 12px;
  align-items: center;
  gap: 11px;
  color: #5c6679;
  font-size: 14px;
  font-weight: 620;
  text-decoration: none;
  transition: color 160ms ease, background-color 160ms ease;
}

.nav-link:hover {
  color: var(--color-text);
  background: #f5f7fb;
}

.nav-link.router-link-active {
  color: var(--color-primary-strong);
  background: var(--color-primary-soft);
}

.nav-link__icon {
  display: grid;
  width: 24px;
  height: 24px;
  place-items: center;
  font-size: 19px;
}

.sidebar__spacer {
  flex: 1;
  min-height: 28px;
}

.quick-add {
  display: flex;
  min-height: 44px;
  margin: 0 2px 14px;
  border-radius: 12px;
  align-items: center;
  justify-content: center;
  gap: 8px;
  color: #fff;
  background: var(--color-primary);
  box-shadow: 0 9px 20px rgba(54, 89, 227, 0.2);
  font-size: 13px;
  font-weight: 680;
  text-decoration: none;
  transition: background-color 160ms ease, box-shadow 160ms ease;
}

.quick-add:hover {
  background: var(--color-primary-strong);
  box-shadow: 0 11px 24px rgba(54, 89, 227, 0.27);
}

.sidebar__account {
  display: flex;
  min-width: 0;
  padding: 13px 10px 4px;
  border-top: 1px solid var(--color-border);
  align-items: center;
  gap: 10px;
}

.account-avatar {
  display: grid;
  width: 34px;
  height: 34px;
  flex: 0 0 auto;
  border-radius: 11px;
  place-items: center;
  color: #334aa0;
  background: #e9edff;
  font-size: 13px;
  font-weight: 750;
}

.account-copy {
  display: grid;
  min-width: 0;
  flex: 1;
  line-height: 1.25;
}

.account-copy strong {
  color: var(--color-text);
  font-size: 12px;
}

.account-copy small {
  margin-top: 2px;
  color: var(--color-text-muted);
  font-size: 10px;
}

.logout-button,
.mobile-settings {
  display: grid;
  width: 36px;
  height: 36px;
  border: 0;
  border-radius: 10px;
  place-items: center;
  color: #7c8698;
  background: transparent;
  font-size: 18px;
  text-decoration: none;
  transition: color 160ms ease, background-color 160ms ease;
}

.logout-button:hover {
  color: var(--color-danger);
  background: var(--color-danger-soft);
}

.workspace {
  min-height: 100vh;
  padding-left: 260px;
}

.main-content {
  width: 100%;
  min-height: 100vh;
  padding: 40px clamp(24px, 4vw, 56px) 48px;
}

.mobile-header,
.mobile-nav {
  display: none;
}

@media (max-width: 900px) {
  .sidebar {
    display: none;
  }

  .workspace {
    padding-left: 0;
  }

  .mobile-header {
    position: sticky;
    top: 0;
    z-index: 25;
    display: flex;
    min-height: 68px;
    padding: 12px 18px;
    border-bottom: 1px solid rgba(224, 229, 239, 0.92);
    align-items: center;
    justify-content: space-between;
    background: rgba(255, 255, 255, 0.9);
    backdrop-filter: blur(16px);
  }

  .main-content {
    padding: 26px 20px 104px;
  }

  .mobile-nav {
    position: fixed;
    right: 14px;
    bottom: max(14px, env(safe-area-inset-bottom));
    left: 14px;
    z-index: 40;
    display: grid;
    min-height: 64px;
    padding: 7px;
    border: 1px solid rgba(217, 223, 235, 0.95);
    border-radius: 19px;
    grid-template-columns: repeat(3, 1fr);
    background: rgba(255, 255, 255, 0.94);
    box-shadow: 0 16px 42px rgba(32, 43, 74, 0.17);
    backdrop-filter: blur(18px);
  }

  .mobile-nav__item {
    display: flex;
    min-height: 48px;
    border-radius: 13px;
    align-items: center;
    justify-content: center;
    flex-direction: column;
    gap: 3px;
    color: #778195;
    font-size: 18px;
    text-decoration: none;
  }

  .mobile-nav__item span {
    font-size: 10px;
    font-weight: 650;
  }

  .mobile-nav__item.router-link-active {
    color: var(--color-primary-strong);
    background: var(--color-primary-soft);
  }
}

@media (max-width: 520px) {
  .main-content {
    padding-right: 16px;
    padding-left: 16px;
  }
}
</style>
