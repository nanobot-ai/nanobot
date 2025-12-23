# Design: Config-Driven Agent System

## Summary
This document defines a configuration-driven agent system: users define **Agents**, **Skills**, and **Tasks** as **Markdown files with YAML front matter**. The runtime loads these definitions, validates schemas, and executes tasks via a tool runtime with safety gates, logging, and replay.

The core idea: **configuration is the product**. Code is a stable execution engine; behavior lives in versioned Markdown.

---

## Goals
- **Config-first**: all agents/skills/tasks defined in Markdown + YAML front matter.
- **Composable**: tasks call skills; skills can call tools; agents provide defaults.
- **Deterministic-ish execution**: stepwise execution with recorded inputs/outputs, retries, and replay.
- **Safety and control**: allowlists, approvals, and scoped permissions.
- **Portable**: a project/repo can ship its own agent packs.
- **Inspectable**: everything human-readable; runs are auditable.

## Non-goals
- Full general AI autonomy without user oversight.
- A fully self-modifying agent that rewrites its own core policies.
- Perfect determinism (LLMs are stochastic); instead we provide *replayability* and *bounded execution*.

---

## Concepts

### Agent
An **Agent** is an executable persona/configuration:
- model + decoding settings
- default system instructions
- allowed tools and permission scopes
- default skills available
- safety policy (approvals, restricted operations)

Think: “who is running the work”.

### Skill
A **Skill** is a reusable capability:
- purpose + instructions
- a defined **input/output contract**
- optional tool requirements
- optional evaluation rubric

Think: “how to do a type of work”.

### Task
A **Task** is a runnable unit of work:
- objective + inputs
- acceptance criteria
- may specify a target agent
- execution plan: either *free-form* (agent plans) or *structured steps* calling skills

Think: “what to do right now”.

### Run
A **Run** is a concrete execution of a task:
- immutable record: task version, agent version, skill versions
- step log with tool calls and outputs
- artifacts (files, patches, generated docs)

---

## Configuration Format

All definitions are Markdown with YAML front matter.

### File layout (recommended)

All configs live under a workspace directory `.nanobot/`.

- Agents: `.nanobot/agents/**/AGENT.md`
- Skills: `.nanobot/skills/**/SKILL.md`
- Tasks: `.nanobot/tasks/**/TASK.md`

Where the **ID** is the relative directory path from the collection root.
Examples:
- `.nanobot/agents/builder/AGENT.md` → agentId = `builder`
- `.nanobot/agents/foo/bar/baz/AGENT.md` → agentId = `foo/bar/baz`

### Common front matter fields

Because IDs are derived from directory names, entities do **not** include `id` or `type` fields in front matter.

All entity types should support:
- `name` (string)
- `description` (string)
- `metadata` (object; free-form)

### Markdown body
The Markdown body is treated as the **instructional canonical source**:
- for agents: behavioral instructions, style, priorities
- for skills: how to perform the capability
- for tasks: details, constraints, context, acceptance criteria

The runtime may attach the body as part of prompts.

---

## Schemas

### Agent schema (front matter)

**Identity**
- `agentId` is derived from the relative directory path under `.nanobot/agents/`.
  - Example: `.nanobot/agents/foo/bar/baz/AGENT.md` → `agentId = "foo/bar/baz"`

Required:
- `name`

Suggested fields:
- `metadata` (object; free-form)
- `model` (e.g., `claude-sonnet`, `gpt-4.1`, etc.)
- `allowed_models`: list of additional models the user may switch to when chatting with this agent.
  - The UI should present these as choices alongside `model`.
  - The runtime MUST reject any model not in `{model} ∪ {allowed_models}`.
- `temperature`, `max_tokens`
- `tools`: list of tools the agent is allowed to call.
  - If omitted, defaults to `inherit`.
  - If set to `inherit`, the agent inherits tools from the host environment defaults.
  - If set to a list:
    - if the list contains `inherit`, the agent inherits tools **and** adds the additional tools from the list.
    - otherwise it is an explicit allowlist (no inherited tools).
  - Built-in tools are referenced by a logical name, e.g. `Read`, `Write`.
  - External tools are referenced as `${mcpServer}/${toolName}`.
- `tool_approvals`: tool approval policy.
  - **Default behavior**: *all tool calls require approval*.
  - Implemented as an ordered list of "allow" rules; the first matching rule wins.
  - Rules can optionally constrain **tool arguments** (a `Map<string, any>`).
- `skills`: list of skill IDs available by default (directory names)
  - If omitted, defaults to `inherit`.
  - If set to `inherit`, the agent inherits skills from the host environment defaults.
  - If set to a list:
    - if the list contains `inherit`, the agent inherits skills **and** adds the additional skills from the list.
    - otherwise it is an explicit allowlist (no inherited skills).
