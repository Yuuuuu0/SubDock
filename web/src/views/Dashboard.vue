<template>
  <div class="page-container dashboard-page">
    <header class="page-header">
      <div class="page-header__copy">
        <p class="page-eyebrow">Overview</p>
        <h1 class="page-title">订阅概览</h1>
        <p class="page-subtitle">先看最需要处理的到期事项，再了解每种币种的周期支出。</p>
      </div>
      <n-button type="primary" @click="openCreate">
        <template #icon><n-icon :component="AddOutline" /></template>
        添加订阅
      </n-button>
    </header>

    <n-alert v-if="store.error && store.items.length > 0" type="warning" :bordered="false" class="dashboard-alert">
      <div class="alert-content">
        <span>刷新失败，当前展示的是上一次成功获取的数据。</span>
        <n-button size="small" @click="store.refresh">重试</n-button>
      </div>
    </n-alert>

    <template v-if="store.loading && !store.loaded">
      <div class="metrics-grid">
        <n-skeleton v-for="index in 4" :key="index" height="146px" :sharp="false" />
      </div>
      <div class="dashboard-grid">
        <n-skeleton height="360px" :sharp="false" />
        <n-skeleton height="360px" :sharp="false" />
      </div>
    </template>

    <section v-else-if="store.error && store.items.length === 0" class="surface-card">
      <empty-state
        :icon="CloudOfflineOutline"
        title="暂时无法读取订阅"
        :description="store.error"
        action-text="重新加载"
        @action="store.refresh"
      />
    </section>

    <section v-else-if="store.items.length === 0" class="surface-card">
      <empty-state
        :icon="AlbumsOutline"
        title="从第一项订阅开始"
        description="添加你正在使用的会员、软件或服务，SubDock 会自动计算到期日并整理提醒。"
        action-text="添加订阅"
        @action="openCreate"
      />
    </section>

    <template v-else>
      <section class="metrics-grid" aria-label="订阅关键指标">
        <metric-card
          label="全部订阅"
          :value="metrics.total"
          :detail="`${metrics.autoRenew} 项开启自动续订`"
          :icon="LayersOutline"
          tone="primary"
        />
        <metric-card
          label="30 天内到期"
          :value="metrics.dueWithin30"
          detail="建议提前检查是否继续使用"
          :icon="CalendarOutline"
          tone="warning"
        />
        <metric-card
          label="已经过期"
          :value="metrics.expired"
          detail="续订或删除不再使用的项目"
          :icon="AlertCircleOutline"
          tone="danger"
        />
        <metric-card
          label="状态正常"
          :value="metrics.active"
          detail="距离提醒日仍有充足时间"
          :icon="CheckmarkCircleOutline"
          tone="success"
        />
      </section>

      <section class="dashboard-grid">
        <article class="dashboard-card surface-card upcoming-card">
          <div class="section-heading">
            <div>
              <h2>近期时间线</h2>
              <p>优先展示已过期与未来 30 天内到期的项目。</p>
            </div>
            <router-link to="/subscriptions" class="text-link">查看全部</router-link>
          </div>

          <div v-if="upcomingSubscriptions.length" class="timeline-list">
            <button
              v-for="item in upcomingSubscriptions"
              :key="item.subscription.id"
              type="button"
              class="timeline-item"
              @click="openSubscription(item.subscription.id)"
            >
              <span class="timeline-date numeric">
                <strong>{{ formatDay(item.subscription.expire_date) }}</strong>
                <small>{{ formatMonth(item.subscription.expire_date) }}</small>
              </span>
              <span class="timeline-copy">
                <strong>{{ item.subscription.name }}</strong>
                <small>{{ getTimelineDescription(item.status.daysRemaining) }}</small>
              </span>
              <subscription-status-pill :status="item.status" />
              <n-icon class="timeline-arrow" :component="ChevronForwardOutline" />
            </button>
          </div>
          <div v-else class="quiet-state">
            <n-icon :component="CheckmarkDoneOutline" />
            <strong>未来 30 天暂无到期项目</strong>
            <span>可以安心把注意力放在更重要的事情上。</span>
          </div>
        </article>

        <article class="dashboard-card surface-card cost-card">
          <div class="section-heading">
            <div>
              <h2>月均费用估算</h2>
              <p>按现有周期折算，并严格按币种分别统计。</p>
            </div>
            <span class="estimate-tag">Estimate</span>
          </div>

          <div class="currency-list">
            <div v-for="summary in monthlySummaries" :key="summary.currency" class="currency-row">
              <span class="currency-code">{{ summary.currency }}</span>
              <span class="currency-copy">
                <strong class="numeric">{{ formatCurrency(summary.monthlyAmount, summary.currency) }}</strong>
                <small>{{ summary.subscriptionCount }} 项订阅 / 月</small>
              </span>
            </div>
          </div>

          <div class="cost-note">
            <n-icon :component="InformationCircleOutline" />
            <span>这里不使用实时汇率，因此不会把不同币种合并成一个总额。</span>
          </div>
        </article>
      </section>

      <section class="dashboard-grid dashboard-grid--lower">
        <article class="dashboard-card surface-card">
          <div class="section-heading">
            <div>
              <h2>费用较高的订阅</h2>
              <p>按折算后的月均费用排序，便于发现预算重点。</p>
            </div>
          </div>

          <div class="cost-ranking">
            <button
              v-for="(item, index) in highestCostSubscriptions"
              :key="item.subscription.id"
              type="button"
              class="ranking-row"
              @click="openSubscription(item.subscription.id)"
            >
              <span class="ranking-index">{{ String(index + 1).padStart(2, '0') }}</span>
              <span class="ranking-name">
                <strong>{{ item.subscription.name }}</strong>
                <small>{{ item.cycleText }}</small>
              </span>
              <strong class="ranking-amount numeric">
                {{ formatCurrency(item.monthlyCost, item.subscription.currency) }}<small>/月</small>
              </strong>
            </button>
          </div>
        </article>

        <article class="dashboard-card surface-card action-card">
          <span class="action-card__icon"><n-icon :component="NotificationsOutline" /></span>
          <div>
            <p class="page-eyebrow">Reminder check</p>
            <h2>通知渠道准备好了吗？</h2>
            <p>在设置中配置 Telegram 或 Bark，并发送一条测试消息确认链路可用。</p>
          </div>
          <n-button secondary type="primary" @click="router.push('/settings')">
            检查通知设置
          </n-button>
        </article>
      </section>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { NAlert, NButton, NIcon, NSkeleton } from 'naive-ui'
