# 次世代タスク管理ツール qai 仕様設計書 (改訂版)

## 1. コンセプト
* **File-Centric Hybrid Storage**: 状態管理の正解（SSOT）はファイルベース（YAML + JSONL）。Markdown (`YYYY-MM-DD.md`) はその日の「タスク一覧のビュー」兼「活動ログ」として扱う。**ファイル上の状態変更（qai pomo による着手や完了など）は、即座に Markdown ファイルへ反映（同期）される。**
* **LLM-Driven Refinement**: 曖昧なアイデアを LLM が分解し、ファイルを介してタスク登録する。
* **Explicit Continuity**: 毎朝 `qai start` を実行することで、未完了タスクを今日の日付ファイルへ一括展開する。

## 2. システム構成
* **言語**: Go
* **Storage**: File-based
    * タスクとログは `~/.config/qai` ディレクトリに保存。
    * **Tasks (YAML)**: 頻繁に更新・参照されるタスク状態の管理。
    * **Logs (JSONL)**: 追記のみの実績・作業記録。
* **UI**: CLI + TUI (Bubble Tea)

## 3. データモデル (Hybrid File Storage)

### Tasks (`tasks.yaml`)
タスクとアイデアを統合管理。人間が読み書きしやすい YAML 形式。
ID はファイル内の最大 ID + 1 で自動採番する (削除済みIDの再利用はしない)。

```yaml
- id: 1
  title: "次世代タスク管理ツールの設計"
  status: "doing"
  priority: 20
  parent_id: null
  started_at: "2026-03-02T10:00:00Z"
  created_at: "2026-03-01T09:00:00Z"

- id: 2
  title: "SQLiteからファイルベースへの移行検討"
  status: "todo"
  priority: 15
  parent_id: 1
  created_at: "2026-03-02T10:05:00Z"
```

### Logs (`logs.jsonl`)
実績記録。高速な追記と堅牢性を重視した JSON Lines 形式（1行1レコード）。

```json
{"id": 1, "todo_id": 1, "content": "25min focus", "duration": 25, "logged_at": "2026-03-02T10:25:00Z"}
{"id": 2, "todo_id": 2, "content": "Task completed", "duration": 10, "logged_at": "2026-03-02T10:40:00Z"}
```

## 4. コマンド体系 (Structured Subcommands)

### 共通 / ユーティリティ
| コマンド | 説明 | ストレージ / 設定 |
| :--- | :--- | :--- |
| `qai init` | 初期化 (設定ファイル作成) | `~/.config/qai/config.toml` 作成・データディレクトリ準備 |
| `qai start` | 1日の開始（プランニング） | 今日の Markdown を生成 |
| `qai preview` | 今日の Markdown 全体を表示 | qai start 出力と同形式 |
| `qai preview [ID]` | 指定 ID の詳細表示 | ID が Ideas の場合は Markdown 形式 (Ideas + 子タスク)、それ以外は title のみ表示 |

### アイデア管理 (Ideas)
| コマンド | 説明 | 状態遷移 |
| :--- | :--- | :--- |
| `qai idea add "内容"` | 漠然としたアイデアの追加 | (新規) -> `idea` |
| `qai idea list` | アイデア一覧の表示 | `idea` の抽出 |

### Todo 管理
| コマンド | 説明 | 状態遷移 / 実績記録 |
| :--- | :--- | :--- |
| `qai todo add "内容" [--parent ID]` | 具体的な Todo を直接追加。`--parent` で親 Idea を指定可能。 | (新規) -> `todo` |
| `qai todo list` | 今日やるべき Todo の一覧 | `todo` / `doing` の抽出 |

### 集中・ポモドーロ
| コマンド | 説明 | 備考 |
| :--- | :--- | :--- |
| **`qai pomo`** | **集中モード (TUI) の起動** | **タスク選択、タイマー、完了を統合管理** |

### 実績・ログ
| コマンド | 説明 | 備考 |
| :--- | :--- | :--- |
| `qai report` | 振り返りレポート | 全実績の集計 |

## 5. 運用ルール
* **Ideas と Todo の分離**:
    - **アイデア (Ideas)**: 漠然とした願いやバックログ。`qai idea add` でここに入る。
    - **Todo**: 今日取り組む具体的な項目。`qai todo add` でここに入る。
