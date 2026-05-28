## 1. 规则和影响确认

- [x] 1.1 重新读取并记录本变更命中的`.agents/rules/openspec.md`、`documentation.md`、`architecture.md`、`plugin.md`、`backend-go.md`、`database.md`、`cache-consistency.md`、`data-permission.md`、`testing.md`和`i18n.md`规则。
- [x] 1.2 在执行记录中确认本变更属于`apps/lina-core`动态插件通用宿主能力，不修改工作台展示契约。
- [x] 1.3 记录影响分析：默认无 HTTP API、前端、目录级 README、插件目录源码、开发工具脚本和用户可见`i18n`文案变更；如实施中发生变化，按对应规则补充任务和验证。
- [x] 1.4 记录数据权限边界：本变更不新增业务数据接口；若改动 host service 数据访问路径，必须追加数据权限测试。

## 2. WASM Bridge 资源边界

- [x] 2.1 梳理动态插件 WASM 执行入口，确认 HTTP 路由、cron discovery、cron job 和生命周期回调均经过统一 bridge 资源 guard。
- [x] 2.2 为 bridge 执行添加无 deadline 时的默认超时，并保证已有调用方 deadline 不被放宽。
- [x] 2.3 为 WASM 模块实例化添加宿主侧内存上限，必要时通过宿主配置读取纯值配置并按`time.Duration`字符串规则解析超时。
- [x] 2.4 补充 WASM 执行超时和内存上限测试，覆盖无限循环、超大分配、已有 deadline 和无 deadline 场景。

## 3. 生命周期 SQL 事务和 Rollback 诊断

- [x] 3.1 将插件 install、upgrade、uninstall 和 rollback 生命周期 SQL 文件执行与`sys_plugin_migration`账本写入纳入同一 PostgreSQL 事务边界。
- [x] 3.2 确认 SQL 转译、语句执行或账本写入任一步失败时，事务回滚且不会留下成功迁移账本。
- [x] 3.3 强化动态插件安装、升级和同版本刷新失败后的 rollback 诊断，聚合原始错误和 rollback SQL、菜单、前端资源、权限治理恢复错误。
- [x] 3.4 补充生命周期 SQL 事务失败和 rollback 失败诊断测试，覆盖 SQL 中途失败、账本写入失败和 rollback 失败。

## 4. Reconciler 集群互斥与恢复

- [x] 4.1 在动态插件协调器共享生命周期副作用前接入 per-plugin 分布式锁，`cluster.enabled=true`时复用 coordination locker，单机模式保留本地互斥分支。
- [x] 4.2 确认未获得 per-plugin 锁的节点跳过当前插件，并等待后续 revision、event 或 safety sweep 重试。
- [x] 4.3 增加 stale `reconciling` 检测与恢复，仅恢复超过阈值的瞬态状态，阈值内状态不得被重置。
- [x] 4.4 为协调器 tick 增加 panic recovery，记录诊断并保证后台循环继续运行。
- [x] 4.5 补充协调器测试，覆盖同插件锁竞争、不同插件独立收敛、stale `reconciling`恢复、阈值内不恢复和 tick panic 后继续运行。

## 5. 缓存、升级和运行时一致性验证

- [x] 5.1 验证动态插件升级或同版本刷新失败时旧 active release 保持有效，失败目标发布不会成为 runtime revision、enabled snapshot、frontend bundle、runtime i18n 或 Wasm 缓存权威来源。
- [x] 5.2 验证 rollback 失败时采用保守暴露策略：不得暴露失败目标发布能力，继续使用旧发布或隐藏插件能力。
- [x] 5.3 记录缓存一致性判断：权威数据源仍为 registry、active release 和 runtime revision，本变更只增加互斥与恢复路径。
- [x] 5.4 运行覆盖变更包的 Go 编译门禁和相关单元测试；涉及启动装配或配置读取时补跑宿主启动绑定包测试。

## 6. 治理验证和审查

- [x] 6.1 运行`openspec validate harden-dynamic-plugin-runtime-safety --strict`并记录结果。
- [x] 6.2 运行必要的静态检索或审查检查，确认没有新增无界`N+1`查询、用户可见文案遗漏、未治理配置或开发工具跨平台影响。
- [x] 6.3 完成实现和验证后调用`lina-review`进行代码、规范和任务状态审查。

## 执行记录

