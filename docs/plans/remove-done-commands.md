# Implementation Plan: `qai todo done` / `delete` — 非対話の完了・削除(エージェント書込み契約)

## Problem
現状の書き込み契約は `add --json` のみで、**一方通行**である:

- 完了: doing→done への遷移は対話 pomo 内部(internal/pomo/model.go:309)でしか起きず、`qai timer` にサブコマンドは無い(headless 遷移コマンド不在)
- 削除: `TaskStorage` に Remove API 自体が無い(internal/storage/task.go:16-116)。結果、エージェントが成果物の掃除・取消をしたくても tasks.yaml 直編集に追いやられ、自動採番(GetMaxID)と logs 記録を壊す

## Objective
resource-centric 体系(`todo`/`idea` の名詞配下)を壊さず、非対話で安全な **完了** と **削除** を追加する。`--json`(既確立の機械契約)と **ログ整合性**(status_change イベント)を最初から持つ。

## Design

### 1. `qai todo done <ID> [--json]`
- 対象: status が `todo` または `doing` のタスクのみ(`idea` は対象外 → stderr メッセージ+exit 1)
- 遷移: `status = done`、**`StartedAt` は触らない**(doing 由来の開始時刻は履歴として温存)。`UpdatedAt` フィールドは存在しないため追加しない
- **ログ必須**: pomo と同一の `EventStatusChange`(FromStatus=旧, ToStatus=done)を `LogStore.AppendNew` へ。これをや忘れると stats/logs の「完了」が二重基準になる(本計画の最大の陷阱)
- 出力: `--json` なし → `i18n` 散文(`cmd.todo_done.success = Completed: %[1]s (ID: %[2]d)`)。`--json` → **遷移後の task レコードを1行**(既存 `model.Task.MarshalJSON` を再利用。トップレベルオブジェクト)
- 存在しない ID / done 済み ID: 明示エラー exit 1(冪等の「もう done」は成功にしない — エージェントに再実行可能さを装わせない)

### 2. `qai idea delete <ID> [--force] [--json]` / `qai todo delete <ID> [--json]`
- `TaskStorage.Remove(tasks []model.Task, id int) ([]model.Task, *model.Task, error)` を internal/storage/task.go に追加(既存 Add/Update と対称の形:スライスを受け取り残りのスライス+削除されたレコードを返す。既存の Add/Update/Save は変更しない)。コマンド側は Load→Remove→Save の順で呼ぶ。**注(設計レビュー反映): 現行コードの全書き込みは自己保存する Add()/Update() 経由であり、Save を外部から直接呼ぶのは本変更が初**。この形態の正当化根拠は「pomo 同形」ではなく **`idea delete --force` の cascade で N 件削除を Save 1 回に集約する**ため(部分保存による不整合窓を作らない)。done 側は既存 `Update()` を使う(外部 Save 不要)
- `todo delete`: 何れかの status のタスクを削除できるが、**子(親=このID)を持つタスクへの削除は ID 単位では行わない**(todo 側の親は idea 固定なので通常発生しない。発生したらエラー)
- `idea delete`: **デフォルトは子がある idea を拒否**(stderr に「N 個のサブタスク。--force でcascade削除」)。`--force` で subtree を再帰削除(削除した全レコードを logs に `EventStatusChange` の代替として記録…はしない — **削除イベントは現状の EventType 体系に存在せず、新 kind を増やさない**。削除は task スナップショット側から消える事実のみ。logs は不変)。この割り切りを Long に一行明記
- 出力: 散文 `Removed ... (ID: N)` / `--json` → **削除された task レコード**(削除後に存在しないことを `.[] .id` で確認できる契約。複数(cascade)なら配列で1行)
- 保護: tasks.yaml に存在しない ID → exit 1。自分自身を参照している親の破棄は考えない(現状のデータは常に整合)

