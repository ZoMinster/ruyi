## 1. 准备与边界确认

- [x] 1.1 复核`apps/lina-core/internal/service/pluginhostservices`、`apps/lina-core/internal/service/pluginruntimecache`、`apps/lina-core/internal/service/plugin`、`apps/lina-core/internal/service/i18n`和`apps/lina-core/internal/cmd`的当前导入关系，记录无 HTTP API、SQL、前端 UI、插件 manifest、运行时文案变更。
- [x] 1.2 确认本变更涉及 Go 后端、插件边界、缓存一致性、测试和`i18n`影响判断；确认数据权限无新增接口但宿主发布服务既有数据权限语义必须保持。
- [x] 1.3 建立迁移前静态基线，记录旧路径导入、`plugin`根目录 Go 文件数量和`pluginruntimecache`调用方列表。

### 1. 准备记录

- 规则命中：已读取`openspec`、文档、架构、插件、后端 Go、缓存一致性、测试、`i18n`和数据权限规则；本阶段无 HTTP API、SQL、前端 UI、插件 manifest、运行时用户可见文案或开发工具脚本变更。
- 影响判断：本变更涉及 Go 后端包组织、插件宿主边界、缓存 freshness 路径和单元测试；数据权限无新增接口，源码插件通过宿主发布服务访问数据的既有数据权限语义必须保持。
- 静态基线：`pluginhostservices`生产导入为`internal/cmd/cmd_http_runtime.go`，测试导入为`internal/service/user/user_test_dependencies_test.go`；`pluginruntimecache`导入位于`plugin`、`plugin/internal/runtime`和`i18n`相关 Go 文件；`plugin`根目录迁移前共有 48 个 Go 文件。

## 2. 合并源码插件宿主服务适配器

- [x] 2.1 创建`apps/lina-core/internal/service/plugin/internal/hostservices`并迁移`pluginhostservices`生产代码、测试和包注释，保持`capability.Services`、`pluginhost.Services`和 scoped services 行为不变。
- [x] 2.2 在`apps/lina-core/internal/service/plugin`根包新增源码插件 host services 构造 facade，逐项接收启动期共享依赖并委托`plugin/internal/hostservices`。
- [x] 2.3 将`internal/cmd`、`internal/service/user`测试依赖和其他调用方从`pluginhostservices.New`迁移到`plugin`根包 facade。
- [x] 2.4 消除 hostservices 对`internal/service/plugin`根包的反向导入，将动态路由元数据读取改为依赖真实 owner 的窄 helper 或显式 resolver。
- [x] 2.5 运行 hostservices 相关单元测试和启动绑定编译门禁，至少覆盖`cd apps/lina-core && go test ./internal/service/plugin/internal/hostservices ./internal/service/plugin ./internal/cmd -count=1`。
- [x] 2.6 静态检索确认生产 Go 代码不再 import `lina-core/internal/service/pluginhostservices`，且`internal/cmd`不 import `lina-core/internal/service/plugin/internal/`。

### 2. 合并记录

- 迁移结果：旧`pluginhostservices`已迁入`apps/lina-core/internal/service/plugin/internal/hostservices`，包名调整为`hostservices`；`plugin`根包新增`NewHostServices`启动期 facade，`internal/cmd`和`internal/service/user`测试依赖改为通过该 facade 构造`capability.Services`。
- 依赖方向：`hostservices`不再导入`internal/service/plugin`根包；动态路由元数据读取改为直接依赖真实 owner`plugin/internal/runtime.GetDynamicRouteMetadata`。为避免`plugin -> hostservices -> auth/apidoc -> plugin`依赖环，facade 入参使用插件公共契约和不会回指`plugin`的窄宿主接口，`cmd`启动装配负责把`auth`、`apidoc`、`notify`、`i18n`的 DTO 转换为插件契约。
- DI 来源检查：未新增运行期服务实例；`auth`、`bizctx`、`config`、`datascope`、`i18n`、`plugin`、`sessionStore`、`orgcap`、`tenantcap`、`notify`和`kvcache`仍由 HTTP 启动装配创建并逐项传入，`NewHostServices`只委托内部子组件，不使用聚合依赖结构体、全局 service locator 或隐式`New()`补齐关键依赖。
- 验证：已通过`cd apps/lina-core && go test ./internal/service/plugin/internal/hostservices ./internal/service/plugin ./internal/cmd -count=1`；静态检索`rg -n '"lina-core/internal/service/pluginhostservices"|internal/service/pluginhostservices' apps/lina-core --glob '*.go'`无结果；静态检索`internal/cmd`没有导入`plugin/internal/hostservices`。

