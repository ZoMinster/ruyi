## Context

当前 OpenSpec 资产已经完成过一次归档聚合，`openspec/changes/archive`下主要是`plugin-framework`、`user-auth`、`distributed-infra`等功能分组目录，而不是大量日期前缀原始归档目录。继续压缩时，问题不再是简单合并碎片，而是归档分组内部和跨分组之间重复承载最终能力契约。

现状基线：

| 指标 | 当前值 |
|---|---:|
| `openspec/changes/archive`体量 | 约`3.5M` |
| `openspec/specs`体量 | 约`1.0M` |
| 归档`spec.md`数量 | `277` |
| 主规范`spec.md`数量 | `98` |
| 跨归档分组重复能力名 | `61` |
| 重复能力涉及归档 spec 文件 | `215` |
| 按能力名计算的额外副本 | `154` |

典型重复包括`cron-job-management`出现在 9 个归档分组，`user-auth`出现在 8 个归档分组，`role-management`、`plugin-manifest-lifecycle`、`online-user`、`config-management`分别出现在 7 个归档分组。若继续让每个归档分组长期保存完整最终规范，AI 在读取历史时会反复消耗上下文，并可能误判不同位置的同名能力规范都是同等事实来源。

## Goals / Non-Goals

**Goals:**

- 将`openspec/specs`确认为当前能力契约唯一事实来源。
- 将`openspec/changes/archive`定位为历史摘要、设计演进、反馈闭环和验证证据的承载位置。
- 为归档`specs/`建立能力 owner 映射，消除跨分组重复全文规范。
- 通过样板分组先行验证压缩规则，再批量处理高体量归档分组。
- 建立量化验收指标，保证压缩是可审查、可验证、可回退的治理动作。

**Non-Goals:**

- 不修改业务代码、前端页面、接口、数据库、缓存、插件源码或运行时行为。
- 不在本提案阶段直接压缩或删除既有归档目录。
- 不改变 OpenSpec CLI 的归档机制。
- 不移除 Git 历史中的完整归档原文追溯能力。
- 不把所有历史细节迁移到新的大型文档中，避免用另一种形式复制冗余。

## Decisions

### 1. 当前契约只由`openspec/specs`承载

`openspec/specs`作为当前能力契约的唯一事实来源。归档分组中的`specs/`只保留该分组作为能力 owner 时仍有维护价值的历史规范摘要，或者保留归档时特有的演进约束。若归档`spec.md`与主规范完全相同，默认删除归档副本，不再保存重复全文。

替代方案是保留所有归档 spec，通过读取顺序约定让 AI 优先读取主规范。该方案不改变体量，也无法消除同名规范的歧义。

### 2. 归档分组必须建立能力 owner 映射

每个能力最多由一个归档分组作为历史 owner。owner 分组负责承载该能力的历史演进、关键决策和保留约束；非 owner 分组不得长期保存该能力完整`spec.md`，只能在`design.md`中保留交叉影响摘要。

初始 owner 映射按能力主要交付物确定，例如：

| 能力 | 归档 owner |
|---|---|
| `plugin-runtime-loading` | `plugin-framework` |
| `plugin-manifest-lifecycle` | `plugin-framework` |
| `plugin-host-service-extension` | `plugin-framework` |
| `plugin-cache-service` | `plugin-framework` |
| `cron-job-management` | `scheduled-jobs` |
| `config-management` | `system-config` |
| `role-management` | `user-management` |
| `menu-management` | `user-management` |
| `dict-management` | `org-structure` |
| `database-bootstrap-commands` | `database-engine` |
| `cluster-deployment-mode` | `distributed-infra` |
| `distributed-cache-coordination` | `distributed-infra` |
| `e2e-suite-organization` | `e2e-testing` |

具体执行时允许根据完整语义阅读结果调整 owner，但必须在任务记录中说明理由。

### 3. 非 owner 分组使用交叉影响摘要替代完整规范

