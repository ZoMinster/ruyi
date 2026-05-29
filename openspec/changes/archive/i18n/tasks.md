# Tasks

## Summary

- [x] 交付 i18n 基础设施：三层模型、文件资源单一事实来源、语言解析、默认双语基线、固定`ltr`、运行时翻译包和语言列表接口。
- [x] 交付运行时性能治理：翻译缓存按语言和扇区分层，显式作用域失效，`ETag`协商，单值翻译热路径避免整包克隆，前端持久缓存和后台校验。
- [x] 交付服务边界：`LocaleResolver`、`Translator`、`BundleProvider`、`Maintainer`小接口，共享`ResourceLoader`，`WASM`section 读取收敛到`pluginbridge`。
- [x] 交付消息治理：`bizerr`结构化错误、消息分类、导入导出本地化、插件桥接错误契约、硬编码中文扫描和前端`messageKey`优先渲染。
- [x] 交付工作台与文档治理：首次语言识别、语言切换刷新、英文布局回归、项目定位统一、README 中英文镜像规则。
- [x] 交叉影响已迁移：配置、菜单、字典、调度、系统 API 文档、系统信息、工作台、登录页、数据库初始化和 demo-control 的完整契约由对应 owner 分组或`openspec/specs`承载。
- [x] 验证：缺失翻译检查、单元测试、基准测试、E2E、硬编码文案扫描、OpenSpec 校验、相关 README 与治理规则更新和`lina-review`均已作为归档维护证据保留。
