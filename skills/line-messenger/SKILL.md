# LINE Messenger スキル

PicoClawからLINEメッセージを送受信するためのスキルです。LINE Messaging APIを使用して、テキストメッセージ、画像、スタンプの送信に対応しています。

## 概要

このスキルは、LINE公式アカウントを通じてユーザーにメッセージを送信したり、Webhookを通じてユーザーからのメッセージを受信したりすることができます。

### 主な機能

- テキストメッセージの送信
- 画像メッセージの送信
- Webhookによるメッセージの受信と解析

## セットアップ

### 1. LINE Developers コンソールでの準備

1. [LINE Developers](https://developers.line.biz/) にアクセスし、ログインします。
2. プロバイダーを作成します（未作成の場合）。
3. 「Messaging API」チャネルを新規作成します。
4. チャネル基本設定から以下の情報を取得します：
   - **チャネルシークレット**: Webhook署名検証に使用
   - **チャネルアクセストークン**: API呼び出しに使用

### 2. チャネルアクセストークンの発行

1. LINE Developers コンソールで対象チャネルを開きます。
2. 「Messaging API設定」タブを選択します。
3. 「チャネルアクセストークン（長期）」の「発行」ボタンをクリックします。
4. 発行されたトークンを安全な場所に保存します。

### 3. 環境変数の設定

以下の環境変数を設定してください：

```
LINE_CHANNEL_ACCESS_TOKEN=your_channel_access_token
LINE_CHANNEL_SECRET=your_channel_secret
```

### 4. Webhook URLの設定

LINE Developers コンソールの「Messaging API設定」で、Webhook URLを設定します：

```
https://your-domain.com/webhook/line-messenger
```

## 使用例

### テキストメッセージの送信

```python
from main import LineMessengerSkill

skill = LineMessengerSkill()

# ユーザーにテキストメッセージを送信
result = skill.send_text(
    to="U1234567890abcdef1234567890abcdef",
    message="こんにちは！今日の予定をお知らせします。"
)
print(result)
```

### 画像メッセージの送信

```python
# ユーザーに画像を送信
result = skill.send_image(
    to="U1234567890abcdef1234567890abcdef",
    image_url="https://example.com/images/photo.jpg"
)
print(result)
```

### Webhookの受信

```python
# Webhookデータの処理
webhook_data = {
    "events": [
        {
            "type": "message",
            "message": {
                "type": "text",
                "text": "こんにちは"
            },
            "source": {
                "userId": "U1234567890abcdef1234567890abcdef",
                "type": "user"
            },
            "replyToken": "reply_token_here"
        }
    ]
}

events = skill.receive_webhook(webhook_data)
for event in events:
    print(f"ユーザー: {event['user_id']}")
    print(f"メッセージ: {event['text']}")
```

## API リファレンス

### `LineMessengerSkill(channel_access_token=None, channel_secret=None)`

スキルのインスタンスを作成します。

| パラメータ | 型 | 必須 | 説明 |
|---|---|---|---|
| `channel_access_token` | str | いいえ | チャネルアクセストークン。省略時は環境変数 `LINE_CHANNEL_ACCESS_TOKEN` を使用 |
| `channel_secret` | str | いいえ | チャネルシークレット。省略時は環境変数 `LINE_CHANNEL_SECRET` を使用 |

### `send_text(to, message)`

テキストメッセージを送信します。

| パラメータ | 型 | 必須 | 説明 |
|---|---|---|---|
| `to` | str | はい | 送信先のユーザーID |
| `message` | str | はい | 送信するテキストメッセージ（最大5000文字） |

**戻り値**: `dict` - 送信結果（成功時: `{"status": "ok"}`）

### `send_image(to, image_url)`

画像メッセージを送信します。

| パラメータ | 型 | 必須 | 説明 |
|---|---|---|---|
| `to` | str | はい | 送信先のユーザーID |
| `image_url` | str | はい | 送信する画像のURL（HTTPS必須） |

**戻り値**: `dict` - 送信結果（成功時: `{"status": "ok"}`）

### `receive_webhook(data)`

Webhookデータを受信して解析します。

| パラメータ | 型 | 必須 | 説明 |
|---|---|---|---|
| `data` | dict | はい | LINE PlatformからのWebhookイベントデータ |

**戻り値**: `list[dict]` - 解析されたイベントのリスト

## 注意事項

- チャネルアクセストークンは定期的にローテーションすることを推奨します。
- 画像URLはHTTPSである必要があります。
- メッセージ送信にはユーザーの同意（友だち追加）が必要です。
- API呼び出しにはレート制限があります。詳細はLINE公式ドキュメントを参照してください。

## ライセンス

MIT License
