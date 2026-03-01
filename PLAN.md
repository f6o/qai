# 次世代タスク管理ツール qai 仕様設計書

## 1. コンセプト
* **Hybrid Storage**: 実績ログは Markdown (`YYYY-MM-DD.md`)、状態管理は SQLite (`qai.db`) で行う。
* **LLM-Driven Refinement**: 曖昧な「やりたいこと」を LLM が実行可能な「タスク」へ具体化する。
* **Zero Friction Continuity**: 昨日の残りタスクを今日の日付ファイルへ自動的に引き継ぎ、作業の断絶をなくす。

## 2. システム構成
* **言語**: Go
* **DB**: SQLite (Single Source of Truth)
* **UI**: CLI + TUI (Bubble Tea)
* **外部連携**: Gemini API / OpenAI API

## 3. データモデル (SQLite)

### Tasks (タスク管理)
| カラム | 型 | 説明 |
| :--- | :--- | :--- |
| `id` | UUID/INT | 一意識別子 |
| `title` | TEXT | タスク名 |
| `status` | STRING | todo / doing / done / archived |
| `priority` | STRING | A (High) / B (Normal) / C (Low) |
| `category` | STRING | work / hobby (カレントディレクトリで判定) |
| `created_at`| DATETIME | 作成日時 |

### Logs (実績記録)
| カラム | 型 | 説明 |
| :--- | :--- | :--- |
| `id` | INT | 一意識別子 |
| `task_id` | INT | 関連するタスクの ID (任意) |
| `content` | TEXT | やったことの内容 |
| `duration` | INT | 作業時間 (分) |
| `logged_at` | DATETIME | 記録日時 |

## 4. コマンド体系

| コマンド | 機能概要 |
| :--- | :--- |
| `qai want "内容"` | 今日の `[やりたいこと]` セクションにアイデアを追加する。 |
| `qai refine` | `[やりたいこと]` を LLM に投げ、DB にタスクとして登録、Markdown に反映する。 |
| `qai sync` | DB の未完了タスクを Markdown に書き出し、Markdown 上の完了を DB に同期する。 |
| `qai pomo "内容"` | 指定された内容でポモドーロを開始。終了後、DB と Markdown に実績を記録する。 |
| `qai done "内容"` | 即時で `[やったこと]` セクションに実績を記録する。 |
| `qai list` | DB 内の未完了タスクを優先度順にターミナルへ一覧表示する。 |
| `qai report` | 指定期間内の全タスク・実績を解析し、サマリーを標準出力/ファイルに出力する。 |

## 5. 運用ユースケース (3日間)

1. **1日目 (月)**: `qai want` で案を出し、`qai refine` でタスク分解。`qai pomo` で作業開始。
2. **2日目 (火)**: 朝、`qai sync` を実行。昨日の未完了タスクが自動で `2026-03-03.md` に出現。
3. **3日目 (水)**: 全タスクを完了させ、`qai report` で 3 日間の活動ログを出力し、報告に活用。

## 6. 技術的な利点
* **Markdown の柔軟性**: ツールを通さずエディタで直接メモを追記しても、運用が壊れない。
* **DB の堅牢性**: 日付をまたぐ集計や、過去の特定タスクの検索が高速かつ正確。
* **Go のポータビリティ**: シングルバイナリで動作するため、仕事（Work）と趣味（Hobby）の環境構築が容易。
