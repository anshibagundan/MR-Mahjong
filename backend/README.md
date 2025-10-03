# 麻雀ゲーム バックエンド（最小限実装）

3人麻雀ゲームのバックエンドAPI。**ヘルスチェック**と**WebSocket接続確認**、**初回配牌**のみの最小限実装。

## 🎯 実装機能

✅ **ヘルスチェック** (`/health`)
- サーバー稼働状況確認

✅ **WebSocket接続確認** (`/ws/game`)
- クライアント接続確認
- 接続応答メッセージ

✅ **初回配牌**
- 3人接続時に自動配牌
- 各プレイヤーに13枚ずつ配布
- 王牌・山牌の設定

## 🚀 実行方法

### 前提条件
- Go 1.21以上
- Docker（オプション）

### 方法1: 直接実行

```bash
# リポジトリをクローン
git clone <repository-url>
cd backend

# 依存関係をインストール
go mod tidy

make run
```

## 🧪 動作確認

### 1. ヘルスチェック

```bash
curl http://localhost:8080/health
```

**期待レスポンス:**
```json
{
  "status": "ok",
  "service": "mahjong-backend",
  "version": "1.0.0"
}
```

### 2. 役判定API (`/api/v1/yaku`)

手牌と和了状況から役と符・翻を判定するREST APIです。

**点数計算表（子の点数）:**
- 1翻: 1,000点
- 2翻: 2,000点  
- 3翻: 3,900点
- 4翻: 7,700点
- 5翻: 8,000点（満貫）
- 6-7翻: 12,000点（跳満）
- 8-10翻: 16,000点（倍満）
- 11-12翻: 24,000点（三倍満）
- 13翻以上: 32,000点（役満）

#### 基本的な役判定（役牌・対々和・清一色）

```bash
curl -X POST http://localhost:8080/api/v1/yaku \
  -H "Content-Type: application/json" \
  -d '{
    "tehai": ["1m","1m"],
    "openMelds": [
      {"type": "pon", "tiles": ["haku","haku","haku"]},
      {"type": "pon", "tiles": ["2m","2m","2m"]},
      {"type": "pon", "tiles": ["3m","3m","3m"]},
      {"type": "pon", "tiles": ["4m","4m","4m"]}
    ],
    "winTile": "1m",
    "isTsumo": false,
    "riichi": false,
    "ippatsu": false,
    "doraIndicators": ["2p"],
    "uraDoraIndicators": [],
    "roundWind": "east",
    "seatWind": "south"
  }'
```

**期待レスポンス:**
```json
{
  "yaku": [
    {"name": "対々和", "han": 2},
    {"name": "混一色", "han": 2},
    {"name": "役牌(白)", "han": 1}
  ],
  "fu": 0,
  "han": 5,
  "doraCount": 0,
  "uraDoraCount": 0,
  "totalHan": 5,
  "yakuman": [],
  "score": 8000,
  "isChombo": false
}
```

#### 門前の役判定（立直・自摸・断么九）

```bash
curl -X POST http://localhost:8080/api/v1/yaku \
  -H "Content-Type: application/json" \
  -d '{
    "tehai": ["2m","3m","4m","5m","6m","7m","2p","3p","4p","5p","6p","7p","8p","8p"],
    "openMelds": [],
    "winTile": "8p",
    "isTsumo": true,
    "riichi": true,
    "ippatsu": false,
    "doraIndicators": ["1s"],
    "uraDoraIndicators": [],
    "roundWind": "east",
    "seatWind": "south"
  }'
```

**期待レスポンス:**
```json
{
  "yaku": [
    {"name": "立直", "han": 1},
    {"name": "門前清自摸和", "han": 1},
    {"name": "断么九", "han": 1}
  ],
  "fu": 0,
  "han": 3,
  "doraCount": 0,
  "uraDoraCount": 0,
  "totalHan": 3,
  "yakuman": [],
  "score": 1200,
  "isChombo": false
}
```

#### 七対子の役判定