## 3. 迁移插件运行时缓存协调组件

- [x] 3.1 创建`apps/lina-core/internal/service/plugin/runtimecache`并迁移`pluginruntimecache`生产代码、测试和包注释，保持`plugin-runtime`domain、change reason、scope、最大陈旧窗口和 failure strategy 不变。
- [x] 3.2 更新`plugin`根包、`plugin/internal/runtime`、`i18n`和相关测试的导入路径到`plugin/runtimecache`。
- [x] 3.3 补充或调整单元测试，覆盖`plugin`根运行时缓存、runtime reconciler 和`i18n`运行时 bundle freshness 各自维护独立`ObservedRevision`。
- [x] 3.4 同步治理引用，例如 panic allowlist、测试路径说明或静态扫描目标中的旧`pluginruntimecache`路径。
- [x] 3.5 运行缓存迁移相关测试，至少覆盖`cd apps/lina-core && go test ./internal/service/plugin/runtimecache ./internal/service/plugin ./internal/service/plugin/internal/runtime ./internal/service/i18n ./internal/cmd -count=1`。
- [x] 3.6 静态检索确认生产 Go 代码不再 import `lina-core/internal/service/pluginruntimecache`，且没有调用方错误依赖`plugin/internal/runtimecache`。

### 3. 迁移记录

- 迁移结果：旧`apps/lina-core/internal/service/pluginruntimecache`已迁入`apps/lina-core/internal/service/plugin/runtimecache`，包名调整为`runtimecache`；`plugin`根包、`plugin/internal/runtime`、`i18n`和相关测试均改为导入`plugin/runtimecache`。
- 缓存语义：`runtimeCacheDomain`仍为`plugin-runtime`，普通运行时变更 reason 仍为`plugin_runtime_changed`，reconciler reason 仍为`plugin_reconciler_changed`，最大陈旧窗口仍为 5 秒，failure strategy 仍为`conservative-hide`。`plugin`根、runtime reconciler 和`i18n`继续分别持有独立`ObservedRevision`实例。
- 治理同步：panic allowlist 路径已更新到`apps/lina-core/internal/service/plugin/runtimecache/runtimecache_controller.go`。
- 验证：已通过`cd apps/lina-core && go test ./internal/service/plugin/runtimecache ./internal/service/plugin ./internal/service/plugin/internal/runtime ./internal/service/i18n ./internal/cmd -count=1`；静态检索`rg -n "internal/service/pluginruntimecache|lina-core/internal/service/pluginruntimecache|pluginruntimecache" apps/lina-core --glob '*.go'`无结果；静态检索`rg -n "plugin/internal/runtimecache" apps/lina-core --glob '*.go'`无结果。

## 4. 收敛 plugin 根目录实现职责

- [x] 4.1 建立`plugin/internal/management`并迁移插件管理列表、管理列表缓存和详情投影中可下沉的实现逻辑，保留`plugin`根包公开方法和投影类型。
- [x] 4.2 建立`plugin/internal/runtimeupgrade`并迁移运行时升级预览和执行中可独立测试的规划、diff、锁名、状态转换 helper，避免向外暴露 DAO、DO、Entity 或内部缓存状态。
- [x] 4.3 建立`plugin/internal/governance`并迁移平台治理、启动一致性、租户开通策略中可分离的判断逻辑，保留根包的宿主生命周期 facade。
- [x] 4.4 评估`plugin_dependency.go`与既有`plugin/internal/dependency`的重叠，迁移纯依赖解析、快照转换和格式化 helper，避免产生只做透传的新 service。
- [x] 4.5 评估`plugin_lifecycle.go`是否适合继续下沉；若下沉会引入大量转发接口或循环依赖，则保留在根包并在任务记录中说明原因。
- [x] 4.6 运行根目录收敛相关测试，至少覆盖`cd apps/lina-core && go test ./internal/service/plugin ./internal/service/plugin/internal/management ./internal/service/plugin/internal/runtimeupgrade ./internal/service/plugin/internal/governance -count=1`，若某些子组件未建立则以实际迁移包替代。

