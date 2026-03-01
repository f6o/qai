# 次世代タスク管理ツール qai 仕様設計書 (改訂版)

## 1. コンセプト
* **DB-Centric Hybrid Storage**: 状態管理の正解（SSOT）は SQLite (`qai.db`)。Markdown (`YYYY-MM-DD.md`) はその日の「タスク一覧のビュー」兼「活動ログ」として扱う。
* **LLM-Driven Refinement**: 曖昧なアイデアを LLM が分解し、DB へ直接タスク登録する。
* **Explicit Continuity**: 毎朝 `qai start` を実行することで、DB 内の未完了タスクを今日の日付ファイルへ一括展開する。

## 2. システム構成
* **言語**: Go
* **DB**: SQLite (Multi-Context Architecture)
    * 実行時のディレクトリ (CWD) に基づき、使用する DB ファイル (`work.db` / `hobby.db` 等) を自動的に切り替える。
* **UI**: CLI + TUI (Bubble Tea)
* **外部連携**: Gemini API / OpenAI API

## 3. データモデル (SQLite)

### Todos (Ideas / Todo 管理)
ひとつのテーブルで「漠然としたアイデア」から「具体的な Todo」までを管理する。

| カラム | 型 | 説明 |
| :--- | :--- | :--- |
| `id` | INT | 一意識別子 (CLI 操作用) |
| `title` | TEXT | 内容 |
| `status` | STRING | idea / todo / doing / done |
| `priority` | INT | 優先度 (正数: 大きいほど高優先。DEFAULT 10, CHECK > 0) |
| `parent_id`| INT | 親 ID (アイデアを分解した場合の紐付け用) |
| `started_at`| DATETIME | 作業開始日時 (最後に `work` を叩いた時刻) |
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
| `qai init` | コンテキストの設定 | `~/.qairc` の作成・更新 |
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
| `qai todo add "内容"` | 具体的な Todo を直接追加 | (新規) -> `todo` |
| `qai todo list` | 今日やるべき Todo の一覧 | `todo` / `doing` の抽出 |
| `qai todo work [ID]` | Todo 着手（作業中へ） | `todo` -> `doing`, `started_at` 更新 |
| `qai todo done [ID]` | Todo 完了 | `doing` -> `done`, `duration` 自動計算して `Logs` 登録 |

### 実績・ログ
| コマンド | 説明 | 備考 |
| :--- | :--- | :--- |
| `qai report` | 振り返りレポート | 全実績の集計 |

## 5. 運用ルール
* **Ideas と Todo の分離**:
    - **アイデア (Ideas)**: 漠然とした願いやバックログ。`qai idea add` でここに入る。
    - **Todo**: 今日取り組む具体的な項目。`qai idea refine` や `qai todo add` を経てここに入る。
* **ワークフロー**:
    1. `qai idea add "家を綺麗にする"` (将来やりたいこと)
    2. `qai todo add "掃除機をかける"` (今日やること)
    3. `qai todo work [ID]` -> 集中開始。
    4. `qai todo done` -> 完了。

## 6. Markdown テンプレート案
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

