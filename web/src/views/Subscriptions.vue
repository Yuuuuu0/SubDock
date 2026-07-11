<template>
  <div class="page-container subscriptions-page">
    <header class="page-header">
      <div class="page-header__copy">
        <p class="page-eyebrow">Subscriptions</p>
        <h1 class="page-title">订阅管理</h1>
        <p class="page-subtitle">检索、筛选和处理全部周期服务，点击订阅名称可查看续订历史与完整信息。</p>
      </div>
      <div class="header-actions">
        <n-dropdown :options="exportOptions" @select="exportSubscriptions">
          <n-button :disabled="store.items.length === 0">
            <template #icon><n-icon :component="DownloadOutline" /></template>
            导出
          </n-button>
        </n-dropdown>
        <n-button type="primary" @click="openCreate">
          <template #icon><n-icon :component="AddOutline" /></template>
          添加订阅
        </n-button>
      </div>
    </header>

    <section class="summary-strip surface-card" aria-label="订阅状态摘要">
      <div>
        <span>全部</span>
        <strong class="numeric">{{ summary.total }}</strong>
      </div>
      <div>
        <span>即将到期</span>
        <strong class="numeric summary-warning">{{ summary.dueSoon }}</strong>
      </div>
      <div>
        <span>已过期</span>
        <strong class="numeric summary-danger">{{ summary.expired }}</strong>
      </div>
      <div>
        <span>自动续订</span>
        <strong class="numeric summary-primary">{{ summary.autoRenew }}</strong>
      </div>
      <n-button quaternary circle aria-label="刷新订阅列表" :loading="store.loading" @click="store.refresh">
        <template #icon><n-icon :component="ReloadOutline" /></template>
      </n-button>
    </section>

    <section class="filter-panel surface-card" aria-label="订阅筛选">
      <div class="filter-topline">
        <n-input v-model:value="searchText" clearable placeholder="搜索名称、备注或币种" class="search-input">
          <template #prefix><n-icon :component="SearchOutline" /></template>
        </n-input>
        <div class="status-tabs" role="group" aria-label="按状态筛选">
          <button
            v-for="option in statusOptions"
            :key="option.value"
            type="button"
            :class="{ active: statusFilter === option.value }"
            @click="statusFilter = option.value"
          >
            {{ option.label }}
            <span class="numeric">{{ option.count }}</span>
          </button>
        </div>
      </div>

      <div class="filter-selects">
        <n-select v-model:value="currencyFilter" :options="currencyOptions" aria-label="按币种筛选" />
        <n-select v-model:value="renewFilter" :options="renewOptions" aria-label="按续订方式筛选" />
        <n-select v-model:value="sortBy" :options="sortOptions" aria-label="排序方式" />
        <button v-if="hasActiveFilters" class="reset-filter" type="button" @click="resetFilters">
          <n-icon :component="CloseCircleOutline" /> 清除筛选
        </button>
      </div>
    </section>

    <section v-if="store.loading && !store.loaded" class="loading-list surface-card" aria-label="正在加载订阅">
      <n-skeleton v-for="index in 7" :key="index" height="58px" :sharp="false" />
    </section>

    <section v-else-if="store.error && store.items.length === 0" class="surface-card">
      <empty-state
        :icon="CloudOfflineOutline"
        title="订阅列表加载失败"
        :description="store.error"
        action-text="重新加载"
        @action="store.refresh"
      />
    </section>

    <section v-else-if="store.items.length === 0" class="surface-card">
      <empty-state
        :icon="AlbumsOutline"
        title="还没有任何订阅"
        description="创建第一项订阅后，你就能在这里统一追踪到期日、周期费用和续订状态。"
        action-text="添加第一项订阅"
        @action="openCreate"
      />
    </section>

    <section v-else class="list-surface surface-card">
      <n-alert v-if="store.error" type="warning" :bordered="false" class="stale-alert">
        刷新失败，当前展示上一次成功加载的数据。
      </n-alert>

      <template v-if="filteredSubscriptions.length > 0">
        <div class="desktop-table">
          <table>
            <thead>
              <tr>
                <th>订阅</th>
                <th>金额</th>
                <th>周期</th>
                <th>到期日期</th>
                <th>状态</th>
                <th>续订</th>
                <th><span class="sr-only">操作</span></th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="subscription in pagedSubscriptions" :key="subscription.id">
                <td>
                  <button class="name-button" type="button" @click="openDetail(subscription)">
                    <span class="name-avatar" aria-hidden="true">{{ getInitial(subscription.name) }}</span>
                    <span class="name-copy">
                      <strong>{{ subscription.name }}</strong>
                      <small>{{ subscription.remark || '暂无备注' }}</small>
                    </span>
                  </button>
                </td>
                <td>
                  <span class="amount-cell numeric">
                    <strong>{{ formatCurrency(subscription.amount, subscription.currency) }}</strong>
                    <small>{{ subscription.currency }}</small>
                  </span>
                </td>
                <td><span class="muted-cell">{{ getCycleText(subscription.cycle_value, subscription.cycle_unit) }}</span></td>
                <td>
                  <span class="date-cell numeric">
                    <strong>{{ formatDate(subscription.expire_date) }}</strong>
                    <small>{{ getRemainingCopy(subscription) }}</small>
                  </span>
                </td>
                <td><subscription-status-pill :status="getSubscriptionStatus(subscription)" /></td>
                <td>
                  <span class="renew-cell">
                    <n-icon :component="subscription.auto_renew ? SyncOutline : HandLeftOutline" />
                    <span><strong>{{ subscription.auto_renew ? '自动' : '手动' }}</strong><small>{{ subscription.renew_count ?? 0 }} 次</small></span>
                  </span>
                </td>
                <td>
                  <div class="row-actions">
                    <n-button
                      size="small"
                      secondary
                      type="primary"
                      :loading="isActionLoading(subscription.id, 'renew')"
                      @click="confirmRenew(subscription)"
                    >
                      续订
                    </n-button>
                    <n-dropdown trigger="click" :options="rowActionOptions" @select="handleRowAction($event, subscription)">
                      <n-button quaternary circle size="small" :aria-label="`${subscription.name} 的更多操作`">
                        <template #icon><n-icon :component="EllipsisHorizontalOutline" /></template>
                      </n-button>
                    </n-dropdown>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <div class="mobile-list">
          <subscription-mobile-card
            v-for="subscription in pagedSubscriptions"
            :key="subscription.id"
            :subscription="subscription"
            :renewing="isActionLoading(subscription.id, 'renew')"
            @view="openDetail"
            @renew="confirmRenew"
          />
        </div>

        <footer v-if="pageCount > 1" class="list-footer">
          <span>共 {{ filteredSubscriptions.length }} 项，当前第 {{ page }} / {{ pageCount }} 页</span>
          <n-pagination v-model:page="page" :page-count="pageCount" :page-slot="5" />
        </footer>
      </template>

      <empty-state
        v-else
        :icon="SearchOutline"
        title="没有符合条件的订阅"
        description="尝试调整关键词或筛选条件，也可以清除筛选查看全部内容。"
        action-text="清除筛选"
        @action="resetFilters"
      />
    </section>

    <subscription-form-drawer
      v-model:show="showForm"
      :subscription="editingSubscription"
      @saved="handleSaved"
    />

    <subscription-detail-drawer
      v-model:show="showDetail"
      :subscription="selectedSubscription"
      :action-loading="selectedActionLoading"
      @edit="openEdit"
      @renew="confirmRenew"
      @notify="sendTestNotification"
      @delete="confirmDelete"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, h, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  NAlert,
  NButton,
  NDropdown,
  NIcon,
  NInput,
  NPagination,
  NSelect,
  NSkeleton,
  useDialog,
  useMessage
} from 'naive-ui'
import {
  AddOutline,
  AlbumsOutline,
  CloseCircleOutline,
  CloudOfflineOutline,
  CreateOutline,
  DocumentTextOutline,
  DownloadOutline,
  EllipsisHorizontalOutline,
  EyeOutline,
  HandLeftOutline,
  NotificationsOutline,
  ReloadOutline,
  SearchOutline,
  SyncOutline,
  TrashOutline
} from '@vicons/ionicons5'