### 4. 收敛记录

- 管理投影：已新增`apps/lina-core/internal/service/plugin/internal/management`，下沉管理列表缓存、缓存 key、列表/详情投影类型、列表克隆、manifest 快照上下文、registry 索引、类型匹配和排序 helper；`plugin`根包通过类型别名继续暴露`PluginItem`、`ListInput`和`ListOutput`，保留`List`、`Get`、`SyncAndList`、`PrewarmManagementList`等 facade 方法。`plugin_list_cache.go`已删除，`plugin`根目录 Go 文件数从阶段中 49 个收敛到 48 个。
- 运行时升级：已新增`plugin/internal/runtimeupgrade`，下沉运行时升级 SQL summary、hostServices diff、风险 hint 组装、可执行状态判断、分布式锁名称、锁 owner、锁 reason 和 lease；根包只保留业务错误包装、权限检查、catalog/runtime/lifecycle 编排和公开 preview/result 契约。子组件不暴露 DAO、DO、Entity、缓存快照或运行时内部状态。
- 治理判断：已新增`plugin/internal/governance`，下沉平台上下文检查、`sys_plugin`启动一致性 enum 组合检查、租户治理支持判断；根包继续承载启动期 capability 注入、生命周期 facade 和错误聚合。
- 依赖解析：已复用既有`plugin/internal/dependency`，新增 projection 和 snapshot helper，迁移依赖结果投影转换、blocker 格式化、首个 blocker 参数提取、reverse dependent ID 提取、registry release dependency snapshot 合成和 snapshot clone；`plugin_dependency.go`继续保留依赖解析编排、catalog 读取和业务错误封装，未新增只做透传的新 service。
- 生命周期评估：`plugin_lifecycle.go`继续保留在根包。原因是该文件跨越平台治理、依赖预检、catalog、runtime、source lifecycle、dynamic lifecycle、integration snapshot、通知和缓存发布；若本阶段下沉，需要为 catalog/runtime/integration/lifecycle/通知/缓存构造大量转发接口，反而增加循环依赖和理解成本。后续若继续收敛，应优先拆分可独立稳定的 install-mode 选择或 lifecycle authorization helper，而不是整体搬迁生命周期 facade。
- 验证：已通过`cd apps/lina-core && go test ./internal/service/plugin ./internal/service/plugin/internal/management ./internal/service/plugin/internal/runtimeupgrade ./internal/service/plugin/internal/governance ./internal/service/plugin/internal/dependency -count=1`；同时通过`cd apps/lina-core && go test ./internal/controller/plugin -count=1`确认控制器投影仍可编译。

## 5. 治理清理与验证

- [x] 5.1 删除旧`apps/lina-core/internal/service/pluginhostservices`和`apps/lina-core/internal/service/pluginruntimecache`目录，或确认仅剩迁移期空目录且无生产用途。
- [x] 5.2 静态检索确认旧路径无生产导入：`rg -n "internal/service/pluginhostservices|internal/service/pluginruntimecache|plugin/internal/runtimecache" apps/lina-core --glob '*.go'`。
- [x] 5.3 静态检索确认`internal/cmd`、控制器和其他插件外部调用方未直接导入`plugin/internal/<subcomponent>`实现包。
- [x] 5.4 静态检索并按需修正本变更触及文档或测试中的旧`pkg/pluginservice`目标态描述，确保新增记录使用`pkg/plugin/capability`。
- [x] 5.5 运行变更范围 Go 编译门禁，至少覆盖`cd apps/lina-core && go test ./internal/service/plugin/... ./internal/service/i18n ./internal/cmd -count=1`。
- [x] 5.6 运行`openspec validate consolidate-plugin-service-boundaries --strict`和本变更文件的格式检查。
- [x] 5.7 在任务记录中补充影响分析：`i18n`无语言资源变更但 freshness 路径已验证；数据权限无新增接口且宿主发布服务边界不变；缓存一致性语义不变且共享实例来源已验证；开发工具跨平台无脚本影响；测试策略为 Go 单元测试、启动绑定测试、静态检索和 OpenSpec 校验。

### 5. 清理与验证记录

