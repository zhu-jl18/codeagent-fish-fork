# fish-agent-wrapper

<p align="center">
  <strong>中文</strong> | <a href="README.en.md">English</a>
</p>

这是从 `cexll/myclaude` **fork 并大幅裁剪**的个人版本：聚焦 `/dev` 工作流及其最小依赖组件。

你会得到什么（Key Concepts）：
- `/dev` 工作流：需求澄清 → 计划 → 并行执行 → 验证
- `fish-agent-wrapper`：Go 写的执行器；统一 3 个后端 `codex/claude/gemini`；核心机制 `--parallel`
- `product-requirements` skill：PRD 生成

你不会得到什么：
- upstream 里那套 agent 映射/复杂编排（已刻意移除）

## 安装（WSL2/Linux + macOS + Windows）

最推荐的安装方式：直接复制 `./dist` 的编译产物（安装时不需要 Go）。

```bash
python3 install.py
```

可选参数：
```bash
python3 install.py --install-dir ~/.claude --force
python3 install.py --skip-wrapper
```

安装器会做这些事：
- `CLAUDE.md`：**追加** managed block（非破坏性覆写；`--force` 刷新 managed block）
- `commands/dev.md`
- `agents/dev-plan-generator.md`
- `skills/fish-agent-wrapper/SKILL.md`
- `skills/product-requirements/SKILL.md`
- `~/.claude/fish-agent-wrapper/*-prompt.md`：每个后端一个空占位文件（用于 prompt 注入）
- `~/.claude/bin/fish-agent-wrapper`（Windows 上是 `.exe`）

提示：
- 在 WSL 里运行 `install.py` 会安装 Linux wrapper；在 macOS（Apple Silicon）里运行会安装 Darwin arm64 wrapper；在 Windows 里运行会安装 Windows `.exe`。
- 如果你使用了非默认目录，请设置 `FISH_AGENT_WRAPPER_CLAUDE_DIR` 指向你的目录。

## 维护（重新编译 dist 二进制）

```bash
bash scripts/build-dist.sh
```

产物：
- `dist/fish-agent-wrapper-linux-amd64`
- `dist/fish-agent-wrapper-darwin-arm64`
- `dist/fish-agent-wrapper-windows-amd64.exe`

## Prompt 注入（默认开启；空文件 = 等价不注入）

默认占位文件（每个后端一个）：
- `~/.claude/fish-agent-wrapper/codex-prompt.md`
- `~/.claude/fish-agent-wrapper/claude-prompt.md`
- `~/.claude/fish-agent-wrapper/gemini-prompt.md`

规则：
- wrapper 会读取对应后端的 prompt 文件；只有在内容非空时才会 prepend 到任务前面
- 文件不存在 / 只有空白字符：等价“无注入”

常用环境变量：
- `FISH_AGENT_WRAPPER_CLAUDE_DIR`：Claude 配置根目录（默认 `~/.claude`）

运行时配置（审批/绕过、超时、并行传播规则）详见：
- `docs/runtime-config.md`

## 使用

在 Claude Code 里：
```text
/dev "实现 X"
```

PRD：
```text
/product-requirements "为功能 X 写 PRD"
```

## 开发/测试

```bash
cd fish-agent-wrapper
go test ./...
```
