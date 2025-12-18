package cache

import (
	"context"
	"fmt"
	"time"
)

// ==================================
// Campaign Details
// ==================================

func (s *Service) SetCampaign(ctx context.Context, campaignID string, campaign any) error {
	key := CampaignKey(campaignID)
	return s.SetJSON(ctx, key, campaign, TTLCampaign)
}

func (s *Service) GetCampaign(ctx context.Context, campaignID string, dest any) error {
	key := CampaignKey(campaignID)
	return s.GetJSON(ctx, key, dest)
}

func (s *Service) InvalidateCampaign(ctx context.Context, campaignID string) error {
	key := CampaignKey(campaignID)
	return s.Delete(ctx, key)
}

func (s *Service) InvalidateOneBrandCampaign(ctx context.Context, CompanyID, CampaignID string) error {
	key := CompanyCampaignsKey(CompanyID)
	return s.client.SRem(ctx, key, []string{CampaignID}).Err()
}

func (s *Service) InvalidateOneUserCampaign(ctx context.Context, userID, CampaignID string) error {
	key := UserCampaignsKey(userID)
	return s.client.SRem(ctx, key, []string{CampaignID}).Err()
}

// ==================================
// Campaign Budget
// ==================================

func (s *Service) SetCampaignBudget(ctx context.Context, campaignID string, amount float64) error {
	key := CampaignBudgetKey(campaignID)
	return s.Set(ctx, key, amount, TTLCampaign)
}

func (s *Service) GetCampaignBudget(ctx context.Context, campaignID string) (float64, error) {
	key := CampaignBudgetKey(campaignID)
	return s.GetFloat(ctx, key)
}

func (s *Service) DecrementCampaignBudget(ctx context.Context, campaignID string, amount float64) error {
	key := CampaignBudgetKey(campaignID)
	return s.IncrByFloat(ctx, key, -amount)
}

// ==================================
// Active Campaigns List
// ==================================

func (s *Service) SetActiveCampaigns(ctx context.Context, campaignIDs []string) error {
	key := ActiveCampaignsKey()

	// Clear existing set
	s.Delete(ctx, key)

	if len(campaignIDs) == 0 {
		return nil
	}

	// Add all campaigns
	members := make([]any, len(campaignIDs))
	for i, id := range campaignIDs {
		members[i] = id
	}

	if err := s.SAdd(ctx, key, members...); err != nil {
		return err
	}

	// Set expiration
	return s.client.Expire(ctx, key, TTLActiveCamps).Err()
}

func (s *Service) GetActiveCampaigns(ctx context.Context) ([]string, error) {
	key := ActiveCampaignsKey()
	return s.SMembers(ctx, key)
}

func (s *Service) AddActiveCampaign(ctx context.Context, campaignID string) error {
	key := ActiveCampaignsKey()
	return s.SAdd(ctx, key, campaignID)
}

func (s *Service) RemoveActiveCampaign(ctx context.Context, campaignID string) error {
	key := ActiveCampaignsKey()
	return s.SRem(ctx, key, campaignID)
}

// ==================================
// Company Campaigns List
// ==================================

func (s *Service) SetUserCampaigns(ctx context.Context, userID, cursor string, campaignIDs []string) error {
	key := UserCampaignsKey(fmt.Sprintf("%s-%s", userID, cursor))

	// Clear existing set
	s.Delete(ctx, key)

	if len(campaignIDs) == 0 {
		return nil
	}

	// Add all campaigns
	members := make([]any, len(campaignIDs))
	for i, id := range campaignIDs {
		members[i] = id
	}

	if err := s.SAdd(ctx, key, members...); err != nil {
		return err
	}

	// Set expiration
	return s.client.Expire(ctx, key, 15*time.Minute).Err()
}

func (s *Service) SetCompanyCampaigns(ctx context.Context, companyID string, campaignIDs []string) error {
	key := CompanyCampaignsKey(companyID)

	// Clear existing set
	s.Delete(ctx, key)

	if len(campaignIDs) == 0 {
		return nil
	}

	// Add all campaigns
	members := make([]any, len(campaignIDs))
	for i, id := range campaignIDs {
		members[i] = id
	}

	if err := s.SAdd(ctx, key, members...); err != nil {
		return err
	}

	// Set expiration
	return s.client.Expire(ctx, key, 15*time.Minute).Err()
}

func (s *Service) GetCompanyCampaigns(ctx context.Context, companyID string) ([]string, error) {
	key := CompanyCampaignsKey(companyID)
	return s.SMembers(ctx, key)
}

func (s *Service) GetUserCampaigns(ctx context.Context, userID, cursor string) ([]string, error) {
	key := UserCampaignsKey(fmt.Sprintf("%s-%s", userID, cursor))
	return s.SMembers(ctx, key)
}

func (s *Service) InvalidateCompanyCampaigns(ctx context.Context, companyID string) error {
	key := CompanyCampaignsKey(companyID)
	return s.Delete(ctx, key)
}

func (s *Service) InvalidateUserCampaigns(ctx context.Context, userID string) error {
	key := UserCampaignsKey(userID)
	return s.Delete(ctx, key)
}

// ==================================
// Pending Applications Count
// ==================================

func (s *Service) SetPendingApplicationsCount(ctx context.Context, companyID string, count int) error {
	key := PendingApplicationsKey(companyID)
	return s.Set(ctx, key, count, 10*time.Minute)
}

func (s *Service) GetPendingApplicationsCount(ctx context.Context, companyID string) (int, error) {
	key := PendingApplicationsKey(companyID)
	return s.GetInt(ctx, key)
}

func (s *Service) IncrementPendingApplicationsCount(ctx context.Context, companyID string) error {
	key := PendingApplicationsKey(companyID)
	return s.Incr(ctx, key)
}

func (s *Service) DecrementPendingApplicationsCount(ctx context.Context, companyID string) error {
	key := PendingApplicationsKey(companyID)
	return s.Decr(ctx, key)
}
