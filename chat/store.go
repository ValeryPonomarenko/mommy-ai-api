package chat

import (
	"sync"
)

// Message is a single chat message (role + content).
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Store holds in-memory chat history per user and thread (userID:threadID).
const DefaultThreadID = "default"

// Store holds in-memory chat history per user and thread.
type Store struct {
	mu      sync.RWMutex
	history map[string][]Message // key = userID + ":" + threadID
}

// NewStore creates a new in-memory chat history store.
func NewStore() *Store {
	return &Store{history: make(map[string][]Message)}
}

func key(userID, threadID string) string {
	if threadID == "" {
		threadID = DefaultThreadID
	}
	return userID + ":" + threadID
}

// Set replaces the thread history with the given messages plus the assistant reply (full thread).
func (s *Store) Set(userID, threadID string, messages []Message, assistantReply string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := make([]Message, 0, len(messages)+1)
	list = append(list, messages...)
	list = append(list, Message{Role: "assistant", Content: assistantReply})
	s.history[key(userID, threadID)] = list
}

// Get returns the full history for a user's thread (copy).
func (s *Store) Get(userID, threadID string) []Message {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := s.history[key(userID, threadID)]
	if len(list) == 0 {
		return nil
	}
	out := make([]Message, len(list))
	copy(out, list)
	return out
}