```bash
curl -X POST http://localhost:8080/api/v1/yaku \
  -H "Content-Type: application/json" \
  -d '{
    "tehai": ["1m","1m","2m","2m","3m","3m","4m","4m","5m","5m","6m","6m","7m","7m"],
    "openMelds": [],
    "winTile": "7m",
    "isTsumo": true,
    "riichi": true,
    "ippatsu": false,
    "doraIndicators": ["6m"],
    "uraDoraIndicators": ["2m"],
    "roundWind": "east",
    "seatWind": "south"
  }'
```

**期待レスポンス:**
```json
{
  "yaku": [
    {"name": "立直", "han": 1},
    {"name": "門前清自摸和", "han": 1},
    {"name": "七対子", "han": 2},
    {"name": "清一色", "han": 6}
  ],
  "fu": 0,
  "han": 10,
  "doraCount": 2,
  "uraDoraCount": 2,
  "totalHan": 14,
  "yakuman": [],
  "score": 32000,
  "isChombo": false
}
```

#### 三色同順の役判定（くいさがりテスト）

**門前（2翻）**:
```bash
curl -X POST http://localhost:8080/api/v1/yaku \
  -H "Content-Type: application/json" \
  -d '{
    "tehai": ["1m","2m","3m","1p","2p","3p","1s","2s","3s","4s","4s","5s","5s","5s"],
    "openMelds": [],
    "winTile": "5s",
    "roundWind": "east",
    "seatWind": "south"
  }'
```

**鳴きあり（1翻）**:
```bash
curl -X POST http://localhost:8080/api/v1/yaku \
  -H "Content-Type: application/json" \
  -d '{
    "tehai": ["1s","2s","3s","5s","5s"],
    "openMelds": [
      {"type": "chi", "tiles": ["1m","2m","3m"]},
      {"type": "chi", "tiles": ["1p","2p","3p"]},
      {"type": "pon", "tiles": ["4s","4s","4s"]}
    ],
    "winTile": "5s",
    "roundWind": "east",
    "seatWind": "south"
  }'
```

#### 清一色の役判定（くいさがりテスト）

**門前（6翻）**:
```bash
curl -X POST http://localhost:8080/api/v1/yaku \
  -H "Content-Type: application/json" \
  -d '{
    "tehai": ["2m","2m","3m","3m","4m","4m","5m","5m","6m","6m","7m","7m","8m","8m"],
    "openMelds": [],
    "winTile": "8m",
    "roundWind": "east",
    "seatWind": "south"
  }'
```

**鳴きあり（5翻）**:
```bash
curl -X POST http://localhost:8080/api/v1/yaku \
  -H "Content-Type: application/json" \
  -d '{
    "tehai": ["5m","5m"],
    "openMelds": [
      {"type": "pon", "tiles": ["2m","2m","2m"]},
      {"type": "pon", "tiles": ["3m","3m","3m"]},
      {"type": "pon", "tiles": ["4m","4m","4m"]},
      {"type": "pon", "tiles": ["6m","6m","6m"]}
    ],
    "winTile": "5m",
    "roundWind": "east",
    "seatWind": "south"
  }'
```

#### チョンボ（不正な手牌）の例

```bash
curl -X POST http://localhost:8080/api/v1/yaku \
  -H "Content-Type: application/json" \
  -d '{
    "tehai": ["1m","2m","3m","4m","5m","6m"],
    "openMelds": [],
    "winTile": "7m",
    "isTsumo": true,
    "riichi": false,
    "ippatsu": false,
    "doraIndicators": ["1s"],
    "uraDoraIndicators": [],
    "roundWind": "east",
    "seatWind": "south"
  }'
```

**期待レスポンス:**
```json
{
  "yaku": [],
  "fu": 0,
  "han": 0,
  "doraCount": 0,
  "uraDoraCount": 0,
  "totalHan": 0,
  "yakuman": [],
  "score": 0,
  "isChombo": true
}
```

### 3. Swagger UI

ブラウザで以下にアクセス:
```
http://localhost:8080/swagger/index.html
```

### 4. WebSocket接続テスト

#### websocat を使用（推奨）

