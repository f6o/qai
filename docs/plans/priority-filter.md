# Implementation Plan: Priority Filter for `qai list`

## Objective
`qai list` に `--above` (`-A`) フラグを追加し、指定した優先度以上のタスクのみを表示できるようにする。

例: `qai list --A 10` → 優先度10以上のアイテムのみ表示

## Target Command

| Command | Flags | Description |
| :--- | :--- | :--- |
| `qai list` | `--above N`, `-A N` | 優先度が N 以上の ideas/todos のみ表示。省略時は全件表示（従来通り）。 |

## Implementation Steps

### 1. `cmd/list.go` の変更

1. パッケージ変数 `listMinPriority int` を追加
2. `init()` 内でフラグを登録:
   ```go
   listCmd.Flags().IntVarP(&listMinPriority, "above", "A", 0, i18n.T("cmd.list.flag.above"))
   ```
3. `RunE` 内で `FilterIdeas` / `FilterTodos` の後にフィルタリングを追加:
   ```go
   if listMinPriority > 0 {
       ideas = filterByMinPriority(ideas, listMinPriority)
       todos = filterByMinPriority(todos, listMinPriority)
   }
   ```
4. ヘルパー関数を追加:
   ```go
   func filterByMinPriority(tasks []model.Task, minPriority int) []model.Task {
       var result []model.Task
       for _, t := range tasks {
           if t.Priority >= minPriority {
               result = append(result, t)
           }
       }
       return result
   }
   ```

### 2. i18n メッセージの追加

`i18n/locales/locale_en-US.ini` (および日本語ロケールがあれば) に以下を追加:
```ini
cmd.list.flag.above = Show only items with priority >= N
```

## File Structure Changes
```text
cmd/
└── list.go          (MODIFY: add --above flag and filter logic)
i18n/locales/
└── locale_en-US.ini (MODIFY: add flag description message)
```

## Verification Plan
1. `go build` が成功すること
2. `go test ./...` が通ること
3. `qai list` でフラグなし時に従来通り全件表示されること
4. `qai list --A 10` で優先度10以上のタスクのみ表示されること
5. `qai list --A 0` で全件表示されること（デフォルト値と同じ動作）
