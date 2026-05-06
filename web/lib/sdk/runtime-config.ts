import type { KefuChatRuntimeConfig } from "@/lib/sdk/config-types"

export function readKefuChatRuntimeConfig(): KefuChatRuntimeConfig {
  if (typeof window === "undefined") {
    return {
      channelId: "",
      baseUrl: "",
      apiBaseUrl: "",
    }
  }

  const query = new URLSearchParams(window.location.search)
  const fallback: KefuChatRuntimeConfig = {
    channelId:
      query.get("channelId") ??
      process.env.NEXT_PUBLIC_OPEN_IM_CHANNEL_ID?.trim() ??
      "",
    baseUrl:
      query.get("baseUrl") ??
      process.env.NEXT_PUBLIC_API_BASE_URL?.trim() ??
      window.location.origin,
    apiBaseUrl:
      query.get("apiBaseUrl") ??
      process.env.NEXT_PUBLIC_API_BASE_URL?.trim() ??
      undefined,
    externalId: query.get("externalId") ?? undefined,
    externalName: query.get("externalName") ?? undefined,
    userToken: query.get("userToken") ?? undefined,
    title: query.get("title") ?? undefined,
    subtitle: query.get("subtitle") ?? undefined,
    position: (query.get("position") as "left" | "right" | null) ?? undefined,
    themeColor: query.get("themeColor") ?? undefined,
    width: query.get("width") ?? undefined,
  }

  if (window.__CS_AGENT_WIDGET_CONFIG__) {
    return window.__CS_AGENT_WIDGET_CONFIG__
  }
  if (window.CSAgentConfig) {
    const { getUserToken: _getUserToken, ...hostConfig } = window.CSAgentConfig
    return {
      ...fallback,
      ...hostConfig,
    }
  }
  return fallback
}

export function setKefuChatRuntimeConfig(config: KefuChatRuntimeConfig) {
  if (typeof window === "undefined") {
    return
  }
  window.__CS_AGENT_WIDGET_CONFIG__ = config
}
