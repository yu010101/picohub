"""Notion Lite スキル

軽量Notion連携スキル。ページの作成・読取・更新に対応。
データベースへのレコード追加、日報テンプレートの自動生成が可能。
"""

import logging
import os
from datetime import date, datetime, timezone
from typing import Any, Optional

import requests

logger = logging.getLogger(__name__)

NOTION_API_BASE_URL = "https://api.notion.com/v1"
NOTION_API_VERSION = "2022-06-28"


class NotionLiteSkill:
    """軽量なNotion連携を提供するスキル。

    Notion APIを使用して、ページの作成・読取、データベースへのレコード追加、
    日報テンプレートの自動生成を行います。

    Attributes:
        api_key: Notion Internal Integration Token。
    """

    def __init__(self, api_key: Optional[str] = None) -> None:
        """NotionLiteSkillを初期化する。

        Args:
            api_key: Notion Internal Integration Token。
                省略時は環境変数 NOTION_API_KEY を使用。

        Raises:
            ValueError: APIキーが設定されていない場合。
        """
        self.api_key = api_key or os.environ.get("NOTION_API_KEY")

        if not self.api_key:
            raise ValueError(
                "Notion APIキーが必要です。引数または環境変数 "
                "NOTION_API_KEY で指定してください。"
            )

        self._session = requests.Session()
        self._session.headers.update(
            {
                "Authorization": f"Bearer {self.api_key}",
                "Content-Type": "application/json",
                "Notion-Version": NOTION_API_VERSION,
            }
        )

    def _parse_rich_text(self, rich_text_array: list[dict]) -> str:
        """Notion のリッチテキスト配列からプレーンテキストを抽出する。

        Args:
            rich_text_array: Notion APIのリッチテキストオブジェクトの配列。

        Returns:
            結合されたプレーンテキスト文字列。
        """
        return "".join(
            item.get("plain_text", "") for item in rich_text_array
        )

    def _build_rich_text(self, text: str) -> list[dict[str, Any]]:
        """プレーンテキストからNotion のリッチテキストオブジェクトを構築する。

        Args:
            text: プレーンテキスト文字列。

        Returns:
            Notion API形式のリッチテキストオブジェクトの配列。
        """
        return [{"type": "text", "text": {"content": text}}]

    def _content_to_blocks(self, content: str) -> list[dict[str, Any]]:
        """テキストコンテンツをNotion ブロックの配列に変換する。

        簡易的なMarkdown記法に対応:
        - "## " で始まる行 -> heading_2 ブロック
        - "### " で始まる行 -> heading_3 ブロック
        - "- " で始まる行 -> bulleted_list_item ブロック
        - 数字 + ". " で始まる行 -> numbered_list_item ブロック
        - その他 -> paragraph ブロック

        Args:
            content: テキストコンテンツ（簡易Markdown形式）。

        Returns:
            Notion API形式のブロックオブジェクトの配列。
        """
        blocks = []

        for line in content.split("\n"):
            stripped = line.strip()
            if not stripped:
                # 空行もパラグラフとして追加
                blocks.append(
                    {
                        "object": "block",
                        "type": "paragraph",
                        "paragraph": {"rich_text": []},
                    }
                )
            elif stripped.startswith("### "):
                blocks.append(
                    {
                        "object": "block",
                        "type": "heading_3",
                        "heading_3": {
                            "rich_text": self._build_rich_text(stripped[4:])
                        },
                    }
                )
            elif stripped.startswith("## "):
                blocks.append(
                    {
                        "object": "block",
                        "type": "heading_2",
                        "heading_2": {
                            "rich_text": self._build_rich_text(stripped[3:])
                        },
                    }
                )
            elif stripped.startswith("- "):
                blocks.append(
                    {
                        "object": "block",
                        "type": "bulleted_list_item",
                        "bulleted_list_item": {
                            "rich_text": self._build_rich_text(stripped[2:])
                        },
                    }
                )
            elif len(stripped) > 2 and stripped[0].isdigit() and ". " in stripped[:5]:
                # "1. " のような番号付きリスト
                text = stripped.split(". ", 1)[1] if ". " in stripped else stripped
                blocks.append(
                    {
                        "object": "block",
                        "type": "numbered_list_item",
                        "numbered_list_item": {
                            "rich_text": self._build_rich_text(text)
                        },
                    }
                )
            else:
                blocks.append(
                    {
                        "object": "block",
                        "type": "paragraph",
                        "paragraph": {
                            "rich_text": self._build_rich_text(stripped)
                        },
                    }
                )

        return blocks

    def create_page(
        self,
        parent_id: str,
        title: str,
        content: Optional[str] = None,
    ) -> dict[str, Any]:
        """新しいページを作成する。

        指定した親ページの下に、新しい子ページを作成します。
        コンテンツは簡易Markdown記法に対応しています。

        Args:
            parent_id: 親ページのID（ハイフンなし32文字 or ハイフン付き36文字）。
            title: ページタイトル。
            content: ページ本文（任意）。簡易Markdown記法に対応。

        Returns:
            作成結果を含む辞書。以下のキーを含む:
            - page_id (str): 作成されたページのID
            - url (str): ページのURL
            - title (str): ページタイトル

        Raises:
            ValueError: parent_id またはタイトルが空の場合。

        Example:
            >>> skill = NotionLiteSkill()
            >>> result = skill.create_page(
            ...     parent_id="abc123...",
            ...     title="新しいページ",
            ...     content="## 見出し\\n内容テキスト"
            ... )
            >>> print(result["url"])
        """
        if not parent_id:
            raise ValueError("親ページIDは空にできません。")
        if not title:
            raise ValueError("ページタイトルは空にできません。")

        url = f"{NOTION_API_BASE_URL}/pages"

        payload: dict[str, Any] = {
            "parent": {"page_id": parent_id},
            "properties": {
                "title": {
                    "title": self._build_rich_text(title),
                }
            },
        }

        if content:
            payload["children"] = self._content_to_blocks(content)

        try:
            response = self._session.post(url, json=payload, timeout=30)
            response.raise_for_status()
            data = response.json()

            page_id = data.get("id", "")
            page_url = data.get("url", "")

            logger.info("ページを作成しました: id=%s, title=%s", page_id, title)
            return {
                "page_id": page_id,
                "url": page_url,
                "title": title,
            }
        except requests.HTTPError as e:
            error_body = {}
            try:
                error_body = e.response.json()
            except (ValueError, AttributeError):
                pass
            logger.error("ページの作成に失敗しました: %s, body=%s", e, error_body)
            return {
                "page_id": "",
                "url": "",
                "title": title,
                "error": str(e),
                "details": error_body,
            }
        except requests.RequestException as e:
            logger.error("ネットワークエラー: %s", e)
            return {
                "page_id": "",
                "url": "",
                "title": title,
                "error": str(e),
            }

    def read_page(self, page_id: str) -> dict[str, Any]:
        """ページの内容を読み取る。

        指定したページのメタデータとブロックコンテンツを取得します。
        ブロックのテキスト内容をプレーンテキストとして結合して返します。

        Args:
            page_id: 読み取るページのID。

        Returns:
            ページ情報を含む辞書。以下のキーを含む:
            - title (str): ページタイトル
            - content (str): ページ本文（プレーンテキスト）
            - last_edited (str): 最終更新日時（ISO 8601形式）
            - created_time (str): 作成日時（ISO 8601形式）

        Raises:
            ValueError: page_id が空の場合。

        Example:
            >>> page = skill.read_page("page_id_here")
            >>> print(page["title"])
            >>> print(page["content"])
        """
        if not page_id:
            raise ValueError("ページIDは空にできません。")

        try:
            # ページのメタデータを取得
            page_url = f"{NOTION_API_BASE_URL}/pages/{page_id}"
            page_response = self._session.get(page_url, timeout=30)
            page_response.raise_for_status()
            page_data = page_response.json()

            # タイトルの抽出
            properties = page_data.get("properties", {})
            title = ""
            for prop in properties.values():
                if prop.get("type") == "title":
                    title = self._parse_rich_text(prop.get("title", []))
                    break

            # ブロックコンテンツの取得
            blocks_url = f"{NOTION_API_BASE_URL}/blocks/{page_id}/children"
            blocks_response = self._session.get(blocks_url, timeout=30)
            blocks_response.raise_for_status()
            blocks_data = blocks_response.json()

            # ブロックからテキストを抽出
            content_parts = []
            for block in blocks_data.get("results", []):
                block_type = block.get("type", "")
                block_content = block.get(block_type, {})
                rich_text = block_content.get("rich_text", [])
                text = self._parse_rich_text(rich_text)

                if block_type in ("heading_1", "heading_2", "heading_3"):
                    prefix = "#" * int(block_type[-1])
                    content_parts.append(f"{prefix} {text}")
                elif block_type == "bulleted_list_item":
                    content_parts.append(f"- {text}")
                elif block_type == "numbered_list_item":
                    content_parts.append(f"* {text}")
                elif block_type == "to_do":
                    checked = block_content.get("checked", False)
                    marker = "[x]" if checked else "[ ]"
                    content_parts.append(f"- {marker} {text}")
                elif text:
                    content_parts.append(text)

            content = "\n".join(content_parts)

            logger.info("ページを読み取りました: id=%s", page_id)
            return {
                "title": title,
                "content": content,
                "last_edited": page_data.get("last_edited_time", ""),
                "created_time": page_data.get("created_time", ""),
            }
        except requests.HTTPError as e:
            error_body = {}
            try:
                error_body = e.response.json()
            except (ValueError, AttributeError):
                pass
            logger.error("ページの読み取りに失敗しました: %s, body=%s", e, error_body)
            return {
                "title": "",
                "content": "",
                "last_edited": "",
                "created_time": "",
                "error": str(e),
                "details": error_body,
            }
        except requests.RequestException as e:
            logger.error("ネットワークエラー: %s", e)
            return {
                "title": "",
                "content": "",
                "last_edited": "",
                "created_time": "",
                "error": str(e),
            }

    def add_database_record(
        self,
        database_id: str,
        properties: dict[str, Any],
    ) -> dict[str, Any]:
        """データベースにレコードを追加する。

        指定したNotionデータベースに新しいレコード（ページ）を追加します。
        プロパティ名と値の辞書を受け取り、Notion API形式に自動変換します。

        サポートされるプロパティ型:
        - 文字列: title または rich_text として自動判定
        - 日付文字列（YYYY-MM-DD形式）: date 型として設定
        - その他: rich_text として設定

        Args:
            database_id: 対象データベースのID。
            properties: プロパティ名と値の辞書。
                例: {"名前": "田中太郎", "期限": "2025-12-31"}

        Returns:
            追加結果を含む辞書。以下のキーを含む:
            - record_id (str): 追加されたレコードのID
            - url (str): レコードのURL

        Raises:
            ValueError: database_id またはプロパティが空の場合。

        Example:
            >>> result = skill.add_database_record(
            ...     database_id="db_id",
            ...     properties={"名前": "タスクA", "ステータス": "進行中"}
            ... )
            >>> print(result["record_id"])
        """
        if not database_id:
            raise ValueError("データベースIDは空にできません。")
        if not properties:
            raise ValueError("プロパティは空にできません。")

        url = f"{NOTION_API_BASE_URL}/pages"

        # プロパティをNotion API形式に変換
        notion_properties = {}
        first_property = True

        for prop_name, prop_value in properties.items():
            if first_property:
                # 最初のプロパティをタイトルとして扱う
                notion_properties[prop_name] = {
                    "title": self._build_rich_text(str(prop_value)),
                }
                first_property = False
            elif self._is_date_string(str(prop_value)):
                notion_properties[prop_name] = {
                    "date": {"start": str(prop_value)},
                }
            elif isinstance(prop_value, bool):
                notion_properties[prop_name] = {
                    "checkbox": prop_value,
                }
            elif isinstance(prop_value, (int, float)):
                notion_properties[prop_name] = {
                    "number": prop_value,
                }
            else:
                notion_properties[prop_name] = {
                    "rich_text": self._build_rich_text(str(prop_value)),
                }

        payload = {
            "parent": {"database_id": database_id},
            "properties": notion_properties,
        }

        try:
            response = self._session.post(url, json=payload, timeout=30)
            response.raise_for_status()
            data = response.json()

            record_id = data.get("id", "")
            record_url = data.get("url", "")

            logger.info("レコードを追加しました: id=%s", record_id)
            return {
                "record_id": record_id,
                "url": record_url,
            }
        except requests.HTTPError as e:
            error_body = {}
            try:
                error_body = e.response.json()
            except (ValueError, AttributeError):
                pass
            logger.error("レコードの追加に失敗しました: %s, body=%s", e, error_body)
            return {
                "record_id": "",
                "url": "",
                "error": str(e),
                "details": error_body,
            }
        except requests.RequestException as e:
            logger.error("ネットワークエラー: %s", e)
            return {
                "record_id": "",
                "url": "",
                "error": str(e),
            }

    def _is_date_string(self, value: str) -> bool:
        """文字列が日付形式（YYYY-MM-DD）かどうかを判定する。

        Args:
            value: 判定する文字列。

        Returns:
            日付形式の場合True。
        """
        try:
            datetime.strptime(value, "%Y-%m-%d")
            return True
        except ValueError:
            return False

    def generate_daily_report(self, database_id: str) -> dict[str, Any]:
        """日報テンプレートを自動生成してデータベースに追加する。

        今日の日付で日報テンプレートを作成し、指定のデータベースに追加します。
        テンプレートには以下のセクションが含まれます:
        - 今日のタスク
        - 完了したタスク
        - 明日の予定
        - メモ・気づき

        Args:
            database_id: 日報データベースのID。

        Returns:
            生成結果を含む辞書。以下のキーを含む:
            - record_id (str): 追加されたレコードのID
            - url (str): レコードのURL
            - date (str): 日報の日付（YYYY-MM-DD形式）
            - template (str): 生成されたテンプレートのテキスト

        Raises:
            ValueError: database_id が空の場合。

        Example:
            >>> result = skill.generate_daily_report("db_id")
            >>> print(f"日報を作成しました: {result['date']}")
        """
        if not database_id:
            raise ValueError("データベースIDは空にできません。")

        today = date.today()
        date_str = today.strftime("%Y-%m-%d")
        weekday_names = ["月", "火", "水", "木", "金", "土", "日"]
        weekday = weekday_names[today.weekday()]

        title = f"日報 {date_str}（{weekday}）"

        template_lines = [
            f"## 日報 {date_str}（{weekday}）",
            "",
            "### 今日のタスク",
            "- [ ] ",
            "- [ ] ",
            "- [ ] ",
            "",
            "### 完了したタスク",
            "- ",
            "",
            "### 明日の予定",
            "- ",
            "",
            "### メモ・気づき",
            "",
        ]
        template = "\n".join(template_lines)

        # まずデータベースにレコードを追加
        url = f"{NOTION_API_BASE_URL}/pages"

        payload: dict[str, Any] = {
            "parent": {"database_id": database_id},
            "properties": {
                "名前": {"title": self._build_rich_text(title)},
                "日付": {"date": {"start": date_str}},
            },
            "children": self._content_to_blocks(template),
        }

        try:
            response = self._session.post(url, json=payload, timeout=30)
            response.raise_for_status()
            data = response.json()

            record_id = data.get("id", "")
            record_url = data.get("url", "")

            logger.info("日報を作成しました: id=%s, date=%s", record_id, date_str)
            return {
                "record_id": record_id,
                "url": record_url,
                "date": date_str,
                "template": template,
            }
        except requests.HTTPError as e:
            error_body = {}
            try:
                error_body = e.response.json()
            except (ValueError, AttributeError):
                pass
            logger.error("日報の作成に失敗しました: %s, body=%s", e, error_body)
            return {
                "record_id": "",
                "url": "",
                "date": date_str,
                "template": template,
                "error": str(e),
                "details": error_body,
            }
        except requests.RequestException as e:
            logger.error("ネットワークエラー: %s", e)
            return {
                "record_id": "",
                "url": "",
                "date": date_str,
                "template": template,
                "error": str(e),
            }
