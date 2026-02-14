---
name: code-council
description: Multi-backend parallel code review using code-dispatcher. Runs 2-3 AI reviewers simultaneously, then the host agent verifies findings and synthesizes a final report.
compatibility: Requires code-dispatcher skill with at least 2 backends configured (codex, claude, gemini).
---

# code-council: Multi-Backend Code Review

## Overview

Leverage code-dispatcher's multi-backend parallel execution to run multiple AI reviewers simultaneously, each performing an independent comprehensive code review. The **host agent then verifies each finding and synthesizes the final report** — ensuring only real issues are presented to the user.

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

`@` references are resolved by code-dispatcher at execution time — the host agent passes them through as-is, no pre-processing needed.

### Step 2: Determine Backends

**Do NOT ask the user** unless the invocation is genuinely ambiguous.

Resolution order:
1. User explicitly specified backends in the invocation → use those
2. Neither specified → default to **all three** (codex, claude, gemini)

### Step 3: Build & Execute Parallel Review

Construct a single `code-dispatcher --parallel` invocation. All backends receive the **same review prompt**.

#### Review Prompt (single source of truth)

```text
You are a code reviewer. Perform a comprehensive review of the target code.

Focus on:
1. Security vulnerabilities (injection, auth bypass, data exposure)
2. Logic errors and bugs
3. Performance issues
4. Error handling completeness
5. Code quality and maintainability
6. Potential edge cases

Target: [@ file references or inline diff]

Output format:
## Code Review Findings
For each finding:
- **[CRITICAL|WARNING|INFO]**: one-line summary
  - Location: file:line
  - Detail: what is wrong and why it matters
  - Suggestion: concrete fix
```

#### Command Template

Generate one `---TASK---` block per selected backend. Each block uses `id: review_<backend>`, sets `backend: <backend>`, and pastes the review prompt above into `---CONTENT---`.

```bash
code-dispatcher --parallel --backend codex <<'EOF'
---TASK---
id: review_codex
backend: codex
---CONTENT---
<review prompt with target filled in>

---TASK---
id: review_claude
backend: claude
---CONTENT---
<review prompt with target filled in>
EOF
```

For 3 backends, add a third `---TASK---` block for gemini. For 1 backend, use a single block.

### Step 4: Host Agent Verification & Synthesis (MANDATORY — DO NOT SKIP)

After code-dispatcher returns, the host agent MUST verify findings and synthesize the final report.

Each code-dispatcher task is a full AI agent with complete filesystem access — it can read any file, run git commands, and explore the entire project. What they lack is the **user's conversation context** — the host agent knows what the user has been discussing, their priorities, and their intent.

#### Verification Rules

1. **Read all review outputs** from each backend
2. **Deduplicate**: merge findings that point to the same issue
3. **Confidence scoring via agreement**:
   - Finding reported by ≥2 backends → **high confidence**, include by default
   - Finding reported by only 1 backend → **needs verification** (see below)
4. **Verify single-source findings**:
   - Read the actual code at the reported location
   - Quote the specific line(s) that confirm or refute the finding
   - If you cannot point to concrete code evidence, mark as **unconfirmed** and include with a caveat
5. **CRITICAL findings require trace**:
   - For any CRITICAL finding, trace the execution path or data flow that leads to the issue
   - If the trace does not hold, downgrade to WARNING or discard
6. **Add missed findings**: if the host agent spots issues that no backend reported, add them with a note
7. **Rank**: CRITICAL → WARNING → INFO
8. **Group by file**
9. **Output final report** titled `## Code Council — Verified Review`
   - Summary: total counts by severity, agreement matrix (which backends flagged each finding)
   - For each finding: severity, location, detail, suggestion, and which backends reported it

### Step 5: Offer Follow-up Actions

After presenting, offer concrete next steps:
- **Fix critical issues**: generate fix tasks and run them through code-dispatcher
- **Review more files**: restart the workflow on another target
- **Save report**: write to a file if the user wants a record

## Severity Definitions

- **CRITICAL**: Will cause bugs, security vulnerabilities, data loss, or crashes in production
- **WARNING**: Code smell, maintainability risk, or potential future bug under reasonable conditions
- **INFO**: Style suggestion, minor improvement, or optimization opportunity

## Edge Cases

**Single file, trivial size (<30 lines)**:
- Still run the review — small code can have critical issues
- Reviewers will naturally produce fewer findings

**User provides a git diff instead of files**:
- Capture the diff content, pass it inline in `---CONTENT---` instead of `@` file references
- Adjust the review prompt to focus on "changes" rather than "code"

**Backend failure or timeout**:
- If ≥1 backend returns results, proceed with available outputs. Note the failure in the final report.
- If all backends fail, report the failure to the user and suggest retrying or switching backends.
- Do NOT silently drop missing results — always disclose which backends succeeded and which did not.

## Critical Rules

1. **MUST perform Step 4 (verification & synthesis)** — this is the core value proposition
2. **MUST use code-dispatcher skill** — this skill depends on code-dispatcher's `--parallel` execution
3. **MUST be called by host agent only** — this skill is NOT designed for sub-agent use
4. **NEVER modify source code in this skill** — code-council is read-only; fixes go through a separate action
5. **NEVER ask backend selection questions** when the user's intent is clear — resolve from invocation or use defaults

## Example Invocations

**Review specific files**:
```
Review @src/auth/login.ts and @src/auth/session.ts using code-council
```

**Review recent changes (PR-style)**:
```
Run code-council on the uncommitted changes
```

**Use specific backends**:
```
Run code-council on @src/core/ using only codex and claude
```
