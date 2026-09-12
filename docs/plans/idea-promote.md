# Implementation Plan: `qai idea promote` — idea→todo の headless 昇進

## Problem
書き込み契約は add/done/delete が揃ったが、**idea→todo の昇進だけ対話 UI か tasks.yaml 直編集に依存**している(task-file-split.md も「idea→todo は idea.yaml 削除+tasks.yaml 追加」とファイル編集前提で、CLI 不在)。エージェントはアイデアをタスクに分解開始できない。

## Objective
`qai idea promote <ID> [--json]` を追加し、status `idea` を `todo` へ非対話で昇進させる。done と同じ契約族に収める。

## Design
- **対象**: status==idea のみ。todo/doing/done の ID 指定は exit 1(`err_not_idea` に status 値を添える)。未発見 ID は exit 1(`err_not_found`)。数値でない引数は exit 1
- **遷移**: `task.Status = model.StatusTodo` のみ。**ID は不変**(子 todo の `parent_id` は ID 参照なので昇進後も自動で正しく繋がる — 実装側で触らないこと)。priority/parent/時刻は一切変更しない
- **既知の帰結(設計レビュー R1 反映・仕様変更なし)**: idea は priority 0 固定(`cmd/idea_add.go`)なので、昇進後も todo は priority 0 のまま → `qai list -A N` 系フィルタでは体系的に非表示になり得る。priority 不変の一貫性(昇進は status 変更のみ)を優先し、Long に一行警告を出す: `Note: promotion does not change priority; a promoted task keeps priority 0 and may be hidden by list --above filters.`
- **ログ**: done と同一形 — `EventStatusChange`、FromStatus=`idea`、ToStatus=`todo`、Content 空。順序は Update(保存)→AppendNew→stdout(AppendNew の戻り値は無視=house pattern 踏襲)
- **保存**: 既存 `TaskStore.Update(tasks, *task)` 経由(自己保存型。外部 Save 呼び出しは Phase B の delete 系のみに限定を維持)
- **出力**: `--json` なし → `cmd.idea_promote.success = Promoted: %[1]s (ID: %[2]d)` の散文。`--json` → 昇進**後**の task レコード1行(`model.Task.MarshalJSON` 再利用、トップレベル=オブジェクト)
- 冪等でなくする: 再 promote は `err_not_idea`(status が既に todo)で exit 1 — done と同じ方針
- `SilenceUsage: true`、SilenceErrors 未設定、Long/Example 最小(テンプレは cmd/todo_done.go を構造ごと踏襲: import/Args/ガード列/Update/ログ/出力/init)

## Files
```
cmd/idea_promote.go            (NEW)
i18n/locales/locale_en-US.ini  (+ # Idea promote セクション: short/success/flag_json/err_not_found/err_not_idea)
```
変更はこの2つだけ。storage/model/add-delete 系は不変更。

## Verification Plan(scratch HOME、B/C で実証済みの方式)
1. gofmt -l cmd / go build ./... / go vet ./... / go test ./...
2. scratch: `idea add --json "p"` → pid → `todo add --json "c" --parent <pid>` → `idea promote <pid> --json` → レコード `.status=="todo"` かつ `.id==pid` / `list --json` で `.ideas` から消え `.todos` に pid が入る / 子 c の `parent_id==pid` が不変であること(jq で突合)
3. logs: scratch の `logs --json` に `status_change` で from=="idea", to=="todo" が1件以上。イベント種別が既知の 5 種に収まること(新 kind 非追加)
4. 異常系 exit 1 + stdout 空: 未発見(9999)/ todo の id / done の id(scratch で promote→done して再 promote)/ 非数値引数
5. 実 HOME 不変: 検証前後で `go run . list --json | jq '[.ideas[].id,.todos[].id]|max'` 一致。検証は scratch 内のみで完結、実 ~/.config/qai への書き込み実行禁止
6. 掃除: mktemp 配下のみ削除。git commit 禁止

## フェーズ分割(Qwen 向け micro-step)
単一フェーズ(ファイル2つ)。ただし毎回「1応答=編集1回」を指示文で強制する: (1) idea_promote.go 骨格+Update/遷移 → build、(2) 出力分岐+init → build、(3) i18n → 検証、(4) 検証+最終報告。読解・思考の往復だけでターンを終わらせないこと(前 Phase B の 11h ハング再発防止)。

## 対象外(記録)
- 逆遷移(todo→idea)、`promote --to doing`(pomo 開始と重複するため do nothing)
- task-file-split.md(idea.yaml/done.yaml 分割)側への波及 — 分割実装時に promote は「ファイル間移動」へ拡張する論点として同文書に残す
</content>
