# Implementation Plan: `qai todo/idea add` への `--json` 出力とエージェント向け `--help` 強化

## Objective
AI コーディングツール(opencode / Claude Code 等)が `todo add` / `idea add` を
安全に非対話実行できる「成功契約」を与える。現状の弱点は:

- 成功出力が `Added todo: <title> (ID: 25)` という i18n 散文で、新 ID が機械抽出できない(`--parent` チェーンが不安定)
- `--start`(pomo 開始)・`-i`(stdin プロンプト)の副作用・非互換が help から読み取れない

本計画は cobra の `Long`/`Example` 整備(L1)と `--json`(L2)まで。ドキュメント生成系(L3)は対象外。

## Output contract

`--json` 指定時は stdout へ **JSON 1 行のみ**を出力し、既存の散文メッセージは出さない。スキーマは `model.Task` の json タグのまま(house pattern: model が自身の JSON 表現を所有しており、`model.Log` が storage で直接 Marshal されている前例に従う)。DTO・`kind` フィールドは新設しない — `status`("todo"/"idea")が kind を兼ねる:

```json
{"id":25,"title":"hoge","status":"todo","priority":10,"created_at":"2026-09-08T21:16:39.6+09:00"}
```

- 注意: `parent_id` は `*int` の `omitempty` で未設定時キー自体が存在しない。`started_at` は `time.Time` のため encoding/json の `omitempty` が利かず、`model.Task.MarshalJSON`(aux struct で `*time.Time` として shadow し、zero 値時に省略)で対応した
- 失敗時は exit 1 + stderr(cobra 既定。stdout には何も出さない)
- `-i` との併用は **エラー終了**(プロンプトが stdout を汚染するため。stderr に理由メッセージ)
- `--start` との併用は可。JSON 1 行を先に出力してから pomo を開始する

## Implementation Steps

### 1. `cmd/todo_add.go`

1. `var todoAddJSON bool` を追加し `init()` で `todoAddCmd.Flags().BoolVar(&todoAddJSON, "json", false, i18n.T("cmd.todo_add.flag_json"))`
2. `RunE` 冒頭で `if todoAddJSON && interactive { return errors.New(...) }`(`-i` 併用エラー)
3. 成功時出力を分岐(新ファイル・ヘルパーは作らない。`encoding/json` を import するのみ):

```go
		if todoAddJSON {
			if err := json.NewEncoder(cmd.OutOrStdout()).Encode(task); err != nil {
				return err
			}
		} else {
			cmd.Println(i18n.T("cmd.todo_add.success", task.Title, task.ID))
		}
```

(`cmd.Println` 系は `OutOrStdout()` 経由なのでテスト容易なまま)
4. `Long` と `Example` を追加。内容は「非対話で使える最小契約」を明示:

```
Long: `Add a new todo non-interactively. Content is a required single argument.
Side effects: --start launches a Pomodoro session (writes locks under ~/.config/qai).
--interactive reads prompts from stdin and is not usable in scripts (and refuses --json).
Use --json to get exactly one machine-readable line of the task record.
Keys parent_id/started_at are omitted when unset.`
Example: `  qai todo add "write tests"
  qai todo add --json "write tests" | jq -r .id
  ID=$(qai idea add --json "new feature" | jq -r .id)
  qai todo add --json "step 1" --parent "$ID"`
```

### 2. `cmd/idea_add.go`

同上(`--json` のみ。`-i`/`--start` はないので併用エラー的分岐は不要)。`Long`/`Example` は idea 向けに簡略版。

### 3. i18n (`i18n/locales/locale_en-US.ini`)

```ini
cmd.todo_add.flag_json = Output exactly one JSON line (stable machine-readable contract)
cmd.idea_add.flag_json = Output exactly one JSON line (stable machine-readable contract)
cmd.todo_add.err_json_interactive = --json cannot be combined with --interactive
```

## Out of scope(次以降の候補として記録)

- 他コマンド(`list`/`logs` 等)への `--json` 波及。波及して共通処理(出力前のチェック等)が生まれた時点で、エンコーダの `internal/` 化か `model.Task` のメソッド化を検討する(今は直 Encode で十分)
- delete/complete 系サブコマンド(現状 add/list のみで、エージェント側で成果物の掃除ができない)
- tasks.yaml 直編集禁止を明文化するエージェント向けマニュアルの生成(L3)

## Verification Plan

1. `gofmt -l cmd` が空 / `go build ./...` / `go test ./...`
2. `qai todo add --json "probe"` → stdout が JSON 1 行で `jq -e '.id and .status == "todo" and .created_at'` が通る(`parent_id` は未設定なのでキーなし)
3. `qai idea add --json "probe-parent" | jq -r .id` → その ID で `qai todo add --json "child" --parent <ID>` → JSON の `parent_id` が一致
4. `qai todo add "probe"`(フラグなし)は従来どおり `Added todo: ... (ID: N)` の散文
5. `qai todo add -i --json "x"` → exit 1、stdout 出力なし
6. 検証で追加したタスクは tasks.yaml の末尾から手動削除する(delete コマンドが存在しないため)。データ本体の書式を壊さないこと
</content>

## 結果(実装済み・レビュー合格・作業ツリー)

実装中に計画の前提と異なる2点が発覚し、承認のうえスコープを拡大して修正した:

1. **失敗時契約**: 「cobra 既定で exit 1 + stdout 空」は誤りだった(`main.go` が `cmd.Execute()` のエラーを無視して全コマンドが exit 0、usage が stdout に出ていた)。→ `main.go` で `os.Exit(1)` 伝播 + add コマンド2つのみ `SilenceUsage: true`(`SilenceErrors` は未設定。`Error:` は cobra が stderr へ)
2. **started_at 漏れ**: 上記のとおり `MarshalJSON` を `internal/model/task.go` に追加

結果、変更は5ファイル(cmd/todo_add.go / cmd/idea_add.go / i18n ini / internal/model/task.go / main.go)となった。`--quality-delta` の `todoAddCmd` verbosity(82→101)は cobra インライン RunE の house style に伴う既知のトレードオフとして ack 相当。検証6項目は実装側・レビュー側で独立に再実行済み。
