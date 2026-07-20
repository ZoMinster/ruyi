## 1. 宿主通用认证契约

- [x] 1.1 在实施前重新读取所有命中规则并记录影响矩阵，明确`apps/lina-core`修改授权、`i18n`、缓存一致性、数据权限例外、测试策略、开发工具影响判断，以及新增运行期依赖的 owner、创建位置、传递路径和共享实例策略。
- [x] 1.2 在`apps/lina-core/pkg/plugin/capability/authcap`及职责明确的`SPI`边界新增`user/machine`主体、认证请求投影、认证结果、接口资源授权请求和认证提供者契约，保持普通消费面不暴露`ghttp`或宿主内部模型。
- [x] 1.3 在`pluginhost`声明和宿主插件集成层增加唯一认证方案注册、插件启用状态治理和原子提供者快照，并补充重复方案、缺失依赖、禁用和升级切换测试。
- [x] 1.4 扩展宿主`bizctx`、插件可见`bizctxcap`和动态插件上下文协议以传播机器客户端、凭证和租户，验证机器主体不能伪造`UserID`、用户角色或在线会话。
- [x] 1.5 增加路由`operation`、`resource`、`action`和`actors`元数据解析、全局目录、唯一性审计及静态、源码插件、动态插件注册校验，保证未声明接口默认拒绝机器主体。
- [x] 1.6 重构宿主认证和权限中间件为按方案精确分派的用户、机器双链路，实现`401/403`边界、接口与资源双重允许校验，并保持现有`Bearer JWT`回归行为。
- [x] 1.7 为通用主体、认证分派、授权目录、中间件早退、机器默认拒绝和用户回归新增宿主单元及启动绑定测试。
- [x] 1.8 审查并同步更新`apps/lina-core/pkg/plugin`下受公共认证契约影响的`README.md`和`README.zh-CN.md`，保证中英文事实一致。

## 2. AKSK插件结构与数据库

- [x] 2.1 创建`apps/lina-plugins/linapro-auth-aksk`官方源码插件前检查插件根`AGENTS.md`，按`source`、`managed`、`tenant_aware`和双语`i18n`配置完整插件目录、嵌入入口、构建配置、菜单及JWT管理权限。
- [x] 2.2 设计机器客户端、访问密钥、策略、策略接口、策略资源和客户端策略关系的PostgreSQL幂等安装与卸载SQL，使用完整`plugin_linapro_auth_aksk_`表前缀、软删除、自动时间字段、业务唯一约束和真实查询路径索引。
- [x] 2.3 通过插件生命周期安装 SQL 初始化本地开发库并使用插件`make dao`生成`DAO/DO/Entity`，核对主键宽度、时间字段、软删除行为、`hack/config.yaml`表范围和`removePrefix`，不得手工修改生成文件。
- [x] 2.4 增加版本化主密钥环、签名时间窗口、`nonce TTL`、缓存陈旧窗口、策略规模和密钥数量上限配置，使用部署环境注入密钥并在缺少活动或历史密钥版本时失败关闭。
- [x] 2.5 增加SQL幂等、唯一约束、索引覆盖、软删除、卸载和重复初始化测试或静态治理验证，并记录Seed、Mock和敏感配置分类结论。

## 3. 客户端、密钥与策略管理后端

- [x] 3.1 按REST语义拆分机器客户端、访问密钥和策略API DTO，补齐`g.Meta`、JWT管理权限、`dc`、`eg`、Unix毫秒时间戳和批量数量上限，再通过`make ctrl`生成控制器骨架。
- [x] 3.2 实现机器客户端分页、详情、创建、更新、状态变更和软删除服务，在数据库查询与写操作前强制平台或当前租户边界，并记录共享治理资源不应用本人或部门范围的数据权限例外。
- [x] 3.3 实现每客户端多密钥创建、列表、状态变更、过期和软删除，使用密码学安全随机源生成唯一`AK/SK`，只在创建响应返回一次明文`SK`，禁止直接修改已有`SK`。
- [x] 3.4 实现`AES-256-GCM`密钥环加密、关联数据绑定、版本化解密和密钥轮换读取，保证明文`SK`不进入实体投影、缓存、错误或日志。
- [x] 3.5 实现策略分页、详情、创建、更新、状态变更、软删除，以及接口授权、资源`read/write`授权和客户端策略关系的事务内集合化替换，拒绝未知接口、资源条件、显式`deny`和超限输入。
- [x] 3.6 实现机器接口和资源目录的有界批量投影，供策略页面一次读取；插件禁用接口暂不生效但保留策略关系，重新启用后按最新目录恢复。
- [x] 3.7 实现客户端和策略列表的聚合或批量关联装配，加入查询次数断言或等价审查证据，证明分页查询不会按行产生密钥、策略或接口`N+1`查询。
- [x] 3.8 为管理服务和控制器补充JWT权限拒绝、机器主体拒绝、同租户共享可见、跨租户不透明拒绝、唯一冲突、批量原子性、软删除和一次性`SK`响应测试。

## 4. HMAC认证提供者

- [x] 4.1 实现`LINA-HMAC-SHA256`头部解析和固定规范签名串，覆盖方法、转义路径、按`RFC 3986`排序编码的查询参数、请求体摘要、Unix秒时间戳和随机`nonce`。
- [x] 4.2 使用固定协议测试向量验证客户端与服务端签名一致性，并覆盖空查询、重复查询键、Unicode、转义路径、空请求体、JSON请求体和签名头格式错误。
- [x] 4.3 实现按唯一`AK`读取密文凭证、客户端状态、过期时间和聚合策略快照，复算请求体摘要、解密当前`SK`并使用恒定时间签名比较。
- [x] 4.4 实现签名过期、请求体篡改、密钥停用、客户端停用、密钥删除、密钥过期、无策略、接口拒绝和资源拒绝的稳定`bizerr`及`401/403`映射。
- [x] 4.5 将AKSK认证器通过宿主通用提供者接缝注册，验证插件安装、启用、禁用、重新启用、卸载和初始化失败时的认证方案可用性，且JWT链路不受影响。
- [x] 4.6 实现`last_used_at`固定窗口合并写入和失败重试或告警，验证高频认证不会每请求写数据库且更新时间失败不改变已授权请求结果。
- [x] 4.7 为HMAC提供者增加并发、无效签名、密钥轮换、明文生命周期和敏感错误脱敏测试，并运行Go竞态或等价并发验证覆盖共享状态。

## 5. 缓存一致性与重放防护

- [x] 5.1 在现有缓存协调体系注册租户维度的`machine-access`修订号域，明确数据库权威源、缓存键、新鲜度窗口、失效原因、指标和恢复路径。
- [x] 5.2 实现只缓存密文凭证和非敏感策略集合的版本化快照，保证密钥、客户端、策略及关系写入在事务提交后跨节点失效，失效发布失败不得静默成功。
- [x] 5.3 实现单机有界本地和集群共享`AK + nonce`原子去重，`TTL`覆盖完整签名窗口，集群协调后端不可用时机器认证失败关闭。
- [x] 5.4 增加单机快照复用、跨节点密钥撤销、跨节点策略变更、重复`nonce`竞争、修订号不可读、快照过期和协调后端恢复测试。

## 6. OpenAPI、审计与安全治理

- [x] 6.1 在宿主OpenAPI构建器增加`LinaHMAC`安全方案，并按`actors`生成`BearerAuth`、`LinaHMAC`或二者择一的操作安全声明，投影稳定`operation/resource/action`扩展字段。
- [x] 6.2 为静态接口、源码插件和动态插件增加机器元数据与OpenAPI一致性测试，验证重复操作码、缺失资源、未知动作和未开放机器接口均失败关闭。
- [x] 6.3 扩展`linapro-monitor-operlog`审计模型、SQL、API和页面投影以区分用户与机器主体，并批量装配机器客户端及脱敏凭证标识而不伪造用户名。
- [x] 6.4 重构操作日志敏感数据处理，使请求头、请求参数、响应结果、动态响应回退和错误文本在截断前统一清理`Authorization`、`SK`、签名、主密钥、密文和加密`nonce`。
- [x] 6.5 增加一次性`SK`完整响应、超长截断响应、动态响应、认证错误和高频无效签名的脱敏及有界审计测试，证明任何持久化日志都不含凭证片段。
- [x] 6.6 增加机器认证失败安全事件的采样或聚合边界、稳定错误码、可观测指标和日志上下文，验证高频攻击不造成无界日志写放大。

## 7. 管理页面、i18n与文档

- [x] 7.1 使用现有`Vben`表格、表单、弹窗、确认和权限组件实现机器客户端列表、筛选、新增、编辑、启停和删除页面，避免创建只服务单页的多层公共包装。
- [x] 7.2 实现客户端密钥列表、创建密钥和一次性`AK/SK`结果弹窗，提供复制操作并在关闭后清除前端明文状态，列表和详情不得重新获取`SK`。
- [x] 7.3 实现策略列表与授权工作流，以接口树精确选择`operation`并按资源类型设置`read/write`，禁止提交资源ID、路径模式、条件或显式`deny`。
- [x] 7.4 按插件菜单和按钮权限完全隐藏无权操作，验证插件禁用时菜单、路由、页面和相关按钮全部隐藏而非仅置灰。
- [x] 7.5 维护插件`en-US`、`zh-CN`运行时文案、错误文案、菜单、字典和`apidoc`资源，运行静态键覆盖与接口文档翻译完整性校验。
- [x] 7.6 创建内容一致的插件`README.md`和`README.zh-CN.md`，记录配置密钥注入、签名规范、固定测试向量、密钥轮换、策略模型、启停影响和调用示例。

## 8. 自动化测试与最终门禁

- [x] 8.1 使用`lina-e2e`技能按插件模块本地编号分配并创建客户端CRUD、密钥一次性展示与轮换、策略授权、权限按钮隐藏、插件启停和异常路径E2E，测试资产保留在插件目录。
- [x] 8.2 增加端到端签名调用测试，使用代表性的机器专用测试路由验证有效签名、接口拒绝、资源读写拒绝、过期、篡改、重放、撤销和跨租户拒绝，同时证明JWT用户路径无回归。
- [x] 8.3 独立运行新增E2E并在页面首次加载、表单、一次性密钥结果、策略授权、提交结果和权限拒绝节点截图，使用多模态审查翻译、布局、遮挡、反馈和敏感信息泄漏。
- [x] 8.4 运行插件数据库初始化两次、卸载和重新安装，验证SQL幂等、DAO生成、软删除、唯一约束、索引、敏感字段和数据恢复语义。
- [x] 8.5 运行覆盖`authcap`、`bizctx`、`middleware`、`plugin`集成、`apidoc`、`linapro-auth-aksk`和`linapro-monitor-operlog`的Go测试及启动绑定测试，确认全部通过。
- [x] 8.6 对宿主和所有受影响插件运行`make lint`或`linactl lint.go`，并运行前端类型检查、构建、`i18n.check`、`apidoc`治理和敏感信息静态扫描。
- [x] 8.7 运行`openspec validate add-aksk-machine-authentication --strict`，更新任务记录中的DI来源、缓存一致性、数据权限例外、接口性能、SQL幂等、跨平台影响、`i18n`和测试证据。
- [x] 8.8 在全部实现和验证完成后调用`lina-review`执行代码与规范审查，修复所有严重问题后再标记变更实施完成。

## 实施影响与验证记录

### 任务 1.1 影响矩阵

| 规则域 | 影响与实施边界 |
|---|---|
| 核心宿主 | 用户已明确授权修改`apps/lina-core`；宿主只新增通用主体、认证提供者、路由元数据、上下文、授权分派和文档投影契约，不持有`AK/SK`业务表、密钥管理或管理页面语义。 |
| `i18n` | 有影响。宿主新增英文接口文档源文本并同步宿主非英文`apidoc`资源；`linapro-auth-aksk`显式启用`en-US`和`zh-CN`，插件运行时、错误、菜单和`apidoc`资源独立维护。 |
| 缓存一致性 | 有影响。插件数据库是机器访问权威源；启动期共享缓存按租户和`AK`保存密文凭证及非敏感权限快照，写事务提交后推进共享`machine-access`修订号；集群无法确认新鲜度或原子登记`nonce`时失败关闭。 |
| 数据权限 | 有明确例外。机器治理记录是平台或当前租户共享资源，由`JWT`管理权限和可信租户边界控制，不应用本人或部门范围；跨租户查询和写入在数据库访问前不透明拒绝。机器访问业务资源仍由资源 owner 应用租户和业务边界。 |
| 接口性能 | 管理列表先分页再批量聚合密钥和策略摘要；目录接口有数量上限；认证冷路径使用有界集合查询装配一次快照，热路径使用集合查找，禁止按客户端、策略、操作或资源逐项查询。 |
| 测试策略 | 宿主覆盖主体、注册、分派、上下文、目录、OpenAPI和`JWT`回归；插件覆盖SQL、管理服务、加密、签名向量、缓存、重放、生命周期和脱敏；页面与端到端签名路径使用插件本地`E2E`并执行关键节点截图审查。 |
| 开发工具与跨平台 | 有影响。动态插件同样需要机器路由元数据，因此扩展`hack/tools/linactl/internal/wasmbuilder`既有 Go AST 提取逻辑，把四类标签写入`RouteContract`；不修改`Makefile`、`make.cmd`或脚本，不引入平台命令和路径差异，并通过工具包测试与模块 lint 验证。 |
| 文档 | 有影响。同步维护宿主插件公共契约和新插件的`README.md`、`README.zh-CN.md`，两种语言表达同一事实。 |

