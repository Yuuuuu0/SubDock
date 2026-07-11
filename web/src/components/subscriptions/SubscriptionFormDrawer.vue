<template>
  <n-drawer
    :show="show"
    width="min(100vw, 620px)"
    placement="right"
    :mask-closable="!submitting"
    @update:show="emit('update:show', $event)"
  >
    <n-drawer-content :native-scrollbar="false" closable>
      <template #header>
        <div class="drawer-title">
          <span class="drawer-title__icon" aria-hidden="true">
            <n-icon :component="subscription ? CreateOutline : AddOutline" />
          </span>
          <div>
            <small>{{ subscription ? 'Edit subscription' : 'New subscription' }}</small>
            <strong>{{ subscription ? '编辑订阅' : '添加新订阅' }}</strong>
          </div>
        </div>
      </template>

      <n-form
        ref="formRef"
        :model="formValue"
        :rules="rules"
        label-placement="top"
        class="subscription-form"
        @submit.prevent="submit"
      >
        <section class="form-section">
          <div class="form-section__heading">
            <span>01</span>
            <div><strong>基本信息</strong><small>名称、价格与结算币种</small></div>
          </div>

          <n-form-item label="订阅名称" path="name">
            <n-input v-model:value="formValue.name" placeholder="例如：流媒体会员、云服务器" maxlength="128" />
          </n-form-item>

          <div class="form-grid form-grid--amount">
            <n-form-item label="每周期金额" path="amount">
              <n-input-number
                v-model:value="formValue.amount"
                :min="0"
                :precision="2"
                :show-button="false"
                placeholder="0.00"
              />
            </n-form-item>
            <n-form-item label="币种" path="currency">
              <n-select v-model:value="formValue.currency" filterable tag :options="currencyOptions" />
            </n-form-item>
          </div>
        </section>

        <section class="form-section">
          <div class="form-section__heading">
            <span>02</span>
            <div><strong>周期与日期</strong><small>到期日将根据开始日期和周期自动计算</small></div>
          </div>

          <n-form-item label="开始日期" path="start_date">
            <n-date-picker
              v-model:formatted-value="formValue.start_date"
              value-format="yyyy-MM-dd"
              type="date"
              clearable
            />
          </n-form-item>

          <div class="form-grid">
            <n-form-item label="周期数值" path="cycle_value">
              <n-input-number v-model:value="formValue.cycle_value" :min="1" :precision="0" />
            </n-form-item>
            <n-form-item label="周期单位" path="cycle_unit">
              <n-select v-model:value="formValue.cycle_unit" :options="cycleUnitOptions" />
            </n-form-item>
          </div>

          <div class="cycle-preview">
            <n-icon :component="CalendarOutline" />
            <span>当前周期：<strong>{{ cyclePreview }}</strong></span>
          </div>
        </section>

        <section class="form-section">
          <div class="form-section__heading">
            <span>03</span>
            <div><strong>续订与提醒</strong><small>决定到期时的处理方式</small></div>
          </div>

          <div class="form-grid">
            <n-form-item label="提前提醒" path="remind_days">
              <n-input-number v-model:value="formValue.remind_days" :min="0" :max="3650" :precision="0">
                <template #suffix>天</template>
              </n-input-number>
            </n-form-item>

            <n-form-item label="自动续订">
              <div class="switch-field">
                <n-switch v-model:value="formValue.auto_renew" />
                <span>{{ formValue.auto_renew ? '到期后自动推进一个周期' : '到期后等待手动确认' }}</span>
              </div>
            </n-form-item>
          </div>

          <div v-if="formValue.remind_days === 0" class="form-note">
            <n-icon :component="InformationCircleOutline" />
            设为 0 天时，仅在到期当天进入提醒窗口。
          </div>
        </section>

        <section class="form-section">
          <div class="form-section__heading">
            <span>04</span>
            <div><strong>备注</strong><small>记录套餐、账号或取消方式等信息</small></div>
          </div>
          <n-form-item path="remark">
            <n-input
              v-model:value="formValue.remark"
              type="textarea"
              :autosize="{ minRows: 3, maxRows: 6 }"
              maxlength="512"
              show-count
              placeholder="可选，例如：家庭套餐，每年 6 月前可取消"
            />
          </n-form-item>
        </section>
      </n-form>

      <template #footer>
        <div class="drawer-actions">
          <n-button :disabled="submitting" @click="emit('update:show', false)">取消</n-button>
          <n-button type="primary" :loading="submitting" @click="submit">
            {{ subscription ? '保存修改' : '创建订阅' }}
          </n-button>
        </div>
      </template>
    </n-drawer-content>
  </n-drawer>
