package main

import (
	"log"
	"net/http"

	"mahjong-backend/internal/interface/handler"
	"mahjong-backend/internal/usecase"

	"github.com/gin-gonic/gin"
)

func main() {
	// usecase
	gameUsecase := usecase.NewGameUsecase()
	wsUsecase := usecase.NewWebSocketUsecase(gameUsecase)

	// handler
	wsHandler := handler.NewWebSocketHandler(wsUsecase)

	// gin
	router := gin.Default()

	// CORS
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// routing
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "mahjong-backend",
			"version": "1.0.0",
		})
	})
	router.GET("/ws/game", wsHandler.HandleWebSocket)

	// server
	port := ":8080"
	log.Printf("Starting server on port %s", port)
	if err := router.Run(port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
