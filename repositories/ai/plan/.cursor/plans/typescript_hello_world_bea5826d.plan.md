---
name: TypeScript Hello World
overview: 在空目录中初始化最小 Node.js + TypeScript 项目，添加入口文件并说明如何用 `tsx` 或 `tsc` 运行。
todos:
  - id: init-npm
    content: 在 plan 目录 npm init 并安装 typescript + tsx（或仅 tsc 方案）
    status: completed
  - id: add-config
    content: 添加 tsconfig.json 与 package.json scripts
    status: completed
  - id: add-entry
    content: 创建 src/index.ts 输出 Hello World 并验证运行
    status: completed
isProject: false
---

# TypeScript Hello World 计划

## 目标

在 [repositories/ai/plan](repositories/ai/plan) 下得到可运行的 TypeScript Hello World：`console.log` 输出 `Hello, World!`（或类似文案）。

## 推荐文件结构

```
plan/
  package.json       # 依赖与 npm scripts
  tsconfig.json      # 编译选项（目标 ES2020+、module 与 Node 对齐）
  src/
    index.ts         # 入口：一行或几行 Hello World
```

## 实现要点

1. **package.json**
  - `name`、`version`、`type: "module"`（若用 ESM；也可用 CommonJS，与 `tsconfig` 的 `module` 一致即可）。
  - `devDependencies`：`typescript`，以及任选其一用于直接跑 TS：
    - **tsx**（推荐，零配置、快），或
    - **ts-node**（需与 `module` 配置配合）。
2. **tsconfig.json**
  - `compilerOptions`：`target`（如 `ES2022`）、`module`（`NodeNext` 或 `ESNext` 与 `package.json` 的 `type` 一致）、`strict: true`、`rootDir: "src"`、`outDir: "dist"`（若走先编译再运行）。
3. **src/index.ts**
  - 示例：`console.log("Hello, World!");`
4. **scripts（示例）**
  - 开发：`"dev": "tsx src/index.ts"`（若选 tsx）。
  - 或编译运行：`"build": "tsc"`，`"start": "node dist/index.js"`。

## 执行顺序（确认计划后由你或代理执行）

1. 在 `plan` 下 `npm init -y` 并安装 `typescript` + `tsx`（或仅 `typescript` 若只用 `tsc`）。
2. 添加 `tsconfig.json` 与 `src/index.ts`。
3. 在 `package.json` 中写入 `dev`/`start` 脚本。
4. 运行 `npm run dev`（或 `npm run build && npm start`）验证输出。

## 说明

- 当前工作区该路径下**尚无**现有 `package.json`/`tsconfig`，无需合并旧配置。
- 若你希望**单文件、无构建**（仅一个 `.ts` + 全局 `ts-node`），也可极简，但不利于复现与协作；上述结构是常见最小可维护方案。

