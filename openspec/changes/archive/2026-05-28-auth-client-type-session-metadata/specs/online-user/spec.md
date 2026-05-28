## MODIFIED Requirements

### Requirement: 在线用户列表查询

系统 SHALL 在`linapro-monitor-online`已安装并启用时为管理员提供查询当前在线用户的能力，支持按用户名和 IP 地址筛选。在线用户列表 SHALL 接入宿主数据权限治理：全部数据范围可查询所有在线会话；本部门数据范围仅查询当前用户所在部门范围内用户的在线会话；仅本人数据范围仅查询当前用户自己的在线会话。在线会话投影 MUST 包含用户会话客户端类型`clientType`。

#### Scenario: 查询在线用户列表

- **当** `linapro-monitor-online`已安装并启用且管理员请求在线用户列表时
- **则** 插件返回宿主会话投影中的在线会话记录列表
- **且** 每条记录仍包含 token_id、用户名、部门名称、IP、登录地点、浏览器、操作系统、登录时间和`clientType`等治理字段

#### Scenario: 本部门范围限制在线用户列表

- **当** 普通用户角色数据范围为本部门数据
- **且** 查询在线用户列表
- **则** 系统仅返回当前用户可见部门范围内用户的在线会话
- **且** 返回记录中的`clientType`来自宿主会话投影，不通过额外逐项查询补齐

#### Scenario: 仅本人范围限制在线用户列表

- **当** 普通用户角色数据范围为仅本人数据
- **且** 查询在线用户列表
- **则** 系统仅返回当前登录用户自己的在线会话

### Requirement: 在线用户列表必须使用 PostgreSQL 投影并可结合 Redis hot state

系统 SHALL 继续以`sys_online_session`作为在线用户管理查询投影。集群模式下请求热状态存储在 Redis，但在线用户列表的数据权限过滤、分页、搜索和治理字段 MUST 通过 PostgreSQL 投影完成。

#### Scenario: 集群模式查询在线用户列表
- **WHEN** 管理员在集群模式下查询在线用户列表
- **THEN** 系统从`sys_online_session`投影查询可见会话
- **AND** 查询继续接入 tenantcap 和 datascope
- **AND** 返回 token_id、用户名、部门、IP、浏览器、操作系统、clientType、登录时间和最后活跃时间

#### Scenario: Redis hot state 与投影短暂不一致
- **WHEN** Redis session 已过期但 PostgreSQL 投影尚未清理
- **THEN** 清理任务最终删除投影
- **AND** 认证链以 Redis hot state 为请求有效性权威
