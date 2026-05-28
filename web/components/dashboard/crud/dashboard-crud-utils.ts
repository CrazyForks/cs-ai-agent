export type DashboardCrudQueryValue = string | number | undefined

export type DashboardCrudQueryFilter = {
  name: string
  trim?: boolean
  allValue?: string | number
  valueType?: "string" | "number"
}

export type DashboardCrudPageResult<T> = {
  results: T[]
  page: {
    page: number
    limit: number
    total: number
  }
}

export function buildDashboardCrudQuery({
  values,
  filters,
  page,
  limit,
}: {
  values: Record<string, string | number | undefined>
  filters: DashboardCrudQueryFilter[]
  page: number
  limit: number
}): Record<string, DashboardCrudQueryValue> {
  const query: Record<string, DashboardCrudQueryValue> = {}

  filters.forEach((filter) => {
    const rawValue = values[filter.name]
    const value =
      filter.trim && typeof rawValue === "string" ? rawValue.trim() : rawValue

    if (
      value === undefined ||
      value === "" ||
      (filter.allValue !== undefined && String(value) === String(filter.allValue))
    ) {
      return
    }

    if (filter.valueType === "number") {
      const numberValue = Number(value)
      if (Number.isFinite(numberValue)) {
        query[filter.name] = numberValue
      }
      return
    }

    query[filter.name] = value
  })

  query.page = page
  query.limit = limit
  return query
}

export function normalizeDashboardCrudPageResult<T>(
  result: Partial<DashboardCrudPageResult<T>> | null | undefined,
  page: number,
  limit: number
): DashboardCrudPageResult<T> {
  return {
    results: Array.isArray(result?.results) ? result.results : [],
    page: {
      page: result?.page?.page ?? page,
      limit: result?.page?.limit ?? limit,
      total: result?.page?.total ?? 0,
    },
  }
}
