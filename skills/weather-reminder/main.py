"""Weather Reminder スキル

天気予報を取得してリマインダーを設定するスキル。
OpenWeatherMap APIを使用。傘リマインダー、洗濯物アラート、熱中症警告に対応。
"""

import logging
import os
from datetime import datetime, timezone
from typing import Any, Optional

import requests

logger = logging.getLogger(__name__)

OPENWEATHERMAP_BASE_URL = "https://api.openweathermap.org/data/2.5"

# 熱中症リスクレベルの閾値
HEATSTROKE_THRESHOLDS = {
    "safe": {"max_wbgt": 21, "label": "安全", "advice": "特に注意は必要ありません。"},
    "caution": {
        "max_wbgt": 25,
        "label": "注意",
        "advice": "こまめに水分補給をしてください。",
    },
    "warning": {
        "max_wbgt": 28,
        "label": "警戒",
        "advice": "激しい運動は避け、適度に休憩を取ってください。",
    },
    "severe": {
        "max_wbgt": 31,
        "label": "厳重警戒",
        "advice": "外出を控え、涼しい環境で過ごしてください。水分・塩分の補給を忘れずに。",
    },
    "danger": {
        "max_wbgt": 99,
        "label": "危険",
        "advice": "外出を避けてください。エアコンの効いた室内で過ごし、こまめに水分・塩分を補給してください。",
    },
}

# 降水に関連する天気コード（OpenWeatherMap）
RAIN_WEATHER_CODES = {
    200, 201, 202, 210, 211, 212, 221, 230, 231, 232,  # 雷雨
    300, 301, 302, 310, 311, 312, 313, 314, 321,        # 霧雨
    500, 501, 502, 503, 504, 511, 520, 521, 522, 531,   # 雨
    600, 601, 602, 611, 612, 613, 615, 616, 620, 621, 622,  # 雪
}