- `tasks`: list of task IDs available to the agent (directory paths under `.nanobot/tasks/`).
  - Semantics match `tools`/`skills` (including inheritance and explicit allowlists).
- `task_approvals`: task approval policy (same structure as `tool_approvals`).
  - **Default behavior**: *all task invocations require approval*.

#### Tool approval rules

`tool_approvals` structure:
- `default`: currently only `approve` is supported (everything requires approval unless allowed)
- `rules`: ordered list of allow rules

Rule shape:
- `tool`: string (must match an entry in `tools`)
- `allow`: boolean
- `when`: optional per-argument matchers

Supported matchers (by value type):
- `equals: <scalar>`
- `in: <array>`
- `startsWith: <string>`
- `matches: <regex>`
- `contains: <scalar>`
- `containsAll: <array>`
- `anyOf: <array of matchers>`
- `allOf: <array of matchers>`

If a key in `when` is missing from the tool args, that rule does not match.

### Skill schema (front matter)

This should follow the Agent Skills specification (https://agentskills.io/specification) closely.

**Identity**
- `skillId` is derived from the relative directory path under `.nanobot/skills/`.
  - Example: `.nanobot/skills/foo/bar/SKILL.md` → `skillId = "foo/bar"`

Required fields (per Agent Skills spec):
- `name`: must match the directory name; lowercase letters, numbers, and hyphens only
- `description`: describes what the skill does and when to use it

Optional fields (per Agent Skills spec):
- `license`
- `compatibility`
- `metadata` (arbitrary key/value mapping)
- `allowed-tools` (space-delimited list of tools that are pre-approved to run; experimental)

Nanobot extensions (optional):
- `tools`: list of tools the skill is allowed to call when activated.
  - If omitted, defaults to `inherit`.
  - If set to `inherit`, the skill inherits tools from its activation context (typically the agent and/or task).
  - If set to a list:
    - if the list contains `inherit`, the skill inherits tools **and** adds the additional tools from the list.
    - otherwise it is an explicit allowlist (no inherited tools).
  - Built-in tools are referenced by a logical name, e.g. `Read`, `Write`.
  - External tools are referenced as `${mcpServer}/${toolName}`.
- `tool_approvals`: tool approval policy for tool calls performed while executing this skill.
  - **Default behavior**: *all tool calls require approval*.
  - Implemented as an ordered list of "allow" rules; the first matching rule wins.
  - Rules can optionally constrain **tool arguments** (a `Map<string, any>`).
- `skills`: additional skills this skill may call when activated.
  - If omitted, defaults to `inherit`.
  - If set to `inherit`, the skill inherits skills from its activation context.
  - If set to a list:
    - if the list contains `inherit`, the skill inherits skills **and** adds the additional skills from the list.
    - otherwise it is an explicit allowlist (no inherited skills).
- `tasks`: tasks this skill may invoke when activated.
  - Semantics match `tools`/`skills`.
- `task_approvals`: task approval policy (same structure as `tool_approvals`).

### Task schema (front matter)

**Identity**
- `taskId` is derived from the relative directory path under `.nanobot/tasks/`.
  - Example: `.nanobot/tasks/foo/bar/TASK.md` → `taskId = "foo/bar"`

A task is a **multi-file** markdown workflow:
- `TASK.md` is required and is always the **first step**.
- If the front matter contains `next`, the runtime loads that next markdown file in the **same directory** and continues.
- Each step is a markdown file with YAML front matter + markdown body.

Required:
- `name`

Suggested fields:
- `metadata` (object; free-form)
- `agent`: agent ID (directory name) (optional; else system default)
- `tools`: tools visible/available during execution of this task/step.
  - If omitted, defaults to `inherit`.
  - If set to `inherit`, inherit tools from the activation context.
  - If set to a list:
    - if the list contains `inherit`, inherit tools **and** add the additional tools from the list.
    - otherwise it is an explicit allowlist (no inherited tools).
- `tool_approvals`: tool approval policy (same semantics as agents/skills).
- `skills`: skills visible/available during execution of this task/step.
  - If omitted, defaults to `inherit`.
  - If set to `inherit`, inherit skills from the activation context.
  - If set to a list:
    - if the list contains `inherit`, inherit skills **and** add the additional skills from the list.
    - otherwise it is an explicit allowlist (no inherited skills).
- `tasks`: tasks this task/step may invoke.
  - Semantics match `tools`/`skills`.
- `task_approvals`: task approval policy (same structure as `tool_approvals`).
- `inputs`: list of task inputs (allowed only in `TASK.md`)
  - Each input has `name` (string identifier), `description` (string), and optional `default` (string)
- `next`: relative path to the next step markdown file (in the same task directory)

---

## Runtime Architecture

### Components
1. **Config Loader / Indexer**
   - Scans configured roots for `*.md`
   - Parses YAML front matter + Markdown body
   - Validates against schemas
   - Builds an index: by id

2. **Orchestrator**
   - Resolves task → agent → skills (by directory-derived IDs)
   - Chooses execution mode:
     - *structured*: execute declared steps
     - *unstructured*: agent plans and calls skills/tools
   - Maintains run state machine

3. **Skill Runner**
   - Converts skill definition + inputs into a prompt/tool plan
   - Enforces tool requirements
   - Normalizes outputs into a structured object

4. **Tool Runtime**
   - Executes tool calls (filesystem, shell, web, git, etc.)
   - Enforces scopes (paths, network domains, command allowlists)
   - Captures stdout/stderr, exit codes, timings

5. **Memory & Storage**
   - **Run store**: run metadata, steps, tool calls (append-only)
   - **Artifact store**: generated files, patches, reports
   - Optional **retrieval index** for docs/code (vector or keyword)

6. **Guardrails**
   - Policy engine: allow/deny tool calls
   - Approval gate: request user confirmation for risky actions
   - Output checks: lint/tests, schema validation, heuristics

7. **Observability**
   - Structured logs + tracing per run/step
   - Replay: re-run with same configs and recorded tool outputs (when possible)

---

## Execution Model

### Run lifecycle
- `created` → `planned` → `running` → (`paused_for_approval`)* → `completed` | `failed` | `canceled`

### Step lifecycle
- `pending` → `running` → `succeeded` | `failed` | `skipped`

### Approvals
Approvals are triggered by:
- policy categories (e.g., any filesystem write outside allowlist)
- tool-level “high risk” flags

Approval request should include:
- proposed action + justification
- exact tool call payload
- diff preview (for file writes)

---

## Tooling Model

### Tool definition
Tools are registered by the host application and referenced by ID in config.

Recommended tool interface:
- `id`: string
- `name`: string
- `category`: `read | write | edit | glob | grep | bash | web | custom`
- `schema`: JSON schema for inputs/outputs
- `risk`: `low|medium|high`
- `execute(input, context) -> output`

### Safety defaults
- Deny by default; allowlist by agent + workspace.
- Path sandboxing: never allow `..` escape.
- Shell execution: prefer high-level tools; if shell is allowed, restrict commands.

---

## Reproducibility

The runtime should record enough information to replay a run:
- resolved agent/task/skill IDs (directory-derived)
- resolved model used
- tool calls + outputs
- hashes of config files (recommended)

---

## Validation

At load time:
- YAML parses successfully
- required fields present
- directory-derived IDs are unique per type
- references resolve (task.agent exists, skill ids exist)

At runtime:
- input schemas validated before executing a skill
- output schemas validated after skill completes
- tool call payload validated

---

## Example Prompts (recommended construction)

A typical execution prompt composition:
1. System safety policy
2. Agent Markdown body
3. Task Markdown body
4. Structured step (if applicable)
5. Skill Markdown body
6. Input JSON

Keep these separable so you can:
- change agent style without changing skills
- reuse skills across tasks

---

## Open Questions
- Do we support multi-agent (manager/worker) out of the gate?
- Do we allow skills to embed sub-steps (nested structured execution)?
- What is the minimal built-in toolset? (fs, shell, web, git)
- Where does secrets management live (env, vault integration, per-run ephemeral tokens)?

---

## Appendix: Built-in Tools

Built-in tool names are the logical names exposed by the host runtime (the same set of tools available to the assistant in this environment):

- `Read` — read files
- `Write` — write/update files
- `Edit` — apply exact replacements to existing files
- `Glob` — find files by glob pattern
- `Grep` — search file contents
- `Bash` — execute shell commands
- `WebFetch` — fetch web URLs

(Your runtime may add more built-ins over time. These names are the ones referenced in `tools:` lists.)

---

## Appendix: Example JSON for a Run Record (illustrative)
```json
{
  "run_id": "run_2025-12-22_001",
  "task": {"id": "scaffold_agent_pack"},
  "agent": {"id": "builder"},
  "status": "running",
  "created_at": "2025-12-22T17:20:00Z",
  "steps": [
    {
      "step_id": "s1",
      "uses": "skill:write_doc",
      "input": {"topic": "design"},
      "output": {"path": "docs/DESIGN.md"},
      "tool_calls": []
    }
  ]
}
```
