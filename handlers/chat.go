package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"mommy-ai-api/llm"
)

// ChatRequest is the JSON body for POST /api/chat.
type ChatRequest struct {
	Message  string `json:"message"`
	Messages []struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"messages,omitempty"`
}

// ChatResponse is the JSON response.
type ChatResponse struct {
	Reply string `json:"reply"`
}

// Chat returns a handler that uses the given LLM client.
func Chat(client *llm.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req ChatRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON: " + err.Error()})
			return
		}

		var messages []llm.ChatMessage
		if len(req.Messages) > 0 {
			messages = make([]llm.ChatMessage, len(req.Messages))
			for i, m := range req.Messages {
				messages[i] = llm.ChatMessage{Role: m.Role, Content: m.Content}
			}
		} else if req.Message != "" {
			messages = []llm.ChatMessage{{Role: "user", Content: req.Message}}
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "provide 'message' or 'messages'"})
			return
		}

		reply, err := client.Chat(c.Request.Context(), messages)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, ChatResponse{Reply: reply})
	}
}