import EmptyState from '../components/feedback/EmptyState.vue'
import SubscriptionDetailDrawer from '../components/subscriptions/SubscriptionDetailDrawer.vue'
import type { SubscriptionAction } from '../components/subscriptions/SubscriptionDetailDrawer.vue'
import SubscriptionFormDrawer from '../components/subscriptions/SubscriptionFormDrawer.vue'
import SubscriptionMobileCard from '../components/subscriptions/SubscriptionMobileCard.vue'
import SubscriptionStatusPill from '../components/subscriptions/SubscriptionStatusPill.vue'
import { getApiError, subscriptionApi } from '../api'
import type { SubscriptionRecord } from '../api'
import { useSubscriptionsStore } from '../stores/subscriptions'
import { formatCurrency } from '../utils/currency'
import { parseLocalDate } from '../utils/date'
import {
  getCycleText,
  getSubscriptionMonthlyCost,
  getSubscriptionStatus
} from '../utils/subscription'
import type { SubscriptionStatusKind } from '../utils/subscription'

type StatusFilter = 'all' | 'active' | 'due' | 'expired'
type RenewFilter = 'all' | 'auto' | 'manual'
type SortOption = 'expire_asc' | 'expire_desc' | 'name_asc' | 'cost_desc'
type RowAction = Exclude<SubscriptionAction, null> | 'view' | 'edit'

