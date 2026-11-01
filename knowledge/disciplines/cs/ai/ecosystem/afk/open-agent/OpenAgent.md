# Open-Agent

Open-Agent 是 AFK 官方 GitHub 组织下的开源 Agent 系统，支持通过聊天提交任务，由多模型、多 Agent 协作完成工作，并可自行部署和修改。本文依据官方 README 及仓库元数据整理，核对日期为 2026-11-01。

## 1. 项目定位

项目将自身定位为 Claude Agent SDK、ChatGPT Agents 和 Manus 的开源替代方案。这是项目对目标范围的描述，不代表已验证与这些产品功能等价，也不能据此将其视为 Claude Agent SDK 的兼容实现。

它提供可使用、部署和修改的 Agent 系统，也可作为开发专用 Agent 的基础。

## 2. 核心能力

| 能力 | 官方说明 |
|---|---|
| 多模型接入 | 集成 OpenAI、Claude、Gemini 和开源模型 |
| 多 Agent 协作 | 通过多 Agent 框架共同完成用户任务 |
| 规划与决策辅助 | 通过规格与上下文组织规划、评估和选项展示，由用户保留最终决策 |
| 自托管与定制 | 支持部署自己的实例，并修改源码构建特定用途的 Agent |

这些为官方宣称的能力；具体工具覆盖、模型兼容性和实际任务效果需要结合代码与部署环境确认。

## 3. 开源与部署

项目采用 Apache-2.0 许可证。官方 README 的部署流程为：复制 `.docker/config.example.json` 和 `.docker/docker-compose.yml` 到部署目录，将配置命名为 `config.json`，填写模型 API Key，再运行 `docker compose up -d`。

仓库默认分支为 `master`。开源与可自托管不意味着外部模型调用免费，所需凭据和模型费用由使用的服务决定。

## 4. 参考资料

- [官方仓库](https://github.com/AFK-surf/open-agent)
- [官方 README](https://github.com/AFK-surf/open-agent/blob/master/README.md)
- [项目网站](https://open-agent.io/)
- [许可证](https://github.com/AFK-surf/open-agent/blob/master/LICENSE)
