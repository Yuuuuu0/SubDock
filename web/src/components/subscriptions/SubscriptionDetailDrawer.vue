<template>
  <n-drawer
    :show="show"
    width="min(100vw, 520px)"
    placement="right"
    @update:show="emit('update:show', $event)"
  >
    <n-drawer-content v-if="subscription" :native-scrollbar="false" closable>
      <template #header>
        <div class="detail-header">
          <span class="detail-header__avatar" aria-hidden="true">{{ subscriptionInitial }}</span>
          <div>
            <small>Subscription detail</small>
            <strong>{{ subscription.name }}</strong>
          </div>
        </div>
      </template>

      <section class="detail-hero">
        <div>
          <subscription-status-pill :status="status" />
          <p>{{ getStatusDescription(status.daysRemaining) }}</p>
        </div>
        <div class="detail-amount numeric">
          <strong>{{ formatCurrency(subscription.amount, subscription.currency) }}</strong>
          <small>/ {{ getCycleText(subscription.cycle_value, subscription.cycle_unit) }}</small>
        </div>
      </section>

      <section class="detail-grid" aria-label="订阅详细信息">
        <div class="detail-field">
          <span><n-icon :component="CalendarOutline" />开始日期</span>
          <strong class="numeric">{{ formatDate(subscription.start_date) }}</strong>
        </div>
        <div class="detail-field">
          <span><n-icon :component="FlagOutline" />到期日期</span>
          <strong class="numeric">{{ formatDate(subscription.expire_date) }}</strong>
        </div>
        <div class="detail-field">
          <span><n-icon :component="RefreshOutline" />续订方式</span>
          <strong>{{ subscription.auto_renew ? '自动续订' : '手动续订' }}</strong>
        </div>
        <div class="detail-field">
          <span><n-icon :component="NotificationsOutline" />提前提醒</span>
          <strong>{{ subscription.remind_days === 0 ? '到期当天' : `${subscription.remind_days} 天` }}</strong>
        </div>
      </section>

      <section v-if="subscription.remark" class="detail-section">
        <div class="detail-section__heading">
          <h3>备注</h3>
        </div>
        <p class="remark-copy">{{ subscription.remark }}</p>
      </section>

      <section class="detail-section renewal-section">
        <div class="detail-section__heading">
          <div>
            <h3>续订历史</h3>
            <p>共完成 {{ subscription.renew_count ?? 0 }} 次续订</p>
          </div>
          <n-button quaternary circle size="small" aria-label="刷新续订历史" @click="fetchRenewals">
            <template #icon><n-icon :component="ReloadOutline" /></template>
          </n-button>
        </div>

        <div v-if="renewalsLoading" class="renewal-loading">
          <n-spin size="small" /> 正在读取历史记录
        </div>

        <n-alert v-else-if="renewalsError" type="error" :bordered="false">
          {{ renewalsError }}
        </n-alert>

        <div v-else-if="renewals.length" class="renewal-timeline">
          <article v-for="renewal in renewals" :key="renewal.id" class="renewal-item">
            <span class="renewal-dot" aria-hidden="true"><n-icon :component="CheckmarkOutline" /></span>
            <div class="renewal-copy">
              <div>
                <strong>第 {{ renewal.renew_count }} 次续订</strong>
                <time :datetime="renewal.renewed_at">{{ formatDateTime(renewal.renewed_at) }}</time>
              </div>
              <p>
                <span>{{ formatDate(renewal.old_expire_date) }}</span>
                <n-icon :component="ArrowForwardOutline" />
                <strong>{{ formatDate(renewal.new_expire_date) }}</strong>
              </p>
            </div>
          </article>
        </div>

        <div v-else class="renewal-empty">
          <n-icon :component="TimeOutline" />
          <span>还没有续订记录</span>
        </div>
      </section>

      <template #footer>
        <div class="detail-actions">
          <n-button
            secondary
            type="primary"
            :loading="actionLoading === 'notify'"
            :disabled="actionLoading !== null && actionLoading !== 'notify'"
            @click="emit('notify', subscription)"
          >
            <template #icon><n-icon :component="NotificationsOutline" /></template>
            测试提醒
          </n-button>
          <n-button
            type="primary"
            :loading="actionLoading === 'renew'"
            :disabled="actionLoading !== null && actionLoading !== 'renew'"
            @click="emit('renew', subscription)"
          >
            <template #icon><n-icon :component="RefreshOutline" /></template>
            续订一次
          </n-button>
          <n-dropdown trigger="click" :options="moreOptions" @select="handleMoreAction">
            <n-button circle aria-label="更多操作">
              <template #icon><n-icon :component="EllipsisHorizontalOutline" /></template>
            </n-button>
          </n-dropdown>
        </div>
      </template>
    </n-drawer-content>
  </n-drawer>
