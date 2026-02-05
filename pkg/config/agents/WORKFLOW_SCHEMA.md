# Workflows

Workflows are Markdown files in `workflows/` that describe a sequence of steps for an AI to execute. Each step's output can be referenced by later steps using `{{step_id}}`.

## Example: Simple Review

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

## Example: With Conditions

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

## Example: With Error Handling

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

## Execution

After running a workflow, write a brief summary to `workflows/.runs/` noting what happened, what succeeded/failed, and any observations worth remembering.
