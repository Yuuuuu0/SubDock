const CURRENCY_SYMBOLS: Readonly<Record<string, string>> = {
  CNY: '¥',
  USD: '$',
  EUR: '€',
  JPY: 'JP¥',
  HKD: 'HK$',
  GBP: '£'
}

/** 标准化币种代码，空值统一归入 UNKNOWN，避免与合法币种混算。 */
export const normalizeCurrencyCode = (currency: string): string => {
  const normalized = currency.trim().toUpperCase()
  return normalized || 'UNKNOWN'
}

/** 获取常见币种符号，未知币种回退到标准化代码。 */
export const getCurrencySymbol = (currency: string): string => {
  const code = normalizeCurrencyCode(currency)
  return CURRENCY_SYMBOLS[code] ?? code
}

/**
 * 使用 Intl 格式化金额，非法或非标准币种使用稳定文本回退。
 * @param amount 金额。
 * @param currency ISO 4217 币种代码。
 * @param locale 展示区域，默认中文环境。
 */
export const formatCurrency = (
  amount: number,
  currency: string,
  locale = 'zh-CN'
): string => {
  if (!Number.isFinite(amount)) return '—'

  const code = normalizeCurrencyCode(currency)
  try {
    return new Intl.NumberFormat(locale, {
      style: 'currency',
      currency: code,
      currencyDisplay: 'narrowSymbol'
    }).format(amount)
  } catch {
    return `${getCurrencySymbol(code)} ${amount.toFixed(2)}`
  }
}
