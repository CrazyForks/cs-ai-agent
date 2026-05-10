"use client"

import {
  HeadphonesIcon,
  Maximize2Icon,
  Minimize2Icon,
  MinusIcon,
  RotateCwIcon,
  XIcon,
} from "lucide-react"
import {
  useCallback,
  useEffect,
  useLayoutEffect,
  useRef,
  useState,
  type ComponentProps,
  type CSSProperties,
} from "react"
import { useShallow } from "zustand/react/shallow"

import { KefuConnectionStatus } from "@/components/kefu/connection-status"
import { getStandaloneClosedUrl } from "@/components/kefu/close-navigation"
import { CustomerMessageEditor } from "@/components/kefu/customer-message-editor"
import {
  KefuMessageList,
  type KefuMessageListHandle,
} from "@/components/kefu/message-list"
import {
  bindKefuHostBridge,
  requestKefuHostClose,
  requestKefuHostMinimize,
  requestKefuHostToggleMaximize,
} from "@/lib/kefu-host-bridge"
import { useKefuChatStore } from "@/lib/stores/kefu-chat"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { cn } from "@/lib/utils"

const windowActionButtonClass =
  "size-8 rounded-md border-0 bg-transparent text-muted-foreground shadow-none hover:bg-foreground/[0.06] hover:text-foreground focus-visible:ring-primary/20 dark:hover:bg-white/10"

function WindowActionButton({
  className,
  ...props
}: ComponentProps<typeof Button>) {
  return (
    <Button
      type="button"
      variant="ghost"
      size="icon"
      className={cn(windowActionButtonClass, className)}
      {...props}
    />
  )
}

function useKefuSystemTheme() {
  useLayoutEffect(() => {
    if (typeof window === "undefined") {
      return
    }
    const root = document.documentElement
    const query = window.matchMedia("(prefers-color-scheme: dark)")
    const previousDarkClass = root.classList.contains("dark")
    const previousColorScheme = root.style.colorScheme

    const syncTheme = () => {
      const isDark = query.matches
      root.classList.toggle("dark", isDark)
      root.style.colorScheme = isDark ? "dark" : "light"
    }

    syncTheme()
    query.addEventListener("change", syncTheme)

    return () => {
      query.removeEventListener("change", syncTheme)
      root.classList.toggle("dark", previousDarkClass)
      root.style.colorScheme = previousColorScheme
    }
  }, [])
}

function isEmbeddedInHost() {
  if (typeof window === "undefined") {
    return false
  }

  try {
    return window.parent !== window
  } catch {
    return false
  }
}

