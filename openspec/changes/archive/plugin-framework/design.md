# Design

## 插件契约和生命周期

插件平台以`plugin.yaml`为统一入口，源码插件使用`apps/lina-plugins/<plugin-id>/`目录，动态插件使用发布产物和 active release 快照。清单承载插件身份、类型、版本、菜单、权限、`hostServices`、依赖声明、运行期配置和 manifest 资源索引；未声明的能力不得由运行时自行推断。

生命周期从“源码目录扫描 + 动态产物装载”演进为显式治理模型。源码插件可被发现、安装、启用、禁用、升级和通过启动引导自动启用；动态插件覆盖上传、安装、授权确认、启用、禁用、卸载、升级、同版本刷新和 active release 切换。插件业务入口在`pending_upgrade`、`abnormal`或`upgrade_failed`状态下必须受控，插件管理和升级入口仍保持可访问。

生命周期回调统一使用`Before*`、`Upgrade`、`Uninstall`和`After*`命名，替代旧`Can*`或 guard 风格契约。`Before*`可阻断操作并返回稳定原因键，`After*`作为 best-effort 通知；卸载相关回调携带`purgeStorageData`，使插件区分保留数据和清理数据两种卸载策略。动态插件生命周期契约由构建期自动发现 guest controller 方法生成，运行时只信任产物内显式生命周期 custom section。

插件生命周期 SQL、mock SQL、upgrade SQL、rollback SQL 和`sys_plugin_migration`账本必须保持事务一致。失败诊断需要保留原始失败、rollback 失败、失败阶段、错误码、消息键和 manifest 快照，避免只记录 warning 后丢失恢复依据。

## 动态运行时和路由资产边界

动态插件运行时以`WASM`产物自定义段为权威输入。`pluginbridge`集中解析`WASM`自定义段，`i18n`、`apidoc`和插件运行时不得维护重复解析器。运行时资源视图绑定 active release checksum 或 generation，包含 manifest、默认配置、前端资产、路由契约、生命周期契约、`hostServices`授权快照和语言资源。

路由边界经历了从`/api/v1/extensions/{pluginId}/...`到统一插件`API`命名空间的收敛。最终设计中源码插件和动态插件`API`均使用`/x/{plugin-id}/api/v1/...`，`/x`只表示插件`API`命名空间；源码插件公开页面、门户、自管静态资源和 fallback 使用非保留路径。公开资产由`plugin.yaml public_assets`声明，并通过`/x-assets/{plugin-id}/{version}/...`托管，不再保留`/plugin-assets`兼容入口。

管理工作台入口从默认根路径收敛为可配置入口，默认`/admin`。工作台`SPA` fallback 只覆盖工作台入口及其子路径，不吞掉源码插件公开路由；`/api`、`/x`、`/x-assets`和已注册插件路由优先级高于工作台 fallback。

所有动态`WASM`执行入口共享宿主资源边界。路由、cron、生命周期和发现执行在调用方已有 deadline 时不得放宽；没有 deadline 时使用宿主默认超时；内存分配超过上限时拒绝或终止本次执行。

## 宿主服务和能力目录

动态插件不直接获得宿主连接、`DAO`、`gdb.Model`、请求对象或内部 service。所有宿主能力通过版本化 host service envelope 进入`pkg/plugin/capability`能力目录或受控适配器。`hostServices`是动态插件的授权申请和运行时授权快照，不是 Go 公共能力集合。

源码插件和动态插件共享同一能力语义，只允许 transport 不同。源码插件通过 registrar、hook、cron 或 lifecycle 上下文接收`capability.Services`；动态插件通过`capability/guest`和`pluginbridge/protocol`发起 host service 调用，宿主先校验授权快照，再委托到同一能力实现。

配置能力被拆为插件作用域配置和宿主公开配置。`HostServices.Config()`读取当前插件运行期配置，实际配置可来自源码插件目录、生产配置根或动态产物默认配置；`HostServices.HostConfig()`只能读取宿主显式白名单键。manifest 资源能力只读当前插件`manifest/`声明型资源，拒绝空根、绝对路径、路径穿越、`URL`和跨插件读取；`manifest/sql`、`manifest/i18n`和`manifest/config`仍由专用生命周期、`i18n`和配置管线处理。

