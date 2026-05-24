"use client"

import { useEffect, useState } from "react"
import { DropletsIcon, PaletteIcon } from "lucide-react"

import { Button } from "@/components/ui/button"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuRadioGroup,
  DropdownMenuRadioItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"

type PaletteMode = "blue" | "green" | "gray"

const PALETTE_STORAGE_KEY = "dashboard_palette"
const DEFAULT_PALETTE: PaletteMode = "green"

const paletteOptions: Array<{
  value: PaletteMode
  label: string
  swatch: string
}> = [
  {
    value: "green",
    label: "温润服务绿",
    swatch: "bg-teal-700",
  },
  {
    value: "gray",
    label: "中性精密灰",
    swatch: "bg-slate-500",
  },
  {
    value: "blue",
    label: "清透科技蓝",
    swatch: "bg-blue-600",
  },
]

function readPalette(): PaletteMode {
  if (typeof window === "undefined") {
    return DEFAULT_PALETTE
  }

  const stored = window.localStorage.getItem(PALETTE_STORAGE_KEY)
  return stored === "blue" || stored === "green" || stored === "gray"
    ? stored
    : DEFAULT_PALETTE
}

function applyPalette(value: PaletteMode) {
  document.documentElement.dataset.palette = value
  window.localStorage.setItem(PALETTE_STORAGE_KEY, value)
}

export function PaletteToggle() {
  const [palette, setPalette] = useState<PaletteMode>(DEFAULT_PALETTE)

  useEffect(() => {
    const storedPalette = readPalette()
    setPalette(storedPalette)
    applyPalette(storedPalette)
  }, [])

  function handleChange(value: string) {
    const nextPalette: PaletteMode =
      value === "blue" || value === "gray" ? value : "green"
    setPalette(nextPalette)
    applyPalette(nextPalette)
  }

  const ActiveIcon = palette === "green" ? DropletsIcon : PaletteIcon

  return (
    <DropdownMenu>
      <DropdownMenuTrigger
        render={<Button variant="outline" size="sm" />}
        aria-label="切换主题色"
      >
        <ActiveIcon />
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" className="w-52 min-w-52">
        <DropdownMenuRadioGroup value={palette} onValueChange={handleChange}>
          {paletteOptions.map((option) => (
            <DropdownMenuRadioItem key={option.value} value={option.value}>
              <span className={`size-2.5 rounded-full ${option.swatch}`} />
              <span className="flex-1">{option.label}</span>
            </DropdownMenuRadioItem>
          ))}
        </DropdownMenuRadioGroup>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
