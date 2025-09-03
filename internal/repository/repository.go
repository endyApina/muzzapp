package repository

import (
	"context"
)

// This interface allows us to mock the mysql db repository in unit tests
// without depending on a real database.
type Repository interface {
	UpsertDecision(ctx context.Context, actorID, recipientID string, liked bool) error
	CheckMutualLike(ctx context.Context, actorID, recipientID string) (bool, error)
	GetLikers(ctx context.Context, recipientID string, paginationToken string) ([]Liker, string, error)
	CountLikes(ctx context.Context, recipientID string) (uint64, error)
	GetNewLikers(ctx context.Context, recipientID string, paginationToken string) ([]Liker, string, error)
	HasRecipientLikedActor(ctx context.Context, recipientID, actorID string) (bool, error)
}
