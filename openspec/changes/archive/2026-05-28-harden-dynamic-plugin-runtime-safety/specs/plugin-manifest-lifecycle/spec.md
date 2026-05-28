## ADDED Requirements

### Requirement: 插件生命周期 SQL 与迁移账本必须事务一致

系统 SHALL 在 PostgreSQL 默认运行环境下，将插件 install、upgrade、uninstall 和 rollback 生命周期 SQL 文件执行与对应 `sys_plugin_migration` 账本记录放入同一事务边界。任一 SQL 文件转译、语句执行或账本写入失败时，系统 MUST 回滚本次生命周期 SQL 和账本写入。

#### Scenario: 生命周期 SQL 中途失败时回滚账本
- **WHEN** 插件 P 安装期间执行多个 `manifest/sql/*.sql` 文件
- **AND** 其中一个 SQL 语句执行失败
- **THEN** 系统回滚本次安装生命周期中已执行的 SQL 语句
- **AND** 系统不得写入表示该失败 SQL 文件已成功完成的 `sys_plugin_migration` 记录

#### Scenario: 迁移账本写入失败时回滚 SQL
- **WHEN** 插件 P 升级期间 SQL 文件已经执行成功
- **AND** 对应 `sys_plugin_migration` 账本写入失败
- **THEN** 系统回滚本次升级生命周期 SQL
- **AND** 插件 P 不得进入升级成功状态

#### Scenario: rollback SQL 使用相同事务语义
- **WHEN** 插件 P 的 rollback SQL 被执行
- **THEN** rollback SQL 文件执行和 rollback 方向迁移账本写入在同一事务中完成
- **AND** 任一步失败都会回滚本次 rollback SQL 与账本写入

### Requirement: 插件生命周期 rollback 失败必须进入权威诊断

系统 SHALL 将插件生命周期失败后的 rollback 失败纳入权威失败诊断。rollback SQL、菜单恢复、前端资源恢复、权限治理恢复或发布状态恢复失败时，系统 MUST 保留原始失败原因和 rollback 失败原因；不得只写 warning 日志后返回原始错误。

#### Scenario: rollback SQL 失败被返回给调用方
- **WHEN** 动态插件 P 安装失败后执行 rollback SQL
- **AND** rollback SQL 执行失败
- **THEN** 系统返回或记录同时包含安装原始失败和 rollback SQL 失败的诊断
- **AND** 插件 P 的运行时状态标记为失败或需要人工处理

#### Scenario: 治理资源恢复失败被记录
- **WHEN** 动态插件 P 升级失败后系统尝试恢复菜单、前端资源或权限治理资源
- **AND** 任一恢复动作失败
- **THEN** 系统将恢复失败写入插件发布、节点状态或 registry 的失败诊断
- **AND** 后续管理或协调流程可以读取该失败原因