const route = useRoute()
const router = useRouter()
const dialog = useDialog()
const message = useMessage()
const store = useSubscriptionsStore()

const searchText = ref('')
const statusFilter = ref<StatusFilter>('all')
const currencyFilter = ref('all')
const renewFilter = ref<RenewFilter>('all')
const sortBy = ref<SortOption>('expire_asc')
const page = ref(1)
const pageSize = 10

const showForm = ref(false)
const editingSubscription = ref<SubscriptionRecord | null>(null)
const showDetail = ref(false)
const selectedSubscriptionId = ref<number | null>(null)
const actionState = ref<{ id: number; action: Exclude<SubscriptionAction, null> } | null>(null)

const summary = computed(() => {
  const statuses = store.items.map((subscription) => getSubscriptionStatus(subscription).kind)
  return {
    total: store.items.length,
    dueSoon: statuses.filter((kind) => kind === 'due_soon' || kind === 'due_today').length,
    expired: statuses.filter((kind) => kind === 'expired').length,
    autoRenew: store.items.filter((subscription) => subscription.auto_renew).length
  }
})

const statusOptions = computed<Array<{ label: string; value: StatusFilter; count: number }>>(() => [
  { label: '全部', value: 'all', count: summary.value.total },
  { label: '正常', value: 'active', count: store.items.filter((item) => getSubscriptionStatus(item).kind === 'active').length },
  { label: '即将到期', value: 'due', count: summary.value.dueSoon },
  { label: '已过期', value: 'expired', count: summary.value.expired }
])

const currencyOptions = computed(() => [
  { label: '全部币种', value: 'all' },
  ...[...new Set(store.items.map((item) => item.currency.trim().toUpperCase()).filter(Boolean))]
    .sort()
    .map((currency) => ({ label: currency, value: currency }))
])

const renewOptions: Array<{ label: string; value: RenewFilter }> = [
  { label: '全部续订方式', value: 'all' },
  { label: '自动续订', value: 'auto' },
  { label: '手动续订', value: 'manual' }
]

const sortOptions: Array<{ label: string; value: SortOption }> = [
  { label: '最近到期优先', value: 'expire_asc' },
  { label: '最晚到期优先', value: 'expire_desc' },
  { label: '名称 A–Z', value: 'name_asc' },
  { label: '月均费用从高到低', value: 'cost_desc' }
]

const statusMatches = (kind: SubscriptionStatusKind) => {
  if (statusFilter.value === 'all') return true
  if (statusFilter.value === 'due') return kind === 'due_soon' || kind === 'due_today'
  return kind === statusFilter.value
}

const filteredSubscriptions = computed(() => {
  const keyword = searchText.value.trim().toLocaleLowerCase('zh-CN')
  const filtered = store.items.filter((subscription) => {
    const matchesKeyword = !keyword || [subscription.name, subscription.remark ?? '', subscription.currency]
      .some((value) => value.toLocaleLowerCase('zh-CN').includes(keyword))
    const matchesStatus = statusMatches(getSubscriptionStatus(subscription).kind)
    const matchesCurrency = currencyFilter.value === 'all' || subscription.currency.toUpperCase() === currencyFilter.value
    const matchesRenew = renewFilter.value === 'all'
      || (renewFilter.value === 'auto' ? subscription.auto_renew : !subscription.auto_renew)
    return matchesKeyword && matchesStatus && matchesCurrency && matchesRenew
  })

  return filtered.sort((left, right) => {
    if (sortBy.value === 'name_asc') return left.name.localeCompare(right.name, 'zh-CN')
    if (sortBy.value === 'cost_desc') {
      return (getSubscriptionMonthlyCost(right) ?? -1) - (getSubscriptionMonthlyCost(left) ?? -1)
    }

    const leftTime = parseLocalDate(left.expire_date)?.getTime() ?? Number.POSITIVE_INFINITY
    const rightTime = parseLocalDate(right.expire_date)?.getTime() ?? Number.POSITIVE_INFINITY
    return sortBy.value === 'expire_desc' ? rightTime - leftTime : leftTime - rightTime
  })
})

