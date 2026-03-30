---
name: scheduled-tasks
description: Load this skill for ANY request that mentions scheduled tasks, recurring jobs, reminders, future runs, or time-based automation.
---

# Scheduled Tasks

Scheduled tasks are persistent task definitions managed through the `nanobot.tasks` tools. When a scheduled task fires, nanobot starts a new chat thread and runs the saved prompt there.

## When to Use Scheduled Tasks

Use scheduled tasks when work should happen later or recur on a schedule:
- Daily, weekly, or monthly reports
- Reminders or recurring checks
- Periodic file updates, cleanups, or syncs

If the user wants the work done right now and not again later, do it directly instead of creating a scheduled task.

## Scheduled Tasks vs Workflows

Do not conflate scheduled tasks with workflows:
- **Workflows** are reusable procedures stored under `workflows/`.
- **Scheduled tasks** decide when nanobot should start a future chat thread.

Use a workflow when the user wants a reusable multi-step procedure. Use a scheduled task when the user wants something to happen later or recur automatically.

If the user wants a workflow to run on a schedule, keep them separate:
1. Create or update the workflow first.
2. Create a scheduled task whose prompt tells nanobot to run that workflow.

## Managing Existing Tasks

Before updating, deleting, or manually running an existing task, call `listScheduledTasks` unless the user already provided the exact `task:///...` URI.

Use the task URI for all follow-up task operations:
- `updateScheduledTask`
- `deleteScheduledTask`
- `startScheduledTask`

## Timezone Rules

When creating a scheduled task, always set the timezone explicitly.

- If the user's current timezone is already known from context, use that by default unless they say otherwise.
- If you do not know the user's timezone, collect it before creating the task.
- Do not guess a timezone from a city, country, or language unless the user explicitly gave enough information to make that unambiguous.

Unless the user asks for a different timezone, new scheduled tasks should use the user's current timezone.

## Valid Schedule Shapes

The task tools only accept five-field cron expressions in these shapes:

- **Daily:** minute hour `* * *`
  - Example: `45 2 * * *` for every day at 02:45
- **Weekly:** minute hour `* * day-of-week`
  - Example: `20 23 * * 1,2,3,4` for Monday through Thursday at 23:20
- **Monthly:** minute hour `day-of-month * *`
  - Example: `45 2 22,24 * *` for the 22nd and 24th of every month at 02:45
- **One-time date:** minute hour `day-of-month month *` with a matching expiration date
  - Example: `45 2 26 3 *` with expiration `2026-03-26` for March 26 at 02:45

Do not use unsupported cron shapes. In particular:
- Do not use six-field cron expressions
- Do not use schedules that combine both day-of-month and day-of-week
- Do not use arbitrary cron patterns outside the daily, weekly, monthly, or one-time date forms above

## Designing Good Task Prompts

Each scheduled run starts a fresh chat thread. The stored prompt should be self-contained:
- State exactly what to do
- Name any files, outputs, or deliverables to produce
- Mention any workflow to run by name if applicable
- Avoid depending on transient context from the current conversation

For recurring actions that send messages or modify external systems, confirm the user's intent unless they have already clearly requested that recurring behavior.
