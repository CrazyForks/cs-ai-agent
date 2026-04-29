"use client"

import { CalendarPlusIcon, GripVerticalIcon } from "lucide-react"

import type {
  AdminAgentTeam,
  AdminAgentTeamSchedule,
  CreateAdminAgentTeamSchedulePayload,
  UpdateAdminAgentTeamSchedulePayload,
} from "@/lib/api/admin"
import { cn, formatDateTime } from "@/lib/utils"

const dayNames = ["一", "二", "三", "四", "五", "六", "日"]
const dayMs = 24 * 60 * 60 * 1000
const minuteMs = 60 * 1000
const minDurationMs = 15 * minuteMs

type ScheduleCalendarProps = {
  weekStart: Date
  teams: AdminAgentTeam[]
  schedules: AdminAgentTeamSchedule[]
  loading: boolean
  savingId: number | null
  onCreate: (defaults: Partial<CreateAdminAgentTeamSchedulePayload>) => void
  onEdit: (item: AdminAgentTeamSchedule) => void
  onMove: (payload: UpdateAdminAgentTeamSchedulePayload) => Promise<void>
  onResize: (payload: UpdateAdminAgentTeamSchedulePayload) => Promise<void>
}

type DragState =
  | {
      type: "move"
      item: AdminAgentTeamSchedule
      startX: number
      startY: number
      moved: boolean
    }
  | {
      type: "resize"
      edge: "start" | "end"
      item: AdminAgentTeamSchedule
      moved: boolean
    }

function addDays(date: Date, days: number) {
  const ret = new Date(date)
  ret.setDate(ret.getDate() + days)
  return ret
}

function startOfDay(date: Date) {
  const ret = new Date(date)
  ret.setHours(0, 0, 0, 0)
  return ret
}

function parseLocalDateTime(value: string) {
  const match = value.match(/^(\d{4})-(\d{2})-(\d{2})[ T](\d{2}):(\d{2})(?::(\d{2}))?/)
  if (!match) {
    return new Date(value)
  }
  return new Date(
    Number(match[1]),
    Number(match[2]) - 1,
    Number(match[3]),
    Number(match[4]),
    Number(match[5]),
    Number(match[6] ?? 0)
  )
}

function formatDate(date: Date) {
  const month = String(date.getMonth() + 1).padStart(2, "0")
  const day = String(date.getDate()).padStart(2, "0")
  return `${date.getFullYear()}-${month}-${day}`
}

function formatDateTimeValue(date: Date) {
  const month = String(date.getMonth() + 1).padStart(2, "0")
  const day = String(date.getDate()).padStart(2, "0")
  const hour = String(date.getHours()).padStart(2, "0")
  const minute = String(date.getMinutes()).padStart(2, "0")
  const second = String(date.getSeconds()).padStart(2, "0")
  return `${date.getFullYear()}-${month}-${day} ${hour}:${minute}:${second}`
}

function formatDayTitle(date: Date) {
  return `${date.getMonth() + 1}/${date.getDate()}`
}

function clamp(value: number, min: number, max: number) {
  return Math.min(Math.max(value, min), max)
}

function roundToQuarterHour(date: Date) {
  const ret = new Date(date)
  ret.setSeconds(0, 0)
  const minutes = ret.getHours() * 60 + ret.getMinutes()
  const rounded = Math.round(minutes / 15) * 15
  ret.setHours(Math.floor(rounded / 60), rounded % 60, 0, 0)
  return ret
}

function getPointerDateInCell(event: PointerEvent | React.PointerEvent, cell: Element) {
  const rect = cell.getBoundingClientRect()
  const day = startOfDay(parseLocalDateTime(`${cell.getAttribute("data-date")} 00:00:00`))
  const ratio = clamp((event.clientX - rect.left) / rect.width, 0, 1)
  return roundToQuarterHour(new Date(day.getTime() + ratio * dayMs))
}

function getDropCell(event: PointerEvent | React.PointerEvent) {
  const element = document.elementFromPoint(event.clientX, event.clientY)
  return element?.closest("[data-schedule-cell]")
}

function getCellTeamAndDate(cell: Element) {
  const teamID = Number(cell.getAttribute("data-team-id"))
  const date = cell.getAttribute("data-date") ?? ""
  return { teamID, date }
}

