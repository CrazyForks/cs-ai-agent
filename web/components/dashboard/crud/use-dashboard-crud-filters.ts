"use client"

import { useEffect, useMemo, useState } from "react"

import {
  buildDashboardCrudInitialFilters,
  type DashboardCrudFilterStateConfig,
} from "./dashboard-crud-utils"

export function useDashboardCrudFilters(
  filters: ReadonlyArray<DashboardCrudFilterStateConfig>
) {
  const initialFilters = useMemo(
    () => buildDashboardCrudInitialFilters(filters),
    [filters]
  )
  const [draftFilters, setDraftFilters] = useState(initialFilters)
  const [appliedFilters, setAppliedFilters] = useState(initialFilters)

  useEffect(() => {
    setDraftFilters(initialFilters)
    setAppliedFilters(initialFilters)
  }, [initialFilters])

  function setDraftFilter(name: string, value: string | number | undefined) {
    setDraftFilters((current) => ({
      ...current,
      [name]: value,
    }))
  }

  function applyFilters() {
    setAppliedFilters(draftFilters)
  }

  return {
    draftFilters,
    appliedFilters,
    setDraftFilter,
    setDraftFilters,
    applyFilters,
  }
}
