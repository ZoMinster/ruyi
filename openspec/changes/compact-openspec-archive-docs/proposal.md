## Why

当前`openspec/changes/archive`约`3.5M`，其中归档`specs/`约`277`个文件，是 OpenSpec 文档体量和 AI 上下文压力的主要来源。归档中大量能力规范跨功能分组重复出现，且部分归档规范与`openspec/specs`主规范完全相同，导致维护者和 AI 难以快速判断当前契约、历史原因和可裁剪内容的边界。

## What Changes

- 建立 OpenSpec 文档信息分层：`openspec/specs`作为当前能力契约唯一事实来源，活跃变更只承载本次增量，`openspec/changes/archive`只承载历史摘要、设计演进、反馈闭环和验证证据。
- 为归档`specs/`引入能力 owner 映射规则，同一能力的历史规范只由一个归档分组长期承载，非 owner 分组中的重复能力规范迁移为交叉影响摘要。
- 定义归档二次压缩流程：先建立体量与重复度基线，再以`plugin-framework`或`user-auth`作为样板分组验证压缩规则，确认后批量处理其他高体量分组。
- 明确`proposal.md`、`design.md`、`tasks.md`和`specs/`的压缩边界，裁剪普通执行流水、重复最终契约和低价值过程记录，保留`FB-*`、根因、验证、审查和治理影响。
- 增加压缩验收指标和静态验证要求，包括归档体量、归档 spec 数量、跨分组重复能力数量、OpenSpec 严格校验和人工语义覆盖记录。
- 本变更只提出归档文档治理和压缩执行方案，不修改运行时代码、HTTP API、数据库、前端 UI、插件源码或业务行为。

## Capabilities

### New Capabilities

- `openspec-archive-document-governance`：定义 OpenSpec 归档文档的信息分层、能力 owner 映射、重复规范裁剪、分阶段压缩、语义覆盖和验证报告要求。

### Modified Capabilities

无。

## Impact

- 影响后续对`openspec/changes/archive`下既有归档分组的文档压缩和结构重写。
- 影响 OpenSpec 归档治理方式，要求未来归档聚合避免在多个分组重复保存同一能力的最终规范全文。
- 不影响`apps/lina-core`、`apps/lina-plugins`、HTTP API、数据库、缓存、前端 UI、运行时文案、插件清单、语言包或生产构建。
- 验证方式以`openspec validate compact-openspec-archive-docs --strict`、归档重复度静态扫描、Markdown 格式检查和样板分组语义覆盖审查为主。
