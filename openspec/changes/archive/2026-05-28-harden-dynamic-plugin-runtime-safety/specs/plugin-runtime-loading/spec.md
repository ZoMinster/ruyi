## ADDED Requirements

### Requirement: 动态插件 WASM 执行必须具备宿主兜底资源边界

系统 SHALL 为动态插件 WASM bridge 执行提供宿主侧默认超时和内存上限。调用方上下文已经包含 deadline 时，系统 MUST 尊重调用方 deadline；调用方未提供 deadline 时，系统 MUST 使用 bridge 默认超时。动态插件 HTTP 路由、cron discovery、cron job 和生命周期回调等所有宿主执行入口 MUST 经过同一资源边界。

#### Scenario: 无调用方 deadline 时使用默认超时
- **WHEN** 动态插件 WASM route handler 执行时调用方 context 不包含 deadline
- **THEN** 宿主 bridge 为本次执行设置默认超时
- **AND** guest 无限循环或长时间不返回时本次执行被取消

#### Scenario: 调用方 deadline 更严格时不放宽
- **WHEN** 动态插件生命周期回调执行时调用方 context 已包含更短 deadline
- **THEN** 宿主 bridge 使用调用方 deadline
- **AND** 不用默认超时延长本次执行窗口

#### Scenario: WASM 内存分配超过上限
- **WHEN** 动态插件在 WASM 执行中请求超过宿主配置或默认内存上限的内存
- **THEN** 宿主拒绝或终止本次 WASM 执行
- **AND** 调用方收到资源耗尽或等价失败诊断

### Requirement: 动态插件协调器必须恢复 stale reconciling 状态

系统 SHALL 检测并恢复动态插件中过期的 `reconciling` 瞬态状态。仅当 `CurrentState=reconciling` 且状态更新时间超过配置或默认阈值时，系统 MAY 将其恢复为由权威安装状态、启用状态和 active release 推导出的稳定状态，并继续后续协调；阈值内的 `reconciling` 状态 MUST 保持不变。

#### Scenario: 过期 reconciling 被恢复
- **WHEN** 动态插件 P 的 `CurrentState` 为 `reconciling`
- **AND** 该状态更新时间超过 stale 阈值
- **THEN** 协调器将 P 恢复到可推导的稳定状态或失败诊断状态
- **AND** 后续协调 tick 可以继续收敛 P

#### Scenario: 活跃 reconciling 不被重置
- **WHEN** 动态插件 P 的 `CurrentState` 为 `reconciling`
- **AND** 该状态更新时间未超过 stale 阈值
- **THEN** 协调器不得重置 P 的当前状态
- **AND** 当前 tick 不得并发执行 P 的生命周期副作用

### Requirement: 动态插件协调器 tick panic 必须被隔离

系统 SHALL 在动态插件协调器 tick 边界恢复 panic。单次 tick 内的 panic MUST 被记录为运行时诊断，并且 MUST NOT 终止后续协调循环。

#### Scenario: 单次 tick panic 后继续运行
- **WHEN** 动态插件协调器在一次 tick 中发生 panic
- **THEN** 系统恢复该 panic 并记录诊断
- **AND** 协调器 goroutine 继续等待并执行后续 tick

#### Scenario: panic 后瞬态状态可继续恢复
- **WHEN** 协调器 panic 发生前插件 P 已进入 `reconciling`
- **AND** P 的 `reconciling` 状态随后超过 stale 阈值
- **THEN** 后续 tick 按 stale `reconciling` 恢复规则处理 P

### Requirement: 动态插件协调器必须按插件串行化共享副作用

系统 SHALL 对动态插件协调器中会修改共享状态的生命周期副作用按插件 ID 串行化。共享副作用包括生命周期 SQL、迁移账本、菜单和权限治理资源同步、active release 切换、frontend bundle 切换以及 runtime revision 发布。

#### Scenario: 同一插件不会并发执行生命周期副作用
- **WHEN** 两个协调触发同时尝试收敛动态插件 P
- **THEN** 系统只允许一个执行方进入 P 的共享生命周期副作用
- **AND** 另一个执行方跳过或等待后续协调机会

#### Scenario: 不同插件可独立收敛
- **WHEN** 动态插件 P 和 Q 同时需要收敛
- **THEN** P 的 per-plugin 互斥不得阻塞 Q 的独立收敛
- **AND** 系统可以在各自锁边界内分别处理 P 和 Q
