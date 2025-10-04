package usecase

import (
	"encoding/json"
	"log"
	"sync"

	"mahjong-backend/internal/domain/entity"

	"github.com/gorilla/websocket"
)

// WebSocket業務ロジック（メモリ保存）
type WebSocketUsecase struct {
	connections map[string]*websocket.Conn
	mutex       sync.RWMutex
	gameUC      *GameUsecase
}

// WebSocketUsecaseのインスタンスを作成
func NewWebSocketUsecase(gameUC *GameUsecase) *WebSocketUsecase {
	return &WebSocketUsecase{
		connections: make(map[string]*websocket.Conn),
		gameUC:      gameUC,
	}
}

// プレイヤーのWebSocket接続を追加
func (u *WebSocketUsecase) AddConnection(playerID string, conn *websocket.Conn) {
	u.mutex.Lock()
	defer u.mutex.Unlock()
	u.connections[playerID] = conn
}

// プレイヤーのWebSocket接続を削除
func (u *WebSocketUsecase) RemoveConnection(playerID string) {
	u.mutex.Lock()
	defer u.mutex.Unlock()

	if conn, exists := u.connections[playerID]; exists {
		conn.Close()
		delete(u.connections, playerID)
	}
}

// 特定プレイヤーにメッセージ送信
func (u *WebSocketUsecase) SendToPlayer(playerID string, message *entity.WebSocketMessage) error {
	u.mutex.RLock()
	defer u.mutex.RUnlock()

	conn, exists := u.connections[playerID]
	if !exists {
		return nil // 接続が見つからない（プレイヤーが切断済み）
	}

	return u.sendMessage(conn, message)
}

// WebSocket接続でメッセージ送信（ヘルパー）
func (u *WebSocketUsecase) sendMessage(conn *websocket.Conn, message *entity.WebSocketMessage) error {
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	return conn.WriteMessage(websocket.TextMessage, data)
}

// 新しいWebSocket接続を処理
func (u *WebSocketUsecase) HandleConnection(conn *websocket.Conn) {
	var currentPlayerID string
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("WebSocket close error: %v", err)
		}
		if currentPlayerID != "" {
			u.RemoveConnection(currentPlayerID)
			// 接続が失われた時にプレイヤーをゲームから削除
			if u.gameUC.IsPlayerInGame(currentPlayerID) {
				if err := u.gameUC.RemovePlayerFromGame(currentPlayerID); err != nil {
					log.Printf("Failed to remove player from game: %v", err)
				}
			}
		}
	}()

	for {
		// クライアントからメッセージ読み取り
		_, messageBytes, err := conn.ReadMessage()
		if err != nil {
			log.Printf("WebSocket read error: %v", err)
			break
		}

		// メッセージ解析
		var message entity.WebSocketMessage
		if err := json.Unmarshal(messageBytes, &message); err != nil {
			log.Printf("Failed to parse WebSocket message: %v", err)
			continue
		}

		// メッセージタイプに応じて処理
		switch message.Type {
		case entity.MessageTypeConnectionCheck:
			playerID := u.handleConnectionCheck(conn, &message)
			if playerID != "" {
				currentPlayerID = playerID
			}
		default:
			log.Printf("Unknown message type: %s", message.Type)
		}
	}
}

// クライアントからの初期接続確認を処理
func (u *WebSocketUsecase) handleConnectionCheck(conn *websocket.Conn, message *entity.WebSocketMessage) string {
	// 接続確認データを解析
	var connCheckData entity.ConnectionCheckMessage
	dataBytes, err := json.Marshal(message.Data)
	if err != nil {
		log.Printf("Failed to marshal connection check data: %v", err)
		return ""
	}

	if err := json.Unmarshal(dataBytes, &connCheckData); err != nil {
		log.Printf("Failed to parse connection check data: %v", err)
		return ""
	}

	playerID := connCheckData.PlayerID
	if playerID == "" {
		log.Printf("Empty player ID in connection check")
		return ""
	}

	// 接続を追加
	u.AddConnection(playerID, conn)

	// プレイヤーをゲームに追加
	game, err := u.gameUC.AddPlayerToGame(playerID)
	if err != nil {
		log.Printf("Failed to add player %s to game: %v", playerID, err)
		return ""
	}

	// 接続応答を送信
	response := &entity.ConnectionResponseMessage{
		PlayerID:     playerID,
		PlayersCount: len(game.Players),
		MaxPlayers:   game.MaxPlayers,
		Status:       game.Status,
		Message:      "接続確認完了",
	}

	responseMessage := &entity.WebSocketMessage{
		Type: entity.MessageTypeConnectionResponse,
		Data: response,
	}

	if err := u.SendToPlayer(playerID, responseMessage); err != nil {
		log.Printf("Failed to send connection response to player %s: %v", playerID, err)
		return ""
	}

	// ゲーム開始可能かチェック（3人接続時）
	if game.CanStart() {
		log.Printf("Game can start with 3 players")
		if err := u.gameUC.StartGame(u); err != nil {
			log.Printf("Failed to start game: %v", err)
		}
	}

	return playerID
}