- 旧目录清理：`apps/lina-core/internal/service/pluginhostservices`和`apps/lina-core/internal/service/pluginruntimecache`已不存在。
- 导入边界：静态检索`rg -n "internal/service/pluginhostservices|internal/service/pluginruntimecache|plugin/internal/runtimecache" apps/lina-core --glob '*.go'`无结果；静态检索`internal/cmd`、控制器、`auth`、`bizctx`、`config`、`i18n`和`user`等插件外部调用方，未发现生产代码直接导入`plugin/internal/<subcomponent>`。唯一命中是既有 panic allowlist 测试路径`plugin/internal/datahost/internal/host/db.go`，不属于运行期业务导入。
- `pkg/pluginservice`治理：静态检索剩余命中仅为本变更设计中的历史风险说明、任务项自身以及既有`capability_boundary_governance_test.go`中用于拒绝旧路径的样例字符串；本次新增记录和实现均使用`pkg/plugin/capability`目标路径，没有扩大旧目标态描述。
- DI 来源检查：未新增关键运行期服务实例。`hostservices`通过`plugin.NewHostServices`由`cmd`逐项传入启动期共享的`auth`、`bizctx`、`config`、`datascope`、`i18n`、`plugin`、`sessionStore`、`orgcap`、`tenantcap`、`notify`和`kvcache`能力；`runtimecache`继续复用启动期`cachecoord`和拓扑；Phase 4 新增的`management`、`runtimeupgrade`、`governance`和`dependency`helper 不持有独立服务图，不调用关键服务`New()`。
- 影响分析：`i18n`无语言资源、API 文档源文本或运行时文案变更，`i18n`runtime bundle freshness 已通过`plugin/runtimecache`路径相关测试覆盖；数据权限无新增接口、路由或数据操作入口，源码插件和动态插件通过宿主发布服务访问数据的既有边界保持；缓存一致性语义不变，`plugin-runtime`domain、scope、reason、observed revision、最大陈旧窗口和 conservative-hide 语义已保留；开发工具跨平台无脚本、`Makefile`、CI 或`linactl`变更；无 HTTP API、SQL、前端 UI、插件 manifest 字段或运行时用户可见文案变更。
- 验证：已通过`cd apps/lina-core && go test ./internal/service/plugin/... ./internal/service/i18n ./internal/cmd -count=1`、`cd apps/lina-core && go test ./internal/controller/plugin -count=1`、`openspec validate consolidate-plugin-service-boundaries --strict`、`git diff --check`和上述静态检索。

## Feedback

- [x] **FB-1**: 收敛`apps/lina-core/internal/cmd`根包内部实现细节，保留命令入口和命令相关主文件
- [x] **FB-2**: 将插件 host service DTO 适配从`httpstartup`收敛回插件宿主边界
- [x] **FB-3**: 继续减少`apps/lina-core/internal/service/plugin`根目录小文件数量

### FB-1 执行记录

