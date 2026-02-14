# code-dispatcher

<p align="center">
  <strong>中文</strong> | <a href="README.en.md">English</a>
</p>

> 接收任务 → 选后端 → 构建参数 → 分发执行 → 收集结果。这就是 dispatch。

原始灵感以及部分代码来源 `cexll/myclaude`，特此感谢。

你会得到什么（Key Concepts）：
- `dev` commands or skill：需求澄清 → 计划 →选择后端 → 并行执行 → 验证
- `code-dispatcher` executor and skill：Go 写的执行器；统一 3 个后端 `codex/claude/gemini`；核心机制 `--parallel` && `--resume`;配套使用guide(给ai看的)
- `code-council` skill：多视角并行代码评审（2-3 个 AI reviewer 并行 + host agent 终审）

## 后端定位（仅推荐，可自由指定）

- `codex`：复杂逻辑、bug修复、优化重构
- `claude`：快速任务、review、补充分析
- `gemini`：前端 UI/UX 原型、样式和交互细化
- 调用入口约束：后端都只通过 `code-dispatcher` 调用；不要直接调用 `codex` / `claude` / `gemini` 命令。


## 安装（WSL2/Linux + macOS + Windows）

默认安装方式：从 GitHub Release 的 `latest` 标签下载当前平台二进制（安装时不需要 Go）。

```bash
python3 install.py
```

可选参数：
```bash
python3 install.py --install-dir ~/.code-dispatcher --force
python3 install.py --skip-dispatcher
python3 install.py --repo zhu-jl18/code-router --release-tag latest
```

安装器会做这些事：
- `~/.code-dispatcher/.env`：运行时唯一配置源
- `~/.code-dispatcher/prompts/*-prompt.md`：每个后端一个空占位文件（用于 prompt 注入）
- `~/.code-dispatcher/bin/code-dispatcher`（Windows 上是 `.exe`）

不会自动做的事（必须手动）：
- 不会自动复制 `skills/` / `dev-workflow/commands` / `dev-workflow/agents` 到你的目标 CLI root 或 project scope
- 需要按你的目标 CLI 自行手动复制：
  - **Skills**：从本仓库 `skills/*` 里挑需要的（例如 `skills/dev`、`skills/code-dispatcher`、`skills/code-council`）
  - **/dev command（Claude Code 等）**：使用 `dev-workflow/commands/dev.md` 与 `dev-workflow/agents/*`

提示：
- 在 WSL 里运行 `install.py` 会安装 Linux 二进制；在 macOS（Apple Silicon）里运行会安装 Darwin arm64 二进制；在 Windows 里运行会安装 Windows `.exe`。
- 需要网络访问 GitHub Release；如只想更新配置文件，使用 `--skip-dispatcher`。

## 本地构建（可选）

```bash
bash scripts/build-dist.sh
```

本地构建产物（默认不提交到 git）：
- `dist/code-dispatcher-linux-amd64`
- `dist/code-dispatcher-darwin-arm64`
- `dist/code-dispatcher-windows-amd64.exe`

## Prompt 注入（默认开启；空文件 = 等价不注入）

默认占位文件（每个后端一个）：
- `~/.code-dispatcher/prompts/codex-prompt.md`
- `~/.code-dispatcher/prompts/claude-prompt.md`
- `~/.code-dispatcher/prompts/gemini-prompt.md`

规则：
- code-dispatcher 会读取对应后端的 prompt 文件；只有在内容非空时才会 prepend 到任务前面
- 文件不存在 / 只有空白字符：等价“无注入”

运行时配置（审批/绕过、超时、并行传播规则）详见：
- `docs/runtime-config.md`

## 使用

在 Claude Code 里：
```text
/dev "实现 X"
```

代码评审：
```text
Review @src/auth/ using code-council
```

## 开发/测试

```bash
cd code-dispatcher
go test ./...
```
