<template>
  <article class="subscription-card surface-card">
    <button class="subscription-card__main" type="button" @click="emit('view', subscription)">
      <span class="subscription-avatar" aria-hidden="true">{{ initial }}</span>
      <span class="subscription-copy">
        <strong>{{ subscription.name }}</strong>
        <small>{{ getCycleText(subscription.cycle_value, subscription.cycle_unit) }}</small>
      </span>
      <span class="subscription-price numeric">
        <strong>{{ formatCurrency(subscription.amount, subscription.currency) }}</strong>
        <small>{{ subscription.currency }}</small>
      </span>
    </button>

    <div class="subscription-card__meta">
      <div>
        <span>到期日期</span>
        <strong class="numeric">{{ formatDate(subscription.expire_date) }}</strong>
      </div>
      <div>
        <span>剩余时间</span>
        <strong>{{ remainingCopy }}</strong>
      </div>
      <subscription-status-pill :status="status" />
    </div>

    <div class="subscription-card__footer">
      <span class="renew-label">
        <n-icon :component="subscription.auto_renew ? SyncOutline : HandLeftOutline" />
        {{ subscription.auto_renew ? '自动续订' : '手动续订' }} · 已续 {{ subscription.renew_count ?? 0 }} 次
      </span>
      <n-button
        size="small"
        secondary
        type="primary"
        :loading="renewing"
        @click="emit('renew', subscription)"
      >
        续订
      </n-button>
    </div>
  </article>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { NButton, NIcon } from 'naive-ui'
import { HandLeftOutline, SyncOutline } from '@vicons/ionicons5'

import type { SubscriptionRecord } from '../../api'
import { formatCurrency } from '../../utils/currency'
import { parseLocalDate } from '../../utils/date'
import { getCycleText, getSubscriptionStatus } from '../../utils/subscription'
import SubscriptionStatusPill from './SubscriptionStatusPill.vue'

const props = defineProps<{
  subscription: SubscriptionRecord
  renewing: boolean
}>()

const emit = defineEmits<{
  view: [subscription: SubscriptionRecord]
  renew: [subscription: SubscriptionRecord]
}>()

const status = computed(() => getSubscriptionStatus(props.subscription))
const initial = computed(() => props.subscription.name.trim().charAt(0).toUpperCase() || 'S')
const remainingCopy = computed(() => {
  const days = status.value.daysRemaining
  if (days === null) return '未知'
  if (days < 0) return `逾期 ${Math.abs(days)} 天`
  if (days === 0) return '今天'
  return `${days} 天`
})

/** 格式化订阅到期日期。 */
const formatDate = (value: string | null) => {
  const date = parseLocalDate(value)
  if (!date) return '未知'
  return new Intl.DateTimeFormat('zh-CN', { month: '2-digit', day: '2-digit', year: 'numeric' }).format(date)
}
</script>

<style scoped>
.subscription-card {
  padding: 17px;
}

.subscription-card__main {
  display: grid;
  width: 100%;
  padding: 0 0 15px;
  border: 0;
  border-bottom: 1px solid #edf0f5;
  grid-template-columns: 42px minmax(0, 1fr) auto;
  align-items: center;
  gap: 11px;
  color: inherit;
  background: transparent;
  text-align: left;
}

.subscription-avatar {
  display: grid;
  width: 42px;
  height: 42px;
  border-radius: 13px;
  place-items: center;
  color: #3550ae;
  background: var(--color-primary-soft);
  font-size: 14px;
  font-weight: 780;
}

.subscription-copy,
.subscription-price {
  display: grid;
  min-width: 0;
}

.subscription-copy strong {
  overflow: hidden;
  color: var(--color-text);
  font-size: 13px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.subscription-copy small,
.subscription-price small {
  margin-top: 4px;
  color: var(--color-text-muted);
  font-size: 9px;
}

.subscription-price {
  text-align: right;
}

.subscription-price strong {
  font-size: 13px;
}

.subscription-card__meta {
  display: grid;
  padding: 15px 0;
  grid-template-columns: repeat(2, minmax(0, 1fr)) auto;
  align-items: end;
  gap: 12px;
}

.subscription-card__meta > div {
  display: grid;
}

.subscription-card__meta span:not(.status-pill) {
  color: var(--color-text-muted);
  font-size: 9px;
}

.subscription-card__meta strong {
  margin-top: 4px;
  color: var(--color-text-secondary);
  font-size: 11px;
}

.subscription-card__footer {
  display: flex;
  padding-top: 13px;
  border-top: 1px solid #edf0f5;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.renew-label {
  display: flex;
  align-items: center;
  gap: 6px;
  color: var(--color-text-muted);
  font-size: 9px;
}

.renew-label .n-icon {
  color: var(--color-primary);
  font-size: 14px;
}

@media (max-width: 380px) {
  .subscription-card__meta {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .subscription-card__meta .status-pill {
    grid-column: 1 / -1;
    justify-self: start;
  }
}
</style>