- 根因：`apps/lina-core/internal/cmd`根包在插件服务边界收敛后继续承载 HTTP 启动装配、路由绑定、前端资源服务、OpenAPI 路由、插件 host service DTO 适配、SQL 资产扫描和 SQL 语句拆分等实现细节。根包职责超出命令入口和命令参数定义，增加理解和维护成本。
- 迁移结果：新增`apps/lina-core/internal/cmd/internal/httpstartup`承载 HTTP runtime、路由绑定、前端资产、OpenAPI 路由、启动 hook 和插件 host service 适配；新增`apps/lina-core/internal/cmd/internal/sqlassets`承载 init/mock SQL 资产来源、扫描、拆分和执行；新增`apps/lina-core/internal/cmd/internal/dbconfig`承载命令期数据库连接读取。`cmd_http.go`只委托`httpstartup.Run(ctx)`，`cmd_init.go`和`cmd_mock.go`只保留命令确认、参数解析和命令语义调用。
- 根目录收敛：`apps/lina-core/internal/cmd`根目录 Go 文件数从 17 个收敛为 8 个，保留`cmd.go`、`cmd_http.go`、`cmd_init.go`、`cmd_init_database.go`、`cmd_mock.go`和既有治理测试文件；HTTP 和 SQL 内部测试随实现迁入对应内部子包。
- DI 来源检查：未新增关键运行期服务实例；`httpstartup`继续由 HTTP 启动过程逐项创建并传递`auth`、`bizctx`、`config`、`datascope`、`i18n`、`plugin`、`sessionStore`、`orgcap`、`tenantcap`、`notify`、`kvcache`等共享实例，`plugin host service`适配只随包迁移到`httpstartup`，未引入聚合依赖结构体、全局 service locator 或隐式`New()`补齐关键依赖。
- 影响分析：无 HTTP API、路由语义、权限标签、DTO、OpenAPI 元数据、SQL 文件、数据库结构、前端 UI、插件 manifest 字段或运行时用户可见文案变更；`i18n`无语言资源、API 文档源文本或翻译缓存影响；数据权限无新增接口或数据操作入口，源码插件和动态插件通过宿主发布服务访问数据的既有边界保持；缓存一致性语义不变，`plugin-runtime`缓存和启动期共享实例路径保持；开发工具跨平台无脚本、`Makefile`、CI 或`linactl`变更。
- 静态检索：`rg -n "internal/service/pluginhostservices|internal/service/pluginruntimecache|plugin/internal/runtimecache" apps/lina-core --glob '*.go'`无结果；`find apps/lina-core/internal/cmd -maxdepth 1 -type f -name '*.go'`显示根目录仅剩 8 个 Go 文件；静态检索确认 HTTP 启动、路由、前端资产、OpenAPI、host service 适配和 SQL 资产处理函数均位于`cmd/internal/*`。
- 验证：已通过`cd apps/lina-core && go test ./internal/cmd ./internal/cmd/internal/... -count=1`、`cd apps/lina-core && go test ./internal/service/plugin/... ./internal/service/i18n ./internal/cmd ./internal/cmd/internal/... -count=1`、`cd apps/lina-core && go test ./internal/controller/plugin -count=1`、`openspec validate consolidate-plugin-service-boundaries --strict`和`git diff --check`。

### FB-2 执行记录

- 根因：FB-1 将 HTTP 启动实现迁入`apps/lina-core/internal/cmd/internal/httpstartup`后，插件 host service 的`apidoc`、`auth`、`i18n`和`notify`DTO 适配仍停留在`httpstartup/plugin_host_services_adapters.go`。这让启动装配包继续理解插件 capability 契约和宿主服务 DTO 之间的转换细节，边界上不符合`plugin`根包 facade +`plugin/internal/hostservices`具体适配的目标。
- 迁移结果：已删除`httpstartup/plugin_host_services_adapters.go`；`httpstartup`仅逐项创建并传递启动期共享服务实例。`plugin.NewHostServices`的窄入参改为宿主真实服务切片，`plugin/internal/hostservices`内部完成`apidoc.RouteTextInput/Output`、`auth`租户令牌 DTO、`i18n.MessageExportOutput`和`notify.NoticePublishInput/SendOutput`到`pkg/plugin/capability/contract`的转换。
- 测试覆盖：已在`apps/lina-core/internal/service/plugin/internal/hostservices/hostservices_adapters_test.go`补充`apidoc`批量 route text、`i18n`消息搜索和`notify`发布/删除的转换断言，并保留既有`auth`与`session`转换断言。
- DI 来源检查：未新增运行期服务实例；`apiDocSvc`、`authSvc`、`authTokenSvc`、`bizCtxSvc`、`hostConfigSvc`、`scopeSvc`、`i18nSvc`、`pluginSvc`、`sessionStore`、`orgCapSvc`、`tenantSvc`、`notifySvc`和`kvCacheSvc`仍由 HTTP 启动期创建或取得并逐项传入`plugin.NewHostServices`，适配器不使用聚合依赖结构体、全局 service locator 或隐式`New()`补齐关键依赖。
- 影响分析：无 HTTP API、路由语义、权限标签、OpenAPI 元数据、SQL、前端 UI、插件 manifest 字段或运行时用户可见文案变更；`i18n`无语言资源或翻译缓存语义变更，只迁移 runtime message 查询的适配归属；数据权限无新增接口或数据操作入口，源码插件和动态插件通过宿主发布服务访问数据的既有边界保持；缓存一致性语义不变，继续复用启动期共享`kvCacheSvc`和既有`plugin-runtime`缓存协调；开发工具跨平台无脚本、`Makefile`、CI 或`linactl`变更。
- 验证：已通过`cd apps/lina-core && go test ./internal/service/plugin/internal/hostservices ./internal/service/plugin ./internal/cmd/internal/httpstartup ./internal/cmd -count=1`；静态检索确认`httpstartup`不再包含`cmdAPIDocAdapter`、`cmdTenantTokenIssuerAdapter`、`cmdI18nAdapter`或`cmdNotifyAdapter`，且`internal/cmd`仍不直接导入`plugin/internal/hostservices`。

