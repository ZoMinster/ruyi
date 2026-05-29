## 1. 规则与影响确认

- [x] 1.1 实施前重新读取`AGENTS.md`、`.agents/rules/openspec.md`、`.agents/rules/documentation.md`、`.agents/rules/testing.md`和`.agents/rules/i18n.md`，确认本变更仅影响 OpenSpec 文档治理和归档文档内容。
- [x] 1.2 在任务记录中明确无运行时代码、HTTP API、数据库、缓存、数据权限、前端 UI、插件目录、运行时文案、运行期依赖、开发工具跨平台入口或生产构建影响。
- [x] 1.3 记录`i18n`影响判断：本变更不修改运行时语言包、`manifest/i18n`、`apidoc i18n JSON`、菜单、路由或用户可见 UI 文案；影响仅限中文 OpenSpec 治理文档。
- [x] 1.4 记录测试策略：本变更为治理和文档压缩，不新增单元测试或 E2E；验证采用 OpenSpec 严格校验、静态扫描、Markdown 格式检查和语义覆盖审查。

实施记录：

- 已读取规则文件：`AGENTS.md`、`.agents/rules/openspec.md`、`.agents/rules/documentation.md`、`.agents/rules/testing.md`和`.agents/rules/i18n.md`。
- 变更范围：仅修改`openspec/changes/compact-openspec-archive-docs/`任务记录和`openspec/changes/archive/`既有归档文档；不修改业务源码、运行时配置或构建入口。
- 运行时影响：无 Go、HTTP API、数据库、缓存、数据权限、前端 UI、插件目录、运行时文案、运行期依赖、开发工具跨平台入口或生产构建影响。
- `i18n`影响：不修改运行时语言包、`manifest/i18n`、`apidoc i18n JSON`、菜单、路由、按钮、错误消息、插件清单或用户可见 UI 文案；影响仅限中文 OpenSpec 治理文档和历史归档文档压缩。
- 测试策略：本变更为 OpenSpec 文档治理和历史归档压缩，不新增单元测试或 E2E；验证采用`openspec validate`、静态重复扫描、Markdown 空白格式检查和语义覆盖审查。

## 2. 当前归档基线盘点

- [x] 2.1 统计`openspec/changes/archive`和`openspec/specs`体量，记录压缩前基线。
- [x] 2.2 统计归档`spec.md`数量、主规范`spec.md`数量、跨分组重复能力数量和重复能力涉及文件数量。
- [x] 2.3 识别与主规范完全相同的归档 spec 副本，作为优先裁剪候选。
- [x] 2.4 输出高体量归档分组排序，确认样板分组候选。

基线记录：

- 体量：`openspec/changes/archive`为`3.5M`，`openspec/specs`为`1.0M`。
- 文件数：归档`spec.md`为`277`个，主规范`spec.md`为`98`个。
- 重复度：跨分组重复能力`61`个，重复能力涉及归档 spec 文件`215`个，按能力名计算额外副本`154`个。
- 与主规范完全相同的归档 spec 副本`16`个：`user-auth/login-page-presentation`、`user-auth/oper-log`、`user-auth/post-management`、`user-auth/login-log`、`user-auth/module-decoupling`、`user-auth/notice-management`、`user-auth/dashboard-workbench`、`user-auth/framework-i18n-runtime-performance`、`user-auth/user-message`、`user-auth/user-management`、`user-auth/dept-management`、`user-auth/user-role-association`、`devops-tooling/agent-skills-link-cli`、`database-engine/database-dialect-abstraction`、`i18n/framework-i18n-foundation`、`system-config/startup-sql-efficiency`。
- 高体量归档分组排序：`user-auth` `580K`、`plugin-framework` `564K`、`distributed-infra` `336K`、`devops-tooling` `252K`、`code-quality` `236K`、`user-management` `208K`、`host-plugin-boundary` `184K`、`foundation` `180K`。
- 样板分组候选：`plugin-framework`和`user-auth`均符合高体量高重复度条件；本轮优先选择`plugin-framework`，因为其是插件运行时、manifest、host service、cache、UI integration 等多项重复能力的主要交付 owner。

