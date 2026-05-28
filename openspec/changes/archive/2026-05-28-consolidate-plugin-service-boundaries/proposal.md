## Why

当前插件系统的宿主私有实现散落在`apps/lina-core/internal/service/plugin`、`apps/lina-core/internal/service/pluginhostservices`和`apps/lina-core/internal/service/pluginruntimecache`三个 service 组件目录中。随着插件运行时、源码插件宿主能力、动态插件 host service 和缓存一致性能力持续扩展，这种分散结构增加了开发者理解宿主插件边界和维护依赖方向的成本。

本变更将插件系统相关核心逻辑收敛到宿主插件服务边界内，并把公共插件契约继续保留在`apps/lina-core/pkg/plugin`体系下，使 LinaPro 的插件扩展能力更符合“面向可持续交付的 AI 原生全栈框架”的长期维护目标。

## What Changes

- 将源码插件宿主服务适配器从`apps/lina-core/internal/service/pluginhostservices`收敛到`apps/lina-core/internal/service/plugin/internal/hostservices`，并由`plugin`根包提供启动期使用的窄构造入口。
- 将插件运行时缓存修订号协调从`apps/lina-core/internal/service/pluginruntimecache`收敛到`apps/lina-core/internal/service/plugin/runtimecache`，保留给`plugin`、`plugin/internal/runtime`和`i18n`等宿主内部组件复用的受控入口。
- 收敛`apps/lina-core/internal/service/plugin`根目录中的大体量实现文件，将管理列表、运行时升级、治理校验和依赖装配等逻辑逐步下沉到职责明确的`plugin/internal/<subcomponent>`子组件。
- 保持动态插件`plugin.yaml`中的`hostServices`字段、授权快照语义、`pluginbridge`协议、HTTP API、SQL、前端 UI 和插件公共 SDK 路径不变。
- 不引入通用 DI 容器、全局 service locator 或新的宿主私有组装层；启动期仍显式注入共享运行期服务实例和共享后端。

## Capabilities

### New Capabilities

- 无。本变更是宿主插件服务边界治理和内部组织结构收敛，不新增业务能力。

### Modified Capabilities

- `core-host-boundary-governance`：补充宿主插件系统私有实现应收敛到`internal/service/plugin`边界内，`plugin`根包应作为启动期和外部宿主模块的稳定 facade。
- `plugin-host-service-extension`：补充源码插件宿主服务适配器的目录归属、构造入口和共享实例要求。
- `plugin-runtime-loading`：补充插件运行时缓存协调组件的归属要求，并明确迁移不得改变`plugin-runtime` revision、scope、observed revision 或 conservative-hide 语义。
- `service-dependency-injection-governance`：补充插件服务边界收敛时的启动期依赖注入、共享实例和导入边界要求。

## Impact

- 影响后端 Go 包组织：`apps/lina-core/internal/service/plugin/**`、`apps/lina-core/internal/service/pluginhostservices/**`、`apps/lina-core/internal/service/pluginruntimecache/**`、`apps/lina-core/internal/service/i18n/**`、`apps/lina-core/internal/cmd/**`。
- 影响插件宿主服务适配器、动态插件 WASM host service 配置、源码插件 registrar 能力目录、插件运行时缓存刷新和管理列表缓存相关测试。
- 不修改 HTTP API、数据库表结构、SQL 迁移、前端页面、插件 manifest 字段、动态插件授权语义或运行时用户可见文案。
- 缓存一致性影响：要求保留现有`plugin-runtime`权威数据源、共享 revision、scope、最大陈旧窗口、故障回退和每个本地缓存域独立 observed revision。
- 数据权限影响：不新增数据操作接口；源码插件和动态插件通过宿主发布服务访问数据的既有数据权限边界必须保持不变。
- `i18n`影响：不新增或修改语言资源；但`i18n`运行时 bundle freshness 仍必须通过迁移后的插件运行时缓存协调入口观察同一`plugin-runtime` revision。
