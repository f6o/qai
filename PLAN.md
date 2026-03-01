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

### Tasks (タスク管理)
| カラム | 型 | 説明 |
| :--- | :--- | :--- |
| `id` | INT | 一意識別子 (CLI 操作用) |
| `title` | TEXT | タスク名 |
| `status` | STRING | todo / doing / done / archived |
| `priority` | INT | 優先度 (正数: 大きいほど高優先。DEFAULT 10, CHECK > 0) |
| `created_at`| DATETIME | 作成日時 |


### Logs (実績記録)
| カラム | 型 | 説明 |
| :--- | :--- | :--- |
| `id` | INT | 一意識別子 |
| `task_id` | INT | 関連するタスクの ID (任意) |
| `content` | TEXT | 内容（ポモドーロや done ログ） |
| `duration` | INT | 作業時間 (分) |
| `logged_at` | DATETIME | 記録日時 |

## 4. コマンド体系 (Pomodoro-Centric)

| コマンド | 説明 | DB 操作 | Markdown 反映 |
| :--- | :--- | :--- | :--- |
| `qai init` | コンテキストの対話的設定 | `~/.qairc` に新しい DB/ログパスを追加 | (なし) |
| `qai start` | 1日の開始（プランニング） | 未完了タスクを抽出 | 今日の `.md` を作成しタスクを列挙 |
| `qai want "内容"` | アイデア・タスクの追加 | `Tasks` に `todo` で登録 | 今日のファイルの `[やりたいこと]` に追記 |
| `qai work [ID]` | タスクの着手（ポモドーロ開始） | ステータスを `doing` に変更 | (任意) ステータス更新 |
| `qai done [ID]` | タスク完了 | ステータスを `done` に | `[やったこと]` に完了時刻と共に移動 |
| `qai pomo [分]` | 汎用的な集中ログ | `Logs` に実績を記録 | `[やったこと]` にポモログを追記 |
| `qai refine` | LLM によるタスク分解 | 既存タスクを詳細化・更新 | 今日のファイルの `[やりたいこと]` を更新 |
| `qai list` | タスク一覧表示 | 未完了タスクをクエリ | (なし) |
| `qai report` | 振り返りレポート | 全実績を解析・集計出力 | (なし) |

## 5. 運用ルール
* **qai init の動作**: 
    - 実行時のディレクトリを `root` とし、DB名とログ保存先をユーザーに問い合せて `~/.qairc` を作成・更新する。
* **ポモドーロの流れ**:
    1. `qai start` で今日やることを確認。
    2. `qai work 1` でタスク1に集中開始。
    3. 完了したら `qai done` (ID省略時は `doing` のものを完了) でログを自動生成。
    4. 予定外の作業は `qai pomo "会議など"` でクイックに記録。

## 6. Markdown テンプレート案
```markdown
# 2026-03-01 (Sun)

## [やりたいこと]
- [ ] 1: 仕様を固める (A)
- [ ] 2: DB設計を行う (B)

## [やったこと]
- [x] 1: 仕様を固める (20:15)
- [pomo] プロトタイプ作成 (25min) (21:00)

## [メモ]
(ここはユーザーが自由に編集可能。qai はこのセクションを壊さない)
```