- 2026-05-27：已按 apply 范围重新读取命中规则文件：`openspec.md`、`documentation.md`、`architecture.md`、`plugin.md`、`backend-go.md`、`database.md`、`cache-consistency.md`、`data-permission.md`、`testing.md`、`i18n.md`。
- 2026-05-27：确认本变更属于`apps/lina-core`动态插件通用宿主能力；不修改管理工作台展示契约，不新增 HTTP API、前端页面、目录级 README、插件目录源码或开发工具脚本。
- 2026-05-27：确认默认无用户可见`i18n`文案变更；本变更不新增业务数据接口，动态插件 host service 数据权限边界保持不变。
- 2026-05-27：核对`temp/plugin-design-review-20260527.md`，确认 P0-1 至 P0-7 均为值得立即改进的安全、数据完整性和集群可靠性问题；P1 及以后继续作为后续迭代候选，不纳入本变更。
- 2026-05-27：WASM 执行入口静态检索`rg -n "ExecuteBridge\\(" apps/lina-core/internal/service/plugin/internal -g'*.go'`确认 HTTP route、cron discovery、cron job 和生命周期 precondition 均经`wasm.ExecuteBridge`；`ExecuteBridge`已在无 deadline 时增加 30 秒默认超时，已有 deadline 不放宽，模块实例化使用 wazero 内存上限`4096`页（256 MiB），未新增宿主配置项。
- 2026-05-27：新增 WASM 测试覆盖默认 deadline、调用方 deadline 保持、无限循环 context 取消和超限内存拒绝；验证命令`cd apps/lina-core && go test ./internal/service/plugin/internal/wasm -count=1`通过。
- 2026-05-27：生命周期 SQL 执行改为在`dao.SysPluginMigration.Transaction`内完成 SQL 文件集执行和成功账本写入；SQL 转译、执行或账本写入失败均回滚本次 SQL 与账本。失败账本不在同一已失败事务中写入，失败诊断由返回错误、发布失败态和节点投影承载。
- 2026-05-27：新增生命周期测试覆盖 SQL 中途失败回滚、账本写入失败回滚和迁移账本无残留；验证命令`cd apps/lina-core && go test ./internal/service/plugin/internal/lifecycle -count=1`通过。
- 2026-05-27：rollback 诊断已聚合原始错误、rollback SQL、菜单删除、旧菜单/权限恢复和发布/节点状态恢复错误；安装、升级和刷新存在失败目标 release 时标记 failed，卸载失败不再误写 registry failed，而是恢复原稳定状态。
- 2026-05-27：协调器在主节点共享生命周期副作用前接入 per-plugin 分布式锁，`cluster.enabled=true`时使用注入的`locker.Service`，单机模式保留进程内互斥；后台协调拿不到锁跳过并等待后续 revision/event/safety sweep，显式运行时升级拿不到锁返回错误，避免误判成功。
- 2026-05-27：新增 stale `reconciling`恢复，仅超过 5 分钟阈值才恢复到稳定态；阈值内状态跳过当前插件，避免破坏其他节点正在执行的生命周期副作用。协调 tick 增加 panic recovery 并记录 panic 值和堆栈，保证后台循环继续。
- 2026-05-27：新增协调器安全测试覆盖同插件锁竞争、不同插件独立锁、显式锁冲突错误、stale `reconciling`恢复、fresh `reconciling`跳过、tick panic recovery 和 rollback 诊断聚合；验证命令`cd apps/lina-core && go test ./internal/service/plugin/internal/runtime -count=1`通过。
- 2026-05-27：失败发布保守暴露策略已补强：runtime upgrade projection 将 registry failed 或 target release failed 投影为`upgrade_failed`，动态路由复用既有`409`和`CodePluginRuntimeUpgradeRequired`阻止失败 release 被业务入口执行；新增 catalog 和 route dispatch 测试覆盖。
- 2026-05-27：缓存一致性判断：权威数据源仍为 registry、active release、runtime revision 和既有派生缓存；本变更只增加分布式互斥、失败态投影和 stale 恢复路径，不新增第二套缓存状态源。失败升级保持旧 active release，失败目标 release 不成为 runtime revision、enabled snapshot、frontend bundle、runtime i18n 或 Wasm 缓存权威来源。
- 2026-05-27：最终 Go 验证通过：`cd apps/lina-core && go test ./internal/service/plugin/internal/wasm -count=1`、`go test ./internal/service/plugin/internal/lifecycle -count=1`、`go test ./internal/service/plugin/internal/catalog -count=1`、`go test ./internal/service/plugin/internal/runtime -count=1`、`go test ./internal/service/plugin -count=1`、`go test ./internal/cmd -count=1`。
- 2026-05-27：运行`openspec validate harden-dynamic-plugin-runtime-safety --strict`通过；运行`git diff --check`通过。
- 2026-05-27：静态检索确认`runtime.New(`调用点仅在插件服务装配和测试服务装配两处，均已传入`locker.New()`；未修改`apps/lina-plugins/<plugin-id>/`，不触发插件本地`AGENTS.md`读取；未新增 SQL 文件、目录级 README、前端页面、DTO、`g.Meta`、用户可见运行时文案或开发工具脚本。API 契约影响为复用既有动态路由失败响应，不新增 HTTP API 或响应字段。
- 2026-05-27：数据权限影响确认：本变更不新增列表、详情、导出、下载、聚合或写操作 HTTP API，不修改动态插件 host service 数据访问授权路径；无新增数据权限过滤测试需求。接口性能和`N+1`影响确认：新增数据库访问发生在生命周期事务、单插件协调锁和失败态投影路径，不在列表/批量数据装配循环中引入逐行查询。
- 2026-05-27：已执行`lina-review`审查。审查范围来自`git status --short`、`git ls-files --others --exclude-standard`和当前 OpenSpec 变更，覆盖 16 个已跟踪 Go 文件、1 个新增 Go 测试文件和 9 个 OpenSpec 文件；重新读取`AGENTS.md`及命中规则文件`openspec.md`、`documentation.md`、`architecture.md`、`plugin.md`、`api-contract.md`、`backend-go.md`、`database.md`、`cache-consistency.md`、`data-permission.md`、`testing.md`、`i18n.md`。结论：未发现阻塞问题；开发工具、前端 UI、插件源码目录和目录级 README 均未命中。

