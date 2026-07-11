<template>
  <article class="metric-card surface-card" :class="`metric-card--${tone}`">
    <span class="metric-card__icon" aria-hidden="true">
      <n-icon :component="icon" />
    </span>
    <div class="metric-card__copy">
      <span>{{ label }}</span>
      <strong class="numeric">{{ value }}</strong>
      <small>{{ detail }}</small>
    </div>
  </article>
</template>

<script setup lang="ts">
import type { Component } from 'vue'
import { NIcon } from 'naive-ui'

withDefaults(defineProps<{
  label: string
  value: string | number
  detail: string
  icon: Component
  tone?: 'primary' | 'success' | 'warning' | 'danger'
}>(), {
  tone: 'primary'
})
</script>

<style scoped>
.metric-card {
  position: relative;
  display: flex;
  min-height: 146px;
  padding: 22px;
  overflow: hidden;
  align-items: flex-start;
  gap: 15px;
}

.metric-card::after {
  position: absolute;
  top: -45px;
  right: -42px;
  width: 120px;
  height: 120px;
  border-radius: 50%;
  content: '';
  background: var(--metric-soft);
}

.metric-card__icon {
  position: relative;
  z-index: 1;
  display: grid;
  width: 42px;
  height: 42px;
  flex: 0 0 auto;
  border-radius: 13px;
  place-items: center;
  color: var(--metric-color);
  background: var(--metric-soft);
  font-size: 20px;
}

.metric-card__copy {
  position: relative;
  z-index: 1;
  display: grid;
  min-width: 0;
}

.metric-card__copy > span {
  color: var(--color-text-muted);
  font-size: 12px;
  font-weight: 650;
}

.metric-card__copy strong {
  margin-top: 8px;
  color: var(--color-text);
  font-size: 30px;
  font-weight: 780;
  letter-spacing: -0.04em;
  line-height: 1;
}

.metric-card__copy small {
  margin-top: 10px;
  color: var(--color-text-muted);
  font-size: 11px;
  line-height: 1.45;
}

.metric-card--primary {
  --metric-color: #3659e3;
  --metric-soft: #eef2ff;
}

.metric-card--success {
  --metric-color: #087f5b;
  --metric-soft: #eaf8f3;
}

.metric-card--warning {
  --metric-color: #b85f00;
  --metric-soft: #fff4e5;
}

.metric-card--danger {
  --metric-color: #c4322b;
  --metric-soft: #fff0ef;
}
</style>
