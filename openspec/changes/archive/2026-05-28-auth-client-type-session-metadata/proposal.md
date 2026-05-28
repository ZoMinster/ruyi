## Why

当前认证生命周期事件中的`clientType`仅在登录和退出 Hook 组装处硬编码为`web`，没有进入登录输入、`pre_token`、JWT、在线会话、请求上下文或插件会话投影。这会让审计、登录日志和未来多客户端入口无法可靠区分真实用户会话来源。

## What Changes

- **BREAKING**：`POST /api/v1/auth/login`请求必须显式提交`clientType`，仅允许`web`、`mobile`、`desktop`、`cli`。
- **BREAKING**：认证服务`Login`、`Logout`、租户令牌签发、租户切换、刷新和 impersonation 会统一携带会话`clientType`，不再在认证内核中隐式补齐`web`。
- 将`clientType`持久化到`sys_online_session`和 Redis session hot state，并写入 JWT claims、`pre_token`和`bizCtx`。
- 认证 Hook 的`clientType`从会话事实源读取；登录成功、登录失败、退出成功事件都发布真实客户端类型。
- 默认 Web 工作台前端登录请求显式传入`clientType=web`。
- 插件可通过宿主发布的 session contract 读取在线会话`clientType`，在线用户插件可展示该字段。

## Capabilities

### New Capabilities

- 无。本变更收紧既有认证和在线会话能力，不新增独立产品能力。

### Modified Capabilities

- `user-auth`：登录、退出、租户令牌、刷新和认证生命周期 Hook 必须保留真实用户客户端类型。
- `session-hot-state`：Redis session hot state payload 必须保存`clientType`。
- `online-user`：在线会话投影和插件会话 contract 必须包含`clientType`。

## Impact

- 后端 API：`apps/lina-core/api/auth/v1/auth_login.go`新增必填请求字段和 API 文档翻译。
- 后端认证：`apps/lina-core/internal/service/auth/**`升级`ClientType`命名类型、JWT claims、`pre_token`、会话创建和 Hook 组装。
- 会话存储：`apps/lina-core/internal/service/session/**`、`sys_online_session`生成模型和 Redis hot state 需要新增`clientType`字段。
- 数据库：更新宿主`manifest/sql/004-online-session.sql`建表源，向`sys_online_session`添加`client_type`列；按项目生成流程重新生成 DAO/DO/Entity。
- 插件契约：`pkg/plugin/capability/contract.Session`和 hostservices session adapter 增加`ClientType`投影；`linapro-monitor-online`在线用户列表同步展示字段。
- 前端：默认 Web 登录 API 调用显式传入`clientType=web`，不新增可见交互。
- 测试：补充认证生命周期、租户选择、刷新、退出 Hook 和会话存储单元测试；运行相关 Go 编译门禁、前端类型检查和 OpenSpec 校验。
