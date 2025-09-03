package service

import (
	"context"
	"time"

	redis_cache "github.com/endyapina/muzzapp/internal/redis"
	"github.com/endyapina/muzzapp/internal/repository"
	pb "github.com/endyapina/muzzapp/proto/gen/github.com/muzzapp/backend-interview-task"
	"github.com/redis/go-redis/v9"
)

type ExploreService struct {
	repo  repository.Repository
	cache redis_cache.Repository
}

func New(repo repository.Repository, cache redis_cache.Repository) *ExploreService {
	return &ExploreService{
		repo:  repo,
		cache: cache,
	}
}

// PutDecision: business logic with caching and mutual likes
func (s *ExploreService) PutDecision(ctx context.Context, actorID, recipientID string, liked bool) (bool, error) {
	if err := s.repo.UpsertDecision(ctx, actorID, recipientID, liked); err != nil {
		return false, err
	}

	if liked {
		s.cache.AddLike(ctx, recipientID, actorID, time.Now().Unix())
	} else {
		s.cache.RemoveLike(ctx, recipientID, actorID)
	}

	mutual, err := s.repo.CheckMutualLike(ctx, actorID, recipientID)
	return mutual, err
}

func (s *ExploreService) CountLikedYou(ctx context.Context, recipientID string) (uint64, error) {
	count, err := s.cache.CountLikes(ctx, recipientID)
	return uint64(count), err
}

func (s *ExploreService) ListLikedYou(ctx context.Context, recipientID string, paginationToken string) ([]*pb.ListLikedYouResponse_Liker, string, error) {
	entries, nextToken, err := s.cache.GetLikers(ctx, recipientID, paginationToken)
	if err != nil && err != redis.Nil {
		return nil, "", err
	}

	var likers []*pb.ListLikedYouResponse_Liker
	for _, e := range entries {
		likers = append(likers, &pb.ListLikedYouResponse_Liker{
			ActorId:       e.Member.(string),
			UnixTimestamp: uint64(int64(e.Score)),
		})
	}
	return likers, nextToken, nil
}

func (s *ExploreService) ListNewLikedYou(ctx context.Context, recipientID string, paginationToken string) ([]*pb.ListLikedYouResponse_Liker, string, error) {
	entries, nextToken, err := s.cache.GetLikers(ctx, recipientID, paginationToken)
	if err != nil && err != redis.Nil {
		return nil, "", err
	}

	var likers []*pb.ListLikedYouResponse_Liker
	for _, e := range entries {
		actorID := e.Member.(string)
		likedBack, err := s.repo.HasRecipientLikedActor(ctx, recipientID, actorID)
		if err != nil {
			return nil, "", err
		}
		if likedBack {
			continue
		}
		likers = append(likers, &pb.ListLikedYouResponse_Liker{
			ActorId:       actorID,
			UnixTimestamp: uint64(int64(e.Score)),
		})
	}

	return likers, nextToken, nil
}
