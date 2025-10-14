package db

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/google/uuid"
)

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

func SeedSubmissions(ctx context.Context, num, status int) []string {
	i := 0
	var ids []string
	tx, _ := MockSubStore.db.BeginTx(ctx, nil)
	for i < num {
		id := fmt.Sprintf("001%d", i)
		query := `
			INSERT INTO submissions 
			(
				id, creator_id, campaign_id, url, status, video_title, video_platform,
				platform_video_id, thumbnail_url, views, like_count, video_status, earnings,
				sync_frequency
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		`
		args := []any{
			id, "0001", "0100", "mock_url", status,
			"Test_Title", "youtube", "testvid001", "example.com", 1000, 100,
			"available", 0.0, 5,
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

	// create mock creator
	creator := generateCreator(ctx, "0001")

	// mock brand
	bid := uuid.New().String()
	generateBrand(bid)

	// mock campaign
	camp := SeedCampaign(ctx, bid, 1)
	defer func() {
		destroyCampaign(ctx, camp)
		destroyBrand(bid)
		destroyCreator(ctx, creator)
		cancel()
	}()
	t.Run("mock submission", func(t *testing.T) {
		sub := Submission{
			Id:         "0001",
			CreatorId:  creator,
			CampaignId: camp[0],
			Url:        "example.com",
			Status:     DraftStatus,
			Views:      100,
			Earnings:   400.0,
		}
		err := MockSubStore.MakeSubmission(ctx, sub)
		if err != nil {
			t.Fail()
		}

		MockSubStore.DeleteSubmission(ctx, "0001")
	})

}

func TestFindSubmission(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)

	// create mock creator
	creator := generateCreator(ctx, "0001")

	// mock brand
	bid := uuid.New().String()
	generateBrand(bid)

	// mock campaign
	camp := SeedCampaign(ctx, bid, 1)

	// mock submission
	sub := Submission{
		Id:         "0001",
		CreatorId:  creator,
		CampaignId: camp[0],
		Url:        "example.com",
		Status:     DraftStatus,
		Views:      100,
		Earnings:   400.0,
	}
	MockSubStore.MakeSubmission(ctx, sub)
	defer func() {
		MockSubStore.DeleteSubmission(ctx, sub.Id)
		destroyCampaign(ctx, camp)
		destroyBrand(bid)
		destroyCreator(ctx, creator)
		cancel()
	}()
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

	// create mock creator
	creator := generateCreator(ctx, "0001")

	// mock brand
	bid := uuid.New().String()
	generateBrand(bid)

	// mock campaign
	camp := SeedCampaign(ctx, bid, 1)

	// mock submissions
	ids := SeedSubmissions(ctx, 10, DraftStatus)
	log.Printf("%d", len(ids))
	defer func() {
		destroySubmissions(ctx, ids)
		destroyCampaign(ctx, camp)
		destroyBrand(bid)
		destroyCreator(ctx, creator)
		cancel()
	}()

	t.Run("filtering submisisons", func(t *testing.T) {
		time_filter := fmt.Sprintf("%02d-%d", int(time.Now().UTC().Month()), time.Now().UTC().Year())
		got, err := MockSubStore.FindSubmissionsByFilters(ctx,
			Filter{
				Time: &time_filter,
			},
			10, // limit
			0,  //offset
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

	// create mock creator
	creator := generateCreator(ctx, "0001")

	// mock brand
	bid := uuid.New().String()
	generateBrand(bid)

	// mock campaign
	camp := SeedCampaign(ctx, bid, 1)

	// mock submissions
	ids := SeedSubmissions(ctx, 1, DraftStatus)
	defer func() {
		destroySubmissions(ctx, ids)
		destroyCampaign(ctx, camp)
		destroyBrand(bid)
		destroyCreator(ctx, creator)
		cancel()
	}()
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

	// create mock creator
	creator := generateCreator(ctx, "0001")

	// mock brand
	bid := uuid.New().String()
	generateBrand(bid)

	// mock campaign
	camp := SeedCampaign(ctx, bid, 1)

	// mock submission
	sub := Submission{
		Id:         "0001",
		CreatorId:  creator,
		CampaignId: camp[0],
		Url:        "example.com",
		Status:     DraftStatus,
		Views:      0,
		Earnings:   400.0,
	}
	MockSubStore.MakeSubmission(ctx, sub)

	defer func() {
		MockSubStore.DeleteSubmission(ctx, sub.Id)
		destroyCampaign(ctx, camp)
		destroyBrand(bid)
		destroyCreator(ctx, creator)
		cancel()
	}()
	t.Run("deleting invalid submission", func(t *testing.T) {
		err := MockSubStore.DeleteSubmission(ctx, "NA") // not possible
		if err == nil {
			t.Fail()
		}
	})
}

func TestChangeViews(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	// create mock creator
	creator := generateCreator(ctx, "0001")
	// mock brand
	bid := uuid.New().String()
	generateBrand(bid)
	// mock campaign
	camp := SeedCampaign(ctx, bid, 1)
	// mock submission
	sub := Submission{
		Id:         "0001",
		CreatorId:  creator,
		CampaignId: camp[0],
		Url:        "example.com",
		Status:     DraftStatus,
		Views:      0,
		Earnings:   400.0,
	}
	defer func() {
		destroyCampaign(ctx, camp)
		destroyBrand(bid)
		destroyCreator(ctx, creator)
		cancel()
	}()

	t.Run("OK", func(t *testing.T) {
		MockSubStore.MakeSubmission(ctx, sub)
		delta := 100
		err := MockSubStore.ChangeViews(ctx, delta, sub.Id)
		// require no error
		if err != nil {
			t.Fail()
		}
		modifiedSub, err := MockSubStore.FindSubmissionById(ctx, sub.Id)
		if err != nil {
			log.Printf("could not get the modified submission\n")
			t.Fail()
		}
		if modifiedSub.Views-sub.Views != delta {
			log.Printf("want: %d, got: %d", delta+sub.Views, modifiedSub.Views)
			t.Fail()
		}
		MockSubStore.DeleteSubmission(ctx, sub.Id)
	})
	t.Run("invalid submission id", func(t *testing.T) {
		delta := 100
		err := MockSubStore.ChangeViews(ctx, delta, "invalid-id")
		// require error
		if err == nil {
			t.Fail()
		}
	})
}

func TestGetSubmissionsForSync(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), TestTimeout)

	// user generation
	uid := "0001"
	generateCreator(ctx, uid)

	// brand genration
	bid := "dummy_brand_id_001"
	generateBrand(bid)

	// launch a campaign
	campaignIds := SeedCampaign(ctx, bid, 1)

	// cleanup
	defer func() {
		destroyCampaign(ctx, campaignIds)
		destroyBrand(bid)
		destroyCreator(ctx, uid)
		cancel()
	}()

	t.Run("Time before sync_frequency", func(t *testing.T) {
		// submissions created
		SubsCount := 10
		submissionIds := SeedSubmissions(ctx, SubsCount, ActiveStatus)
		polling_subs, err := MockSubStore.GetSubmissionsForSync(ctx)
		if err != nil {
			log.Printf("could not get submissions for polling\n")
			t.Fail()
		}

		if len(polling_subs) > 0 {
			log.Printf("wanted: %d, got: %d\n", 0, len(polling_subs))
			t.Fail()
		}

		// clean up
		destroySubmissions(ctx, submissionIds)
	})

	t.Run("OK", func(t *testing.T) {
		// submissions created
		SubsCount := 1
		sub := Submission{
			Id:         "0001",
			CreatorId:  uid,
			CampaignId: campaignIds[0],
			Url:        "example.com",
			Status:     ActiveStatus,
			Views:      10000, // dummy values
			Earnings:   400.0, // dummy values
			// shift the created time back by 10 min
			// now the DB will see this as valid to be synced submission
			// as this will by default have sync_frequency 5 min
			LastSyncedAt: time.Now().Add(-time.Minute * 10).Format("2006-01-02 15:04:05-07:00"),
			CreatedAt:    time.Now().Add(-time.Minute * 10).Format("2006-01-02 15:04:05-07:00"),
		}

		query := `
        INSERT INTO submissions
        (
            id, creator_id, campaign_id, url, status, video_title, video_platform,
            platform_video_id, thumbnail_url, views, like_count, video_status, earnings,
            created_at, last_synced_at
        )
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
    `
		_, err := MockSubStore.db.ExecContext(ctx, query,
			sub.Id, sub.CreatorId, sub.CampaignId, sub.Url, sub.Status,
			sub.VideoTitle, sub.VideoPlatform, sub.VideoID, sub.ThumbnailURL,
			sub.Views, sub.LikeCount, sub.VideoStatus, sub.Earnings, sub.CreatedAt,
			sub.LastSyncedAt,
		)
		if err != nil {
			// internal server error
			log.Printf("error making submission: %v\n", err.Error())
			t.Fail()
		}

		polling_subs, err := MockSubStore.GetSubmissionsForSync(ctx)
		if err != nil {
			log.Printf("could not get submissions for polling\n")
			t.Fail()
		}

		if len(polling_subs) != SubsCount {
			log.Printf("wanted: %d, got: %d\n", 1, len(polling_subs))
			t.Fail()
		}

		// clean up
		destroySubmissions(ctx, []string{sub.Id})
	})
}
