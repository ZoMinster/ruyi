## ADDED Requirements

### Requirement: 动态插件协调器集群模式必须使用 per-plugin 分布式锁

系统 SHALL 在 `cluster.enabled=true` 时，使用现有 distributed locker 或 coordination lock 为动态插件协调器提供 per-plugin 互斥。锁名称 MUST 包含稳定插件 ID，并且同一插件的生命周期 SQL、迁移账本、治理资源同步、发布状态切换和 runtime revision 发布 MUST 只由持锁节点执行。未获得锁的节点 MUST 跳过当前插件并等待后续 revision、event 或 safety sweep。

#### Scenario: 集群节点竞争同一插件协调锁
- **WHEN** 集群节点 A 和 B 同时尝试协调动态插件 P
- **THEN** 只有成功获取 P 的 per-plugin 分布式锁的节点执行 P 的共享生命周期副作用
- **AND** 未获得锁的节点不执行 P 的 SQL、菜单同步或发布状态写入

#### Scenario: 锁名称按插件隔离
- **WHEN** 节点同时协调动态插件 P 和 Q
- **THEN** P 的协调锁名称与 Q 的协调锁名称不同
- **AND** P 的锁竞争不得阻止 Q 在独立锁下收敛

#### Scenario: 单机模式不强制依赖 coordination lock
- **WHEN** `cluster.enabled=false`
- **THEN** 动态插件协调器可以使用进程内互斥或单机锁分支保护本节点并发
- **AND** 系统不得要求 Redis coordination lock 存在