</template>

<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import {
  NButton,
  NDatePicker,
  NDrawer,
  NDrawerContent,
  NForm,
  NFormItem,
  NIcon,
  NInput,
  NInputNumber,
  NSelect,
  NSwitch,
  useMessage
} from 'naive-ui'
import type { FormInst, FormRules } from 'naive-ui'
import {
  AddOutline,
  CalendarOutline,
  CreateOutline,
  InformationCircleOutline
} from '@vicons/ionicons5'

import { getApiError, subscriptionApi } from '../../api'
import type { CycleUnit, SubscriptionRecord, SubscriptionWriteRequest } from '../../api'
import { formatLocalDateKey } from '../../utils/date'
import { getCycleText } from '../../utils/subscription'

interface SubscriptionFormModel {
  name: string
  amount: number
  currency: string
  start_date: string
  cycle_value: number
  cycle_unit: CycleUnit
  auto_renew: boolean
  remind_days: number
  remark: string
}

const props = defineProps<{
  show: boolean
  subscription: SubscriptionRecord | null
}>()

const emit = defineEmits<{
  'update:show': [show: boolean]
  saved: [subscription: SubscriptionRecord]
}>()

const message = useMessage()
const formRef = ref<FormInst | null>(null)
const submitting = ref(false)

const createInitialForm = (): SubscriptionFormModel => ({
  name: '',
  amount: 0,
  currency: 'CNY',
  start_date: formatLocalDateKey(),
  cycle_value: 1,
  cycle_unit: 'month',
  auto_renew: false,
  remind_days: 7,
  remark: ''
})

const formValue = reactive<SubscriptionFormModel>(createInitialForm())

const currencyOptions = ['CNY', 'USD', 'EUR', 'JPY', 'HKD', 'GBP'].map((currency) => ({
  label: currency,
  value: currency
}))

const cycleUnitOptions: Array<{ label: string; value: CycleUnit }> = [
  { label: '天', value: 'day' },
  { label: '月', value: 'month' },
  { label: '季度', value: 'quarter' },
  { label: '半年', value: 'half_year' },
  { label: '年', value: 'year' }
]

const rules: FormRules = {
  name: { required: true, message: '请输入订阅名称', trigger: ['blur', 'input'] },
  amount: { required: true, type: 'number', message: '请输入有效金额', trigger: ['blur', 'change'] },
  currency: { required: true, message: '请选择或输入币种', trigger: ['blur', 'change'] },
  start_date: { required: true, message: '请选择开始日期', trigger: ['blur', 'change'] },
  cycle_value: { required: true, type: 'number', message: '周期必须大于 0', trigger: ['blur', 'change'] },
  cycle_unit: { required: true, message: '请选择周期单位', trigger: ['blur', 'change'] },
  remind_days: { required: true, type: 'number', message: '请输入提醒天数', trigger: ['blur', 'change'] }
}

const cyclePreview = computed(() => getCycleText(formValue.cycle_value, formValue.cycle_unit))

/** 根据创建或编辑场景重置表单，避免上一次输入残留。 */
const resetForm = () => {
  const source: SubscriptionFormModel = props.subscription
    ? {
        name: props.subscription.name,
        amount: props.subscription.amount,
        currency: props.subscription.currency,
        start_date: props.subscription.start_date.slice(0, 10),
        cycle_value: props.subscription.cycle_value,
        cycle_unit: props.subscription.cycle_unit,
        auto_renew: props.subscription.auto_renew ?? false,
        remind_days: props.subscription.remind_days,
        remark: props.subscription.remark ?? ''
      }
    : createInitialForm()

  Object.assign(formValue, source)
  formRef.value?.restoreValidation()
}

