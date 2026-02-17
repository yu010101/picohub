# Weather Reminder スキル

天気予報を取得してリマインダーを設定するためのスキルです。OpenWeatherMap APIを使用して、傘リマインダー、洗濯物アラート、熱中症警告などの便利な通知機能を提供します。

## 概要

このスキルは、指定した都市の天気予報を取得し、天候に応じた各種リマインダーを自動的に判定します。日常生活に役立つ天気ベースのアドバイスを提供します。

### 主な機能

- 都市別の天気予報取得（現在の天気＋5日間予報）
- 雨の予報がある場合の傘リマインダー
- 天候に応じた洗濯物アラート
- 気温に基づく熱中症警告

## セットアップ

### 1. OpenWeatherMap APIキーの取得

1. [OpenWeatherMap](https://openweathermap.org/) にアクセスし、アカウントを作成します。
2. ログイン後、「API keys」タブからAPIキーを取得します。
3. 無料プランでも十分な機能を利用できます（60回/分のAPI呼び出し）。

### 2. 環境変数の設定

以下の環境変数を設定してください：

```
OPENWEATHERMAP_API_KEY=your_api_key
```

## 使用例

### 天気予報の取得

```python
from main import WeatherReminderSkill

skill = WeatherReminderSkill()

# 東京の天気予報を取得
forecast = skill.get_forecast(city="Tokyo")
print(f"現在の天気: {forecast['current']['description']}")
print(f"気温: {forecast['current']['temperature']}°C")
print(f"湿度: {forecast['current']['humidity']}%")

for day in forecast["daily"]:
    print(f"{day['date']}: {day['description']} ({day['temp_min']}°C - {day['temp_max']}°C)")
```

### 傘リマインダーの確認

```python
# 今日傘が必要かチェック
umbrella = skill.check_umbrella(city="Osaka")
print(f"傘が必要: {umbrella['needed']}")
print(f"理由: {umbrella['reason']}")
print(f"降水確率: {umbrella['rain_probability']}%")
```

### 洗濯物アラートの確認

```python
# 洗濯物を外に干せるかチェック
laundry = skill.check_laundry(city="Nagoya")
print(f"外干し推奨: {laundry['recommended']}")
print(f"アドバイス: {laundry['advice']}")
print(f"乾燥指数: {laundry['drying_index']}")
```

### 熱中症警告の確認

```python
# 熱中症リスクをチェック
heatstroke = skill.check_heatstroke(city="Fukuoka")
print(f"リスクレベル: {heatstroke['risk_level']}")
print(f"WBGT推定値: {heatstroke['wbgt_estimate']}°C")
print(f"アドバイス: {heatstroke['advice']}")
```

## API リファレンス

### `WeatherReminderSkill(api_key=None)`

スキルのインスタンスを作成します。

| パラメータ | 型 | 必須 | 説明 |
|---|---|---|---|
| `api_key` | str | いいえ | OpenWeatherMap APIキー。省略時は環境変数 `OPENWEATHERMAP_API_KEY` を使用 |

### `get_forecast(city)`

指定した都市の天気予報を取得します。

| パラメータ | 型 | 必須 | 説明 |
|---|---|---|---|
| `city` | str | はい | 都市名（英語表記。例: "Tokyo", "Osaka"） |

**戻り値**: `dict` - 現在の天気と5日間予報（current, daily を含む）

### `check_umbrella(city)`

傘の必要性を判定します。

| パラメータ | 型 | 必須 | 説明 |
|---|---|---|---|
| `city` | str | はい | 都市名（英語表記） |

**戻り値**: `dict` - 傘リマインダー情報（needed, reason, rain_probability を含む）

### `check_laundry(city)`

洗濯物の外干し適性を判定します。

| パラメータ | 型 | 必須 | 説明 |
|---|---|---|---|
| `city` | str | はい | 都市名（英語表記） |

**戻り値**: `dict` - 洗濯物アラート情報（recommended, advice, drying_index を含む）

### `check_heatstroke(city)`

熱中症リスクを判定します。

| パラメータ | 型 | 必須 | 説明 |
|---|---|---|---|
| `city` | str | はい | 都市名（英語表記） |

**戻り値**: `dict` - 熱中症警告情報（risk_level, wbgt_estimate, advice を含む）

## リスクレベルの定義

### 熱中症リスクレベル

| レベル | WBGT推定値 | 説明 |
|---|---|---|
| 安全 | 21°C未満 | 特に注意なし |
| 注意 | 21-25°C | こまめな水分補給を推奨 |
| 警戒 | 25-28°C | 激しい運動は避ける |
| 厳重警戒 | 28-31°C | 外出を控えることを推奨 |
| 危険 | 31°C以上 | 外出を避け、涼しい環境で過ごす |

## 注意事項

- 天気予報データは目安です。最新の情報は気象庁の公式発表を確認してください。
- WBGT（暑さ指数）は気温・湿度・風速から推定した値であり、正式な計測値ではありません。
- 都市名は英語表記で指定してください（日本語対応予定）。
- 無料プランのAPIキーでは呼び出し回数に制限があります。

## ライセンス

MIT License
