# 插件目录结构规则

## 适用范围

本规则约束 LinaPro 插件的通用资源目录、源码插件与动态插件共享的后端开发目录结构、源码插件编译嵌入对接、动态插件 WASM 运行时对接、动态插件产物资源视图、插件 manifest 资源、安装卸载 SQL、前端页面、宿主能力接缝和生命周期资源归属。

源码插件和动态插件必须共享插件清单、生命周期资源、SQL、i18n、前端静态资源和后端业务开发结构约定。两类插件的差异仅体现在与宿主的对接方式、运行时加载方式和交付形态上：源码插件随宿主编译嵌入，动态插件通过 WASM artifact、`pluginbridge`和`hostServices`协议接入。禁止让动态插件绕过统一的`api`、`controller`、`service`分层开发结构，也禁止让动态插件绕过通用插件资源约定。

## 强制遵守场景

以下任一场景命中时，必须先读取并遵守本规则：

- 新增或修改`apps/lina-plugins/<plugin-id>/`下的源码插件或动态插件源码目录
- 修改插件通用资源，包括`plugin.yaml`、`frontend/`、`manifest/`、`manifest/sql/`、`manifest/i18n/`或插件生命周期资源
- 修改源码插件与动态插件共享的后端开发结构，包括`backend/api/`、`backend/plugin.go`、`backend/internal/controller/`或`backend/internal/service/`
- 修改插件数据库访问结构，包括`backend/hack/config.yaml`、`backend/internal/dao/`或`backend/internal/model/{do,entity}/`
- 修改源码插件编译嵌入对接结构，包括`plugin_embed.go`、registrar、provider 或 adapter
- 修改动态插件运行时对接结构，包括`main.go`、`go.mod`、WASM 构建入口、`pluginbridge`路由声明、`hostServices`声明或动态插件产物资源视图
- 修改宿主与插件之间的资源归属、发布产物、打包资源或生命周期扫描逻辑

## 插件通用资源要求

源码插件和动态插件都必须遵守以下通用资源约定：

- 插件源码目录统一放在`apps/lina-plugins/<plugin-id>/`下，`<plugin-id>`必须与`plugin.yaml`中的`id`一致。
- 插件必须维护`plugin.yaml`和`manifest/`。
- 插件前端页面或公开静态资源统一放在`frontend/pages/`或由`plugin.yaml`的`public_assets`显式声明的目录下。
- 插件安装 SQL 放在`manifest/sql/`。
- 插件卸载 SQL 放在`manifest/sql/uninstall/`。
- 插件 Mock 数据 SQL 放在`manifest/sql/mock-data/`。
- 插件多语言资源放在`manifest/i18n/<locale>/`，API 文档翻译资源放在`manifest/i18n/<locale>/apidoc/`。
- 不得把插件生命周期资源回流到宿主目录中。
- 插件 SQL 必须遵守`.agents/rules/database.md`。
- 插件 i18n 资源必须遵守`.agents/rules/i18n.md`。

## 插件后端同构开发结构要求

源码插件和动态插件必须保持一致的后端业务开发结构，以降低开发者学习、迁移和维护成本。两类插件必须遵守以下结构：

- 每个插件必须同时维护`plugin.yaml`、`backend/`、`frontend/`与`manifest/`。
- 插件后端统一采用`backend/api/`、`backend/plugin.go`、`backend/internal/controller/`、`backend/internal/service/`结构。
- `backend/api/`用于声明构建期可解析的 API DTO、请求响应契约和路由元数据。
- `backend/plugin.go`用于声明插件后端入口、路由注册、生命周期接入或动态路由桥接入口。
- `backend/internal/controller/`用于承载插件侧请求处理、参数转换、调用服务和响应投影逻辑。动态插件中的 controller 是 WASM guest 内部的请求处理分层，不等同于宿主原生 controller，但目录和职责必须保持一致。
- `backend/internal/service/`用于承载插件业务编排、领域逻辑、中间件实现和对宿主能力接缝的调用。动态插件中的 service 通过`pluginbridge`、WASM host call 或版本化 host service 协议访问宿主能力。
- 禁止再将业务`service`目录直接放在`backend/service/`下。
- 插件业务编排、领域逻辑和中间件实现必须收敛到`backend/internal/service/`。
- 只有实现宿主稳定能力接缝的 provider/adapter 才允许放在`backend/provider/`等非`internal`目录中。
- 插件后端 Go 代码必须遵守`.agents/rules/backend-go.md`。

## 插件数据库访问要求

- 插件若需要自有数据库访问，必须在插件自己的`backend/`下维护`hack/config.yaml`。
- 插件的`gf gen dao`生成结果必须放在`backend/internal/dao/`与`backend/internal/model/{do,entity}/`。
- 禁止插件重新依赖宿主的`dao/do/entity`生成工件。
- 动态插件涉及宿主数据访问时，必须通过`plugin.yaml`的`hostServices`资源边界和宿主授权的 host service 协议，不得直接依赖宿主私有 DAO、DO 或 Entity 工件。

## 源码插件对接要求

源码插件是随宿主源码编译和嵌入交付的插件。源码插件必须遵守以下对接要求：

- 源码插件必须维护`plugin_embed.go`作为宿主编译嵌入和静态资源装配入口。
- 源码插件应通过 registrar 或等价上下文把`backend/plugin.go`中声明的 controller、service、路由、中间件和生命周期能力接入宿主。
- 源码插件控制器和服务应通过宿主稳定能力接缝获取`pkg/pluginservice/*`适配器，不得直接耦合宿主私有实现。
- 源码插件 provider/adapter 只能承载宿主稳定能力接缝实现，业务编排和领域逻辑仍必须放在`backend/internal/service/`。

