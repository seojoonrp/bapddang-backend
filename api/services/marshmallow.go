// api/services/marshmallow.go

package services

import (
	"context"

	"github.com/seojoonrp/bapddang-server/api/repositories"
	"github.com/seojoonrp/bapddang-server/apperr"
	"github.com/seojoonrp/bapddang-server/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MarshmallowService interface {
	GetUserMarshmallows(ctx context.Context, userID string) ([]models.Marshmallow, error)
}

type marshmallowService struct {
	marshmallowRepo repositories.MarshmallowRepository
}

func NewMarshmallowService(mr repositories.MarshmallowRepository) MarshmallowService {
	return &marshmallowService{
		marshmallowRepo: mr,
	}
}

func (s *marshmallowService) GetUserMarshmallows(ctx context.Context, userID string) ([]models.Marshmallow, error) {
	uID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, apperr.InternalServerError("invalid user ID in token", err)
	}

	marshmallows, err := s.marshmallowRepo.FindByUserID(ctx, uID)
	if err != nil {
		return nil, apperr.InternalServerError("failed to fetch marshmallows", err)
	}
	if marshmallows == nil {
		marshmallows = []models.Marshmallow{}
	}

	return marshmallows, nil
}
