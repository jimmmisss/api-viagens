package utils

import (
	"sync"
	"time"
)

// TokenBlacklist manages a list of invalidated tokens
type TokenBlacklist struct {
	blacklist map[string]time.Time
	mutex     sync.RWMutex
}

// NewTokenBlacklist creates a new token blacklist
func NewTokenBlacklist() *TokenBlacklist {
	return &TokenBlacklist{
		blacklist: make(map[string]time.Time),
	}
}

// Add adds a token to the blacklist with an expiration time
func (tb *TokenBlacklist) Add(token string, expiry time.Time) {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()
	tb.blacklist[token] = expiry
}

// IsBlacklisted checks if a token is in the blacklist
func (tb *TokenBlacklist) IsBlacklisted(token string) bool {
	tb.mutex.RLock()
	defer tb.mutex.RUnlock()
	_, exists := tb.blacklist[token]
	return exists
}

// Cleanup removes expired tokens from the blacklist
func (tb *TokenBlacklist) Cleanup() {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()
	now := time.Now()
	for token, expiry := range tb.blacklist {
		if now.After(expiry) {
			delete(tb.blacklist, token)
		}
	}
}

// Global token blacklist instance
var globalTokenBlacklist = NewTokenBlacklist()

// GetTokenBlacklist returns the global token blacklist instance
func GetTokenBlacklist() *TokenBlacklist {
	return globalTokenBlacklist
}