### 3. ファイル構成
```
internal/storage/task.go   (MODIFY: Remove 追加のみ。Add/Update/Save は不変更)
cmd/                       (新規: todo_done.go, todo_delete.go, idea_delete.go —
                            各コマンド1ファイルは既存 cmd 配置の_house style。RunE はインラインで)
i18n/locales/locale_en-US.ini (各キー)
```
cmd 側3ファイルは構造がほぼ同一(Load→FindByID/検証→変更→Save→log→出力)。共通化はしない(3ファイルの類似 init/ガード行は duplication 30t 以下に収まる設計。抽象化で RunE を分割しない)。ただし `delete` の2コマンド間は**同じヘルパーを使わない**(親検証・cascade が非対称。無理な共通化を禁じる)

### 4. 契約の一般則(this change で決着)
- 書込み系3種(add/done/delete)すべて: `--json` は遷移**後**のレコード(削除は削除**前**のスナップショット)を stdout に1行。副作用の順序: storage Save 成功 → log Append → stdout 出力(出力前に永続化完了を保証。エージェントが続けて `list --json` で読める)
- SilenceUsage: true、SilenceErrors 未設定、exit 1 は main.go 伝播(既設)に委ねる
- **新 EventType を model に追加しない**(EventTaskDelete は作らない)。将来 logs へ削除を出したい時は本計画 Out of scope の再設計に回す

## Verification Plan(HOME 分離で実データを安全に保つ)
全機能検証は `TMP=$(mktemp -d); HOME=$TMP ...` の_scratch ストレージ_で行う(実 ~/.config/qai を触らない)。go run . は repo から実行し HOME だけ差し替える。

1. gofmt -l cmd internal / go build ./... / go vet ./... / go test ./...
2. storage.Remove の単体確認はコマンド経由(scratch HOME):`todo add --json "x"` → 返った id を `todo done <id> --json` → レコードの status=="done"、続けて `list --json | jq '[.todos[].id]'` に含まれない(done は list に出ない仕様)こと
3. scratch で `idea add --json "p"` → `todo add --json "c" --parent <pid>` → `idea delete <pid>` が**子ありで拒否**され exit 1 → `--force` で両方消え、`list --json` が `.ideas==[] && .todos==[]`
4. logs 整合: scratch で done 遷移後 `logs --json | jq '[.[]|select(.event_type=="status_change")]|length'` ≥1、`stats --json` の today/完了数が人間表示と突合一致
5. 存在しない ID: `todo done 9999 --json` → exit 1・stdout 空(エラーは stderr)/ done 済み ID の再 done も exit 1
6. 実データで従来コマンドの回帰無し: 実 HOME のまま `list`/`logs`/`stats` がレビュー済み Phase C までと同じ出力(実データへ add/done/delete を実行しない検証のみ)
7. 検証後の掃除: scratch は mktemp 配下のみ。`rm -rf $TMP` まで実行して完了

## Out of scope
- 削除イベントの logs 化(EventTaskDelete の新設)— 再設計時に logs 契約ごと扱う
- `idea→todo` の昇進コマンド(task-file-split.md の分割計画と絡む。次以降)
- 完了済み(done)の取消(`todo undone` 相当。要求が出てから)
- server ブランチ(gRPC)との統合 — 別系統のまま

## 実装のフェーズ分割(Qwen 小コンテキスト向け。info-commands-json.md と同じ流儀)
- **Phase A**: `qai todo done --json`(既存 Update() 経由+status_change ログ)のみ。cmd/todo_done.go + i18n
- **Phase B**: `TaskStorage.Remove` + `todo delete` + `idea delete(--force cascade)`(初外部 Save の書込み経路。A 完了・コミット後に着手)
- 1依頼=1フェーズのみ。「他フェーズは未承認」を毎回明記。レビュー観点: 検証2〜5が scratch HOME で完結していること(実 ~/.config/qai 不変)、done のログ欠落、cascade の Save 回数

---
*Rev1: 設計レビュー(Claude ペイン)反映 — 「pomo 保存パス同形」の誤りを訂正(全書込みは Add/Update 自己保存経由。外部 Save 初導入の根拠=cascade の Save 集約)。done は Update() 使用と明記。scratch HOME の初回 add は EnsureDirectories+MkdirAll で実測 OK(ブロッカーなし、判定:実装に進める)*
</content>

