package handler

import (
	"context"

	"github.com/endyapina/muzzapp/internal/service"
	pb "github.com/endyapina/muzzapp/proto/gen/muzzapp/proto"
)

type ExploreHandler struct {
	service *service.ExploreService
	pb.UnimplementedExploreServiceServer
}

func New(service *service.ExploreService) *ExploreHandler {
	return &ExploreHandler{service: service}
}

func (h *ExploreHandler) PutDecision(ctx context.Context, req *pb.PutDecisionRequest) (*pb.PutDecisionResponse, error) {
	mutual, err := h.service.PutDecision(ctx, req.ActorUserId, req.RecipientUserId, req.LikedRecipient)
	if err != nil {
		return nil, err
	}
	return &pb.PutDecisionResponse{MutualLikes: mutual}, nil
}

func (h *ExploreHandler) ListLikedYou(ctx context.Context, req *pb.ListLikedYouRequest) (*pb.ListLikedYouResponse, error) {
	var token string
	if req.PaginationToken != nil {
		token = *req.PaginationToken
	}

	likers, nextPaginationToken, err := h.service.ListLikedYou(ctx, req.RecipientUserId, token)
	if err != nil {
		return nil, err
	}

	return &pb.ListLikedYouResponse{
		Likers:              likers,
		NextPaginationToken: &nextPaginationToken,
	}, nil
}

func (h *ExploreHandler) CountLikedYou(ctx context.Context, req *pb.CountLikedYouRequest) (*pb.CountLikedYouResponse, error) {
	count, err := h.service.CountLikedYou(ctx, req.RecipientUserId)
	if err != nil {
		return nil, err
	}
	return &pb.CountLikedYouResponse{Count: count}, nil
}

func (h *ExploreHandler) ListNewLikedYou(ctx context.Context, req *pb.ListLikedYouRequest) (*pb.ListLikedYouResponse, error) {
	var token string
	if req.PaginationToken != nil {
		token = *req.PaginationToken
	}

	likers, nextPaginationToken, err := h.service.ListNewLikedYou(ctx, req.RecipientUserId, token)
	if err != nil {
		return nil, err
	}

	return &pb.ListLikedYouResponse{
		Likers:              likers,
		NextPaginationToken: &nextPaginationToken,
	}, nil
}