### 运行期依赖来源

- 宿主认证提供者目录由`lina-core`插件启动装配拥有并创建，源码插件通过受治理 registrar 注册；中间件和路由目录复用同一原子快照，不在请求路径临时构造。
- `linapro-auth-aksk`服务由插件启动入口拥有并创建；数据库、配置、宿主集群协调、修订号和日志依赖从宿主 registrar 逐项注入，控制器、认证提供者和缓存组件复用该启动期服务图。
- 单机缓存和`nonce`仓库由插件启动期创建并共享；集群模式复用宿主统一协调后端，禁止创建节点本地替代实现后继续放行。
- 主密钥环属于插件运行配置值，由部署环境注入并在插件启动期解析；只作为纯值配置传递，不作为可自行创建的运行期服务。

### 任务 1.2 验证与审查

- 新增`authcap.Actor`、只读认证请求投影、认证结果、允许集合快照以及`operation + resource + access`双重授权请求；机器主体契约不包含`UserID`、角色或在线会话字段。
- 新增独立`authcap/authspi.Provider`与`ProviderFactory`，普通`authcap.Service`消费面未暴露 provider 注册、`ghttp.Request`或宿主内部模型。
- 运行`mise exec -- go test lina-core/pkg/plugin/capability/authcap/... -count=1`通过。
- 任务级`lina-review`读取`AGENTS.md`、`.agents/rules/openspec.md`、`.agents/rules/architecture.md`、`.agents/rules/plugin.md`、`.agents/rules/backend-go.md`、`.agents/rules/testing.md`和`.agents/rules/i18n.md`后未发现阻塞问题。
- 本任务没有新增运行期服务实例、数据库访问、缓存状态、HTTP文案或用户可观察页面，因此`DI`、数据权限、缓存一致性、`i18n`和`E2E`均无本任务级影响；对应行为由后续任务实施和验证。

### 任务 1.3 验证与审查

- `pluginhost.ProviderDeclarations`负责声明 scheme 与 factory，`plugin.Service.RegisterSourcePluginProviderFactories`负责把全部源码插件声明注册到唯一启动期 manager；跨插件和大小写归一后的重复 scheme 在发布前返回错误。
- `authspi.Manager`由`newHTTPRuntime`创建，owner 为宿主 HTTP 启动装配，创建位置是`internal/cmd/internal/httpstartup/httpstartup_runtime.go`，经插件集成注册后由同一`httpRuntime`共享；请求路径不会调用`NewManager()`。
- provider enablement 读取同一`pluginRuntime`实例，环境通过`pluginRuntime.AuthenticationProviderEnv`传递；factory 缺依赖或初始化失败不缓存错误并失败关闭，禁用前后均检查状态，同 owner factory 替换通过不可变 map 的`atomic.Pointer`发布。
- scheme 与 factory 声明是各节点相同的代码权威源；插件可用性复用宿主已有跨节点治理和 enablement 快照失效路径。本任务不新增独立跨节点业务缓存，provider 目录没有陈旧允许窗口，未知或不可确认状态均拒绝。
- 运行`mise exec -- go test lina-core/pkg/plugin/capability/authcap/authspi -count=1`、`mise exec -- go test -race lina-core/pkg/plugin/capability/authcap/authspi -count=1`、`mise exec -- go test lina-core/pkg/plugin/pluginhost -count=1`、`mise exec -- go test lina-core/internal/service/plugin -run 'TestSourceProviderAvailabilityFollowsEnabledSnapshot|TestNonExistent' -count=1`和`mise exec -- go test lina-core/internal/cmd -count=1`均通过。
- 任务级`lina-review`重新读取`AGENTS.md`及命中的 OpenSpec、架构、插件、后端、测试和缓存一致性规则后未发现阻塞问题；本任务无数据库、数据权限、用户可见`i18n`或`E2E`页面影响。

### 任务 1.4 验证与审查

- 宿主`model.Context`、`bizctx.Service`和插件可见`bizctxcap.CurrentContext`传播统一 actor；`SetUser`建立 user actor，`SetActor`建立 machine actor并清空用户 ID、用户名、token/session、impersonation和角色数据范围字段。
- 动态插件`IdentitySnapshotV1`追加`actorKind`、`subjectId`和`credentialId`，protobuf wire 使用新字段`15-17`并保留既有字段编号；WASM host call只从宿主 identity注入 actor和租户，machine identity携带的用户权限字段在进入 capability context前被清空。
- 复用任务 1.3 中由启动装配创建的同一`bizctx.Service`和动态 runtime/host service，不新增运行期依赖、缓存或数据库访问；数据权限 owner 后续只会看到可信 machine tenant而不会看到伪造用户数据范围。
- 运行`mise exec -- go test lina-core/internal/service/bizctx lina-core/pkg/plugin/capability/bizctxcap -count=1`、`mise exec -- go test lina-core/pkg/plugin/pluginbridge/internal/codec -count=1`、`mise exec -- go test lina-core/internal/service/plugin/internal/wasm -run 'TestContextWithHostCallBizContextRejectsMachineUserFields|TestHandleHostServiceInvokeOrgMethods' -count=1`、`mise exec -- go test lina-core/internal/service/plugin/internal/runtime -count=1`和`mise exec -- go test lina-core/internal/cmd -count=1`均通过。
- 任务级`lina-review`重新读取`AGENTS.md`及命中的 OpenSpec、架构、插件、后端、测试和数据权限规则后未发现阻塞问题；本任务没有用户可见行为，未触发`E2E`质量审查，`i18n`和缓存一致性无影响。

### 任务 1.5 验证与审查

- 新增`authcap.RouteAuthorization`与启动期共享`RouteAuthorizationCatalogue`，缺少`actors`默认只允许`user`；声明`machine`时强制完整`operation/resource/action`，并拒绝未知 actor、未知 action、重复路由和全局重复 operation。
- 目录 owner 为宿主 HTTP 启动装配，创建位置是`internal/cmd/internal/httpstartup/httpstartup_runtime.go`；同一实例经`plugin.Service`显式传入 dynamic runtime，静态、源码插件和动态插件 owner 更新使用不可变快照及`atomic.Pointer`发布，请求和管理读取不临时创建目录。
- 静态路由在`server.GetRoutes()`完成宿主与源码插件绑定后审计；源码插件在 registrar 捕获并校验 DTO 标签；动态插件在`RouteContract`产物解析、安装、启用、刷新和显式升级前验证，生命周期收敛后按 owner 原子替换，卸载移除。
- 目录保存已安装插件声明，插件当前是否生效继续由宿主统一 enablement 治理；因此禁用不会破坏策略 operation 引用，重新启用无需重建策略。未声明机器 actor 的现有接口不会进入机器目录，保持默认拒绝。
- 动态 WASM 构建器使用现有 Go AST 跨平台提取路径写入一等`RouteContract`字段；没有新增 shell、平台命令或路径语义，Linux、macOS和 Windows 共享同一 Go 实现。运行`mise exec -- go test ./hack/tools/linactl/internal/wasmbuilder -count=1`通过。
- 运行`mise exec -- go test -race ./pkg/plugin/capability/authcap -count=1`、`mise exec -- go test ./pkg/plugin/pluginbridge/... ./pkg/plugin/pluginhost ./pkg/plugin/capability/routecap ./internal/service/plugin/internal/integration ./internal/service/plugin/internal/runtime ./internal/service/plugin ./internal/cmd -count=1`、`openspec validate add-aksk-machine-authentication --strict`和`git diff --check`均通过。
- 本任务没有数据库访问、业务数据读取、缓存后端、用户可见文案或页面变化，因此数据权限、分布式缓存一致性、`i18n`和`E2E`均无任务级影响；路由目录是代码/active release声明的进程级不可变派生快照，不缓存业务授权结果。

### 任务 1.6 至 1.7 验证与审查

- 新增`authspi.Dispatcher`窄调用契约，由`newHTTPRuntime`创建的同一`authspi.Manager`实现，并同时显式注入宿主 middleware 与`plugin.Service -> dynamic runtime`；owner、创建位置、传递路径和共享实例与任务 1.3 的 provider manager一致，请求路径没有调用`NewManager()`或构造 provider。
- `Authorization`按 scheme token精确分派：`Bearer`只调用既有`AuthenticateAccessToken`、session、tenant和 role链；非 Bearer只调用匹配 provider一次，未知、不可用、格式错误或 provider失败返回`401`且不回退。认证成功后 actor未获路由、operation或 resource-action授权返回`403`。
- 静态与源码路由从当前 DTO handler读取`actors/operation/resource/action`；动态路由从 active `RouteContract`读取同一语义。机器请求忽略用户`permission`并要求 operation与资源动作双重允许；用户请求继续使用现有`permission`和角色快照，machine-only接口拒绝用户。
- Tenancy middleware识别可信 machine actor并保留 provider注入的固定 tenant，禁止请求头、查询参数或用户租户解析覆盖；动态机器 identity只投影 actor、subject、credential和 tenant，不创建用户 ID、token、session、角色或数据范围。
- 新增真实 GoFrame middleware链测试覆盖 machine成功、默认拒绝、operation拒绝、resource拒绝、provider失败早退、固定 tenant和 Bearer回归；新增 dynamic runtime测试覆盖 machine成功、actor拒绝及双重授权拒绝。顺带修复既有 tenancy测试使用`http://:port`导致的顺序相关 EOF，统一改为`127.0.0.1 + GetListenedPort()`。
- 运行`mise exec -- go test -race ./internal/service/middleware ./internal/service/plugin/internal/runtime -run 'TestMachineAuthenticationMiddlewareChain|TestBearerAuthenticationMiddlewareRegression|TestDynamicMachineRouteAuthorization' -count=1`和`mise exec -- go test ./pkg/plugin/capability/authcap/... ./internal/service/middleware ./internal/service/plugin/internal/runtime ./internal/service/plugin ./internal/cmd -count=1`均通过；`openspec validate add-aksk-machine-authentication --strict`与`git diff --check`通过。
- 本阶段只改变认证、请求上下文和授权控制流，不新增数据库查询、业务数据装配、缓存后端、前端页面或用户可见运行时文案资源；数据权限、分布式缓存一致性、`i18n`和`E2E`无本任务级新增影响，资源 owner的租户与数据权限接入仍由后续机器可访问业务接口实施。

### 任务 1.8 验证与审查

- 同步更新`apps/lina-core/pkg/plugin/README.md`与`README.zh-CN.md`，记录`ProvideAuthentication`声明边界、scheme唯一与精确分派、普通消费面和 provider SPI隔离，以及机器路由四类元数据和双重允许规则。
- 中英文镜像均明确缺少`actors`默认 user-only，静态、源码和动态路由分别在启动、注册、产物及生命周期边界失败关闭；文档没有提前承诺尚未实施的 AKSK插件内部配置或管理页面。
- 运行`openspec validate add-aksk-machine-authentication --strict`、`git diff --check`和双语关键事实静态检索通过。本任务只影响技术文档，无运行时、数据权限、缓存、`i18n`资源或`E2E`影响。

### 任务 2.1 验证与审查

- 创建前检查确认`apps/lina-plugins/linapro-auth-aksk`目录不存在，因此没有插件根`AGENTS.md`需要读取；创建后继续遵守顶层规范和已读取的插件、前端、i18n、文档与测试规则。
- 新插件声明`source`、`managed`、`tenant_aware`、`supports_multi_tenant: true`和`tenant_scoped`，显式启用`en-US/zh-CN`；通过`plugin_embed.go`嵌入清单、前端和 manifest资源，并由`backend/plugin.go`注册源码插件。
- 管理菜单挂载宿主`setting`目录，JWT权限按客户端、密钥、策略的查看、新增、修改、删除及策略授权拆分；英中`menu.json`覆盖全部13个菜单和按钮 key，插件名称与说明在独立`plugin.json`维护。
- 当前骨架不创建 service、controller、认证 provider、数据库访问或缓存实例，因此 DI、数据权限和缓存一致性无任务级影响；前端仅保留后续管理页面的无文案`Page`入口，不新增未翻译可见文本。
- 运行`GOWORK=off mise exec -- go test ./... -count=1`通过；根`go.work`按现有约定只包含`lina-core`和`linactl`，插件独立模块不修改 workspace。`openspec validate add-aksk-machine-authentication --strict`与`git diff --check`通过。

### 任务 2.2 验证与审查