</template>

<script setup lang="ts">
import { computed, h, ref, watch } from 'vue'
import {
  NAlert,
  NButton,
  NDrawer,
  NDrawerContent,
  NDropdown,
  NIcon,
  NSpin
} from 'naive-ui'
import {
  ArrowForwardOutline,
  CalendarOutline,
  CheckmarkOutline,
  CreateOutline,
  EllipsisHorizontalOutline,
  FlagOutline,
  NotificationsOutline,
  RefreshOutline,
  ReloadOutline,
  TimeOutline,
  TrashOutline
} from '@vicons/ionicons5'

import { getApiError, subscriptionApi } from '../../api'
import type { SubscriptionRecord, SubscriptionRenewal } from '../../api'
import { formatCurrency } from '../../utils/currency'
import { parseLocalDate } from '../../utils/date'
import { getCycleText, getSubscriptionStatus } from '../../utils/subscription'
import SubscriptionStatusPill from './SubscriptionStatusPill.vue'

export type SubscriptionAction = 'renew' | 'notify' | 'delete' | null

const props = defineProps<{
  show: boolean
  subscription: SubscriptionRecord | null
  actionLoading: SubscriptionAction
}>()

const emit = defineEmits<{
  'update:show': [show: boolean]
  edit: [subscription: SubscriptionRecord]
  renew: [subscription: SubscriptionRecord]
  notify: [subscription: SubscriptionRecord]
  delete: [subscription: SubscriptionRecord]
}>()

const renewals = ref<SubscriptionRenewal[]>([])
const renewalsLoading = ref(false)
const renewalsError = ref<string | null>(null)
let requestSequence = 0

const status = computed(() => props.subscription
  ? getSubscriptionStatus(props.subscription)
  : { kind: 'unknown' as const, label: '日期未知', daysRemaining: null })
const subscriptionInitial = computed(() => props.subscription?.name.trim().charAt(0).toUpperCase() || 'S')

const moreOptions = [
  {
    label: '编辑订阅',
    key: 'edit',
    icon: () => h(NIcon, null, { default: () => h(CreateOutline) })
  },
  {
    label: '删除订阅',
    key: 'delete',
    icon: () => h(NIcon, { color: '#C4322B' }, { default: () => h(TrashOutline) })
  }
]

/** 获取当前订阅的续订历史，并丢弃过期请求结果。 */
const fetchRenewals = async () => {
  if (!props.subscription) return
  const currentRequest = ++requestSequence
  renewalsLoading.value = true
  renewalsError.value = null

  try {
    const response = await subscriptionApi.listRenewals(props.subscription.id)
    if (currentRequest === requestSequence) renewals.value = response.data
  } catch (error: unknown) {
    if (currentRequest === requestSequence) {
      renewalsError.value = getApiError(error, '读取续订历史失败')
    }
  } finally {
    if (currentRequest === requestSequence) renewalsLoading.value = false
  }
}

/** 格式化仅含日期语义的后端字段。 */
const formatDate = (value: string | null) => {
  const date = parseLocalDate(value)
  if (!date) return '日期未知'
  return new Intl.DateTimeFormat('zh-CN', { year: 'numeric', month: 'short', day: 'numeric' }).format(date)
}