const pageCount = computed(() => Math.max(1, Math.ceil(filteredSubscriptions.value.length / pageSize)))
const pagedSubscriptions = computed(() => {
  const start = (page.value - 1) * pageSize
  return filteredSubscriptions.value.slice(start, start + pageSize)
})
const hasActiveFilters = computed(() => (
  searchText.value.trim() !== ''
  || statusFilter.value !== 'all'
  || currencyFilter.value !== 'all'
  || renewFilter.value !== 'all'
  || sortBy.value !== 'expire_asc'
))
const selectedSubscription = computed(() => (
  store.items.find((subscription) => subscription.id === selectedSubscriptionId.value) ?? null
))
const selectedActionLoading = computed<SubscriptionAction>(() => (
  actionState.value?.id === selectedSubscriptionId.value ? actionState.value.action : null
))

const rowActionOptions = [
  { label: '查看详情', key: 'view', icon: () => h(NIcon, null, { default: () => h(EyeOutline) }) },
  { label: '编辑', key: 'edit', icon: () => h(NIcon, null, { default: () => h(CreateOutline) }) },
  { label: '测试提醒', key: 'notify', icon: () => h(NIcon, null, { default: () => h(NotificationsOutline) }) },
  { type: 'divider', key: 'divider' },
  { label: '删除', key: 'delete', icon: () => h(NIcon, { color: '#C4322B' }, { default: () => h(TrashOutline) }) }
]

const exportOptions = [
  { label: '导出为 JSON', key: 'json', icon: () => h(NIcon, null, { default: () => h(DocumentTextOutline) }) },
  { label: '导出为 CSV', key: 'csv', icon: () => h(NIcon, null, { default: () => h(DownloadOutline) }) }
]

/** 打开创建抽屉。 */
const openCreate = () => {
  editingSubscription.value = null
  showForm.value = true
}

/** 打开订阅编辑抽屉。 */
const openEdit = (subscription: SubscriptionRecord) => {
  showDetail.value = false
  editingSubscription.value = subscription
  showForm.value = true
}

/** 打开订阅详情抽屉。 */
const openDetail = (subscription: SubscriptionRecord) => {
  selectedSubscriptionId.value = subscription.id
  showDetail.value = true
}

/** 保存后刷新列表，并保留当前详情的最新数据。 */
const handleSaved = async (subscription: SubscriptionRecord) => {
  selectedSubscriptionId.value = subscription.id
  await store.refresh()
}

/** 请求手动续订前展示不可逆操作确认。 */
const confirmRenew = (subscription: SubscriptionRecord) => {
  dialog.info({
    title: '确认续订',
    content: `将“${subscription.name}”的到期日推进一个完整周期，并写入续订历史。`,
    positiveText: '确认续订',
    negativeText: '取消',
    onPositiveClick: () => renewSubscription(subscription)
  })
}

/** 执行续订并刷新列表。 */
const renewSubscription = async (subscription: SubscriptionRecord) => {
  actionState.value = { id: subscription.id, action: 'renew' }
  try {
    await subscriptionApi.renew(subscription.id)
    message.success(`“${subscription.name}”已续订`)
    await store.refresh()
  } catch (error: unknown) {
    message.error(getApiError(error, '续订失败'))
  } finally {
    actionState.value = null
  }
}

/** 使用当前通知配置发送单条测试提醒。 */
const sendTestNotification = async (subscription: SubscriptionRecord) => {
  actionState.value = { id: subscription.id, action: 'notify' }
  try {
    await subscriptionApi.testNotify(subscription.id)
    message.success('测试通知已发送')
  } catch (error: unknown) {
    message.error(getApiError(error, '测试通知发送失败'))
  } finally {
    actionState.value = null
  }
}

