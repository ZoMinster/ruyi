## ADDED Requirements

### Requirement: OpenAPI必须发布LINA-HMAC安全方案

系统 SHALL 在宿主和已启用插件的`OpenAPI`文档中发布稳定`LinaHMAC`安全方案，说明认证头、时间戳、`nonce`、请求体摘要和签名版本。文档示例 MUST 使用不可用占位值，不得包含真实`AK`、`SK`或签名。

#### Scenario: 查看机器认证安全方案

- **WHEN** 管理员打开接口文档并查看安全方案
- **THEN** 文档同时展示现有`BearerAuth`和新增`LinaHMAC`
- **AND** `LinaHMAC`说明签名所需请求头和协议版本

#### Scenario: 接口文档本地化

- **WHEN** 管理员以`en-US`或`zh-CN`请求启用`i18n`插件的接口文档
- **THEN** 机器认证说明和接口元数据使用对应语言资源
- **AND** 文档不显示内部翻译键或明文凭证

### Requirement: 接口安全声明必须匹配允许主体

系统 SHALL 根据接口`actors`元数据生成操作级安全声明。只允许用户的接口使用`BearerAuth`，只允许机器的接口使用`LinaHMAC`，同时允许二者的接口使用可选认证方案语义；公开接口不得被误标为需要机器凭证。

#### Scenario: 用户和机器均可访问

- **WHEN** 接口声明`actors:"user,machine"`
- **THEN** `OpenAPI`操作安全声明表示`BearerAuth`或`LinaHMAC`任一方案可用
- **AND** 不错误要求调用方同时提供两种凭证

#### Scenario: 未开放机器访问

- **WHEN** 受保护接口没有声明允许`machine`
- **THEN** 文档只显示现有用户认证方案
- **AND** 不暗示`AK/SK`可以访问该接口

### Requirement: OpenAPI必须投影机器授权元数据

系统 SHALL 为允许机器访问的接口投影稳定`operation`、`resource`和`read/write`动作，以便接口文档、策略管理目录和治理测试使用同一事实源。

#### Scenario: 查看机器接口详情

- **WHEN** 接口声明完整机器授权元数据
- **THEN** `OpenAPI`操作包含对应稳定扩展字段或等价结构化投影
- **AND** 投影值与运行时授权目录一致

#### Scenario: 动态插件机器接口

- **WHEN** 已启用动态插件声明允许机器访问的路由
- **THEN** 宿主合并该路由的认证主体和授权元数据到统一文档
- **AND** 动态路由不因共享分发入口而丢失自身操作码和资源码