class WeatherReminderSkill:
    """天気予報に基づくリマインダーを提供するスキル。

    OpenWeatherMap APIを利用して天気予報を取得し、
    傘リマインダー、洗濯物アラート、熱中症警告などを判定します。

    Attributes:
        api_key: OpenWeatherMap APIキー。
    """

    def __init__(self, api_key: Optional[str] = None) -> None:
        """WeatherReminderSkillを初期化する。

        Args:
            api_key: OpenWeatherMap APIキー。
                省略時は環境変数 OPENWEATHERMAP_API_KEY を使用。

        Raises:
            ValueError: APIキーが設定されていない場合。
        """
        self.api_key = api_key or os.environ.get("OPENWEATHERMAP_API_KEY")

        if not self.api_key:
            raise ValueError(
                "OpenWeatherMap APIキーが必要です。引数または環境変数 "
                "OPENWEATHERMAP_API_KEY で指定してください。"
            )

        self._session = requests.Session()

    def _fetch_current_weather(self, city: str) -> dict[str, Any]:
        """現在の天気情報を取得する。

        Args:
            city: 都市名（英語表記）。

        Returns:
            OpenWeatherMap APIからの現在の天気データ。

        Raises:
            requests.HTTPError: API呼び出しが失敗した場合。
        """
        url = f"{OPENWEATHERMAP_BASE_URL}/weather"
        params = {
            "q": city,
            "appid": self.api_key,
            "units": "metric",
            "lang": "ja",
        }
        response = self._session.get(url, params=params, timeout=30)
        response.raise_for_status()
        return response.json()

    def _fetch_forecast(self, city: str) -> dict[str, Any]:
        """5日間の天気予報を取得する。

        Args:
            city: 都市名（英語表記）。

        Returns:
            OpenWeatherMap APIからの予報データ。

        Raises:
            requests.HTTPError: API呼び出しが失敗した場合。
        """
        url = f"{OPENWEATHERMAP_BASE_URL}/forecast"
        params = {
            "q": city,
            "appid": self.api_key,
            "units": "metric",
            "lang": "ja",
        }
        response = self._session.get(url, params=params, timeout=30)
        response.raise_for_status()
        return response.json()

    def _estimate_wbgt(self, temperature: float, humidity: float) -> float:
        """WBGT（暑さ指数）を気温と湿度から推定する。

        簡易的な推定式を使用。正式なWBGTは黒球温度等も必要だが、
        日常的な目安として気温と湿度から算出する。

        Args:
            temperature: 気温（摂氏）。
            humidity: 湿度（%）。

        Returns:
            推定WBGT値（摂氏）。
        """
        return 0.725 * temperature + 0.0368 * humidity + 0.00364 * temperature * humidity - 3.246

    def get_forecast(self, city: str) -> dict[str, Any]:
        """指定した都市の天気予報を取得する。

        現在の天気情報と5日間の天気予報を取得し、整形して返します。

        Args:
            city: 都市名（英語表記。例: "Tokyo", "Osaka", "Nagoya"）。

        Returns:
            天気予報情報を含む辞書。以下のキーを含む:
            - city (str): 都市名
            - current (dict): 現在の天気（description, temperature, humidity, wind_speed, weather_code）
            - daily (list[dict]): 日別予報のリスト

        Raises:
            ValueError: 都市名が空の場合。

        Example:
            >>> skill = WeatherReminderSkill()
            >>> forecast = skill.get_forecast("Tokyo")
            >>> print(forecast["current"]["temperature"])
        """
        if not city:
            raise ValueError("都市名は空にできません。")

        try:
            current_data = self._fetch_current_weather(city)
            forecast_data = self._fetch_forecast(city)

            weather = current_data.get("weather", [{}])[0]
            main_data = current_data.get("main", {})

            current = {
                "description": weather.get("description", "不明"),
                "temperature": main_data.get("temp", 0),
                "feels_like": main_data.get("feels_like", 0),
                "humidity": main_data.get("humidity", 0),
                "wind_speed": current_data.get("wind", {}).get("speed", 0),
                "weather_code": weather.get("id", 0),
            }

            # 予報データを日別に集約
            daily_forecasts = {}
            for entry in forecast_data.get("list", []):
                dt = datetime.fromtimestamp(entry["dt"], tz=timezone.utc)
                date_key = dt.strftime("%Y-%m-%d")

                if date_key not in daily_forecasts:
                    daily_forecasts[date_key] = {
                        "date": date_key,
                        "temp_min": float("inf"),
                        "temp_max": float("-inf"),
                        "descriptions": [],
                        "weather_codes": [],
                        "rain_probability": 0,
                    }

                day = daily_forecasts[date_key]
                temp = entry.get("main", {}).get("temp", 0)
                day["temp_min"] = min(day["temp_min"], temp)
                day["temp_max"] = max(day["temp_max"], temp)

                entry_weather = entry.get("weather", [{}])[0]
                day["descriptions"].append(entry_weather.get("description", ""))
                day["weather_codes"].append(entry_weather.get("id", 0))

                pop = entry.get("pop", 0) * 100
                day["rain_probability"] = max(day["rain_probability"], pop)

            daily = []
            for date_key in sorted(daily_forecasts.keys()):
                day = daily_forecasts[date_key]
                # 最も頻出する天気説明を代表値にする
                descriptions = day["descriptions"]
                most_common = max(set(descriptions), key=descriptions.count) if descriptions else "不明"
                daily.append(
                    {
                        "date": day["date"],
                        "description": most_common,
                        "temp_min": round(day["temp_min"], 1),
                        "temp_max": round(day["temp_max"], 1),
                        "rain_probability": round(day["rain_probability"], 0),
                    }
                )

            return {"city": city, "current": current, "daily": daily}

        except requests.HTTPError as e:
            logger.error("天気予報の取得に失敗しました: city=%s, error=%s", city, e)
            return {"city": city, "current": {}, "daily": [], "error": str(e)}
        except requests.RequestException as e:
            logger.error("ネットワークエラー: %s", e)
            return {"city": city, "current": {}, "daily": [], "error": str(e)}

    def check_umbrella(self, city: str) -> dict[str, Any]:
        """傘の必要性を判定する。

        今日の天気予報を確認し、雨の可能性がある場合に傘リマインダーを返します。
        降水確率30%以上で傘の携帯を推奨します。

        Args:
            city: 都市名（英語表記）。

        Returns:
            傘リマインダー情報を含む辞書。以下のキーを含む:
            - needed (bool): 傘が必要かどうか
            - reason (str): 判定理由
            - rain_probability (float): 降水確率（%）
            - current_weather (str): 現在の天気

        Raises:
            ValueError: 都市名が空の場合。

        Example:
            >>> result = skill.check_umbrella("Osaka")
            >>> if result["needed"]:
            ...     print(f"傘を持っていきましょう！{result['reason']}")
        """
        if not city:
            raise ValueError("都市名は空にできません。")

        try:
            forecast = self.get_forecast(city)

            if "error" in forecast:
                return {
                    "needed": False,
                    "reason": "天気情報を取得できませんでした。",
                    "rain_probability": 0,
                    "current_weather": "",
                    "error": forecast["error"],
                }

            current = forecast.get("current", {})
            weather_code = current.get("weather_code", 0)
            current_description = current.get("description", "不明")

            # 現在の天気が雨関連かチェック
            is_raining_now = weather_code in RAIN_WEATHER_CODES

            # 今日の予報から降水確率を取得
            daily = forecast.get("daily", [])
            today_rain_probability = daily[0]["rain_probability"] if daily else 0

            # 判定
            if is_raining_now:
                return {
                    "needed": True,
                    "reason": f"現在、{current_description}です。傘を持っていきましょう。",
                    "rain_probability": max(today_rain_probability, 80),
                    "current_weather": current_description,
                }
            elif today_rain_probability >= 60:
                return {
                    "needed": True,
                    "reason": f"降水確率が{today_rain_probability:.0f}%です。傘を持っていくことを強くお勧めします。",
                    "rain_probability": today_rain_probability,
                    "current_weather": current_description,
                }
            elif today_rain_probability >= 30:
                return {
                    "needed": True,
                    "reason": f"降水確率が{today_rain_probability:.0f}%です。折りたたみ傘を持っていくと安心です。",
                    "rain_probability": today_rain_probability,
                    "current_weather": current_description,
                }
            else:
                return {
                    "needed": False,
                    "reason": "今日は雨の心配はなさそうです。",
                    "rain_probability": today_rain_probability,
                    "current_weather": current_description,
                }

        except Exception as e:
            logger.error("傘リマインダーの判定に失敗しました: %s", e)
            return {
                "needed": False,
                "reason": "判定に失敗しました。",
                "rain_probability": 0,
                "current_weather": "",
                "error": str(e),
            }

    def check_laundry(self, city: str) -> dict[str, Any]:
        """洗濯物の外干し適性を判定する。

        天候、気温、湿度、風速を総合的に評価し、洗濯物を外に干せるかどうかを判定します。
        乾燥指数（0-100）で乾きやすさを数値化します。

        Args:
            city: 都市名（英語表記）。

        Returns:
            洗濯物アラート情報を含む辞書。以下のキーを含む:
            - recommended (bool): 外干しを推奨するかどうか
            - advice (str): アドバイスメッセージ
            - drying_index (int): 乾燥指数（0-100、高いほど乾きやすい）
            - conditions (dict): 現在の気象条件

        Raises:
            ValueError: 都市名が空の場合。

        Example:
            >>> result = skill.check_laundry("Nagoya")
            >>> print(f"乾燥指数: {result['drying_index']}")
        """
        if not city:
            raise ValueError("都市名は空にできません。")

        try:
            forecast = self.get_forecast(city)

            if "error" in forecast:
                return {
                    "recommended": False,
                    "advice": "天気情報を取得できませんでした。",
                    "drying_index": 0,
                    "conditions": {},
                    "error": forecast["error"],
                }

            current = forecast.get("current", {})
            temperature = current.get("temperature", 0)
            humidity = current.get("humidity", 0)
            wind_speed = current.get("wind_speed", 0)
            weather_code = current.get("weather_code", 0)

            conditions = {
                "temperature": temperature,
                "humidity": humidity,
                "wind_speed": wind_speed,
                "weather": current.get("description", "不明"),
            }

            # 雨の場合は外干し不可
            if weather_code in RAIN_WEATHER_CODES:
                return {
                    "recommended": False,
                    "advice": "現在雨が降っています。室内干しをお勧めします。",
                    "drying_index": 0,
                    "conditions": conditions,
                }

            # 乾燥指数を計算（気温、湿度、風速から総合評価）
            temp_score = min(max((temperature - 5) / 30 * 40, 0), 40)
            humidity_score = max((100 - humidity) / 100 * 35, 0)
            wind_score = min(wind_speed / 5 * 25, 25)
            drying_index = int(temp_score + humidity_score + wind_score)
            drying_index = min(max(drying_index, 0), 100)

            # 今日の降水確率も考慮
            daily = forecast.get("daily", [])
            today_rain_prob = daily[0]["rain_probability"] if daily else 0

            if today_rain_prob >= 50:
                return {
                    "recommended": False,
                    "advice": f"午後の降水確率が{today_rain_prob:.0f}%です。室内干しをお勧めします。",
                    "drying_index": max(drying_index - 30, 0),
                    "conditions": conditions,
                }
            elif drying_index >= 60:
                return {
                    "recommended": True,
                    "advice": "絶好の洗濯日和です！外干しをお勧めします。",
                    "drying_index": drying_index,
                    "conditions": conditions,
                }
            elif drying_index >= 40:
                return {
                    "recommended": True,
                    "advice": "外干しは可能ですが、厚手の衣類は乾きにくいかもしれません。",
                    "drying_index": drying_index,
                    "conditions": conditions,
                }
            else:
                return {
                    "recommended": False,
                    "advice": "気温が低く湿度が高いため、室内干しまたは乾燥機の使用をお勧めします。",
                    "drying_index": drying_index,
                    "conditions": conditions,
                }

        except Exception as e:
            logger.error("洗濯物アラートの判定に失敗しました: %s", e)
            return {
                "recommended": False,
                "advice": "判定に失敗しました。",
                "drying_index": 0,
                "conditions": {},
                "error": str(e),
            }

    def check_heatstroke(self, city: str) -> dict[str, Any]:
        """熱中症リスクを判定する。

        気温と湿度からWBGT（暑さ指数）を推定し、熱中症リスクレベルを判定します。
        リスクレベルに応じたアドバイスを提供します。

        Args:
            city: 都市名（英語表記）。

        Returns:
            熱中症警告情報を含む辞書。以下のキーを含む:
            - risk_level (str): リスクレベル（安全/注意/警戒/厳重警戒/危険）
            - wbgt_estimate (float): 推定WBGT値（摂氏）
            - advice (str): アドバイスメッセージ
            - conditions (dict): 現在の気象条件

        Raises:
            ValueError: 都市名が空の場合。

        Example:
            >>> result = skill.check_heatstroke("Fukuoka")
            >>> print(f"リスクレベル: {result['risk_level']}")
        """
        if not city:
            raise ValueError("都市名は空にできません。")

        try:
            current_data = self._fetch_current_weather(city)

            main_data = current_data.get("main", {})
            temperature = main_data.get("temp", 0)
            humidity = main_data.get("humidity", 0)

            conditions = {
                "temperature": temperature,
                "feels_like": main_data.get("feels_like", 0),
                "humidity": humidity,
            }

            wbgt = self._estimate_wbgt(temperature, humidity)
            wbgt = round(wbgt, 1)

            # リスクレベルを判定
            risk_level = "安全"
            advice = HEATSTROKE_THRESHOLDS["safe"]["advice"]

            for level_key in ["danger", "severe", "warning", "caution", "safe"]:
                threshold = HEATSTROKE_THRESHOLDS[level_key]
                if level_key == "safe" and wbgt < threshold["max_wbgt"]:
                    risk_level = threshold["label"]
                    advice = threshold["advice"]
                    break
                elif level_key == "danger" and wbgt >= HEATSTROKE_THRESHOLDS["severe"]["max_wbgt"]:
                    risk_level = threshold["label"]
                    advice = threshold["advice"]
                    break
                elif level_key == "severe" and wbgt >= HEATSTROKE_THRESHOLDS["warning"]["max_wbgt"]:
                    risk_level = threshold["label"]
                    advice = threshold["advice"]
                    break
                elif level_key == "warning" and wbgt >= HEATSTROKE_THRESHOLDS["caution"]["max_wbgt"]:
                    risk_level = threshold["label"]
                    advice = threshold["advice"]
                    break
                elif level_key == "caution" and wbgt >= HEATSTROKE_THRESHOLDS["safe"]["max_wbgt"]:
                    risk_level = threshold["label"]
                    advice = threshold["advice"]
                    break

            return {
                "risk_level": risk_level,
                "wbgt_estimate": wbgt,
                "advice": advice,
                "conditions": conditions,
            }

        except requests.HTTPError as e:
            logger.error("熱中症リスクの判定に失敗しました: city=%s, error=%s", city, e)
            return {
                "risk_level": "不明",
                "wbgt_estimate": 0,
                "advice": "天気情報を取得できませんでした。",
                "conditions": {},
                "error": str(e),
            }
        except requests.RequestException as e:
            logger.error("ネットワークエラー: %s", e)
            return {
                "risk_level": "不明",
                "wbgt_estimate": 0,
                "advice": "ネットワークエラーが発生しました。",
                "conditions": {},
                "error": str(e),
            }