- 新增单一迭代安装 SQL 和对应卸载 SQL，覆盖机器客户端、访问密钥、策略、策略接口、策略资源和客户端策略六张表；安装文件仅含 PostgreSQL DDL，不含 Seed 或 Mock 数据，也没有显式写入自增主键。
- 所有业务表、索引、唯一索引、外键和检查约束均使用完整`plugin_linapro_auth_aksk_`前缀；主键采用`BIGINT GENERATED ALWAYS AS IDENTITY`，关联字段保持`BIGINT`同宽，敏感密文和加密`nonce`使用`BYTEA`。
- 所有表均声明`created_at`、`updated_at`和`deleted_at`，后续由 GoFrame 自动维护时间和软删除；活动业务唯一性通过`WHERE deleted_at IS NULL`部分唯一索引表达，软删除后允许重新使用名称、客户端键、`AK`和关系组合。
- 安装 SQL 的表和索引均使用`IF NOT EXISTS`，卸载 SQL 使用`DROP TABLE IF EXISTS`并按客户端策略、策略资源、策略接口、访问密钥、策略、客户端的外键依赖逆序删除，满足重复初始化和重复卸载的幂等要求。
- 认证按未删除`AK`全局唯一索引定位；客户端与策略列表使用`tenant_id + status + created_at + id`索引；密钥、策略接口、策略资源和客户端策略关系均提供以租户和父实体开头的集合化查询索引，支持分页后批量装配并避免逐行查询。
- 本任务不新增运行期服务、接口、缓存或用户可见文案，因此 DI、数据权限执行、缓存一致性、`i18n`和 E2E 无任务级新增影响；数据库仍是后续机器认证和策略快照的权威源。
- 运行 SQL 静态检索、`openspec validate add-aksk-machine-authentication --strict`和`git diff --check`通过；真实数据库重复初始化、卸载和重新安装验证由任务 2.3、2.5 与 8.4继续覆盖。

### 任务 2.3 验证与审查

- 实施时确认插件共享`Makefile`只提供`ctrl`和`dao`，根`db.init`只初始化宿主 SQL；源码插件 SQL 的真实生产入口是宿主插件生命周期。因此将任务文字校正为先在本地开发库重放同一插件安装 SQL，再执行插件`make dao`，未修改 Makefile、`linactl`或其他开发工具。
- 在项目当前 PostgreSQL 开发库连续重放安装 SQL 两次，第二次六张表和全部索引均通过`IF NOT EXISTS`安全重入；数据库查询确认六张`plugin_linapro_auth_aksk_*`表均存在。
- `hack/config.yaml`只配置六张插件自有表，`removePrefix`为完整`plugin_linapro_auth_aksk_`，生成输出位于插件`backend/internal`；本地生成时临时使用当前开发库连接，完成后恢复仓库标准开发连接配置，未提交本地凭证。
- `DAO/DO/Entity`全部由`mise exec -- make dao`生成且未手工修改。生成实体主键、`client_id`和`policy_id`均为`int64`，与数据库`BIGINT`同宽；`tenant_id`、审计用户和状态为`int`；密文与加密`nonce`为`[]byte`；所有时间和软删除字段为`*time.Time`。
- GoFrame 通过`created_at`、`updated_at`和`deleted_at`元信息自动维护时间与软删除，后续服务不得手工写入这些字段或追加冗余`deleted_at IS NULL`过滤。本任务只生成数据访问层，不新增运行期服务实例、缓存、接口或用户可见文案，DI、缓存、数据权限执行、`i18n`和 E2E 无新增影响。
- 运行`GOWORK=off mise exec -- go test ./... -count=1`通过，覆盖插件根包、注册入口及全部生成包的编译门禁。

### 任务 2.4 验证与审查

- 新增插件内部启动期配置加载器，显式接收插件作用域`plugincap.ConfigService`和部署环境查询函数；运行期依赖 owner 仍是后续认证提供者工厂，配置只在提供者初始化阶段解析一次并以纯值快照传递，请求路径不读取环境变量或临时创建配置服务。
- 版本化主密钥环以版本到环境变量名的映射声明，活动版本必须存在，每个活动或历史版本都必须从部署环境解析为标准 Base64 编码的 32 字节`AES-256`密钥；任一声明版本缺失、解码失败或长度错误均失败关闭。仓库运行配置不提供内置密钥，示例只记录环境变量名。
- 配置默认签名窗口为`5m`、`nonce TTL`为`10m`、缓存最大陈旧时间为`30s`；校验要求`nonce TTL`至少为签名窗口两倍，缓存陈旧时间不得超过签名窗口，避免未来时间偏移请求在`nonce`过期后仍可重放。
- 每客户端密钥数、策略数、每策略接口数和资源数均提供默认值与绝对上限，后续管理服务和认证快照统一消费同一配置；密钥环读取返回副本，版本枚举不暴露密钥材料。
- 本任务有缓存配置影响但尚未创建缓存：数据库仍是权威源，后续任务 5.1 至 5.4实现事务后修订号、跨节点失效和故障关闭；本任务不新增 HTTP 接口、数据访问或用户可见文案，因此数据权限、接口性能、`i18n`和 E2E 无任务级新增影响。
- 运行`GOWORK=off mise exec -- go test ./backend/internal/config ./... -count=1`通过，覆盖多版本密钥环、可变副本隔离、缺失活动或历史密钥、不安全窗口和超限配置；`openspec validate add-aksk-machine-authentication --strict`与`git diff --check`通过。

### 任务 2.5 验证与审查

- 新增嵌入 SQL 契约测试，固定六张表的`CREATE TABLE IF NOT EXISTS`、`GENERATED ALWAYS AS IDENTITY`、自动时间与软删除字段、DDL-only 数据分类、部分唯一索引和外键逆序卸载顺序；测试同时禁止 Mock 目录、`ON DUPLICATE KEY`和`SERIAL`主键。
- 在项目当前 PostgreSQL 开发库连续执行两次卸载 SQL，再连续执行两次安装 SQL；重复卸载通过`DROP TABLE IF EXISTS`，重复安装通过表和索引的`IF NOT EXISTS`保护，最终六张表及 21 个主键、唯一和查询索引恢复完整。
- 在回滚事务中验证同租户活动客户端名称重复触发唯一冲突，设置`deleted_at`后同一业务名称可重新插入，证明部分唯一索引与软删除恢复语义一致；事务回滚后没有残留测试数据。
- SQL 分类结论为仅 DDL，无 Seed 和 Mock 数据。敏感配置分类结论为主密钥值只来自部署环境，`config.yaml`与`config.example.yaml`只包含空活动版本、环境变量名和非敏感阈值，不包含`SK`、主密钥或测试密钥。
- 高频路径索引覆盖按唯一`AK`认证、租户分页、客户端密钥批量读取、策略接口与资源集合读取以及客户端策略双向关联；后续任务 3.7和 4.3继续通过查询次数测试证明应用层不产生`N+1`。
- 运行`GOWORK=off mise exec -- go test ./... -count=1`、SQL 静态检索、`openspec validate add-aksk-machine-authentication --strict`和`git diff --check`均通过。本任务无新增运行期 DI、HTTP 接口、缓存实例、用户文案或 E2E 页面影响。

### 任务 3.1 验证与审查

- 按`machine-clients`、嵌套`access-keys`和`access-policies`三个 REST 资源拆分 17 个管理端点：读取使用`GET`、创建使用`POST`、更新与状态及授权替换使用`PUT`、删除使用`DELETE`，路径不包含动作动词。
- 所有管理接口只声明对应插件 JWT 菜单权限且不声明`actors`，因此沿用宿主缺省 user-only 行为，机器主体不能管理自身客户端、密钥或策略。客户端与策略列表均分页且`pageSize`上限为 100，列表直接投影关联计数；每客户端密钥列表由配置数量上限天然有界。
- 策略授权使用一个`PUT /access-policies/{id}/authorization`请求原子替换精确`operationCodes`、整体资源`read/write`和客户端关系，DTO 硬上限分别为 2048、512和 128，运行期配置可进一步收紧；契约不接受资源 ID、路径、条件或显式`deny`。
- 所有公开字段均包含`json`、`dc`和`eg`，响应时间点使用`int64`或`*int64` Unix 毫秒时间戳；主外键与生成实体保持`int64`同宽。普通密钥响应只含`AK`、状态和时间，不包含`SK`、密文、加密`nonce`或主密钥版本；只有创建响应包含一次性`secretKey`。
- 通过`mise exec -- make ctrl`生成三个 API 接口和 17 个控制器方法骨架，未手工创建或修改生成契约。控制器将在任务 3.2 至 3.8注入服务并绑定启动路由，当前不提前发布返回未实现错误的管理接口。
- `i18n`有影响：源 DTO 使用英文文档文本并新增`en-US/apidoc`空占位；最终`zh-CN`接口文档翻译和完整性门禁由任务 7.5在 DTO 稳定后统一完成。当前没有前端页面，E2E 由任务 8.1执行；缓存和数据权限执行从后续服务任务开始。
- 新增 API 契约测试固定 REST 方法与路径、JWT权限、user-only、文档标签、毫秒时间和密钥字段泄漏边界。运行`GOWORK=off mise exec -- go test ./backend/api/... ./backend/internal/controller/... ./... -count=1`通过。

### 任务 3.2 验证与审查

- 新增机器客户端`Service`，实现分页、详情、创建、更新、状态变更和软删除；服务显式接收启动期共享`bizctxcap.Service`，owner 为后续源码插件 HTTP 注册回调，创建位置将在任务 3.8收口，控制器与服务复用同一实例，请求路径不创建业务上下文服务。
- 所有方法先读取宿主可信`bizctx`并要求`ActorKindUser`、正`UserID`和非负租户，再执行任何数据库访问；查询和写入始终显式使用`tenant_id = current.TenantID`。即使平台上下文具有`PlatformBypass`也不会绕过到其他租户，跨租户与不存在 ID 均返回`AKSK_CLIENT_NOT_FOUND`。
- 数据权限例外：机器客户端是平台或当前租户共享治理资源，同一租户内具有 JWT 管理权限的管理员共享可见，不应用本人、部门或角色数据范围；服务仅使用可信租户和创建/更新用户审计字段，禁止请求参数指定租户。
- 创建使用`crypto/rand`生成带`mc_`前缀的 144 位随机稳定客户端键；同租户活动名称冲突通过预检和 PostgreSQL 约束双层映射为稳定`bizerr`，客户端键碰撞最多重试三次。业务状态使用命名类型与`enabled/disabled`常量。
- 删除在一个 GoFrame 事务内软删除客户端、该客户端全部访问密钥和客户端策略关系；所有时间与`deleted_at`继续由 GoFrame 自动维护，不手工写时间或冗余过滤。客户端名称在软删除后可安全复用。
- 列表先分页读取当前租户客户端，响应模型已预留批量密钥和策略计数；任务 3.7将用固定数量集合查询填充计数，当前实现没有按行数据库查询。缓存尚未创建，写后跨节点失效由任务 5.1至 5.2接入，在认证提供者启用前不存在陈旧授权放行路径。
- 新增英文和中文结构化错误资源，覆盖管理用户要求、not-found、名称必填/冲突、状态和密钥生成失败；本任务无新增页面文案，E2E 在任务 8.1覆盖。
- 使用当前开发 PostgreSQL 显式运行`LINA_TEST_PGSQL_LINK=... GOWORK=off mise exec -- go test ./backend/internal/service/client -count=1`通过，覆盖租户共享与隔离、冲突、更新、启停、机器拒绝、事务软删除及名称复用；`GOWORK=off mise exec -- go test ./... -count=1`与严格 OpenSpec 校验通过。

### 任务 3.3 验证与审查

- 新增每客户端多密钥服务，构造函数逐项显式接收启动期共享`bizctxcap.Service`、机器客户端`Service`、密钥加密`Service`和纯值密钥数量上限；owner 为任务 3.8的插件 HTTP 启动装配，控制器与后续认证提供者复用同一服务图，不在请求路径临时创建依赖。
- `AK`使用 144 位`crypto/rand`随机值和`lak_`前缀，`SK`使用 256 位独立随机值和`lsk_`前缀；全局活动`AK`唯一冲突最多重试三次。每客户端非删除密钥数达到配置上限即拒绝创建，软删除后释放名额。
- 创建支持可选未来过期时间，服务端拒绝过去时间；列表受每客户端配置上限约束并只返回公开`AK`、状态和 Unix 毫秒时间，不投影`SK`、密文、加密`nonce`或主密钥版本。公开服务不存在修改`SK`方法，轮换只能创建新密钥后停用或删除旧密钥。
- 所有密钥查询和写入先通过客户端服务验证 user-only 与精确租户，再使用`tenant_id + client_id + id`过滤；错客户端、跨租户和不存在密钥统一返回 not-found。状态和删除不改变其他密钥，满足不中断轮换。
- 新增稳定`bizerr`及英中资源，覆盖密钥不存在、数量上限、过期时间、状态、生成和加密失败；普通错误和日志从不拼接`SK`。缓存失效将在任务 5.2统一接入，认证提供者尚未启用，因此当前没有撤销陈旧放行窗口。
- PostgreSQL 集成测试覆盖两把独立密钥、第三把超限、过去过期时间、密文存储解密、普通列表、跨租户和错客户端、启停、软删除及名额释放。使用随机租户隔离并通过`-p=1`避免共享开发库 DDL 并发。

### 任务 3.4 验证与审查

