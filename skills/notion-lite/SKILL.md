# Notion Lite スキル

軽量なNotion連携スキルです。Notion APIを使用して、ページの作成・読取・更新、データベースへのレコード追加、日報テンプレートの自動生成を行います。

## 概要

このスキルは、PicoClawからNotionのワークスペースに簡単にアクセスするための軽量インターフェースを提供します。複雑なNotion APIの操作を簡潔なメソッドで実行できます。

### 主な機能

- ページの作成（テキスト・見出し・箇条書きを含むリッチコンテンツ対応）
- ページの読み取り
- データベースへのレコード追加
- 日報テンプレートの自動生成

## セットアップ

### 1. Notion インテグレーションの作成

1. [Notion Developers](https://www.notion.so/my-integrations) にアクセスします。
2. 「新しいインテグレーション」をクリックします。
3. インテグレーション名を入力（例: "PicoClaw連携"）。
4. 関連するワークスペースを選択します。
5. 機能で「コンテンツを読み取る」「コンテンツを挿入する」「コンテンツを更新する」を有効にします。
6. 「送信」をクリックしてインテグレーションを作成します。
7. 表示された **Internal Integration Token** を控えます。

### 2. ページ/データベースへの接続許可

操作対象のNotionページまたはデータベースで、作成したインテグレーションを接続します：

1. 対象のページ/データベースを開きます。
2. 右上の「...」メニューから「接続」を選択します。
3. 作成したインテグレーションを検索して追加します。

### 3. 環境変数の設定

以下の環境変数を設定してください：

```
NOTION_API_KEY=your_internal_integration_token
```

## 使用例

### ページの作成

```python
from main import NotionLiteSkill

skill = NotionLiteSkill()

# 新しいページを作成
result = skill.create_page(
    parent_id="parent_page_id_here",
    title="週次ミーティング議事録",
    content="## 参加者\n- 田中\n- 鈴木\n\n## 議題\n1. 進捗報告\n2. 来週の計画"
)
print(f"ページID: {result['page_id']}")
print(f"URL: {result['url']}")
```

### ページの読み取り

```python
# ページの内容を取得
page = skill.read_page(page_id="target_page_id_here")
print(f"タイトル: {page['title']}")
print(f"内容: {page['content']}")
print(f"最終更新: {page['last_edited']}")
```

### データベースへのレコード追加

```python
# データベースにレコードを追加
result = skill.add_database_record(
    database_id="database_id_here",
    properties={
        "名前": "田中太郎",
        "ステータス": "進行中",
        "期限": "2025-12-31",
        "優先度": "高",
    }
)
print(f"レコードID: {result['record_id']}")
```

### 日報テンプレートの自動生成

```python
# 今日の日報を自動生成してデータベースに追加
result = skill.generate_daily_report(
    database_id="daily_report_db_id_here"
)
print(f"日報ID: {result['record_id']}")
print(f"日付: {result['date']}")
```

## API リファレンス

### `NotionLiteSkill(api_key=None)`

スキルのインスタンスを作成します。

| パラメータ | 型 | 必須 | 説明 |
|---|---|---|---|
| `api_key` | str | いいえ | Notion Internal Integration Token。省略時は環境変数 `NOTION_API_KEY` を使用 |

### `create_page(parent_id, title, content=None)`

新しいページを作成します。

| パラメータ | 型 | 必須 | 説明 |
|---|---|---|---|
| `parent_id` | str | はい | 親ページまたはデータベースのID |
| `title` | str | はい | ページタイトル |
| `content` | str | いいえ | ページ本文（Markdown形式の簡易記法に対応） |

**戻り値**: `dict` - 作成結果（page_id, url を含む）

### `read_page(page_id)`

ページの内容を読み取ります。

| パラメータ | 型 | 必須 | 説明 |
|---|---|---|---|
| `page_id` | str | はい | 読み取るページのID |

**戻り値**: `dict` - ページ情報（title, content, last_edited を含む）

### `add_database_record(database_id, properties)`

データベースにレコードを追加します。

| パラメータ | 型 | 必須 | 説明 |
|---|---|---|---|
| `database_id` | str | はい | 対象データベースのID |
| `properties` | dict | はい | プロパティ名と値の辞書 |

**戻り値**: `dict` - 追加結果（record_id を含む）

### `generate_daily_report(database_id)`

日報テンプレートを自動生成してデータベースに追加します。

| パラメータ | 型 | 必須 | 説明 |
|---|---|---|---|
| `database_id` | str | はい | 日報データベースのID |

**戻り値**: `dict` - 生成結果（record_id, date, template を含む）

## 日報テンプレートの構成

自動生成される日報テンプレートには以下のセクションが含まれます：

1. **日付** - 当日の日付
2. **今日のタスク** - タスクチェックリスト（空欄）
3. **完了したタスク** - 完了タスクの記録欄
4. **明日の予定** - 翌日の予定記入欄
5. **メモ・気づき** - 自由記述欄

## 注意事項

- Notion APIにはレート制限があります（平均3リクエスト/秒）。
- ページの内容はブロック単位で取得されます。大量のブロックがあるページは取得に時間がかかる場合があります。
- インテグレーションに接続されていないページ/データベースにはアクセスできません。
- 日報テンプレートのカスタマイズは今後のバージョンで対応予定です。

## ライセンス

MIT License
