## Why

LinaPro现有受保护接口只接受`Bearer JWT`并以登录用户、角色菜单权限和用户数据范围作为授权主体，无法让外部系统以独立机器身份访问明确开放的非用户级资源。项目需要在不伪造用户会话、不复用第三方云厂商凭证语义的前提下，提供可审计、可撤销、可按接口和资源类型授权的自有`AK/SK`访问能力。

## What Changes

- 在`apps/lina-core`新增通用认证主体和认证提供者契约，使`user`与`machine`主体可以通过同一宿主认证入口进入后续授权链，同时保持现有`JWT`用户认证语义。
- 为可受机器访问的接口增加稳定`operation`、`resource`、`action`和`actors`元数据；未显式声明`machine`的接口默认拒绝`AK/SK`访问。
- 新增`linapro-auth-aksk`官方源码插件，闭环管理机器客户端、访问密钥和机器访问策略，并通过宿主通用认证接缝提供`LINA-HMAC-SHA256`认证。
- 将机器授权拆分为精确接口白名单和资源类型级`read`、`write`权限；请求只有在接口与资源权限同时满足时才允许继续执行。
- 支持一个机器客户端持有多个访问密钥，以便进行不中断轮换；`SK`只在创建时返回一次，服务端使用版本化主密钥加密保存，禁止在列表、详情、日志或审计响应中回显。
- 增加时间窗口、随机`nonce`和请求体摘要校验以防止签名重放；密钥停用、删除、过期或策略失效后按失败关闭语义拒绝请求。
- 增加机器认证和策略快照的单机、集群缓存一致性治理，复用现有协调基础设施完成事务后跨节点失效和分布式`nonce`去重。
- 扩展`OpenAPI`安全声明和操作元数据，使文档能够准确展示接口支持`Bearer JWT`、`LINA-HMAC-SHA256`或两者之一。
- 扩展统一审计上下文以记录机器客户端、访问密钥标识、接口和资源类型，并补齐请求、响应、头部和错误链路中的凭证脱敏。
- 提供与现有工作台一致的机器客户端、密钥和策略管理页面；管理接口仅允许`JWT`管理员访问，禁止机器身份修改自身密钥或策略。

## Capabilities

### New Capabilities

- `machine-authentication-contract`: 定义宿主通用认证主体、认证提供者注册、认证方案分派、接口与资源双重授权、机器访问默认拒绝和请求上下文传播契约。
- `linapro-auth-aksk`: 定义机器客户端、访问密钥、访问策略、`HMAC`签名、安全存储、管理接口、管理页面、租户边界和插件生命周期行为。

### Modified Capabilities

- `plugin-host-domain-capabilities`: 增加机器认证提供者接缝、机器主体`CapabilityContext`传播和源码插件提供认证实现的边界要求。
- `distributed-cache-coordination`: 增加机器认证策略快照修订号、跨节点失效、分布式`nonce`去重和协调后端故障时的失败关闭要求。
- `system-api-docs`: 增加`LINA-HMAC-SHA256`安全方案、接口认证主体和机器授权元数据的文档投影要求。
- `oper-log`: 增加机器主体审计字段以及`AK`、`SK`、签名、授权头和一次性密钥响应的脱敏要求。

## Impact

- 宿主后端：`apps/lina-core/internal/service/auth`、`apps/lina-core/internal/service/middleware`、`apps/lina-core/internal/service/bizctx`、`apps/lina-core/internal/service/apidoc`、启动装配和对应测试。
- 插件公共契约：`apps/lina-core/pkg/plugin/capability/authcap`、`apps/lina-core/pkg/plugin/capability/bizctxcap`、`apps/lina-core/pkg/plugin/pluginhost`及源码插件认证提供者注册实现；需要同步审查`apps/lina-core/pkg/plugin`下的`README`文档。
- 新增插件：`apps/lina-plugins/linapro-auth-aksk`，采用`source`、`managed`、`tenant_aware`和中英文`i18n`资源，插件业务表使用完整`plugin_linapro_auth_aksk_`前缀。
- 数据与权限：机器策略不复用用户角色菜单授权；资源权限仅控制资源类型整体`read`或`write`，不增加资源`ID`、路径模式或属性条件。机器客户端固定属于平台或单一租户，原有资源租户隔离继续生效。
- 接口性能：请求认证使用按`AK`唯一索引查询或版本化快照读取，策略在单次请求内以集合结构完成常数时间判断，不允许按权限项或资源项循环查询数据库；`last_used_at`采用合并写入，避免每次请求同步更新。
- 安全与配置：新增版本化主密钥配置和轮换治理，优先使用Go标准库密码组件并复用现有协调、缓存和锁能力，不引入独立认证服务。
- 首批机器资源：不开放既有生产业务接口；由`linapro-auth-aksk`作为资源 owner 提供无持久业务副作用的机器专用读写验证资源，用于验证客户端、访问密钥、策略、签名认证和资源读写授权完整链路。所有其他现有接口继续默认拒绝机器主体，后续仍由各资源 owner 为目标接口声明完整机器元数据并接入租户边界。
- 前端与测试：新增插件管理页面、后端单元和集成测试、管理工作流`E2E`、签名协议测试、重放与失效测试、单机和集群缓存一致性测试，以及关键页面截图审查。
- `i18n`：插件启用`en-US`与`zh-CN`，新增菜单、按钮、表单、提示、错误码和插件接口文档翻译资源。
- 开发工具与脚本：扩展`hack/tools/linactl/internal/wasmbuilder`的既有 Go AST 路由提取逻辑，把`operation`、`resource`、`action`和`actors`写入动态插件`RouteContract`；不新增 shell、平台命令或平台专属路径语义，Linux、macOS和 Windows 复用同一 Go 实现。
