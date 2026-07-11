import apiClient from './client'
import type { PublicConfig } from './contracts'

/** 公开配置接口。 */
export const configApi = {
  /** 获取无需登录即可读取的站点配置。 */
  get() {
    return apiClient.get<PublicConfig>('/config')
  }
}