- 新增插件内部`secretcrypto.Service`，使用标准库`AES-256-GCM`、每次随机 12 字节`nonce`和版本化主密钥；关联数据按固定协议包含版本标识、租户、客户端 ID 和`AK`，任一绑定字段变化都会导致认证失败。
- 新密文使用活动主密钥版本，解密按数据库保存版本读取活动或历史密钥；缺失版本、密钥长度错误、密文/nonce格式错误和关联数据篡改全部失败关闭。密钥环返回的副本在每次加解密后清零，组件不保留明文`SK`。
- 实施测试定位到 GoFrame PostgreSQL驱动`v2.10.0`会把非 JSON 切片中的`[`/`]`字节替换为`{`/`}`，随机破坏`BYTEA`密文。宿主和插件统一升级至官方已修复该问题的`github.com/gogf/gf/contrib/drivers/pgsql/v2 v2.10.1`，保留正确`BYTEA`存储而不引入 Base64 补救模型。
- 驱动属于宿主数据库运行期依赖，由 Go 模块启动装配加载；只升级同系列补丁版本，不新增服务实例或传递路径。插件加密服务由插件启动入口创建一次并同时供密钥管理和认证提供者使用，密钥环来自任务 2.4的不可变配置快照。
- 同步修复 SQL 治理遗漏：六表列标识符全部使用 PostgreSQL双引号，每个建表语句前增加中英用途注释；卸载/安装/重复安装、插件 SQL 契约测试和宿主`pkg/dialect`治理测试通过。
- 单元测试覆盖当前版本往返、随机 nonce、密文不含明文、租户/客户端/AK篡改、历史版本轮换读取和缺失历史版本；密钥数据库往返连续运行 20 次通过，客户端与密钥 PostgreSQL 测试串行重复 5 次通过。`mise exec -- go test ./pkg/dialect ./internal/cmd -count=1`与插件全包测试通过。

### 任务 3.5 验证与审查

- 新增策略`Service`并实现分页、详情、创建、更新、启停和事务软删除；所有数据库读取与写入先要求可信`user`主体，再使用精确`tenant_id`过滤。策略继续作为平台或当前租户共享治理资源，不应用本人或部门数据范围，跨租户与不存在 ID 统一返回不透明 not-found。
- 策略授权替换先对宿主机器路由目录、资源整体`read/write`、客户端租户和配置上限执行集合化校验，再在一个 GoFrame事务内软删除旧接口、资源和客户端关系并批量插入归一化后的新集合；重复接口和客户端去重，重复资源授权合并为允许并集，不接受资源 ID、条件、路径模式或显式`deny`字段。
- 每客户端策略上限通过一次客户端集合查询和一次分组计数查询验证；详情固定使用三次关系集合查询，写入使用有界批次，不按接口、资源或客户端逐项查询。列表关联计数仍由任务 3.7统一批量装配。
- 新增 PostgreSQL集成测试覆盖 CRUD、同租户共享、跨租户不透明、机器主体拒绝、名称冲突、未知接口和资源、无效读写授权、操作与资源上限、重复项归一化、跨租户客户端、每客户端策略上限、拒绝时无部分关系、真实插入失败事务回滚以及策略和三类关系软删除。
- `i18n`有影响：新增的 11 个策略`bizerr`已同步维护`en-US`和`zh-CN`结构化错误资源；当前没有前端页面或用户交互变更，E2E继续由任务 8.1覆盖。缓存失效尚未接入，数据库仍为唯一权威源，认证提供者尚未启用，任务 5.1至 5.2负责事务提交后的跨节点失效。
- 策略服务的运行期 owner 是任务 3.8的插件启动装配；构造函数逐项接收启动期共享`bizctxcap.Service`、宿主路由目录和纯值配置上限，请求路径不创建服务或目录。运行`LINA_TEST_PGSQL_LINK=<current-project-link> GOWORK=off mise exec -- go test -p=1 ./backend/internal/service/policy -count=5`、`GOWORK=off mise exec -- go test ./... -count=1`、`openspec validate add-aksk-machine-authentication --strict`和`git diff --check`均通过。

### 任务 3.6 验证与审查

- 扩展 core-owned`routecap.Service`，新增最多 10000 条声明的机器接口与资源有界只读投影；接口项包含 owner、方法、路径、稳定`operation/resource/action`和`active`，资源项同时表达已声明与当前 active 的`read/write`模式。超出调用上限、目录依赖缺失或插件状态无法解析时失败关闭，不返回截断目录。
- 目录 owner 仍是宿主 HTTP启动装配创建的同一`authcap.RouteAuthorizationCatalogue`。该实例显式传入`NewHostServices`并由`capabilityhost`构造`Route()`只读适配器，同时继续传入插件 service负责路由生命周期；请求路径不创建目录或插件状态服务。
- 插件 owner启用状态通过新增`ResolveBusinessEntryEnablement`批量契约解析：一次 manifest扫描、一次启动数据快照、一次注册表读取和一次租户状态集合查询覆盖全部 distinct owner；租户状态通过 request context批量快照复用，避免按路由或插件执行`N+1`查询。宿主路由始终 active，源码和动态插件路由按当前平台或租户治理状态标记。
- 插件禁用不会从全局声明目录删除策略所引用的 operation；目录将其投影为`active=false`，策略服务拒绝新提交的 inactive operation或资源动作但不删除已有关系。重新启用后同一声明按最新批量状态恢复 active，已有策略无需重建。
- 源码插件通过`capability.Services.Route()`直接消费；动态插件通过新增受治理`host:route/machine_authorizations.list`方法、统一 JSON信封、guest SDK和 WASM dispatcher消费同一 DTO。宿主插件公共中英文 README已同步目录能力说明和生成 host-service表。
- 新增 JWT user-only的`GET /machine-authorizations`管理 DTO和生成控制器骨架，使策略页面可用一次请求获得完整有界快照；最终控制器委托由任务 3.8接入。`i18n`有接口文档影响，英文源文本已完成，`zh-CN apidoc`翻译与完整性门禁由任务 7.5统一维护；当前没有页面行为，E2E仍在任务 8.1。
- 测试覆盖 host/source/dynamic owner、目录上限、一次 owner批量解析、禁用保留声明、资源 active模式和重新启用恢复；动态 host service测试覆盖同契约投影，策略 PostgreSQL测试覆盖 inactive拒绝和 user-only目录读取。运行`mise exec -- go test -race ./internal/service/plugin/internal/capabilityhost ./internal/service/plugin/internal/integration ./internal/service/plugin/internal/wasm ./pkg/plugin/capability/routecap ./pkg/plugin/pluginbridge/internal/domainhostcall -count=1`、核心插件/协议/启动包测试、`LINA_TEST_PGSQL_LINK=<current-project-link> GOWORK=off mise exec -- go test -p=1 ./backend/internal/service/policy -count=3`、插件全包测试、严格 OpenSpec校验和`git diff --check`均通过。

### 任务 3.7 验证与审查

- 客户端列表先执行租户过滤后的总数和当前页查询，再对当前页客户端 ID 分别执行访问密钥总数、启用且未过期密钥数和关联策略数三次`GROUP BY`查询；非空页面最多固定 5 次数据库查询，查询次数不随页面行数增长。
- 策略列表先执行租户过滤后的总数和当前页查询，再对当前页策略 ID 分别执行接口数、资源数和关联客户端数三次`GROUP BY`查询；非空页面同样最多固定 5 次数据库查询，所有结果通过内存映射合并，数据库调用均位于动态行循环之外。
- 所有聚合查询同时使用可信业务上下文中的精确`tenant_id`与当前页 ID 集合，空关系保持零值，不会通过计数、总数或关联项泄露其他租户数据。机器治理资源继续采用同租户共享的数据权限例外，不应用本人或部门范围。
- PostgreSQL集成测试覆盖一个客户端的 3 把密钥中仅 1 把启用且未过期、1 个关联策略，以及一个策略的 2 个接口、1 个资源和 1 个客户端关联；同时覆盖无关系策略的三个零计数。静态审查固定查询语句与循环位置，作为等价查询次数证据。
- 运行`GOWORK=off mise exec -- go test ./backend/internal/service/client ./backend/internal/service/policy -count=1`和`LINA_TEST_PGSQL_LINK=<current-project-link> GOWORK=off mise exec -- go test -p=1 ./backend/internal/service/client ./backend/internal/service/policy -count=3`通过。本任务不新增运行期依赖、缓存状态、HTTP契约或用户可见文案，因此 DI、缓存一致性、`i18n`和 E2E 无任务级新增影响。

### 任务 3.8 验证与审查

- `backend/plugin.go`在源码插件 HTTP 启动回调中加载一次插件配置并创建共享`secretcrypto.Service`、客户端服务、密钥服务、策略服务和三个控制器；`BizCtx`、`Route`与插件配置 owner均为宿主 registrar，主密钥来自部署环境并只在启动期解析。控制器和后续认证路径复用该启动期服务图，请求方法只做 DTO 转换与 service委托，不临时创建关键服务。
- 三组管理控制器统一绑定在宿主`NeverDoneCtx -> HandlerResponse -> CORS -> RequestBodyLimit -> Ctx -> Auth -> Tenancy -> Permission`链上；API 契约测试固定 18 个端点的 JWT菜单权限且`actors`为空，插件装配测试固定`Permission`为控制器前最后一道门禁，因此无权限 JWT在数据库访问前返回`403`，管理接口不会向 machine actor开放。
- 客户端、密钥和策略控制器构造函数逐项拒绝缺失 service；插件回调同时拒绝缺失 registrar、路由、中间件、`BizCtx`、`Route`或插件配置能力，并以 error返回初始化失败，不在可预期配置错误上 panic。
- 控制器测试覆盖客户端详情聚合投影、机器授权目录投影和密钥创建的一次性`secretKey`响应；普通密钥列表 DTO没有`SK`、密文、加密`nonce`或主密钥版本字段。客户端详情复用三次固定分组聚合，响应中的密钥和策略计数与公开契约一致。
- PostgreSQL服务测试覆盖客户端、密钥和策略三类机器主体拒绝、同租户共享、跨租户不透明、名称唯一冲突、策略批量替换原子回滚、客户端及密钥和策略关系软删除、密钥数量释放和一次性密钥加密持久化；连续三次串行执行均通过。
- 运行`LINA_TEST_PGSQL_LINK=<current-project-link> GOWORK=off mise exec -- go test -p=1 ./backend/internal/service/client ./backend/internal/service/accesskey ./backend/internal/service/policy -count=3`、`GOWORK=off mise exec -- go test ./... -count=1`、`openspec validate add-aksk-machine-authentication --strict`和两层`git diff --check`均通过。缓存失效仍由任务 5.1至 5.2接入；当前新增源 API 文本的`zh-CN apidoc`和页面 E2E分别由任务 7.5与 8.1统一完成，开发工具与跨平台入口无修改。

### 任务 4.1 至 4.2 验证与审查

- 新增插件内部`hmacprotocol`组件，严格解析`LINA-HMAC-SHA256 Credential=<AK>,Signature=<hex>`、`X-Lina-Date`、`X-Lina-Nonce`和`X-Lina-Content-SHA256`；拒绝未知或重复认证字段、多值签名头、非规范 Unix秒、非小写 64 位十六进制摘要/签名、填充或长度不安全的 Base64URL nonce，以及与宿主正文摘要不一致的输入。
- 固定规范串依次包含`LINA-HMAC-SHA256-V1`、大写 HTTP方法、逐段解码再按 RFC 3986重编码的路径、按编码后键和值排序且保留重复键的查询、正文摘要、Unix秒和 nonce。路径不执行会改变路由语义的 clean，编码斜线保持为`%2F`，空路径稳定为`/`。
- 百分号编码只保留 RFC 3986非保留字符，空格使用`%20`而不是`+`，十六进制统一大写；查询重复键、Unicode键值、加号、空值和空查询均使用同一集合化排序逻辑，没有依赖平台 URL序列化差异。
- 固定测试向量使用 JSON正文`{"name":"Lina","enabled":true}`、重复和 Unicode查询及转义路径，规范串签名固定为`a2d5f047e687d88b88b7ed3e17d86bab7ebad3b9efe311e7c99e759a9d8a52b3`，并通过独立`openssl`计算核对。测试同时覆盖空正文标准 SHA-256、JSON正文摘要和认证格式失败路径。
- 运行`GOWORK=off mise exec -- go test ./backend/internal/hmacprotocol -count=20`、`GOWORK=off mise exec -- go test ./... -count=1`和插件工作区`git diff --check`通过。本阶段只新增无状态协议逻辑，不新增运行期 DI、数据库、缓存、数据权限、用户可见文案或 E2E影响；固定向量将在任务 7.6同步到双语文档。

### 任务 4.3 至 4.4 验证与审查

