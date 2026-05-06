export type CSAgentConfig = {
  channelId: string
  baseUrl?: string
  apiBaseUrl?: string
  widgetBaseUrl?: string
  /** 外部访客稳定标识；未传时使用浏览器本地访客 ID */
  externalId?: string
  /** 访客展示名，仅用于首次换取客服会话 token */
  externalName?: string
  /** 打开客服前按需获取业务系统签发的前台用户 JWT */
  getUserToken?: () => string | Promise<string>
  title?: string
  subtitle?: string
  position?: "left" | "right"
  themeColor?: string
  width?: string
}

export type KefuChatRuntimeConfig = Omit<CSAgentConfig, "getUserToken"> & {
  /** 仅用于 /kefu/chat 运行时换取客服会话 token，不属于 CSAgentConfig 接入参数 */
  userToken?: string
}

export type CSAgentWidget = {
  mount: (config?: CSAgentConfig) => void
  destroy: () => void
  open: () => Promise<void>
  close: () => void
  getChatUrl: () => Promise<string>
}

declare global {
  interface Window {
    CSAgentConfig?: CSAgentConfig
    CSAgentWidget?: CSAgentWidget
    __CS_AGENT_WIDGET_CONFIG__?: KefuChatRuntimeConfig
    __CS_AGENT_WIDGET_STATE__?: unknown
  }
}