function buildMovePayload(item: AdminAgentTeamSchedule, teamId: number, date: string): UpdateAdminAgentTeamSchedulePayload {
  const originalStart = parseLocalDateTime(item.startAt)
  const originalEnd = parseLocalDateTime(item.endAt)
  const duration = originalEnd.getTime() - originalStart.getTime()
  const nextDay = startOfDay(parseLocalDateTime(`${date} 00:00:00`))
  const nextStart = new Date(nextDay)
  nextStart.setHours(originalStart.getHours(), originalStart.getMinutes(), originalStart.getSeconds(), 0)
  const nextEnd = new Date(nextStart.getTime() + duration)

  return {
    id: item.id,
    teamId,
    startAt: formatDateTimeValue(nextStart),
    endAt: formatDateTimeValue(nextEnd),
    sourceType: item.sourceType,
    remark: item.remark,
  }
}

function buildResizePayload(
  item: AdminAgentTeamSchedule,
  edge: "start" | "end",
  nextTime: Date
): UpdateAdminAgentTeamSchedulePayload | null {
  const startAt = parseLocalDateTime(item.startAt)
  const endAt = parseLocalDateTime(item.endAt)
  if (edge === "start") {
    if (endAt.getTime() - nextTime.getTime() < minDurationMs) {
      return null
    }
    startAt.setTime(nextTime.getTime())
  } else {
    if (nextTime.getTime() - startAt.getTime() < minDurationMs) {
      return null
    }
    endAt.setTime(nextTime.getTime())
  }
  return {
    id: item.id,
    teamId: item.teamId,
    startAt: formatDateTimeValue(startAt),
    endAt: formatDateTimeValue(endAt),
    sourceType: item.sourceType,
    remark: item.remark,
  }
}

function sliceScheduleForDay(item: AdminAgentTeamSchedule, day: Date) {
  const dayStart = startOfDay(day)
  const dayEnd = addDays(dayStart, 1)
  const scheduleStart = parseLocalDateTime(item.startAt)
  const scheduleEnd = parseLocalDateTime(item.endAt)
  const visibleStart = new Date(Math.max(scheduleStart.getTime(), dayStart.getTime()))
  const visibleEnd = new Date(Math.min(scheduleEnd.getTime(), dayEnd.getTime()))
  if (!visibleEnd.getTime() || visibleEnd <= visibleStart) {
    return null
  }
  const left = ((visibleStart.getTime() - dayStart.getTime()) / dayMs) * 100
  const width = ((visibleEnd.getTime() - visibleStart.getTime()) / dayMs) * 100
  return { left, width, visibleStart, visibleEnd }
}

