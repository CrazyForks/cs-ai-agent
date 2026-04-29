import assert from "node:assert/strict"
import test from "node:test"

import { addDays, formatWeekTitle, startOfWeek } from "./calendar-date-range.ts"

test("builds Monday-based week range and title", () => {
  const start = startOfWeek(new Date(2026, 3, 29, 14, 0, 0))

  assert.equal(start.getFullYear(), 2026)
  assert.equal(start.getMonth(), 3)
  assert.equal(start.getDate(), 27)
  assert.equal(formatWeekTitle(start), "2026-04-27 - 2026-05-03")
  assert.equal(addDays(start, 7).getDate(), 4)
})
