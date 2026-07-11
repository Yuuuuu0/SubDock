const DATE_PREFIX_PATTERN = /^(\d{4})-(\d{2})-(\d{2})(?:$|[T\s])/
const MILLISECONDS_PER_DAY = 24 * 60 * 60 * 1000

/** 将 Date 规整为本地日历日，移除时分秒影响。 */
export const startOfLocalDay = (date: Date): Date => {
  return new Date(date.getFullYear(), date.getMonth(), date.getDate())
}

/**
 * 按本地日历语义解析后端日期。
 * 对 ISO 时间只读取 YYYY-MM-DD 部分，避免浏览器按 UTC 解析导致日期偏移。
 */
export const parseLocalDate = (value: string | Date | null | undefined): Date | null => {
  if (value instanceof Date) {
    if (Number.isNaN(value.getTime())) return null
    return startOfLocalDay(value)
  }

  if (!value) return null
  const match = DATE_PREFIX_PATTERN.exec(value.trim())
  if (!match) return null

  const year = Number(match[1])
  const monthIndex = Number(match[2]) - 1
  const day = Number(match[3])
  const date = new Date(year, monthIndex, day)

  if (
    date.getFullYear() !== year ||
    date.getMonth() !== monthIndex ||
    date.getDate() !== day
  ) {
    return null
  }

  return date
}

/** 将日期格式化为本地 YYYY-MM-DD，适用于日期表单默认值。 */
export const formatLocalDateKey = (date: Date = new Date()): string => {
  const year = String(date.getFullYear()).padStart(4, '0')
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

/**
 * 计算目标日期相对今天的日历天数。
 * @returns 未来为正数、今天为 0、过去为负数；非法日期返回 null。
 */
export const getRemainingDays = (
  value: string | Date | null | undefined,
  today: Date = new Date()
): number | null => {
  const target = parseLocalDate(value)
  if (!target || Number.isNaN(today.getTime())) return null

  const targetTimestamp = Date.UTC(target.getFullYear(), target.getMonth(), target.getDate())
  const todayTimestamp = Date.UTC(today.getFullYear(), today.getMonth(), today.getDate())
  return Math.round((targetTimestamp - todayTimestamp) / MILLISECONDS_PER_DAY)
}
