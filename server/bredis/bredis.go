package bredis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Client wraps redis client
type Client struct {
	*redis.Client
	ctx       context.Context
	keyPrefix string
}

// New creates a new Redis client
func New(addr, password string, db int, keyPrefix string) *Client {
	client := &Client{
		Client: redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: password,
			DB:       db,
		}),
		ctx:       context.Background(),
		keyPrefix: keyPrefix,
	}

	// Test connection
	if _, err := client.Ping(client.ctx).Result(); err != nil {
		return nil
	}

	return client
}

func (c *Client) key(k string) string {
	if c.keyPrefix == "" {
		return k
	}
	return fmt.Sprintf("%s:%s", c.keyPrefix, k)
}

// Set stores a value with expiration
func (c *Client) Set(key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.Client.Set(c.ctx, c.key(key), data, ttl).Err()
}

// Get retrieves a value and unmarshals into dest
func (c *Client) Get(key string, dest interface{}) error {
	data, err := c.Client.Get(c.ctx, c.key(key)).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

// Delete removes keys
func (c *Client) Delete(keys ...string) error {
	prefixedKeys := make([]string, len(keys))
	for i, k := range keys {
		prefixedKeys[i] = c.key(k)
	}
	return c.Client.Del(c.ctx, prefixedKeys...).Err()
}

// Incr increments a counter
func (c *Client) Incr(key string) (int64, error) {
	return c.Client.Incr(c.ctx, c.key(key)).Result()
}

// Expire sets TTL on a key
func (c *Client) Expire(key string, ttl time.Duration) error {
	return c.Client.Expire(c.ctx, c.key(key), ttl).Err()
}

// GetTTL returns remaining TTL
func (c *Client) GetTTL(key string) time.Duration {
	ttl, _ := c.Client.TTL(c.ctx, c.key(key)).Result()
	return ttl
}

// ============ Rate Limiting ============

// RateLimitResult holds rate limit check result
type RateLimitResult struct {
	Allowed    bool
	Remaining  int64
	RetryAfter time.Duration
}

// CheckRateLimit checks if identifier is within limit
func (c *Client) CheckRateLimit(identifier string, limit int64, window time.Duration) *RateLimitResult {
	key := "rl:" + identifier
	count, err := c.Incr(key)
	if err != nil {
		return &RateLimitResult{Allowed: true, Remaining: limit}
	}

	if count == 1 {
		_ = c.Expire(key, window)
	}

	if count > limit {
		return &RateLimitResult{
			Allowed:    false,
			Remaining:  0,
			RetryAfter: c.GetTTL(key),
		}
	}

	return &RateLimitResult{
		Allowed:   true,
		Remaining: limit - count,
	}
}

// ResetRateLimit resets rate limit for identifier
func (c *Client) ResetRateLimit(identifier string) {
	_ = c.Delete("rl:" + identifier)
}
