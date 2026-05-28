## ADDED Requirements

### Requirement: 插件服务边界收敛必须保持启动期显式依赖注入

系统 SHALL 在收敛插件服务边界时继续使用启动期显式依赖注入。`plugin`根包新增的任何 facade 构造入口 MUST 逐项接收接口型运行期依赖并委托内部子组件，不得通过聚合依赖结构体、全局 service locator、隐式`New()`或包级默认实例补齐`auth`、`session`、`plugin`、`i18n`、`cachecoord`、`kvcache`、`notify`、`orgcap`、`tenantcap`等关键依赖。

#### Scenario: NewHostServices 构造源码插件能力目录
- **WHEN** `plugin`根包提供源码插件 host services 构造 facade
- **THEN** 该 facade 的签名逐项接收所需宿主服务实例
- **AND** 它不得使用`Dependencies`、`Deps`或`Options`等聚合结构体承载多个接口型依赖
- **AND** 它不得在依赖缺失时临时创建关键服务实例

#### Scenario: ConfigureWasmHostServices 保持共享后端
- **WHEN** 宿主启动配置动态插件 WASM host service
- **THEN** 配置入口继续使用启动期传入的共享 cache、lock、notify、config、host services 和 manifest/config factory
- **AND** 迁移 hostservices 或 runtimecache 包路径不得导致 WASM host service 回退到包级默认孤立实例

### Requirement: 插件内部子组件导入边界必须可静态验证

系统 SHALL 通过静态检索、编译门禁或治理测试验证插件服务边界收敛后的导入方向。宿主启动层和插件外部调用方 MUST 只依赖`internal/service/plugin`根 facade 或`plugin/runtimecache`等明确允许的受控子包，不得直接依赖`plugin/internal/<subcomponent>`实现包。

#### Scenario: 启动层不导入 plugin internal 子组件
- **WHEN** 审查`apps/lina-core/internal/cmd`生产 Go 代码
- **THEN** 不得发现它导入`lina-core/internal/service/plugin/internal/`
- **AND** 它通过`lina-core/internal/service/plugin`根 facade 获取插件服务、host services 构造和 WASM host service 配置入口

#### Scenario: 旧独立插件 service 包无生产导入
- **WHEN** 审查迁移后的生产 Go 代码
- **THEN** 不得发现生产代码 import `lina-core/internal/service/pluginhostservices`
- **AND** 不得发现生产代码 import `lina-core/internal/service/pluginruntimecache`

#### Scenario: 子组件不扩大导出面规避循环依赖
- **WHEN** 插件内部实现迁入`plugin/internal/<subcomponent>`
- **THEN** 子组件只导出父组件或授权边界内调用所需的窄契约
- **AND** 不得为了测试便利、临时复用或规避循环依赖暴露缓存快照、DAO、DO、Entity、私有配置或运行时状态结构