数据能力通过结构化 data service 和`capability/data`受限 DSL 承载。动态插件按表、方法、字段、分页、排序、用户上下文和数据权限执行治理，不暴露 raw SQL。列表和投影接口必须批量化、有界装配，避免插件或宿主通过循环单条详情制造`N+1`。

缓存、锁、网络、通知和存储服务均以插件作用域授权为边界。缓存是有损缓存，不作为权限、配置、插件状态、业务数据或修订号事实源；集群模式使用 coordination 后端，单机模式可用 SQL table 后端。锁使用 ticket/owner token 隔离；网络按授权`URL`模式 default-deny；通知按授权 channel 发送；存储只暴露逻辑空间和对象，不泄漏物理路径。

## 包边界和内部实现

插件公共契约统一收敛到`pkg/plugin`命名空间。`pluginhost`只负责源码插件贡献入口，`pluginbridge`只负责动态插件`ABI`、transport 和公开协议出口，`capability`只负责插件消费宿主能力。历史`pluginservice`、`plugindb`和`sourceupgrade`公共入口被移除或内化，能力 client、data DSL 和 guest SDK 使用`pkg/plugin/capability/**`。

宿主插件运行时治理收敛到`apps/lina-core/internal/service/plugin`及其职责明确的子组件。插件 catalog、runtime、host service adapter、runtime cache、lifecycle、integration、frontend、OpenAPI、WASM host service、管理投影和升级治理不再作为`internal/service`根层级的平行组件扩散。启动层只依赖`plugin`根 facade 或明确允许的受控子包，不直接导入`plugin/internal/<subcomponent>`。

框架能力 provider 使用窄接口和强类型 provider env。组织、租户等能力由独立 capability 组件维护 DTO、消费 service、fallback、delegation 和 provider factory；provider adapter 留在提供方插件内部，消费方通过能力 service 或插件依赖治理接缝访问，不直接 import 其他插件`backend/internal/**`。

## 插件 UI、菜单和管理读模型

插件页面支持`iframe`、新标签页和宿主嵌入式挂载。管理工作台只通过 manifest 菜单声明感知插件贡献的导航、权限和动态页面；源码插件通过代码注册的公开 HTTP 路由不自动投影为菜单、权限节点或 OpenAPI 路径。

插件菜单复用宿主菜单与角色授权体系，以`menu_key`作为生命周期锚点。插件禁用时菜单和按钮权限临时失效但授权关系保留，重新启用后恢复；动态路由生成的按钮权限必须挂在所属插件菜单下，而不是形成游离权限集合。

插件管理列表从副作用查询演进为只读读模型。`GET /plugins`不得写治理表；同步、上传、安装、卸载、启用、禁用、升级和租户供应策略变更显式失效读模型。列表读模型可在启动后预热，但必须保留详情、安装、卸载、升级和授权弹窗所需字段，不能通过删字段换取首屏速度。

## 启动引导、依赖和升级

`plugin.autoEnable`从字符串列表演进为结构化条目，包含`id`和可选`withMockData`。启动引导在插件路由、cron 和动态前端包预热前执行，语义是按需安装再启用；列出的插件失败必须阻塞启动。引导阶段和后续接线复用同一启动快照，源码插件安装写入后必须同步更新快照，避免同一轮启用检查读取旧的未安装状态。

依赖治理最终收敛为硬插件依赖和框架版本约束。安装、启用、卸载、升级和发布切换前校验依赖存在性、版本范围、循环和反向依赖；插件清单不再维护软依赖或自动安装策略字段。若必须硬依赖 provider 插件，仍通过`dependencies.plugins`声明；可选 capability 通过`Available(ctx)`降级，不新增 capability 依赖模型。

运行时升级显式分离发现版本和有效版本。启动扫描只标记`normal`、`pending_upgrade`、`abnormal`或失败状态，不自动升级、不切换 release、不执行 SQL。升级通过只读预览和有副作用执行`API`完成，执行链路按锁、状态重读、依赖校验、`BeforeUpgrade`、`Upgrade`、upgrade SQL、治理同步、release 切换、缓存失效和`AfterUpgrade`顺序推进。

