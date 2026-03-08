package main

import (
	"log"
	"net/http"
	"os"

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

	r := gin.Default()
	r.Use(corsMiddleware())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	r.GET("/api/hello", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Hello from Mommy AI API"})
	})
	r.POST("/api/chat", handlers.Chat(client))

	r.Run(":8080")
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
