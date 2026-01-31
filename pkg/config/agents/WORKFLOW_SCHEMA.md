# Workflows

Workflows are Markdown files in `workflows/` that describe a sequence of steps for an AI to execute. Each step's output can be referenced by later steps using `{{step_id}}`.

## Example: Simple Review

```markdown
# Workflow: code-review

Review code changes for quality issues.

## Inputs

- **target** (optional): Files to review. Default: `.`

## Steps

### 1. find_changes
Find all modified code files in {{input.target}}.

### 2. review_code
Review these files for quality issues:
{{find_changes}}

Focus on: error handling, edge cases, and readability.

## Output

{{review_code}}
```

## Example: With Conditions

```markdown
# Workflow: smart-fix

Analyze an issue and apply a fix only if it's safe to do so.

## Steps

### 1. analyze_issue
Analyze the reported issue and determine severity.

### 2. check_safety
Based on this analysis, determine if an automated fix is safe:
{{analyze_issue}}

End your response with SAFE or UNSAFE.

### 3. apply_fix
Apply the fix for: {{analyze_issue}}

**Condition:** {{check_safety}} contains SAFE

### 4. create_manual_report
Create a report explaining why manual intervention is needed:
{{analyze_issue}}

**Condition:** {{check_safety}} contains UNSAFE
```

## Example: With Error Handling

```markdown
# Workflow: deploy-with-rollback

Deploy changes with automatic rollback on failure.

## Inputs

- **environment** (required): Target environment (staging, production)

## Steps

### 1. run_tests
Run the full test suite. Report any failures.

### 2. deploy
Deploy to {{input.environment}}.

**On error:** rollback

### 3. verify
Verify the deployment is healthy.

**On error:** rollback

### 4. rollback
Roll back to the previous version and report what went wrong.

## Output

Deployment to {{input.environment}} complete.
{{verify}}
```

## Key Concepts

**Variables:** Use `{{input.name}}` for inputs and `{{step_id}}` for outputs from previous steps.

**Conditions:** Add `**Condition:**` to make a step run only when the condition is true. Common patterns:
- `{{step_id}} contains X` - check if output contains text
- `{{step_id}} not empty` - check if output exists
- `{{input.flag}} equals yes` - check exact value

**Error handling:** Add `**On error:**` to control what happens when a step fails:
- `stop` (default) - halt the workflow
- `continue` - log the error and move to the next step
- `step_id` - jump to a specific step (useful for cleanup/rollback)

**Output:** The `## Output` section defines the final result. If omitted, the last step's output is used.

## Execution

After running a workflow, write a brief summary to `workflows/.runs/` noting what happened, what succeeded/failed, and any observations worth remembering.