## 3. 能力 owner 映射

- [x] 3.1 为重复能力建立初始 owner 映射，优先按能力主要交付物和长期维护归属判定。
- [x] 3.2 对样板分组中的每个`specs/<capability>/spec.md`判定 owner、非 owner 或保留待确认。
- [x] 3.3 对非 owner 能力准备交叉影响摘要，摘要必须指向当前契约位置和历史 owner。
- [x] 3.4 对无法安全判定 owner 或语义覆盖的能力保留最短摘要，并记录阻断原因。

样板 owner 记录：

- `plugin-framework`保留 owner 能力`29`个：`framework-capability-registry`、`official-plugin-workspace-decoupling`、`plugin-api-query-performance`、`plugin-cache-service`、`plugin-capability-boundary-governance`、`plugin-config-service`、`plugin-data-service`、`plugin-dependency-management`、`plugin-embed-snapshot-packaging`、`plugin-hook-slot-extension`、`plugin-host-service-extension`、`plugin-id-governance`、`plugin-install-enable-shortcut`、`plugin-lock-service`、`plugin-manifest-lifecycle`、`plugin-mock-data-installation`、`plugin-network-service`、`plugin-notify-service`、`plugin-package-boundary-governance`、`plugin-permission-governance`、`plugin-runtime-loading`、`plugin-runtime-upgrade`、`plugin-startup-bootstrap`、`plugin-storage-service`、`plugin-ui-integration`、`plugin-upgrade-governance`、`plugin-workspace-management`、`pluginbridge-subcomponent-architecture`、`workspace-route-boundary`。
- `plugin-framework`非 owner 裁剪能力`20`个：`cluster-deployment-mode`、`cluster-topology-boundaries`、`config-duration-unification`、`core-host-boundary-governance`、`cron-jobs`、`demo-control-guard`、`distributed-locker`、`e2e-suite-organization`、`leader-election`、`menu-management`、`module-decoupling`、`online-user`、`project-setup`、`release-image-build`、`role-management`、`server-monitor`、`service-dependency-injection-governance`、`source-upgrade-governance`、`system-api-docs`、`user-auth`。
- 非 owner 语义覆盖：已在`openspec/changes/archive/plugin-framework/design.md`的`Cross-Domain Impacts`记录当前契约位置和历史 owner，当前契约以`openspec/specs/<capability>/spec.md`为准；其中集群和分布式锁归`distributed-infra`，调度归`scheduled-jobs`，菜单/角色归`user-management`，认证归`user-auth`，发布和工具归`devops-tooling`，E2E 归`e2e-testing`，宿主边界和 DI 归治理主规范。
- 待确认能力：本轮样板未发现必须保留原文但无法判定 owner 的能力；保留的`29`个 owner 规范仍作为插件框架历史摘要入口，后续可在批量阶段继续二次压缩。

## 4. 样板分组压缩

- [x] 4.1 选择`plugin-framework`或`user-auth`作为样板分组，并完整读取该分组`proposal.md`、`design.md`、`tasks.md`和`specs/`全部内容。
- [x] 4.2 压缩样板分组`proposal.md`，仅保留背景、目标、范围和影响。
- [x] 4.3 重写样板分组`design.md`，保留架构决策、方案演进、废弃方案原因、关键约束和交叉影响摘要。
- [x] 4.4 压缩样板分组`tasks.md`，只保留`FB-*`、根因、修复、验证、审查和治理影响的最小维护摘要。
- [x] 4.5 删除或迁移样板分组中非 owner 的重复`spec.md`全文，确保语义已由主规范、owner 分组或交叉影响摘要覆盖。
- [x] 4.6 运行样板分组后的重复能力扫描、OpenSpec 校验和 Markdown 格式检查，并记录结果。

样板压缩记录：

