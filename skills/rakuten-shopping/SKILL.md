# Rakuten Shopping スキル

楽天市場の商品検索・価格比較を行うためのスキルです。楽天商品検索APIを使用して、カテゴリ別検索、価格帯フィルター、ポイント還元率の表示に対応しています。

## 概要

このスキルを使用すると、PicoClawから楽天市場の膨大な商品データベースにアクセスし、商品の検索、価格比較、ポイント還元率の確認を簡単に行うことができます。

### 主な機能

- キーワード・カテゴリによる商品検索
- 価格帯を指定した絞り込み検索
- 同一商品の価格比較
- ポイント還元率の確認

## セットアップ

### 1. 楽天APIアプリケーションの登録

1. [楽天ウェブサービス](https://webservice.rakuten.co.jp/) にアクセスします。
2. 楽天IDでログインします（未登録の場合は新規登録）。
3. 「アプリID発行」からアプリケーションを登録します。
4. 発行された **アプリID（Application ID）** を控えます。

### 2. 環境変数の設定

以下の環境変数を設定してください：

```
RAKUTEN_APP_ID=your_application_id
RAKUTEN_AFFILIATE_ID=your_affiliate_id  # 任意
```

## 使用例

### 商品検索

```python
from main import RakutenShoppingSkill

skill = RakutenShoppingSkill()

# キーワードで検索
results = skill.search(keyword="ワイヤレスイヤホン")
for item in results["items"]:
    print(f"{item['name']} - {item['price']}円")

# カテゴリと価格帯を指定して検索
results = skill.search(
    keyword="コーヒー豆",
    category="100227",
    min_price=1000,
    max_price=3000
)
```

### 価格比較

```python
# 同一キーワードの商品を価格順で比較
comparison = skill.compare_prices(keyword="Nintendo Switch")
for item in comparison["items"]:
    print(f"{item['shop_name']}: {item['price']}円 (ポイント{item['point']}倍)")
```

### ポイント還元率の確認

```python
# 商品コードからポイント還元率を取得
point_info = skill.get_point_rate(item_code="shop_12345")
print(f"通常ポイント: {point_info['base_rate']}倍")
print(f"ボーナスポイント: {point_info['bonus_rate']}倍")
print(f"合計: {point_info['total_rate']}倍")
```

## API リファレンス

### `RakutenShoppingSkill(app_id=None, affiliate_id=None)`

スキルのインスタンスを作成します。

| パラメータ | 型 | 必須 | 説明 |
|---|---|---|---|
| `app_id` | str | いいえ | 楽天アプリID。省略時は環境変数 `RAKUTEN_APP_ID` を使用 |
| `affiliate_id` | str | いいえ | アフィリエイトID。省略時は環境変数 `RAKUTEN_AFFILIATE_ID` を使用 |

### `search(keyword, category=None, min_price=None, max_price=None)`

楽天市場の商品を検索します。

| パラメータ | 型 | 必須 | 説明 |
|---|---|---|---|
| `keyword` | str | はい | 検索キーワード |
| `category` | str | いいえ | 楽天カテゴリID |
| `min_price` | int | いいえ | 最低価格（円） |
| `max_price` | int | いいえ | 最高価格（円） |

**戻り値**: `dict` - 検索結果（items, total_count, page_info を含む）

### `compare_prices(keyword)`

同一キーワードの商品を価格の安い順にソートして比較します。

| パラメータ | 型 | 必須 | 説明 |
|---|---|---|---|
| `keyword` | str | はい | 比較対象の商品キーワード |

**戻り値**: `dict` - 価格比較結果（安い順にソート済み）

### `get_point_rate(item_code)`

指定商品のポイント還元率を取得します。

| パラメータ | 型 | 必須 | 説明 |
|---|---|---|---|
| `item_code` | str | はい | 楽天商品コード |

**戻り値**: `dict` - ポイント還元情報（base_rate, bonus_rate, total_rate を含む）

## カテゴリIDの一覧（主要なもの）

| カテゴリID | カテゴリ名 |
|---|---|
| 100371 | パソコン・周辺機器 |
| 100227 | 水・ソフトドリンク |
| 558885 | スマートフォン・タブレット |
| 100533 | キッチン用品 |
| 100316 | メンズファッション |
| 100431 | レディースファッション |

## 注意事項

- 楽天APIには1秒あたりのリクエスト数に制限があります（無料枠: 1リクエスト/秒）。
- 検索結果は最大30件/ページです。
- 価格はすべて税込み表示です。
- ポイント還元率はキャンペーン等により変動する場合があります。

## ライセンス

MIT License
