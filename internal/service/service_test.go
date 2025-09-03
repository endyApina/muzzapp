package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/endyapina/muzzapp/internal/redis"
	redis_mocks "github.com/endyapina/muzzapp/internal/redis/mocks"
	db_mocks "github.com/endyapina/muzzapp/internal/repository/mocks"
)

func TestExploreService_PutDecision(t *testing.T) {
	tests := []struct {
		name          string
		actorID       string
		recipientID   string
		liked         bool
		mockUpsertErr error
		mockMutual    bool
		mockMutualErr error
		wantMutual    bool
		wantErr       bool
	}{
		{
			name:        "success - mutual like",
			actorID:     "user1",
			recipientID: "user2",
			liked:       true,
			mockMutual:  true,
			wantMutual:  true,
			wantErr:     false,
		},
		{
			name:          "failure - repo upsert error",
			actorID:       "user1",
			recipientID:   "user2",
			liked:         true,
			mockUpsertErr: errors.New("db error"),
			wantErr:       true,
		},
		{
			name:          "failure - mutual check error",
			actorID:       "user1",
			recipientID:   "user2",
			liked:         false,
			mockMutualErr: errors.New("db error"),
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			mockRepo := db_mocks.NewRepository(t)
			mockCache := redis_mocks.NewRepository(t)

			mockRepo.EXPECT().
				UpsertDecision(ctx, tt.actorID, tt.recipientID, tt.liked).
				Return(tt.mockUpsertErr)

			if tt.mockUpsertErr == nil {
				mockRepo.EXPECT().
					CheckMutualLike(ctx, tt.actorID, tt.recipientID).
					Return(tt.mockMutual, tt.mockMutualErr)
			}

			if tt.mockUpsertErr == nil {
				if tt.liked {
					mockCache.EXPECT().
						AddLike(ctx, tt.recipientID, tt.actorID, mock.AnythingOfType("int64")).
						Return(nil)
				} else {
					mockCache.EXPECT().
						RemoveLike(ctx, tt.recipientID, tt.actorID).
						Return(nil)
				}
			}

			svc := New(mockRepo, mockCache)

			gotMutual, err := svc.PutDecision(ctx, tt.actorID, tt.recipientID, tt.liked)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantMutual, gotMutual)
			}
		})
	}
}

func TestExploreService_ListLikedYou(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name            string
		recipientID     string
		paginationToken string
		mockCacheData   []redis.Z
		mockCacheNext   string
		mockCacheErr    error
		wantLikers      []string
		wantNextToken   string
		wantErr         bool
	}{
		{
			name:            "success",
			recipientID:     "user2",
			paginationToken: "",
			mockCacheData: []redis.Z{
				{Member: "user1", Score: 1000},
				{Member: "user3", Score: 1100},
			},
			mockCacheNext: "token_from_cache",
			wantLikers:    []string{"user1", "user3"},
			wantNextToken: "token_from_cache",
			wantErr:       false,
		},
		{
			name:            "cache error",
			recipientID:     "user2",
			paginationToken: "",
			mockCacheErr:    errors.New("redis error"),
			wantErr:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := db_mocks.NewRepository(t)
			mockCache := redis_mocks.NewRepository(t)

			mockCache.EXPECT().
				GetLikers(ctx, tt.recipientID, tt.paginationToken).
				Return(tt.mockCacheData, tt.mockCacheNext, tt.mockCacheErr).
				Once()

			svc := New(mockRepo, mockCache)
			got, nextToken, err := svc.ListLikedYou(ctx, tt.recipientID, tt.paginationToken)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)

			gotIDs := make([]string, len(got))
			for i, l := range got {
				gotIDs[i] = l.ActorId
			}
			assert.Equal(t, tt.wantLikers, gotIDs)

			assert.Equal(t, tt.wantNextToken, nextToken)
		})
	}
}

func TestExploreService_ListNewLikedYou(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name             string
		recipientID      string
		paginationToken  string
		mockCacheData    []redis.Z
		mockCacheErr     error
		mockCacheNext    string
		mockHasLikedBack map[string]bool
		mockHasLikedErr  error
		wantLikers       []string
		wantNextToken    string
		wantErr          bool
	}{
		{
			name:            "success - filter liked back",
			recipientID:     "user2",
			paginationToken: "",
			mockCacheData: []redis.Z{
				{Member: "user1", Score: 1000},
				{Member: "user3", Score: 1100},
			},
			mockCacheNext: "token_from_cache",
			mockHasLikedBack: map[string]bool{
				"user1": false,
				"user3": true,
			},
			wantLikers:    []string{"user1"},
			wantNextToken: "token_from_cache",
			wantErr:       false,
		},
		{
			name:            "cache error",
			recipientID:     "user2",
			paginationToken: "",
			mockCacheErr:    errors.New("redis error"),
			wantErr:         true,
		},
		{
			name:            "HasRecipientLikedActor error",
			recipientID:     "user2",
			paginationToken: "",
			mockCacheData:   []redis.Z{{Member: "user1", Score: 1}},
			mockCacheNext:   "token_from_cache",
			mockHasLikedErr: errors.New("db error"),
			wantErr:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := db_mocks.NewRepository(t)
			mockCache := redis_mocks.NewRepository(t)

			mockCache.EXPECT().
				GetLikers(ctx, tt.recipientID, tt.paginationToken).
				Return(tt.mockCacheData, tt.mockCacheNext, tt.mockCacheErr).
				Once()

			for _, z := range tt.mockCacheData {
				member := z.Member.(string)
				mockRepo.EXPECT().
					HasRecipientLikedActor(ctx, tt.recipientID, member).
					Return(tt.mockHasLikedBack[member], tt.mockHasLikedErr).
					Maybe()
			}

			svc := New(mockRepo, mockCache)
			got, nextToken, err := svc.ListNewLikedYou(ctx, tt.recipientID, tt.paginationToken)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)

			gotIDs := make([]string, len(got))
			for i, l := range got {
				gotIDs[i] = l.ActorId
			}
			assert.Equal(t, tt.wantLikers, gotIDs)

			assert.Equal(t, tt.wantNextToken, nextToken)
		})
	}
}
