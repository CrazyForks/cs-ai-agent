package factory

// DefaultProjectInstruction 为当前项目统一注入的全局项目规则。
//
// 默认回退到内置常量；运行时优先由 ProjectInstructionProvider 读取仓库中的 AGENTS.md。
// TODO 这个内容不合适，需要重新整理内容，内容需要面向客服系统
const DefaultProjectInstruction = `# AGENTS.md

本文件定义本项目内 AI Agent 的强制开发规则。除非用户明确要求偏离，否则必须遵循。

## 1. 基本原则

- 适用范围：仓库根目录及所有子目录
- 优先级：用户明确指令 > 本文件 > 默认实现习惯
- 若与用户要求冲突：先执行用户要求，并在变更说明中标注偏离点

## 2. 固定技术栈

- 后端：Golang + Iris + GORM + github.com/mlogclub/simple
- 数据库：同时兼容 SQLite 和 MySQL
- 前端：Next.js(App Router) + React + shadcn/ui + Tailwind CSS
- 前端包管理器：pnpm

## 3. 后端分层

必须遵循单向依赖：models -> repositories -> services -> controllers

- models：只定义实体和表映射
- repositories：只封装数据访问
- services：负责业务规则、事务编排、聚合逻辑
- controllers：只做参数解析、权限校验、service 调用、响应封装

禁止：

- controller 直接调用 repository
- 直接将 GORM model 返回前端
- 在 models/repositories 中写业务编排

## 4. simple 使用约定

- DB 初始化后必须执行：sqls.SetDB(db)
- 查询条件优先使用：sqls.Cnd
- 参数绑定优先使用：web/params
- HTTP 响应统一使用：web.JsonData、web.JsonPageData、web.JsonError

## 5. 数据库兼容规则

- 字段类型使用兼容集合：varchar、text、int、bigint、datetime
- 主键统一使用 int64
- 避免数据库私有语法和方言特性

## 6. 接口与 DTO

- DTO 分离：request / response 分开定义
- JSON 字段统一使用 camelCase
- 禁止透传底层 SQL 错误
- controller 入参使用 request DTO
- controller 出参使用 response DTO
- 禁止直接返回 models 到前端

## 7. Go 代码规范

- 日志统一使用标准库 log/slog
- 新增 Go 代码统一使用 any，禁止新增 interface{}
- 修改 Go 代码后必须执行 gofmt

## 8. 前端规范

- 前端目录：web
- 框架：Next.js 16 + App Router
- 基础组件优先使用 shadcn/ui
- 前端业务接口统一通过 web/lib/api/* 发起，禁止页面里散落裸 fetch
- 新增或修改前端页面后至少执行：cd web && pnpm typecheck

## 9. 提交前检查清单

每次修改后至少确认：

1. 没有跨层调用或反向依赖
2. 写操作有明确事务边界
3. 返回仍符合统一 JsonResult 结构
4. 兼容 SQLite 与 MySQL
5. 补充了必要测试，至少覆盖 service 核心路径
6. Go 改动已执行 gofmt
7. 前端改动至少通过 pnpm lint 或 pnpm typecheck（在 web 目录）`
