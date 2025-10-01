package entity

// WebSocketメッセージタイプ
type MessageType string

const (
	// クライアント→サーバー
	MessageTypeConnectionCheck MessageType = "connection_check"

	// サーバー→クライアント
	MessageTypeConnectionResponse MessageType = "connection_response"
	MessageTypeGameStart          MessageType = "game_start"
)

// WebSocketメッセージ構造体
type WebSocketMessage struct {
	Type MessageType `json:"type"`
	Data interface{} `json:"data,omitempty"`
}

// クライアントからの初期接続確認メッセージ
type ConnectionCheckMessage struct {
	PlayerID string `json:"playerId"`
}

// サーバーからの接続応答メッセージ
type ConnectionResponseMessage struct {
	PlayerID     string `json:"playerId"`
	PlayersCount int    `json:"playersCount"`
	MaxPlayers   int    `json:"maxPlayers"`
	Status       string `json:"status"`
	Message      string `json:"message"`
}

// ゲーム開始メッセージ（プレイヤーの手牌含む）
type GameStartMessage struct {
	PlayerID string    `json:"playerId"`
	Tehai    []Tile    `json:"tehai"`
	Wanpai   *Wanpai   `json:"wanpai"`
	Yama     []Tile    `json:"yama"`
	Players  []*Player `json:"players"`
}
