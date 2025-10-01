package handler

import (
	"log"
	"mahjong-backend/internal/usecase"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// WebSocket接続処理
type WebSocketHandler struct {
	wsUsecase *usecase.WebSocketUsecase
	upgrader  websocket.Upgrader
}

// WebSocketHandlerのインスタンスを作成
func NewWebSocketHandler(wsUsecase *usecase.WebSocketUsecase) *WebSocketHandler {
	return &WebSocketHandler{
		wsUsecase: wsUsecase,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// 開発環境では全オリジンを許可
				// 本番環境では適切なオリジンチェックを実装
				return true
			},
		},
	}
}

// WebSocket接続を処理（GET /ws/game）
func (h *WebSocketHandler) HandleWebSocket(c *gin.Context) {
	// HTTP接続をWebSocketにアップグレード
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection to WebSocket: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to upgrade to WebSocket"})
		return
	}

	// WebSocket接続を処理
	go h.wsUsecase.HandleConnection(conn)
}