/** 删除前二次确认，成功后关闭关联详情。 */
const confirmDelete = (subscription: SubscriptionRecord) => {
  dialog.warning({
    title: '删除订阅',
    content: `确定删除“${subscription.name}”吗？这会将订阅移入数据库软删除状态。`,
    positiveText: '删除',
    negativeText: '取消',
    onPositiveClick: () => deleteSubscription(subscription)
  })
}

/** 执行订阅删除并刷新列表。 */
const deleteSubscription = async (subscription: SubscriptionRecord) => {
  actionState.value = { id: subscription.id, action: 'delete' }
  try {
    await subscriptionApi.delete(subscription.id)
    if (selectedSubscriptionId.value === subscription.id) showDetail.value = false
    message.success('订阅已删除')
    await store.refresh()
  } catch (error: unknown) {
    message.error(getApiError(error, '删除订阅失败'))
  } finally {
    actionState.value = null
  }
}

/** 处理表格更多菜单操作。 */
const handleRowAction = (key: string, subscription: SubscriptionRecord) => {
  const action = key as RowAction
  if (action === 'view') openDetail(subscription)
  if (action === 'edit') openEdit(subscription)
  if (action === 'notify') sendTestNotification(subscription)
  if (action === 'delete') confirmDelete(subscription)
}

/** 判断指定行是否正在执行某项异步操作。 */
const isActionLoading = (id: number, action: Exclude<SubscriptionAction, null>) => (
  actionState.value?.id === id && actionState.value.action === action
)

/** 清理全部筛选条件并返回第一页。 */
const resetFilters = () => {
  searchText.value = ''
  statusFilter.value = 'all'
  currencyFilter.value = 'all'
  renewFilter.value = 'all'
  sortBy.value = 'expire_asc'
  page.value = 1
}

/** 获取订阅名称首字符作为无图片头像。 */
const getInitial = (name: string) => name.trim().charAt(0).toUpperCase() || 'S'

/** 格式化到期日期。 */
const formatDate = (value: string | null) => {
  const date = parseLocalDate(value)
  if (!date) return '日期未知'
  return new Intl.DateTimeFormat('zh-CN', { year: 'numeric', month: '2-digit', day: '2-digit' }).format(date)
}

/** 生成剩余或逾期天数文案。 */
const getRemainingCopy = (subscription: SubscriptionRecord) => {
  const days = getSubscriptionStatus(subscription).daysRemaining
  if (days === null) return '请检查日期'
  if (days < 0) return `逾期 ${Math.abs(days)} 天`
  if (days === 0) return '今天到期'
  return `剩余 ${days} 天`
}

/** 下载 JSON 或 CSV 格式的当前全部订阅数据。 */
const exportSubscriptions = (format: string) => {
  const dateKey = new Date().toISOString().slice(0, 10)
  if (format === 'json') {
    downloadFile(
      `subdock-subscriptions-${dateKey}.json`,
      JSON.stringify(store.items, null, 2),
      'application/json;charset=utf-8'
    )
    message.success('JSON 导出已生成')
    return
  }

  const headers = ['name', 'amount', 'currency', 'start_date', 'cycle_value', 'cycle_unit', 'expire_date', 'auto_renew', 'renew_count', 'remind_days', 'remark']
  const rows = store.items.map((item) => [
    item.name,
    item.amount,
    item.currency,
    item.start_date.slice(0, 10),
    item.cycle_value,
    item.cycle_unit,
    item.expire_date?.slice(0, 10) ?? '',
    item.auto_renew ? 'true' : 'false',
    item.renew_count ?? 0,
    item.remind_days,
    item.remark ?? ''
  ])
  const csv = [headers, ...rows]
    .map((row) => row.map((value) => escapeCsv(String(value))).join(','))
    .join('\n')
  downloadFile(`subdock-subscriptions-${dateKey}.csv`, `\uFEFF${csv}`, 'text/csv;charset=utf-8')
  message.success('CSV 导出已生成')
}