- 已完整读取`openspec/changes/archive/plugin-framework/proposal.md`、`design.md`、`tasks.md`和`specs/`下`49`个规范文件。
- `proposal.md`从`101`行压缩为面向背景、范围、能力和影响的摘要；`design.md`从按迭代/能力铺陈重写为按插件契约、运行时、host service、包边界、UI、启动、升级、工作区、一致性和交叉影响组织；`tasks.md`从`267`行压缩为最小维护证据。
- 删除非 owner 规范全文`20`个，删除后`plugin-framework/specs`保留`29`个 owner 规范。
- 体量变化：`plugin-framework`约从基线`564K`降至`356K`；`openspec/changes/archive`约从`3.5M`降至`3.3M`。
- 样板验证：`openspec validate compact-openspec-archive-docs --strict`通过；`openspec validate --all`通过，`104 passed, 0 failed`；`git diff --check -- openspec/changes/archive/plugin-framework openspec/changes/compact-openspec-archive-docs`无输出。
- 样板重复扫描：跨归档重复能力从`61`降至`57`，重复能力涉及文件从`215`降至`193`，额外副本从`154`降至`136`。

## 5. 批量归档压缩

- [x] 5.1 根据样板结果批量处理`user-auth`、`distributed-infra`、`devops-tooling`、`code-quality`等高体量分组。
- [x] 5.2 对每个分组重复执行完整语义读取、owner 判定、交叉影响摘要、文件职责压缩和非 owner spec 裁剪。
- [x] 5.3 对低体量分组执行剩余重复能力清理，确保跨分组重复能力数量接近 0。
- [x] 5.4 对无法安全压缩的分组或能力保留原文或最短摘要，并在最终报告中列出原因。

批量压缩记录：

- `user-auth`已完整读取`proposal.md`、`design.md`、`tasks.md`和`specs/`下全部规范；保留`14`个认证、会话、租户、登录态和安全治理 owner 规范，删除`34`个非 owner 规范全文；压缩后体量约`128K`，`specs`约`100K`。
- `distributed-infra`已完整读取`proposal.md`、`design.md`、`tasks.md`和`20`个规范；保留`7`个集群、协调 provider、分布式缓存、分布式锁和 leader election owner 规范，删除`13`个非 owner 规范全文；压缩后体量约`80K`，`specs`约`56K`。
- `devops-tooling`已完整读取`proposal.md`、`design.md`、`tasks.md`和`17`个规范；保留`linactl`、Agent 资源、release、upgrade、installer、monthly archive 和 perf audit owner 规范，删除非 owner 或与主规范完全相同的规范全文；压缩后体量约`156K`。
- `code-quality`已完整读取`proposal.md`、`design.md`、`tasks.md`和`22`个规范；保留`7`个 API 合同、后端一致性、i18n 前端刷新性能、Go 单测效率、宿主运行能力、显式 DI 和启动 SQL 效率 owner 规范，删除`15`个非 owner 规范全文；压缩后体量约`92K`，`specs`约`68K`。
- 批量校验：每个分组压缩后均运行`openspec validate compact-openspec-archive-docs --strict`通过；`git diff --check -- openspec/changes/archive/user-auth openspec/changes/archive/distributed-infra openspec/changes/archive/devops-tooling openspec/changes/archive/code-quality openspec/changes/compact-openspec-archive-docs`无输出。
- 当前归档统计：`openspec/changes/archive`约`2.3M`，归档`spec.md`为`191`个；剩余重复能力主要集中在`user-auth`、`user-management`、`plugin-manifest-lifecycle`、`cron-job-management`、`config-management`等低体量分组，需要继续执行 5.3。
- `user-management`已完整读取并压缩用户、角色、菜单、用户角色、登录页和租户平台访问 owner 规范；删除`7`个非 owner 规范全文；压缩后体量约`76K`。
- `system-governance`已完整读取并压缩操作日志、登录日志、在线用户、服务监控、系统 API 文档、系统信息、组件演示、字典导入和宿主数据权限治理 owner 规范；删除`6`个非 owner 规范全文；压缩后体量约`84K`。
- `system-config`已完整读取并压缩`config-management`和`login-home-sql-efficiency` owner 规范；删除`10`个非 owner 规范全文；压缩后体量约`28K`，`specs`约`12K`。
- `host-plugin-boundary`已完整读取并压缩宿主边界、模块解耦、demo-control、HTTP seam 和 pluginbridge 边界样板；删除`15`个非 owner 规范全文，`pluginbridge-subcomponent-architecture`最终归`plugin-framework`；压缩后体量约`52K`。
- `i18n`已完整读取并压缩 i18n 基础设施、工作台 i18n、消息治理、插件治理、项目定位和 README 本地化 owner 规范；删除`12`个非 owner 或与主规范完全相同的规范全文；压缩后体量约`80K`。
- `foundation`已完整读取并压缩`project-setup`和`base-layout` owner 规范；删除`8`个非 owner 规范全文；压缩后体量约`20K`，`specs`约`8K`。
- `database-engine`已完整读取并压缩 PostgreSQL-only、数据库方言、引导命令、SQL 源语法和易失性表 owner 规范；删除`12`个非 owner 或与主规范完全相同的规范全文；压缩后体量约`60K`。
- 低体量重复清理：`org-structure/user-management`、`e2e-testing/project-setup`、`notification/base-layout`、`host-plugin-boundary/pluginbridge-subcomponent-architecture`、`devops-tooling/plugin-upgrade-governance`均已迁移为交叉影响摘要并删除重复规范全文。
- 完全重复主规范副本清理：删除`database-engine/database-dialect-abstraction`、`devops-tooling/agent-skills-link-cli`、`i18n/framework-i18n-foundation`，对应历史原因已由各分组`proposal.md`或`design.md`摘要承载。
- 低体量阶段校验：重复能力扫描为`0`；与主规范完全相同的归档 spec 副本为`0`；未跟踪的`.DS_Store`归档噪音文件已清理；`openspec validate compact-openspec-archive-docs --strict`通过；`git diff --check -- openspec/changes/archive openspec/changes/compact-openspec-archive-docs`无输出。
- 无法安全压缩项：未发现必须保留重复全文且无法迁移的能力；保留的归档`spec.md`均为分组 owner 能力或仍有历史维护价值的能力摘要。非 owner 能力均已在对应分组`design.md`记录交叉影响摘要，并指向`openspec/specs/<capability>/spec.md`与历史 owner 分组。

