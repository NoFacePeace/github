# DSH

DeepSeek Harness（DSH，命令名为 `dsh`）是 DeepSeek 开源的智能体运行框架（Agent Harness），用于把模型、工具和执行流程组织起来，让智能体完成编程及自动化任务。

## 1. 核心设计

DSH 采用“一切皆插件”（Everything is a Plugin）的架构，通过插件组合与扩展能力，底层使用 Cordis。模型提供理解、推理与生成能力，运行框架负责组织模型和工具之间的交互。

以修复代码为例，智能体可以在工具及权限允许的范围内读取项目文件、修改代码、执行命令，再依据执行结果继续处理。

## 2. 本地运行

安装符合项目要求的 Node.js 后，可以通过 npm 启动：

```bash
npx @deepseek-ai/dsh web
```

该命令启动 Web 界面，本地启动时默认打开浏览器，默认访问地址为 `http://127.0.0.1:3080`。模型接入和其他配置以官方文档为准。

## 3. 项目状态

截至 2026-10-25，官方将项目标记为开发者预览版（Developer Preview），仍在快速迭代，更新可能包含不兼容改动。安装要求、命令和插件行为应以官方仓库的当前说明为准。

## 4. 参考资料

- [官方 GitHub 仓库](https://github.com/deepseek-ai/deepseek-harness)
- [官方文档](https://deepseek-harness.github.io/deepseek-harness/)