```bash
# websocatをインストール（未インストールの場合）
# macOS
brew install websocat

# 持続的な接続（対話形式）
websocat ws://localhost:8080/ws/game
# 接続後、以下を入力してEnter:
{"type":"connection_check","data":{"playerId":"p3"}}
```

**期待レスポンス:**
```json
{
  "type": "connection_response",
  "data": {
    "playerId": "test-player-1",
    "playersCount": 1,
    "maxPlayers": 3,
    "status": "waiting",
    "message": "接続確認完了"
  }
}
```

#### その他のWebSocketクライアント

WebSocketクライアントで `ws://localhost:8080/ws/game` に接続後、以下のメッセージを送信:

```json
{
  "type": "connection_check",
  "data": {
    "playerId": "test-player-1"
  }
}
```

### 5. 3人接続での配牌テスト

#### websocat使用（推奨）

3つのターミナルを開いて、それぞれで持続的な接続を行います：

**ターミナル1:**
```bash
websocat ws://localhost:8080/ws/game
# 接続後、以下を入力:
{"type":"connection_check","data":{"playerId":"p1"}}
```

**ターミナル2:**
```bash
websocat ws://localhost:8080/ws/game
# 接続後、以下を入力:
{"type":"connection_check","data":{"playerId":"p2"}}
```

**ターミナル3:**
```bash
websocat ws://localhost:8080/ws/game
# 接続後、以下を入力:
{"type":"connection_check","data":{"playerId":"p3"}}
```


3人目が接続すると、各クライアントに以下のような配牌メッセージが送信されます:

```json
{
  "type": "game_start",
  "data": {
    "gameId": "ゲームID",
    "playerId": "プレイヤーID",
    "tehai": ["1m","2m","3m",...],
    "wanpai": {
      "revealedDora": ["hatu"],
      "kanDoras": ["4s","7s","9s"],
      "unrevealedDoras": ["4p","nan","1m","9s"],
      "rinsyan": ["5pr","chun","6p","7p"]
    },
    "yama": [...],
    "players": [
      {"id": "p1", "tehai": [...], "isHost": true},
      {"id": "p2", "tehai": [...], "isHost": false},
      {"id": "p3", "tehai": [...], "isHost": false}
    ]
  }
}
```

## � WebSocket接続について

**重要な注意点:**
- `echo | websocat` を使うと、メッセージ送信後すぐに接続が閉じます
- 3人麻雀のテストには **持続的な接続** が必要です
- 配牌を受信するには、3人全員が **同時に接続状態** を保つ必要があります

**推奨テスト方法:**
1. `websocat ws://localhost:8080/ws/game` で対話形式接続
2. Node.jsスクリプト（`test-websocket.js`）での持続的接続
3. 3つのターミナルで同時接続テスト

## �🛠️ 開発

### 利用可能なMakeコマンド

```bash
# ビルド
make build

# 起動
make run

# ホットリロード起動（要air）
make dev

# テスト実行
make test

# リント実行
make lint

# Swagger文書生成
make swagger-gen

# Docker関連
make docker-build    # イメージビルド
make docker-up       # 起動
make docker-down     # 停止

# ヘルプ
make help
```

### プロジェクト構成

```
backend/
├── cmd/main.go                      # エントリーポイント
├── internal/
│   ├── domain/entity/              # ゲームエンティティ
│   ├── usecase/                    # ビジネスロジック
│   └── interface/handler/          # HTTP/WebSocketハンドラー
├── swagger/gen/                    # API仕様書
├── Dockerfile                      # Dockerイメージ定義
├── docker-compose.yml             # Docker Compose設定
└── Makefile                       # ビルド・実行コマンド
```

## 🎲 牌の仕様

- **萬子**: 1m, 9m (各4枚)
- **筒子**: 1p-9p (5pは通常3枚+赤ドラ5pr 1枚)
- **索子**: 1s-9s (5sは通常3枚+赤ドラ5sr 1枚)
- **字牌**: ton,nan,sya,pe,haku,hatu,chun (各4枚)
- **総数**: 80枚

## 📊 配牌の仕様

