## ADDED Requirements

### Requirement: 宿主插件系统私有实现必须收敛到插件服务边界

系统 SHALL 将宿主插件系统的私有实现组织在`apps/lina-core/internal/service/plugin`边界下，并将插件公共契约、SDK、bridge 和 capability 能力继续组织在`apps/lina-core/pkg/plugin`边界下。除明确服务多个宿主领域且不属于插件语义的共享基础组件外，插件 catalog、runtime、host service adapter、runtime cache、lifecycle、integration、frontend、openapi、WASM host service、管理投影和插件治理实现 MUST 不作为`internal/service`根层级的独立 service 组件长期存在。

#### Scenario: 开发者查找宿主插件实现
- **WHEN** 开发者需要理解宿主插件 catalog、runtime、host service adapter、runtime cache、lifecycle 或管理投影实现
- **THEN** 相关私有实现位于`apps/lina-core/internal/service/plugin`及其子目录下
- **AND** 开发者不需要在`internal/service/pluginhostservices`、`internal/service/pluginruntimecache`等平行根组件中继续查找插件系统核心逻辑

#### Scenario: 公共插件契约仍归属 pkg plugin
- **WHEN** 源码插件、动态插件或构建工具需要访问插件公共契约、guest SDK、bridge 协议或 capability 服务接口
- **THEN** 这些稳定契约继续通过`apps/lina-core/pkg/plugin`体系暴露
- **AND** 宿主私有实现不得迁入`pkg/plugin`公共边界

### Requirement: plugin 根包必须作为宿主插件服务 facade

系统 SHALL 将`apps/lina-core/internal/service/plugin`根包维护为宿主内部稳定 facade。根包 MUST 保留`Service`契约、公开投影类型、启动期构造入口、轻量编排和必要适配；具体实现逻辑 MUST 优先下沉到职责明确的同包文件或`plugin/internal/<subcomponent>`子组件。根包不得继续积累可独立测试、跨文件共享状态、缓存协调、插件桥接、运行时升级或管理投影等多职责实现。

#### Scenario: 启动装配构造插件宿主服务
- **WHEN** `internal/cmd`需要构造插件服务、源码插件宿主能力目录或 WASM host service 配置
- **THEN** 它通过`internal/service/plugin`根包的稳定 facade 完成
- **AND** 它不得直接导入`plugin/internal/<subcomponent>`实现包

#### Scenario: 新增插件内部职责
- **WHEN** 变更新增或迁移插件管理列表、运行时升级、平台治理、启动一致性、host service adapter 或 runtime cache 等实现
- **THEN** 实现必须归入职责明确的`plugin/internal/<subcomponent>`或`plugin/runtimecache`等目标子组件
- **AND** 子组件命名必须体现领域职责，不得使用`util`、`common`或`helper`等兜底名称

#### Scenario: 根包保留核心跨组件编排
- **WHEN** 某段逻辑跨越 catalog、runtime、integration、lifecycle、cache 和 i18n 等多个插件子组件
- **THEN** 只有在下沉会引入更多转发接口或循环依赖风险时，该逻辑才可暂留`plugin`根包
- **AND** 审查必须记录暂留原因和后续可继续收敛的判断