/** 转义 CSV 单元格中的引号、逗号和换行。 */
const escapeCsv = (value: string) => /[",\n\r]/.test(value) ? `"${value.split('"').join('""')}"` : value

/** 通过浏览器 Blob 下载生成的导出文件。 */
const downloadFile = (filename: string, content: string, mimeType: string) => {
  const url = URL.createObjectURL(new Blob([content], { type: mimeType }))
  const link = document.createElement('a')
  link.href = url
  link.download = filename
  link.click()
  URL.revokeObjectURL(url)
}

/** 消费来自概览页和侧栏的创建/详情深链意图。 */
const consumeRouteIntent = async () => {
  if (route.query.create === '1') {
    openCreate()
    await router.replace('/subscriptions')
    return
  }

  if (typeof route.query.subscription === 'string') {
    const id = Number.parseInt(route.query.subscription, 10)
    if (Number.isInteger(id) && id > 0) {
      selectedSubscriptionId.value = id
      showDetail.value = true
    }
    await router.replace('/subscriptions')
  }
}

watch([searchText, statusFilter, currencyFilter, renewFilter, sortBy], () => {
  page.value = 1
})
watch(pageCount, (count) => {
  if (page.value > count) page.value = count
})
watch(() => route.fullPath, consumeRouteIntent, { immediate: true })

onMounted(store.fetch)
</script>

<style scoped>
.header-actions {
  display: flex;
  align-items: center;
  gap: 10px;
}

.summary-strip {
  display: grid;
  min-height: 82px;
  padding: 0 18px;
  grid-template-columns: repeat(4, minmax(0, 1fr)) 48px;
  align-items: center;
}

.summary-strip > div {
  display: grid;
  min-height: 42px;
  padding: 0 20px;
  border-right: 1px solid var(--color-border);
  align-content: center;
}

.summary-strip > div:first-child {
  padding-left: 4px;
}

.summary-strip span {
  color: var(--color-text-muted);
  font-size: 10px;
}

.summary-strip strong {
  margin-top: 3px;
  color: var(--color-text);
  font-size: 20px;
}

.summary-strip .summary-warning {
  color: var(--color-warning);
}

.summary-strip .summary-danger {
  color: var(--color-danger);
}

.summary-strip .summary-primary {
  color: var(--color-primary);
}

.filter-panel {
  margin-top: 16px;
  padding: 18px;
}

.filter-topline,
.filter-selects {
  display: flex;
  align-items: center;
  gap: 12px;
}

.filter-topline {
  justify-content: space-between;
}

.search-input {
  width: min(100%, 340px);
  flex: 0 1 340px;
}

.status-tabs {
  display: flex;
  padding: 4px;
  border-radius: 12px;
  gap: 3px;
  background: #f1f3f7;
}

.status-tabs button {
  display: flex;
  min-height: 34px;
  padding: 0 10px;
  border: 0;
  border-radius: 9px;
  align-items: center;
  gap: 7px;
  color: #6b7587;
  background: transparent;
  font-size: 10px;
  font-weight: 650;
  white-space: nowrap;
}

.status-tabs button:hover {
  color: var(--color-text-secondary);
}

.status-tabs button.active {
  color: var(--color-primary-strong);
  background: #fff;
  box-shadow: 0 3px 10px rgba(35, 48, 83, 0.08);
}

.status-tabs button span {
  display: grid;
  min-width: 18px;
  height: 18px;
  padding: 0 4px;
  border-radius: 999px;
  place-items: center;
  color: inherit;
  background: rgba(54, 89, 227, 0.08);
  font-size: 8px;
}

.filter-selects {
  margin-top: 12px;
}

.filter-selects .n-select {
  width: 170px;
}

.filter-selects .n-select:last-of-type {
  width: 200px;
}

.reset-filter {
  display: inline-flex;
  min-height: 34px;
  padding: 0 8px;
  border: 0;
  align-items: center;
  gap: 5px;
  color: var(--color-primary);
  background: transparent;
  font-size: 10px;
  font-weight: 650;
}

.loading-list {
  display: grid;
  margin-top: 16px;
  padding: 20px;
  gap: 10px;
}

.list-surface {
  margin-top: 16px;
  overflow: hidden;
}

.stale-alert {
  margin: 14px 14px 0;
}

.desktop-table {
  overflow-x: auto;
}

table {
  width: 100%;
  min-width: 900px;
  border-collapse: collapse;
}

th {
  height: 48px;
  padding: 0 15px;
  color: #7a8497;
  background: #f8f9fc;
  font-size: 10px;
  font-weight: 700;
  letter-spacing: 0.035em;
  text-align: left;
  text-transform: uppercase;
}

td {
  height: 74px;
  padding: 10px 15px;
  border-top: 1px solid #edf0f5;
  color: var(--color-text-secondary);
  font-size: 11px;
}

tbody tr {
  transition: background-color 150ms ease;
}

tbody tr:hover {
  background: #fafbff;
}

.name-button {
  display: flex;
  max-width: 260px;
  padding: 0;
  border: 0;
  align-items: center;
  gap: 11px;
  color: inherit;
  background: transparent;
  text-align: left;
}

.name-avatar {
  display: grid;
  width: 38px;
  height: 38px;
  flex: 0 0 auto;
  border-radius: 12px;
  place-items: center;
  color: #3653b5;
  background: var(--color-primary-soft);
  font-size: 12px;
  font-weight: 780;
}

.name-copy,
.amount-cell,
.date-cell,
.renew-cell > span {
  display: grid;
  min-width: 0;
}

.name-copy strong {
  overflow: hidden;
  color: var(--color-text);
  font-size: 12px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.name-copy small {
  max-width: 200px;
  overflow: hidden;
  margin-top: 4px;
  color: var(--color-text-muted);
  font-size: 9px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.amount-cell strong,
.date-cell strong {
  color: var(--color-text-secondary);
  font-size: 11px;
}

.amount-cell small,
.date-cell small {
  margin-top: 4px;
  color: var(--color-text-muted);
  font-size: 9px;
}

.muted-cell {
  color: #5f6b80;
  white-space: nowrap;
}

.renew-cell {
  display: flex;
  align-items: center;
  gap: 7px;
}

.renew-cell > .n-icon {
  color: var(--color-primary);
  font-size: 16px;
}

.renew-cell strong {
  font-size: 10px;
}

.renew-cell small {
  margin-top: 2px;
  color: var(--color-text-muted);
  font-size: 8px;
}

.row-actions {
  display: flex;
  justify-content: flex-end;
  gap: 5px;
}

.mobile-list {
  display: none;
}

.list-footer {
  display: flex;
  min-height: 64px;
  padding: 12px 18px;
  border-top: 1px solid var(--color-border);
  align-items: center;
  justify-content: space-between;
  gap: 18px;
}

.list-footer > span {
  color: var(--color-text-muted);
  font-size: 10px;
}

@media (max-width: 1050px) {
  .filter-topline {
    align-items: stretch;
    flex-direction: column;
  }

  .search-input {
    width: 100%;
    flex-basis: auto;
  }

  .status-tabs {
    align-self: flex-start;
    overflow-x: auto;
    max-width: 100%;
  }
}

@media (max-width: 900px) {
  .desktop-table {
    display: none;
  }

  .list-surface {
    overflow: visible;
    border: 0;
    background: transparent;
    box-shadow: none;
  }

  .mobile-list {
    display: grid;
    gap: 12px;
  }

  .list-footer {
    margin-top: 12px;
    padding: 16px;
    border: 1px solid var(--color-border);
    border-radius: 16px;
    align-items: stretch;
    flex-direction: column;
    background: #fff;
  }

  .list-footer > span {
    text-align: center;
  }

  .list-footer .n-pagination {
    justify-content: center;
  }
}

@media (max-width: 760px) {
  .header-actions {
    display: grid;
    width: 100%;
    grid-template-columns: 0.75fr 1.25fr;
  }

  .header-actions .n-button {
    width: 100%;
  }

  .summary-strip {
    min-height: 0;
    padding: 14px;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 9px;
  }

  .summary-strip > div {
    min-height: 58px;
    padding: 10px 13px;
    border: 1px solid #e9ecf2;
    border-radius: 12px;
    background: #fafbfe;
  }

  .summary-strip > div:first-child {
    padding-left: 13px;
  }

  .summary-strip > .n-button {
    position: absolute;
    opacity: 0;
    pointer-events: none;
  }

  .filter-selects {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .filter-selects .n-select,
  .filter-selects .n-select:last-of-type {
    width: 100%;
  }

  .filter-selects .n-select:last-of-type {
    grid-column: 1 / -1;
  }

  .reset-filter {
    justify-self: start;
  }

}

@media (max-width: 460px) {
  .filter-panel {
    padding: 14px;
  }

  .status-tabs {
    width: 100%;
  }

  .status-tabs button {
    min-width: max-content;
    flex: 1;
  }
}
</style>