import {
  AddOutline,
  AlbumsOutline,
  AlertCircleOutline,
  CalendarOutline,
  CheckmarkCircleOutline,
  CheckmarkDoneOutline,
  ChevronForwardOutline,
  CloudOfflineOutline,
  InformationCircleOutline,
  LayersOutline,
  NotificationsOutline
} from '@vicons/ionicons5'

import MetricCard from '../components/dashboard/MetricCard.vue'
import EmptyState from '../components/feedback/EmptyState.vue'
import SubscriptionStatusPill from '../components/subscriptions/SubscriptionStatusPill.vue'
import type { SubscriptionRecord } from '../api'
import { useSubscriptionsStore } from '../stores/subscriptions'
import { formatCurrency } from '../utils/currency'
import { parseLocalDate } from '../utils/date'
import {
  getCycleText,
  getSubscriptionMonthlyCost,
  getSubscriptionStatus,
  summarizeMonthlyCostsByCurrency
} from '../utils/subscription'

const router = useRouter()
const store = useSubscriptionsStore()

const statuses = computed(() => store.items.map((subscription) => ({
  subscription,
  status: getSubscriptionStatus(subscription)
})))

const metrics = computed(() => ({
  total: store.items.length,
  autoRenew: store.items.filter((item) => item.auto_renew).length,
  dueWithin30: statuses.value.filter(({ status }) => (
    status.daysRemaining !== null && status.daysRemaining >= 0 && status.daysRemaining <= 30
  )).length,
  expired: statuses.value.filter(({ status }) => status.kind === 'expired').length,
  active: statuses.value.filter(({ status }) => status.kind === 'active').length
}))

const monthlySummaries = computed(() => summarizeMonthlyCostsByCurrency(store.items))

