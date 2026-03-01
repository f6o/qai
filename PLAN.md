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

### Tasks (Ideas / TODO 管理)
ひとつのテーブルで「漠然としたアイデア」から「具体的なタスク」までを管理する。

| カラム | 型 | 説明 |
| :--- | :--- | :--- |
| `id` | INT | 一意識別子 (CLI 操作用) |
| `title` | TEXT | 内容 |
| `status` | STRING | want (Idea) / todo (Task) / doing / done / archived |
| `priority` | INT | 優先度 (正数: 大きいほど高優先。DEFAULT 10, CHECK > 0) |
| `parent_id`| INT | 親タスクの ID (アイデアを分解した場合の紐付け用) |
| `created_at`| DATETIME | 作成日時 |

### Logs (実績記録)

## 4. コマンド体系 (Structured Subcommands)

### 共通 / 設定
| コマンド | 説明 | DB / 設定 |
| :--- | :--- | :--- |
| `qai init` | コンテキストの設定 | `~/.qairc` の作成・更新 |
| `qai start` | 1日の開始（プランニング） | 今日の Markdown を生成 |

### アイデア管理 (Wants)
| コマンド | 説明 | DB 状態遷移 |
| :--- | :--- | :--- |
| `qai idea add "内容"` | 漠然としたアイデアの追加 | (新規) -> `want` |
| `qai idea list` | アイデア一覧の表示 | `want` の抽出 |
| `qai idea refine [ID]` | LLM によるタスク分解 | `want` -> `todo` (複数可) |

### タスク管理 (TODO/Pomodoro)
| コマンド | 説明 | DB 状態遷移 |
| :--- | :--- | :--- |
| `qai task add "内容"` | 具体的なタスクを直接追加 | (新規) -> `todo` |
| `qai task list` | 今日やるべきタスクの一覧 | `todo` / `doing` の抽出 |
| `qai task work [ID]` | タスク着手（ポモドーロ開始） | `todo` -> `doing` |
| `qai task done [ID]` | タスク完了 | `doing` -> `done` |

### 実績・ログ
| コマンド | 説明 | 備考 |
| :--- | :--- | :--- |
| `qai pomo [分] "内容"` | 汎用的な集中ログ | タスク外の記録用 |
| `qai report` | 振り返りレポート | 全実績の集計 |

## 5. 運用ルール
* **Wants と TODO の分離**:
    - **やりたいこと (Wants)**: 漠然とした願いやバックログ。`qai idea add` でここに入る。
    - **TODO**: 今日取り組む具体的なタスク。`qai idea refine` や `qai task add` を経てここに入る。
* **ワークフロー**:
    1. `qai idea add "家を綺麗にする"`
    2. `qai idea refine [ID]` -> `task` (todo) が生成される。
    3. `qai task work [ID]` -> 集中開始。
    4. `qai task done` -> 完了。

## 6. Markdown テンプレート案
```markdown
# 2026-03-01 (Sun)

## [やりたいこと (Wants)]
- [ ] 1: 新機能のアイデアを練る
- [ ] 2: 旅行の計画を立てる

## [TODO]
- [ ] 3: 掃除機をかける (10)
- [/] 4: ゴミを出す (10) [doing]

## [やったこと]
- [x] 5: 昨日の日報を書く (09:15)

## [メモ]
```

