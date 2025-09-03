package redis

import (
	"context"

	goredis "github.com/redis/go-redis/v9"
)

// Z alias to goredis.Z for cleaner service usage
type Z = goredis.Z

// Repository is an interface that defines the operations we need from Redis.
// This allows us to mock the cache implementation when running unit tests.
type Repository interface {
	AddLike(ctx context.Context, recipientID, actorID string, timestamp int64) error
	RemoveLike(ctx context.Context, recipientID, actorID string) error
	GetLikers(ctx context.Context, recipientID string, paginationToken string) ([]Z, string, error)
	CountLikes(ctx context.Context, recipientID string) (int64, error)
}