当某个分组确实对非 owner 能力产生过影响时，不直接丢弃语义，而是在该分组`design.md`的交叉影响章节中记录最短维护摘要。摘要包含影响主题、最终契约位置和历史 owner，不保留完整 requirement 与 scenario。

示例：

```markdown
## Cross-Domain Impacts

- `user-auth`影响`plugin-runtime-loading`的登录态可见性和权限过滤边界；当前契约由`openspec/specs/plugin-runtime-loading/spec.md`承载，历史 owner 为`archive/plugin-framework`。
```

### 4. 压缩按样板分组推进

先选择一个高体量且重复明显的归档分组作为样板，建议优先使用`plugin-framework`或`user-auth`。样板压缩必须完整覆盖`proposal.md`、`design.md`、`tasks.md`和`specs/`，验证指标达标后再批量处理其他分组。

推荐顺序：

1. `plugin-framework`
2. `user-auth`
3. `distributed-infra`
4. `devops-tooling`
5. `code-quality`

该顺序优先处理体量大、重复能力多、跨域影响复杂的分组，尽早暴露规则缺陷。

### 5. 文件级压缩边界固定

| 文件 | 压缩后职责 |
|---|---|
| `proposal.md` | 背景、目标、范围、影响，不保留实施流水 |
| `design.md` | 架构决策、方案演进、废弃原因、关键约束、交叉影响摘要 |
| `tasks.md` | `FB-*`、根因、修复、验证、审查、治理影响的最小维护摘要 |
| `specs/` | 仅保留 owner 能力历史契约或无法安全迁移的能力摘要 |

普通 checklist、重复命令、逐文件记录、与主规范完全相同的最终契约副本必须优先裁剪。无法确认是否安全裁剪时，保留最短摘要并记录阻断原因。

### 6. 验证以静态指标和语义审查结合

压缩不是纯文本变短，必须同时满足确定性指标和语义覆盖：

- `openspec validate compact-openspec-archive-docs --strict`通过。
- 执行样板或批量压缩后，`openspec validate --all`通过。
- 归档重复能力扫描显示跨分组重复能力持续下降。
- `tasks.md`不再包含普通执行流水主导的大段 checklist。
- 每个删除的非 owner spec 都有 owner 映射或交叉影响摘要覆盖。
- 任务记录明确`i18n`、缓存一致性、数据权限、DI、开发工具跨平台和测试策略影响判断。

## Risks / Trade-offs

- 语义摘要遗漏历史细节 → 通过逐目录阅读、owner 映射、交叉影响摘要和语义覆盖记录降低风险；无法确认时不删除原文。
- 删除归档 spec 副本降低离线可读性 → 以`openspec/specs`作为当前契约入口，归档只承载历史原因，减少重复歧义。
- 样板分组规则不适配其他分组 → 先处理一个分组并复盘，再批量推进。
- 压缩后追溯完整执行流水不如原文直接 → 完整原文仍可通过 Git 历史追溯，归档目录只保留维护所需摘要。

## Migration Plan

1. 创建本 OpenSpec 变更，明确归档文档治理规则和任务计划。
2. 建立当前归档体量、spec 数量和重复能力基线。
3. 选择样板分组，完整读取其`proposal.md`、`design.md`、`tasks.md`和`specs/`。
4. 为样板分组建立能力 owner 判定，压缩非 owner spec 为交叉影响摘要。
5. 压缩样板分组的`proposal.md`、`design.md`和`tasks.md`。
6. 运行 OpenSpec 校验、静态重复扫描和 Markdown 格式检查。
7. 根据样板结果批量压缩其他高体量分组。
8. 复核总体指标，记录压缩前后体量、重复能力数量、保留信息类别和未压缩原因。

## Open Questions

- 样板分组首选`plugin-framework`还是`user-auth`，可在实施前根据当前活跃变更归档状态最终确认。
- 是否需要后续为能力 owner 映射建立持久清单文件，还是仅在归档压缩任务记录中维护。本变更先不引入新的长期配置文件。
