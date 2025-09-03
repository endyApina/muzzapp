package redis

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/endyapina/muzzapp/internal/config"

	"github.com/redis/go-redis/v9"
)

type Cache struct {
	client *redis.Client
	config *config.AppConfig
}

// NewCache creates and returns a new redis client wrapped inside our Cache struct.
//
// why redis here?
// - redis is an in-memory data store that excels at fast reads/writes.
// - perfect for caching, counters, leaderboards, and sorted sets (like in this app).
//
// for production at scale, you might:
// - use redis clusters or redis sentinel for high availability.
// - configure connection pooling, retries, and timeouts.
// - add monitoring and metrics around redis usage to detect bottlenecks.
func NewCache(config *config.AppConfig) (*Cache, error) {
	if config == nil {
		return nil, errors.New("missing redis config")
	}
	return &Cache{
		client: redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%s", config.RedisHost, config.RedisPort),
			Password: config.RedisPassword,
			DB:       config.RedisDB,
		}),
		config: config,
	}, nil
}

// Add a like to sorted set
func (c *Cache) AddLike(ctx context.Context, recipientID, actorID string, timestamp int64) error {
	key := fmt.Sprintf("liked:%s", recipientID)
	return c.client.ZAdd(ctx, key, redis.Z{
		Score:  float64(timestamp),
		Member: actorID,
	}).Err()
}

// Remove a like from sorted set (used for updates/passes)
func (c *Cache) RemoveLike(ctx context.Context, recipientID, actorID string) error {
	key := fmt.Sprintf("liked:%s", recipientID)
	return c.client.ZRem(ctx, key, actorID).Err()
}

// generateNextToken creates an opaque pagination token
func generateNextToken(last redis.Z) string {
	tokenStr := fmt.Sprintf("%f:%s", last.Score, last.Member.(string))
	return base64.URLEncoding.EncodeToString([]byte(tokenStr))
}

// parseNextToken decodes a pagination token
func parseNextToken(token string) (float64, string, error) {
	if token == "" {
		return 0, "", nil
	}
	data, err := base64.URLEncoding.DecodeString(token)
	if err != nil {
		return 0, "", err
	}
	var score float64
	var member string
	_, err = fmt.Sscanf(string(data), "%f:%s", &score, &member)
	if err != nil {
		return 0, "", err
	}
	return score, member, nil
}

// GetLikes fetches likes from Redis with keyset pagination
func (c *Cache) GetLikers(ctx context.Context, recipientID string, paginationToken string) ([]Z, string, error) {
	key := fmt.Sprintf("liked:%s", recipientID)

	startScore, _, err := parseNextToken(paginationToken)
	if err != nil {
		return nil, "", err
	}

	zs, err := c.client.ZRangeByScoreWithScores(ctx, key, &redis.ZRangeBy{
		Min:    fmt.Sprintf("(%f", startScore),
		Max:    "+inf",
		Offset: 0,
		Count:  c.config.PaginationSize,
	}).Result()
	if err != nil {
		return nil, "", err
	}

	var nextToken string
	if len(zs) == int(c.config.PaginationSize) {
		nextToken = generateNextToken(zs[len(zs)-1])
	}

	return zs, nextToken, nil
}

func (c *Cache) CountLikes(ctx context.Context, recipientID string) (int64, error) {
	key := fmt.Sprintf("liked:%s", recipientID)
	return c.client.ZCard(ctx, key).Result()
}
