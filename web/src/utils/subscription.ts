import type { CycleUnit, Subscription } from '../api'
import { normalizeCurrencyCode } from './currency'
import { getRemainingDays } from './date'

const DAYS_PER_MONTH = 365.2425 / 12

/** 订阅状态标识。 */
export type SubscriptionStatusKind = 'unknown' | 'expired' | 'due_today' | 'due_soon' | 'active'

/** 包含状态文案和剩余天数的展示模型。 */
export interface SubscriptionStatus {
  kind: SubscriptionStatusKind
  label: string
  daysRemaining: number | null
}

/** 按币种隔离的月均费用汇总，禁止跨币种相加。 */
export interface CurrencyMonthlySummary {
  currency: string
  monthlyAmount: number
  subscriptionCount: number
}

/** 根据到期日与提醒天数计算订阅状态。 */
export const getSubscriptionStatus = (
  subscription: Pick<Subscription, 'expire_date' | 'remind_days'>,
  today: Date = new Date()
): SubscriptionStatus => {
  const daysRemaining = getRemainingDays(subscription.expire_date, today)
  if (daysRemaining === null) {
    return { kind: 'unknown', label: '日期未知', daysRemaining: null }
  }
  if (daysRemaining < 0) {
    return { kind: 'expired', label: '已过期', daysRemaining }
  }
  if (daysRemaining === 0) {
    return { kind: 'due_today', label: '今日到期', daysRemaining }
  }

  const remindDays = Number.isFinite(subscription.remind_days)
    ? Math.max(0, subscription.remind_days)
    : 0
  if (daysRemaining <= remindDays) {
    return { kind: 'due_soon', label: '即将到期', daysRemaining }
  }
  return { kind: 'active', label: '正常', daysRemaining }
}

/** 将周期配置转换为自然中文。 */
export const getCycleText = (cycleValue: number, cycleUnit: CycleUnit): string => {
  if (!Number.isInteger(cycleValue) || cycleValue <= 0) return '周期配置异常'

  if (cycleValue === 1) {
    const singularLabels: Record<CycleUnit, string> = {
      day: '每天',
      month: '每月',
      quarter: '每季度',
      half_year: '每半年',
      year: '每年'
    }
    return singularLabels[cycleUnit]
  }

  const unitLabels: Record<CycleUnit, string> = {
    day: '天',
    month: '个月',
    quarter: '个季度',
    half_year: '个半年',
    year: '年'
  }
  return `每 ${cycleValue} ${unitLabels[cycleUnit]}`
}

/**
 * 将单次费用按周期折算为估算月均费用。
 * @returns 配置非法时返回 null，避免把异常历史数据静默计为 0。
 */
export const calculateMonthlyEquivalent = (
  amount: number,
  cycleValue: number,
  cycleUnit: CycleUnit
): number | null => {
  if (!Number.isFinite(amount) || amount < 0 || !Number.isFinite(cycleValue) || cycleValue <= 0) {
    return null
  }

  switch (cycleUnit) {
    case 'day':
      return amount * (DAYS_PER_MONTH / cycleValue)
    case 'month':
      return amount / cycleValue
    case 'quarter':
      return amount / (cycleValue * 3)
    case 'half_year':
      return amount / (cycleValue * 6)
    case 'year':
      return amount / (cycleValue * 12)
  }
}

/** 计算单个订阅的估算月均费用。 */
export const getSubscriptionMonthlyCost = (
  subscription: Pick<Subscription, 'amount' | 'cycle_value' | 'cycle_unit'>
): number | null => {
  return calculateMonthlyEquivalent(
    subscription.amount,
    subscription.cycle_value,
    subscription.cycle_unit
  )
}

/**
 * 按币种汇总订阅月均费用。
 * 不执行汇率换算，每种币种始终生成独立汇总项。
 */
export const summarizeMonthlyCostsByCurrency = (
  subscriptions: readonly Pick<
    Subscription,
    'amount' | 'currency' | 'cycle_value' | 'cycle_unit'
  >[]
): CurrencyMonthlySummary[] => {
  const summaries = new Map<string, CurrencyMonthlySummary>()

  for (const subscription of subscriptions) {
    const monthlyAmount = getSubscriptionMonthlyCost(subscription)
    if (monthlyAmount === null) continue

    const currency = normalizeCurrencyCode(subscription.currency)
    const current = summaries.get(currency)
    if (current) {
      current.monthlyAmount += monthlyAmount
      current.subscriptionCount += 1
    } else {
      summaries.set(currency, {
        currency,
        monthlyAmount,
        subscriptionCount: 1
      })
    }
  }

  return [...summaries.values()].sort((left, right) =>
    left.currency.localeCompare(right.currency)
  )
}
