# Mercari Lister スキル

メルカリ出品テキストを自動生成するためのスキルです。商品名、状態、ブランド名から説明文を作成し、適切なカテゴリと価格を提案します。

## 概要

このスキルは、メルカリへの出品作業を効率化します。商品情報を入力するだけで、魅力的な商品説明文を自動生成し、市場価格に基づいた適正価格を提案します。

### 主な機能

- 商品説明文の自動生成（テンプレートベース）
- 商品状態・ブランドに基づく価格提案
- 出品情報の一括生成（説明文＋カテゴリ＋価格）

## セットアップ

### 1. インストール

```bash
pip install -r requirements.txt
```

### 2. 環境変数の設定（任意）

価格提案の精度を向上させるために、外部APIキーを設定できます：

```
MERCARI_PRICE_API_KEY=your_api_key  # 任意: 価格分析APIキー
```

**注意**: 環境変数が未設定でもスキルは動作します。その場合、内蔵の価格推定ロジックを使用します。

## 使用例

### 商品説明文の生成

```python
from main import MercariListerSkill

skill = MercariListerSkill()

# 商品説明文を生成
description = skill.generate_description(
    item_name="ワイヤレスイヤホン AirPods Pro 第2世代",
    condition="目立った傷や汚れなし",
    brand="Apple"
)
print(description["text"])
```

出力例:
```
【Apple】ワイヤレスイヤホン AirPods Pro 第2世代

ご覧いただきありがとうございます。

■ 商品名
ワイヤレスイヤホン AirPods Pro 第2世代

■ ブランド
Apple

■ 商品の状態
目立った傷や汚れなし

■ 商品説明
Appleのワイヤレスイヤホン AirPods Pro 第2世代です。
目立った傷や汚れはなく、良好な状態です。

■ 発送について
・匿名配送対応
・24時間以内に発送予定
・丁寧に梱包してお届けします

#Apple #ワイヤレスイヤホン #AirPods
```

### 価格提案

```python
# 適正価格を提案
price = skill.suggest_price(
    item_name="AirPods Pro 第2世代",
    condition="目立った傷や汚れなし"
)
print(f"推奨価格: {price['suggested_price']}円")
print(f"価格レンジ: {price['min_price']}円 - {price['max_price']}円")
```

### 出品情報の一括生成

```python
# 説明文・カテゴリ・価格をまとめて生成
listing = skill.generate_listing(
    item_name="ナイキ エアマックス90 27cm",
    condition="やや傷や汚れあり",
    brand="NIKE",
    photos=["photo1.jpg", "photo2.jpg"]
)
print(f"説明文: {listing['description']}")
print(f"カテゴリ: {listing['category']}")
print(f"推奨価格: {listing['price']}円")
print(f"写真枚数: {listing['photo_count']}枚")
```

## API リファレンス

### `MercariListerSkill()`

スキルのインスタンスを作成します。パラメータは不要です。

### `generate_description(item_name, condition, brand=None)`

商品説明文を生成します。

| パラメータ | 型 | 必須 | 説明 |
|---|---|---|---|
| `item_name` | str | はい | 商品名 |
| `condition` | str | はい | 商品の状態（例: "新品、未使用", "目立った傷や汚れなし"） |
| `brand` | str | いいえ | ブランド名 |

**戻り値**: `dict` - 生成された説明文（text, hashtags を含む）

### `suggest_price(item_name, condition)`

適正価格を提案します。

| パラメータ | 型 | 必須 | 説明 |
|---|---|---|---|
| `item_name` | str | はい | 商品名 |
| `condition` | str | はい | 商品の状態 |

**戻り値**: `dict` - 価格提案（suggested_price, min_price, max_price を含む）

### `generate_listing(item_name, condition, brand=None, photos=None)`

出品情報を一括生成します。

| パラメータ | 型 | 必須 | 説明 |
|---|---|---|---|
| `item_name` | str | はい | 商品名 |
| `condition` | str | はい | 商品の状態 |
| `brand` | str | いいえ | ブランド名 |
| `photos` | list[str] | いいえ | 写真ファイルパスのリスト |

**戻り値**: `dict` - 出品情報（description, category, price, photo_count を含む）

## 商品状態の一覧

メルカリで使用される商品状態は以下の通りです：

| 状態 | 価格への影響 |
|---|---|
| 新品、未使用 | 定価の70-90% |
| 未使用に近い | 定価の60-80% |
| 目立った傷や汚れなし | 定価の40-70% |
| やや傷や汚れあり | 定価の30-50% |
| 傷や汚れあり | 定価の15-35% |
| 全体的に状態が悪い | 定価の5-20% |

## 注意事項

- 生成された説明文はテンプレートベースです。必要に応じて編集してください。
- 価格提案は市場の一般的な傾向に基づく推定値です。実際の売れ行きは異なる場合があります。
- 写真の解析機能は将来のバージョンで対応予定です。
- メルカリの利用規約に従って出品してください。

## ライセンス

MIT License
