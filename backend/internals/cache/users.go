package cache

import (
	"context"

	"github.com/Alter-Sitanshu/campaignHub/internals/db"
)

// models
type UserResponse struct {
	Id             string     `json:"id" binding:"required"`
	FirstName      string     `json:"first_name" binding:"required"`
	LastName       string     `json:"last_name" binding:"required"`
	Email          string     `json:"email" binding:"required"`
	IsVerified     bool       `json:"is_verified" binding:"required"`
	Gender         string     `json:"gender" binding:"required"`
	Amount         float64    `json:"amount" binding:"required,min=0"`
	Currency       string     `json:"currency"`
	Age            int        `json:"age" binding:"required"`
	ProfilePicture string     `json:"picture"`
	PlatformLinks  []db.Links `json:"links" binding:"required"`
}

// ==================================
// User Balance
// ==================================

func (s *Service) SetUserBalance(ctx context.Context, userID string, balance float64) error {
	key := UserBalanceKey(userID)
	return s.Set(ctx, key, balance, TTLBalance)
}

func (s *Service) GetUserBalance(ctx context.Context, userID string) (float64, error) {
	key := UserBalanceKey(userID)
	return s.GetFloat(ctx, key)
}

func (s *Service) UpdateUserBalance(ctx context.Context, userID string, delta float64) error {
	key := UserBalanceKey(userID)
	return s.IncrByFloat(ctx, key, delta)
}

// ==================================
// User Profile
// ==================================

func (s *Service) SetUserProfile(ctx context.Context, userID string, profile UserResponse) error {
	key := UserProfileKey(userID)
	return s.SetJSON(ctx, key, profile, TTLUserProfile)
}

func (s *Service) SetUserProfileByMail(ctx context.Context, mail string, profile UserResponse) error {
	key := UserProfileKey(mail)
	return s.SetJSON(ctx, key, profile, TTLUserProfile)
}

func (s *Service) GetUserProfile(ctx context.Context, userID string, dest *UserResponse) error {
	key := UserProfileKey(userID)
	return s.GetJSON(ctx, key, dest)
}

func (s *Service) GetUserProfileByMail(ctx context.Context, mail string, dest *UserResponse) error {
	key := UserProfileKey(mail)
	return s.GetJSON(ctx, key, dest)
}

func (s *Service) InvalidateUserProfile(ctx context.Context, userID string) error {
	key := UserProfileKey(userID)
	return s.Delete(ctx, key)
}
