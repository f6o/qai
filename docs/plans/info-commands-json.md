# Implementation Plan: 情報確認系コマンド全体のエージェント対応(`--json`)

## Objective
書き込み側(add)と `qai list` で確立した機械可読契約を、残りの情報確認系コマンドへ波及させる。Qwen ペインのコンテキストが小さいため、**フェーズ分割で逐次実装**する(各フェーズ独立にレビュー・コミット可能)。

## 共通契約(全フェーズ共通の憲法。add/list で既確立済みの方針の継続)

- `--json` 時は stdout へ **JSON 1 行のみ**、nil/空は常に**空配列**(null 不可)。出力直前に `if x == nil { x = []model.Task{} }`(logs なら `[]model.Log{}`)を代入して正規化
- 各コマンドに `SilenceUsage: true`。**`SilenceErrors` は設定しない**(`Error:` は cobra が stderr へ)。exit 伝播は main.go 側で済んでいる
- task レコードは `model.Task`(既存 `MarshalJSON`)、log レコードは `model.Log`(json タグ所有済み)をそのまま利用。**DTO・新ファイル・モデル変更は禁止**
- 人間表示(フラグなし)の出力は一字も変えない
- 匿名 struct は `if` ブランチ内に閉じる(list と同形)
- `Long`/`Example` は最小限(verbosity gate 超過を避ける。100 行超えが見えたら分岐内の計算を小さなヘルパー関数へ切り出す)
- import は `encoding/json` と(必要なら)`model` のみ追加

## Phase A: `idea list` / `todo list`(S)

### Output contract
- `qai idea list --json` → `[{"id":1,"title":"…","status":"idea",…}, …]`(**トップレベルが配列**。このスコープの view は 1 kind なのでオブジェクトで包まない)
- `qai todo list --json` → 同じく配列。**top-level `list` での `-A` は未実装なので、この2コマンドにも優先度フィルタを新設しない**(計画にない API 増加禁止)

### Steps
1. `cmd/idea_list.go`: `var ideaListJSON bool` + `Flags().BoolVar` + `SilenceUsage: true` + 最小 Long/Example。**JSON 分岐は空チェックの散文 early-return より前**に置く(`if len(ideas)==0 { prose; return nil }` の内側に入れると空のとき散文になる — 現行コードの陷阱)
2. `cmd/todo_list.go`: 同上。JSON 分岐は **sort 適用後**に置く
3. i18n: `cmd.idea_list.flag_json` / `cmd.todo_list.flag_json`(各1行、説明のみ。フォーマット違いなし)

### Verification A
1. gofmt / build / vet / test
2. `go run . idea list --json | jq -e 'type=="array" and length==2'`(現状 idea は id 1, 6 の2件)
3. `go run . todo list --json | jq -e 'type=="array"'`、id 集合が人間表示と一致・priority 降順
4. 空契約: 現状データでは作れないので**データを書き換えず** `jq` での null 回避はコード確認で代替(正規化が散文 early-return より前かを expand で確認)
5. フラグなし出力不変

## Phase B: `logs`(S)

### Output contract
- `qai logs --json` → `[{"id":1,"todo_id":20,"logged_at":"…","event_type":"focus_complete",…}, …]`、`--type` フィルタ適用後
- **event_type の正規化**: `model.Log.EventType` は omitempty で未設定時キーが消える。人間表示は `EffectiveEventType()`(focus_complete にフォールバック)で出力しているため、JSON へ渡す前に各レコードの `EventType` を `EffectiveEventType()` の結果で埋めてから出力する(**ストレージは書かない**。メモリ上のコピーへの正規化)。契約は「event_type は常に存在」
- duration は `*int` のまま(なし=キー欠落を許容。ゼロと区別したい语义なので埋めない)

### Steps
1. `cmd/logs.go`: `var logsJSON bool` + 登録 + `SilenceUsage: true` + Long/Example
2. RunE: `--type` フィルタ適用後、散文 early-return の前に JSON 分岐。正規化は以下:

```go
		if logsJSON {
			out := make([]model.Log, len(logs))
			for i, l := range logs {
				l.EventType = l.EffectiveEventType()
				out[i] = l
			}
			return json.NewEncoder(cmd.OutOrStdout()).Encode(out)
		}
```

### Verification B
1. gofmt / build / vet / test
2. `go run . logs --json | jq -e 'type=="array" and all(.event_type != null)'`
3. `go run . logs --json --type focus_complete | jq 'length'` が人間表示 `logs --type focus_complete` の行数と一致
4. `jq` で logged_at/todo_id のキー名が `model.Log` の json タグどおりであることを1レコード確認
5. フラグなし出力不変

## Phase C: `stats`(M・リファクタリング含む)

### Output contract

```json
{"tasks":{"total":26,"ideas":2,"todos":4,"done":18},
 "focus":{"sessions":12,"total_minutes":300},
 "today":{"date":"2026-09-09","sessions":1,"focus_minutes":25,"logs":[…Log records…]}}
```

- `today.logs` は Phase B と同一の正規化(EffectiveEventType 適用)を適用。空なら `[]`
- `date` は RFC3339 日付部(`"2006-01-02"`)。**エージェントがローカルタイムゾーンを再推定できない**ため明示
- 各数値は人間表示と同じ計算式(from: 現行 RunE のロジック)を維持

### Steps
1. `cmd/stats.go` の RunE を**計算と描画に分離**(`statsData` 匿名 struct へ寄せて両レンダラが読む形。これが「計算を分岐間で重複させない」= clone 回避の本題でもある):
   - tasks{total,ideas,todos,done} / focus{sessions,totalMinutes} / today{date,sessions,minutes,logs} を1回だけ計算
2. `statsJSON bool` + 登録 + `SilenceUsage: true` + Long/Example。`--json` 時はオブジェクト1個を Encode、それ以外は既存散文描画(数値・並びは現状と同一であること)
3. i18n: `cmd.stats.flag_json`

### Verification C
1. gofmt / build / vet / test
2. `go run . stats --json | jq -e '.tasks.total and (.tasks|has("ideas" and "todos" and "done")) and .today|has("date")'`(jq 構文は実装側で正しく調整)
3. 人間表示 `stats` と数値を突き合わせ(total/sessions/focus_minutes/date)
4. `stats --json` 出力の `today.logs` が `logs --json` の同日サブセットと一致
5. フラグなし出力がリファクタ前後で**バイト一致**(リダイレクト不可なら目視 diff 可)

## 対象外(方針の記録)
- `md`: Markdown 自体がエージェント可読。契約化しない
- `timer`: 対話専用・副作用(pomo/lock)。エージェントからは呼ばせない。各 Long に「非対話では使わないこと」は書かない(フェーズスコープ外でのファイル接触を増やさない)
- `init`: 一回セットアップ。`--json` の语义が薄い

## 実装・レビューの流れ
フェーズ A → B → C の順に**別々に** Qwen へ依頼(1依頼=1フェーズのみ、「計画書の他フェーズは未承認」と明記)。各フェーズ完了ごとにレビュー→コミット。レビュー観点: 共通契約の遵守(空=配列、早期散文return より前、モデル無変更)と人間出力の不変性。
</content>
