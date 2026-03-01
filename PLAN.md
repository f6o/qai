# 次世代タスク管理ツール qai 仕様設計書 (改訂版)

## 1. コンセプト
* **DB-Centric Hybrid Storage**: 状態管理の正解（SSOT）は SQLite (`work.db` 等)。Markdown (`YYYY-MM-DD.md`) はその日の「タスク一覧のビュー」兼「活動ログ」として扱う。**DB 上の状態変更（qai pomo による着手や完了など）は、即座に Markdown ファイルへ反映（同期）される。**
* **LLM-Driven Refinement**: 曖昧なアイデアを LLM が分解し、DB へ直接タスク登録する。
* **Explicit Continuity**: 毎朝 `qai start` を実行することで、DB 内の未完了タスクを今日の日付ファイルへ一括展開する。

## 2. システム構成
* **言語**: Go
* **DB**: SQLite (Multi-Context Architecture)
    * 実行時のディレクトリ (CWD) に基づき、使用する DB ファイル (`work.db` / `hobby.db` 等) を自動的に切り替える。
* **UI**: CLI + TUI (Bubble Tea)

## 3. データモデル (SQLite)

### Todos (Ideas / Todo 管理)
ひとつのテーブルで「漠然としたアイデア」から「具体的な Todo」までを管理する。

| カラム | 型 | 説明 |
| :--- | :--- | :--- |
| `id` | INT | 一意識別子 (CLI 操作用) |
| `title` | TEXT | 内容 |
| `status` | STRING | idea / todo / doing / done |
| `priority` | INT | 優先度 (正数: 大きいほど高優先。DEFAULT 10, CHECK > 0)。TUIおよびMarkdownで優先度降順で表示・出力される。 |
| `parent_id`| INT | 親 ID。`qai todo add --parent` で手動指定、または将来の LLM 分解機能で自動設定。 |
| `started_at`| DATETIME | 作業開始日時 (最後に `pomo` でタスク選択した時刻) |
| `created_at`| DATETIME | 作成日時 |

### Logs (実績記録)
Todo ごとの作業ログや、汎用的な集中記録を保持する。

| カラム | 型 | 説明 |
| :--- | :--- | :--- |
| `id` | INT | 一意識別子 |
| `todo_id` | INT | 関連 Todo の ID (任意) |
| `content` | TEXT | ログ内容 |
| `duration` | INT | 作業時間 (分) |
| `logged_at`| DATETIME | 記録日時 |

## 4. コマンド体系 (Structured Subcommands)

### 共通 / ユーティリティ
| コマンド | 説明 | DB / 設定 |
| :--- | :--- | :--- |
| `qai init` | コンテキストの設定 | `~/.qairc` 作成・ポモドーロ設定 (25/5分) |
| `qai start` | 1日の開始（プランニング） | 今日の Markdown を生成 |
| `qai show [ID]` | 指定した Todo/Idea の詳細表示 | 全カラム情報の出力 |

### アイデア管理 (Ideas)
| コマンド | 説明 | DB 状態遷移 |
| :--- | :--- | :--- |
| `qai idea add "内容"` | 漠然としたアイデアの追加 | (新規) -> `idea` |
| `qai idea list` | アイデア一覧の表示 | `idea` の抽出 |

### Todo 管理
| コマンド | 説明 | DB 状態遷移 / 実績記録 |
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


## 6. 設定ファイル (`~/.qairc`) 案
```toml
# ポモドーロ基本設定
[pomodoro]
work_minutes = 25
break_minutes = 5

# コンテキスト設定 (ディレクトリと DB/Log の紐付け)
[[contexts]]
name = "work"
path = "~/workspace/work"
db_path = "~/qai/work.db"
log_dir = "~/workspace/work/logs"

[[contexts]]
name = "hobby"
path = "~/workspace/hobby"
db_path = "~/qai/hobby.db"
log_dir = "~/workspace/hobby/logs"
```

## 7. Markdown テンプレート案
```markdown
# 2026-03-01 (Sun)

## [アイデア (Ideas)]
- [ ] 1: 新機能のアイデアを練る
- [ ] 2: 旅行の計画を立てる

## [TODO]
- [ ] 3: 掃除機をかける (10)
- [/] 4: ゴミを出す (10) [doing]

## [やったこと]
- [x] 5: 昨日の日報を書く (09:15)

## [メモ]
```

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
* **シームレスな状態遷移**: タスクを選択すると、DB 上のステータスが自動的に `doing` に更新され、`started_at` カラムが現在時刻で設定される。以降、同じタスクを継続する場合は `started_at` は更新されず、最初の選択時刻を維持する。
* **Markdown 同期**: タスクの追加・状態変更（着手、完了など）が発生した際、当日の Markdown ファイルへ即座に反映される。`idea` / `todo` / `doing` / `done` いずれのステータスも同期対象。
* **自動記録**: タイマーが 1 セッション（デフォルト 25 分）経過するごとに、`Logs` テーブルへ「25分集中」の実績を自動で追加する。
* **休憩への遷移**: 1 セッション終了時、または `d` キーでタスクを完了させた際、タイマーの残り時間に関わらず一度セッションを終了し、休憩を開始するかどうかの選択画面を表示する（自動開始はしない）。
* **完了連携**: タイマー画面で `d` を押すと、その場でステータスを `done` に変更、経過時間を `Logs` へ保存し、休憩選択またはタスク選択画面に戻る。
* **中断制御**: `q` で中断。それまでの経過時間を保存するか選択可能。
* **通知機能**: セッション終了、および休憩終了時にシステムのデスクトップ通知、またはターミナルのベル音で通知を行う。
* **拡張性**: 将来的には 4 ポモドーロごとの「ロング休憩」への対応や、設定ファイルによる休憩時間のカスタマイズを容易にする設計とする。