## 动态插件目录结构要求

动态插件是以运行时 WASM artifact 交付和加载的插件。动态插件必须保持与源码插件一致的后端业务开发结构，并额外遵守以下运行时对接结构：

- 动态插件必须维护`plugin.yaml`，并在其中声明`type: dynamic`。
- 动态插件源码目录应维护`go.mod`和`main.go`作为 WASM guest 构建入口。
- 动态插件必须维护`backend/api/`、`backend/plugin.go`、`backend/internal/controller/`和`backend/internal/service/`，并与源码插件保持相同的职责边界。
- 动态插件的`backend/plugin.go`或同职责文件必须作为构建期路由声明和 WASM guest 接入入口，例如通过`pluginbridge.DynamicRouteRegistrar`声明路由分组，并把动态请求分发到插件侧 controller。
- 动态插件可以维护 host call 或 host service 分发相关文件，例如`backend/plugin_host_dispatcher.go`，但这些文件只能承载 guest/bridge 适配逻辑，不得替代`backend/internal/controller/`和`backend/internal/service/`承载业务逻辑。
- 动态插件的 controller/service 是 guest 内部开发分层，宿主不得把它们当作源码插件原生 controller/service 直接加载；宿主只能通过`pluginbridge`、WASM host call 或版本化 host service 协议与动态插件交互。
- 动态插件涉及 Go guest 代码、WASM host service、host call 协议或插件桥接时，必须遵守`.agents/rules/backend-go.md`中关于动态插件 host service、WASM host service、错误处理和共享实例的要求。

## 动态插件 manifest 与授权要求

- 动态插件的`plugin.yaml`必须通过`hostServices`声明所需宿主服务、方法和资源边界。
- 动态插件访问 storage、network、data、cache、lock、notify、config、runtime 等宿主能力时，必须通过受治理的 host service 协议，不得绕过授权模型直接访问宿主资源。
- 动态插件声明的 data 资源边界必须遵守`.agents/rules/data-permission.md`。
- 动态插件声明或修改缓存相关 host service 时，必须遵守`.agents/rules/cache-consistency.md`。
- 动态插件菜单、路由、按钮权限和用户可见文本必须按`plugin.yaml`、`manifest/i18n`和相关规则统一治理。

## 动态插件产物资源视图要求

动态插件打包后的 WASM artifact 或发布产物必须保留与源码目录一致的插件资源语义：

- 发布产物必须携带`plugin.yaml`或等价 manifest 快照。
- 发布产物必须携带`manifest/sql/`、`manifest/sql/uninstall/`、`manifest/sql/mock-data/`中适用的资源，并保持相同路径语义。
- 发布产物必须携带启用 i18n 时的`manifest/i18n/<locale>/`资源和`apidoc`资源，并由宿主运行时按动态插件扇区加载。
- 动态插件打包工具不得为 SQL、Mock 数据、i18n 或前端静态资源引入不同于源码插件的额外路径约定或额外清单字段，除非先通过 OpenSpec 变更明确设计。
- 动态插件同版本刷新、升级、启用、禁用或卸载后，必须按插件运行时和缓存规则触发对应派生资源失效或刷新。

## 宿主接缝要求

- 源码插件应通过宿主稳定能力接缝访问宿主能力，不得直接耦合宿主私有实现。
- 源码插件控制器和服务应通过 registrar 或等价上下文获取宿主发布的`pkg/pluginservice/*`适配器。
- 动态插件必须通过`pluginbridge`、WASM host call 或版本化 host service 协议访问宿主能力。
- 动态插件的`pluginbridge`、host call dispatcher、WASM export/import 等对接层必须只做协议适配、路由分发和宿主能力调用，不得绕过插件内部 controller/service 分层直接堆叠业务逻辑。
- 生产路径不得自行构造孤立宿主服务适配器或绕过启动期共享服务实例。
- 涉及缓存、权限、数据权限、`i18n`或运行时配置时，必须继续读取对应规则文件。

## 验证要求

- 插件目录结构变更必须通过静态文件存在性检查确认通用资源、共享后端开发结构、源码插件对接结构或动态插件运行时对接结构完整。
- 源码插件后端变更必须运行覆盖变更包的 Go 编译门禁。
- 动态插件 guest、controller、service、bridge 或 WASM 入口变更必须运行对应 WASM 构建、`linactl wasm`、插件构建 smoke、Go 编译门禁或更窄但能覆盖构建期契约的验证。
- 插件 API、SQL、前端、i18n 和 E2E 变更必须分别执行对应规则文件要求的验证。
- 纯治理文档变更可以使用`openspec validate`、静态检索、文件存在性检查、格式检查或审查结论作为验证证据。

## 审查要求

- 审查必须先区分变更对象是插件通用资源、共享后端开发结构、源码插件对接结构、动态插件运行时对接结构还是动态插件发布产物资源视图。
- 审查必须确认源码插件和动态插件均符合统一的`backend/api/`、`backend/plugin.go`、`backend/internal/controller/`和`backend/internal/service/`结构，且插件业务逻辑没有放到非`internal`目录或宿主目录。
- 审查必须确认源码插件和动态插件数据库生成工件没有依赖宿主 DAO/DO/Entity。
- 审查必须拒绝动态插件用`main.go`、`pluginbridge`、host call dispatcher 或其他桥接文件替代 controller/service 分层承载业务逻辑。
- 审查必须确认动态插件通过`type: dynamic`、WASM 构建入口、`pluginbridge`路由声明和`hostServices`授权模型表达运行时能力边界。
- 审查必须确认源码插件和动态插件共享的 SQL、Mock 数据、i18n 和前端资源路径语义一致。
- 审查必须确认插件生命周期资源没有回流到宿主目录。
