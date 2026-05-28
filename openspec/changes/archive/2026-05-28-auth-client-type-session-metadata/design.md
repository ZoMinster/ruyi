## Context

认证 Hook payload 已经公开`clientType`字段，但当前实现把它当作局部默认值处理：登录失败、登录成功、退出成功都写死`web`，插件分发层也会在空值时补`web`。这会把“默认 Web 工作台入口”和“用户会话客户端类型”混在一起。

本项目没有历史兼容负担，因此本变更采用破坏式收敛：后端认证内核不再猜测客户端类型，所有正式用户会话必须从登录入口带入受控枚举，并沿会话生命周期传递。

## Decisions

### D1：`ClientType`只表示用户会话客户端

`ClientType`的底层 JWT claim 契约由`pkg/authtoken`统一维护，`auth`服务复用该命名类型并把非法值包装为认证业务错误。取值仅包括：

| 值 | 语义 |
|----|------|
| `web` | 浏览器 Web 工作台或 Web 应用 |
| `mobile` | 移动端 App 或移动端专用登录入口 |
| `desktop` | 桌面端客户端 |
| `cli` | 人通过命令行客户端登录形成的用户会话 |

不包含`service`或`plugin`。服务调用、插件调用、后台任务、机器令牌等不是用户客户端类型，未来如需表达应使用`actorType`、`grantType`或单独的 host-service 授权主体模型。

### D2：登录入口显式传入并校验

`LoginReq.clientType`为必填字段。服务层通过`ParseClientType`校验，只接受上述受控枚举；未知值返回结构化认证参数错误。动态插件路由等 JWT 旁路解析路径使用同一`pkg/authtoken`契约校验 claims 中的`clientType`。默认 Web 前端在 API 适配层固定传入`web`，而不是让后端隐式兜底。

该字段是认证生命周期事实的一部分，必须写入：

- `LoginInput.ClientType`
- `preTokenRecord.ClientType`
- JWT `Claims.ClientType`
- `session.Session.ClientType`
- Redis session hot state payload
- `sys_online_session.client_type`
- `model.Context.ClientType`
- auth lifecycle Hook payload
- dynamic plugin route user context

### D3：二阶段租户选择继承登录来源

当登录用户需要选择租户时，首次密码校验不会签发正式 token。此时`clientType`写入`pre_token`记录；`IssueTenantToken`消费`pre_token`后使用同一个`clientType`签发正式 token 和在线会话。

租户切换使用当前 JWT claims 中的`ClientType`继承旧会话来源，不允许切换租户时改变客户端类型。

### D4：退出 Hook 使用会话事实源

`Logout`改为接收`LogoutInput`，其中`ClientType`来自认证中间件注入的`bizCtx`。为保证撤销前后信息一致，控制器在调用`Logout`前只使用当前请求上下文的会话信息；`Logout`不再读取 HTTP 参数或硬编码`web`。

### D5：impersonation 继承调用方用户会话来源

impersonation token 是平台管理员在当前登录会话中的代理行为，不是`plugin`或`service`客户端。`IssueImpersonationToken`优先从`bizCtx.ClientType`继承调用方会话来源；若调用方没有认证上下文则拒绝，避免生成缺少来源的用户会话。

### D6：插件契约只增加会话投影字段

插件 Hook payload 既有`clientType`字段保持不变，只改变其事实来源。插件 session contract 增加`ClientType`字符串字段，使`linapro-monitor-online`等插件可以展示在线会话来源。该字段不泄露宿主 DAO/Entity，只是稳定投影。

## Data Model

更新宿主在线会话建表源`apps/lina-core/manifest/sql/004-online-session.sql`：

```sql
CREATE TABLE IF NOT EXISTS sys_online_session (
    ...
    "client_type" VARCHAR(32) NOT NULL,
    ...
);

COMMENT ON COLUMN sys_online_session."client_type" IS 'User session client type: web, mobile, desktop, cli';
```

本项目不考虑旧数据兼容，因此不保留独立追加迁移、不设置默认值，也不做历史回填；初始化新库时建表 SQL 直接保持字段必填。

索引判断：本次不新增按`client_type`筛选或排序的查询路径，暂不创建索引，避免低价值索引。未来若在线用户列表需要按客户端类型筛选，再按实际查询路径增加组合索引。

## Performance

`clientType`随已有登录、会话创建、会话读取和列表投影一起写入或读取，不引入额外数据库查询。在线用户列表仍使用单次分页查询和既有数据权限过滤；新增字段只是同一行投影，不产生`N+1`或前端瀑布式调用。

Redis session hot state 在已有 payload 中增加一个字符串字段，不改变 key 维度、TTL、失效或跨实例一致性机制。

## Validation

- `openspec validate auth-client-type-session-metadata --strict`
- `cd apps/lina-core && go test ./internal/service/auth ./internal/service/session ./internal/service/plugin/internal/hostservices -count=1`
- 涉及 API DTO 或路由绑定后运行`cd apps/lina-core && go test ./internal/cmd -count=1`
- 涉及默认 Web 前端 API 适配后运行`cd apps/lina-vben && pnpm -F @lina/web-antd typecheck`
- 静态检索确认认证实现中不再出现`ClientType: "web"`硬编码

## Impact Notes

- `i18n`：修改宿主 API 文档源文本，需同步宿主`zh-CN` apidoc 翻译资源；默认前端无新增用户可见文案。
- 缓存一致性：涉及 Redis session hot state payload 扩展，权威来源仍为登录签发时的认证输入和会话状态；不新增缓存域或失效路径。
- 数据权限：不新增数据读取或写操作边界；在线用户列表继续通过宿主 session contract 和既有 tenant/data-scope 过滤。
- 开发工具：仅运行既有`make dao`生成流程，不修改工具脚本或跨平台入口。
