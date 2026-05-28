## Context

`temp/plugin-design-review-20260527.md` 的 P0 审查点集中在动态插件运行时安全：WASM bridge 目前依赖调用方自行提供 deadline，模块实例化没有宿主侧内存上限；生命周期 SQL 文件执行和 `sys_plugin_migration` 账本记录不在同一事务边界内；动态插件安装、升级或刷新失败后的 rollback 诊断可能只进入 warning 日志；协调器只用进程内互斥，无法覆盖集群节点间并发；`reconciling` 这类瞬态状态在崩溃后缺少过期恢复；协调 tick 中的 panic 可能终止后台 goroutine。

动态插件属于`apps/lina-core`的核心宿主扩展能力，本变更收紧已有运行时、生命周期和集群协调语义，不新增管理工作台页面、HTTP API、插件 manifest 字段或独立产品能力。当前仓库仍有多个已完成但未归档的活跃变更；本变更作为新的后续 OpenSpec 变更，只定义后续实施边界，不修改既有已完成变更状态。

## Goals / Non-Goals

**Goals:**

- 为所有动态插件 WASM 执行入口提供统一宿主兜底资源边界，覆盖超时和内存上限，并尊重调用方已有 deadline。
- 让插件生命周期 SQL 文件执行和迁移账本记录在 PostgreSQL 默认环境下同成同败。
- 将 rollback SQL、菜单、前端资源、权限治理等恢复失败纳入结构化诊断和失败状态，避免只写日志后继续隐藏不一致。
- 在集群模式下用 per-plugin 分布式锁保护动态插件协调器共享生命周期副作用。
- 增加 stale `reconciling` 检测恢复和 tick 级 panic recovery，使后台协调器可继续收敛。
- 明确测试、缓存一致性、数据权限、`i18n`、开发工具和数据库影响边界。

**Non-Goals:**

- 不实现 P1/P2/P3 审查点，例如主机调用配额、存储读写大小限制、Facade 层生命周期串行化、能力接口版本化或网络响应头模型。
- 不新增动态插件管理页、HTTP API、OpenAPI DTO 或用户可见运行时文案。
- 不引入新的协调系统、外部分布式锁依赖或数据库 advisory lock 方言分支。
- 不扩展插件 manifest schema；若后续需要可配置资源阈值，优先放在宿主配置中。
- 不支持 MySQL DDL 补偿方案。当前项目默认 SQL 源和运行环境按 PostgreSQL 14+ 约束治理，本变更只要求 PostgreSQL 事务一致性。

## Decisions

1. WASM bridge 在宿主侧统一加资源 guard。

   `ExecuteBridge` 应成为默认资源边界入口。调用方已经传入 deadline 时继续沿用更严格的调用方上下文；调用方没有 deadline 时，bridge 创建默认超时上下文。模块实例化使用明确内存页上限，避免动态插件通过 guest 分配消耗宿主无边界内存。HTTP 路由、cron discovery、cron job 和生命周期回调都继续走同一 bridge 入口，避免各调用方各自实现超时逻辑。

   备选方案是在 route、cron、lifecycle 各入口分别加超时。该方案容易遗漏新入口，并使默认值分散，拒绝采用。

2. 配置采用宿主配置治理，默认值可直接在 bridge 配置层生效。

   如果实现阶段需要暴露 `timeout`、`memoryLimitPages` 等配置，应使用宿主配置文件和配置读取层，时间长度使用带单位字符串并解析为 `time.Duration`。纯值配置可以使用专门配置结构体，不得为此新增运行期依赖聚合结构或新的 DI 容器。

   备选方案是在插件 manifest 中声明资源上限。该方案会扩大插件发布契约和治理面，且本次需求是宿主兜底保护，拒绝采用。

3. 生命周期 SQL 使用 PostgreSQL 事务包裹 SQL 文件集和迁移账本。

   install、upgrade、uninstall、rollback 等生命周期 SQL 执行应在同一事务内完成 SQL 语句执行与 `sys_plugin_migration` 账本写入。任一 SQL 文件转译、语句执行或账本写入失败时，事务回滚，调用方看到失败状态。Mock SQL 已有事务路径时应保持语义一致，不引入第二套执行模型。

   备选方案是继续逐语句 auto-commit，再依赖 rollback SQL 补偿。该方案无法保证中途失败的数据完整性，拒绝采用。数据库 advisory lock 或 MySQL DDL 补偿不属于本变更范围。

4. rollback 失败必须进入失败诊断。

   动态插件安装、升级、同版本刷新或回滚期间，原始失败和 rollback SQL、菜单恢复、前端资源恢复、权限治理恢复失败都必须聚合到返回错误、发布诊断、节点状态或 registry 失败状态中。日志只作为辅助观测，不能替代权威失败状态。

   备选方案是继续只 warning 记录 rollback 失败。该方案会让管理端和后续自动恢复流程无法判断数据是否一致，拒绝采用。

