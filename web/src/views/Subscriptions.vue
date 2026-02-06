<template>
  <div class="page-container">
    <div class="page-header">
      <div>
        <h1 class="page-title">订阅列表</h1>
        <p class="page-subtitle">管理您的所有定期订阅服务</p>
      </div>
    </div>

    <n-card class="content-card" :bordered="false">
      <div class="action-bar">
        <n-input v-model:value="searchText" placeholder="搜索订阅..." class="search-input">
          <template #prefix>
            <n-icon :component="SearchOutline" />
          </template>
        </n-input>
        <n-button type="primary" class="add-btn" @click="handleCreate">
          <template #icon>
            <n-icon :component="AddOutline" />
          </template>
          添加订阅
        </n-button>
      </div>

      <n-data-table
        :columns="columns"
        :data="filteredData"
        :loading="loading"
        :pagination="pagination"
        :row-class-name="rowClassName"
        table-layout="fixed"
        :scroll-x="1120"
      />
    </n-card>

    <!-- Create/Edit Modal -->
    <n-modal
      v-model:show="showModal"
      preset="card"
      :title="modalTitle"
      style="width: 1000px; max-width: 90vw;"
      :mask-closable="false"
    >
      <n-form
        ref="formRef"
        :model="formValue"
        :rules="rules"
        label-placement="left"
        label-width="100"
        require-mark-placement="right-hanging"
      >
        <n-form-item label="订阅名称" path="name">
          <n-input v-model:value="formValue.name" placeholder="例如：Netflix" />
        </n-form-item>
        
        <n-grid :cols="2" :x-gap="12">
          <n-form-item-gi label="金额" path="amount">
            <n-input-number v-model:value="formValue.amount" :min="0" :precision="2" style="width: 100%" />
          </n-form-item-gi>
          <n-form-item-gi label="货币" path="currency">
            <n-select v-model:value="formValue.currency" :options="currencyOptions" />
          </n-form-item-gi>
        </n-grid>

        <n-grid :cols="2" :x-gap="12">
          <n-form-item-gi label="周期数值" path="cycle_value">
            <n-input-number v-model:value="formValue.cycle_value" :min="1" style="width: 100%" />
          </n-form-item-gi>
          <n-form-item-gi label="周期单位" path="cycle_unit">
            <n-select v-model:value="formValue.cycle_unit" :options="cycleUnitOptions" />
          </n-form-item-gi>
        </n-grid>

        <n-form-item label="开始日期" path="start_date">
          <n-date-picker v-model:formatted-value="formValue.start_date" value-format="yyyy-MM-dd" type="date" style="width: 100%" />
        </n-form-item>

        <n-form-item label="提醒天数" path="remind_days">
          <n-input-number v-model:value="formValue.remind_days" :min="1" style="width: 100%">
            <template #suffix>天</template>
          </n-input-number>
        </n-form-item>

        <n-form-item label="自动续订" path="auto_renew">
          <n-switch v-model:value="formValue.auto_renew" />
        </n-form-item>

        <n-form-item label="备注" path="remark">
          <n-input v-model:value="formValue.remark" type="textarea" placeholder="可选备注信息" />
        </n-form-item>
      </n-form>
      <template #footer>
        <div class="modal-actions">
          <n-button @click="showModal = false">取消</n-button>
          <n-button type="primary" @click="handleSubmit" :loading="submitting">保存</n-button>
        </div>
      </template>
    </n-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, h } from 'vue'
import { useMessage, useDialog, NButton, NTag, NSpace, NIcon, NTooltip } from 'naive-ui'
import type { DataTableColumns, FormInst } from 'naive-ui'
import { SearchOutline, AddOutline, CreateOutline, TrashOutline, NotificationsOutline, RefreshOutline } from '@vicons/ionicons5'
import { subscriptionApi } from '../api'
import type { Subscription } from '../api'

type SubscriptionSubmitPayload = Pick<Subscription, 'name' | 'amount' | 'currency' | 'start_date' | 'cycle_value' | 'cycle_unit' | 'remind_days' | 'remark' | 'auto_renew'>