* **ワークフロー**:
    1. `qai idea add "家を綺麗にする"` (将来やりたいこと)
    2. `qai todo add "掃除機をかける"` (今日やること)
    3. `qai pomo` -> **TUI でタスク「掃除機をかける」を選択して集中開始。**
    4. タイマー終了、または `d` キーでタスク完了。ログが自動保存される。


## 6. 設定ファイル (`~/.config/qai/config.toml`) 案
```toml
# ポモドーロ基本設定
[pomodoro]
work_minutes = 25
break_minutes = 5

# データ保存先
[data]
todofile = "~/.config/qai/tasks.yaml"
logfile = "~/.config/qai/logs.jsonl"
```

## 7. Markdown テンプレート案
```markdown
# 2026-03-05 (Tue)

## アイディア1

- [x] 完了したタスク
- [/] 着手中のタスク
- [ ] 未着手のタスク1
- [ ] 未着手のタスク2
- [ ] 未着手のタスク3

## アイディア2

タスクに分解されていません。

## アイディア3

タスクに分解されていません。

## 雑多なタスク

- [x] 雑多なタスク2
- [ ] 雑多なタスク1
- [ ] 雑多なタスク3
```

## 実装検討事項

### 通知機能
* セッション終了・休憩終了時の通知方式
    * ターミナルベル (標準出力で `\a` または tty チェック)
    * macOS デスクトップ通知 (`osascript` または notify-rust 等)
    * 設定で切り替え可能にすることが望ましい
* 実装優先度: ターミナルベル → デスクトップ通知の順で検討

## 8. 集中モード (qai pomo) 仕様
`qai pomo` 実行時に表示される TUI 画面の動作とインターフェース。

### ステップ 1: タスク選択
起動後、`todo` / `doing` ステータスのタスク一覧を優先度降順で表示。
```text
? Select a task to focus on:
> [ ] 3: 掃除機をかける
  [ ] 4: ゴミを出す
  [ ] 6: ドキュメント作成
```

### ステップ 2: 集中タイマー
タスク選択後、ポモドーロタイマーが開始。
```text
Focusing on: [ID] 掃除機をかける
[██████████░░░░░░░░░] 15:20 / 25:00
(q) quit | (d) done | (p) pause | (s) skip session
```

### ステップ 3: 休憩選択 / 休憩タイマー
集中セッション終了後、またはタスク完了後に表示。**自動では開始せず、ユーザーの操作を待機する。**
```text
Great job! Start a short break?
(b) Start Short Break (5:00) | (n) Next Task | (q) Quit
```
休憩開始後：
```text
Short Break
[███████████████████] 05:00 / 05:00
(s) Skip break | (q) Quit
```

### ステップ 4: 休憩終了後の遷移
休憩終了後に表示。
```text
Break finished. What's next?
(c) Continue same task | (n) Next task | (q) Quit
```

### 振る舞い
* **シームレスな状態遷移**: タスクを選択すると、ファイル上のステータスが自動的に `doing` に更新され、`started_at` が現在時刻で設定される。
* **Markdown 同期**: タスクの追加・状態変更が発生した際、当日の Markdown ファイルへ即座に反映される。
* **自動記録**: タイマーが 1 セッション経過するごとに、`logs.jsonl` へ実績を自動で追記する。
* **休憩への遷移**: 1 セッション終了時、または `d` キーでタスクを完了させた際、タイマーの残り時間に関わらず一度セッションを終了し、休憩開始の選択画面を表示する。
* **完了連携**: タイマー画面で `d` を押すと、その場でステータスを `done` に変更、経過時間を `logs.jsonl` へ保存し、休憩選択またはタスク選択画面に戻る。
* **中断制御**: `q` で中断。それまでの経過時間を保存するか選択可能。
* **通知機能**: セッション終了、および休憩終了時にシステムのデスクトップ通知、またはターミナルのベル音で通知を行う。
* **拡張性**: 将来的には 4 ポモドーロごとの「ロング休憩」への対応や、設定ファイルによる休憩時間のカスタマイズを容易にする設計とする。
