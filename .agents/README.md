# Agent Harness

This directory contains the working harness for AI-assisted development on
Codda.

Before making changes, read:

1. `.agents/harness/project-context.md`
2. the relevant document in `docs/`
3. the code that implements the behavior

The codebase is the source of truth. `docs/` is the official documentation.
`local-docs/` is historical reference and is ignored by git.

## Expected Workflow

1. Restate the task and identify the likely files to inspect.
2. Read the relevant code before proposing changes.
3. Keep edits scoped to the requested behavior.
4. Update docs when behavior or public contracts change.
5. Run focused tests first, then broader tests when appropriate.
6. Report commands run and any validation that could not be performed.

Use `.agents/harness/task-brief-template.md` to create precise task prompts.
