package db

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/google/uuid"
)

func generateBrand(bid string) {
	// I need a base brand which will post campaign
	// Cause of dependancy of campaigns on brands
	mockBrand := &Brand{
		Id:        bid,
		Name:      "MockBrand",
		Email:     "mockbrand@gmail.com",
		Sector:    "skin_care",
		Website:   "mockbrand.com",
		Address:   "Guwahati",
		Campaigns: 0,
	}
	mockBrand.Password.Hash("random_pass")
	MockBrandStore.RegisterBrand(context.Background(), mockBrand)
}

func destroyBrand(bid string) {
	MockBrandStore.DeregisterBrand(context.Background(), bid)
}

func SeedCampaign(ctx context.Context, bid string, status, num int) []string {
	i := 0
	var ids []string
	tx, _ := MockCampaignStore.db.BeginTx(ctx, nil)
	for i < num {
		id := fmt.Sprintf("010%d", i)
		query := `
			INSERT INTO campaigns (id, brand_id, title, budget, cpm, requirements, platform, doc_link, status)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`
		args := []any{
			id, bid, fmt.Sprintf("title_%d", i), 1000.0, 101.0, "", "youtube", "", status,
		}
		_, err := tx.ExecContext(ctx, query, args...)
		if err != nil {
			tx.Rollback()
			log.Printf("Error seeding table: %v", err.Error())
			return nil
		}
		ids = append(ids, id)
		i++
	}
	tx.Commit()
	return ids
}

func destroyCampaign(ctx context.Context, ids []string) {
	for _, id := range ids {
		MockCampaignStore.DeleteCampaign(ctx, id)
	}
}

func TestLaunchCampaign(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)

	bid := uuid.New().String()
	generateBrand(bid)

	temp_camp := Campaign{
		Id:       "0001",
		BrandId:  bid,
		Title:    "mock_title",
		Budget:   1000.0,
		CPM:      101.0,
		Req:      "mock_requirements",
		Platform: "youtube",
		DocLink:  "mock_link",
		Status:   DraftStatus,
	}
	defer func() {
		MockCampaignStore.DeleteCampaign(ctx, temp_camp.Id)
		destroyBrand(bid)
		cancel()
	}()
	t.Run("creating a new campaign", func(t *testing.T) {
		err := MockCampaignStore.LaunchCampaign(ctx, &temp_camp)
		if err != nil {
			t.Fail()
		}
	})
}

func TestGetCampaign(t *testing.T) {
	bid := uuid.New().String()
	generateBrand(bid)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)

	temp_camp := Campaign{
		Id:       "0001",
		BrandId:  bid,
		Title:    "mock_title",
		Budget:   1000.0,
		CPM:      101.0,
		Req:      "mock_requirements",
		Platform: "youtube",
		DocLink:  "mock_link",
		Status:   0,
	}
	MockCampaignStore.LaunchCampaign(ctx, &temp_camp)
	defer func() {
		MockCampaignStore.DeleteCampaign(ctx, temp_camp.Id)
		destroyBrand(bid)
		cancel()
	}()
	t.Run("fetching valid campaign", func(t *testing.T) {
		_, err := MockCampaignStore.GetCampaign(ctx, temp_camp.Id)
		if err != nil {
			log.Printf("Error fetching valid campaign: %v\n", err.Error())
			t.Fail()
		}
	})
	t.Run("fetching invalid campaign", func(t *testing.T) {
		_, err := MockCampaignStore.GetCampaign(ctx, "NA")
		if err == nil {
			t.Fail()
		}
	})
}

func TestGetRecentCampaigns(t *testing.T) {
	bid := uuid.New().String()
	generateBrand(bid)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer func() {
		destroyBrand(bid)
		cancel()
	}()
	GenIds := SeedCampaign(ctx, bid, ActiveStatus, 10)
	got, _, _, err := MockCampaignStore.GetRecentCampaigns(ctx, 10, "")
	if err != nil {
		log.Printf("Fetching error in campaign feed: %v", err.Error())
		t.Fail()
	}
	if len(got) != 10 {
		log.Printf("got: %d, want %d", len(got), 10)
		t.Fail()
	}
	if len(got) > 0 && got[0].Id != "0109" {
		log.Printf("sorting failed in campaign feed: top: %s", got[0].Id)
		t.Fail()
	}
	for _, v := range GenIds {
		// Clean up the seeding
		MockCampaignStore.DeleteCampaign(ctx, v)
	}
}

func TestEndCampaign(t *testing.T) {
	bid := uuid.New().String()
	generateBrand(bid)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	temp_camp := Campaign{
		Id:       "0001",
		BrandId:  bid,
		Title:    "mock_title",
		Budget:   1000.0,
		CPM:      101.0,
		Req:      "mock_requirements",
		Platform: "youtube",
		DocLink:  "mock_link",
		Status:   0,
	}
	defer func() {
		MockCampaignStore.DeleteCampaign(ctx, temp_camp.Id)
		destroyBrand(bid)
		cancel()
	}()
	MockCampaignStore.LaunchCampaign(ctx, &temp_camp)
	t.Run("ending a campaign", func(t *testing.T) {
		err := MockCampaignStore.EndCampaign(ctx, temp_camp.Id)
		if err != nil {
			t.Fail()
		}
	})

}

