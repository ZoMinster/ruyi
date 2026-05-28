## Why

`temp/plugin-design-review-20260527.md` 中的 P0 审查点暴露了动态插件运行时在 WASM 沙箱资源边界、插件生命周期 SQL 原子性、回滚诊断和集群协调恢复上的缺口。动态插件是 LinaPro 作为`面向可持续交付的 AI 原生全栈框架`的核心扩展面，这些缺口会影响宿主稳定性、数据完整性和多节点一致性，需要作为独立变更收敛。

## What Changes

- 为动态插件 WASM bridge 增加宿主侧默认执行超时和内存上限；调用方已提供 deadline 时尊重调用方 deadline，无 deadline 时由 bridge 统一兜底。
- 将动态插件 HTTP 路由、cron discovery、cron job、生命周期回调等所有 WASM 执行入口纳入同一资源边界，不让单个插件执行永久占用 goroutine 或无边界占用宿主内存。
- 将插件 install、upgrade、uninstall、rollback 等生命周期 SQL 执行与迁移账本记录放入一致的事务边界；在 PostgreSQL 默认运行环境下要求 SQL 与 `sys_plugin_migration` 账本同成同败。
- 强化动态插件安装、升级和刷新失败后的回滚诊断；rollback SQL、菜单恢复或前端/权限治理回滚失败必须进入错误链或失败状态，不得只写 warning 后丢失。
- 在集群模式下为动态插件协调器的共享生命周期副作用增加 per-plugin 分布式互斥，复用宿主 coordination/locker 能力，避免领导权切换或并发触发时重复执行 SQL、菜单同步或发布状态写入。
- 增加 stale `reconciling` 状态检测与恢复策略，使崩溃、panic 或进程退出后遗留的过期瞬态状态可被后续协调 tick 明确修复或重新收敛。
- 增加协调器 tick 级 panic recovery，保证单次插件协调异常不会永久杀死后台协调 goroutine。
- 不新增 HTTP API、前端页面、插件 manifest 字段或用户可见运行时文案；若实现阶段需要新增配置项，应复用宿主配置治理和 `time.Duration` 解析规则。

## Capabilities

### New Capabilities

- 无。本变更收紧既有动态插件运行时、生命周期、升级治理和分布式锁能力，不新增独立产品能力。

### Modified Capabilities

- `plugin-runtime-loading`：增加动态插件 WASM 执行超时、内存上限、协调器分布式互斥、stale `reconciling` 恢复和 panic recovery 要求。
- `plugin-manifest-lifecycle`：增加插件生命周期 SQL 事务边界、迁移账本一致性和 rollback 失败诊断要求。
- `plugin-upgrade-governance`：要求动态插件升级、同版本刷新和失败回滚路径保留旧有效发布，并记录失败发布与 rollback 失败诊断。
- `distributed-locker`：明确动态插件协调器在集群模式下应使用现有 distributed locker/coordination lock 保护 per-plugin 生命周期副作用。

## Impact

- 影响后端运行时：`apps/lina-core/internal/service/plugin/internal/wasm/**`、`apps/lina-core/internal/service/plugin/internal/runtime/**`、`apps/lina-core/internal/service/plugin/internal/lifecycle/**`。
- 影响宿主启动装配和配置读取：如果新增 WASM 执行资源配置，需要同步 `apps/lina-core/internal/service/config/**`、`manifest/config/config.yaml`和`manifest/config/config.template.yaml`。
- 影响数据库操作路径：生命周期 SQL 与 `sys_plugin_migration` 账本需要事务化；原则上不新增数据库表或列，若实现需要索引或状态字段，必须通过当前迭代 SQL 文件和 DAO 生成流程完成。
- 影响缓存与集群一致性：动态插件状态、enabled snapshot、runtime revision、frontend bundle、runtime i18n 和 Wasm 编译缓存仍以现有权威数据源和 revision 机制为准；本变更补充分布式互斥和恢复路径，不引入第二套状态源。
- 数据权限影响：不新增业务数据读写接口；动态插件 host service 数据访问继续遵守既有 `hostServices` 授权、数据权限和租户边界。
- `i18n`影响：默认无运行时用户可见文案、菜单、API 文档源文本或语言包资源变更；若实现阶段新增用户可见错误码或 API 文档文本，必须同步维护对应资源。
- 测试影响：需要新增或更新后端单元测试、集成测试或可控替身测试，覆盖 WASM 超时/内存限制、SQL 失败回滚、rollback 失败诊断、分布式锁跳过并发协调、stale `reconciling` 恢复和 reconciler panic recovery。
