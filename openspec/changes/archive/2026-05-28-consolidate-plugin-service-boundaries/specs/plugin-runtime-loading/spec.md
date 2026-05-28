## ADDED Requirements

### Requirement: 插件运行时缓存协调组件必须归属 plugin runtimecache 子组件

系统 SHALL 将`plugin-runtime`缓存域的 revision controller、observed revision、change reason、scope 和 domain policy 实现归属到`apps/lina-core/internal/service/plugin/runtimecache`子组件。该子组件属于宿主插件服务边界，但 MUST 可被`plugin`根包、`plugin/internal/runtime`和`i18n`等宿主内部组件通过受控路径复用。旧`apps/lina-core/internal/service/pluginruntimecache`不得作为长期生产入口保留。

#### Scenario: i18n 观察插件运行时修订号
- **WHEN** `i18n`运行时消息包需要确认 source plugin 或 dynamic plugin 资源 freshness
- **THEN** 它通过`plugin/runtimecache`创建或持有自身的 revision controller
- **AND** 它不依赖`plugin/internal/runtimecache`
- **AND** 它不通过导入`plugin`根包绕过真实缓存协调 owner

#### Scenario: runtime reconciler 使用独立 scope
- **WHEN** 动态插件 reconciler 发布或观察 wake-up revision
- **THEN** 它继续通过`plugin/runtimecache`使用 reconciler scope 和 reconciler change reason
- **AND** 该 scope 不得与普通插件运行时缓存失效 scope 混用

#### Scenario: 旧缓存协调包被移除
- **WHEN** runtime cache 迁移完成
- **THEN** 生产 Go 代码不得继续 import `lina-core/internal/service/pluginruntimecache`
- **AND** 测试和 panic allowlist 等治理文件必须同步到新路径或说明已删除

### Requirement: 插件运行时缓存迁移不得改变一致性语义

系统 SHALL 在迁移`plugin-runtime`缓存协调组件时保持现有一致性语义不变。迁移不得改变权威数据源、缓存域名称、change reason、scope、最大可接受陈旧时间、故障回退策略、跨实例同步机制或各调用方的本地 observed revision 独立性。

#### Scenario: 多个本地缓存域独立观察同一 revision
- **WHEN** `plugin`根 facade、`plugin/internal/runtime` reconciler 和`i18n`运行时 bundle 分别消费`plugin-runtime`revision
- **THEN** 每个调用方维护自己的`ObservedRevision`
- **AND** 一个调用方记录 observed revision 不得让另一个调用方跳过自身 refresh 或 invalidate

#### Scenario: 集群模式继续复用统一 cachecoord
- **WHEN** `cluster.enabled=true`且插件安装、启用、禁用、卸载、升级或 active release 切换发布运行时变更
- **THEN** 系统继续通过宿主统一`cachecoord`后端发布`plugin-runtime`revision 和 event
- **AND** 其他节点继续按现有路径刷新 enabled snapshot、frontend bundle、runtime i18n 和 Wasm 派生缓存

#### Scenario: freshness 不可确认时保持 conservative-hide
- **WHEN** 节点无法确认`plugin-runtime`revision freshness
- **THEN** 动态插件能力继续按既有 conservative-hide 或调用方定义的安全降级处理
- **AND** 迁移不得因包路径变化退化为继续暴露可能已禁用、卸载或权限变化的插件能力
