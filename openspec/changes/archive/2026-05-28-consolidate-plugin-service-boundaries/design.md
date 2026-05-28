## Context

`apps/lina-core`当前已经把大部分插件宿主实现收敛到`internal/service/plugin`下，并通过`internal/service/plugin/internal/{catalog,runtime,integration,lifecycle,frontend,openapi,wasm}`等子组件承载具体职责。但仍有两个插件系统核心组件位于`internal/service`根层级：

- `internal/service/pluginhostservices`：负责把宿主运行期的`auth`、`bizctx`、`i18n`、`notify`、`session`、`kvcache`、`orgcap`、`tenantcap`等服务适配成插件可消费的`capability.Services`。
- `internal/service/pluginruntimecache`：负责`plugin-runtime`缓存域的共享修订号控制、本地 observed revision 和跨节点刷新。

同时，`internal/service/plugin`根目录仍保留大量跨职责 Go 文件，既包含`Service` facade 和构造装配，也包含管理列表、依赖治理、运行时升级、平台治理、启动一致性等具体实现。该结构会让新开发者难以判断哪些文件是公开宿主 facade，哪些只是插件系统内部实现。

本变更属于宿主通用插件扩展能力的内部边界治理，不是工作台展示适配，不修改 HTTP API、SQL、前端 UI 或插件 manifest 字段。

## Goals / Non-Goals

**Goals:**

- 将插件系统宿主私有实现统一收敛到`apps/lina-core/internal/service/plugin`边界下。
- 保持`apps/lina-core/pkg/plugin`作为宿主与插件共享的公共契约、SDK、bridge 和 capability 组件边界。
- 让`plugin`根包成为启动装配、控制器、中间件、i18n 等宿主内部调用方可依赖的稳定 facade。
- 将`pluginhostservices`迁入`plugin/internal/hostservices`，并通过`plugin`根包暴露窄构造入口。
- 将`pluginruntimecache`迁入`plugin/runtimecache`，保留给`plugin`、`plugin/internal/runtime`和`i18n`复用的受控内部公共入口。
- 逐步将`plugin`根目录的具体实现下沉到`plugin/internal/<subcomponent>`，降低根目录文件数量和职责噪声。
- 保留现有缓存一致性、数据权限、授权、`i18n` freshness 和共享实例语义。

**Non-Goals:**

- 不重命名动态插件`plugin.yaml`中的`hostServices`字段。
- 不改变动态插件 host service service/method 字符串、授权快照、bridge envelope 或 guest SDK 公共行为。
- 不新增或修改 HTTP API、DTO、OpenAPI 元数据、SQL 迁移、前端页面或运行时用户可见文案。
- 不引入通用 DI 容器、全局 service locator、聚合依赖结构体或新的宿主私有启动编排层。
- 不在本变更中重构`apps/lina-core/pkg/plugin`公共包层级。

## Decisions

### Decision 1: `pluginhostservices`迁入`plugin/internal/hostservices`

`pluginhostservices`不是独立业务模块，而是宿主插件服务对源码插件和能力目录发布的适配器集合。其实现依赖宿主内部`auth`、`apidoc`、`bizctx`、`datascope`、`i18n`、`notify`、`session`和`kvcache`等服务，并返回`pkg/plugin/capability.Services`。将它迁入`plugin/internal/hostservices`可以清晰表达它属于插件宿主服务实现细节。

启动装配层不能直接 import `plugin/internal/hostservices`，因此`plugin`根包提供一个窄入口，例如`NewHostServices(...) (capability.Services, error)`。该入口只做参数校验、轻量编排和委托，不持有新状态，也不创建共享服务依赖之外的孤立服务图。

替代方案是保留`pluginhostservices`独立目录。该方案迁移成本最低，但继续让插件核心逻辑分散在`internal/service`根层级，不能解决开发者理解成本问题。

### Decision 2: `pluginruntimecache`迁入`plugin/runtimecache`而不是`plugin/internal/runtimecache`

`pluginruntimecache`当前同时被`plugin`根包、`plugin/internal/runtime`和`i18n`使用。若迁入`plugin/internal/runtimecache`，`i18n`因 Go `internal`可见性规则无法导入；若让`i18n`改为依赖`plugin`根包，又会扩大`plugin`facade并增加循环依赖风险。

因此目标路径为`apps/lina-core/internal/service/plugin/runtimecache`。该路径仍位于插件服务边界内，但不是`internal`子包，允许`apps/lina-core/internal/service/i18n`在宿主内部通过受控入口复用同一个`plugin-runtime`缓存协调协议。

迁移必须保持以下语义不变：

- 缓存域仍为`plugin-runtime`。
- 普通运行时缓存变更和 reconciler wake-up 继续使用不同 change reason 和 scope。
- `plugin`根包、`plugin/internal/runtime`和`i18n`各自维护独立`ObservedRevision`。
- 集群模式继续复用宿主统一`cachecoord`和拓扑抽象。
- revision freshness 不可确认时继续执行 conservative-hide 或对应调用方既有降级策略。

