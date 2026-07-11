<template>
  <div class="brand" :class="{ 'brand--compact': compact }">
    <span class="brand__mark" aria-hidden="true">
      <n-icon :component="AlbumsOutline" />
    </span>
    <span class="brand__copy">
      <strong>{{ title }}</strong>
      <small v-if="!compact">让每一笔续费都有数</small>
    </span>
  </div>
</template>

<script setup lang="ts">
import { NIcon } from 'naive-ui'
import { AlbumsOutline } from '@vicons/ionicons5'

withDefaults(defineProps<{
  title: string
  compact?: boolean
}>(), {
  compact: false
})
</script>

<style scoped>
.brand {
  display: inline-flex;
  align-items: center;
  gap: 12px;
  min-width: 0;
}

.brand__mark {
  display: grid;
  width: 42px;
  height: 42px;
  flex: 0 0 auto;
  place-items: center;
  border-radius: 13px;
  color: #fff;
  background: linear-gradient(145deg, #5575ef 0%, #2948bd 100%);
  box-shadow: 0 10px 22px rgba(54, 89, 227, 0.24);
  font-size: 22px;
}

.brand__copy {
  display: grid;
  min-width: 0;
  line-height: 1.2;
}

.brand__copy strong {
  overflow: hidden;
  color: var(--color-text);
  font-size: 18px;
  font-weight: 750;
  letter-spacing: -0.02em;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.brand__copy small {
  margin-top: 4px;
  color: var(--color-text-muted);
  font-size: 11px;
  letter-spacing: 0.06em;
}

.brand--compact .brand__mark {
  width: 38px;
  height: 38px;
  border-radius: 12px;
}

.brand--compact .brand__copy strong {
  font-size: 16px;
}
</style>
