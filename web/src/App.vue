<template>
  <n-config-provider :theme-overrides="themeOverrides" :locale="zhCN" :date-locale="dateZhCN">
    <n-global-style />
    <n-message-provider>
      <n-dialog-provider>
        <router-view />
      </n-dialog-provider>
    </n-message-provider>
  </n-config-provider>
</template>

<script setup lang="ts">
import { NConfigProvider, NGlobalStyle, NMessageProvider, NDialogProvider, zhCN, dateZhCN } from 'naive-ui'
import type { GlobalThemeOverrides } from 'naive-ui'

const themeOverrides: GlobalThemeOverrides = {
  common: {
    primaryColor: '#1A1A1A',
    primaryColorHover: '#333333',
    primaryColorPressed: '#000000',
    primaryColorSuppl: '#333333',
    infoColor: '#1A1A1A',
    successColor: '#10B981',
    warningColor: '#F59E0B',
    errorColor: '#EF4444',
    bodyColor: '#F3F4F6',
    textColorBase: '#1A1A1A',
    textColor1: '#1A1A1A',
    textColor2: '#374151',
    textColor3: '#6B7280',
    borderRadius: '6px'
  },
  Button: {
    borderRadiusMedium: '6px',
    heightMedium: '40px',
    textColorPrimary: '#FFFFFF',
    colorFocusPrimary: '#333333',
  },
  Input: {
    borderRadius: '6px',
    heightMedium: '40px',
  },
  Card: {
    borderRadius: '12px',
  },
  DataTable: {
    borderRadius: '8px'
  }
}
</script>
