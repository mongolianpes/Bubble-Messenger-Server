package ipblocker

import (
	"sync"
	"time"
)

type IPBlocker struct {
	mu          sync.RWMutex
	strikes     map[string]int
	blocked     map[string]time.Time
	maxStrikes  int
	banDuration time.Duration
}

type Blocker interface {
	RegisterStrike(ip string)
	IsBlocked(ip string) bool
}

func New(maxStrikes int, banDuration time.Duration) *IPBlocker {
	return &IPBlocker{
		strikes:     make(map[string]int),
		blocked:     make(map[string]time.Time),
		maxStrikes:  maxStrikes,
		banDuration: banDuration,
	}
}

func (b *IPBlocker) RegisterStrike(ip string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if until, ok := b.blocked[ip]; ok && time.Now().Before(until) {
		return
	}

	b.strikes[ip]++
	if b.strikes[ip] >= b.maxStrikes {
		b.blocked[ip] = time.Now().Add(b.banDuration)
		b.strikes[ip] = 0
	}
}

func (b *IPBlocker) IsBlocked(ip string) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()

	until, ok := b.blocked[ip]
	if !ok {
		return false
	}
	if time.Now().After(until) {
		delete(b.blocked, ip)
		return false
	}
	return true
}