### FB-3 执行记录

- 根因：阶段 4 已把较大职责下沉到`plugin/internal/<subcomponent>`，但`apps/lina-core/internal/service/plugin`根目录仍保留多个少于 50 行的小型 Go 文件，例如`plugin_topology.go`、`plugin_source_manifest.go`、`plugin_source_route_binding.go`、`plugin_openapi.go`、`plugin_platform_guard.go`、`plugin_install_context.go`、`plugin_enabled_snapshot.go`、`plugin_frontend.go`以及小型测试占位文件。根包文件数仍偏高，且后端 Go 规则要求超过 10 个文件的组件应尽量合并少于 50 行的小文件。
- 合并结果：已将`Topology`和`ScanRegisteredSourceManifests`收敛进`plugin.go`；将`ListSourceRouteBindings`收敛进`plugin_integration.go`；将动态运行时 frontend facade 和 OpenAPI projection facade 收敛进`plugin_runtime.go`；将平台治理守卫收敛进`plugin_startup_consistency.go`；将 install mock-data context helper 和 enabled snapshot helper 收敛进`plugin_lifecycle.go`；将 data-table comment helper 测试收敛进`plugin_list_test.go`；将测试数据库驱动导入收敛进`plugin_test.go`。对应空壳文件已删除，`plugin`根目录 Go 文件数从 48 降为 38，其中生产文件从 30 降为 22，测试文件从 18 降为 16。
- 边界判断：本反馈只做同包源码组织收敛，不改变`Service`接口、导出符号、HTTP API、路由、DTO、SQL、前端 UI、插件 manifest、动态插件授权语义、缓存 key、缓存失效策略或运行时用户可见文案；未新增运行期依赖，未创建新的 service、adapter、registry、manager 或启动装配层。
- 影响分析：`i18n`无语言资源、API 文档源文本、错误消息或翻译缓存语义影响；数据权限无新增接口、数据操作或可见性边界影响；缓存一致性无新增缓存、快照、失效触发点或跨实例同步影响，既有`plugin-runtime`发布路径保持；开发工具跨平台无脚本、`Makefile`、CI 或`linactl`影响；测试策略为 Go 编译门禁、受影响子组件测试、静态符号检索、OpenSpec 严格校验和格式检查。
- 已通过验证：`gofmt`已覆盖本反馈修改的 Go 文件；`openspec validate consolidate-plugin-service-boundaries --strict`通过；`git diff --check -- apps/lina-core/internal/service/plugin openspec/changes/consolidate-plugin-service-boundaries/tasks.md`通过；静态检索确认被删除小文件名无残留引用；静态检索确认`Topology`、`ScanRegisteredSourceManifests`、`BuildDynamicRoutePublicPath`、`ListSourceRouteBindings`、`ensurePlatformGovernance`、`withInstallMockData`、`syncEnabledSnapshotAndPublishRuntimeChange`和`PrewarmRuntimeFrontendBundles`仍在根包内定义；已通过`cd apps/lina-core && go test ./internal/service/plugin/internal/catalog ./internal/service/plugin/internal/frontend ./internal/service/plugin/internal/integration ./internal/service/plugin/internal/runtime ./internal/service/plugin/internal/lifecycle ./internal/service/plugin/internal/openapi ./internal/service/plugin/runtimecache -count=1`和`cd apps/lina-core && go test ./internal/service/plugin/internal/dependency -count=1`。
- 后续根因确认：GitHub Actions run `26515440723`中 plugin-full Go 单测已越过此前`auth-client-type-session-metadata`的 DAO/Entity 阻断，新的失败点为`lina-core/internal/service/plugin/internal/runtime`包的`TestRunBundledDynamicSampleBeforeInstallLifecycleAllowsRuntimeLog`。该测试执行真实打包的`linapro-demo-dynamic` WASM 样例；CI plugin-full 命令带`-race`，WASM 冷启动和 guest 初始化超过样例默认 5 秒生命周期回调超时，触发`module closed with context deadline exceeded`。生产生命周期默认超时策略正确，不应为测试冷启动放宽生产契约。
- 后续修复：仅在该真实样例单测中把`BeforeInstall` handler 的测试用 manifest `TimeoutMs`提高到 2 分钟，保留生产默认`pluginhost.DefaultLifecycleHookTimeout=5s`和动态插件 manifest 自身契约不变。扩大到 2 分钟的原因是 CI `-race`下真实 WASM 样例冷启动耗时接近 40 秒，30 秒仍存在抖动风险。
- 测试 fixture 修复：本地扩大验证时发现`TestValidateStartupConsistencyRejectsPlatformUserMembership`清理租户成员关系前假设`linapro-tenant-core`插件表已存在，若测试数据库未安装插件 schema 会失败于`plugin_linapro_tenant_core_user_membership`缺表。已在该测试文件内新增自包含建表 helper，使用插件 schema 等价的幂等 DDL 和索引，仅服务测试隔离，不修改生产 SQL，不引入插件 DAO/DO/Entity 依赖。
- 后续影响分析：无 HTTP API、SQL、前端 UI、插件 manifest、运行时文案、语言包、数据权限或缓存一致性变更；只调整测试内存中的 manifest 副本、Go 测试 fixture 和任务记录。开发工具跨平台无脚本或`linactl`入口变更；测试 DDL 为当前测试显式调用的幂等 fixture，不属于交付迁移。
- 后续验证：已通过`cd apps/lina-core && go test ./internal/service/plugin -run TestValidateStartupConsistencyRejectsPlatformUserMembership -count=1`、`cd apps/lina-core && go test -race ./internal/service/plugin/internal/runtime -run TestRunBundledDynamicSampleBeforeInstallLifecycleAllowsRuntimeLog -count=1 -v`、`cd apps/lina-core && go test ./internal/service/plugin ./internal/service/plugin/internal/runtime ./internal/service/plugin/internal/hostservices -count=1`、`openspec validate consolidate-plugin-service-boundaries --strict`和`git diff --check`。

