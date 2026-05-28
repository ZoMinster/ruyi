## MODIFIED Requirements

### Requirement: 用户名密码登录

系统 SHALL 支持用户名 + 密码登录，登录请求 MUST 携带受控的用户会话客户端类型`clientType`。允许值仅包括`web`、`mobile`、`desktop`、`cli`；系统 MUST 拒绝缺失或未知的客户端类型。验证成功后系统返回 JWT Token。登录过程中（无论成功或失败）SHALL 发出统一的登录生命周期事件，且事件 MUST 包含登录请求中的`clientType`。登录成功后，SHALL 在宿主维护的`sys_online_session`表中创建会话记录；如果`linapro-monitor-loginlog`已启用，插件根据登录事件完成登录日志入库。

#### Scenario: 登录成功
- **当** 用户向`POST /api/v1/auth/login`提交正确的用户名、密码和`clientType=web`时
- **则** 系统返回 JWT Token，响应格式为`{code: 0, message: "ok", data: {accessToken: "...", refreshToken: "..."}}`
- **且** 宿主在`sys_online_session`表中创建会话记录（包含 token_id、用户信息、IP、浏览器、操作系统、client_type 等）
- **且** 宿主签发的 access token 和 refresh token claims 均包含`clientType=web`
- **且** 宿主发出登录成功事件，事件`clientType`为`web`；如果`linapro-monitor-loginlog`已启用，插件写入登录成功日志

#### Scenario: 登录失败且日志插件缺失
- **当** 用户提交错误凭证和合法`clientType=mobile`
- **且** `linapro-monitor-loginlog`未安装、未启用或初始化失败
- **则** 系统仍返回正确的登录失败结果
- **且** 宿主发出登录失败事件，事件`clientType`为`mobile`
- **且** 宿主不因缺少特定登录日志持久化实现而报错

#### Scenario: 拒绝未知客户端类型
- **WHEN** 用户向`POST /api/v1/auth/login`提交`clientType=plugin`
- **THEN** 系统拒绝登录请求
- **AND** 不签发 JWT
- **AND** 不创建在线会话

### Requirement: 用户退出

系统 SHALL 支持用户退出操作。退出操作 SHALL 发出统一的登录生命周期事件并删除宿主维护的在线会话记录。退出事件 MUST 使用当前认证上下文中的会话`clientType`，不得硬编码或重新推断客户端类型。

#### Scenario: 退出成功
- **当** 已登录用户调用`POST /api/v1/auth/logout`时
- **则** 系统返回成功响应，从`sys_online_session`表中删除该用户的会话记录，前端清除本地存储的 Token
- **且** 宿主发出退出成功事件，事件`clientType`等于该 token claims 和会话记录中的`clientType`
- **且** 如果`linapro-monitor-loginlog`已启用，插件写入对应日志

### Requirement: 认证生命周期事件可供插件订阅

系统 SHALL 将登录成功、登录失败和退出成功等认证生命周期事件作为受控 Hook 发布给已启用的插件。事件 payload 中的`clientType` MUST 来自受控用户会话客户端类型，不得包含`service`、`plugin`或其他非用户客户端主体值。

#### Scenario: 登录成功后发布认证事件
- **当** 用户使用`clientType=desktop`登录成功时
- **则** 宿主向订阅了`auth.login.succeeded`的插件分发事件
- **且** 事件包含宿主暴露的用户身份和客户端上下文
- **且** 事件`clientType`为`desktop`

#### Scenario: 退出成功后发布认证事件
- **当** `clientType=cli`的用户退出成功时
- **则** 宿主向订阅了`auth.logout.succeeded`的插件分发事件
- **且** 事件`clientType`为`cli`
- **且** 事件分发不改变原始退出成功语义

### Requirement: pre_token 必须使用 Redis 单次 TTL 状态

系统 SHALL 在集群模式下使用 Redis 存储`pre_token`、候选租户、用户会话`clientType`和 single-use 消费状态。`pre_token` MUST 短期有效且只能使用一次。

#### Scenario: pre_token 首次选择租户
- **WHEN** 用户使用`clientType=mobile`完成密码验证并获得有效`pre_token`
- **AND** 用户使用该`pre_token`调用 select-tenant
- **THEN** 系统原子消费 Redis 中的`pre_token`
- **AND** 签发正式 JWT
- **AND** 正式 JWT 和在线会话的`clientType`均为`mobile`
- **AND** 后续同一`pre_token`不可再次使用

#### Scenario: pre_token 重放
- **WHEN** 客户端第二次使用同一`pre_token`
- **THEN** 系统拒绝请求
- **AND** 不签发正式 JWT