/** 格式化续订发生时间。 */
const formatDateTime = (value: string) => {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return '时间未知'
  return new Intl.DateTimeFormat('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  }).format(date)
}

/** 根据剩余天数生成详情页状态说明。 */
const getStatusDescription = (days: number | null) => {
  if (days === null) return '请检查到期日期是否正确'
  if (days < 0) return `已经逾期 ${Math.abs(days)} 天`
  if (days === 0) return '订阅将在今天到期'
  return `距离到期还有 ${days} 天`
}

/** 分发编辑或删除操作。 */
const handleMoreAction = (key: string) => {
  if (!props.subscription) return
  if (key === 'edit') emit('edit', props.subscription)
  if (key === 'delete') emit('delete', props.subscription)
}

watch(
  () => [props.show, props.subscription?.id, props.subscription?.renew_count] as const,
  ([show]) => {
    if (show && props.subscription) fetchRenewals()
  }
)
</script>

<style scoped>
.detail-header {
  display: flex;
  min-width: 0;
  align-items: center;
  gap: 12px;
}

.detail-header__avatar {
  display: grid;
  width: 42px;
  height: 42px;
  flex: 0 0 auto;
  border-radius: 13px;
  place-items: center;
  color: #3450b6;
  background: var(--color-primary-soft);
  font-size: 16px;
  font-weight: 780;
}

.detail-header > div {
  display: grid;
  min-width: 0;
}

.detail-header small {
  color: var(--color-text-muted);
  font-size: 9px;
  font-weight: 700;
  letter-spacing: 0.1em;
  text-transform: uppercase;
}

.detail-header strong {
  overflow: hidden;
  margin-top: 3px;
  font-size: 17px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.detail-hero {
  display: flex;
  padding: 20px;
  border: 1px solid #e5e9f2;
  border-radius: 17px;
  align-items: flex-start;
  justify-content: space-between;
  gap: 18px;
  background:
    radial-gradient(circle at 100% 0, rgba(94, 122, 232, 0.13), transparent 12rem),
    #fafbfe;
}

.detail-hero p {
  margin: 9px 0 0;
  color: var(--color-text-muted);
  font-size: 11px;
}

.detail-amount {
  display: grid;
  text-align: right;
}

.detail-amount strong {
  font-size: 23px;
  letter-spacing: -0.035em;
}

.detail-amount small {
  margin-top: 5px;
  color: var(--color-text-muted);
  font-size: 9px;
}

.detail-grid {
  display: grid;
  margin-top: 16px;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
}

.detail-field {
  display: grid;
  min-height: 83px;
  padding: 14px;
  border: 1px solid #e9ecf2;
  border-radius: 14px;
  align-content: center;
  background: #fff;
}

.detail-field span {
  display: flex;
  align-items: center;
  gap: 6px;
  color: var(--color-text-muted);
  font-size: 10px;
}

.detail-field span .n-icon {
  color: var(--color-primary);
  font-size: 14px;
}

.detail-field strong {
  margin-top: 7px;
  color: var(--color-text-secondary);
  font-size: 12px;
}

.detail-section {
  margin-top: 22px;
  padding-top: 20px;
  border-top: 1px solid var(--color-border);
}

.detail-section__heading {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

.detail-section__heading h3 {
  margin: 0;
  font-size: 14px;
}

.detail-section__heading p {
  margin: 4px 0 0;
  color: var(--color-text-muted);
  font-size: 9px;
}

.remark-copy {
  margin: 12px 0 0;
  padding: 14px;
  border-radius: 13px;
  color: #556177;
  background: #f7f8fb;
  font-size: 11px;
  line-height: 1.75;
  white-space: pre-wrap;
}

.renewal-loading,
.renewal-empty {
  display: flex;
  min-height: 120px;
  align-items: center;
  justify-content: center;
  gap: 8px;
  color: var(--color-text-muted);
  font-size: 11px;
}

.renewal-empty {
  flex-direction: column;
}

.renewal-empty .n-icon {
  color: #a5adbb;
  font-size: 25px;
}

.renewal-timeline {
  display: grid;
  margin-top: 16px;
}

.renewal-item {
  position: relative;
  display: grid;
  min-height: 72px;
  padding: 0 0 18px;
  grid-template-columns: 28px minmax(0, 1fr);
  gap: 10px;
}

.renewal-item:not(:last-child)::before {
  position: absolute;
  top: 27px;
  bottom: 0;
  left: 13px;
  width: 1px;
  content: '';
  background: #dfe4ed;
}

.renewal-dot {
  z-index: 1;
  display: grid;
  width: 28px;
  height: 28px;
  border-radius: 9px;
  place-items: center;
  color: var(--color-success);
  background: var(--color-success-soft);
  font-size: 14px;
}

.renewal-copy {
  padding: 1px 0;
}

.renewal-copy > div,
.renewal-copy p {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
}

.renewal-copy strong {
  color: var(--color-text-secondary);
  font-size: 11px;
}

.renewal-copy time {
  color: var(--color-text-muted);
  font-size: 9px;
}

.renewal-copy p {
  margin: 8px 0 0;
  justify-content: flex-start;
  color: var(--color-text-muted);
  font-size: 10px;
}

.renewal-copy p .n-icon {
  color: #9aa4b5;
  font-size: 13px;
}

.detail-actions {
  display: flex;
  width: 100%;
  gap: 9px;
}

.detail-actions > .n-button:nth-child(1),
.detail-actions > .n-button:nth-child(2) {
  flex: 1;
}

@media (max-width: 460px) {
  .detail-hero {
    align-items: flex-start;
    flex-direction: column;
  }

  .detail-amount {
    text-align: left;
  }

  .detail-grid {
    grid-template-columns: 1fr;
  }
}
</style>
