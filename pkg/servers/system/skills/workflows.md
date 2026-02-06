---
name: workflows
description: Design and execute workflows - structured multi-step processes with inputs, conditions, and error handling.
---

# Workflows

Workflows are Markdown files in `workflows/` that codify repeatable processes. Each step's output can be referenced by later steps using `{{Step Name}}`.

## When to Use Workflows

Workflows are for repeatable tasks. If it's a one-time thing, just execute it directly.

## Workflow Format

Workflows are Markdown files with:
- **Title**: `# Workflow: workflow-name`
- **Description**: Brief explanation of the workflow's purpose
- **Inputs**: Optional parameters with defaults
- **Steps**: Numbered steps with clear instructions
- **Output**: Optional template for the final result

### Example: Simple Review

```markdown
# Workflow: code-review

Review code changes for quality issues.

## Inputs

- **target** (optional): Files to review. Default: `.`

## Steps

### 1. Find Changes
Find all modified code files in {{input.target}}.

### 2. Review Code
Review these files for quality issues:
{{Find Changes}}

Focus on: error handling, edge cases, and readability.

## Output

{{Review Code}}
```

### Example: With Conditions

```markdown
# Workflow: smart-fix

Analyze an issue and apply a fix only if it's safe to do so.

## Steps

### 1. Analyze Issue
Analyze the reported issue and determine severity.

### 2. Check Safety
Based on this analysis, determine if an automated fix is safe:
{{Analyze Issue}}

End your response with SAFE or UNSAFE.

### 3. Apply Fix
Apply the fix for: {{Analyze Issue}}

**Condition:** {{Check Safety}} contains SAFE

### 4. Create Manual Report
Create a report explaining why manual intervention is needed:
{{Analyze Issue}}

**Condition:** {{Check Safety}} contains UNSAFE
```

### Example: With Error Handling

```markdown
# Workflow: deploy-with-rollback

Deploy changes with automatic rollback on failure.

## Inputs

- **environment** (required): Target environment (staging, production)

## Steps

### 1. Run Tests
Run the full test suite. Report any failures.

### 2. Deploy
Deploy to {{input.environment}}.

**On error:** Rollback

### 3. Verify
Verify the deployment is healthy.

**On error:** Rollback

### 4. Rollback
Roll back to the previous version and report what went wrong.

## Output

Deployment to {{input.environment}} complete.
{{Verify}}
```

## Key Concepts

**Variables:** Use `{{input.name}}` for inputs and `{{Step Name}}` for outputs from previous steps.

**Conditions:** Add `**Condition:**` to make a step run only when the condition is true. Common patterns:
- `{{Step Name}} contains X` - check if output contains text
- `{{Step Name}} not empty` - check if output exists
- `{{input.flag}} equals yes` - check exact value

**Error handling:** Add `**On error:**` to control what happens when a step fails:
- `stop` (default) - halt the workflow
- `continue` - log the error and move to the next step
- `Step Name` - jump to a specific step (useful for cleanup/rollback)

**Output:** The `## Output` section defines the final result. If omitted, the last step's output is used.

## Designing Workflows

When creating a workflow, ask clarifying questions to understand what it should accomplish, what inputs it needs, and how errors should be handled. Use descriptive step names in title case (e.g., "Fetch Issues" not "step1") - these become variables for later steps.

For deterministic operations (data transformation, calculations), consider having steps use Python scripts (see `python-scripts` skill). For external services (Slack, Jira, databases), consider MCP servers (see `mcp-curl` skill).

Write workflows to `workflows/<name>.md`.

## Executing Workflows

Load the workflow from `workflows/<name>.md`, collect any required inputs, then execute each step in order. Check conditions before running conditional steps, interpolate variables (`{{input.name}}` and `{{Step Name}}`), and follow error handling directives.

**IMPORTANT:** Use TodoWrite to track workflow execution. Create a todo for each step and update its status as you progress. Todos are displayed to the user, so this is how you keep them informed of what's happening.

Provide running updates in normal chat messages as you work through each step. If errors occur, judgment calls are needed, or unexpected conditions arise, mention them in chat messages and add new todos as needed.

Complete tasks fully, follow requested formats exactly, and provide specific analysis rather than vague summaries.

After execution, write a brief summary to `workflows/.runs/` noting what succeeded, failed, or was skipped. Report results to the user using the workflow's Output section if present.

## Improving Workflows

After executing a workflow, you can analyze the execution and propose improvements to the workflow file.

**What to look for:**
- Steps that needed retries or failed - unclear prompt? missing format spec?
- Steps that produced poor output - missing context? too vague?
- Steps that should be split into smaller steps
- Missing error handling for steps that failed
- Hardcoded values that should be inputs

**Process:** Review which steps succeeded/failed/retried and what fixes were applied. For each issue, determine the concrete fix (rewrite prompt, add format spec, split step, add error handling, etc.). Present each proposed change with the issue, current text, and proposed text.

**IMPORTANT:** Get user approval before editing the workflow file. Never apply changes automatically.

**Example:**
```
## Proposed Changes to issue-triage.md

### 1. Step: Analyze Issues
**Issue:** Agent returned prose instead of JSON. Had to retry.
**Change:** Add explicit format requirement.

Current:
> Analyze each issue and categorize by priority.
> Return the results as JSON.

Proposed:
> Analyze each issue and categorize by priority.
>
> Return as a JSON array:
> [{"number": 1, "title": "...", "priority": "High|Medium|Low", "reason": "..."}]

Apply these changes? (yes/no/select specific)
```

Don't change things that worked fine or add complexity without a concrete problem to solve.
