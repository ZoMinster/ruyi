# Tasks

## Summary

- [x] 交付跨平台`linactl`主入口、Windows`make.cmd`、Makefile 薄包装、环境检查/初始化、服务启停、构建、镜像、Wasm、测试、init/mock 和 GoFrame 代码生成命令。
- [x] 整合独立工具模块到`linactl/internal/`，保留公开命令稳定，移除默认路径中对旧`image-builder`、`build-wasm`、`runtime-i18n`和本地`gf`的依赖。
- [x] 建立 Agent 多资源桥接、月度 OpenSpec 自动归档、release/nightly 发布治理、受控 tag 创建、source-plugin 升级、安装脚本、demo/test Compose 和`lina-perf-audit`手动审计技能。
- [x] 反馈闭环：升级治理、数据库启动、安装脚本、性能审计、跨平台命令、工具整合、环境命令、镜像发布、release 版本治理和 monthly archive 共处理`FB-*`系列问题；关键修复包括 source-plugin effective/discovered version 分离、SQL 全量重放、安装脚本目录规范、CI 复用模板、OpenSpec 自动化 fail-fast、AI 工具运行时凭据注入、E2E/CI 稳定性和 Agent 命令显示收敛。
- [x] 验证：覆盖 Go 单元测试、命令 smoke、Windows 基本命令、YAML/shell 语法、Docker/Redis/CI 等价验证、安装脚本 smoke、性能审计 dry-run、OpenSpec 校验、diff 空白检查和`lina-review`审查。
- [x] 治理：本历史分组主要影响开发工具、CI、文档、OpenSpec 和 Agent skill；运行时代码、HTTP API、数据库 schema、权限、数据权限、插件运行时、前端 UI、运行时 i18n 和业务缓存影响由对应 owner 变更承载。本次压缩不修改运行时资源。
