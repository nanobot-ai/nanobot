---
name: learn
description: Review workflow execution results and apply improvements to the workflow file.
---

Analyze the workflow execution that just completed and propose improvements.

## What to Look For

- Steps that needed retries or failed - unclear prompt? missing format spec?
- Steps that produced poor output - missing context? too vague?
- Steps that should be split into smaller steps
- Missing error handling for steps that failed
- Hardcoded values that should be inputs

## Process

1. **Review** - Look at which steps succeeded/failed/retried and what fixes were applied
2. **Identify changes** - For each issue, determine the concrete fix (rewrite prompt, add format spec, split step, add error handling, etc.)
3. **Propose** - Present each change with the issue, current text, and proposed text
4. **Get approval** - Ask the user which changes to apply
5. **Apply** - Only edit the workflow file after user approval

## Example

```
## Proposed Changes to issue-triage.md

### 1. Step: analyze_issues
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

## What NOT to Change

- Things that worked fine
- The workflow's fundamental purpose
- Don't add complexity without a concrete problem to solve