const message = useMessage()
const dialog = useDialog()

// State
const loading = ref(false)
const subscriptions = ref<Subscription[]>([])
const searchText = ref('')
const showModal = ref(false)
const submitting = ref(false)
const formRef = ref<FormInst | null>(null)
const isEdit = ref(false)
const editingId = ref<number | null>(null)

// Form Data
const formValue = ref<Subscription>({
  name: '',
  amount: 0,
  currency: 'CNY',
  start_date: new Date().toISOString().slice(0, 10),
  cycle_value: 1,
  cycle_unit: 'month',
  expire_date: null,
  auto_renew: false,
  renew_count: 0,
  remind_days: 3,
  remark: ''
})

// Constants
const currencyOptions = [
  { label: 'CNY (¥)', value: 'CNY' },
  { label: 'USD ($)', value: 'USD' },
  { label: 'EUR (€)', value: 'EUR' },
  { label: 'JPY (¥)', value: 'JPY' },
  { label: 'HKD ($)', value: 'HKD' }
]

const cycleUnitOptions = [
  { label: '天', value: 'day' },
  { label: '月', value: 'month' },
  { label: '季', value: 'quarter' },
  { label: '半年', value: 'half_year' },
  { label: '年', value: 'year' }
]

const rules = {
  name: { required: true, message: '请输入订阅名称', trigger: 'blur' },
  amount: { required: true, type: 'number', message: '请输入金额', trigger: 'blur' },
  currency: { required: true, message: '请选择货币', trigger: 'blur' },
  start_date: { required: true, message: '请选择开始日期', trigger: 'blur' },
  cycle_value: { required: true, type: 'number', message: '请输入周期数值', trigger: 'blur' },
  cycle_unit: { required: true, message: '请选择周期单位', trigger: 'blur' }
}

// Computed
const modalTitle = computed(() => isEdit.value ? '编辑订阅' : '添加订阅')

const filteredData = computed(() => {
  const lower = searchText.value.toLowerCase()
  const filtered = !searchText.value
    ? [...subscriptions.value]
    : subscriptions.value.filter(item => 
      item.name.toLowerCase().includes(lower) || 
      item.remark?.toLowerCase().includes(lower)
    )

  return filtered.sort((a, b) => {
    const aRaw = a.expire_date ? new Date(a.expire_date).getTime() : Number.POSITIVE_INFINITY
    const bRaw = b.expire_date ? new Date(b.expire_date).getTime() : Number.POSITIVE_INFINITY
    const aTime = Number.isNaN(aRaw) ? Number.POSITIVE_INFINITY : aRaw
    const bTime = Number.isNaN(bRaw) ? Number.POSITIVE_INFINITY : bRaw
    return aTime - bTime
  })
})

const pagination = {
  pageSize: 10
}

// Helpers
const getCurrencySymbol = (code: string) => {
  switch(code) {
    case 'CNY': return '¥'
    case 'USD': return '$'
    case 'EUR': return '€'
    case 'JPY': return '¥'
    case 'HKD': return '$'
    default: return code
  }
}

const getCycleText = (val: number, unit: string) => {
  const unitMap: Record<string, string> = {
    day: '天',
    month: '月',
    quarter: '季',
    half_year: '半年',
    year: '年'
  }
  return `每 ${val} ${unitMap[unit]}`
}

const getDaysRemaining = (expireDate: string | undefined | null) => {
  if (!expireDate) return 0
  const today = new Date()
  today.setHours(0, 0, 0, 0)
  const expire = new Date(expireDate)
  if (isNaN(expire.getTime())) return 0
  expire.setHours(0, 0, 0, 0)
  const diffTime = expire.getTime() - today.getTime()
  return Math.ceil(diffTime / (1000 * 60 * 60 * 24))
}

const formatStatus = (expireDate: string, remindDays: number): { type: 'error' | 'warning' | 'success'; text: string } => {
  const days = getDaysRemaining(expireDate)
  if (days < 0) return { type: 'error', text: '已过期' }
  if (days <= remindDays) return { type: 'warning', text: '即将到期' }
  return { type: 'success', text: '正常' }
}

