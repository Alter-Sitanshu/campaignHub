package db

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"
)

var MockSubStore SubmissionStore

func init() {
	MockSubStore.db = MockDB
}

func generateCreator(ctx context.Context, mockUserId string) string {
	query := `
		INSERT INTO users (id, first_name, last_name, email, password, gender, age, role)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	args := []any{
		mockUserId, "mock_first_name", "mock_last_name", "email@gmail.com",
		"password", "O", 20, "LVL1",
	}
	MockUserStore.db.ExecContext(ctx, query, args...)
	return mockUserId
}

func destroyCreator(ctx context.Context, id string) {
	query := `
		DELETE FROM users WHERE id = $1
	`
	MockUserStore.db.ExecContext(ctx, query, id)
}

func SeedSubmissions(ctx context.Context, num int) []string {
	i := 0
	var ids []string
	tx, _ := MockSubStore.db.BeginTx(ctx, nil)
	for i < num {
		id := fmt.Sprintf("001%d", i)
		query := `
			INSERT INTO submissions (id, creator_id, campaign_id, url, status)
			VALUES ($1, $2, $3, $4, $5)
		`
		args := []any{
			id, "0001", "0100", "mock_url", DraftStatus,
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

func destroySubmissions(ctx context.Context, ids []string) {
	for _, v := range ids {
		MockSubStore.DeleteSubmission(ctx, v)
	}
}

func TestMakeSubmission(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	// create mock creator
	creator := generateCreator(ctx, "0001")
	defer destroyCreator(ctx, creator)
	// mock brand
	generateBrand() // brandid - 0001
	defer destroyBrand()
	// mock campaign
	camp := SeedCampaign(ctx, 1)[0]
	defer destroyCampaign(ctx, camp)
	t.Run("mock submission", func(t *testing.T) {
		sub := Submission{
			Id:         "0001",
			CreatorId:  creator,
			CampaignId: camp,
			Url:        "example.com",
			Status:     DraftStatus,
			Views:      100,
			Earnings:   400.0,
		}
		err := MockSubStore.MakeSubmission(ctx, &sub)
		if err != nil {
			t.Fail()
		}

		MockSubStore.DeleteSubmission(ctx, "0001")
	})

}

func TestFindSubmission(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	// create mock creator
	creator := generateCreator(ctx, "0001")
	defer destroyCreator(ctx, creator)
	// mock brand
	generateBrand() // brandid - 0001
	defer destroyBrand()
	// mock campaign
	camp := SeedCampaign(ctx, 1)[0]
	defer destroyCampaign(ctx, camp)

	// mock submission
	sub := Submission{
		Id:         "0001",
		CreatorId:  creator,
		CampaignId: camp,
		Url:        "example.com",
		Status:     DraftStatus,
		Views:      100,
		Earnings:   400.0,
	}
	MockSubStore.MakeSubmission(ctx, &sub)
	defer MockSubStore.DeleteSubmission(ctx, sub.Id)
	t.Run("finding valid submission", func(t *testing.T) {
		_, err := MockSubStore.FindSubmissionById(ctx, sub.Id)
		if err != nil {
			t.Fail()
		}
	})
	t.Run("finding invalid submission", func(t *testing.T) {
		_, err := MockSubStore.FindSubmissionById(ctx, "NA")
		if err == nil {
			t.Fail()
		}
		_, err = MockSubStore.FindSubmissionById(ctx, "")
		if err == nil {
			t.Fail()
		}
	})
}

func TestFilteringSubmissions(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	// create mock creator
	creator := generateCreator(ctx, "0001")
	defer destroyCreator(ctx, creator)
	// mock brand
	generateBrand() // brandid - 0001
	defer destroyBrand()
	// mock campaign
	camp := SeedCampaign(ctx, 1)[0]
	defer destroyCampaign(ctx, camp)

	// mock submissions
	ids := SeedSubmissions(ctx, 10)
	log.Printf("%d", len(ids))
	defer destroySubmissions(ctx, ids)

	t.Run("filtering submisisons", func(t *testing.T) {
		time_filter := fmt.Sprintf("%02d-%d", int(time.Now().UTC().Month()), time.Now().UTC().Year())
		got, err := MockSubStore.FindSubmissionsByFilters(ctx,
			Filter{
				Time: &time_filter,
			},
		)

		if err != nil {
			t.Fail()
		}
		if len(got) != 10 {
			log.Printf("filtering by time failed: got(%d)", len(got))
			t.Fail()
		}
	})
}

func TestUpdateSubmissions(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	// create mock creator
	creator := generateCreator(ctx, "0001")
	defer destroyCreator(ctx, creator)
	// mock brand
	generateBrand() // brandid - 0001
	defer destroyBrand()
	// mock campaign
	camp := SeedCampaign(ctx, 1)[0]
	defer destroyCampaign(ctx, camp)

	// mock submissions
	ids := SeedSubmissions(ctx, 1)
	defer destroySubmissions(ctx, ids)
	new_stat := 3
	new_url := "new_url.com"
	new_views := 1000
	new_earnings := 1000.0
	t.Run("updating with valid params", func(t *testing.T) {
		payload := UpdateSubmission{
			Id:       ids[0],
			Status:   &new_stat,
			Url:      &new_url,
			Views:    &new_views,
			Earnings: &new_earnings,
		}
		err := MockSubStore.UpdateSubmission(ctx, payload)
		if err != nil {
			t.Fail()
		}
		updatedSub, _ := MockSubStore.FindSubmissionById(ctx, ids[0])
		if updatedSub.Status != new_stat ||
			updatedSub.Url != new_url ||
			updatedSub.Views != new_views ||
			updatedSub.Earnings != new_earnings {
			t.Fail()
		}
	})
	t.Run("invalid arguments payload", func(t *testing.T) {
		payload := UpdateSubmission{}
		err := MockSubStore.UpdateSubmission(ctx, payload)
		if err == nil {
			t.Fail()
		}
	})
}

func TestDeleteSub(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	// create mock creator
	creator := generateCreator(ctx, "0001")
	defer destroyCreator(ctx, creator)
	// mock brand
	generateBrand() // brandid - 0001
	defer destroyBrand()
	// mock campaign
	camp := SeedCampaign(ctx, 1)[0]
	defer destroyCampaign(ctx, camp)

	// mock submission
	sub := Submission{
		Id:         "0001",
		CreatorId:  creator,
		CampaignId: camp,
		Url:        "example.com",
		Status:     DraftStatus,
		Views:      100,
		Earnings:   400.0,
	}
	MockSubStore.MakeSubmission(ctx, &sub)
	defer MockSubStore.DeleteSubmission(ctx, sub.Id)
	t.Run("deleting invalid submission", func(t *testing.T) {
		err := MockSubStore.DeleteSubmission(ctx, "NA") // not possible
		if err == nil {
			t.Fail()
		}
	})
}
