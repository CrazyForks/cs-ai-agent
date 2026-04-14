package engine

import runtimeeino "cs-agent/internal/ai/infra/eino"

// TODO 这个地方为什么要定义类型别名，不能直接用吗？
type RunInput = runtimeeino.RunInput
type ResumeInput = runtimeeino.ResumeInput
type InterruptContextSummary = runtimeeino.InterruptContextSummary
type RunResult = runtimeeino.RunResult

type Request = RunInput
type ResumeRequest = ResumeInput
type Summary = RunResult
