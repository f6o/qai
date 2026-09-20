---
name: qai
description: Manage tasks, ideas, and Pomodoro focus sessions with the qai CLI. Use when the user wants to add, list, complete, delete, or promote tasks/ideas, check focus statistics, view activity logs, or generate the daily markdown report.
allowed-tools: Bash(qai:*)
---

# qai — task management CLI

Data lives under `~/.config/qai` (config `config.toml`, activity log `logs6.jsonl`).
Task statuses: `idea` -> `todo` -> `doing` -> `done`.

## Setup

`qai` must be on PATH. If `command -v qai` fails, build it from the repo:

```bash
go build -o ~/.local/bin/qai .   # from /Users/yappy/workspace/qai
qai init                          # create ~/.config/qai if missing
```

This document is what `qai skills` prints (front matter stripped), so the
instructions always match the binary that prints them.

## Machine-readable contract

Data commands accept `--json`: stdout is exactly ONE JSON line, nothing else.
Failures exit 1 with a message on stderr (stdout stays empty). Chain with jq:

```bash
ID=$(qai todo add --json "write tests" | jq -r .id)
qai todo add --json "step 1" --parent "$ID"
qai todo done --json "$ID"
```

Task record: `{"id":N,"title":"...","status":"idea|todo|doing|done","priority":N,"created_at":"..."}`
Keys `parent_id` / `started_at` are omitted when unset.
`list`/`stats`/`logs --json` emit `{"ideas":[...],"todos":[...]}`, a stats object, or a log array respectively.

## Commands

### Todos

```bash
qai todo add "content"              # add (default priority from config)
qai todo add "content" --parent 3   # link under a parent task
qai todo list                       # all todos, priority desc
qai todo done 42                    # mark done (writes status_change log)
qai todo delete 42                  # delete; refused if it has children
```

### Ideas

```bash
qai idea add "content"              # add (priority 0)
qai idea list
qai idea promote 42                 # idea -> todo; ID/priority/parent untouched
qai idea delete 7                   # leaf ideas only
qai idea delete --force 7           # delete the whole subtree
```

### Read / report

```bash
qai list                            # ideas + todos together
qai list --above 5                 # only priority >= 5
qai stats                          # task counts + focus time + today
qai logs                           # activity log
qai logs --type status_change      # filter: focus_complete | focus_skip | focus_quit | task_create | task_continue | status_change
qai logs path
qai md                             # today's markdown report
qai md --save                      # also save under ~/.config/qai/markdown/
```

### Pomodoro

`qai timer` (alias `qai pomo`) is an interactive TUI. Do NOT run it in an agent
flow — it blocks on a terminal. Tell the user to run it in their own terminal.
`qai todo add "..." --start` also launches the TUI; avoid it in scripts.

## Pitfalls

- `todo done` and `idea promote` are NOT idempotent: unknown ID, wrong status, or already-done exits 1.
- `idea promote` does not change priority: a promoted idea keeps priority 0 and may be hidden by `list --above N`.
- `todo delete` refuses tasks that have children — delete children first. `idea delete` only deletes ideas and needs `--force` for subtrees.
- Deletions are NOT logged (no log event).
- `todo add -i` (interactive) refuses `--json`.
- Before mutating, read current state with `qai list --json` to avoid duplicate titles and stale IDs.
