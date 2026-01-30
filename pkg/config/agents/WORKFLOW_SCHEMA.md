# Workflow Schema

This document defines the schema for workflows that will be executed by the executor.

## Overview

A workflow is a sequence of steps where each step can:

- Execute a task
- Capture output for use in later steps
- Use conditional logic to branch execution

## File Location

Workflows are stored in `workflows/` as Markdown files (`.md`).

## Workflow Structure

```markdown
# Workflow: <name>

<description>

## Inputs

- **input_name** (required): Description.
- **input_name** (optional): Description. Default: `value`

## Steps

### 1. step_id 
Task description...

**On error:** stop | continue | step_id (optional)
**Condition:** `{{step_id}} contains X` (optional)

---

### 2. next_step_id 
...

---

## Output

Final output template with {{variable}} interpolation
```

## Field Reference

### Workflow Header

- **Title**: `# Workflow: <name>` - Required. The name is a unique identifier.
- **Description**: The paragraph immediately after the title. Required.

### Inputs Section (Optional)

Define workflow inputs that can be referenced in steps:

```markdown
## Inputs

- **input_name** (required): Description of the input.
- **input_name** (optional): Description. Default: `value`
```

- **Input name**: The identifier used to reference this input via `{{input.name}}`
- **Required/Optional**: Indicates whether the input must be provided
- **Description**: Explains what the input is for
- **Default**: For optional inputs, the value used if not provided

### Steps Section (Required)

Define the sequence of steps:

```markdown
## Steps

### 1. step_id 
Task description with {{step_id}} interpolation...

**On error:** stop | continue | step_id
**Condition:** `{{step_id}} contains X`
```

#### Step Header

`### N. step_id` - Required for each step.

- **N**: Step number (for readability, not used for ordering)
- **step_id**: Unique identifier for the step. Used to reference output in later steps (via `{{step_id}}`) and in `On error` jumps. Use descriptive names like `fetch_issues` or `analyze_code`.

#### Task Description

The prose content after the step header. This is the task to execute. Can include `{{variable}}` interpolation.

#### Step Fields

