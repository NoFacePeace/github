# h5-main — AI Agent 指南

面向在本仓库中改代码的自动化助手与人类协作者：说明项目是什么、如何跑、约定是什么。

## 项目定位

- **名称**：`h5-main`（package: `h5-main`）
- **角色**：基于 **Vue 3** 的 H5 主应用骨架，使用 **Vite** 构建、**Vue Router** 做路由；预留 **子应用注册** 入口（`registerSubApps.ts` 当前为空实现），便于后续接入微前端（如 qiankun，同 monorepo 下另有相关目录时可对接）。

## 技术栈

| 类别 | 选型 |
|------|------|
| 运行时 | Vue 3（`<script setup>` + Composition API） |
| 路由 | Vue Router 4，`createWebHistory`，`BASE_URL` 来自 `import.meta.env.BASE_URL` |
| 构建 | Vite 8，插件：`@vitejs/plugin-vue`、`vite-plugin-vue-devtools` |
| 语言 | TypeScript；模板校验用 `vue-tsc` |
| Node | `^20.19.0 \|\| >=22.12.0`（见 `package.json` `engines`） |

## 目录与职责（`src/`）

- `main.ts`：创建应用、挂载 `router`、挂载到 `#app`
- `App.vue`：根布局，内含 `<RouterView />`
- `router/index.ts`：路由表；新页面用懒加载 `() => import('@/views/...')`
- `views/`：页面级组件
- `registerSubApps.ts`：子应用注册钩子（待实现时在此集中配置）

路径别名：**`@` → `src/`**（与 `vite.config.ts`、`tsconfig.app.json` 一致）。

## 常用命令

在项目根目录（本文件所在目录）执行：

- `npm run dev` — 本地开发
- `npm run build` — 生产构建
- `npm run preview` — 预览构建产物
- `npm run type-check` — TypeScript / Vue 类型检查

改完 TS/Vue 后建议至少跑 `type-check`，避免 CI 或编辑器才发现问题。

## 编码约定

- 新路由：在 `router/index.ts` 增加 `path` / `name` / 懒加载组件；页面放 `views/`，命名与路由语义一致。
- 导入优先使用别名：`import X from '@/...'`，避免深层 `../../`。
- 样式：页面/组件内优先 `scoped`；全局样式若后续增加，应集中管理并避免污染子应用容器（微前端场景）。
- 微前端：实现 `registerSubApps` 时在 `main.ts` 或启动流程中显式调用，并保持与路由激活规则一致；子应用资源地址、跨域、生命周期需与部署环境对齐后再写死或配置化。

## 协作时注意

- 仅改与任务相关的文件；不要顺带大范围格式化或无关重构。
- 不要提交 `node_modules`、构建产物目录（如 `dist/`）除非项目明确要求。
- 用户界面文案当前为简体中文示例（如首页）；新增用户可见文案保持语言一致，除非产品另有要求。

## 相关路径（monorepo）

本包路径：`repositories/frontend/mini/h5-main`。若仓库内存在 `qiankun`、`single-spa` 等实验或封装目录，对接时以本包 `package.json` 实际依赖与入口为准，避免假设已安装未声明的包。