// Columns - 使用 computed 避免 slot 警告
const columns = computed<DataTableColumns<Subscription>>(() => [
  {
    title: '名称',
    key: 'name',
    width: 280,
    render(row) {
      const remark = row.remark?.trim() || ''
      const remarkNode = remark
        ? h(
            NTooltip,
            { trigger: 'hover' },
            {
              default: () => remark,
              trigger: () =>
                h(
                  'div',
                  {
                    class: 'name-remark-ellipsis'
                  },
                  remark
                )
            }
          )
        : null

      return h('div', { class: 'name-cell' }, [
        h('div', { style: 'font-weight: 600; color: #1A1A1A;' }, row.name),
        ...(remarkNode ? [remarkNode] : [])
      ])
    }
  },
  {
    title: '金额',
    key: 'amount',
    width: 130,
    render(row) {
      return h('div', { style: 'color: #10B981; font-family: monospace; font-weight: 600;' }, 
        `${getCurrencySymbol(row.currency)} ${row.amount}`
      )
    }
  },
  {
    title: '周期',
    key: 'cycle',
    width: 120,
    render(row) {
      return getCycleText(row.cycle_value, row.cycle_unit)
    }
  },
  {
    title: '到期日期',
    key: 'expire_date',
    width: 180,
    render(row) {
      if (!row.expire_date) return h('div', {}, '-')
      const days = getDaysRemaining(row.expire_date)
      const dateOnly = row.expire_date.split('T')[0]
      return h('div', {}, [
        h('div', { style: 'font-weight: 600;' }, dateOnly),
        h('div', { style: 'font-size: 0.8em; color: #6B7280;' }, 
          days < 0 ? `逾期 ${Math.abs(days)} 天` : `剩余 ${days} 天`
        )
      ])
    }
  },
  {
    title: '状态',
    key: 'status',
    width: 110,
    render(row) {
      const status = formatStatus(row.expire_date || '', row.remind_days)
      return h(NTag, { type: status.type, bordered: false, round: true }, () => status.text)
    }
  },
  {
    title: '续订',
    key: 'renew',
    width: 130,
    render(row) {
      const renewCount = row.renew_count ?? 0
      const autoRenew = row.auto_renew ? '自动' : '手动'
      return h('div', {}, [
        h('div', { style: 'font-weight: 600;' }, `${renewCount} 次`),
        h('div', { style: 'font-size: 0.8em; color: #6B7280;' }, autoRenew)
      ])
    }
  },
  {
    title: '操作',
    key: 'actions',
    width: 220,
    render(row) {
      return h(NSpace, null, () => [
        h(NButton, {
          size: 'small',
          quaternary: true,
          circle: true,
          type: 'success',
          onClick: () => handleRenew(row)
        }, { icon: () => h(NIcon, null, () => h(RefreshOutline)) }),
        h(NButton, {
          size: 'small',
          quaternary: true,
          circle: true,
          onClick: () => handleTestNotify(row)
        }, { icon: () => h(NIcon, null, () => h(NotificationsOutline)) }),
        h(NButton, {
          size: 'small',
          quaternary: true,
          circle: true,
          onClick: () => handleEdit(row)
        }, { icon: () => h(NIcon, null, () => h(CreateOutline)) }),
        h(NButton, {
          size: 'small',
          quaternary: true,
          circle: true,
          type: 'error',
          onClick: () => handleDelete(row)
        }, { icon: () => h(NIcon, null, () => h(TrashOutline)) })
      ])
    }
  }
])

const rowClassName = (_row: Subscription) => {
  return 'sub-row'
}

// Methods
const fetchData = async () => {
  loading.value = true
  try {
    const res = await subscriptionApi.list()
    subscriptions.value = res.data.map(item => ({
      ...item,
      expire_date: item.expire_date || null,
      remark: item.remark || ''
    }))
  } catch (error) {
    message.error('获取订阅列表失败')
  } finally {
    loading.value = false
  }
}

