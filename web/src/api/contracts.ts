/** 订阅周期单位，与后端枚举值保持一致。 */
export type CycleUnit = 'day' | 'month' | 'quarter' | 'half_year' | 'year'

/** 登录请求。 */
export interface LoginRequest {
  username: string
  password: string
}

/** 登录成功响应。 */
export interface LoginResponse {
  token: string
}

/** 修改密码请求。 */
export interface PasswordChange {
  old_password: string
  new_password: string
}

/** 后端返回的订阅数据。 */
export interface Subscription {
  id?: number
  name: string
  amount: number
  currency: string
  start_date: string
  cycle_value: number
  cycle_unit: CycleUnit
  expire_date: string | null
  auto_renew?: boolean
  renew_count?: number
  remind_days: number
  remark?: string
  created_at?: string
  updated_at?: string
}

/** 已持久化且拥有确定 ID 的订阅数据。 */
export interface SubscriptionRecord extends Subscription {
  id: number
}

/** 创建或完整更新订阅时可提交的字段。 */
export type SubscriptionWriteRequest = Pick<
  Subscription,
  'name' | 'amount' | 'currency' | 'start_date' | 'cycle_value' | 'cycle_unit' | 'remind_days'
> &
  Partial<Pick<Subscription, 'auto_renew' | 'remark'>>

/** 更新订阅请求，支持后端已有的部分更新语义。 */
export type UpdateSubscriptionRequest = Partial<SubscriptionWriteRequest>

/** 单次续订历史记录。 */
export interface SubscriptionRenewal {
  id: number
  subscription_id: number
  renewed_at: string
  old_expire_date: string
  new_expire_date: string
  renew_count: number
}

/** 页面使用的通知设置。 */
export interface Settings {
  notify_hours: number[]
  telegram_bot_token: string
  telegram_chat_id: string
  bark_url: string
}

/** 后端设置格式，通知时段使用逗号分隔字符串。 */
export interface SettingsResponse {
  notify_hours: string
  telegram_bot_token: string
  telegram_chat_id: string
  bark_url: string
}

/** 支持的通知渠道。 */
export type NotificationType = 'telegram' | 'bark'

/** 公开站点配置。 */
export interface PublicConfig {
  website_title: string
}

/** 通用操作成功响应。 */
export interface MessageResponse {
  message: string
}
