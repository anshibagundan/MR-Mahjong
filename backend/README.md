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

### 2. Swagger UI

ブラウザで以下にアクセス:
```
http://localhost:8080/swagger/index.html
```

### 3. WebSocket接続テスト

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

### 4. 3人接続での配牌テスト

#### websocat使用（推奨）

3つのターミナルを開いて、それぞれで持続的な接続を行います：

**ターミナル1:**
```bash
websocat ws://localhost:8080/ws/game
# 接続後、以下を入力:
{"type":"connection_check","data":{"playerId":"player-1"}}
```

**ターミナル2:**
```bash
websocat ws://localhost:8080/ws/game
# 接続後、以下を入力:
{"type":"connection_check","data":{"playerId":"player-2"}}
```

**ターミナル3:**
```bash
websocat ws://localhost:8080/ws/game
# 接続後、以下を入力:
{"type":"connection_check","data":{"playerId":"player-3"}}
```


3人目が接続すると、各クライアントに以下のような配牌メッセージが送信されます:

```json
{
  "type": "game_start",
  "data": {
    "gameId": "ゲームID",
    "playerId": "プレイヤーID",
    "tehai": ["1m","2m","3m",...],
    "wanpai": {...},
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