#### FB-3 Lina 审查记录

- 审查范围：反馈级，文件包括`apps/lina-core/internal/service/plugin/internal/runtime/lifecycle_precondition_sample_test.go`、`apps/lina-core/internal/service/plugin/plugin_startup_consistency_test.go`和本`tasks.md`记录；`git ls-files --others --exclude-standard`无未跟踪文件。
- 已读取规则文件：`AGENTS.md`、`.agents/rules/openspec.md`、`.agents/rules/documentation.md`、`.agents/rules/architecture.md`、`.agents/rules/backend-go.md`、`.agents/rules/database.md`、`.agents/rules/cache-consistency.md`、`.agents/rules/data-permission.md`、`.agents/rules/plugin.md`、`.agents/rules/testing.md`、`.agents/rules/i18n.md`；技能：`lina-feedback`、`lina-review`、`goframe-v2`。
- 规则域结论：Go 测试变更自包含且顺序无关；动态样例仅放宽测试内 manifest 副本，不改变生产生命周期默认超时或插件 manifest；测试 fixture DDL 仅为当前测试显式准备插件表，不是交付 SQL 迁移，不引入宿主生产路径对插件 DAO/DO/Entity 的依赖；无 HTTP API、前端 UI、语言资源、缓存策略、数据权限边界、插件 manifest 或开发工具入口影响。
- 验证证据：`cd apps/lina-core && go test ./internal/service/plugin -run TestValidateStartupConsistencyRejectsPlatformUserMembership -count=1`通过；`cd apps/lina-core && go test -race ./internal/service/plugin/internal/runtime -run TestRunBundledDynamicSampleBeforeInstallLifecycleAllowsRuntimeLog -count=1 -v`通过；`cd apps/lina-core && go test ./internal/service/plugin ./internal/service/plugin/internal/runtime ./internal/service/plugin/internal/hostservices -count=1`通过；`openspec validate consolidate-plugin-service-boundaries --strict`通过；`git diff --check`通过。严重问题 0，警告 0。
