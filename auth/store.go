package auth

import (
	"sync"
)

// User is stored in memory (prototype).
type User struct {
	ID           string
	Email        string
	PasswordHash string
}

// Profile holds onboarding data.
type Profile struct {
	PregnancyWeek int `json:"pregnancy_week"` // 1-42
	Feelings      int `json:"feelings"`       // 1-5
}

// Store is an in-memory user and profile store.
type Store struct {
	mu       sync.RWMutex
	users    map[string]*User    // email -> user
	profiles map[string]*Profile // userID -> profile
	nextID   int
}

// NewStore creates a new in-memory store.
func NewStore() *Store {
	return &Store{
		users:    make(map[string]*User),
		profiles: make(map[string]*Profile),
		nextID:   1,
	}
}

// CreateUser adds a user. Returns error if email exists.
func (s *Store) CreateUser(email, passwordHash string) (*User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.users[email]; exists {
		return nil, ErrEmailExists
	}
	id := nextIDString(s.nextID)
	s.nextID++
	u := &User{ID: id, Email: email, PasswordHash: passwordHash}
	s.users[email] = u
	return u, nil
}

// UserByEmail returns user by email.
func (s *Store) UserByEmail(email string) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.users[email]
	if !ok {
		return nil, ErrNotFound
	}
	return u, nil
}

// UserByID returns user by id.
func (s *Store) UserByID(id string) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, u := range s.users {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, ErrNotFound
}

// GetProfile returns profile for userID.
func (s *Store) GetProfile(userID string) *Profile {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.profiles[userID]
}

// SetProfile saves onboarding profile for userID.
func (s *Store) SetProfile(userID string, p Profile) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.profiles[userID] = &p
}

func nextIDString(n int) string {
	return fmtID(n)
}