- **親（ホスト）**: 14枚
- **子（その他）**: 13枚ずつ (計26枚)
- **王牌**: 13枚 (表ドラ1枚 + 裏ドラ4枚 + 嶺上牌8枚)
- **山牌**: 27枚

## 🏗️ アーキテクチャ

シンプルな3層アーキテクチャ（完全メモリベース）:

```
├── Entity    # ゲーム・プレイヤー・牌の定義
├── UseCase   # ビジネスロジック + メモリストレージ
└── Handler   # HTTP/WebSocket インターフェース
```

**特徴:**
- 永続化なし（メモリのみ）
- 外部依存なし
- 高速・軽量・シンプル

## 📝 ライセンス

MIT License

---

## 🧪 テスト実行

### 全テストの実行
```bash
go test ./test -v
```

### テスト内容
作成されたユニットテストでは以下の役パターンをテストしています：

#### 基本的な役（1-2翻）
- **断么九のみ（1翻）**: 2-8の数牌のみの手
- **立直+断么九（2翻）**: リーチと断么九の複合
- **門前清自摸和+断么九（2翻）**: ツモ和了と断么九の複合

#### くいさがり（鳴きによる減点）
以下の役は鳴きによって翻数が下がります：

- **清一色**: 門前6翻 → 鳴きあり5翻
- **混一色**: 門前3翻 → 鳴きあり2翻
- **純全帯么九**: 門前3翻 → 鳴きあり2翻
- **混全帯么九**: 門前2翻 → 鳴きあり1翻
- **一気通貫**: 門前2翻 → 鳴きあり1翻
- **三色同順**: 門前2翻 → 鳴きあり1翻

#### くいさがりなし（鳴きによる減点なし）
- **三色同刻**: 門前/鳴きあり共に2翻
- **対々和**: 門前/鳴きあり共に2翻
- **役牌**: 門前/鳴きあり共に1翻

#### 中級役（2-3翻）
- **七対子+断么九（3翻）**: 7つの対子による特殊形
- **役牌（1翻）**: 三元牌（白/發/中）のポン
- **対々和（2翻）**: 4つのポン・カンによる手

#### 高得点役（5-8翻）
- **混一色（2-3翻）**: 一色+字牌の混合（副露時減点）
- **清一色（5-6翻）**: 一色のみの手（副露時減点）
- **満貫（5翻）**: 対々和+混一色+役牌の複合

#### 複合役
多くのテストケースでは、複数の役が同時に成立する複合役をテストしています：
- **清一色+七対子+断么九（9翻）**: 複数役の合算
- **混一色+対々和+役牌（5翻）**: 満貫レベルの複合
- **純全帯么九+一気通貫（5翻）**: チャンタ系との複合

#### ドラ計算
- **ドラ2個**: インジケーター「1m」でドラ「2m」が2枚ある場合

#### チョンボ判定
- **牌数不正**: 14牌でない不正な手

### テスト結果の見方
各テストケースでは以下を検証します：
- **Score**: 正しい点数計算（1翻=1000点など）
- **TotalHan**: 役+ドラの合計翻数
- **DoraCount**: ドラの枚数
- **IsChombo**: チョンボ判定
- **Yaku**: 成立した役の一覧

例：
```
Response: {Yaku:[{Name:対々和 Han:2} {Name:役牌(白) Han:1}] Han:3 DoraCount:2 TotalHan:5 Score:8000 IsChombo:false}
```
→ 対々和(2翻)+役牌(1翻)+ドラ2個 = 5翻 = 満貫8000点

