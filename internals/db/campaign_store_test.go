package db

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/joho/godotenv"
)

var MockCampaignStore CampaignStore

func init() {
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Fatal("Error loading .env", err.Error())
	}
	MockCampaignStore = CampaignStore{
		db: MockDB,
	}
}

func generateBrand() {
	// I need a base brand which will post campaign
	// Cause of dependancy of campaigns on brands
	mockBrand := &Brand{
		Id:        "0001",
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

func destroyBrand() {
	MockBrandStore.DeregisterBrand(context.Background(), "0001")
}

func SeedCampaign(ctx context.Context, num int) []string {
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
			id, "0001", fmt.Sprintf("title_%d", i), 1000.0, 101.0, "", "youtube", "", 0,
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

func destroyCampaign(ctx context.Context, id string) {
	MockCampaignStore.DeleteCampaign(ctx, id)
}

func TestLaunchCampaign(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	generateBrand()
	defer destroyBrand()
	temp_camp := Campaign{
		Id:       "0001",
		BrandId:  "0001",
		Title:    "mock_title",
		Budget:   1000.0,
		CPM:      101.0,
		Req:      "mock_requirements",
		Platform: "youtube",
		DocLink:  "mock_link",
		Status:   0,
	}
	t.Run("creating a new campaign", func(t *testing.T) {
		err := MockCampaignStore.LaunchCampaign(ctx, &temp_camp)
		if err != nil {
			t.Fail()
		}
	})
	MockCampaignStore.DeleteCampaign(ctx, temp_camp.Id)
}

func TestGetCampaign(t *testing.T) {
	generateBrand()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	defer destroyBrand()
	temp_camp := Campaign{
		Id:       "0001",
		BrandId:  "0001",
		Title:    "mock_title",
		Budget:   1000.0,
		CPM:      101.0,
		Req:      "mock_requirements",
		Platform: "youtube",
		DocLink:  "mock_link",
		Status:   0,
	}
	MockCampaignStore.LaunchCampaign(ctx, &temp_camp)
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
	MockCampaignStore.DeleteCampaign(ctx, temp_camp.Id)
}

func TestGetRecentCampaigns(t *testing.T) {
	generateBrand()
	defer destroyBrand()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	GenIds := SeedCampaign(ctx, 10)
	got, err := MockCampaignStore.GetRecentCampaigns(ctx, 0, 10)
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
	generateBrand()
	defer destroyBrand()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	temp_camp := Campaign{
		Id:       "0001",
		BrandId:  "0001",
		Title:    "mock_title",
		Budget:   1000.0,
		CPM:      101.0,
		Req:      "mock_requirements",
		Platform: "youtube",
		DocLink:  "mock_link",
		Status:   0,
	}
	MockCampaignStore.LaunchCampaign(ctx, &temp_camp)
	t.Run("ending a campaign", func(t *testing.T) {
		err := MockCampaignStore.EndCampaign(ctx, temp_camp.Id)
		if err != nil {
			t.Fail()
		}
	})
	MockCampaignStore.DeleteCampaign(ctx, temp_camp.Id)
}

func TestUpdateCampaign(t *testing.T) {
	generateBrand()
	defer destroyBrand()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	temp_camp := Campaign{
		Id:       "0001",
		BrandId:  "0001",
		Title:    "mock_title",
		Budget:   1000.0,
		CPM:      101.0,
		Req:      "mock_requirements",
		Platform: "youtube",
		DocLink:  "mock_link",
		Status:   0,
	}
	MockCampaignStore.LaunchCampaign(ctx, &temp_camp)
	NewTitle := "new_title"
	NewBudget := 1000.01
	NewReq := "new_req"
	NewDocLink := "new_doc_link"
	t.Run("updating a campaign", func(t *testing.T) {
		payload := UpdateCampaign{
			Id:      temp_camp.Id,
			Title:   &NewTitle,
			Budget:  &NewBudget,
			Req:     &NewReq,
			DocLink: &NewDocLink,
		}
		err := MockCampaignStore.UpdateCampaign(ctx, payload)
		if err != nil {
			t.Fail()
		}
		updatedCampaign, _ := MockCampaignStore.GetCampaign(ctx, payload.Id)
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
		MockCampaignStore.DeleteCampaign(ctx, temp_camp.Id)
	})
}
