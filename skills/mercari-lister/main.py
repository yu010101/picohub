"""Mercari Lister スキル

メルカリ出品テキストを自動生成するスキル。
商品写真から説明文を作成し、適切なカテゴリ・価格を提案。
"""

import logging
import re
from typing import Any, Optional

logger = logging.getLogger(__name__)

# 商品状態と価格倍率のマッピング
CONDITION_PRICE_MULTIPLIER = {
    "新品、未使用": {"min": 0.70, "max": 0.90, "label": "新品・未使用のため、大変綺麗な状態です。"},
    "未使用に近い": {"min": 0.60, "max": 0.80, "label": "ほぼ未使用で、非常に良好な状態です。"},
    "目立った傷や汚れなし": {"min": 0.40, "max": 0.70, "label": "目立った傷や汚れはなく、良好な状態です。"},
    "やや傷や汚れあり": {"min": 0.30, "max": 0.50, "label": "多少の使用感はありますが、問題なくご使用いただけます。"},
    "傷や汚れあり": {"min": 0.15, "max": 0.35, "label": "使用感がありますが、まだご使用いただけます。"},
    "全体的に状態が悪い": {"min": 0.05, "max": 0.20, "label": "全体的に使用感がございます。ご理解の上ご購入ください。"},
}

# 簡易カテゴリ推定用キーワードマッピング
CATEGORY_KEYWORDS = {
    "レディース": ["レディース", "ワンピース", "スカート", "ブラウス", "パンプス"],
    "メンズ": ["メンズ", "Tシャツ", "ジーンズ", "スニーカー", "ジャケット"],
    "ベビー・キッズ": ["ベビー", "キッズ", "子供", "幼児"],
    "インテリア・住まい": ["インテリア", "家具", "照明", "カーテン", "クッション"],
    "本・音楽・ゲーム": ["本", "漫画", "CD", "DVD", "ゲーム", "PlayStation", "Nintendo", "Switch"],
    "おもちゃ・ホビー": ["おもちゃ", "フィギュア", "プラモデル", "トレカ", "カード"],
    "コスメ・美容": ["コスメ", "化粧品", "美容", "香水", "スキンケア"],
    "家電・スマホ": [
        "家電", "スマホ", "iPhone", "iPad", "MacBook", "パソコン", "PC",
        "イヤホン", "AirPods", "カメラ", "テレビ",
    ],
    "スポーツ・レジャー": ["スポーツ", "ゴルフ", "テニス", "ランニング", "キャンプ", "アウトドア"],
    "ハンドメイド": ["ハンドメイド", "手作り", "手編み"],
    "チケット": ["チケット", "入場券", "観戦券"],
    "自動車・オートバイ": ["自動車", "バイク", "オートバイ", "カー用品"],
    "その他": [],
}

# ブランド別の基準価格帯（簡易的な参考値）
BRAND_BASE_PRICES = {
    "apple": 30000,
    "nike": 8000,
    "adidas": 7000,
    "uniqlo": 2000,
    "gu": 1500,
    "zara": 3000,
    "louis vuitton": 50000,
    "gucci": 40000,
    "chanel": 60000,
    "hermes": 80000,
    "sony": 15000,
    "nintendo": 20000,
    "dyson": 25000,
    "panasonic": 10000,
}

# デフォルトの基準価格
DEFAULT_BASE_PRICE = 5000


