package db

import (
	"context"
	"fmt"
	"log"
	"testing"
)

func TestGetBrand(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), TestTimeout)
	defer cancel()
	mockBrand := &Brand{
		Id:        "0001",
		Name:      "MockBrand",
		Email:     fmt.Sprintf("%s@mockbrand.com", "0001"),
		Sector:    "skin_care",
		Website:   "mockbrand.com",
		Address:   "Guwahati",
		Campaigns: 0,
	}
	mockBrand.Password.Hash("random_pass")
	// ensure any pre-existing brand with this id is removed to avoid duplicate key errors
	MockBrandStore.DeregisterBrand(ctx, mockBrand.Id)
	// also ensure no other brand exists with this email
	MockBrandStore.DeregisterBrand(ctx, "0001")
	err := MockBrandStore.RegisterBrand(ctx, mockBrand)
	if err != nil {
		t.Fail()
	}
	t.Run("Getting brand by id", func(t *testing.T) {
		_, err = MockBrandStore.GetBrandById(ctx, "0001") // This is a given id that will always exist
		if err != nil {
			t.Fail()
		}
	})
	t.Run("Fetching Invalid brand id", func(t *testing.T) {
		_, err := MockBrandStore.GetBrandById(ctx, "NA") // Some random id
		if err == nil {
			t.Fail()
		}
		_, err = MockBrandStore.GetBrandById(ctx, "")
		if err == nil {
			// We expect an occur to occur
			t.Fail()
		}
	})
	t.Run("Filtering Brands", func(t *testing.T) {
		// sector | campaigns | name
		_, err := MockBrandStore.GetBrandsByFilter(ctx, "sector", "skin_care")
		// valid sector
		if err != nil {
			t.Fail()
		}
		_, err = MockBrandStore.GetBrandsByFilter(ctx, "campaigns", 10)
		// valid companies with no. of campaigns over 10
		if err != nil {
			t.Fail()
		}
		_, err = MockBrandStore.GetBrandsByFilter(ctx, "sector", "skin_care")
		// valid name
		if err != nil {
			t.Fail()
		}
		_, err = MockBrandStore.GetBrandsByFilter(ctx, "NA", "NA")
		// invalid filter NA
		if err == nil {
			t.Fail()
		}
	})
	MockBrandStore.DeregisterBrand(ctx, "0001")
}

func TestCreateBrand(t *testing.T) {
	ctx := t.Context()
	t.Run("Creating a valid brand entity", func(t *testing.T) {
		mockBrand := &Brand{
			Id:        "0001",
			Name:      "MockBrand",
			Email:     fmt.Sprintf("%s@mockbrand.com", "0001"),
			Sector:    "skin_care",
			Website:   "mockbrand.com",
			Address:   "Guwahati",
			Campaigns: 0,
		}
		mockBrand.Password.Hash("random_pass")
		// ensure no pre-existing brand with this id
		MockBrandStore.DeregisterBrand(ctx, mockBrand.Id)
		err := MockBrandStore.RegisterBrand(ctx, mockBrand)
		if err != nil {
			t.Fail()
		}
		MockBrandStore.DeregisterBrand(ctx, "0001")
	})
}

func TestUpdateBrand(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), TestTimeout)
	defer cancel()
	// creating a mock entity
	mockBrand := &Brand{
		Id:        "0001",
		Name:      "MockBrand",
		Email:     fmt.Sprintf("%s@mockbrand.com", "0001"),
		Sector:    "skin_care",
		Website:   "mockbrand.com",
		Address:   "Guwahati",
		Campaigns: 0,
	}
	mockBrand.Password.Hash("random_pass")
	// ensure no pre-existing brand with this id
	MockBrandStore.DeregisterBrand(ctx, mockBrand.Id)
	MockBrandStore.RegisterBrand(ctx, mockBrand)

	// Mock the changes to be made
	NewEmail := "newmail@gmail.com"
	NewWebsite := "newwebsite.com"
	NewAddress := "newaddress"
	t.Run("Updating Brand Entity", func(t *testing.T) {
		payload := BrandUpdatePayload{
			Email:   &NewEmail,
			Website: &NewWebsite,
			Address: &NewAddress,
		}
		err := MockBrandStore.UpdateBrand(ctx, mockBrand.Id, payload)
		if err != nil {
			t.Fail()
		}
		updatedBrand, _ := MockBrandStore.GetBrandById(ctx, mockBrand.Id)
		if updatedBrand.Address != NewAddress ||
			updatedBrand.Email != NewEmail ||
			updatedBrand.Website != NewWebsite {
			log.Printf("got: %s, want: %s\n", updatedBrand.Address, NewAddress)
			log.Printf("got: %s, want: %s\n", updatedBrand.Email, NewEmail)
			log.Printf("got: %s, want: %s\n", updatedBrand.Website, NewWebsite)
			t.Fail()
		}
	})
	MockBrandStore.DeregisterBrand(ctx, mockBrand.Id)
}

func TestUpdateBrandPassword(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), TestTimeout)

	// creating a dummy brand
	bid := "dummy_brand_01"
	generateBrand(bid)
	defer func() {
		destroyBrand(bid)
		cancel()
	}()
	t.Run("OK", func(t *testing.T) {
		new_pass := "random_password"
		err := MockBrandStore.ChangePassword(ctx, bid, new_pass)
		if err != nil {
			t.Fail()
		}
	})
	t.Run("Invalid ID", func(t *testing.T) {
		new_pass := "random_password"
		err := MockBrandStore.ChangePassword(ctx, "NA", new_pass)
		if err == nil {
			t.Fail()
		}
	})
	t.Run("violating minimum password length", func(t *testing.T) {
		new_pass := "123456" // min length is 8
		err := MockBrandStore.ChangePassword(ctx, bid, new_pass)
		if err == nil {
			t.Fail()
		}
	})
}

func TestDeleteBrand(t *testing.T) {
	// context
	ctx, cancel := context.WithTimeout(context.Background(), TestTimeout)

	// generate a dummy brand
	bid := "dummy_brand_01"
	generateBrand(bid)

	defer func() {
		cancel()
		destroyBrand(bid)
	}()

	t.Run("Delete brand with an active campaign", func(t *testing.T) {
		campaign := Campaign{
			Id:       "dummy_campaign_001",
			BrandId:  bid,
			Title:    "mock_title",
			Budget:   1000.0,
			CPM:      101.0,
			Req:      "mock_requirements",
			Platform: "youtube",
			DocLink:  "mock_link",
			Status:   ActiveStatus,
		}
		err := MockCampaignStore.LaunchCampaign(ctx, &campaign)
		if err != nil {
			// could not create a campaign
			t.FailNow()
		}
		err = MockBrandStore.DeregisterBrand(ctx, bid)
		if err == nil {
			// we should not be able to deregister
			// a brand with an active campaign
			t.Fail()
		}
		// clean up
		MockCampaignStore.DeleteCampaign(ctx, campaign.Id)
	})
}