替代方案是把该能力提升到`internal/service/cachecoord`或`pkg`。该方案会削弱插件运行时语义归属，且`pkg`不应承载宿主私有缓存实现。

### Decision 3: `plugin`根包只保留 facade、公开投影和轻量装配

`plugin.go`作为组件主文件，应保留`Service`契约、公开类型别名、核心 DTO、`serviceImpl`字段、`New()`构造和必要的启动期 facade。具体业务流程应进入同包其他文件或`plugin/internal/<subcomponent>`。

本变更按风险从低到高分阶段迁移：

1. 先迁移独立适配器和缓存协调包。
2. 再迁移管理列表、运行时升级、平台治理、启动一致性和依赖装配等相对可分割逻辑。
3. 最后评估`plugin_lifecycle.go`等核心跨组件编排是否需要继续下沉。

不要求一次性把所有根文件清空。核心原则是每个新增子组件必须有明确职责，并且能降低当前理解成本，而不是为了减少文件数量制造转发型抽象。

### Decision 4: 路由元数据依赖通过窄接口消除反向 import

当前`pluginhostservices`中的 route adapter 通过 import `internal/service/plugin`访问动态路由元数据 helper。迁入`plugin/internal/hostservices`后若保留该依赖会形成不合理依赖方向。实施时需要将动态路由元数据读取改为以下任一方式：

- 直接依赖`plugin/internal/runtime`中稳定的窄 helper。
- 或通过`NewHostServices`显式注入一个窄的 route metadata resolver。

优先选择直接复用`plugin/internal/runtime`的窄 helper，前提是不会扩大运行时子组件导出面到具体内部状态。若该 helper 当前只由`plugin`根包 alias 暴露，则将 alias 调用点迁移到真实 owner。

### Decision 5: 分阶段实施和验证

本变更会拆成四个实施阶段：

1. 合并 host services 适配器。
2. 迁移 runtime cache 协调组件。
3. 瘦身`plugin`根目录。
4. 清理旧路径、测试、静态检索和 OpenSpec 校验。

每个阶段完成后必须运行可覆盖当前阶段的 Go 测试和静态导入扫描。涉及启动装配或构造入口的阶段必须运行`cd apps/lina-core && go test ./internal/cmd -count=1`或等价启动绑定编译门禁。

## Risks / Trade-offs

- `plugin/internal/hostservices`不可被`internal/cmd`直接导入 → 通过`plugin.NewHostServices(...)`作为窄 facade，启动层只依赖`plugin`根包。
- route metadata helper 依赖方向错误 → 在迁移 host services 前先消除`hostservices -> plugin`反向 import，改为窄 helper 或显式 resolver。
- `runtimecache`迁移破坏`i18n` freshness → 使用`plugin/runtimecache`而非`plugin/internal/runtimecache`，并保留`i18n`独立 observed revision 测试。
- 缓存 observed revision 被误共享 → 在任务中要求单元测试覆盖`plugin`根运行时缓存、`runtime` reconciler 和`i18n` bundle freshness 各自独立观察同一共享 revision。
- 纯搬目录造成大量导入变更但没有架构收益 → 每次迁移必须对应明确目标子组件职责，并用静态扫描证明旧根目录和旧独立 service 包不再承担相同职责。
- 历史 OpenSpec 基线仍残留旧`pkg/pluginservice`文字 → 本变更在相关 delta 中记录目标态，实施阶段通过静态检索和必要规范修正避免继续扩大旧命名。

## Migration Plan

1. 建立`plugin/internal/hostservices`并迁移`pluginhostservices`代码和测试。
2. 在`plugin`根包新增 host services 构造 facade，并将`internal/cmd`和测试调用方迁移到该 facade。
3. 消除 host services 对`plugin`根包的反向导入。
4. 建立`plugin/runtimecache`并迁移`pluginruntimecache`代码和测试，更新`plugin`、`plugin/internal/runtime`、`i18n`和相关测试导入。
5. 按职责建立`plugin/internal/management`、`plugin/internal/runtimeupgrade`、`plugin/internal/governance`等子组件，并迁移低耦合逻辑。
6. 删除旧`internal/service/pluginhostservices`和`internal/service/pluginruntimecache`目录。
7. 运行 Go 测试、静态导入扫描、OpenSpec 严格校验和格式检查。

回退策略为按阶段回退：每个阶段只改变一类边界，若测试或导入扫描失败，可保留已通过阶段并回退当前阶段的目录迁移。

## Open Questions

- `plugin_lifecycle.go`是否在本变更内继续下沉，取决于阶段三完成后根目录剩余文件的职责密度和新增接口数量。若下沉需要引入大量转发接口，应保留在根包或另开变更评估。
- `runtimeupgrade`子组件是否直接持有`catalog/lifecycle/runtime/integration`依赖，还是只作为根包私有 helper 包，需要在实施阶段根据循环依赖情况决定。
