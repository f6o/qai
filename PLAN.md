# 次世代タスク管理ツール qai 仕様設計書 (改訂版)

## 1. コンセプト
* **DB-Centric Hybrid Storage**: 状態管理の正解（SSOT）は SQLite (`qai.db`)。Markdown (`YYYY-MM-DD.md`) はその日の「タスク一覧のビュー」兼「活動ログ」として扱う。
* **LLM-Driven Refinement**: 曖昧なアイデアを LLM が分解し、DB へ直接タスク登録する。
* **Explicit Continuity**: 毎朝 `qai start` を実行することで、DB 内の未完了タスクを今日の日付ファイルへ一括展開する。

## 2. システム構成
* **言語**: Go
* **DB**: SQLite (Single Source of Truth)
* **UI**: CLI + TUI (Bubble Tea)
* **外部連携**: Gemini API / OpenAI API

## 3. データモデル (SQLite)

### Tasks (タスク管理)
| カラム | 型 | 説明 |
| :--- | :--- | :--- |
| `id` | INT | 一意識別子 (CLI 操作用) |
| `title` | TEXT | タスク名 |
| `status` | STRING | todo / doing / done / archived |
| `priority` | STRING | A (High) / B (Normal) / C (Low) |
| `category` | STRING | work / hobby (ディレクトリ判定) |
| `created_at`| DATETIME | 作成日時 |

### Logs (実績記録)
| カラム | 型 | 説明 |
| :--- | :--- | :--- |
| `id` | INT | 一意識別子 |
| `task_id` | INT | 関連するタスクの ID (任意) |
| `content` | TEXT | 内容（ポモドーロや done ログ） |
| `duration` | INT | 作業時間 (分) |
| `logged_at` | DATETIME | 記録日時 |

## 4. コマンド体系

| コマンド | DB 操作 | Markdown 反映 |
| :--- | :--- | :--- |
| `qai start` | 未完了タスクを抽出 | `YYYY-MM-DD.md` を生成。タスクを `[やりたいこと]` に列挙。 |
| `qai want "内容"` | タスクを新規登録 | 今日のファイルの `[やりたいこと]` に追記。 |
| `qai refine` | 既存タスクを分解・更新 | 分解後のタスクを `[やりたいこと]` に反映。 |
| `qai done [ID]` | ステータスを `done` に | `[やったこと]` に完了時刻と共に追記。 |
| `qai pomo "内容"` | `Logs` に実績を記録 | `[やったこと]` に時間・時刻と共に追記。 |
| `qai list` | 未完了タスクを一覧表示 | (なし) |
| `qai report` | 全実績を解析・出力 | (なし) |

## 5. 運用ルール
* **Markdown 手動編集の扱い**: ユーザーが Markdown を直接編集（メモ追記など）することは自由だが、その内容は **DB には同期されない**。タスクの完了や追加は必ず `qai` コマンドを通じて行う。
* **カテゴリ判定**: 実行時のカレントディレクトリが設定済みの `work_dir` 配下であれば自動的に `work` カテゴリとする。
* **ID の再利用**: `qai list` 等で表示される ID は、CLI での打ちやすさを考慮した短い数値とする。

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