func TestUpdateCampaign(t *testing.T) {
	bid := uuid.New().String()
	generateBrand(bid)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	temp_camp := Campaign{
		Id:       "0001",
		BrandId:  bid,
		Title:    "mock_title",
		Budget:   1000.0,
		CPM:      101.0,
		Req:      "mock_requirements",
		Platform: "youtube",
		DocLink:  "mock_link",
		Status:   0,
	}
	defer func() {
		MockCampaignStore.DeleteCampaign(ctx, temp_camp.Id)
		destroyBrand(bid)
		cancel()
	}()
	MockCampaignStore.LaunchCampaign(ctx, &temp_camp)
	NewTitle := "new_title"
	NewBudget := 1000.01
	NewReq := "new_req"
	NewDocLink := "new_doc_link"
	campaign_id := temp_camp.Id
	t.Run("updating a campaign", func(t *testing.T) {
		payload := UpdateCampaign{
			Title:   &NewTitle,
			Budget:  &NewBudget,
			Req:     &NewReq,
			DocLink: &NewDocLink,
		}
		err := MockCampaignStore.UpdateCampaign(ctx, campaign_id, payload)
		if err != nil {
			t.Fail()
		}
		updatedCampaign, _ := MockCampaignStore.GetCampaign(ctx, campaign_id)
		if updatedCampaign.Title != NewTitle ||
			updatedCampaign.Budget != NewBudget ||
			updatedCampaign.Req != NewReq ||
			updatedCampaign.DocLink != NewDocLink {
			log.Printf("got: %v, want: %v", updatedCampaign.Title, NewTitle)
			log.Printf("got: %v, want: %v", updatedCampaign.Budget, NewBudget)
			log.Printf("got: %v, want: %v", updatedCampaign.Req, NewReq)
			log.Printf("got: %v, want: %v", updatedCampaign.DocLink, NewDocLink)
			t.Fail()
		}

	})
}

func TestGetBrandCampaigns(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)

	brandID := uuid.New().String()
	generateBrand(brandID)
	defer func() {
		destroyBrand(brandID)
		cancel()
	}()
	t.Run("fetching brand campaigns", func(t *testing.T) {
		campaign := &Campaign{
			Id:        uuid.New().String(),
			BrandId:   brandID,
			Title:     "Brand Campaign",
			Budget:    1000,
			CPM:       10,
			Req:       "Requirements",
			Platform:  "Instagram",
			DocLink:   "http://doc.link",
			Status:    1,
			CreatedAt: time.Now().Format(time.RFC3339),
		}
		err := MockCampaignStore.LaunchCampaign(ctx, campaign)
		if err != nil {
			t.Fail()
		}

		campaigns, _, _, err := MockCampaignStore.GetBrandCampaigns(ctx, brandID, 1, "")
		if err != nil || len(campaigns) == 0 {
			log.Printf("got: %v, want: %v", len(campaigns), 1)
			t.Fail()
		}
		MockCampaignStore.DeleteCampaign(ctx, campaign.Id)
	})

}

func TestGetUserCampaigns(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)

	// create the user
	userID := uuid.New().String()
	creator := generateCreator(ctx, userID)

	// create the brand
	brandID := uuid.New().String()
	generateBrand(brandID)
	campaignID := uuid.New().String()
	defer func() {
		destroyBrand(brandID)
		destroyCreator(ctx, creator)
		cancel()
	}()
	t.Run("fetching user campaigns", func(t *testing.T) {
		// Insert campaign
		campaign := &Campaign{
			Id:        campaignID,
			BrandId:   brandID,
			Title:     "User Campaign",
			Budget:    2000,
			CPM:       20,
			Req:       "Requirements",
			Platform:  "YouTube",
			DocLink:   "http://doc.link",
			Status:    1,
			CreatedAt: time.Now().Format(time.RFC3339),
		}
		err := MockCampaignStore.LaunchCampaign(ctx, campaign)
		if err != nil {
			t.Fail()
		}
		user_submission := Submission{
			Id:         uuid.New().String(),
			CreatorId:  userID,
			CampaignId: campaignID,
			Url:        "http://submission.link",
			Status:     0,
			Views:      0,
			Earnings:   0.0,
		}
		// Make a submission for the user
		err = MockSubStore.MakeSubmission(ctx, user_submission)
		if err != nil {
			t.Fail()
		}

		campaigns, _, _, err := MockCampaignStore.GetUserCampaigns(ctx, userID, 1, "")
		if err != nil || len(campaigns) == 0 || campaigns[0].Id != campaignID {
			log.Printf("got: %v, want: %v", len(campaigns), 1)
			t.Fail()
		}
		MockSubStore.DeleteSubmission(ctx, user_submission.Id)
		MockCampaignStore.DeleteCampaign(ctx, campaign.Id)
	})

}
