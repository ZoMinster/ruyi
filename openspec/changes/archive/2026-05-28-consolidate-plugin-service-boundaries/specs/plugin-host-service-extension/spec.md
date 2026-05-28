## ADDED Requirements

### Requirement: 源码插件宿主服务适配器必须归属 plugin 内部 hostservices 子组件

系统 SHALL 将源码插件宿主服务适配器实现归属到`apps/lina-core/internal/service/plugin/internal/hostservices`子组件。该子组件负责把宿主启动期共享的`auth`、`apidoc`、`bizctx`、`datascope`、`i18n`、`notify`、`session`、`kvcache`、`orgcap`、`tenantcap`和插件生命周期能力适配为`pkg/plugin/capability.Services`。`apps/lina-core/internal/service/pluginhostservices`不得作为长期生产入口保留。

#### Scenario: 启动期构造源码插件宿主服务目录
- **WHEN** 宿主 HTTP runtime 需要构造源码插件可消费的宿主服务目录
- **THEN** 启动期通过`internal/service/plugin`根包暴露的窄构造入口创建`capability.Services`
- **AND** 该入口委托`plugin/internal/hostservices`完成具体适配
- **AND** `internal/cmd`不得直接导入`plugin/internal/hostservices`

#### Scenario: 适配器复用共享运行期实例
- **WHEN** `plugin/internal/hostservices`构造`capability.Services`
- **THEN** 所有接口型运行期依赖必须由启动期逐项显式传入
- **AND** 适配器不得在构造函数、插件回调路径或 host service 调用路径中创建独立的`auth`、`session`、`plugin`、`i18n`、`notify`、`kvcache`、`orgcap`或`tenantcap`服务实例

#### Scenario: 源码插件获取插件作用域能力
- **WHEN** 源码插件 registrar、hook、route 或 cron 回调需要插件作用域的 host services
- **THEN** 其获取的目录仍满足`capability.Services`和必要的`pluginhost.Services`契约
- **AND** cache、config 和 manifest 等插件作用域能力继续按插件 ID 绑定

### Requirement: hostservices 子组件不得反向依赖 plugin 根包

系统 SHALL 保证`plugin/internal/hostservices`不导入`apps/lina-core/internal/service/plugin`根包。需要读取动态路由元数据或其他插件运行时上下文时，hostservices MUST 依赖真实 owner 的窄接口、窄 helper 或由`plugin`根包构造入口显式注入的 resolver。

#### Scenario: route adapter 解析动态路由元数据
- **WHEN** 源码插件通过宿主服务目录读取当前动态路由元数据
- **THEN** hostservices 通过运行时子组件的窄能力或显式注入的 resolver 完成读取
- **AND** hostservices 不得通过 import `internal/service/plugin`调用 facade alias

#### Scenario: 导入边界审查
- **WHEN** 审查 hostservices 迁移后的生产 Go 代码
- **THEN** 静态检索不得发现`plugin/internal/hostservices`导入`internal/service/plugin`
- **AND** 静态检索不得发现生产代码继续导入旧`internal/service/pluginhostservices`
