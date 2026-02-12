---
name: code-council
description: Multi-backend parallel code review using code-router. Runs 2-3 AI reviewers simultaneously with the same comprehensive review prompt, then the host agent verifies findings and synthesizes a final report.
compatibility: Requires code-router skill with at least 2 backends configured (codex, claude, gemini).
---

# code-council: Multi-Backend Code Review

## Overview

Leverage code-router's multi-backend parallel execution to run multiple AI reviewers simultaneously, each performing an independent comprehensive code review. The **host agent then verifies each finding and synthesizes the final report** — ensuring only real issues are presented to the user.

## When to Use

- User explicitly requests code review, architecture audit, or quality check
- PR review or pre-merge quality gate
- Module-level or file-level review
- A skill or command definition declares a dependency on this skill

## Workflow

### Step 1: Determine Review Scope

Resolve target files from user input:

- **Explicit references**: user provides file/directory paths → use `@` syntax
- **Git diff**: user mentions PR, diff, or recent changes → capture diff output, pass as context
- **Module**: user names a package or directory → glob relevant source files

If scope is ambiguous, ask the user to clarify.

### Step 2: Select Backends

Ask the user which backends to enable for parallel review:

**Options**:
- codex (recommended)
- claude (recommended)
- gemini (recommended)

**Multi-select**: Allow user to pick one or more backends

**Default**: All three enabled

**Fallback** (if AskUserQuestion is not available in your environment):
- Stop and ask the user explicitly: "Which backends do you want to use for review? Options: codex, claude, gemini. Default: all three."
- Wait for user response before proceeding

### Step 3: Build & Execute Parallel Review

Construct a single `code-router --parallel` invocation. All backends run the **same comprehensive review prompt**.

**Template (3 backends)**:

```bash
code-router --parallel --backend codex <<'EOF'
---TASK---
id: review_codex
backend: codex
---CONTENT---
You are a code reviewer. Perform a comprehensive review of the target code.

Focus on:
1. Security vulnerabilities (injection, auth bypass, data exposure)
2. Logic errors and bugs
3. Performance issues
4. Error handling completeness
5. Code quality and maintainability
6. Potential edge cases

Target: [@ file references]

Output format:
## Code Review Findings
For each finding:
- **[CRITICAL|WARNING|INFO]**: one-line summary
  - Location: file:line
  - Detail: what is wrong and why it matters
  - Suggestion: concrete fix

---TASK---
id: review_claude
backend: claude
---CONTENT---
You are a code reviewer. Perform a comprehensive review of the target code.

Focus on:
1. Security vulnerabilities (injection, auth bypass, data exposure)
2. Logic errors and bugs
3. Performance issues
4. Error handling completeness
5. Code quality and maintainability
6. Potential edge cases

Target: [@ file references]

Output format:
## Code Review Findings
For each finding:
- **[CRITICAL|WARNING|INFO]**: one-line summary
  - Location: file:line
  - Detail: what is wrong and why it matters
  - Suggestion: concrete fix

---TASK---
id: review_gemini
backend: gemini
---CONTENT---
You are a code reviewer. Perform a comprehensive review of the target code.

Focus on:
1. Security vulnerabilities (injection, auth bypass, data exposure)
2. Logic errors and bugs
3. Performance issues
4. Error handling completeness
5. Code quality and maintainability
6. Potential edge cases

Target: [@ file references]

Output format:
## Code Review Findings
For each finding:
- **[CRITICAL|WARNING|INFO]**: one-line summary
  - Location: file:line
  - Detail: what is wrong and why it matters
  - Suggestion: concrete fix
EOF
```

**Template (2 backends)**: Omit one `---TASK---` block

**Template (1 backend)**: Use single `---TASK---` block

### Step 4: Host Agent Verification & Synthesis (MANDATORY — DO NOT SKIP)

After code-router returns, the host agent MUST verify findings and synthesize the final report.

Note: each code-router task is a full AI agent with complete filesystem access — it can read any file, run git commands, and explore the entire project. Tasks are NOT "blind" to the project. What they lack is the **user's conversation context** — the host agent knows what the user has been discussing, their priorities, and their intent.

**Why this step exists**:
- **User context**: The host agent carries the ongoing conversation — it knows what the user cares about, prior discussions, and implicit priorities
- **Verification**: Backends may produce false positives or miss real issues — the host agent must verify each finding
- **Interactive ability**: The host agent can ask the user follow-up questions; code-router tasks are fire-and-forget
- **Quality gate**: Raw outputs should be verified and synthesized, not dumped as-is

**What the host agent does**:

1. **Read all review outputs** from each backend
2. **Deduplicate**: merge findings that point to the same issue, note which backends agree
3. **Verify each finding**:
   - Re-examine the code to confirm the issue is real
   - If a finding is false positive, mark it and explain why
   - If backends missed something, add your own findings
4. **Rank by severity**: CRITICAL first, then WARNING, then INFO
5. **Group by file**
6. **Output final report** titled "## Code Council — Verified Review"
   - Include a summary: total counts by severity
   - Note which backends reported each finding

### Step 5: Offer Follow-up Actions

After presenting, offer concrete next steps:
- **Fix critical issues**: generate fix tasks and run them through code-router
- **Review more files**: restart the workflow on another target
- **Save report**: write to a file if the user wants a record

## Severity Definitions

- **CRITICAL**: Will cause bugs, security vulnerabilities, data loss, or crashes in production
- **WARNING**: Code smell, maintainability risk, or potential future bug under reasonable conditions
- **INFO**: Style suggestion, minor improvement, or optimization opportunity

## Edge Cases

**Single file, trivial size (<30 lines)**:
- Still run the full council — small code can have critical issues
- Reviewers will naturally produce fewer findings

**User provides a git diff instead of files**:
- Capture the diff content, pass it inline in the `---CONTENT---` sections instead of `@` file references
- Adjust reviewer prompts to focus on "changes" rather than "code"

## Critical Rules

1. **MUST perform Step 4 (verification & synthesis)** — this is the core value proposition
2. **MUST use code-router skill** — this skill depends on code-router's `--parallel` execution
3. **MUST be called by host agent only** — this skill is NOT designed for sub-agent use
4. **NEVER modify source code in this skill** — code-council is read-only; fixes go through a separate action

## Example Invocations

**Review specific files**:
```
Review @src/auth/login.ts and @src/auth/session.ts using code-council
```

**Review recent changes (PR-style)**:
```
Run code-council on the uncommitted changes
```

**Audit a module**:
```
code-council audit @src/payments/
```

**Use specific backends**:
```
Run code-council on @src/core/ using only codex and claude
```