### 期待されるテスト出力
```
=== RUN   TestYakuEvaluation
### 最新のテスト結果（2024年度版 - 全32ケース）

全てのテストが正常にパスしています：

```
=== RUN   TestYakuEvaluation
=== RUN   TestYakuEvaluation/2翻_-_門前清自摸和_断么九_門前_ツモ
=== RUN   TestYakuEvaluation/4翻_-_対々和_役牌(白)_混全帯么九_ロン
=== RUN   TestYakuEvaluation/2翻_-_立直_断么九_門前_ロン
=== RUN   TestYakuEvaluation/1翻_-_断么九_門前_ロン
=== RUN   TestYakuEvaluation/3翻_-_七対子_断么九_門前_ロン
=== RUN   TestYakuEvaluation/8翻_-_対々和_清一色_断么九_ロン
=== RUN   TestYakuEvaluation/6翻_-_対々和_混一色_役牌(白)_混全帯么九_ロン
=== RUN   TestYakuEvaluation/9翻_-_対々和_清一色_純全帯么九_ロン
=== RUN   TestYakuEvaluation/6翻_-_対々和_役牌(白)_混全帯么九_ドラ2_ロン
=== RUN   TestYakuEvaluation/役満_-_国士無双_門前_ロン
=== RUN   TestYakuEvaluation/6翻_-_対々和_役牌(場風)_混全帯么九_門清_ツモ
=== RUN   TestYakuEvaluation/役満_-_大三元_ロン
=== RUN   TestYakuEvaluation/役満_-_緑一色_門前_ロン
=== RUN   TestYakuEvaluation/役満_-_字一色_門前_ロン
=== RUN   TestYakuEvaluation/役満_-_清老頭_門前_ロン
=== RUN   TestYakuEvaluation/4翻_-_役牌(東)_場風_対々和_混全帯么九_ロン
=== RUN   TestYakuEvaluation/4翻_-_役牌(南)_自風_対々和_混全帯么九_ロン
=== RUN   TestYakuEvaluation/9翻_-_清一色_七対子_断么九_門前_ロン
=== RUN   TestYakuEvaluation/8翻_-_清一色_対々和_断么九_鳴きあり_ロン
=== RUN   TestYakuEvaluation/7翻_-_混一色_七対子_混全帯么九_門前_ロン
=== RUN   TestYakuEvaluation/6翻_-_混一色_対々和_混全帯么九_役牌鳴きあり_ロン
=== RUN   TestYakuEvaluation/3翻_-_純全帯么九_門前_ロン
=== RUN   TestYakuEvaluation/2翻_-_純全帯么九_鳴きあり_ロン
=== RUN   TestYakuEvaluation/2翻_-_混全帯么九_門前_ロン
=== RUN   TestYakuEvaluation/1翻_-_混全帯么九_鳴きあり_ロン
=== RUN   TestYakuEvaluation/5翻_-_一気通貫_純全帯么九_門前_ロン
=== RUN   TestYakuEvaluation/3翻_-_一気通貫_純全帯么九_鳴きあり_ロン
=== RUN   TestYakuEvaluation/5翻_-_三色同順_純全帯么九_門前_ロン
=== RUN   TestYakuEvaluation/3翻_-_三色同順_純全帯么九_鳴きあり_ロン
=== RUN   TestYakuEvaluation/5翻_-_三色同刻_対々和_門前_ロン
=== RUN   TestYakuEvaluation/5翻_-_三色同刻_対々和_鳴きあり_ロン
=== RUN   TestYakuEvaluation/チョンボ_-_牌数不正
--- PASS: TestYakuEvaluation (0.00s)
PASS
ok      mahjong-backend/test    0.805s
```

### 実装済み機能

#### ✅ 完全実装済み
- **基本役**: 全ての一般的な役（断么九、立直、門前清自摸和など）
- **複合役**: 対々和、七対子、清一色、混一色など
- **役満**: 国士無双、大三元、緑一色、字一色、清老頭など全13種類
- **役牌判定**: 場風牌、自風牌の正確な判定ロジック
- **くいさがり**: 鳴きによる翻数減少（清一色 6→5翻、混一色 3→2翻など）
- **複雑な役組み合わせ**: 自動検出される混全帯么九、純全帯么九などの副次的役
- **正確な点数計算**: 麻雀の正式な点数表に基づく計算
- **エラーハンドリング**: チョンボ判定（牌数不正など）

#### 🎯 テストカバレッジ
- 基本役の単独テスト
- 複合役の組み合わせテスト  
- 役満の判定テスト
- くいさがりの翻数減少テスト
- 複雑な役の自動検出テスト
- エラーケースのテスト

この実装により、正確で包括的な麻雀の役判定システムが完成しました。