const handleCreate = () => {
  isEdit.value = false
  editingId.value = null
  formValue.value = {
    name: '',
    amount: 0,
    currency: 'CNY',
    start_date: new Date().toISOString().slice(0, 10),
    cycle_value: 1,
    cycle_unit: 'month',
    expire_date: null,
    auto_renew: false,
    renew_count: 0,
    remind_days: 3,
    remark: ''
  }
  showModal.value = true
}

const handleEdit = (row: Subscription) => {
  isEdit.value = true
  editingId.value = row.id as number
  formValue.value = { 
    ...row,
    start_date: row.start_date.slice(0, 10),
    expire_date: row.expire_date ? row.expire_date.slice(0, 10) : null,
    auto_renew: row.auto_renew ?? false,
    renew_count: row.renew_count ?? 0
  }
  showModal.value = true
}

const handleRenew = (row: Subscription) => {
  dialog.info({
    title: '确认续订',
    content: `确定要为订阅 "${row.name}" 续订 1 次吗？`,
    positiveText: '确定',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        await subscriptionApi.renew(row.id as number)
        message.success('续订成功')
        fetchData()
      } catch (error: any) {
        message.error(error.response?.data?.error || '续订失败')
      }
    }
  })
}

const handleDelete = (row: Subscription) => {
  dialog.warning({
    title: '确认删除',
    content: `确定要删除订阅 "${row.name}" 吗？`,
    positiveText: '确定',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        await subscriptionApi.delete(row.id as number)
        message.success('删除成功')
        fetchData()
      } catch (error) {
        message.error('删除失败')
      }
    }
  })
}

const handleTestNotify = async (row: Subscription) => {
  try {
    await subscriptionApi.testNotify(row.id as number)
    message.success('测试通知已发送')
  } catch (error: any) {
    message.error(error.response?.data?.error || '发送测试通知失败')
  }
}

const handleSubmit = (e: Event) => {
  e.preventDefault()
  formRef.value?.validate(async (errors) => {
    if (!errors) {
      submitting.value = true
      try {
        const payload: SubscriptionSubmitPayload = {
          name: formValue.value.name,
          amount: formValue.value.amount,
          currency: formValue.value.currency,
          start_date: formValue.value.start_date,
          cycle_value: formValue.value.cycle_value,
          cycle_unit: formValue.value.cycle_unit,
          remind_days: formValue.value.remind_days,
          remark: formValue.value.remark || '',
          auto_renew: formValue.value.auto_renew ?? false
        }
        if (isEdit.value && editingId.value) {
          await subscriptionApi.update(editingId.value, payload as Subscription)
          message.success('更新成功')
        } else {
          await subscriptionApi.create(payload as Subscription)
          message.success('创建成功')
        }
        showModal.value = false
        fetchData()
      } catch (error: any) {
        message.error(error.response?.data?.error || '操作失败')
      } finally {
        submitting.value = false
      }
    }
  })
}

onMounted(() => {
  fetchData()
})
</script>

<style scoped>
.page-container {
  max-width: 1280px;
  margin: 0 auto;
}

.name-cell {
  min-width: 0;
}

.name-remark-ellipsis {
  margin-top: 2px;
  font-size: 12px;
  color: #6B7280;
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
}

.page-header {
  margin-bottom: 24px;
}

.page-title {
  font-size: 1.8rem;
  font-weight: 700;
  margin: 0;
  color: #1A1A1A;
}

.page-subtitle {
  color: #6B7280;
  margin-top: 4px;
}

.content-card {
  box-shadow: 0 1px 3px 0 rgba(0, 0, 0, 0.1), 0 1px 2px 0 rgba(0, 0, 0, 0.06);
}

.action-bar {
  display: flex;
  justify-content: space-between;
  margin-bottom: 20px;
}

.search-input {
  width: 300px;
}

.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

:deep(.sub-row td) {
  padding: 16px 12px;
}
</style>