- 新增`authprovider.SnapshotStore`及 PostgreSQL权威实现：先按活动唯一`AK`索引读取密文凭证，再按凭证固定`tenant_id + client_id`读取客户端，之后以客户端策略、启用策略、策略接口和策略资源四次有界集合查询装配允许并集。有效冷读最多固定 6 次查询，数据库调用均在动态关系循环外，禁用凭证或客户端只需前两次查询。
- 客户端策略数、每策略接口数和资源数均使用启动配置计算`limit + 1`失败关闭边界；查询只投影客户端状态、策略 ID、操作码和资源读写位等必要字段。跨租户边界来自唯一 AK记录并在每个后续查询中显式使用同一`tenant_id`，未知、已删除或孤立凭证统一返回不透明`AKSK_AUTH_CREDENTIAL_INVALID`。
- `authprovider.Provider`逐项接收快照仓库、共享`secretcrypto.Service`、必需的`ReplayGuard`和纯值时间配置；构造阶段拒绝缺失重放保护或不安全时间窗。提供者先验证正文摘要与签名时间，再读取快照、检查密钥和客户端状态及过期时间、按关联数据解密 SK、使用`hmac.Equal`恒定时间比较签名，最后才原子消费 nonce。
- 解密后的 SK只存在于当前调用字节切片并在返回前清零，不进入 actor、授权快照、错误、缓存或日志。返回 actor只包含`machine`、稳定客户端键、数字凭证记录 ID和可信租户；授权结果通过`authcap.NewAuthorizationSnapshot`构造不可变语义的接口与资源允许集合。
- 新增`AKSK_AUTH_REQUEST_INVALID`、时间过期、凭证无效、密钥停用、客户端停用、密钥过期、签名无效、重放拒绝和后端不可用稳定`bizerr`，全部映射为`401`并同步英中错误资源。有效凭证无策略时认证成功但返回空允许集合；接口或资源任一不满足继续由宿主稳定`MIDDLEWARE_HTTP_FORBIDDEN`映射为`403`，两者都满足才进入控制器。
- 单元测试覆盖有效签名、接口与资源双重允许、写权限默认拒绝、正文篡改、过期时间、密钥/客户端状态、密钥过期、签名不匹配、重复 nonce、协调错误和无策略默认拒绝。PostgreSQL测试通过真实管理服务创建加密凭证与策略，验证完整认证、策略停用、密钥停用、客户端停用和未知 AK；连续三次通过。
- 运行`LINA_TEST_PGSQL_LINK=<current-project-link> GOWORK=off mise exec -- go test -p=1 ./backend/internal/authprovider -count=3`、`GOWORK=off mise exec -- go test -race ./backend/internal/authprovider ./backend/internal/hmacprotocol -count=1`、插件全包测试、双语 JSON解析和静态查询循环审查通过。缓存权威封装与真实重放实现由任务 5.2至 5.3接入，当前 provider在没有`ReplayGuard`时无法构造；本阶段无前端或 E2E新增影响。

### 任务 4.6 至 4.7 验证与审查

- 新增启动期共享`lastused.Service`，按访问密钥 ID维护固定窗口、`inFlight`和最近成功时间；同一密钥窗口内最多一次更新，并发请求在首个更新执行期间直接合并。数据库更新使用生成`DO`和精确`tenant_id + id`条件写入`last_used_at`，删除或撤销后的缺失行是安全无操作。
- 合并状态使用固定容量和最久未使用淘汰，永不淘汰执行中条目；容量全部被执行中条目占用时跳过非权威使用时间更新，不扩张内存。窗口与容量新增为配置项，默认`5m`和 10000，分别限制为不超过`24h`和 100000，非法值启动失败。
- 更新时间失败会记录只含数字凭证 ID的 warning，不推进最近成功时间，因此后续请求可重试；`Record`不向 provider返回错误，已完成签名、重放和授权快照验证的请求不会因非权威审计时间写入失败改写为认证失败。
- 提供者并发测试使用 64 个独立 nonce同时验证同一 AK，全部成功且重放接缝恰好调用 64 次；无效签名在 replay和 last-used之前拒绝。错误文本静态断言不含 AK、SK、完整签名或 nonce，解密器包装测试确认`Authenticate`返回后同一明文字节切片全部清零。
- 密钥轮换由版本化`secretcrypto`测试覆盖：旧版本密文可由包含历史版本的新密钥环读取，缺失历史版本失败关闭；每次加密使用独立 nonce且关联租户、客户端和 AK。Provider、合并记录器和密钥组件连续三次`-race`通过。
- PostgreSQL完整认证测试使用真实`lastused.NewDatabaseUpdater`，成功请求后验证`last_used_at`已持久化；单元测试覆盖窗口合并、64 请求并发单写、失败后重试和容量淘汰。运行`GOWORK=off mise exec -- go test -race ./backend/internal/authprovider ./backend/internal/lastused ./backend/internal/secretcrypto -count=3`、数据库测试、插件全包测试、严格 OpenSpec校验和`git diff --check`通过。
- 任务 4.5仍等待任务 5.1至 5.3提供启动期共享修订号、版本化快照和原子 nonce依赖后再注册真实 provider factory；当前构造函数强制`ReplayGuard`非空，禁止提前发布仅验证签名而没有重放保护的认证路径。本阶段无数据权限、前端或 E2E新增影响，`i18n`只新增配置与内部日志，无用户可见新文案。

### 任务 5.1 验证与审查

- 新增 core-owned`authcap/machinecoord.Service`，固定注册`machine-access`域并由宿主按调用插件 ID绑定；修订号 scope 使用`provider-<plugin-id> + tenant`，插件输入不能指定其他 owner。数据库中的机器凭证和访问策略表是权威源，允许的失效原因只有`credential`、`policy`和显式`recovery`。
- 单机模式使用启动期共享进程修订号，集群模式使用共享修订号和协调事件；域声明记录配置的最大陈旧窗口、`fail-closed`故障策略、权威源和同步机制。现有`cachecoord.Snapshot`继续提供本地/共享修订号、最后同步时间、后端健康、订阅状态、最近错误和陈旧秒数，恢复路径是后端恢复后重新读取修订号并由调用方重建数据库快照，或显式发布`recovery`修订。
- `newHTTPRuntime`是`cluster.Service`、`coordination.Service`和`cachecoord.Service`的 owner；生产装配改用`cachecoord.DefaultWithCoordination(clusterSvc, coordinationSvc)`复用同一共享后端，并把三项依赖逐项传入`capabilityhost.New`。源码插件通过 plugin-scoped capability直接调用，动态插件通过受治理`host:auth:machine_coordination`协议和 WASM dispatcher调用，请求路径不构造协调实例。
- 共享重放方法只接受 32 字节小写 SHA-256摘要，协调键由宿主使用插件、租户和摘要构造，不暴露原始 AK或 nonce；只在集群模式且共享 KV与 KeyBuilder均可用时调用原子`SetNX`，后端缺失或错误均失败关闭。实际本地有界去重和 provider接入由任务 5.3完成。
- 测试覆盖未绑定调用拒绝、插件与租户修订隔离、域一致性元数据、集群共享`SetNX`首次接受与重复拒绝、租户重放隔离、共享后端不可用失败关闭、动态协议目录和显式 WASM注册表。运行`mise exec -- go test ./internal/service/plugin/internal/capabilityhost ./pkg/plugin/pluginbridge/... ./internal/service/plugin/internal/wasm ./internal/cmd/internal/httpstartup -count=1`通过。
- 本任务没有新增数据库表、HTTP API或用户可见文案，数据权限和`i18n`无新增影响；同步更新了`apps/lina-core/pkg/plugin/README.md`与`README.zh-CN.md`的生成式 host-service表，开发工具和跨平台入口无影响。

### 任务 4.5、5.2 至 5.4 验证与审查

- `linapro-auth-aksk`通过`ProvideAuthentication("LINA-HMAC-SHA256", factory)`注册真实 provider。宿主`authspi.ProviderEnv`只盖章插件 ID、插件作用域配置和插件绑定`MachineCoordination`；factory一次性构造配置、版本密钥环、数据库权威读取器、版本快照、重放保护和`last_used_at`合并器，`authspi.Manager`缓存该 provider实例，请求路径不重建服务图。缺少任一环境依赖、主密钥或协调能力时 scheme初始化失败关闭。
- provider生命周期继续由宿主原子 scheme快照和`IsProviderEnabled`在调用前后治理。现有 manager测试覆盖未安装或禁用不可用、重新启用后恢复、factory失败后可重试和同 owner升级原子切换；源码插件安装、启用和卸载使用宿主统一生命周期测试。`TestBearerAuthenticationMiddlewareRegression`证明 provider注册不改变 JWT分派、用户上下文或角色权限链。
- 新增固定容量 10000 的 LRU版本快照，缓存内容仅包含公开 AK标识、客户端和凭证状态、过期时间、`AES-256-GCM`密文及非敏感 operation/resource允许集合；每次返回深复制，解密 SK永不进入缓存。首次唯一 AK查询得到租户后读取前修订号，完成总计最多六次的有界数据库装配后读取后修订号；修订不一致最多重试一次，禁止把旧快照标记成新版本。
- 缓存命中读取当前租户修订号：版本相同则复用并更新确认时间，版本变化则立即丢弃并从数据库重建；协调读取失败只允许使用最近确认且未超过`cacheMaxStaleness`的密文快照，超过窗口返回`AKSK_AUTH_UNAVAILABLE`。固定容量采用 LRU淘汰，数据库和协调恢复后无需重启即可重新确认或重建。
- 客户端和访问密钥服务逐项注入同一 plugin-scoped`machinecoord.Service`并在成功数据库写或事务提交后发布`credential`；策略元数据与 operation/resource/client关系在提交后发布`policy`。发布失败向管理调用返回错误，不静默报告完整成功。真实 PostgreSQL测试强制发布失败并确认客户端、密钥、策略和关系替换四条写路径均返回错误。
- 新增启动期共享重放 guard，使用固定容量 100000 的本地摘要表和 TTL到期清理；摘要按长度前缀编码`AK + nonce`后计算 SHA-256，内存与协调键不保存原始 AK、nonce、SK、签名或规范串。单机并发由互斥临界区保证只有一个赢家；集群模式还必须调用宿主共享`SetNX`，共享拒绝或后端错误均失败关闭。配置已保证`nonceTTL >= 2 * signatureWindow`，覆盖允许的完整时间偏移窗口。
- 单元和竞态测试覆盖本地重复与到期、64并发单赢家、容量耗尽失败关闭和到期恢复、跨节点共享单赢家、共享后端不可用、快照副本隔离、版本变化重载、修订加载竞态、LRU淘汰、陈旧窗口失败关闭和协调恢复。宿主测试覆盖两个独立`cachecoord`节点共享租户修订、插件与租户隔离、修订后端不可读及恢复发布。
- 真实 PostgreSQL测试使用两个独立本地快照 store预热同一凭证，再通过真实服务禁用策略、密钥和客户端；两侧下一次读取都观察共享修订并返回撤销后的数据库状态。运行`LINA_TEST_PGSQL_LINK='pgsql:minster:@tcp(127.0.0.1:5432)/ruyi?sslmode=disable' GOWORK=off mise exec -- go test -p=1 ./backend/internal/authprovider ./backend/internal/service/client ./backend/internal/service/accesskey ./backend/internal/service/policy -count=3`通过。
- 运行插件全包测试、`GOWORK=off mise exec -- go test -race ./backend/internal/authprovider ./backend/internal/replayguard ./backend/internal/lastused ./backend/internal/secretcrypto -count=3`及宿主`authcap`、`pluginbridge`、`pluginhost`、`middleware`、`plugin/...`和 HTTP启动测试通过；严格 OpenSpec校验与两层`git diff --check`通过。本阶段无新 HTTP DTO、数据库结构、前端文案或 E2E行为，数据权限和`i18n`无新增影响；开发工具和跨平台入口无影响。

### 任务 6.1 至 6.2 验证与审查

- `apidoc.Service`显式接收宿主 HTTP 启动装配创建的同一`authcap.RouteAuthorizationCatalogue`；owner、创建位置和传递路径均与路由审计、中间件及插件生命周期共享，文档请求路径不创建独立目录或认证服务。
- 宿主文档固定发布`LinaHMAC`安全方案；只有在共享目录中显式允许`machine`的操作才生成`LinaHMAC`操作声明和`x-lina-operation/resource/action/actors`投影。`user,machine`用两个独立 security requirement 表达或语义，未开放机器的接口继续继承现有`BearerAuth`默认且不暴露机器元数据。
- 动态插件公开文档路径会映射回插件内路由路径，并同时校验`dynamic`的 owner 类型和插件 ID，不会把其他插件的同名路由误投影。静态、源码插件和动态插件都使用同一目录事实源。
- 中文`apidoc`资源和嵌入镜像已同步`LinaHMAC`说明，`en-US`保持空占位并使用 Go 英文源文本。本任务无数据库、数据权限、缓存、前端页面或 E2E 新增影响，开发工具与跨平台入口不变。
- 测试覆盖宿主、源码插件、动态插件的 machine-only 和 user-or-machine 声明、未开放机器接口、中文安全方案，并复用目录套件对缺失资源、未知动作和重复 operation 的失败关闭验证。运行`mise exec -- go test ./pkg/plugin/capability/authcap ./pkg/plugin/pluginbridge/contract ./internal/service/plugin/internal/integration ./internal/service/plugin/internal/runtime ./internal/service/apidoc ./internal/cmd/internal/httpstartup -count=1`通过，`git diff --check`和双份 JSON 解析通过。

### 任务 6.3 至 6.6 验证与审查