## 工作区、构建和验证边界

官方插件工作区可选。host-only 模式下，宿主构建、宿主测试、前端扫描和源码插件发现不得要求`apps/lina-plugins`存在；plugin-full 模式下才要求官方插件工作区、官方插件测试、动态插件`WASM`构建和插件自有`E2E`。插件工作区管理命令负责从`hack/config.yaml`来源安装、更新、诊断和锁定插件目录，不把来源仓库作为嵌套 Git 仓库写入。

动态插件构建器支持`go:embed`资源声明，并在迁移期保留目录扫描回退。最终`.wasm`和中间产物收敛到仓库根`temp/output/`或显式输出目录，不回写插件源码目录。发布产物必须保留与源码插件一致的 manifest、配置、metadata、SQL、i18n 和 public asset 路径语义。

验证策略按 host-only 与 plugin-full 区分。根`E2E`只验证宿主插件框架和通用动态插件 fixture；官方源码插件业务能力由各插件自己的`hack/tests/e2e/`闭环。完整发布链路按共享测试模板执行简要测试门禁，nightly 覆盖完整`E2E`。

## 运行时一致性和故障恢复

插件运行时变化必须按插件、sector、locale 或 global scope 精细失效 frontend bundle、runtime i18n、WASM 编译缓存、manifest 资源视图和默认配置视图。集群模式通过统一 coordination revision/event 唤醒其他节点和 reconciler；freshness 不可确认且超过窗口时采用 conservative-hide，不暴露可能已禁用、卸载或权限变化的插件能力。

动态插件协调器在集群模式下按插件 ID 串行化共享副作用，生命周期 SQL、迁移账本、菜单权限同步、active release 切换、frontend bundle 切换和 runtime revision 发布只能由持锁节点执行。协调器需要恢复 stale `reconciling`状态，并在 tick 边界隔离 panic，避免单个插件故障停止后台收敛。

## Cross-Domain Impacts

- `cluster-deployment-mode`、`cluster-topology-boundaries`、`leader-election`和`distributed-locker`为插件运行时提供主节点、副本、revision/event 和 per-plugin 锁边界；当前契约由`openspec/specs`对应能力承载，历史 owner 为`archive/distributed-infra`。
- `cron-jobs`影响插件 cron 声明、主节点任务分类和动态任务接线；当前契约由`openspec/specs/cron-jobs/spec.md`与调度相关规范承载，历史 owner 为`archive/scheduled-jobs`。
- `menu-management`、`role-management`和`plugin-permission-governance`共同定义插件菜单、按钮权限和角色授权关系；插件侧历史保留`plugin-permission-governance`，菜单与角色通用契约由`archive/user-management`及主规范承载。
- `user-auth`只作为插件登录/登出 hook、鉴权上下文和失败隔离的交叉影响；当前认证契约由`openspec/specs/user-auth/spec.md`承载，历史 owner 为`archive/user-auth`。
- `online-user`和`server-monitor`只保留时长配置、会话触碰节流、监控采集和清理任务与插件运行时的交叉影响；当前契约由系统监控或用户会话相关主规范承载。
- `project-setup`、`release-image-build`和`e2e-suite-organization`只保留 host-only/plugin-full、工作台入口、插件 API 路径、发布和测试范围的交叉影响；长期 owner 分别为`archive/foundation`、`archive/devops-tooling`和`archive/e2e-testing`。
- `core-host-boundary-governance`、`module-decoupling`和`service-dependency-injection-governance`提供宿主边界、可选插件降级和显式依赖注入约束；当前契约由主规范和对应治理分组承载，插件分组只保留对插件公共包、host services 和 provider 边界的影响。
- `source-upgrade-governance`和`system-api-docs`分别影响源码插件升级内部化和动态插件 OpenAPI 投影；当前契约由主规范承载，插件分组通过运行时升级和动态路由设计保留必要历史原因。
- `config-duration-unification`和`demo-control-guard`分别影响插件配置读取、启动引导和演示只读插件安装方式；当前契约由系统配置和官方插件治理相关主规范承载。
