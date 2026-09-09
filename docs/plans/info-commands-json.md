# Implementation Plan: 情報確認系コマンド全体のエージェント対応(`--json`)

## Objective
書き込み側(add)と `qai list` で確立した機械可読契約を、残りの情報確認系コマンドへ波及させる。Qwen ペインのコンテキストが小さいため、**フェーズ分割で逐次実装**する(各フェーズ独立にレビュー・コミット可能)。

## 共通契約(全フェーズ共通の憲法。add/list で既確立済みの方針の継続)

- `--json` 時は stdout へ **JSON 1 行のみ**、nil/空は常に**空配列**(null 不可)。出力直前に `if x == nil { x = []model.Task{} }`(logs なら `[]model.Log{}`)を代入して正規化
- 各コマンドに `SilenceUsage: true`。**`SilenceErrors` は設定しない**(`Error:` は cobra が stderr へ)。exit 伝播は main.go 側で済んでいる
- task レコードは `model.Task`(既存 `MarshalJSON`)、log レコードは `model.Log`(json タグ所有済み)をそのまま利用。**DTO・新ファイル・モデル変更は禁止**
- 匿名 struct を `if` ブランチ内に閉じる(list と同形)— **例外: Phase C のみ適用外**。`statsData` は人間/JSON 両レンダラが参照するため RunE 本体(if の外)に置く
- 人間表示(フラグなし)の出力は一字も変えない
- `Long`/`Example` は最小限(verbosity gate 超過を避ける。100 行超えが見えたら分岐内の計算を小さなヘルパー関数へ切り出す)
- import は `encoding/json` と(必要なら)`model` のみ追加
- **検証の数値アサーションは実データ非依存にする**: tasks.yaml の件数は人の操作で変わるため、`length==N` 型のハードコードは禁止。構造検証(jq)と人間表示との動的突合で検証する

## Phase A: `idea list` / `todo list`(S)

### Output contract
- `qai idea list --json` → `[{"id":1,"title":"…","status":"idea",…}, …]`(**トップレベルが配列**。このスコープの view は 1 kind なのでオブジェクトで包まない)
- `qai todo list --json` → 同じく配列。**`idea/todo list` には `-A/--above` を追加しない**(top-level `list` には実装済みだが、本計画のスコープ外=計画にない API 増加禁止のため)

### Steps
1. `cmd/idea_list.go`: `var ideaListJSON bool` + `Flags().BoolVar` + `SilenceUsage: true` + 最小 Long/Example。**JSON 分岐は空チェックの散文 early-return より前**に置く(`if len(ideas)==0 { prose; return nil }` の内側に入れると空のとき散文になる — 現行コードの陷阱)
2. `cmd/todo_list.go`: 同上。JSON 分岐は **sort 適用後**に置く
3. i18n: `cmd.idea_list.flag_json` / `cmd.todo_list.flag_json`(各1行、説明のみ)

### Verification A
1. gofmt / build / vet / test
2. `go run . idea list --json | jq -e 'type=="array" and all(.status=="idea")'`、`go run . todo list --json | jq -e 'type=="array" and all(.status=="todo" or .status=="doing")'`
3. 人間表示との突合(件数非依存): `diff <(go run . idea list --json | jq -r '.[].id') <(go run . idea list | grep -oE '^  \[[0-9]+\]' | tr -d ' []')` が空であること。todo list も同理(id 列のみ比較)
4. 空配列契約の実行時検証(**HOME へ生データを書き換えない方法**): `HOME=$(mktemp -d) go run . idea list --json | jq -e '. == []'` と `HOME=$(mktemp -d) go run . todo list --json | jq -e '. == []'` が通ること(生 tasks.yaml に触れない。CWD 直下の config.toml は上書きが利かないため使わないこと)
5. フラグなし出力不変(実データで `go run . idea list` を実行し従来フォーマットであることを確認)

## Phase B: `logs`(S)

### Output contract
- `qai logs --json` → `[{"id":1,"todo_id":20,"logged_at":"…","event_type":"focus_complete",…}, …]`、`--type` フィルタ適用後
- **event_type の正規化**: `model.Log.EventType` は omitempty で未設定時キーが消える。人間表示は `EffectiveEventType()`(focus_complete へフォールバック)のため、JSON へ渡す前にメモリ上のコピーで `EventType` を埋めてから出力(**ストレージは書かない**)。契約は「event_type は常に存在」
- duration は `*int` のまま(なし=キー欠落を許容。ゼロと区別したい语义なので埋めない)

### Steps
1. `cmd/logs.go`: `var logsJSON bool` + 登録 + `SilenceUsage: true` + Long/Example
2. 正規化は**ファイル内ヘルパーとして定義**(Phase C の today.logs との重複=clone を避けるため、インラインコピーにしない。新ファイル不要、`cmd/logs.go` 内に置く):

```go
func normalizeEventTypes(logs []model.Log) []model.Log {
	out := make([]model.Log, len(logs))
	for i, l := range logs {
		l.EventType = l.EffectiveEventType()
		out[i] = l
	}
	return out
}
```

3. RunE: `--type` フィルタ適用後、散文 early-return の前に JSON 分岐:

```go
		if logsJSON {
			return json.NewEncoder(cmd.OutOrStdout()).Encode(normalizeEventTypes(logs))
		}
```

(`normalizeEventTypes` は len 0 でも nil ではない `[]model.Log{}` を返すので空契約は成立)

