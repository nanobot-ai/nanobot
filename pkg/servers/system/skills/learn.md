---
name: learn
description: Review workflow execution results and apply improvements to the workflow file. Use after running a workflow to analyze what succeeded, failed, or needed retries, then propose and apply targeted fixes to step prompts, format specs, error handling, or agent selection.
---

Analyze the workflow execution that just completed and propose improvements to the workflow file itself.

## What to Look For

**Step Issues:**
- Steps that needed retries - why? Unclear prompt? Missing format spec?
- Steps that failed or timed out - too broad? Wrong agent?
- Steps that produced poor output - missing context? Vague instructions?
- Steps that had to be fixed on-the-fly - what fix worked?

**Structural Issues:**
- Steps that should be split into multiple smaller steps
- Missing error handling (`On error:`) for steps that failed
- Conditions that didn't evaluate as expected

**Input/Output Issues:**
- Values that should have been inputs but were hardcoded
- Missing required inputs that caused problems
- Optional inputs that should have had different defaults
- Output format that didn't serve downstream steps well

**Agent Selection:**
- Wrong agent for the task (e.g., used `general` when `explore` was better)
- Agent that struggled with the task scope

## Process

1. **Review the execution** - Look at:
   - Which steps succeeded/failed/were retried
   - What fixes were applied during execution
   - Any observations or issues noted in the summary
   - The actual prompts that were sent vs. what worked

2. **Identify improvements** - For each issue, determine the concrete change:
   - Rewrite a step's task description to be clearer
   - Add explicit format requirements (JSON schema, list format, etc.)
   - Add examples to step prompts
   - Split a step into multiple steps
   - Change the agent for a step
   - Add/modify `On error:` or `Condition:`
   - Add/modify inputs with better defaults

3. **Propose changes** - Present each change to the user:
   ```
   ## Proposed Changes to <workflow-name>.md

   ### 1. Step: <step_id>
   **Issue:** <what went wrong>
   **Change:** <what you want to modify>

   Current:
   > <existing text>

   Proposed:
   > <new text>

   ### 2. ...
   ```

4. **Get approval** - Ask the user which changes to apply:
   - "Apply all changes?"
   - "Apply changes 1, 3, 4?"
   - "Skip all?"

5. **Apply changes** - Only after user approval, edit the workflow file

## What NOT to Change

- Don't add verbose explanations or comments to the workflow
- Don't change things that worked fine
- Don't add complexity that isn't justified by an actual problem
- Don't change the workflow's fundamental purpose or scope
- Don't remove steps that succeeded without issues

## Example

After an execution where step `analyze_issues` failed because the agent returned prose instead of JSON:

```
## Proposed Changes to issue-triage.md

### 1. Step: analyze_issues
**Issue:** Agent returned prose instead of JSON array. Had to retry with explicit format instructions.
**Change:** Add explicit JSON format requirement with example.

Current:
> Analyze each issue and categorize by priority.
> Return the results as JSON.

Proposed:
> Analyze each issue and categorize by priority.
>
> Return as a JSON array with this exact structure:
> ```json
> [{"number": 1, "title": "...", "priority": "High|Medium|Low", "reason": "..."}]
> ```
>
> Process ALL issues. Do not truncate.

---

Apply these changes? (yes/no/select specific changes)
```
