# タスクファイルの分割計画

## 問題
tasks.yaml のタスク数が増えると編集しにくくなる。

## 解決策
ステータスごとにファイルを分割する。

## 分割後の構造

```
~/.config/qai/
├── tasks.yaml   # todo + doing（アクティブなタスク）
├── idea.yaml    # アイデア
└── done.yaml    # 完了済み
```

## 移動ルール

| 移動元 | 移動先 | 動作 |
|--------|--------|------|
| idea | todo | idea.yaml → 削除、tasks.yaml → 追加 |
| todo | doing | tasks.yaml 内でのステータス変更 |
| doing | done | tasks.yaml → 削除、done.yaml → 追加 |

## 利点
- 日常的に編集するタスクは tasks.yaml 一つで完結
- idea/done はアーカイブとして分離され、情報整理がしやすい

## 実装タイミング
実装必要时、`internal/storage/task.go` を修正して、状态ごとに别々のファイルに读写する逻辑を追加。
