## ADDED Requirements

### Requirement: 动态插件升级失败必须保留旧有效发布并记录目标失败诊断

系统 SHALL 在动态插件升级或同版本刷新失败时保留升级前的有效发布和 active release。失败的目标发布、artifact 校验和、生命周期阶段、原始错误和 rollback 错误 MUST 被记录为可诊断状态；系统不得将失败目标发布切换为有效发布。

#### Scenario: 动态插件升级 SQL 失败保留旧发布
- **WHEN** 动态插件 P 从 release A 升级到 release B
- **AND** release B 的升级 SQL 执行失败
- **THEN** P 的有效发布继续指向 release A
- **AND** release B 记录升级失败诊断
- **AND** 系统不得暴露 release B 的动态路由、前端资源或 runtime i18n 作为有效能力

#### Scenario: 同版本刷新 rollback 失败保留旧产物并记录诊断
- **WHEN** 动态插件 P 以同版本新 artifact 刷新
- **AND** 刷新失败后的 rollback 也失败
- **THEN** P 的 active release 继续指向刷新前 artifact
- **AND** 系统记录刷新原始失败和 rollback 失败诊断
- **AND** 后续协调不得把失败 artifact 误判为成功刷新

### Requirement: 动态插件升级失败后的运行时缓存不得指向失败目标

系统 SHALL 在动态插件升级或同版本刷新失败后，确保 runtime revision、enabled snapshot、frontend bundle、runtime i18n 和 Wasm 编译缓存继续以旧有效发布或明确失败状态为准。失败目标发布不得成为派生缓存的权威来源。

#### Scenario: 失败升级不刷新为目标缓存
- **WHEN** 动态插件 P 升级到 release B 失败
- **THEN** 系统不得发布使其他节点加载 release B 为有效发布的 runtime revision
- **AND** 其他节点继续使用 release A 的有效缓存或按失败状态隐藏 P

#### Scenario: rollback 失败时采用保守暴露策略
- **WHEN** 动态插件 P 升级失败且 rollback 恢复失败
- **THEN** 系统不得暴露失败目标发布的能力
- **AND** 系统根据权威 active release 可用性继续使用旧发布或隐藏该插件能力
