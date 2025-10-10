package db

import (
	"context"
	"log"
	"testing"
	"time"
)

func TestGetBrand(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
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
			Email:     "mockbrand@gmail.com",
			Sector:    "skin_care",
			Website:   "mockbrand.com",
			Address:   "Guwahati",
			Campaigns: 0,
		}
		mockBrand.Password.Hash("random_pass")
		err := MockBrandStore.RegisterBrand(ctx, mockBrand)
		if err != nil {
			t.Fail()
		}
		MockBrandStore.DeregisterBrand(ctx, "0001")
	})
}

func TestUpdateBrand(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	// creating a mock entity
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
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

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
