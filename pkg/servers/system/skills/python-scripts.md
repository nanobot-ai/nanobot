---
name: python-scripts
description: Write and execute Python scripts for workflow steps that benefit from programmatic execution. Use for data transformation, API calls, calculations, file processing, or any task where code is more reliable than natural language. Covers both inline one-off scripts and reusable script skills.
---

Use this skill when a workflow step (or sub-step) would benefit from programmatic execution rather than pure LLM reasoning.

## When to Use Python

**Good candidates for Python:**
- Data transformation (JSON manipulation, CSV processing, format conversion)
- API calls (especially with authentication, pagination, or complex request bodies)
- Calculations (math, statistics, aggregations)
- File processing (parsing, filtering, bulk operations)
- Deterministic operations (sorting, deduplication, exact string matching)
- Operations on large datasets where LLM might truncate or hallucinate

**Keep as pure LLM:**
- Tasks requiring judgment, analysis, or interpretation
- Natural language generation (summaries, explanations, reports)
- Tasks where flexibility and context-awareness matter more than precision
- One-off simple operations where writing code is overkill

## Two Approaches

### 1. Inline Scripts (One-off)

For simple, single-use scripts executed during workflow runs.

Write the script and execute it directly with `uv run`:

```python
#!/usr/bin/env python3
# /// script
# requires-python = ">=3.11"
# dependencies = [
#     "requests",
#     "rich",
# ]
# ///

import json
import requests

# Your code here
data = requests.get("https://api.example.com/data").json()
print(json.dumps(data, indent=2))
```

Execute with:
```bash
uv run script.py
```

The `# /// script` block declares inline dependencies — `uv` installs them automatically.

### 2. Reusable Script Skills

For scripts that will be used across multiple workflows or need documentation.

Create a skill directory following the agentskills.io specification:

```
my-script-skill/
├── SKILL.md           # Required: describes when/how to use
└── scripts/
    └── main.py        # The executable script
```

**SKILL.md format:**
```markdown
---
name: my-script-skill
description: What this script does and when to use it.
---

## Usage

Describe inputs, outputs, and how to invoke.

## Examples

Show example invocations and expected output.
```

**Script conventions:**
- Accept input via command-line arguments or stdin
- Output results to stdout (JSON preferred for structured data)
- Use exit codes: 0 for success, non-zero for failure
- Write errors to stderr

## Using uv

`uv` handles Python version management and dependency installation automatically.

**Run a script with inline dependencies:**
```bash
uv run script.py
```

**Run with additional arguments:**
```bash
uv run script.py --input data.json --output results.json
```

**Specify Python version:**
```python
# /// script
# requires-python = ">=3.11"
# ///
```

**Common dependencies:**
```python
# /// script
# dependencies = [
#     "requests",      # HTTP requests
#     "rich",          # Pretty output
#     "pydantic",      # Data validation
#     "pandas",        # Data processing
#     "jinja2",        # Templating
# ]
# ///
```

## Output Conventions

For the executor to capture and use script output:

**Structured data — use JSON to stdout:**
```python
import json
result = {"status": "success", "items": [...]}
print(json.dumps(result))
```

**Multiple outputs — use JSON with named fields:**
```python
print(json.dumps({
    "summary": "Processed 42 items",
    "data": [...],
    "errors": []
}))
```

**Progress/debug info — use stderr:**
```python
import sys
print("Processing...", file=sys.stderr)
```

**Status reporting — match workflow conventions:**
```python
# Success
print(json.dumps({"result": data}))
print("STATUS: COMPLETE", file=sys.stderr)

# Partial completion
print(json.dumps({"result": partial_data, "missing": [...]}))
print("STATUS: PARTIAL - Could not process 3 items", file=sys.stderr)

# Blocked
print("STATUS: BLOCKED - API rate limited", file=sys.stderr)
sys.exit(1)
```

## Example: Inline Script in Workflow Step

A workflow step that uses an inline Python script:

```markdown
### 3. transform_data (agent: build)

Write and execute a Python script to transform the issue data.

Input data:
{{fetch_issues}}

The script should:
1. Parse the JSON input
2. Filter issues older than 30 days
3. Group by label
4. Output as JSON with structure: {"by_label": {"bug": [...], "feature": [...]}}

Use `uv run` with inline dependencies. Output the result as JSON to stdout.
```

## Example: Creating a Reusable Skill

When the planner identifies a reusable script need:

```markdown
### 2. create_aggregator_skill (agent: build)

Create a reusable script skill at `aggregate-metrics/` following agentskills.io spec.

The skill should:
- Accept a JSON file path as input
- Aggregate numeric fields (sum, avg, min, max)
- Output aggregated results as JSON

Include:
- SKILL.md with usage documentation
- scripts/aggregate.py with inline uv dependencies
```

## Error Handling

**In scripts:**
```python
import sys

try:
    result = do_work()
    print(json.dumps(result))
except Exception as e:
    print(f"Error: {e}", file=sys.stderr)
    print("STATUS: BLOCKED - Script execution failed", file=sys.stderr)
    sys.exit(1)
```

**In workflow steps:**
Use `**On error:** continue` if the workflow can proceed without this step, or let it default to `stop` for critical steps.
