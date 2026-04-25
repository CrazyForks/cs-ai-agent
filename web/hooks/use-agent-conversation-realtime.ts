"use client"

import { useEffect, useRef } from "react"
import { toast } from "sonner"

import { createAdminWebSocketUrl } from "@/lib/api/admin"
import { readSession } from "@/lib/auth"
import { createRealtimeConnectionManager } from "@/lib/realtime-connection"
import { getNotificationBody, showNotification } from "@/lib/services/notification"
import { useAgentConversationsStore } from "@/lib/stores/agent-conversations"

type AgentRealtimeConnection = ReturnType<typeof createRealtimeConnectionManager>

export function useAgentConversationRealtime() {
  const selectedConversationId = useAgentConversationsStore(
    (state) => state.selectedConversationId
  )
  const loadConversations = useAgentConversationsStore(
    (state) => state.loadConversations
  )
  const syncLatestMessages = useAgentConversationsStore(
    (state) => state.syncLatestMessages
  )
  const setRealtimeStatus = useAgentConversationsStore(
    (state) => state.setRealtimeStatus
  )
  const realtimeRef = useRef<AgentRealtimeConnection | null>(null)
  const subscribedConversationIdRef = useRef<number | null>(null)
  const selectedConversationIdRef = useRef<number | null>(selectedConversationId)
  const currentUserIdRef = useRef<number>(readSession()?.user.id ?? 0)

  useEffect(() => {
    selectedConversationIdRef.current = selectedConversationId
  }, [selectedConversationId])

  useEffect(() => {
    const realtime = createRealtimeConnectionManager({
      createSocket: () => new WebSocket(createAdminWebSocketUrl()),
      onStatusChange: setRealtimeStatus,
      onOpen: (socket) => {
        console.info("[agent-realtime] websocket connected", {
          url: socket.url,
        })

        const conversationId = selectedConversationIdRef.current
        if (conversationId) {
          socket.send(
            JSON.stringify({
              type: "subscribe",
              topics: [`conversation:${conversationId}`],
            })
          )
          subscribedConversationIdRef.current = conversationId
        } else {
          subscribedConversationIdRef.current = null
        }
      },
      onMessage: (event, socket) => {
        try {
          const payload = JSON.parse(event.data) as {
            eventId?: string
            type?: string
            data?: {
              conversationId?: number
              messageId?: number
              status?: number
              currentAssigneeId?: number
              senderType?: string
              messageType?: string
              content?: string
            }
          }
          const eventType = payload.type ?? ""
          const conversationId = payload.data?.conversationId ?? 0
          const eventId = payload.eventId?.trim() ?? ""

          if (
            eventType === "" ||
            eventType === "connected" ||
            eventType === "pong" ||
            eventType === "subscribed" ||
            eventType === "unsubscribed"
          ) {
            return
          }

          if (eventId && socket.readyState === WebSocket.OPEN) {
            socket.send(JSON.stringify({ type: "ack", eventId }))
          }

          if (eventType === "message.created" && conversationId > 0) {
            const senderType = payload.data?.senderType ?? ""
            const status = payload.data?.status ?? 0
            const currentAssigneeId = payload.data?.currentAssigneeId ?? 0

            void loadConversations()
              .then(() => {
                const store = useAgentConversationsStore.getState()
                const shouldNotify =
                  senderType === "customer" &&
                  status === 2 &&
                  currentAssigneeId > 0 &&
                  currentAssigneeId === currentUserIdRef.current &&
                  typeof document !== "undefined" &&
                  document.visibilityState !== "visible"

                if (!shouldNotify) {
                  return
                }

                showNotification(
                  "新消息",
                  getNotificationBody({
                    messageType: payload.data?.messageType ?? "",
                    content: payload.data?.content ?? "",
                  }),
                  () => {
                    void store.selectConversation(conversationId)
                  }
                )
              })
              .catch((error) => {
                toast.error(error instanceof Error ? error.message : "加载消息失败")
              })
          } else {
            void loadConversations().catch((error) => {
              toast.error(error instanceof Error ? error.message : "加载会话列表失败")
            })
          }

          if (
            conversationId > 0 &&
            selectedConversationIdRef.current === conversationId
          ) {
            void syncLatestMessages(conversationId)
          }
        } catch {
          // ignore invalid ws payload
        }
      },
      onClose: (event, socket) => {
        console.log("[agent-realtime] websocket closed", {
          url: socket.url,
          readyState: socket.readyState,
          code: event.code,
          reason: event.reason,
          wasClean: event.wasClean,
        })
        subscribedConversationIdRef.current = null
      },
      onError: (_event, socket) => {
        console.log("[agent-realtime] websocket error", {
          url: socket.url,
          readyState: socket.readyState,
        })
      },
      onConnectError: (error) => {
        toast.error(error instanceof Error ? error.message : "连接实时服务失败")
      },
    })

    realtimeRef.current = realtime
    realtime.connect()

    return () => {
      realtimeRef.current = null
      realtime.disconnect()
      subscribedConversationIdRef.current = null
    }
  }, [loadConversations, setRealtimeStatus, syncLatestMessages])

  useEffect(() => {
    const socket = realtimeRef.current?.getSocket()
    if (!socket || socket.readyState !== WebSocket.OPEN) {
      return
    }

    const previousConversationId = subscribedConversationIdRef.current
    const nextConversationId = selectedConversationId ?? null

    if (previousConversationId && previousConversationId !== nextConversationId) {
      socket.send(
        JSON.stringify({
          type: "unsubscribe",
          topics: [`conversation:${previousConversationId}`],
        })
      )
    }

    if (nextConversationId && nextConversationId !== previousConversationId) {
      socket.send(
        JSON.stringify({
          type: "subscribe",
          topics: [`conversation:${nextConversationId}`],
        })
      )
      subscribedConversationIdRef.current = nextConversationId
      return
    }

    if (!nextConversationId) {
      subscribedConversationIdRef.current = null
    }
  }, [selectedConversationId])
}