### Verification B
1. gofmt / build / vet / test
2. `go run . logs --json | jq -e 'type=="array" and all(.event_type != null)'`
3. `go run . logs --json --type focus_complete | jq 'length'` が人間表示 `go run . logs --type focus_complete | grep -c '^  \['` と一致
4. 空契約の実行時検証(**実データ非依存。モデルに実在する type 名を「存在しない」前提で使わないこと** — `focus_quit` は実データに存在するため旧案は却下): `HOME=$(mktemp -d) go run . logs --json | jq -e '. == []'` が通ること(A-4/C-6 と同型の HOME 方式に統一)
5. フラグなし出力不変

## Phase C: `stats`(M・リファクタリング含む)

### Output contract

```json
{"tasks":{"total":0,"ideas":0,"todos":0,"done":0},
 "focus":{"sessions":0,"total_minutes":0},
 "today":{"date":"2006-01-02形式の文字列","sessions":0,"focus_minutes":0,"logs":[]}}
```

(値は実データ由来。上記は**型・キーの定義**であり実例の数字ではない)
- `today.logs` は Phase B の `normalizeEventTypes` を再利用。空なら `[]`
- `date` は `"2006-01-02"` 形式。**エージェントがローカルタイムゾーンを再推定できない**ため明示
- 各数値は現行 RunE と同一の計算式を維持

### Steps
1. `cmd/stats.go` の RunE を**計算と描画に分離**。計算は `statsData` に1回だけ寄せる(if の外。共通契約の匿名structルールの適用外は明記済み)。型は以下で固定する:

```go
		var statsData struct {
			Tasks struct {
				Total int `json:"total"`
				Ideas int `json:"ideas"`
				Todos int `json:"todos"`
				Done  int `json:"done"`
			} `json:"tasks"`
			Focus struct {
				Sessions     int `json:"sessions"`
				TotalMinutes int `json:"total_minutes"`
			} `json:"focus"`
			Today struct {
				Date         string      `json:"date"`
				Sessions     int         `json:"sessions"`
				FocusMinutes int         `json:"focus_minutes"`
				Logs         []model.Log `json:"logs"`
			} `json:"today"`
		}
```

		現行の各計算(`len(tasks)`、doneCount、`totalMinutes`、today系)は statsData への代入として一度だけ書き、人間表示側の `cmd.Printf` は statsData の値を参照する(**計算ロジック自体は変更しない。移動と参照のみ**)。`statsData.Today.Logs` には `normalizeEventTypes(todayLogs)` を格納し、人間表示側の today ログループも同じ値を参照してよい(表示は現状と同一フィールドのみを使うこと)
2. `statsJSON bool` + 登録 + `SilenceUsage: true` + Long/Example。`--json` 時は statsData をそのまま Encode(計算済みなので分岐は3行)。`Today.Logs` は nil になり得ない(`make` 由来 or 正規化済み)が、防御的に `if statsData.Today.Logs == nil { statsData.Today.Logs = []model.Log{} }` を Encode 直前に置く
3. i18n: `cmd.stats.flag_json`

### Verification C
1. gofmt / build / vet / test
2. `go run . stats --json | jq -e '(.tasks|has("total") and has("ideas") and has("todos") and has("done")) and (.focus|has("sessions") and has("total_minutes")) and (.today|has("date") and has("logs"))'`
3. 人間表示との動的突合: `go run . stats --json | jq -r '.focus.sessions'` と `go run . stats` のセッション数表示が一致(同様に total_minutes、today の各値。件数のハードコード禁止)
4. today.logs のサブセット突合: `diff <(go run . stats --json | jq -r '.today.logs[].id' | sort) <(go run . logs --json | jq -r --arg d "$(date +%F)" '[.[] | select(.logged_at | startswith($d))][] | .id' | sort)` が空であること
5. フラグなし出力がリファクタ前後で一致(`go run . stats` の出力を目視/前出力保存と比較)
6. HOME=$(mktemp -d) での空データ検証: `HOME=$(mktemp -d) go run . stats --json | jq -e '.today.logs == [] and .tasks.total == 0'` が通ること(tz による日付ズレを date キーが正当に吸収することも確認)

## 対象外(方針の記録)
- `md`: Markdown 自体がエージェント可読。契約化しない
- `timer`: 対話専用・副作用(pomo/lock)。エージェントからは呼ばせない。各 Long に「非対話では使わないこと」は書かない(フェーズスコープ外でのファイル接触を増やさない)
- `init`: 一回セットアップ。`--json` の语义が薄い

## 実装・レビューの流れ
フェーズ A → B → C の順に**別々に** Qwen へ依頼(1依頼=1フェーズのみ、「計画書の他フェーズは未承認」と明記)。各フェーズ完了ごとにレビュー→コミット。レビュー観点: 共通契約の遵守(空=配列、早期散文return より前、モデル無変更)、人間出力の不変性、検証のアサーションが実データ非依存であること。

---
*Rev2: 設計書レビュー(Claude ペイン)反映 — B1 jq構文修正、B2 件数ハードコード排除、R1 HOME=$(mktemp -d) による空検証、R2 normalizeEventTypes 共有ヘルパー化、R3 -A 実装済みの事実修正、R4 匿名struct例外の明記、R5 statsData 具体化、Q1 サブセット比較コマンド明記*
*Rev2-1: 再レビュー反映 — B-4 の `--type focus_quit`(実在するため空にならない)を HOME 方式へ統一。再レビューで A/C の検証・statsData 型定義は実測確認済み、判定「Phase A は実装可」*
</content>
