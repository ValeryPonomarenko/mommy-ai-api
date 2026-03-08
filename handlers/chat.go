package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"mommy-ai-api/auth"
	"mommy-ai-api/chat"
	"mommy-ai-api/llm"
)

// ChatRequest is the JSON body for POST /api/chat.
type ChatRequest struct {
	Message  string `json:"message"`
	ThreadID string `json:"thread_id,omitempty"`   // optional; default "default" for main chat
	Context  string `json:"context,omitempty"`    // optional; e.g. "Событие: Второй скрининг — 15 Июн" or "Анализ: Общий анализ крови"
	Messages []struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"messages,omitempty"`
}

// ChatResponse is the JSON response.
type ChatResponse struct {
	Reply string `json:"reply"`
}

// Chat returns a handler that uses the given LLM client. If user is set (optional auth), appends to history.
func Chat(client *llm.Client, historyStore *chat.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req ChatRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON: " + err.Error()})
			return
		}

		var messages []llm.ChatMessage
		var apiMessages []chat.Message
		if len(req.Messages) > 0 {
			messages = make([]llm.ChatMessage, len(req.Messages))
			apiMessages = make([]chat.Message, len(req.Messages))
			for i, m := range req.Messages {
				messages[i] = llm.ChatMessage{Role: m.Role, Content: m.Content}
				apiMessages[i] = chat.Message{Role: m.Role, Content: m.Content}
			}
		} else if req.Message != "" {
			messages = []llm.ChatMessage{{Role: "user", Content: req.Message}}
			apiMessages = []chat.Message{{Role: "user", Content: req.Message}}
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "provide 'message' or 'messages'"})
			return
		}

		reply, err := client.Chat(c.Request.Context(), messages, req.Context)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if userID, ok := c.Get(auth.ContextUserIDKey); ok && historyStore != nil {
			if uid, _ := userID.(string); uid != "" {
				threadID := req.ThreadID
				if threadID == "" {
					threadID = chat.DefaultThreadID
				}
				historyStore.Set(uid, threadID, apiMessages, reply)
			}
		}

		c.JSON(http.StatusOK, ChatResponse{Reply: reply})
	}
}

// GetChatHistory returns the in-memory chat history for the current user and thread (auth required).
// Query: thread_id (optional, default "default").
func GetChatHistory(historyStore *chat.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get(auth.ContextUserIDKey)
		uid, _ := userID.(string)
		if uid == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "требуется авторизация"})
			return
		}
		threadID := c.Query("thread_id")
		if threadID == "" {
			threadID = chat.DefaultThreadID
		}
		list := historyStore.Get(uid, threadID)
		if list == nil {
			list = []chat.Message{}
		}
		c.JSON(http.StatusOK, gin.H{"messages": list})
	}
}