## Feedback

- [x] **FB-1**: 固化新增运行期依赖的 DI 来源检查和`lina-review`显式依赖注入审查门禁
- [x] **FB-2**: 修正`plugin.New`中临时创建`locker.Service`导致的显式依赖注入违规

## Feedback Execution Records

- 2026-05-27：FB-1 根因确认：本次动态插件协调器新增`locker.Service`依赖时，只检查了`runtime.New(`调用点是否传入依赖，没有追溯依赖 owner、创建位置、传递路径和共享实例策略；审查结论也只记录调用点已传入`locker.New()`，未确认该关键服务没有在业务构造函数中临时创建独立服务图。
- 2026-05-27：FB-1 已更新`.agents/rules/openspec.md`，要求新增或修改运行期依赖、服务构造函数、启动装配、插件宿主服务适配器或`WASM host service`时，任务完成记录必须包含 DI 来源检查，覆盖 owner、创建位置、传递路径、是否复用启动期共享实例或共享后端；无新增运行期依赖时必须记录无影响判断。
- 2026-05-27：FB-1 已更新`.agents/rules/backend-go.md`，要求`lina-review`显式依赖注入审查追溯新增或修改运行期依赖从消费者到启动装配层的完整来源，并检查新增或修改的`New()`、`NewV1()`、插件宿主服务适配器构造函数和`WASM host service`构造函数中没有临时调用关键服务的`New()`创建独立服务图。
- 2026-05-27：FB-1 影响分析：本次为项目治理规则和 OpenSpec 任务记录变更，不修改 Go 生产代码、HTTP API、SQL、前端 UI、插件源码目录、运行时用户可见文案、语言包、缓存实现、数据权限路径或开发工具脚本；无新增运行期行为测试需求，采用 OpenSpec 严格校验、Markdown 静态检查和审查结论作为治理验证。
- 2026-05-27：FB-2 根因确认：`plugin.New`在插件服务内部装配`runtime.Service`时直接调用`locker.New()`，只满足了`runtime.New`调用点参数传递，却绕过启动装配层，违反显式依赖注入和跨实例协调依赖共享要求。
- 2026-05-27：FB-2 修复：`plugin.New`新增显式`locker.Service`参数并校验非空，`runtime.New`改用该注入依赖；`newHTTPRuntime`在启动装配层创建`lockerSvc := locker.New()`，并复用同一实例传给`pluginsvc.New`和`hostlock.New`；受签名影响的测试 fixture 同步显式传入测试级`locker.Service`。
- 2026-05-27：FB-2 DI 来源检查：新增依赖 owner 为`locker.Service`；生产创建位置为`apps/lina-core/internal/cmd/cmd_http_runtime.go`启动装配层；传递路径为`newHTTPRuntime` → `pluginsvc.New` → `runtime.New`；共享实例策略为同一`lockerSvc`同时供动态插件 reconciler lock 和 WASM host lock 使用，底层继续由`locker.ConfigureCoordination`按部署选择 SQL 或 coordination 后端。
- 2026-05-27：FB-2 影响分析：本次修改 Go 生产代码和测试 fixture，属于运行期依赖注入修复；不新增 HTTP API、DTO、SQL、前端 UI、插件源码目录、运行时用户可见文案、语言包或数据权限路径。缓存一致性影响为正向收敛：跨实例协调依赖改为启动期共享实例；无新增缓存源或失效路径。开发工具跨平台无影响。
- 2026-05-27：FB-2 已通过验证：`cd apps/lina-core && go test ./internal/cmd -count=1`、`go test ./internal/service/auth -count=1`、`go test ./internal/controller/i18n -count=1`、`go test ./internal/service/user -count=1`、`go test ./internal/service/plugin/internal/runtime -count=1`、`go test ./internal/service/plugin -count=1`、`go test ./internal/service/plugin/... -count=1`；`openspec validate harden-dynamic-plugin-runtime-safety --strict`和`git diff --check`通过。
