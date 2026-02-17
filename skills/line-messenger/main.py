"""LINE Messenger スキル

PicoClawからLINEメッセージを送受信するスキル。
LINE Messaging APIを使用してテキスト、画像、スタンプの送信に対応。
"""

import json
import logging
import os
from typing import Any, Optional

import requests

logger = logging.getLogger(__name__)

LINE_API_BASE_URL = "https://api.line.me/v2/bot"


class LineMessengerSkill:
    """LINE Messaging APIを利用してメッセージの送受信を行うスキル。

    Attributes:
        channel_access_token: LINEチャネルアクセストークン。
        channel_secret: LINEチャネルシークレット。
    """

    def __init__(
        self,
        channel_access_token: Optional[str] = None,
        channel_secret: Optional[str] = None,
    ) -> None:
        """LineMessengerSkillを初期化する。

        Args:
            channel_access_token: LINEチャネルアクセストークン。
                省略時は環境変数 LINE_CHANNEL_ACCESS_TOKEN を使用。
            channel_secret: LINEチャネルシークレット。
                省略時は環境変数 LINE_CHANNEL_SECRET を使用。

        Raises:
            ValueError: チャネルアクセストークンが設定されていない場合。
        """
        self.channel_access_token = channel_access_token or os.environ.get(
            "LINE_CHANNEL_ACCESS_TOKEN"
        )
        self.channel_secret = channel_secret or os.environ.get("LINE_CHANNEL_SECRET")

        if not self.channel_access_token:
            raise ValueError(
                "チャネルアクセストークンが必要です。引数または環境変数 "
                "LINE_CHANNEL_ACCESS_TOKEN で指定してください。"
            )

        self._session = requests.Session()
        self._session.headers.update(
            {
                "Authorization": f"Bearer {self.channel_access_token}",
                "Content-Type": "application/json",
            }
        )

    def _send_push_message(self, to: str, messages: list[dict[str, Any]]) -> dict:
        """LINE Push Message APIを呼び出してメッセージを送信する。

        Args:
            to: 送信先のユーザーID。
            messages: 送信するメッセージオブジェクトのリスト。

        Returns:
            送信結果を含む辞書。

        Raises:
            requests.HTTPError: API呼び出しが失敗した場合。
        """
        url = f"{LINE_API_BASE_URL}/message/push"
        payload = {
            "to": to,
            "messages": messages,
        }

        try:
            response = self._session.post(url, json=payload, timeout=30)
            response.raise_for_status()
            logger.info("メッセージを送信しました: to=%s", to)
            return {"status": "ok"}
        except requests.HTTPError as e:
            error_body = {}
            try:
                error_body = e.response.json()
            except (ValueError, AttributeError):
                pass
            logger.error(
                "メッセージの送信に失敗しました: to=%s, status=%s, body=%s",
                to,
                e.response.status_code if e.response is not None else "N/A",
                error_body,
            )
            return {
                "status": "error",
                "error": str(e),
                "details": error_body,
            }
        except requests.RequestException as e:
            logger.error("ネットワークエラー: %s", e)
            return {"status": "error", "error": str(e)}

    def send_text(self, to: str, message: str) -> dict:
        """テキストメッセージを送信する。

        Args:
            to: 送信先のユーザーID。LINE Platform上の一意な識別子。
            message: 送信するテキストメッセージ。最大5000文字。

        Returns:
            送信結果を含む辞書。
            成功時: {"status": "ok"}
            失敗時: {"status": "error", "error": "...", "details": {...}}

        Raises:
            ValueError: メッセージが空または5000文字を超える場合。

        Example:
            >>> skill = LineMessengerSkill()
            >>> result = skill.send_text("U1234...", "こんにちは！")
            >>> print(result)
            {"status": "ok"}
        """
        if not message:
            raise ValueError("メッセージは空にできません。")
        if len(message) > 5000:
            raise ValueError(
                f"メッセージが長すぎます（{len(message)}文字）。最大5000文字です。"
            )

        messages = [{"type": "text", "text": message}]
        return self._send_push_message(to, messages)

    def send_image(self, to: str, image_url: str) -> dict:
        """画像メッセージを送信する。

        Args:
            to: 送信先のユーザーID。
            image_url: 送信する画像のURL。HTTPS必須。
                JPEG または PNG 形式を推奨。最大ファイルサイズ: 10MB。

        Returns:
            送信結果を含む辞書。
            成功時: {"status": "ok"}
            失敗時: {"status": "error", "error": "...", "details": {...}}

        Raises:
            ValueError: URLがHTTPSでない場合。

        Example:
            >>> skill = LineMessengerSkill()
            >>> result = skill.send_image("U1234...", "https://example.com/photo.jpg")
            >>> print(result)
            {"status": "ok"}
        """
        if not image_url.startswith("https://"):
            raise ValueError("画像URLはHTTPSである必要があります。")

        messages = [
            {
                "type": "image",
                "originalContentUrl": image_url,
                "previewImageUrl": image_url,
            }
        ]
        return self._send_push_message(to, messages)

    def receive_webhook(self, data: dict) -> list[dict[str, Any]]:
        """Webhookデータを受信して解析する。

        LINE Platformから送られてくるWebhookイベントデータを解析し、
        テキストメッセージイベントの情報を抽出して返す。

        Args:
            data: LINE PlatformからのWebhookリクエストボディ。
                "events"キーにイベントのリストが含まれている必要がある。

        Returns:
            解析されたイベント情報の辞書のリスト。各辞書には以下のキーが含まれる:
            - event_type (str): イベントタイプ（"message"など）
            - message_type (str): メッセージタイプ（"text"など）
            - text (str): メッセージのテキスト（テキストメッセージの場合）
            - user_id (str): 送信者のユーザーID
            - reply_token (str): リプライトークン

        Example:
            >>> webhook_data = {"events": [{"type": "message", ...}]}
            >>> events = skill.receive_webhook(webhook_data)
            >>> for event in events:
            ...     print(event["text"])
        """
        parsed_events = []
        events = data.get("events", [])

        if not events:
            logger.warning("Webhookデータにイベントが含まれていません。")
            return parsed_events

        for event in events:
            try:
                event_type = event.get("type", "unknown")
                source = event.get("source", {})
                user_id = source.get("userId", "")
                reply_token = event.get("replyToken", "")

                parsed_event = {
                    "event_type": event_type,
                    "user_id": user_id,
                    "reply_token": reply_token,
                }

                if event_type == "message":
                    message = event.get("message", {})
                    message_type = message.get("type", "unknown")
                    parsed_event["message_type"] = message_type

                    if message_type == "text":
                        parsed_event["text"] = message.get("text", "")
                    elif message_type == "image":
                        parsed_event["content_id"] = message.get("id", "")
                    elif message_type == "sticker":
                        parsed_event["sticker_id"] = message.get("stickerId", "")
                        parsed_event["package_id"] = message.get("packageId", "")

                parsed_events.append(parsed_event)
                logger.info(
                    "イベントを解析しました: type=%s, user_id=%s",
                    event_type,
                    user_id,
                )
            except Exception as e:
                logger.error("イベントの解析に失敗しました: %s", e)
                continue

        return parsed_events
