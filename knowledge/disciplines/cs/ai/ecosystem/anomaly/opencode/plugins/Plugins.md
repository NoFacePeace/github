# OpenCode 插件

OpenCode 插件通过事件钩子、自定义工具及相关配置扩展宿主行为。插件可以在模型请求前后或工具执行相关阶段接入处理，不只是向模型提供文字说明。

## 1. 加载方式

官方插件文档介绍两种入口：

- 本地 JavaScript 或 TypeScript 文件：项目目录 `.opencode/plugins/`，或全局目录 `~/.config/opencode/plugins/`。
- npm 包：通过 OpenCode 配置的 `plugin` 列表声明。

具体安装命令与插件接口取决于 OpenCode 和插件版本，应同时核对二者说明。

## 2. 工具与钩子

- 工具：模型生成调用请求，插件代码执行操作并返回结果。
- 钩子：运行到指定阶段时，宿主调用插件处理函数。

上下文管理可以结合两者：模型通过工具提交处理选择，插件在后续请求发送前转换消息。

## 3. 社区插件

- [DCP](DCP.md)：通过模型可调用的压缩工具和自动清理管理上下文；它是社区插件，不是 OpenCode 官方内置功能。

## 4. 参考资料

核对日期：2026-10-21。

- [OpenCode 插件文档](https://opencode.ai/docs/plugins/)
- [OpenCode 社区生态](https://opencode.ai/docs/ecosystem/)
