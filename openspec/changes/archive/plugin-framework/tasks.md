# Tasks

## Summary

- [x] 建立插件框架核心能力：统一`plugin.yaml`、源码/动态插件生命周期、动态`WASM`运行时、host service、能力目录、插件 UI、菜单权限、启动引导、依赖、升级、工作区和包边界。
- [x] 修复和收敛关键演进：源码插件自动安装后同步启动快照；旧`Can*`/guard 生命周期替换为`Before*`/`After*`；动态生命周期由构建期自动发现；插件 API 与资产路径分别收敛到`/x/{plugin-id}/api/v1/...`和`/x-assets/{plugin-id}/{version}/...`。
- [x] 治理：插件公共契约收敛到`pkg/plugin`；宿主私有实现收敛到`internal/service/plugin`；能力目录不暴露`DAO`、`DO`、`Entity`、`*gdb.Model`、`*ghttp.Request`、写入路径或数据权限注入能力。
- [x] 性能：插件列表查询保持只读；完整治理读模型可预热并按插件生命周期、动态产物和租户供应策略显式失效；列表装配复用快照并避免逐插件重复扫描。
- [x] 一致性：插件运行时、frontend bundle、runtime i18n、WASM、manifest 资源视图和默认配置视图按插件和资源作用域失效；集群模式通过 coordination revision/event 和 per-plugin 锁收敛。
- [x] 测试与验证：历史实现覆盖后端单元测试、前端类型检查、插件管理与动态插件`E2E`、host-only/plugin-full 构建测试、`WASM`构建、静态扫描、OpenSpec 校验和发布链路验证。
- [x] 反馈：原分散归档中的反馈、根因、修复和验证结论已按能力主题并入设计与规范；未保留普通执行流水和逐文件 checklist。
- [x] 交叉影响：集群、分布式锁、菜单、角色、认证、监控、发布、`E2E`、DI、宿主边界和系统 API 文档等非插件框架 owner 能力已迁移为`design.md`交叉影响摘要。