const upcomingSubscriptions = computed(() => statuses.value
  .filter(({ status }) => status.daysRemaining !== null && status.daysRemaining <= 30)
  .sort((left, right) => (left.status.daysRemaining ?? 0) - (right.status.daysRemaining ?? 0))
  .slice(0, 6))

const highestCostSubscriptions = computed(() => store.items
  .map((subscription) => ({
    subscription,
    monthlyCost: getSubscriptionMonthlyCost(subscription),
    cycleText: getCycleText(subscription.cycle_value, subscription.cycle_unit)
  }))
  .filter((item): item is { subscription: SubscriptionRecord; monthlyCost: number; cycleText: string } => (
    item.monthlyCost !== null
  ))
  .sort((left, right) => right.monthlyCost - left.monthlyCost)
  .slice(0, 5))

/** 跳转到订阅页并打开创建表单。 */
const openCreate = () => router.push({ path: '/subscriptions', query: { create: '1' } })

/** 跳转到订阅页并打开指定订阅详情。 */
const openSubscription = (id: number) => router.push({ path: '/subscriptions', query: { subscription: String(id) } })

/** 将后端日期格式化为两位日。 */
const formatDay = (value: string | null) => {
  const date = parseLocalDate(value)
  return date ? String(date.getDate()).padStart(2, '0') : '--'
}

/** 将后端日期格式化为短月份。 */
const formatMonth = (value: string | null) => {
  const date = parseLocalDate(value)
  return date ? `${date.getMonth() + 1} 月` : '未知'
}

/** 根据剩余天数生成时间线辅助文案。 */
const getTimelineDescription = (days: number | null) => {
  if (days === null) return '到期日期未知'
  if (days < 0) return `已逾期 ${Math.abs(days)} 天`
  if (days === 0) return '今天到期'
  return `还有 ${days} 天到期`
}

onMounted(store.fetch)
</script>

<style scoped>
.dashboard-alert {
  margin-bottom: 20px;
}

.alert-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

.metrics-grid {
  display: grid;
  margin-bottom: 22px;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 16px;
}

.dashboard-grid {
  display: grid;
  margin-top: 18px;
  grid-template-columns: minmax(0, 1.45fr) minmax(300px, 0.8fr);
  gap: 18px;
}

.dashboard-grid--lower {
  grid-template-columns: minmax(0, 1.2fr) minmax(300px, 0.8fr);
}

.dashboard-card {
  min-width: 0;
  padding: 24px;
}

.text-link {
  flex: 0 0 auto;
  color: var(--color-primary);
  font-size: 12px;
  font-weight: 680;
  text-decoration: none;
}

.text-link:hover {
  text-decoration: underline;
}

.timeline-list {
  display: grid;
  margin-top: 18px;
}

.timeline-item {
  display: grid;
  width: 100%;
  min-height: 68px;
  padding: 10px 4px;
  border: 0;
  border-bottom: 1px solid #edf0f5;
  grid-template-columns: 46px minmax(0, 1fr) auto 20px;
  align-items: center;
  gap: 12px;
  color: inherit;
  background: transparent;
  text-align: left;
  transition: background-color 160ms ease;
}

.timeline-item:hover {
  border-radius: 10px;
  background: #f8faff;
}

.timeline-item:last-child {
  border-bottom: 0;
}

.timeline-date {
  display: grid;
  width: 44px;
  height: 48px;
  border: 1px solid #e4e8f1;
  border-radius: 12px;
  place-content: center;
  color: #3b4860;
  background: #fafbfe;
  text-align: center;
}

.timeline-date strong {
  font-size: 16px;
  line-height: 1;
}

.timeline-date small {
  margin-top: 4px;
  color: var(--color-text-muted);
  font-size: 9px;
}

.timeline-copy {
  display: grid;
  min-width: 0;
}

