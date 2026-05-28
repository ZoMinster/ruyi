## MODIFIED Requirements

### Requirement: 会话热状态必须保留租户和 token 维度

Redis session key SHALL 至少包含 tenant ID 和 token ID。会话 payload SHALL 包含 user ID、username、tenant ID、clientType、login time、last active time 和必要客户端上下文。`clientType` MUST 与 JWT claims 和 PostgreSQL 会话投影一致。

#### Scenario: 同一用户多个租户会话隔离
- **WHEN** 用户 U 同时持有租户 A token 和租户 B token
- **THEN** Redis 中存在两个不同 session hot key
- **AND** 每个 session hot state 都包含该 token 签发时的`clientType`
- **AND** 删除租户 A token 不影响租户 B token

#### Scenario: token 租户不匹配
- **WHEN** JWT claims 中`TenantId=1001`
- **AND** Redis session hot key 仅存在于`tenant=1002`
- **THEN** 认证链拒绝请求

#### Scenario: clientType 在热状态中保持一致
- **WHEN** 用户以`clientType=desktop`登录并访问受保护 API
- **THEN** Redis session hot state 中的`clientType`为`desktop`
- **AND** 认证中间件注入的业务上下文`clientType`也为`desktop`
