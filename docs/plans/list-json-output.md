# Implementation Plan: `qai list` への `--json` 出力

## Objective
書き込み側(`todo/idea add --json`、docs/plans/add-json-output.md 参照)が完成したので、読み取り側 `qai list` にも機械可読契約を与える。エージェントの「追加→状態確認」ループを散文パースなしで完結させる。

## Output contract

`--json` 指定時は stdout へ **JSON 1 行のみ**:

```json
{"ideas":[{...task...}],"todos":[{...task...}]}
```

- task レコードは `model.Task`(既存の `MarshalJSON` を再利用。キーは add と同一契約)
- 人間表示と同一パイプラインの**適用後**を返す: `-A/--above` フィルタ → todos の priority 降順ソート(並び・絞り込みを再実装しない。既存の変数 `ideas`/`todos` をそのまま載せる)
- **空は空配列**(`{"ideas":[],"todos":[]}`, exit 0)。`FilterIdeas`/`FilterTodos` が nil を返すと `"ideas":null` になり契約違反 → 出力直前に nil を `[]model.Task{}` へ正規化すること(`cmd.list.empty` 散文は出さない)
- `--json` と `-A` の同時指定は有効な合成。フラグなしの出力は従来と一字も変わらないこと
- listCmd にも `SilenceUsage: true`(エージェント出力に usage を混ぜない方針の統一。他コマンドは触らない)

## Implementation Steps

### 1. `cmd/list.go`(変更はこの1ファイル+下記 i18n のみ)

1. `var listJSON bool` を追加、`init()` で `listCmd.Flags().BoolVar(&listJSON, "json", false, i18n.T("cmd.list.flag_json"))`、`listCmd` に `SilenceUsage: true`
2. `Long`/`Example` を簡潔に追加(数行で。コマンドリテラルの過膨張を避ける — add 側で verbosity 指摘済みなので最小限に留める)
3. `RunE` の `-A` フィルタ・ソート適用後、`len(ideas) > 0` 等の表示ロジックへ分岐を入れる:

```go
		if listJSON {
			if ideas == nil {
				ideas = []model.Task{}
			}
			if todos == nil {
				todos = []model.Task{}
			}
			return json.NewEncoder(cmd.OutOrStdout()).Encode(struct {
				Ideas []model.Task `json:"ideas"`
				Todos []model.Task `json:"todos"`
			}{ideas, todos})
		}
```

- 匿名 struct を `if` 内に閉じるので DTO 新設・新ファイルは不要(add と同じ方針)
- import に `encoding/json` を追加(`model` は `-A` 実装済みで既にある)

### 2. i18n (`i18n/locales/locale_en-US.ini`)

```ini
cmd.list.flag_json = Output exactly one JSON line: {"ideas":[...],"todos":[...]} after filters
```

## Out of scope

- `logs --json` / 単一タスク詳細の機械出力(次点)
- 既存 add コマンド・model への変更(MarshalJSON は再利用するだけ。触らない)

## Verification Plan

1. `gofmt -l cmd` 空 / `go build ./...` / `go vet ./...` / `go test ./...`
2. `qai list --json | jq -e 'has("ideas") and has("todos")'` が通る
3. 空正規化: 現在の tasks.yaml は priority 最大 30 なので `qai list -A 50 --json` → `jq -e '.ideas == [] and .todos == []'`(null ではなく空配列。exit 0)
4. `-A 10 --json | jq -r '.todos[].id'` の id 集合が人間表示 `qai list -A 10` の todo 行と一致し、priority 降順であること
5. フラグなしの出力が実装前(`git stash` 比較でなく目視で可)と同一であること
6. list は読み取りのみなので tasks.yaml の掃除は不要(検証でデータを変更するコマンドを実行しないこと)
</content>