class MercariListerSkill:
    """メルカリ出品テキストを自動生成するスキル。

    商品名、状態、ブランド名から出品用の説明文を生成し、
    適切な価格を提案します。テンプレートベースのアプローチを使用。
    """

    def __init__(self) -> None:
        """MercariListerSkillを初期化する。"""
        logger.info("MercariListerSkillを初期化しました。")

    def _estimate_base_price(self, item_name: str, brand: Optional[str] = None) -> int:
        """商品名とブランドから基準価格を推定する。

        Args:
            item_name: 商品名。
            brand: ブランド名（任意）。

        Returns:
            推定基準価格（円）。
        """
        if brand:
            brand_lower = brand.lower()
            for brand_key, price in BRAND_BASE_PRICES.items():
                if brand_key in brand_lower or brand_lower in brand_key:
                    return price

        # 商品名からブランドを推定
        item_lower = item_name.lower()
        for brand_key, price in BRAND_BASE_PRICES.items():
            if brand_key in item_lower:
                return price

        return DEFAULT_BASE_PRICE

    def _estimate_category(self, item_name: str, brand: Optional[str] = None) -> str:
        """商品名とブランドからカテゴリを推定する。

        Args:
            item_name: 商品名。
            brand: ブランド名（任意）。

        Returns:
            推定カテゴリ名。
        """
        search_text = item_name
        if brand:
            search_text = f"{brand} {item_name}"

        for category, keywords in CATEGORY_KEYWORDS.items():
            for keyword in keywords:
                if keyword.lower() in search_text.lower():
                    return category

        return "その他"

    def _generate_hashtags(
        self, item_name: str, brand: Optional[str] = None
    ) -> list[str]:
        """商品名とブランドからハッシュタグを生成する。

        Args:
            item_name: 商品名。
            brand: ブランド名（任意）。

        Returns:
            ハッシュタグのリスト。
        """
        hashtags = []

        if brand:
            hashtags.append(f"#{brand}")

        # 商品名からキーワードを抽出（2文字以上の単語）
        words = re.findall(r"[A-Za-zぁ-んァ-ヶ一-龥0-9]{2,}", item_name)
        for word in words[:3]:
            tag = f"#{word}"
            if tag not in hashtags:
                hashtags.append(tag)

        return hashtags

    def generate_description(
        self,
        item_name: str,
        condition: str,
        brand: Optional[str] = None,
    ) -> dict[str, Any]:
        """商品説明文を生成する。

        テンプレートベースで、メルカリ出品に適した商品説明文を自動生成します。
        商品名、状態、ブランド名から魅力的な説明文を組み立てます。

        Args:
            item_name: 商品名。
            condition: 商品の状態（例: "新品、未使用", "目立った傷や汚れなし"）。
            brand: ブランド名（任意）。

        Returns:
            生成された説明文情報を含む辞書。以下のキーを含む:
            - text (str): 説明文全体
            - hashtags (list[str]): ハッシュタグのリスト
            - character_count (int): 文字数

        Raises:
            ValueError: 商品名または状態が空の場合。

        Example:
            >>> skill = MercariListerSkill()
            >>> result = skill.generate_description("AirPods Pro", "目立った傷や汚れなし", "Apple")
            >>> print(result["text"])
        """
        if not item_name:
            raise ValueError("商品名は空にできません。")
        if not condition:
            raise ValueError("商品の状態は空にできません。")

        condition_info = CONDITION_PRICE_MULTIPLIER.get(
            condition,
            {"label": f"{condition}の状態です。"},
        )
        condition_description = condition_info["label"]

        hashtags = self._generate_hashtags(item_name, brand)
        hashtag_text = " ".join(hashtags)

        # タイトル部分
        if brand:
            title = f"【{brand}】{item_name}"
        else:
            title = item_name

        # テンプレートで説明文を組み立て
        lines = [
            title,
            "",
            "ご覧いただきありがとうございます。",
            "",
            "■ 商品名",
            item_name,
            "",
        ]

        if brand:
            lines.extend(["■ ブランド", brand, ""])

        lines.extend(
            [
                "■ 商品の状態",
                condition,
                "",
                "■ 商品説明",
            ]
        )

        if brand:
            lines.append(f"{brand}の{item_name}です。")
        else:
            lines.append(f"{item_name}です。")

        lines.extend(
            [
                condition_description,
                "",
                "■ 発送について",
                "・匿名配送対応",
                "・24時間以内に発送予定",
                "・丁寧に梱包してお届けします",
                "",
                hashtag_text,
            ]
        )

        text = "\n".join(lines)

        return {
            "text": text,
            "hashtags": hashtags,
            "character_count": len(text),
        }

    def suggest_price(
        self,
        item_name: str,
        condition: str,
    ) -> dict[str, Any]:
        """適正価格を提案する。

        商品名と状態から、メルカリでの適正な出品価格を推定します。
        ブランドの基準価格と状態による減価率を元に計算します。

        Args:
            item_name: 商品名。
            condition: 商品の状態。

        Returns:
            価格提案情報を含む辞書。以下のキーを含む:
            - suggested_price (int): 推奨価格（円）
            - min_price (int): 最低推奨価格（円）
            - max_price (int): 最高推奨価格（円）
            - base_price (int): 推定基準価格（円）
            - condition_factor (str): 状態による価格調整の説明

        Raises:
            ValueError: 商品名または状態が空の場合。

        Example:
            >>> price = skill.suggest_price("AirPods Pro 第2世代", "目立った傷や汚れなし")
            >>> print(f"推奨価格: {price['suggested_price']}円")
        """
        if not item_name:
            raise ValueError("商品名は空にできません。")
        if not condition:
            raise ValueError("商品の状態は空にできません。")

        base_price = self._estimate_base_price(item_name)

        condition_info = CONDITION_PRICE_MULTIPLIER.get(
            condition,
            {"min": 0.30, "max": 0.50},
        )

        min_multiplier = condition_info["min"]
        max_multiplier = condition_info["max"]
        avg_multiplier = (min_multiplier + max_multiplier) / 2

        min_price = int(base_price * min_multiplier)
        max_price = int(base_price * max_multiplier)
        suggested_price = int(base_price * avg_multiplier)

        # 100円単位に丸める
        suggested_price = max(round(suggested_price / 100) * 100, 300)
        min_price = max(round(min_price / 100) * 100, 300)
        max_price = max(round(max_price / 100) * 100, 300)

        return {
            "suggested_price": suggested_price,
            "min_price": min_price,
            "max_price": max_price,
            "base_price": base_price,
            "condition_factor": f"状態「{condition}」による価格倍率: {min_multiplier:.0%}-{max_multiplier:.0%}",
        }

    def generate_listing(
        self,
        item_name: str,
        condition: str,
        brand: Optional[str] = None,
        photos: Optional[list[str]] = None,
    ) -> dict[str, Any]:
        """出品情報を一括生成する。

        商品説明文、カテゴリ推定、価格提案をまとめて生成します。
        メルカリ出品に必要な情報を一度に準備できます。

        Args:
            item_name: 商品名。
            condition: 商品の状態。
            brand: ブランド名（任意）。
            photos: 写真ファイルパスのリスト（任意）。

        Returns:
            出品情報を含む辞書。以下のキーを含む:
            - description (str): 商品説明文
            - category (str): 推定カテゴリ
            - price (int): 推奨出品価格（円）
            - price_range (dict): 価格レンジ（min, max）
            - photo_count (int): 写真枚数
            - hashtags (list[str]): ハッシュタグ
            - tips (list[str]): 出品のコツ

        Raises:
            ValueError: 商品名または状態が空の場合。

        Example:
            >>> listing = skill.generate_listing(
            ...     "ナイキ エアマックス90 27cm",
            ...     "やや傷や汚れあり",
            ...     brand="NIKE",
            ...     photos=["photo1.jpg", "photo2.jpg"]
            ... )
            >>> print(listing["description"])
        """
        if not item_name:
            raise ValueError("商品名は空にできません。")
        if not condition:
            raise ValueError("商品の状態は空にできません。")

        # 説明文の生成
        description_result = self.generate_description(item_name, condition, brand)

        # カテゴリの推定
        category = self._estimate_category(item_name, brand)

        # 価格の提案
        price_result = self.suggest_price(item_name, condition)

        # 写真に関する情報
        photo_list = photos or []
        photo_count = len(photo_list)

        # 出品のコツを生成
        tips = self._generate_listing_tips(condition, photo_count)

        return {
            "description": description_result["text"],
            "category": category,
            "price": price_result["suggested_price"],
            "price_range": {
                "min": price_result["min_price"],
                "max": price_result["max_price"],
            },
            "photo_count": photo_count,
            "photos": photo_list,
            "hashtags": description_result["hashtags"],
            "tips": tips,
        }

    def _generate_listing_tips(
        self, condition: str, photo_count: int
    ) -> list[str]:
        """出品のコツを生成する。

        Args:
            condition: 商品の状態。
            photo_count: 写真の枚数。

        Returns:
            出品のコツのリスト。
        """
        tips = []

        if photo_count == 0:
            tips.append("写真を追加してください。写真があると売れやすくなります（推奨: 4枚以上）。")
        elif photo_count < 4:
            tips.append(f"現在{photo_count}枚の写真があります。4枚以上あると売れやすくなります。")
        else:
            tips.append(f"{photo_count}枚の写真が設定されています。")

        if condition in ["やや傷や汚れあり", "傷や汚れあり", "全体的に状態が悪い"]:
            tips.append("傷や汚れがある場合は、該当箇所の写真を追加すると購入者の安心感が高まります。")

        tips.append("タイトルにブランド名・サイズ・色を含めると検索に引っかかりやすくなります。")
        tips.append("週末（金曜夜〜日曜）に出品すると閲覧数が上がる傾向があります。")

        return tips
