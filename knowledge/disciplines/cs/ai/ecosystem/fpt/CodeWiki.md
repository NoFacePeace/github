# CodeWiki

CodeWiki 是 FSoft-AI4Code 的开源代码库文档生成框架，使用静态分析和 AI Agent，将源码组织成模块文档、整体概述和 Mermaid 图，采用 MIT 许可证。

## 1. 定位

CodeWiki 面向整个代码仓库的理解与文档生成，适合新成员熟悉项目、梳理模块关系，以及随代码变更维护技术文档。这里指的是 FSoft-AI4Code/CodeWiki，与 Google 的 Code Wiki 是独立项目。

## 2. 实现流程

1. 分析源码，提取代码组件和依赖关系，构建依赖图。
2. 将依赖图递归划分为层次化模块树。
3. 由 Agent 读取模块源码并生成页面，再根据子模块页面生成上层概述。
4. 输出 Markdown、Mermaid 图、模块树和生成元数据。
5. 增量更新时比较已保存的图与当前代码，更新受影响模块及相关上层页面；变更过大时回退到全量生成。

2.0 还将构建文件、CI、容器配置、包清单和 schema 等技术文件纳入分析。静态依赖与生成说明不能直接证明运行时行为，关键结论仍需核对源码和实际运行证据。

## 3. 使用方式

项目以 Python 实现，支持 CLI 生成文档，也提供 MCP 模式供 AI IDE 使用其分析工具链。具体运行条件、模型接入和参数以所选版本文档为准。

```bash
cd /path/to/project
codewiki generate
codewiki generate --update
```

首次使用前需安装依赖并配置模型接入。默认产物写入 `docs/`，包含概述、模块页面、模块树和元数据。

## 4. 参考资料

- [CodeWiki 源码与使用说明](https://github.com/FSoft-AI4Code/CodeWiki)。
- [项目主页与生成示例](https://fsoft-ai4code.github.io/CodeWiki/)。
- [CodeWiki 论文](https://aclanthology.org/2026.findings-acl.288/)。

本文依据 2026-10-23 查阅的上游 README 整理；上游能力与下游移植版本应分别核实。
