package repository

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/endyapina/muzzapp/internal/config"
	"github.com/endyapina/muzzapp/internal/models"

	pb "github.com/endyapina/muzzapp/proto/gen/github.com/muzzapp/backend-interview-task"

	"gorm.io/gorm"
)

type DBRepository struct {
	db     *gorm.DB
	config *config.AppConfig
}

type Liker = pb.ListLikedYouResponse_Liker

func New(db *gorm.DB, config *config.AppConfig) (*DBRepository, error) {
	if config == nil {
		return nil, errors.New("database config is required")
	}
	return &DBRepository{db: db, config: config}, nil
}

func (r *DBRepository) UpsertDecision(ctx context.Context, actorID, recipientID string, liked bool) error {
	return r.db.WithContext(ctx).Save(&models.Decision{
		ActorUserID:     actorID,
		RecipientUserID: recipientID,
		Liked:           liked,
		UnixTimestamp:   time.Now().Unix(),
	}).Error
}

func (r *DBRepository) CheckMutualLike(ctx context.Context, actorID, recipientID string) (bool, error) {
	var count int64

	// count both decisions: actor liked recipient AND recipient liked actor
	err := r.db.WithContext(ctx).Model(&models.Decision{}).
		Where("(actor_user_id = ? AND recipient_user_id = ? AND liked = ?) OR (actor_user_id = ? AND recipient_user_id = ? AND liked = ?)",
			actorID, recipientID, true,
			recipientID, actorID, true,
		).Count(&count).Error
	if err != nil {
		return false, err
	}

	// if count == 2, then it is mutual
	return count == 2, nil
}

// GetLikers returns likers of a recipient with optional pagination
func (r *DBRepository) GetLikers(ctx context.Context, recipientID string, paginationToken string) ([]Liker, string, error) {
	pageSize := int(r.config.PaginationSize)
	var likers []Liker
	query := r.db.WithContext(ctx).Where("recipient_user_id = ? AND liked = ?", recipientID, true).Order("unix_timestamp ASC, actor_user_id ASC").Limit(pageSize + 1)

	if paginationToken != "" {
		ts, actor, err := decodePaginationToken(paginationToken)
		if err != nil {
			return nil, "", err
		}
		query = query.Where("(unix_timestamp > ?) OR (unix_timestamp = ? AND actor_user_id > ?)", ts, ts, actor)
	}

	var results []models.Decision
	if err := query.Find(&results).Error; err != nil {
		return nil, "", err
	}

	nextToken := ""
	if len(results) > pageSize {
		nextToken = encodePaginationToken(results[pageSize].UnixTimestamp, results[pageSize].ActorUserID)
		results = results[:pageSize]
	}

	for _, d := range results {
		likers = append(likers, Liker{
			ActorId:       d.ActorUserID,
			UnixTimestamp: uint64(d.UnixTimestamp),
		})
	}

	return likers, nextToken, nil
}

// CountLikes returns number of likes a recipient has
func (r *DBRepository) CountLikes(ctx context.Context, recipientID string) (uint64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.Decision{}).Where("recipient_user_id = ? AND liked = ?", recipientID, true).Count(&count).Error; err != nil {
		return 0, err
	}
	return uint64(count), nil
}

// GetNewLikers excludes users who the recipient has already liked
func (r *DBRepository) GetNewLikers(ctx context.Context, recipientID string, paginationToken string) ([]Liker, string, error) {
	pageSize := int(r.config.PaginationSize)
	var likers []Liker
	query := r.db.WithContext(ctx).Table("decisions as d1").
		Select("d1.actor_user_id, d1.unix_timestamp").
		Joins("LEFT JOIN decisions as d2 ON d1.actor_user_id = d2.recipient_user_id AND d2.actor_user_id = ?", recipientID).
		Where("d1.recipient_user_id = ? AND d1.liked = ? AND (d2.liked IS NULL OR d2.liked = ?)", recipientID, true, false).
		Order("d1.unix_timestamp ASC, d1.actor_user_id ASC").
		Limit(pageSize + 1)

	if paginationToken != "" {
		ts, actor, err := decodePaginationToken(paginationToken)
		if err != nil {
			return nil, "", err
		}
		query = query.Where("(d1.unix_timestamp > ?) OR (d1.unix_timestamp = ? AND d1.actor_user_id > ?)", ts, ts, actor)
	}

	var results []models.Decision
	if err := query.Scan(&results).Error; err != nil {
		return nil, "", err
	}

	nextToken := ""
	if len(results) > pageSize {
		nextToken = encodePaginationToken(results[pageSize].UnixTimestamp, results[pageSize].ActorUserID)
		results = results[:pageSize]
	}

	for _, d := range results {
		likers = append(likers, Liker{
			ActorId:       d.ActorUserID,
			UnixTimestamp: uint64(d.UnixTimestamp),
		})
	}

	return likers, nextToken, nil
}

// HasRecipientLikedActor checks if recipient has liked the actor
func (r *DBRepository) HasRecipientLikedActor(ctx context.Context, recipientID, actorID string) (bool, error) {
	var decision models.Decision
	err := r.db.WithContext(ctx).First(&decision, "actor_user_id = ? AND recipient_user_id = ? AND liked = ?", recipientID, actorID, true).Error
	if err == gorm.ErrRecordNotFound {
		return false, nil
	}
	return err == nil, err
}

// Helper functions for encoding/decoding pagination tokens
func encodePaginationToken(ts int64, actorID string) string {
	return base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%d|%s", ts, actorID)))
}

func decodePaginationToken(token string) (int64, string, error) {
	bytes, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return 0, "", err
	}
	parts := string(bytes)
	var ts int64
	var actor string
	_, err = fmt.Sscanf(parts, "%d|%s", &ts, &actor)
	return ts, actor, err
}