| Field | Required | Description |
|-------|----------|-------------|
| `**On error:**` | No | Error handling: `stop` (default), `continue`, or a `step_id` to jump to |
| `**Condition:**` | No | Step only runs if condition is true. See [Condition Syntax](#condition-syntax) |

Step outputs are automatically stored and can be referenced in later steps using `{{step_id}}`.

### Condition Syntax

Conditions control whether a step executes. The step runs only if the condition evaluates to true.

#### Operators

| Operator | Syntax | Description |
|----------|--------|-------------|
| `contains` | `{{var}} contains X` | True if var contains substring X (case-insensitive) |
| `equals` | `{{var}} equals X` | True if var exactly equals X (case-sensitive) |
| `empty` | `{{var}} empty` | True if var is empty, missing, or whitespace-only |
| `not` | `{{var}} not <op>` | Negates the operator that follows |

#### Examples

```markdown
# Run only if safety check passed
**Condition:** `{{safety_check}} contains SAFE`

# Run only if status is exactly "success"
**Condition:** `{{status}} equals success`

# Run only if optional input was provided
**Condition:** `{{input.filter}} not empty`

# Run only if no errors found
**Condition:** `{{result}} not contains error`

# Run only if status is not "skipped"
**Condition:** `{{status}} not equals skipped`
```

#### Truthy Check

A variable reference alone checks if the value is non-empty:

```markdown
**Condition:** `{{input.optional_flag}}`
```

This is equivalent to `{{input.optional_flag}} not empty`.

### Output Section (Optional)

Define the final workflow output:

```markdown
## Output

Final output template with {{variable}} interpolation.
```

The output section supports full variable interpolation and conditional content. If omitted, the output of the last executed step is used.

## Variable Interpolation

Use `{{variable_name}}` syntax to reference values:

| Syntax | Description |
|--------|-------------|
| `{{input.name}}` | Workflow input value |
| `{{step_id}}` | Output from step with that ID |
| `{{steps.step_id}}` | Same as above (explicit) |
| `{{workflow.name}}` | Workflow name |
| `{{workflow.cwd}}` | Current working directory |

## Examples

### Simple Sequential Workflow

```markdown
# Workflow: code-review

Review code changes for quality and security.

## Inputs

- **target** (optional): Files to review. Default: `.`

## Steps

### 1. find_changes 
Find all code files in {{input.target}} that have been modified.

---

### 2. review_code 
Review these files for code quality:
{{find_changes}}

---

## Output

{{review_code}}
```

### Workflow with Conditions

```markdown
# Workflow: smart-fix

Analyze issue and apply fix if safe.

## Steps

### 1. analyze_issue 
Analyze the reported issue and determine severity.

---

### 2. check_safety 
Based on this analysis, determine if an automated fix is safe:
{{analyze_issue}}

Respond with SAFE or UNSAFE and explanation.

---

### 3. apply_fix 
Apply the fix for: {{analyze_issue}}

**Condition:** `{{check_safety}} contains SAFE`

---

### 4. create_manual_report 
Create a report for manual review: {{analyze_issue}}

**Condition:** `{{check_safety}} contains UNSAFE`
```

## Execution Model

1. **Initialization**: Parse workflow Markdown, validate structure, collect inputs
2. **Step Execution**: For each step (in sequence):
   - Check condition (skip if false)
   - Interpolate variables in task
   - Execute the step
   - Store output in context
   - Handle errors per on_error setting
3. **Completion**: Interpolate and return final output

## Error Handling

- `**On error:** stop` (default) - Halt workflow, report error
- `**On error:** continue` - Log error, continue to next step
- `**On error:** <step_id>` - Jump to specified step (for recovery/cleanup)

## Step Output Format

Step outputs should end with a status marker to indicate completion state:

| Status             | Meaning                                                                |
|--------------------|------------------------------------------------------------------------|
| `STATUS: COMPLETE` | Task fully accomplished, all items processed                           |
| `STATUS: PARTIAL`  | Some items processed but not all (explain what's missing)              |
| `STATUS: BLOCKED`  | Cannot complete due to missing information or access (explain blocker) |

This helps the executor track progress and enables better error handling and reporting in execution summaries.

---

## Execution Summaries

After each workflow run, the executor writes a summary to `workflows/.runs/`.

### File Naming

```
workflows/.runs/execution-<datetime>-summary.md
```

Example: `workflows/.runs/execution-2024-01-19T10-30-00-summary.md`

### Summary Structure

```markdown
# Workflow Execution Summary

**Workflow:** <name>
**Executed:** <ISO 8601>
**Duration:** <seconds>s
**Result:** success | failed | partial

## Inputs

- **input_name:** value

## Steps

| Step | Result | Notes |
|------|--------|-------|
| step_id | success | Brief note |
| step_id | failed | Error description |
| step_id | skipped | Condition not met |

## Failure Details (if failed)

**Failed at step:** step_id
**Error:** What went wrong
**Impact:** Workflow halted, steps X, Y, Z did not run

## Observations

- Notable finding 1
- Notable finding 2

## Issues

- **step_id:** Issue description and any remediation attempted
```

### Example: Successful Run

```markdown
# Workflow Execution Summary

**Workflow:** code-review
**Executed:** 2024-01-19T10:30:00Z
**Duration:** 87s
**Result:** success

## Inputs

- **target:** src/

## Steps

| Step | Result | Notes |
|------|--------|-------|
| discover_files | success | Found 23 files |
| review_quality | success | 5 issues identified |
| generate_report | success | Report generated |

## Observations

- High test coverage in src/utils/
- Several files lack documentation headers
```

### Example: Failed Run

```markdown
# Workflow Execution Summary

**Workflow:** code-review
**Executed:** 2024-01-19T14:22:00Z
**Duration:** 58s
**Result:** failed

## Inputs

- **target:** src/

## Steps

| Step | Result | Notes |
|------|--------|-------|
| discover_files | success | Found 23 files |
| review_quality | failed | Timeout after 45s |
| generate_report | not run | — |

## Failure Details

**Failed at step:** review_quality
**Error:** Timed out scanning large file set
**Impact:** Workflow halted. Steps not run: generate_report

## Issues

- **review_quality:** Timeout on 23 files. Consider batching or increasing timeout for large file sets.
```

---

## Best Practices

1. **Use descriptive step IDs** - They appear in logs, error messages, and are used to reference outputs (e.g., `fetch_issues` not `step1`)
2. **Keep tasks focused** - One clear objective per step
3. **Use conditions for branching** - Rather than complex logic in prompts
4. **Provide defaults for inputs** - Makes workflows easier to run
5. **Review execution logs** - Learn from past runs to improve workflows
6. **Maintain Lessons Learned** - Document solutions to recurring issues