- `linapro-monitor-operlog`通过新增幂等`002-add-aksk-machine-authentication.sql`保存`actor_kind`、`subject_id`、脱敏`credential_id`、`operation_code`、`resource_code`和`access_mode`；日志写入时使用宿主可信`bizctx`和路由元数据去归一化，列表、详情和导出无需跨插件逐行查询，比批量补查更严格地保持零额外数据库读取。机器日志的`oper_name`保持空值，不伪造用户名。
- `AKSK`provider将完整 AK 作为请求期凭证标识放入可信 actor，审计写入前立即转为只保留六位前缀和四位后缀的脱敏值；持久化、API和页面均不包含完整 AK。用户 actor 继续使用真实用户名和用户主体 ID。
- 请求头不进入审计模型；请求 JSON、URL query、普通响应、动态响应回退和错误文本共用结构化与文本双层清理，统一移除`Authorization`、AK/SK、签名、主密钥、密文和`nonce`，且始终在 2000 字节截断前脱敏。测试覆盖一次性 SK JSON、动态纯文本响应、超长边界、错误行和 URL 签名参数。
- 新增启动期共享`securityevent.Service`，owner 是`linapro-auth-aksk`的 provider factory，创建位置为`backend/plugin.go`，逐项传入同一 provider 实例并由`authspi.Manager`缓存复用。它以 SHA-256 摘要键维护最多 10000 个“稳定错误码 + 来源 + 脱敏凭证”桶，每分钟最多输出一条日志；其余计入`suppressed`，容量耗尽计入`dropped`，不保存 nonce、签名、SK 或规范串。
- SQL 先安装`001`再执行`002`，随后连续两次重复执行`002`均通过；使用本机`minster/ruyi`运行插件`make dao`生成 DAO/DO/Entity，开发连接配置已恢复且无差异。迁移只包含 DDL，无 Seed 或 Mock 变化；新索引支持租户 + 主体类型/主体 + 时间的真实审计查询路径。
- 运行`GOWORK=off mise exec -- go test ./... -count=1`覆盖操作日志插件，`GOWORK=off mise exec -- go test -race ./backend/internal/securityevent ./backend/internal/authprovider -count=3`和 AKSK 全包测试通过；真实 PostgreSQL 认证测试连续三次通过，宿主`bizctx/middleware/apidoc/httpstartup`回归通过，严格 OpenSpec 校验与两层`git diff --check`通过。`i18n`已同步运行时页面与插件`apidoc`字段；用户可观察页面的 E2E 和截图门禁在任务 8.1 至 8.3 统一执行。

### 任务 7.1 至 7.4 验证与审查

- 新增单页直连的机器访问工作台，使用现有`Page`、`useVbenVxeGrid`、`useVbenForm`、`useVbenDrawer`、`useVbenModal`、`Popconfirm`和`ghost-button`完成客户端筛选、增改、启停、删除、密钥管理与策略管理；页面 API均为分页或有界批量读取，没有逐行详情补查和前端瀑布式装配。
- 密钥列表 DTO只包含公开`AK`及状态、过期和使用时间；创建响应中的`SK`仅写入一次性弹窗局部`shallowRef`，弹窗关闭后置空，列表刷新和详情请求均不会重新获取`SK`。复制操作只在该弹窗内可用。
- 策略授权一次并行读取策略详情、机器授权目录和最多 100 个客户端候选，接口树只提交稳定`operationCodes`，资源区域只提交`resourceCode/read/write`，不会构造资源 ID、路径、条件或显式`deny`字段。
- 页面内所有新增、编辑、启停、删除、密钥和授权入口都通过`useAccess().hasAccessByCodes`按`plugin.yaml`中的精确按钮权限完全隐藏；缺少策略查看权限时整个策略标签不渲染。插件禁用时菜单、动态页装配、后端路由和认证提供者继续由宿主统一插件生命周期移除，数据保留；启停的浏览器级验证在任务 8.1 至 8.3执行。
- 使用临时外部源码类型检查配置运行`mise exec -- pnpm exec vue-tsc --noEmit --skipLibCheck -p apps/web-antd/tsconfig.aksk-check.json`通过后已删除该临时文件；随后运行页面`prettier`、`mise exec -- make i18n.check`和`GOWORK=off mise exec -- go test ./... -count=1`均通过。本阶段无新增后端运行期依赖、数据库、缓存或数据权限变化；`i18n`和 E2E分别继续由任务 7.5与 8.1 至 8.3收口，开发工具和跨平台入口无修改。

### 任务 7.5 至 7.6 验证与审查

- 插件显式启用`en-US`和`zh-CN`，两种语言均维护运行时页面、菜单和结构化`bizerr`资源；英文 API 文档继续直接使用 DTO源文本，`en-US/apidoc`保持空占位，`zh-CN/apidoc`完整覆盖客户端、访问密钥、策略和机器授权目录的`tags`、`summary`、接口说明与字段说明。
- 在插件`backend/api/api_contract_test.go`增加本地接口文档完整性门禁，从全部公开 DTO结构标签推导`plugins.linapro_auth_aksk.*`稳定键，缺失任一中文翻译或英文占位出现重复翻译都会失败。运行`GOWORK=off mise exec -- go test ./backend/api ./backend/internal/hmacprotocol ./backend/internal/config -count=1`、全部资源`jq empty`和`mise exec -- make i18n.check`均通过。
- 新建内容镜像一致的`README.md`与`README.zh-CN.md`，记录 32 字节主密钥通过环境变量注入、活动与历史版本失败关闭、轮换顺序、`LINA-HMAC-SHA256-V1`规范串、固定请求向量、调用报文、接口与资源双重允许策略、一次性`SK`及插件启停数据保留语义。固定正文摘要和签名继续由`TestFixedProtocolVector`与`TestJSONBodyDigestMatchesVector`验证。
- 运行`openspec validate add-aksk-machine-authentication --strict`通过。文档和翻译没有新增运行期依赖、数据库、缓存或数据权限变化；接口文档翻译归属插件自身，不写入宿主运行时语言包。开发工具与跨平台入口无修改，页面翻译和用户可观察结果将在任务 8.1 至 8.3的 E2E及截图中继续验证。

### 任务 8.1 验证与审查

- 按`lina-e2e`技能扫描插件目录后从本地`TC001`连续分配至`TC005`：`TC001-machine-client-crud.ts`覆盖客户端新增、查询、编辑和删除；`TC002-access-key-rotation.ts`覆盖一次性`SK`关闭清理、第二把密钥和旧密钥停用；`TC003-policy-authorization.ts`覆盖授权抽屉仅暴露接口与资源整体读写及未知关系整体拒绝；`TC004-permission-visibility.ts`覆盖只读用户按钮隐藏和后端独立拒绝；`TC005-plugin-lifecycle.ts`覆盖禁用隐藏、重新启用和数据恢复。
- 所有测试资产闭环在`apps/lina-plugins/linapro-auth-aksk/hack/tests/`，页面操作集中于`pages/MachineAccessPage.ts`，API准备与清理集中于`support/aksk.ts`。每个文件只有一个`test.describe`、无跨文件状态依赖，并在自身`afterAll`清理客户端、策略、角色和用户；涉及插件生命周期的用例最终恢复启用状态。
- 前端新增局部`data-testid`用于客户端/策略表单、一次性`AK/SK`和策略授权区域定位，不改变 API或视觉语义。用临时窄范围配置运行`mise exec -- pnpm exec tsc --noEmit -p tsconfig.aksk-e2e.json`通过后已删除配置；页面与测试目录`prettier`通过。
- 项目标准 Playwright发现入口已选择 5 个 AKSK文件并按插件治理串行执行，但在加载共享既有`hack/tests/fixtures/auth-state.ts`时触发`ReferenceError: exports is not defined in ES module scope`，尚未进入任何新增测试。全量`test:validate`同时存在与本变更无关的宿主 E2E治理失败。新增用例的实际运行、截图和该环境问题复核保留在任务 8.3；本任务本身仅创建和静态验证测试资产。

### 任务 8.2 验证与审查

- 扩展宿主`middleware_machine_auth_test.go`中的真实 GoFrame HTTP链路，新增 machine-only的`POST /machine-write`代表性测试路由，验证精确`records.create`与`records/write`同时允许才执行控制器，只有 read资源权限时返回`403`且不进入控制器；现有读路由继续覆盖接口拒绝、资源拒绝和未开放 machine的用户接口默认拒绝。
- `TestSignedMachineRequestProjectionAndReplayEarlyExit`发送完整`LINA-HMAC-SHA256`形态请求，断言宿主传给提供者的 scheme、方法、`X-Lina-Date`、`X-Lina-Nonce`、`X-Lina-Content-SHA256`和服务端正文摘要准确；同一 nonce首次返回`200`，第二次提供者拒绝后返回`401`且控制器总计只执行一次。`TestBearerAuthenticationMiddlewareRegression`继续证明 Bearer请求不进入机器分派器。
- 插件`authprovider`固定向量与 provider测试覆盖有效签名、正文篡改、时间过期、密钥/客户端停用、密钥过期、无策略、签名错误和重放后端失败；`replayguard`竞态测试覆盖单机及跨节点原子单赢家。真实 PostgreSQL测试使用同一 AK装配加密凭证与策略，并覆盖策略、密钥、客户端撤销后的跨节点失效；客户端、密钥和策略服务测试覆盖跨租户不透明拒绝。
- 运行`mise exec -- go test ./internal/service/middleware -run 'Test(MachineAuthenticationMiddlewareChain|SignedMachineRequestProjectionAndReplayEarlyExit|BearerAuthenticationMiddlewareRegression)' -count=1`、`GOWORK=off mise exec -- go test -race ./backend/internal/authprovider ./backend/internal/replayguard -count=1`和带`LINA_TEST_PGSQL_LINK`的四个插件 PostgreSQL包测试均通过。本任务只增加测试路由和测试 stub，不新增生产 API、运行期依赖、缓存、数据库、数据权限或`i18n`影响；开发工具和跨平台入口无修改。

### 任务 8.3 验证与审查

- 使用安装的 Google Chrome 依次独立运行`TC001`至`TC005`均通过，再串行运行整个`apps/lina-plugins/linapro-auth-aksk/hack/tests/e2e`目录，5 个用例在 1.3 分钟内全部通过；用例各自创建和清理数据，生命周期用例最终恢复插件启用状态，证明不存在跨文件状态依赖。
- E2E 实际运行发现并修复密钥创建页把未导入的`<a-modal>`当作全局组件使用的问题，改为显式导入 Ant Design Vue `Modal`；同时按 VXE 固定操作列、Vben Modal/Drawer 实际 DOM和原生`403 Forbidden`契约收敛 POM定位及权限断言。
- 最新截图保存在`temp/20260716/`，覆盖页面初始态、客户端表单、提交结果、一次性凭证、策略授权、策略提交、只读权限和插件禁用态。多模态审查确认中文翻译完整，无原始 i18n key、文本截断、元素重叠或错误反馈；一次性结果只显示 AK，SK保持掩码，关闭后 DOM断言确认明文已清除。
- 当前运行目录没有声明为`machine`的生产路由，授权目录按默认拒绝语义显示空接口和空资源集合；E2E在目录非空时选择首个可用项，目录为空时保存空允许集合并绑定客户端，随后仍以未知 operation/resource验证关系整体拒绝。临时插件 ESM边界文件已删除，未引入长期测试运行时配置。

### 任务 8.4 验证与审查

- 在本机 PostgreSQL `ruyi`库使用`ON_ERROR_STOP`连续两次执行插件原始`001-add-aksk-machine-authentication.sql`，已有对象均以`IF NOT EXISTS`安全跳过；随后执行原始卸载 SQL，六张插件表全部删除且查询计数为 0，再执行安装 SQL完整恢复。
- 重装后静态目录为 6 张表和 21 个索引。事务内验证活动`tenant_id + client_key/name`重复写入触发唯一约束，设置`deleted_at`后相同业务键可以重新创建，事务最终回滚，不保留门禁数据；卸载脚本按外键依赖逆序执行，无残留对象。
- 访问密钥表仅保存公开`access_key`、`BYTEA secret_ciphertext`、`BYTEA secret_nonce`和`master_key_version`，不存在明文 SK列。迁移仍只有 DDL，无 Seed 或 Mock数据；卸载清理数据，普通禁用不执行卸载 SQL且 E2E已证明启停保留数据。
- 临时将插件`hack/config.yaml`连接切到当前开发库后运行`GOWORK=off mise exec -- make dao`成功生成六组 DAO/DO/Entity，再恢复默认模板连接。生成结果确认主键和策略/客户端外键为`int64`、密文与 nonce为`[]byte`、自动时间及软删除字段为`*time.Time`，未手工修改生成文件。

### 任务 8.5 验证与审查

- 宿主使用`mise exec -- go test -p=1`覆盖`authcap`、`bizctxcap`、`pluginbridge/...`、`pluginhost`、`bizctx`、`middleware`、`plugin/...`、`apidoc`、`httpstartup`和`internal/cmd`，包括公共契约、认证分派、插件生命周期、动态桥接、文档投影和启动绑定，全部通过。
- 首次不限制包并行度的宿主矩阵中，`plugin/internal/integration`两个动态插件菜单卸载用例因跨包共享测试状态失败；两用例单独重跑通过，完整矩阵以`-p=1`重跑也通过，确认不是本变更生产行为回归，并记录串行验证边界。
- `linapro-auth-aksk`与`linapro-monitor-operlog`分别运行`GOWORK=off mise exec -- go test ./... -count=1`全包通过。AKSK再运行`-race`覆盖`authprovider`、`replayguard`、`lastused`、`secretcrypto`和`securityevent`，共享快照、重放表、合并更新、明文清理和安全事件状态均无竞态。
- 使用`LINA_TEST_PGSQL_LINK=<current-project-link>`串行运行 AKSK `authprovider`、客户端、访问密钥和策略四个真实 PostgreSQL集成包全部通过，覆盖密文凭证、软删除、唯一约束、事务授权替换、跨租户拒绝、缓存修订和撤销语义。

