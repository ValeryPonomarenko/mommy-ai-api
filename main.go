package main

import (
	"log"
	"net/http"
	"os"

	"mommy-ai-api/auth"
	"mommy-ai-api/chat"
	"mommy-ai-api/handlers"
	"mommy-ai-api/llm"

	"github.com/gin-gonic/gin"
)

func main() {
	apiKey := os.Getenv("YANDEX_API_KEY") //Insert your API key here

	cfg := llm.Config{
		APIKey:   apiKey,
		BaseURL:  "https://ai.api.cloud.yandex.net/v1",
		Project:  "b1gectqd6v1vksue8oi4",
		PromptID: "fvtch4g1pmhosuh3r2cu",
	}
	client, err := llm.NewClient(cfg)
	if err != nil {
		log.Fatalf("llm client: %v", err)
	}
	store := auth.NewStore()
	chatHistory := chat.NewStore()

	r := gin.Default()
	r.Use(corsMiddleware())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	r.GET("/api/hello", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Hello from Mommy AI API"})
	})
	r.POST("/api/register", handlers.Register(store))
	r.POST("/api/login", handlers.Login(store))
	me := r.Group("/api/me")
	me.Use(auth.RequireAuth(store))
	{
		me.GET("", handlers.GetMe(store))
		me.PUT("/onboarding", handlers.PutOnboarding(store))
	}
	r.POST("/api/chat", auth.OptionalAuth(store), handlers.Chat(client, chatHistory))
	r.GET("/api/chat/history", auth.RequireAuth(store), handlers.GetChatHistory(chatHistory))

	r.Run(":8080")
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
