"""Rakuten Shopping スキル

楽天市場の商品検索・価格比較スキル。
カテゴリ別検索、価格帯フィルター、ポイント還元率の表示に対応。
"""

import logging
import os
from typing import Any, Optional

import requests

logger = logging.getLogger(__name__)

RAKUTEN_API_BASE_URL = "https://app.rakuten.co.jp/services/api"
ICHIBA_SEARCH_ENDPOINT = f"{RAKUTEN_API_BASE_URL}/IchibaItem/Search/20220601"


class RakutenShoppingSkill:
    """楽天市場の商品検索・価格比較を行うスキル。

    楽天商品検索APIを利用して、商品の検索、価格比較、
    ポイント還元率の確認を行います。

    Attributes:
        app_id: 楽天アプリケーションID。
        affiliate_id: 楽天アフィリエイトID（任意）。
    """

    def __init__(
        self,
        app_id: Optional[str] = None,
        affiliate_id: Optional[str] = None,
    ) -> None:
        """RakutenShoppingSkillを初期化する。

        Args:
            app_id: 楽天アプリケーションID。
                省略時は環境変数 RAKUTEN_APP_ID を使用。
            affiliate_id: 楽天アフィリエイトID。
                省略時は環境変数 RAKUTEN_AFFILIATE_ID を使用。

        Raises:
            ValueError: アプリケーションIDが設定されていない場合。
        """
        self.app_id = app_id or os.environ.get("RAKUTEN_APP_ID")
        self.affiliate_id = affiliate_id or os.environ.get("RAKUTEN_AFFILIATE_ID")

        if not self.app_id:
            raise ValueError(
                "楽天アプリケーションIDが必要です。引数または環境変数 "
                "RAKUTEN_APP_ID で指定してください。"
            )

        self._session = requests.Session()

    def _build_base_params(self) -> dict[str, str]:
        """API呼び出しに必要な基本パラメータを構築する。

        Returns:
            基本パラメータの辞書。
        """
        params = {
            "applicationId": self.app_id,
            "format": "json",
            "formatVersion": "2",
        }
        if self.affiliate_id:
            params["affiliateId"] = self.affiliate_id
        return params

    def _parse_item(self, raw_item: dict[str, Any]) -> dict[str, Any]:
        """API レスポンスの商品データを整形する。

        Args:
            raw_item: 楽天APIから返された生の商品データ。

        Returns:
            整形済みの商品情報辞書。
        """
        return {
            "name": raw_item.get("itemName", ""),
            "price": raw_item.get("itemPrice", 0),
            "item_code": raw_item.get("itemCode", ""),
            "item_url": raw_item.get("itemUrl", ""),
            "shop_name": raw_item.get("shopName", ""),
            "shop_url": raw_item.get("shopUrl", ""),
            "image_url": (raw_item.get("mediumImageUrls") or [""])[0]
            if raw_item.get("mediumImageUrls")
            else "",
            "review_average": raw_item.get("reviewAverage", 0.0),
            "review_count": raw_item.get("reviewCount", 0),
            "point": raw_item.get("pointRate", 1),
            "availability": raw_item.get("availability", 0) == 1,
        }

    def search(
        self,
        keyword: str,
        category: Optional[str] = None,
        min_price: Optional[int] = None,
        max_price: Optional[int] = None,
    ) -> dict[str, Any]:
        """楽天市場の商品を検索する。

        指定されたキーワード、カテゴリ、価格帯で楽天市場の商品を検索します。
        結果は最大30件返却されます。

        Args:
            keyword: 検索キーワード。
            category: 楽天カテゴリID（例: "100227"）。省略時は全カテゴリ。
            min_price: 最低価格（円）。省略時は下限なし。
            max_price: 最高価格（円）。省略時は上限なし。

        Returns:
            検索結果を含む辞書。以下のキーを含む:
            - items (list[dict]): 商品情報のリスト
            - total_count (int): 検索結果の総件数
            - page_info (dict): ページング情報

        Raises:
            ValueError: キーワードが空の場合。

        Example:
            >>> skill = RakutenShoppingSkill()
            >>> results = skill.search("コーヒー豆", min_price=1000, max_price=3000)
            >>> print(results["total_count"])
        """
        if not keyword:
            raise ValueError("検索キーワードは空にできません。")

        params = self._build_base_params()
        params["keyword"] = keyword
        params["hits"] = "30"

        if category:
            params["genreId"] = category
        if min_price is not None:
            if min_price < 0:
                raise ValueError("最低価格は0以上で指定してください。")
            params["minPrice"] = str(min_price)
        if max_price is not None:
            if max_price < 0:
                raise ValueError("最高価格は0以上で指定してください。")
            params["maxPrice"] = str(max_price)
        if min_price is not None and max_price is not None and min_price > max_price:
            raise ValueError("最低価格は最高価格以下で指定してください。")

        try:
            response = self._session.get(
                ICHIBA_SEARCH_ENDPOINT, params=params, timeout=30
            )
            response.raise_for_status()
            data = response.json()

            items = [self._parse_item(item) for item in data.get("Items", [])]

            return {
                "items": items,
                "total_count": data.get("count", 0),
                "page_info": {
                    "page": data.get("page", 1),
                    "page_count": data.get("pageCount", 0),
                    "hits": data.get("hits", 0),
                },
            }
        except requests.HTTPError as e:
            logger.error("楽天API呼び出しに失敗しました: %s", e)
            return {"items": [], "total_count": 0, "page_info": {}, "error": str(e)}
        except requests.RequestException as e:
            logger.error("ネットワークエラー: %s", e)
            return {"items": [], "total_count": 0, "page_info": {}, "error": str(e)}

    def compare_prices(self, keyword: str) -> dict[str, Any]:
        """同一キーワードの商品を価格の安い順にソートして比較する。

        指定したキーワードで検索し、結果を価格の昇順でソートして返します。
        ショップ名、価格、ポイント還元率を含む比較情報を提供します。

        Args:
            keyword: 比較対象の商品キーワード。

        Returns:
            価格比較結果を含む辞書。以下のキーを含む:
            - items (list[dict]): 価格の安い順にソートされた商品リスト
            - lowest_price (int): 最安値
            - highest_price (int): 最高値
            - average_price (float): 平均価格

        Raises:
            ValueError: キーワードが空の場合。

        Example:
            >>> comparison = skill.compare_prices("Nintendo Switch")
            >>> print(f"最安値: {comparison['lowest_price']}円")
        """
        if not keyword:
            raise ValueError("検索キーワードは空にできません。")

        params = self._build_base_params()
        params["keyword"] = keyword
        params["hits"] = "30"
        params["sort"] = "+itemPrice"

        try:
            response = self._session.get(
                ICHIBA_SEARCH_ENDPOINT, params=params, timeout=30
            )
            response.raise_for_status()
            data = response.json()

            items = [self._parse_item(item) for item in data.get("Items", [])]

            prices = [item["price"] for item in items if item["price"] > 0]

            return {
                "items": items,
                "lowest_price": min(prices) if prices else 0,
                "highest_price": max(prices) if prices else 0,
                "average_price": round(sum(prices) / len(prices), 0) if prices else 0,
                "total_count": len(items),
            }
        except requests.HTTPError as e:
            logger.error("価格比較に失敗しました: %s", e)
            return {
                "items": [],
                "lowest_price": 0,
                "highest_price": 0,
                "average_price": 0,
                "total_count": 0,
                "error": str(e),
            }
        except requests.RequestException as e:
            logger.error("ネットワークエラー: %s", e)
            return {
                "items": [],
                "lowest_price": 0,
                "highest_price": 0,
                "average_price": 0,
                "total_count": 0,
                "error": str(e),
            }

    def get_point_rate(self, item_code: str) -> dict[str, Any]:
        """指定商品のポイント還元率を取得する。

        楽天市場の商品コードを指定して、その商品のポイント還元情報を取得します。
        通常ポイント、ボーナスポイント、合計ポイント倍率を返します。

        Args:
            item_code: 楽天商品コード（例: "shop_12345"）。

        Returns:
            ポイント還元情報を含む辞書。以下のキーを含む:
            - item_name (str): 商品名
            - base_rate (int): 通常ポイント倍率
            - bonus_rate (int): ボーナスポイント倍率
            - total_rate (int): 合計ポイント倍率
            - estimated_points (int): 推定獲得ポイント数

        Raises:
            ValueError: 商品コードが空の場合。

        Example:
            >>> info = skill.get_point_rate("shop_12345")
            >>> print(f"合計ポイント倍率: {info['total_rate']}倍")
        """
        if not item_code:
            raise ValueError("商品コードは空にできません。")

        params = self._build_base_params()
        params["itemCode"] = item_code

        try:
            response = self._session.get(
                ICHIBA_SEARCH_ENDPOINT, params=params, timeout=30
            )
            response.raise_for_status()
            data = response.json()

            items = data.get("Items", [])
            if not items:
                return {
                    "item_name": "",
                    "base_rate": 0,
                    "bonus_rate": 0,
                    "total_rate": 0,
                    "estimated_points": 0,
                    "error": "商品が見つかりませんでした。",
                }

            item = items[0]
            base_rate = item.get("pointRate", 1)
            bonus_rate = item.get("pointRateStartTime", 0)
            # ボーナスポイントがある場合は pointRate に含まれるケースを考慮
            if isinstance(bonus_rate, str):
                bonus_rate = 0
            total_rate = base_rate + bonus_rate

            price = item.get("itemPrice", 0)
            estimated_points = int(price * total_rate / 100)

            return {
                "item_name": item.get("itemName", ""),
                "price": price,
                "base_rate": base_rate,
                "bonus_rate": bonus_rate,
                "total_rate": total_rate,
                "estimated_points": estimated_points,
            }
        except requests.HTTPError as e:
            logger.error("ポイント情報の取得に失敗しました: %s", e)
            return {
                "item_name": "",
                "base_rate": 0,
                "bonus_rate": 0,
                "total_rate": 0,
                "estimated_points": 0,
                "error": str(e),
            }
        except requests.RequestException as e:
            logger.error("ネットワークエラー: %s", e)
            return {
                "item_name": "",
                "base_rate": 0,
                "bonus_rate": 0,
                "total_rate": 0,
                "estimated_points": 0,
                "error": str(e),
            }