export function KefuChatShell() {
  useKefuSystemTheme()

  const messageListRef = useRef<KefuMessageListHandle | null>(null)
  const [isEmbedded, setIsEmbedded] = useState(false)
  const [isMaximized, setIsMaximized] = useState(false)
  const [isCloseDialogOpen, setIsCloseDialogOpen] = useState(false)
  const [isClosingConversation, setIsClosingConversation] = useState(false)

  const {
    title,
    subtitle,
    themeColor,
    conversation,
    messages,
    messagesHasMore,
    messagesLoadingMore,
    loadOlderMessages,
    status,
    error,
    isOpen,
    isVisible,
    setIsOpen,
    setIsVisible,
    bootstrap,
    handleSendMessage,
    uploadMessageImage,
    sendAttachment,
    retry,
    disconnectSocket,
    markConversationRead,
    closeConversation,
  } = useKefuChatStore(
    useShallow((state) => ({
      title: state.title,
      subtitle: state.subtitle,
      themeColor: state.themeColor,
      conversation: state.conversation,
      messages: state.messages,
      messagesHasMore: state.messagesHasMore,
      messagesLoadingMore: state.messagesLoadingMore,
      loadOlderMessages: state.loadOlderMessages,
      status: state.status,
      error: state.error,
      isOpen: state.isOpen,
      isVisible: state.isVisible,
      setIsOpen: state.setIsOpen,
      setIsVisible: state.setIsVisible,
      bootstrap: state.bootstrap,
      handleSendMessage: state.handleSendMessage,
      uploadMessageImage: state.uploadMessageImage,
      sendAttachment: state.sendAttachment,
      retry: state.retry,
      disconnectSocket: state.disconnectSocket,
      markConversationRead: state.markConversationRead,
      closeConversation: state.closeConversation,
    }))
  )
  const safeMessages = Array.isArray(messages) ? messages : []

  useEffect(() => {
    setIsEmbedded(isEmbeddedInHost())
  }, [])

  const maybeMarkConversationRead = useCallback(() => {
    if (!isVisible || !conversation || typeof document === "undefined") {
      return
    }
    if (document.visibilityState !== "visible") {
      return
    }
    void markConversationRead().catch((readError) => {
      console.error("Failed to mark kefu conversation read", readError)
    })
  }, [conversation?.id, isVisible, markConversationRead])

  useEffect(() => {
    return bindKefuHostBridge({
      onOpen: () => {
        setIsOpen(true)
        setIsVisible(true)
      },
      onMinimize: () => {
        setIsVisible(false)
      },
      onMaximizedChange: (nextIsMaximized) => {
        setIsMaximized(nextIsMaximized)
      },
    })
  }, [setIsOpen, setIsVisible])

  useEffect(() => {
    bootstrap()

    return () => {
      if (!isOpen) {
        disconnectSocket()
      }
    }
  }, [isOpen, bootstrap, disconnectSocket])

  useEffect(() => {
    maybeMarkConversationRead()
  }, [maybeMarkConversationRead, safeMessages.length])

  useEffect(() => {
    const handleVisibilityChange = () => {
      if (document.visibilityState === "visible") {
        maybeMarkConversationRead()
      }
    }
    const handleFocus = () => {
      maybeMarkConversationRead()
    }

    document.addEventListener("visibilitychange", handleVisibilityChange)
    window.addEventListener("focus", handleFocus)
    return () => {
      document.removeEventListener("visibilitychange", handleVisibilityChange)
      window.removeEventListener("focus", handleFocus)
    }
  }, [maybeMarkConversationRead])

  async function handleSend(content: string) {
    await handleSendMessage(content)
    messageListRef.current?.scrollToBottom()
  }

  function handleMinimize() {
    setIsVisible(false)
    requestKefuHostMinimize()
  }

  function handleToggleMaximize() {
    requestKefuHostToggleMaximize()
  }

  async function confirmCloseConversation() {
    if (isClosingConversation) {
      return
    }
    setIsClosingConversation(true)
    try {
      if (conversation?.id) {
        await closeConversation()
      }
      setIsCloseDialogOpen(false)
      if (isEmbedded) {
        requestKefuHostClose()
      } else {
        window.location.replace(getStandaloneClosedUrl())
      }
    } catch (closeError) {
      window.alert(closeError instanceof Error ? closeError.message : "关闭会话失败")
    } finally {
      setIsClosingConversation(false)
    }
  }

  useEffect(() => {
    if (!isCloseDialogOpen) {
      return
    }
    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape" && !isClosingConversation) {
        setIsCloseDialogOpen(false)
      }
    }
    window.addEventListener("keydown", handleKeyDown)
    return () => window.removeEventListener("keydown", handleKeyDown)
  }, [isCloseDialogOpen, isClosingConversation])

  return (
    <main
      className="relative flex h-[100dvh] min-h-[100dvh] overflow-hidden bg-muted text-foreground supports-not-[height:100dvh]:h-screen supports-not-[height:100dvh]:min-h-screen"
      style={{ "--primary": themeColor } as CSSProperties}
    >
      <section className="flex h-full w-full flex-col overflow-hidden bg-card text-card-foreground">
        <header className="shrink-0 border-b border-primary/[0.10] bg-primary/[0.06] px-3 py-2.5 shadow-[0_10px_26px_rgba(15,23,42,0.06)] dark:border-primary/20 dark:bg-primary/10 dark:shadow-none sm:px-4 sm:py-3">
          <div className="flex min-w-0 items-center justify-between gap-2 sm:gap-3">
            <div className="flex min-w-0 items-center gap-2.5">
              <div className="flex size-9 shrink-0 items-center justify-center rounded-lg bg-primary text-primary-foreground shadow-[0_8px_18px_rgba(37,99,235,0.18)]">
                <HeadphonesIcon className="size-[18px]" />
              </div>
              <div className="min-w-0">
                <div className="truncate text-base font-semibold text-foreground">
                  {title}
                </div>
                <div className="mt-0.5 truncate text-xs text-muted-foreground sm:mt-1">
                  {subtitle}
                </div>
              </div>
            </div>
            <div className="flex shrink-0 items-center gap-1 sm:gap-2">
              {status !== "connected" ? (
                <KefuConnectionStatus status={status} />
              ) : null}
              <div className="flex items-center gap-0.5 rounded-lg bg-background/55 p-0.5 shadow-sm ring-1 ring-border/70 dark:bg-background/25 dark:ring-white/10">
                <WindowActionButton
                  onClick={retry}
                  aria-label="重新连接"
                  title="重新连接"
                >
                  <RotateCwIcon className="size-4" />
                </WindowActionButton>
                {isEmbedded ? (
                  <>
                    <WindowActionButton
                      onClick={handleMinimize}
                      aria-label="收起聊天窗口"
                      title="收起聊天窗口"
                    >
                      <MinusIcon className="size-4" />
                    </WindowActionButton>
                    <WindowActionButton
                      onClick={handleToggleMaximize}
                      aria-label={isMaximized ? "取消最大化" : "最大化聊天窗口"}
                      title={isMaximized ? "取消最大化" : "最大化聊天窗口"}
                    >
                      {isMaximized ? (
                        <Minimize2Icon className="size-4" />
                      ) : (
                        <Maximize2Icon className="size-4" />
                      )}
                    </WindowActionButton>
                  </>
                ) : null}
                <WindowActionButton
                  onClick={() => setIsCloseDialogOpen(true)}
                  aria-label="关闭聊天窗口"
                  title="关闭聊天窗口"
                  className="hover:bg-rose-50 hover:text-rose-600 dark:hover:bg-rose-950/45 dark:hover:text-rose-300"
                >
                  <XIcon className="size-4" />
                </WindowActionButton>
              </div>
            </div>
          </div>
        </header>

        <div className="grid min-h-0 flex-1 grid-rows-[minmax(0,1fr)_auto] overflow-hidden bg-muted/60 dark:bg-muted/30">
          <KefuMessageList
            ref={messageListRef}
            messages={safeMessages}
            onNearBottomVisible={maybeMarkConversationRead}
            hasMoreOlder={messagesHasMore}
            loadingOlder={messagesLoadingMore}
            onLoadOlder={loadOlderMessages}
          />
          <div className="shrink-0 border-t border-border/80 bg-card/95 pb-[env(safe-area-inset-bottom)] shadow-[0_-10px_24px_rgba(15,23,42,0.05)] dark:bg-card/90 dark:shadow-none">
            <CustomerMessageEditor
              disabled={!conversation}
              onSend={handleSend}
              onUploadImage={uploadMessageImage}
              onSendAttachment={sendAttachment}
            />
          </div>
        </div>

        {error ? (
          <div className="border-t border-destructive/20 bg-destructive/10 px-4 py-3 text-sm text-destructive">
            {error}
          </div>
        ) : null}
      </section>

      <Dialog
        open={isCloseDialogOpen}
        onOpenChange={(open) => {
          if (!isClosingConversation) {
            setIsCloseDialogOpen(open)
          }
        }}
      >
        <DialogContent className="max-w-[320px]" showCloseButton={!isClosingConversation}>
          <DialogHeader>
            <DialogTitle>结束当前对话？</DialogTitle>
            <DialogDescription className="text-xs leading-5">
              结束会话，客服将无法再查看您的消息记录，如需再次联系请重新发起对话。
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              disabled={isClosingConversation}
              onClick={() => setIsCloseDialogOpen(false)}
            >
              继续对话
            </Button>
            <Button
              type="button"
              variant="destructive"
              disabled={isClosingConversation}
              onClick={() => void confirmCloseConversation()}
            >
              {isClosingConversation ? "结束中..." : "确认结束"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </main>
  )
}
