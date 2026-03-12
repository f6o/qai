# Implementation Plan: Command Reorganization

## Objective
Reorganize the command structure from a mix of "verbs" and "nouns" to a resource-centric (noun-based) architecture to provide a consistent and intuitive CLI experience.

## Target Command Structure

| Resource | Action | Command Example | Old Command | Flags | Description |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **`md`** | preview | `qai md` | `qai preview` | | Show today's markdown content in stdout. |
| | save | `qai md -s` | `qai start` | `--save`, `-s` | Generate and save today's markdown file. |
| **`stats`** | show | `qai stats` | `qai report` | | Show focus statistics and activity report. |
| **`logs`** | path | `qai logs path` | (Same) | | Display the path to the log file. |
| **`idea`** | add/list | `qai idea add` | (Same) | | Manage ideas (remains independent). |
| **`todo`** | add/list | `qai todo add` | (Same) | | Manage todos (remains independent). |
| **`timer`** | start | `qai timer` | `qai pomo` | | Start a Pomodoro timer session (alias: `pomo`). |
| **`list`** | list all | `qai list` | (Same) | | Summary view of both ideas and todos. |

## Implementation Steps

### Phase 1: Markdown Integration (`md`)
1. Create `cmd/md.go`.
   - Define `mdCmd` and add it to `rootCmd`.
   - Implement `RunE` to handle logic from `qai start` and `qai preview`.
   - Default behavior: Print generated markdown to stdout.
   - Add `--save` (`-s`) flag to save the output to a file in the configured directory.
2. Remove old files: `cmd/start.go`, `cmd/preview.go`, and `cmd/preview_test.go`.

### Phase 2: Statistics (`stats`)
1. Rename `cmd/report.go` to `cmd/stats.go`.
   - Change command name to `stats` and register it to `rootCmd`.
   - `qai logs` and `qai logs path` remain unchanged.

### Phase 3: Timer Command (`timer`)
1. Rename `cmd/pomo.go` to `cmd/timer.go`.
   - Change command name from `pomo` to `timer`.
   - Add `Aliases: []string{"pomo"}` to the `cobra.Command` definition.

### Phase 4: I18n and Message Cleanup
1. Update `i18n/locales/locale_en-US.ini`.
   - Update message keys: e.g., `cmd.report.short` -> `cmd.stats.short`.
   - Update descriptions to be resource-focused.
2. Verify all `Short` and `Long` descriptions in `cmd/*.go` match the new terminology.

## File Structure Changes
```text
cmd/
├── md.go           (NEW: integrated start + preview)
├── logs.go         (No change)
├── logs_path.go    (No change)
├── stats.go        (RENAME: report.go)
└── timer.go        (RENAME: pomo.go)
```

## Verification Plan
1. Run `go build` to ensure no compilation errors.
2. Verify `qai md` prints the markdown to the terminal without creating a file.
3. Verify `qai md -s` creates a markdown file in the configured directory.
4. Verify `qai stats` displays the summary report correctly.
5. Verify `qai logs path` prints the correct log file path.
6. Verify `qai timer` (and `qai pomo`) starts the TUI timer.