## 6. 验证与审查

- [x] 6.1 运行`openspec validate compact-openspec-archive-docs --strict`并记录结果。
- [x] 6.2 运行`openspec validate --all`并记录结果。
- [x] 6.3 复跑归档体量、归档 spec 数量、跨分组重复能力和完全重复主规范副本统计，输出压缩前后对比。
- [x] 6.4 运行 Markdown 空白格式检查，确认文档无格式噪音。
- [x] 6.5 执行`lina-review`审查本变更，重点检查 OpenSpec 规范、文档治理、语义覆盖、`i18n`影响判断、测试策略和归档压缩风险。

验证记录：

- `openspec validate compact-openspec-archive-docs --strict`通过。
- `openspec validate --all`通过，结果为`104 passed, 0 failed`。
- 压缩后统计：`openspec/changes/archive`约`1.5M`，归档`spec.md`为`116`个，主规范`spec.md`为`98`个，跨分组重复能力为`0`，与主规范完全相同的归档 spec 副本为`0`。
- 压缩前后对比：归档体量约从`3.5M`降至`1.5M`，归档`spec.md`从`277`降至`116`，跨分组重复能力从`61`降至`0`，重复能力涉及归档 spec 文件从`215`降至`0`，完全重复主规范副本从`16`降至`0`。
- Markdown 空白格式检查：`git diff --check -- openspec/changes/archive openspec/changes/compact-openspec-archive-docs`无输出。
- 噪音文件检查：归档目录内`.DS_Store`扫描无输出。
- `lina-review`审查：已读取`AGENTS.md`、`.agents/rules/openspec.md`、`.agents/rules/documentation.md`、`.agents/rules/testing.md`和`.agents/rules/i18n.md`；范围来自`git status --short`、未跟踪文件展开和`compact-openspec-archive-docs`上下文，覆盖归档压缩变更和活跃变更5个文件；未发现阻塞问题。规则域结论：OpenSpec流程通过，文档治理通过，测试策略使用治理验证且通过，`i18n`无运行时资源影响，缓存一致性/数据权限/DI/开发工具跨平台/运行时代码/API/SQL/前端/插件源码均无影响。
