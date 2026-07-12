# AI-Assisted Development

This project includes a lightweight harness for AI-assisted work in
`.agents/harness`.

The harness exists to make agent sessions repeatable:

- give the agent a stable project context;
- define what must be checked before changes;
- keep task briefs focused;
- make review expectations explicit.

## Harness Files

- `.agents/README.md`: entry point for agents.
- `.agents/harness/project-context.md`: concise architecture and source of truth.
- `.agents/harness/task-brief-template.md`: template for starting a focused task.
- `.agents/harness/review-checklist.md`: checklist for implementation review.
- `.agents/harness/issue-template.md`: template for turning backlog items into issues.

## Working Rules

Agents should:

1. Treat code as the source of truth.
2. Use `docs/` as official documentation.
3. Use `local-docs/` only as historical reference when available.
4. Keep implementation, tests, and docs in sync.
5. Run the narrowest useful tests first, then broaden validation.
6. Never change unrelated files just to clean up style.

## Suggested Agent Prompt

```text
Read .agents/README.md and .agents/harness/project-context.md first.
Use docs/ as the official documentation source.
For this task: <describe the task>.
Before editing, summarize the files you expect to touch and why.
After editing, run the relevant tests and report results.
```
