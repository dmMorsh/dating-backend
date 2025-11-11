package realtime

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// SessionStore is an abstraction for one-time WebSocket session tokens.
// Implementations must honor TTL semantics: a token set with a TTL should
// become unavailable after the TTL elapses.
type SessionStore interface {
    Set(token string, userID int64, ttl time.Duration) error
    Get(token string) (int64, bool, error)
    Delete(token string) error
}

// In-memory implementation -------------------------------------------------
type inMemoryEntry struct {
    userID    int64
    expiresAt time.Time
}

type InMemorySessionStore struct {
    mu sync.RWMutex
    m  map[string]inMemoryEntry
}

func NewInMemorySessionStore() *InMemorySessionStore {
    s := &InMemorySessionStore{m: make(map[string]inMemoryEntry)}
    go s.cleaner()
    return s
}

func (s *InMemorySessionStore) Set(token string, userID int64, ttl time.Duration) error {
    s.mu.Lock()
    s.m[token] = inMemoryEntry{userID: userID, expiresAt: time.Now().Add(ttl)}
    s.mu.Unlock()
    return nil
}

func (s *InMemorySessionStore) Get(token string) (int64, bool, error) {
    s.mu.RLock()
    e, ok := s.m[token]
    s.mu.RUnlock()
    if !ok {
        return 0, false, nil
    }
    if time.Now().After(e.expiresAt) {
        // expired — eagerly delete
        s.mu.Lock()
        delete(s.m, token)
        s.mu.Unlock()
        return 0, false, nil
    }
    return e.userID, true, nil
}

func (s *InMemorySessionStore) Delete(token string) error {
    s.mu.Lock()
    delete(s.m, token)
    s.mu.Unlock()
    return nil
}

func (s *InMemorySessionStore) cleaner() {
    ticker := time.NewTicker(1 * time.Minute)
    defer ticker.Stop()
    for range ticker.C {
        now := time.Now()
        s.mu.Lock()
        for k, v := range s.m {
            if now.After(v.expiresAt) {
                delete(s.m, k)
            }
        }
        s.mu.Unlock()
    }
}

// Redis-backed implementation ----------------------------------------------
type RedisSessionStore struct {
    client *redis.Client
}

func NewRedisSessionStore(opts *redis.Options) *RedisSessionStore {
    c := redis.NewClient(opts)
    return &RedisSessionStore{client: c}
}

func (r *RedisSessionStore) Set(token string, userID int64, ttl time.Duration) error {
    ctx := context.Background()
    return r.client.Set(ctx, token, strconv.FormatInt(userID, 10), ttl).Err()
}

func (r *RedisSessionStore) Get(token string) (int64, bool, error) {
    ctx := context.Background()
    s, err := r.client.Get(ctx, token).Result()
    if err == redis.Nil {
        return 0, false, nil
    }
    if err != nil {
        return 0, false, err
    }
    id, err := strconv.ParseInt(s, 10, 64)
    if err != nil {
        return 0, false, err
    }
    return id, true, nil
}

func (r *RedisSessionStore) Delete(token string) error {
    ctx := context.Background()
    return r.client.Del(ctx, token).Err()
}

// DefaultSessionStore is the store used by handlers. By default it's an
// in-memory store. Main or tests can replace it with a Redis-backed store
// by assigning a different implementation to this variable before the server
// starts handling requests.
var DefaultSessionStore SessionStore = NewInMemorySessionStore()
