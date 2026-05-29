## 为什么

LinaPro 需要一个稳定、可扩展且可治理的插件平台，使业务能力能够通过源码插件和动态插件持续交付，而不是反复侵入`apps/lina-core`核心宿主。插件平台必须同时覆盖清单契约、生命周期、动态`WASM`运行时、前后端集成、宿主服务授权、依赖与升级治理、启动自动化、包边界和多节点一致性，才能支撑“面向可持续交付的`AI`原生全栈框架”定位。

早期方案同时牵引了集群、发布、菜单、角色、认证、监控、`E2E`和开发工具等能力。压缩后的归档将插件框架自身作为长期历史 owner，只保留插件平台的核心演进和决策；其他能力只在设计文档中保留交叉影响摘要，当前契约以`openspec/specs/<capability>/spec.md`为准。

## 变更内容

- 建立统一插件契约，以`plugin.yaml`、源码插件目录和动态插件发布产物作为插件身份、资源、依赖、菜单、权限和生命周期的事实入口。
- 定义源码插件和动态插件生命周期，覆盖发现、安装、启用、禁用、卸载、升级、同版本刷新、租户级生命周期和失败诊断。
- 构建动态`WASM`插件运行时，包括自定义段解析、运行时资源视图、前端资产、动态路由、桥接协议、生命周期自动发现和执行资源边界。
- 建立统一宿主服务模型，通过`hostServices`授权快照和`pkg/plugin/capability`能力目录向动态插件和源码插件暴露受治理的配置、manifest、数据、缓存、锁、网络、通知、组织、租户和业务上下文能力。
- 收敛插件公共包边界：`pluginhost`负责源码插件贡献，`pluginbridge`负责动态插件`ABI`与 transport，`capability`负责插件消费宿主能力；宿主插件运行时治理保留在`internal/service/plugin`。
- 支持插件页面、菜单、公开资产、工作台动态路由和插件管理读模型，并将统一插件`API`命名空间收敛到`/x/{plugin-id}/api/v1/...`，公开资产收敛到`/x-assets/{plugin-id}/{version}/...`。
- 引入`plugin.autoEnable`启动引导、事务性 mock data 安装、安装并启用快捷操作、插件依赖检查、运行时升级预览和显式升级执行。
- 支持官方插件工作区可选化和插件工作区管理命令，使宿主可在 host-only 与 plugin-full 模式下分别构建、测试和发布。

## Capabilities

### New Capabilities

- `plugin-manifest-lifecycle`
- `plugin-runtime-loading`
- `plugin-host-service-extension`
- `plugin-capability-boundary-governance`
- `plugin-package-boundary-governance`
- `pluginbridge-subcomponent-architecture`
- `plugin-config-service`
- `plugin-data-service`
- `plugin-cache-service`
- `plugin-lock-service`
- `plugin-network-service`
- `plugin-notify-service`
- `plugin-storage-service`
- `plugin-hook-slot-extension`
- `plugin-ui-integration`
- `plugin-embed-snapshot-packaging`
- `plugin-id-governance`
- `plugin-dependency-management`
- `plugin-startup-bootstrap`
- `plugin-mock-data-installation`
- `plugin-install-enable-shortcut`
- `plugin-runtime-upgrade`
- `plugin-upgrade-governance`
- `plugin-workspace-management`
- `official-plugin-workspace-decoupling`
- `framework-capability-registry`
- `workspace-route-boundary`
- `plugin-api-query-performance`
- `plugin-permission-governance`

### Modified Capabilities

- `menu-management`、`role-management`、`user-auth`、`cron-jobs`、`cluster-deployment-mode`、`cluster-topology-boundaries`、`distributed-locker`、`leader-election`、`project-setup`、`release-image-build`、`e2e-suite-organization`、`server-monitor`、`online-user`、`core-host-boundary-governance`、`module-decoupling`、`service-dependency-injection-governance`、`source-upgrade-governance`和`system-api-docs`只保留插件相关交叉影响摘要，不再由本分组长期保存完整规范全文。

## Impact

- 后端影响集中在插件注册、运行时加载、生命周期编排、`WASM`桥接、host service、能力目录、启动引导、升级治理、缓存一致性和插件管理读模型。
- 前端影响集中在插件管理、动态页面承载、插件菜单和路由刷新、公开资产引用、安装与升级弹窗，以及插件状态变化后的用户体验保护。
- 数据和配置影响集中在插件治理表、发布快照、迁移账本、资源引用、`plugin.autoEnable`、插件运行期配置、manifest 资源和动态产物快照。
- 构建与工具影响集中在`build-wasm`、host-only/plugin-full 构建测试、插件工作区管理命令和发布链路。
- 本归档压缩不修改运行时代码、数据库、`API`、前端页面或插件源码；当前能力契约以`openspec/specs`为准，归档仅保留历史设计和治理原因。
