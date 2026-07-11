<template>
  <div class="empty-state">
    <span class="empty-state__icon" aria-hidden="true">
      <n-icon :component="icon" />
    </span>
    <h3>{{ title }}</h3>
    <p>{{ description }}</p>
    <n-button v-if="actionText" type="primary" @click="$emit('action')">
      {{ actionText }}
    </n-button>
  </div>
</template>

<script setup lang="ts">
import type { Component } from 'vue'
import { NButton, NIcon } from 'naive-ui'

defineProps<{
  icon: Component
  title: string
  description: string
  actionText?: string
}>()

defineEmits<{
  action: []
}>()
</script>

<style scoped>
.empty-state {
  display: grid;
  min-height: 280px;
  padding: 36px 24px;
  place-items: center;
  align-content: center;
  text-align: center;
}

.empty-state__icon {
  display: grid;
  width: 62px;
  height: 62px;
  margin-bottom: 18px;
  border-radius: 20px;
  place-items: center;
  color: var(--color-primary);
  background: var(--color-primary-soft);
  font-size: 30px;
}

.empty-state h3 {
  margin: 0;
  color: var(--color-text);
  font-size: 18px;
}

.empty-state p {
  max-width: 390px;
  margin: 9px 0 20px;
  color: var(--color-text-muted);
  font-size: 13px;
  line-height: 1.7;
}
</style>