export function ScheduleCalendar({
  weekStart,
  teams,
  schedules,
  loading,
  savingId,
  onCreate,
  onEdit,
  onMove,
  onResize,
}: ScheduleCalendarProps) {
  const days = Array.from({ length: 7 }, (_, index) => startOfDay(addDays(weekStart, index)))

  function handleBlankCellClick(teamId: number, day: Date) {
    const startAt = new Date(day)
    startAt.setHours(9, 0, 0, 0)
    const endAt = new Date(day)
    endAt.setHours(18, 0, 0, 0)
    onCreate({
      teamId,
      startAt: formatDateTimeValue(startAt),
      endAt: formatDateTimeValue(endAt),
      sourceType: "manual",
      remark: "",
    })
  }

  function handlePointerDown(event: React.PointerEvent, item: AdminAgentTeamSchedule, type: DragState["type"], edge?: "start" | "end") {
    event.preventDefault()
    event.stopPropagation()
    const target = event.currentTarget as HTMLElement
    target.setPointerCapture(event.pointerId)
    const state: DragState =
      type === "resize"
        ? { type: "resize", edge: edge ?? "end", item, moved: false }
        : { type: "move", item, startX: event.clientX, startY: event.clientY, moved: false }

    function handlePointerMove(moveEvent: PointerEvent) {
      if (state.type === "move") {
        if (Math.abs(moveEvent.clientX - state.startX) > 4 || Math.abs(moveEvent.clientY - state.startY) > 4) {
          state.moved = true
        }
      } else {
        state.moved = true
      }
    }

    async function handlePointerUp(upEvent: PointerEvent) {
      target.releasePointerCapture(event.pointerId)
      window.removeEventListener("pointermove", handlePointerMove)
      window.removeEventListener("pointerup", handlePointerUp)
      if (!state.moved) {
        onEdit(item)
        return
      }
      const cell = getDropCell(upEvent)
      if (!cell) {
        return
      }
      if (state.type === "move") {
        const next = getCellTeamAndDate(cell)
        if (!next.teamID || !next.date) {
          return
        }
        await onMove(buildMovePayload(item, next.teamID, next.date))
        return
      }
      const payload = buildResizePayload(item, state.edge, getPointerDateInCell(upEvent, cell))
      if (payload) {
        await onResize(payload)
      }
    }

    window.addEventListener("pointermove", handlePointerMove)
    window.addEventListener("pointerup", handlePointerUp)
  }

  if (teams.length === 0 && !loading) {
    return (
      <div className="flex min-h-64 items-center justify-center rounded-lg border bg-background text-sm text-muted-foreground">
        暂无客服组，无法展示排班日历
      </div>
    )
  }

  return (
    <div className="overflow-x-auto rounded-lg border bg-background">
      <div className="min-w-[980px]">
        <div className="grid grid-cols-[168px_repeat(7,minmax(112px,1fr))] border-b bg-muted/40">
          <div className="flex h-14 items-center px-4 text-sm font-medium text-muted-foreground">客服组</div>
          {days.map((day, index) => (
            <div key={day.toISOString()} className="flex h-14 flex-col justify-center border-l px-3">
              <div className="text-sm font-medium">周{dayNames[index]}</div>
              <div className="text-xs text-muted-foreground">{formatDayTitle(day)}</div>
            </div>
          ))}
        </div>

        <div className={cn("relative", loading && "opacity-60")}>
          {teams.map((team) => (
            <div key={team.id} className="grid min-h-28 grid-cols-[168px_repeat(7,minmax(112px,1fr))] border-b last:border-b-0">
              <div className="flex min-h-28 items-center px-4">
                <div className="min-w-0">
                  <div className="truncate text-sm font-medium">{team.name}</div>
                  <div className="text-xs text-muted-foreground">组ID：{team.id}</div>
                </div>
              </div>
              {days.map((day) => {
                const date = formatDate(day)
                const daySchedules = schedules.filter((item) => item.teamId === team.id && sliceScheduleForDay(item, day))
                return (
                  <button
                    key={`${team.id}-${date}`}
                    type="button"
                    data-schedule-cell
                    data-team-id={team.id}
                    data-date={date}
                    className="relative min-h-28 border-l bg-background p-2 text-left transition-colors hover:bg-muted/20"
                    onClick={(event) => {
                      if (event.target === event.currentTarget) {
                        handleBlankCellClick(team.id, day)
                      }
                    }}
                  >
                    {daySchedules.length === 0 ? (
                      <div className="flex h-full min-h-20 items-center justify-center text-xs text-muted-foreground/70">
                        <CalendarPlusIcon className="mr-1 size-3.5" />
                        新增
                      </div>
                    ) : null}
                    {daySchedules.map((item, index) => {
                      const slice = sliceScheduleForDay(item, day)
                      if (!slice) {
                        return null
                      }
                      const busy = savingId === item.id
                      return (
                        <div
                          key={`${item.id}-${date}`}
                          role="button"
                          tabIndex={0}
                          className={cn(
                            "absolute top-2 z-10 h-20 cursor-grab overflow-hidden rounded-md border border-primary/20 bg-primary/10 px-2 py-1.5 text-primary shadow-sm outline-none transition active:cursor-grabbing",
                            busy && "pointer-events-none opacity-60"
                          )}
                          style={{
                            left: `calc(${slice.left}% + 8px)`,
                            width: `calc(${slice.width}% - 16px)`,
                            top: `${8 + index * 28}px`,
                            minWidth: "42px",
                          }}
                          onPointerDown={(event) => handlePointerDown(event, item, "move")}
                          onKeyDown={(event) => {
                            if (event.key === "Enter" || event.key === " ") {
                              event.preventDefault()
                              onEdit(item)
                            }
                          }}
                        >
                          <div
                            className="absolute left-0 top-0 flex h-full w-2 cursor-ew-resize items-center justify-center bg-primary/15"
                            onPointerDown={(event) => handlePointerDown(event, item, "resize", "start")}
                          >
                            <GripVerticalIcon className="size-3" />
                          </div>
                          <div
                            className="absolute right-0 top-0 flex h-full w-2 cursor-ew-resize items-center justify-center bg-primary/15"
                            onPointerDown={(event) => handlePointerDown(event, item, "resize", "end")}
                          >
                            <GripVerticalIcon className="size-3" />
                          </div>
                          <div className="truncate pl-2 pr-2 text-xs font-medium">{item.sourceType}</div>
                          <div className="truncate pl-2 pr-2 text-xs">
                            {formatDateTime(item.startAt).slice(11, 16)} - {formatDateTime(item.endAt).slice(11, 16)}
                          </div>
                          {item.remark ? (
                            <div className="truncate pl-2 pr-2 text-[11px] text-primary/80">{item.remark}</div>
                          ) : null}
                        </div>
                      )
                    })}
                  </button>
                )
              })}
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}