### 任务 8.6 验证与审查

- 分别运行`make lint dir=apps/lina-core plugins=0`、`make lint dir=apps/lina-plugins/linapro-auth-aksk plugins=1`和`make lint dir=apps/lina-plugins/linapro-monitor-operlog plugins=1`，三个受影响 Go模块的`golangci-lint`与`staticcheck U1000`均为 0 issues。
- lint发现并修复本次新增身份 wire解码函数复杂度超限、LRU缓存未检查`list.Element.Value`类型断言和英式`catalogue`拼写。身份字段 11至 17抽成单用途扩展解码函数且 wire编号与错误语义不变；缓存异常元素现在会从映射和链表同时丢弃并回源重建，正常路径仍为 O(1)。相关单元、插件全包和 route/apidoc/runtime测试重跑通过。
- 使用临时外置源码配置运行`vue-tsc --noEmit --skipLibCheck`覆盖宿主与 AKSK全部 Vue/TS文件通过，随后删除临时配置。仓库`build:antd`仍过滤旧包名`@vben/web-antd`，改用当前包名执行`pnpm run build --filter=@lina/web-antd`，11 个 Turbo任务和`@lina/web-antd`生产 Vite构建全部成功；构建入口元数据问题与本变更无关且未修改开发工具。
- `make i18n.check`通过运行时硬编码、消息和前端键覆盖门禁；宿主`apidoc/i18n`及插件`api/hmacprotocol/config`治理测试通过，所有受影响 i18n JSON均经`jq empty`解析成功。API英文源文本统一使用`catalog`后，中文稳定键完整性测试继续通过。
- 定向扫描受影响生产后端未发现把`secret`、签名、Authorization、密文或 nonce传给 logger、`g.Log`或`fmt.Print`；配置只保存主密钥环境变量名。前端`secretKey`仅存在于一次性响应类型和一次性 Modal显示/复制路径，E2E已证明关闭后清除，未进入列表、持久状态或普通页面。

### 任务 8.7 验证与审查

- `openspec validate add-aksk-machine-authentication --strict`通过；`make plugins.check`扫描 28 个清单、20 个配置和 836 个 Go文件，共 884 个文件且 findings=0。主仓与`apps/lina-plugins`子模块分别运行`git diff --check`均通过。
- DI来源记录覆盖宿主启动期 route authorization catalog、machine coordination和缓存协调共享实例，以及插件 factory一次构造 provider、密钥环、快照、重放、last-used和安全事件服务的 owner、创建位置、传递路径与复用策略；请求路径不创建独立服务图。
- 缓存记录明确数据库权威源、租户修订号一致性、事务后失效、单机/集群分支、最大陈旧窗口、共享`SetNX`重放防护、后端不可用失败关闭和恢复重建；竞态、跨节点撤销、修订故障与恢复测试均有证据。
- 数据权限例外记录明确机器客户端/密钥/策略是租户共享治理资源，不应用本人/部门范围，但 JWT管理权限与平台/租户边界始终先于查询和写入；机器资源 owner仍负责目标资源租户过滤，跨租户不透明拒绝测试通过。
- 接口性能记录覆盖列表批量聚合、策略集合化替换、认证最多六次有界查询、常数时间集合授权、固定容量 LRU、固定窗口`last_used_at`合并和无动态结果集 N+1；SQL记录覆盖幂等、卸载逆序、唯一/检查约束、真实索引和软删除复用。
- 开发工具与跨平台有受控影响：`wasmbuilder`使用既有 Go AST路径提取四类机器路由标签，不新增 shell或平台专属语义；`mise exec -- go test ./hack/tools/linactl/internal/wasmbuilder -count=1`与`mise exec -- make lint dir=hack/tools/linactl plugins=0`通过。`build:antd`旧包名过滤问题作为既有元数据记录，使用当前包名的等价 Turbo入口完成构建。`i18n`、apidoc、敏感扫描、Go测试、E2E与截图证据均已在任务 7.5至 8.6记录，临时 ESM和类型检查配置已删除。

### 任务 8.8 Lina审查报告

**变更：** `add-aksk-machine-authentication`

**范围：** 全部变更。

**审查文件数：** 292。

**范围来源：** 主仓与`apps/lina-plugins`子仓的`git status --short`、`git diff --name-only`、`git diff --cached --name-only`和`git ls-files --others --exclude-standard`，并展开`linapro-auth-aksk`未跟踪目录及所有生成的`DAO/DO/Entity`文件。`linapro-auth-aksk`与`linapro-monitor-operlog`插件根目录均不存在本地`AGENTS.md`，适用项目顶层规范。

**已读取规则文件：** `AGENTS.md`、`.agents/rules/openspec.md`、`.agents/rules/documentation.md`、`.agents/rules/architecture.md`、`.agents/rules/data-permission.md`、`.agents/rules/cache-consistency.md`、`.agents/rules/plugin.md`、`.agents/rules/api-contract.md`、`.agents/rules/backend-go.md`、`.agents/rules/database.md`、`.agents/rules/dev-tooling.md`、`.agents/rules/frontend-ui.md`、`.agents/rules/testing.md`和`.agents/rules/i18n.md`。

#### 发现的问题

- **严重，已修复：** `apps/lina-plugins/linapro-monitor-operlog/backend/internal/service/middleware/middleware_audit_sanitize.go`对截断或损坏 JSON 中带引号的`secretKey`、`Authorization`、签名和加密字段无法可靠脱敏，存在一次性凭证进入审计存储的风险。新增带引号字段容错规则、字段边界和 URL 参数边界，并在`middleware_audit_test.go`覆盖截断 JSON、引号赋值、普通字段误匹配和非敏感查询参数保留。
- **严重，已修复：** `apps/lina-plugins/linapro-auth-aksk/backend/internal/service/accesskey/accesskey_impl.go`的密钥上限采用事务外先计数后插入，并发创建可能突破`maxKeysPerClient`。现改为事务内锁定客户端行、重新计数、生成加密并插入；真实 PostgreSQL 并发测试连续 5 次均保持一个成功、一个稳定超限拒绝。
- **严重，已修复：** `apps/lina-plugins/linapro-auth-aksk/backend/internal/service/policy/policy_impl.go`的每客户端策略上限在事务外校验，不同策略并发绑定可能突破`maxPoliciesPerClient`。现改为授权替换事务内按客户端 ID 稳定排序锁行后计数；真实 PostgreSQL 并发测试连续 5 次均保持关系数不超过上限。
- **警告：** `apps/lina-plugins/linapro-auth-aksk/hack/tests/e2e/TC003-policy-authorization.ts:35`在当前生产机器路由目录为空时只能验证空允许集合提交、客户端绑定和未知关系整体拒绝，不能在浏览器中选择真实`operation`与资源读写项。当前需求未指定首批资源 owner，所有生产接口按设计保持 machine 默认拒绝；非空目录、接口与资源双重允许已由宿主、目录、策略服务和动态路由自动化测试覆盖。首个资源 owner 开放生产接口时必须补充对应非空目录 E2E。

#### 规则域结论

- OpenSpec、文档与架构：通过。用户已明确授权修改`apps/lina-core`；宿主只承载 core-owned 通用主体、路由授权、协调和桥接契约，AKSK业务闭环在官方源码插件。公共中英文 README 与插件中英文 README事实一致。
- 数据权限与 API：通过。管理接口只接受 JWT用户，查询和写入在数据库访问前使用可信平台或租户边界；共享治理资源不应用本人或部门范围的例外已记录。REST方法、DTO标签、毫秒时间戳、主键宽度、分页和批量上限均有测试。
- 缓存与安全：通过。数据库为权威源，事务提交后推进租户修订号，单机与集群共享后端分支、最大陈旧窗口、失败关闭、原子 nonce、防重放、密钥版本和明文生命周期均有并发或集成测试。
- 后端、数据库与性能：通过。DI owner、创建位置、传递路径和共享实例已追溯；生成文件由工具维护。SQL幂等、软删除、唯一约束、索引和卸载顺序已验证；认证、列表和策略装配均为固定次数集合查询，并发规模上限已补强。
- 前端、E2E与`i18n`：通过并保留上述 1 项非阻塞警告。E2E质量审查因页面、权限、插件生命周期和端到端工作流变化而触发；5 个插件 E2E独立及串行通过，关键截图确认中文翻译、布局、反馈和一次性`SK`掩码无异常。英文与中文资源由静态键、API文档完整性和 JSON解析门禁覆盖。
- 开发工具与跨平台：通过。`wasmbuilder`只扩展现有 Go AST结构化提取，Linux、macOS和 Windows 共用同一代码路径；工具测试和模块 lint通过，无 shell或平台专属入口变化。

#### 验证证据

- 宿主认证、`bizctx`、中间件、插件集成、动态运行时、`apidoc`和启动绑定测试：通过。
- `linapro-auth-aksk`、`linapro-monitor-operlog`全包测试及 AKSK安全核心竞态测试：通过。
- AKSK真实 PostgreSQL认证、客户端、密钥、策略及新增并发上限测试：通过。
- `apps/lina-core`、`linapro-auth-aksk`、`linapro-monitor-operlog`与`hack/tools/linactl`对应 lint：通过，0 issues。
- 当前前端类型检查、生产构建、`make i18n.check`、`make plugins.check`、敏感信息扫描、SQL安装卸载与截图审查：通过。
- `openspec validate add-aksk-machine-authentication --strict`、主仓与插件子仓`git diff --check`：通过。

#### 摘要

- 严重：0 个未解决，3 个已修复。
- 警告：1 个非阻塞测试缺口。
- 结论：未发现阻塞问题，变更实施完成；等待用户验收后归档。

## Feedback

- [x] **FB-1**: 缺少`LINA_AKSK_MASTER_KEY_V1`时 AKSK 管理路由注册终止宿主冷启动，应保持宿主与`Bearer JWT`可用并让机器认证及新密钥创建失败关闭
- [x] **FB-2**: PostgreSQL集成测试直接在开发库执行插件安装 SQL，导致 AKSK 表所有者与运行账号不一致，机器访问列表和新建客户端均失败
- [x] **FB-3**: 修复访问密钥创建和空策略授权抽屉运行时错误，新增插件机器专用读写验证接口，并以客户端、访问密钥、策略授权和真实 HMAC 请求覆盖完整链路
- [x] **FB-4**: 修复一次性访问凭证弹窗确认按钮无响应和警告提示未渲染问题，明确提醒保存`AK/SK`并提供可见的`SK`复制按钮，更新`TC002`验证复制和确认关闭后清除明文凭证

### FB-1 验证与审查

- 根因是`linapro-auth-aksk`在注册 JWT 管理路由时调用完整`pluginconfig.Load`解析部署主密钥；默认配置声明`v1 -> LINA_AKSK_MASTER_KEY_V1`，环境变量缺失错误经`registerRoutes`冒泡到`httpstartup.Run`并终止整个宿主。修复前使用未设置主密钥的标准`make dev`稳定复现后端启动即退出。
- 配置层新增只验证非密钥运行参数的管理面加载入口；启动装配优先使用完整密钥环，缺失或无效时注入显式失败关闭的`secretcrypto.Service`并记录告警。JWT 管理路由、客户端及已有非密钥管理操作继续可用；创建访问密钥复用既有`AKSK_ACCESS_KEY_ENCRYPTION_FAILED`业务错误且不落库；HMAC提供者仍使用完整配置并失败关闭，不生成默认密钥或明文兜底。
- DI 来源检查：没有新增外部运行期接口依赖。`plugincap.ConfigService`、`machinecoord.Service`、`bizctxcap.Service`和`routecap.Service`仍由宿主启动期 registrar提供并复用共享实例；降级 cipher由插件启动装配创建后注入唯一 access-key service图，请求路径不创建服务、配置读取器、缓存或协调实例。
- 影响分析：无 API DTO、HTTP方法、路由、数据库结构、查询装配或数据权限边界变化；无缓存权威源、修订号、失效或集群协调语义变化；无前端、API文档源文本、业务错误定义、插件清单或语言包变化，因此`i18n`资源无影响；无开发工具、脚本或跨平台入口修改。E2E质量审查因宿主冷启动工作流回归而触发，采用自动化路由注册测试、真实`make dev`冷启动与 HTTP就绪探测作为等价端到端验证，无页面或视觉变化，浏览器截图不适用。
- 新增回归覆盖缺少主密钥时三组 JWT 管理控制器仍绑定、HMAC provider初始化失败关闭、不可用 cipher拒绝加解密，以及真实 PostgreSQL中创建访问密钥返回稳定错误且凭证行数保持为零。插件全包测试、受影响包`-race`、宿主`internal/cmd`启动绑定测试、官方插件目标 lint和严格 OpenSpec校验全部通过。
- 最终在显式移除`LINA_AKSK_MASTER_KEY_V1`的环境中运行`make dev`通过：Lina Core与 Lina Vben均完成就绪探测，`GET /api.json`返回`200`，JWT管理路由未认证返回`401`，缺密钥 HMAC请求返回`401`；日志记录 AKSK受控降级告警后正常监听`:9120`。`lina-review`反馈级范围审查读取全部命中规则，未发现阻塞问题或新增警告。

