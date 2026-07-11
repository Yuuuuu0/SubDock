import type { GlobalThemeOverrides } from 'naive-ui'

/** SubDock 的 Naive UI 主题配置，统一颜色、圆角和控件尺寸。 */
export const themeOverrides: GlobalThemeOverrides = {
  common: {
    primaryColor: '#3659E3',
    primaryColorHover: '#4B6BE8',
    primaryColorPressed: '#2947BC',
    primaryColorSuppl: '#4B6BE8',
    infoColor: '#3659E3',
    successColor: '#087F5B',
    warningColor: '#C76A00',
    errorColor: '#C4322B',
    bodyColor: '#F7F8FC',
    cardColor: '#FFFFFF',
    modalColor: '#FFFFFF',
    popoverColor: '#FFFFFF',
    textColorBase: '#182033',
    textColor1: '#182033',
    textColor2: '#4E5A70',
    textColor3: '#667085',
    borderColor: '#E4E8F0',
    dividerColor: '#E9ECF2',
    borderRadius: '10px',
    fontFamily:
      'Inter, ui-sans-serif, -apple-system, BlinkMacSystemFont, "Segoe UI", "PingFang SC", "Hiragino Sans GB", "Microsoft YaHei", sans-serif'
  },
  Button: {
    borderRadiusMedium: '10px',
    borderRadiusLarge: '12px',
    heightMedium: '42px',
    heightLarge: '48px',
    fontWeight: '600',
    textColorPrimary: '#FFFFFF'
  },
  Input: {
    borderRadius: '10px',
    heightMedium: '42px'
  },
  InputNumber: {
    borderRadius: '10px',
    heightMedium: '42px'
  },
  Select: {
    peers: {
      InternalSelection: {
        borderRadius: '10px',
        heightMedium: '42px'
      }
    }
  },
  Card: {
    borderRadius: '16px',
    paddingMedium: '22px 24px'
  },
  DataTable: {
    borderRadius: '14px',
    thColor: '#F8F9FC',
    thTextColor: '#667085',
    tdColorHover: '#F8FAFF'
  },
  Drawer: {
    color: '#FFFFFF'
  },
  Modal: {
    borderRadius: '18px'
  },
  Tag: {
    borderRadius: '999px'
  },
  Dialog: {
    borderRadius: '16px'
  }
}