.timeline-copy strong {
  overflow: hidden;
  color: var(--color-text);
  font-size: 13px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.timeline-copy small {
  margin-top: 4px;
  color: var(--color-text-muted);
  font-size: 10px;
}

.timeline-arrow {
  color: #a3abb9;
}

.quiet-state {
  display: grid;
  min-height: 250px;
  place-items: center;
  align-content: center;
  color: var(--color-text-muted);
  text-align: center;
}

.quiet-state > .n-icon {
  margin-bottom: 12px;
  color: var(--color-success);
  font-size: 36px;
}

.quiet-state strong {
  color: var(--color-text-secondary);
  font-size: 14px;
}

.quiet-state span {
  margin-top: 6px;
  font-size: 11px;
}

.estimate-tag {
  padding: 5px 8px;
  border-radius: 999px;
  color: #667085;
  background: #f1f3f7;
  font-size: 9px;
  font-weight: 750;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.currency-list {
  display: grid;
  margin-top: 22px;
  gap: 10px;
}

.currency-row {
  display: flex;
  min-height: 70px;
  padding: 13px;
  border: 1px solid #e9ecf2;
  border-radius: 14px;
  align-items: center;
  gap: 12px;
  background: #fafbfe;
}

.currency-code {
  display: grid;
  width: 44px;
  height: 38px;
  flex: 0 0 auto;
  border-radius: 11px;
  place-items: center;
  color: #3d55ad;
  background: #edf1ff;
  font-size: 10px;
  font-weight: 780;
}

.currency-copy {
  display: grid;
}

.currency-copy strong {
  color: var(--color-text);
  font-size: 17px;
}

.currency-copy small {
  margin-top: 3px;
  color: var(--color-text-muted);
  font-size: 10px;
}

.cost-note {
  display: flex;
  margin-top: 18px;
  padding-top: 16px;
  border-top: 1px solid var(--color-border);
  align-items: flex-start;
  gap: 8px;
  color: var(--color-text-muted);
  font-size: 10px;
  line-height: 1.55;
}

.cost-note .n-icon {
  flex: 0 0 auto;
  margin-top: 1px;
  color: var(--color-primary);
  font-size: 15px;
}

.cost-ranking {
  display: grid;
  margin-top: 16px;
}

.ranking-row {
  display: grid;
  min-height: 62px;
  padding: 10px 4px;
  border: 0;
  border-bottom: 1px solid #edf0f5;
  grid-template-columns: 34px minmax(0, 1fr) auto;
  align-items: center;
  gap: 10px;
  color: inherit;
  background: transparent;
  text-align: left;
}

.ranking-row:last-child {
  border-bottom: 0;
}

.ranking-row:hover {
  border-radius: 10px;
  background: #f8faff;
}

.ranking-index {
  color: #a1a9b8;
  font-size: 10px;
  font-weight: 720;
}

.ranking-name {
  display: grid;
  min-width: 0;
}

.ranking-name strong {
  overflow: hidden;
  font-size: 12px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.ranking-name small {
  margin-top: 3px;
  color: var(--color-text-muted);
  font-size: 10px;
}

.ranking-amount {
  color: var(--color-text-secondary);
  font-size: 12px;
}

.ranking-amount small {
  margin-left: 2px;
  color: var(--color-text-muted);
  font-size: 9px;
  font-weight: 500;
}

.action-card {
  display: flex;
  min-height: 300px;
  align-items: flex-start;
  justify-content: center;
  flex-direction: column;
  background:
    radial-gradient(circle at 100% 0, rgba(82, 113, 232, 0.16), transparent 15rem),
    rgba(255, 255, 255, 0.96);
}

.action-card__icon {
  display: grid;
  width: 50px;
  height: 50px;
  margin-bottom: 24px;
  border-radius: 16px;
  place-items: center;
  color: var(--color-primary);
  background: var(--color-primary-soft);
  font-size: 24px;
}

.action-card h2 {
  margin: 0;
  font-size: 22px;
  letter-spacing: -0.025em;
}

.action-card p:not(.page-eyebrow) {
  margin: 10px 0 22px;
  color: var(--color-text-muted);
  font-size: 12px;
  line-height: 1.7;
}

@media (max-width: 1180px) {
  .metrics-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 760px) {
  .dashboard-grid,
  .dashboard-grid--lower {
    grid-template-columns: 1fr;
  }

  .dashboard-card {
    padding: 20px;
  }
}

@media (max-width: 560px) {
  .metrics-grid {
    grid-template-columns: 1fr;
  }

  .timeline-item {
    grid-template-columns: 44px minmax(0, 1fr) 18px;
  }

  .timeline-item .status-pill {
    display: none;
  }
}
</style>