### FB-2 验证与审查

- 根因是客户端、访问密钥、策略和认证提供者四组 PostgreSQL集成测试直接把`LINA_TEST_PGSQL_LINK`配置为 GoFrame默认数据库，并在该数据库执行插件安装 SQL。此前测试连接使用`minster`访问开发库`ruyi`，因此六张 AKSK表及其 identity序列由`minster`持有；宿主运行账号`ruyi`没有这些对象的读写权限，`GET`和`POST /x/linapro-auth-aksk/api/v1/machine-clients`均在数据库层返回`permission denied`。路由、JWT权限、Controller和前端页面装配均正常，不是页面吞错或主密钥缺失问题。
- 新增插件内部 PostgreSQL测试支撑：每个集成测试从测试连接创建唯一`linapro_auth_aksk_test_*`数据库，把 GoFrame默认组切换到该临时库，用`SELECT current_database()`确认隔离后才执行安装 SQL；测试结束关闭连接、恢复原配置并使用 PostgreSQL 14支持的`DROP DATABASE ... WITH (FORCE)`清理。四组 setup统一复用该入口，不再各自在开发库执行 DDL。
- 新增单元测试固定 GoFrame PostgreSQL连接串的数据库路径替换、查询参数保留、非 PostgreSQL或空数据库拒绝以及 DDL标识符转义。真实 PostgreSQL测试以具备`CREATEDB`权限的测试账号运行，客户端、访问密钥、策略、认证提供者和 testsupport五个包全部通过；测试结束查询确认临时数据库数量为 0，开发库 AKSK对象 owner未被测试改回。
- 对当前开发库执行一次性运维修复，将六张 AKSK表、六个 identity序列以及同根因污染的操作日志表和序列 owner恢复为运行账号`ruyi`；没有重建数据库、修改表结构或改动业务数据。浏览器实际刷新机器访问页面后列表正常，新建客户端显示“创建成功”并出现在列表，删除显示“删除成功”且测试数据已清理。
- E2E质量审查因页面列表和表单提交属于用户可观察回归而触发。仓库已有`TC001-machine-client-crud.ts`独立运行通过，覆盖新建、API查询、编辑和删除；首次运行因本机 Playwright浏览器缓存缺失未进入用例，改用测试配置原生支持的`E2E_BROWSER_CHANNEL=chrome`后 1 个用例通过。初始页、创建抽屉和提交结果三张自动截图已逐张审查，无错误提示、原始`i18n`键、布局遮挡或数据渲染异常。
- 影响分析：只修改插件测试支撑和既有 PostgreSQL集成测试 setup，无生产 API、HTTP路由、DTO、权限标签、数据权限或租户边界变化；无生产数据库迁移、DAO/DO/Entity、查询路径、索引、软删除和业务数据语义变化；无缓存权威源、失效、修订号或集群协调影响；无运行期 DI、服务实例和宿主能力契约变化；无前端代码、用户文案、API文档源文本、插件清单或语言包变化，因此`i18n`资源无影响；无 Makefile、脚本、开发工具或跨平台入口变化。插件根目录无本地`AGENTS.md`，根目录无`.contributing`，本反馈未修改`apps/lina-core`、`apps/lina-vben`或`hack`。
- 本次按`AGENTS.md`读取并遵守`openspec.md`、`documentation.md`、`architecture.md`、`plugin.md`、`backend-go.md`、`database.md`、`testing.md`、`frontend-ui.md`、`api-contract.md`、`data-permission.md`、`cache-consistency.md`、`dev-tooling.md`和`i18n.md`；其中前端、API、数据权限、缓存和开发工具规则域经审查确认无代码影响，仅记录边界与验证判断。
- 当前工作区重新运行插件全包测试、真实 PostgreSQL五包`-race`、官方插件目标 lint、`openspec validate add-aksk-machine-authentication --strict`和主仓/插件子仓`git diff --check`均通过。`lina-review`首轮发现新增 helper缺少部分注释且相关变量未分组，修正并重跑门禁后未发现阻塞问题；剩余风险仅为集成测试账号必须具备创建和删除临时数据库权限，这是显式测试环境前置条件，不影响宿主运行账号。

### FB-3 验证与审查

- 根因包含三个相互关联的缺口。第一，本地宿主进程未注入`LINA_AKSK_MASTER_KEY_V1`，因此按既定失败关闭语义使用不可用 cipher，新建访问密钥返回`AKSK_ACCESS_KEY_ENCRYPTION_FAILED`且不落库；这不是允许默认密钥或明文兜底的代码缺陷。第二，空策略的`clientIds`使用`append([]int64(nil), ...)`投影，Go JSON把空关系序列化为`null`，前端授权抽屉执行展开运算时触发`clientIds is not iterable`。第三，既有 E2E只有宿主测试代表路由，实际插件授权目录没有机器资源，无法验证管理面创建到真实签名放行的完整链路。
- 策略 DTO映射现在把`operationCodes`、`resources`和`clientIds`全部稳定序列化为数组，并新增 JSON回归测试锁定空关系契约。未修改前端适配层；现有`accessPolicyDetail()`继续解包`detail`，`TC003`证明空策略抽屉可正常打开和提交。
- 插件新增`GET/PUT /x/linapro-auth-aksk/api/v1/machine-test-resource`，分别声明`linapro-auth-aksk.test-resource.read/write`、同一`linapro-auth-aksk.test-resource`资源、`read/write`动作和`actors:"machine"`。资源不访问数据库、不持久化载荷，只投影可信机器主体、客户端标识、脱敏凭证前后缀和租户；不返回完整`AK`、`SK`、签名、`nonce`、密文、主密钥或请求头。
- DI来源检查：验证资源 owner是`linapro-auth-aksk`。`bizctxcap.Service`继续由宿主启动期`HTTPRegistrar.Services().BizCtx()`提供，`registerRoutes`只创建一个`testresource.Service`并注入一个`ControllerV1`，随后与既有控制器绑定到同一认证、租户和权限中间件链；请求路径不调用`New()`，不创建独立服务图、缓存、协调后端或数据库连接。
- 数据权限影响：验证资源没有持久业务数据、列表、详情目标或聚合，不读取其他租户记录；响应租户只来自宿主可信机器上下文，载荷不能覆盖。机器客户端、密钥和策略管理边界保持原有`JWT`管理权限加平台或当前租户共享治理例外。数据库影响：没有 SQL、表、索引、DAO、DO、Entity、软删除或查询路径变更，也不存在`N+1`。
- 缓存一致性影响：没有新增或修改缓存权威源、修订号、失效路径、陈旧窗口或集群协调代码。`TC006`通过既有事务后修订推进验证策略授权立即放行、关系清空后新请求立即`403`，并验证同一`AK + nonce`重放返回`401`。开发工具与跨平台无代码影响；没有修改`Makefile`、脚本、runner或宿主工具。为规避仓库已记录的插件 ESM测试加载问题，执行期临时边界文件和定向类型配置均已删除，未进入交付内容。
- `i18n`有接口文档和稳定错误影响。插件仍按`en-US`源文本和`zh-CN`翻译治理，新增验证资源 DTO的中文`apidoc`映射与机器主体错误英中资源；`en-US/apidoc`保持空对象。中英文 README同步记录固定路由、operation、resource、动作和验收步骤。`make i18n.check`、API文档翻译完整性测试和 JSON解析通过。
- E2E质量审查因访问密钥创建、策略抽屉、权限拒绝和端到端签名工作流变化而触发。新增`TC006-machine-authentication-flow.ts`通过 Node`crypto`按正式协议创建正文摘要、时间戳、随机`nonce`和 HMAC；独立创建并清理客户端、访问密钥和策略，验证 JWT调用机器专用接口`403`、无策略签名请求`403`、读写授权成功、主体与租户投影、重放`401`及策略解绑后`403`。`TC002`和`TC003`回归通过；三张截图确认一次性`SK`遮罩、策略抽屉展示两条 operation与资源整体读写、提交成功且无错误提示、原始 i18n键、布局重叠或敏感信息泄漏。
- 本地服务使用运行期随机生成且未写入文件的 32字节 Base64主密钥重启；真实管理 API创建临时客户端和访问密钥成功，返回`lak_`凭证和一次性`SK`后已清理。最终服务保持运行于`http://127.0.0.1:9120/`和`http://127.0.0.1:5666/`。部署或后续重启仍必须显式注入同一版本化主密钥；没有把开发密钥提交到配置、日志或数据库。
- 验证证据全部基于最终工作区：插件`go test ./...`、变更包`-race`、真实 PostgreSQL客户端/密钥/策略/认证提供者/testsupport五包`-race`、官方插件 lint均通过；AKSK前端定向`vue-tsc`、E2E定向`tsc`和`@lina/web-antd`生产构建通过；`TC002`、`TC003`、`TC006`通过；`make i18n.check`、`make plugins.check`、`openspec validate add-aksk-machine-authentication --strict`及主仓/插件子仓`git diff --check`通过。
- `lina-review`反馈级审查重新读取所有命中规则，发现并修复验证接口完整 AK投影及 README契约缺失后重跑门禁，最终严重问题 0、警告 0，未发现阻塞问题。插件根无本地`AGENTS.md`，根目录无`.contributing`；本反馈只修改`linapro-auth-aksk`与对应 OpenSpec文档，没有新增`apps/lina-core`、`apps/lina-vben`或根`hack`变更。

### FB-4 验证与审查

- 根因是一次性访问凭证组件使用`useVbenModal`时未注册`onConfirm`，底部确认按钮触发后没有调用`modalApi.close()`；同时模板使用未注册的`a-alert`导致保存提醒不渲染，`Input.Password`内置显隐后缀又占用了秘密密钥复制按钮的后缀位置。浏览器实际点击确认后仅获得焦点、弹窗保持打开，控制台同时记录`Failed to resolve component: a-alert`，与用户截图一致。
- 组件现在通过`onConfirm`关闭弹窗，并继续在`onClosed`清空前端凭证引用；警告改用显式导入的`Alert`，中英文文案明确要求立即保存`AK`和`SK`并说明`SK`关闭后无法再次查看或找回。`SK`复制按钮移到密码框外作为独立可见图标按钮，避免与密码显隐控件冲突；`AK`复制按钮也使用显式组件和稳定测试标识。
- `i18n`影响限定在已启用双语的`linapro-auth-aksk`插件运行时 UI，已同步更新`en-US`和`zh-CN`资源；无 API文档源文本、插件清单或宿主语言资源变化。`make i18n.check`通过，`TC002`精确断言中文提醒和“已复制”反馈，不依赖双语正则或原始 key。
- E2E质量审查因一次性凭证弹窗、复制和确认关闭属于用户可观察工作流而触发。既有`TC002-access-key-rotation.ts`改为通过底部确认关闭，不再用右上角关闭按钮绕过原问题，并验证提醒、`SK`复制按钮、复制成功、弹窗关闭、DOM明文清除、列表不回显`SK`以及后续密钥轮换。用例独立创建唯一客户端并在`afterAll`清理，使用`E2E_BROWSER_CHANNEL=chrome`独立运行通过，1个用例、0失败。
- 当前截图`temp/20260720/211312931-aksk-one-time-key-result.png`确认提醒完整、`AK/SK`复制按钮可见、密码显隐与复制按钮无重叠；`temp/20260720/211313755-aksk-one-time-key-confirmed.png`确认结果弹窗已关闭、复制成功反馈可见且页面不含`SK`。未发现原始`i18n`键、文字截断、布局遮挡或敏感信息泄漏。
- 影响分析：无 HTTP API、DTO、后端 Go、数据库、查询性能、数据权限、缓存权威源或失效、运行期 DI、服务构造和开发工具跨平台影响；没有新增`apps/lina-core`、`apps/lina-vben`或根`hack`代码变更。插件根不存在本地`AGENTS.md`，根目录不存在`.contributing`，修复闭环在`linapro-auth-aksk`插件及对应 OpenSpec记录内。
- 最终工作区的`@lina/web-antd`生产构建、AKSK前端定向`vue-tsc`、`TC002`定向 TypeScript检查、`make i18n.check`、`openspec validate add-aksk-machine-authentication --strict`和主仓/插件子仓`git diff --check`均通过。仓库级 E2E TypeScript检查仍有其他插件和宿主既有错误，本次定向配置及真实 E2E均证明新增文件无错误；临时类型检查配置已删除。
- `lina-review`反馈级审查读取`AGENTS.md`、`openspec.md`、`documentation.md`、`plugin.md`、`frontend-ui.md`、`testing.md`和`i18n.md`，审查并补齐确认操作后的截图证据后，严重问题 0、警告 0，未发现阻塞问题。
