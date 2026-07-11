<template>
  <span class="status-pill" :class="`status-pill--${status.kind}`">
    <span class="status-pill__dot" aria-hidden="true" />
    {{ status.label }}
  </span>
</template>

<script setup lang="ts">
import type { SubscriptionStatus } from '../../utils/subscription'

defineProps<{
  status: SubscriptionStatus
}>()
</script>

<style scoped>
.status-pill {
  display: inline-flex;
  min-height: 26px;
  padding: 4px 9px;
  border-radius: 999px;
  align-items: center;
  gap: 6px;
  color: var(--status-color);
  background: var(--status-bg);
  font-size: 11px;
  font-weight: 680;
  white-space: nowrap;
}

.status-pill__dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: currentColor;
}

.status-pill--active {
  --status-color: #087f5b;
  --status-bg: #eaf8f3;
}

.status-pill--due_soon,
.status-pill--due_today {
  --status-color: #a85600;
  --status-bg: #fff2de;
}

.status-pill--expired {
  --status-color: #b72d27;
  --status-bg: #fff0ef;
}

.status-pill--unknown {
  --status-color: #687386;
  --status-bg: #f0f2f6;
}
</style>
