package db

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

var MockBrandStore BrandStore

func init() {
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Fatal("Error loading .env", err.Error())
	}
	MockBrandStore = BrandStore{
		db: MockDB,
	}
}

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
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
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
	NewPass := "newpass"
	NewAddress := "newaddress"
	t.Run("Updating Brand Entity", func(t *testing.T) {
		payload := BrandUpdatePayload{
			Id:      mockBrand.Id,
			Email:   &NewEmail,
			Website: &NewWebsite,
			NewPass: &NewPass,
			Address: &NewAddress,
		}
		err := MockBrandStore.UpdateBrand(ctx, payload)
		if err != nil {
			t.Fail()
		}
		updatedBrand, _ := MockBrandStore.GetBrandById(ctx, payload.Id)
		if updatedBrand.Address != NewAddress ||
			updatedBrand.Email != NewEmail ||
			updatedBrand.Website != NewWebsite {
			log.Printf("got: %s, want: %s\n", updatedBrand.Address, NewAddress)
			log.Printf("got: %s, want: %s\n", updatedBrand.Email, NewEmail)
			log.Printf("got: %s, want: %s\n", updatedBrand.Website, NewWebsite)
			t.Fail()
		}
		checkPass := bcrypt.CompareHashAndPassword(updatedBrand.Password.hashed_pass, []byte(NewPass))
		if checkPass != nil {
			log.Printf("Passwords did not match\n")
			t.Fail()
		}
	})
	MockBrandStore.DeregisterBrand(ctx, mockBrand.Id)
}