5. 集群模式下按插件 ID 获取 distributed locker。

   动态插件协调器在执行 SQL、菜单/权限同步、active release 切换、frontend bundle 切换、runtime revision 发布等共享副作用前，应获取 per-plugin 锁。`cluster.enabled=true`时复用既有 `locker`/coordination lock 能力；未获得锁的节点跳过当前插件并等待下一次 revision 或 safety sweep。`cluster.enabled=false`时保留进程内互斥或 SQL/local 分支，不强制依赖 Redis。

   备选方案是 PostgreSQL advisory lock。该方案引入方言专属语义，与现有 distributed locker 能力重复，拒绝采用。

6. stale `reconciling` 使用阈值检测和保守恢复。

   协调 tick 开始前或单插件协调前，应识别 `CurrentState=reconciling` 且 `updated_at` 超过阈值的记录。只有超过阈值的瞬态状态才允许恢复；仍在阈值内的记录视为可能有其他节点正在处理，不能盲目重置。恢复目标应基于权威安装标志、启用状态和 active release 推导出的稳定状态，然后继续正常收敛或标记失败诊断。

   备选方案是每次看到 `reconciling` 都重置。该方案会破坏正在执行的合法协调流程，拒绝采用。

7. 协调器按 tick recover panic。

   后台循环每次 tick 应有 panic recovery 边界，记录 panic 值和可用堆栈后继续下一轮。panic 不应被当作成功收敛；相关插件若已进入 transient 状态，应依赖 stale 恢复或错误状态机制后续收敛。

   备选方案是在最外层 goroutine 只 recover 一次后退出。该方案仍会丢失后续协调能力，拒绝采用。

## Risks / Trade-offs

- WASM 默认超时可能中断合法长任务。缓解：默认值只作为无 deadline 兜底；长任务调用方可提供更明确 deadline，配置项如有新增需通过宿主配置治理。
- 内存上限可能暴露现有示例插件的隐含大内存需求。缓解：用单元测试和动态插件 smoke 覆盖示例插件常规路由、cron 和生命周期执行。
- 生命周期 SQL 事务会让长迁移持有事务时间变长。缓解：插件 SQL 必须保持幂等和可分批设计；实施阶段通过集成测试覆盖失败回滚和账本一致性。
- 分布式锁不可用时可能延迟动态插件收敛。缓解：集群模式使用既有 coordination lock 失败语义，未获锁节点跳过并重试；单机模式不依赖分布式后端。
- stale `reconciling` 阈值过短可能误伤慢迁移，过长会延迟恢复。缓解：阈值使用宿主配置或清晰默认值，测试覆盖阈值内不恢复、阈值外恢复。
- rollback 诊断聚合可能改变错误返回链。缓解：仅增强失败可观测性，不新增 HTTP API 字段；若实现触达用户可见错误码或文案，按`i18n`和 API 文档规则补齐治理。

## Impact Analysis

- 宿主边界：本变更属于`apps/lina-core`动态插件通用运行时能力，和管理工作台展示结构无关。
- 数据权限：不新增业务数据读写接口；动态插件通过 host service 访问业务数据的权限边界不变。实施阶段若改动 host service 数据访问路径，必须重新接入数据权限验证。
- 缓存一致性：继续以现有插件 registry、active release、runtime revision、frontend bundle、runtime i18n 和 Wasm 编译缓存为权威状态源与派生缓存；本变更新增协调互斥和恢复路径，不新增第二套缓存状态源。
- 数据库：原则上不新增表、列或索引；生命周期 SQL 和 `sys_plugin_migration` 账本进入事务边界。若实施发现需要状态诊断字段或索引，必须通过当前迭代 SQL 文件、幂等 SQL 和 DAO 生成流程处理。
- API 契约：默认不新增或修改 HTTP API、DTO、路由和 OpenAPI 元数据。
- 开发工具：默认不修改脚本、CI 或 `linactl`；如果为 WASM resource guard 增加构建或测试工具入口，需记录跨平台影响。
- `i18n`：默认不新增用户可见文案、菜单、API 文档源文本或语言包。若新增业务错误码、fallback 或 API 文档文本，必须同步治理。
- 测试：需要后端单元测试、可控替身测试或集成测试覆盖 WASM 超时和内存上限、生命周期 SQL 事务失败回滚、rollback 失败诊断、集群锁未获得时跳过、stale `reconciling` 恢复、tick panic recovery。纯 OpenSpec 文档阶段以`openspec validate --strict`作为治理验证。