/** 校验并创建或更新订阅，返回服务端最新记录。 */
const submit = async () => {
  try {
    await formRef.value?.validate()
  } catch {
    return
  }

  submitting.value = true
  const payload: SubscriptionWriteRequest = {
    name: formValue.name.trim(),
    amount: formValue.amount,
    currency: formValue.currency.trim().toUpperCase(),
    start_date: formValue.start_date,
    cycle_value: formValue.cycle_value,
    cycle_unit: formValue.cycle_unit,
    auto_renew: formValue.auto_renew,
    remind_days: formValue.remind_days,
    remark: formValue.remark.trim()
  }

  try {
    const response = props.subscription
      ? await subscriptionApi.update(props.subscription.id, payload)
      : await subscriptionApi.create(payload)

    message.success(props.subscription ? '订阅已更新' : '订阅已创建')
    emit('saved', response.data)
    emit('update:show', false)
  } catch (error: unknown) {
    message.error(getApiError(error, props.subscription ? '更新订阅失败' : '创建订阅失败'))
  } finally {
    submitting.value = false
  }
}

watch(() => [props.show, props.subscription] as const, ([show]) => {
  if (show) resetForm()
})
</script>

<style scoped>
.drawer-title {
  display: flex;
  align-items: center;
  gap: 12px;
}

.drawer-title__icon {
  display: grid;
  width: 42px;
  height: 42px;
  border-radius: 13px;
  place-items: center;
  color: var(--color-primary);
  background: var(--color-primary-soft);
  font-size: 20px;
}

.drawer-title > div {
  display: grid;
}

.drawer-title small {
  color: var(--color-text-muted);
  font-size: 9px;
  font-weight: 700;
  letter-spacing: 0.1em;
  text-transform: uppercase;
}

.drawer-title strong {
  margin-top: 3px;
  font-size: 17px;
}

.subscription-form {
  display: grid;
  gap: 18px;
}

.form-section {
  padding: 20px;
  border: 1px solid #e7eaf1;
  border-radius: 16px;
  background: #fbfcfe;
}

.form-section__heading {
  display: flex;
  margin-bottom: 20px;
  align-items: center;
  gap: 11px;
}

.form-section__heading > span {
  display: grid;
  width: 32px;
  height: 32px;
  border-radius: 10px;
  place-items: center;
  color: var(--color-primary);
  background: var(--color-primary-soft);
  font-size: 10px;
  font-weight: 760;
}

.form-section__heading > div {
  display: grid;
}

.form-section__heading strong {
  font-size: 13px;
}

.form-section__heading small {
  margin-top: 2px;
  color: var(--color-text-muted);
  font-size: 10px;
}

.form-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 14px;
}

.form-grid--amount {
  grid-template-columns: minmax(0, 1.4fr) minmax(130px, 0.6fr);
}

.form-section :deep(.n-input-number),
.form-section :deep(.n-date-picker) {
  width: 100%;
}

.cycle-preview,
.form-note {
  display: flex;
  padding: 11px 13px;
  border-radius: 11px;
  align-items: center;
  gap: 8px;
  color: #536078;
  background: #f0f3fa;
  font-size: 11px;
}

.cycle-preview .n-icon,
.form-note .n-icon {
  flex: 0 0 auto;
  color: var(--color-primary);
  font-size: 16px;
}

.switch-field {
  display: flex;
  min-height: 42px;
  align-items: center;
  gap: 10px;
}

.switch-field span {
  color: var(--color-text-muted);
  font-size: 10px;
  line-height: 1.45;
}

.form-note {
  margin-top: 2px;
  color: #8a520d;
  background: #fff6e9;
}

.form-note .n-icon {
  color: #b56500;
}

.drawer-actions {
  display: flex;
  width: 100%;
  justify-content: flex-end;
  gap: 10px;
}

@media (max-width: 560px) {
  .form-section {
    padding: 17px;
  }

  .form-grid,
  .form-grid--amount {
    grid-template-columns: 1fr;
    gap: 0;
  }

  .drawer-actions .n-button {
    flex: 1;
  }
}
</style>
