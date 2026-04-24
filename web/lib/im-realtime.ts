import { createWebSocketBaseUrl } from "@/lib/api/websocket"
import { getImVisitorId } from "@/lib/api/im"

const OPEN_IM_CHANNEL_ID =
  process.env.NEXT_PUBLIC_OPEN_IM_CHANNEL_ID?.trim() || ""
const OPEN_IM_EXTERNAL_SOURCE =
  process.env.NEXT_PUBLIC_OPEN_IM_EXTERNAL_SOURCE?.trim() || "web_chat"

export type ImRealtimeEnvelope = {
  type: string
  topic?: string
  data?: {
    conversationId?: number
    messageId?: number
  }
  payload?: {
    conversationId?: number
    messageId?: number
  }
}

export function createImRealtimeConnection() {
  const baseUrl = createWebSocketBaseUrl()
  const externalId = encodeURIComponent(getImVisitorId())
  const externalSource = encodeURIComponent(OPEN_IM_EXTERNAL_SOURCE)
  const channelId = encodeURIComponent(OPEN_IM_CHANNEL_ID)
  return new WebSocket(
    `${baseUrl}/api/open/im/ws?externalId=${externalId}&externalSource=${externalSource}&channelId=${channelId}`
  )
}
