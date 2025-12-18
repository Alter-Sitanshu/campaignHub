package db

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"
)

func destroyApplications(ctx context.Context, applIDs []string) {
	for i := range applIDs {
		MockApplicationStore.DeleteApplication(ctx, applIDs[i])
	}
}

func SeedApplications(ctx context.Context, campIDs []string, uid string) []string {
	i := 0
	var ids []string
	tx, _ := MockApplicationStore.db.BeginTx(ctx, nil)
	for i < len(campIDs) {
		id := fmt.Sprintf("010%d", i)
		query := `
			INSERT INTO applications (id, campaign_id, creator_id)
			VALUES ($1, $2, $3)
		`
		args := []any{id, campIDs[i], uid}
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

func TestCreateApplications(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	uid := "dummy_application_user_01"
	bid := "dummy_brand_01"
	appl_id := "dummy_application_01"

	// create a dummy creator
	generateCreator(ctx, uid)

	// create a dummy brand
	generateBrand(bid)

	// launch a dummy campaign
	camp := SeedCampaign(ctx, bid, ActiveStatus, 1)
	defer func() {
		destroyCampaign(ctx, camp)
		destroyBrand(bid)
		destroyCreator(ctx, uid)
		cancel()
	}()

	t.Run("OK", func(t *testing.T) {
		// Delay to ensure the creation of the entities
		time.Sleep(time.Millisecond * 50)
		appl := CampaignApplication{
			Id:         appl_id,
			CampaignId: camp[0],
			CreatorId:  uid,
		}
		err := MockApplicationStore.CreateApplication(
			ctx, appl,
		)
		// no errors should occur
		if err != nil {
			t.Fail()
		}
		MockApplicationStore.DeleteApplication(ctx, appl_id)
	})
	t.Run("invalid campaign id", func(t *testing.T) {
		appl := CampaignApplication{
			Id:         appl_id,
			CampaignId: "NA", //invalid id
			CreatorId:  uid,
		}
		err := MockApplicationStore.CreateApplication(
			ctx, appl,
		)
		// errors should occur
		if err == nil {
			t.Fail()
		}
	})
	t.Run("invalid creator id", func(t *testing.T) {
		appl := CampaignApplication{
			Id:         appl_id,
			CampaignId: camp[0],
			CreatorId:  "NA",
		}
		err := MockApplicationStore.CreateApplication(
			ctx, appl,
		)
		// errors should occur
		if err == nil {
			t.Fail()
		}
	})
}

func TestGetCreatorApplications(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	uid := "dummy_user_01"
	bid := "dummy_brand_01"

	// create a dummy creator
	generateCreator(ctx, uid)

	// create a dummy brand
	generateBrand(bid)

	campIDs := SeedCampaign(ctx, bid, ActiveStatus, 10)
	applIDs := SeedApplications(ctx, campIDs, uid)

	defer func() {
		destroyApplications(ctx, applIDs)
		destroyCampaign(ctx, campIDs)
		destroyCreator(ctx, uid)
		destroyBrand(bid)
		cancel()
	}()
	t.Run("OK", func(t *testing.T) {
		appls, err := MockApplicationStore.GetCreatorApplications(ctx, uid, 0, 10)
		// errors are not expected
		if err != nil {
			t.Fail()
		}
		if len(appls) != 10 {
			log.Printf("wanted: %d, got: %d\n", 10, len(appls))
			t.Fail()
		}
	})
	t.Run("invalid uid", func(t *testing.T) {
		// invalids user id requested
		_, err := MockApplicationStore.GetCreatorApplications(ctx, "", 0, 10)
		// errors are expected
		if err == nil {
			t.Fail()
		}
	})
	t.Run("negative offset/limit", func(t *testing.T) {
		_, err := MockApplicationStore.GetCreatorApplications(ctx, uid, -1, 10)
		// errors are expected
		if err == nil {
			t.Fail()
		}
		_, err = MockApplicationStore.GetCreatorApplications(ctx, uid, 0, -1)
		// errors are expected
		if err == nil {
			t.Fail()
		}
	})
}

func TestGetCampaignApplications(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	uid := "dummy_user_01"
	bid := "dummy_brand_01"

	// create a dummy creator
	generateCreator(ctx, uid)

	// create a dummy brand
	generateBrand(bid)

	campIDs := SeedCampaign(ctx, bid, ActiveStatus, 1)
	applIDs := SeedApplications(ctx, campIDs, uid)

	defer func() {
		destroyApplications(ctx, applIDs)
		destroyCampaign(ctx, campIDs)
		destroyCreator(ctx, uid)
		destroyBrand(bid)
		cancel()
	}()
	t.Run("OK", func(t *testing.T) {
		appls, err := MockApplicationStore.GetCampaignApplications(ctx, campIDs[0], 0, 10)
		// errors are not expected
		if err != nil {
			t.Fail()
		}
		if len(appls) != 1 {
			log.Printf("wanted: %d, got: %d\n", 1, len(appls))
			t.Fail()
		}
	})
	t.Run("invalid campaign_id", func(t *testing.T) {
		// invalids campaign id requested
		_, err := MockApplicationStore.GetCampaignApplications(ctx, "", 0, 10)
		// errors are expected
		if err == nil {
			t.Fail()
		}
	})
	t.Run("negative offset/limit", func(t *testing.T) {
		_, err := MockApplicationStore.GetCampaignApplications(ctx, campIDs[0], -1, 10)
		// errors are expected
		if err == nil {
			t.Fail()
		}
		_, err = MockApplicationStore.GetCampaignApplications(ctx, campIDs[0], 0, -1)
		// errors are expected
		if err == nil {
			t.Fail()
		}
	})
}

func TestDestroyApplications(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	// create dummy uid and bid
	uid := "dummy_user_01"
	bid := "dummy_brand_01"

	// generate the dummy user
	generateCreator(ctx, uid)
	// generate the dummy brand
	generateBrand(bid)
	campIDs := SeedCampaign(ctx, bid, ActiveStatus, 3)
	applIDs := SeedApplications(ctx, campIDs, uid)
	// clean up before exit
	defer func() {
		destroyApplications(ctx, applIDs)
		destroyCampaign(ctx, campIDs)
		destroyBrand(bid)
		destroyCreator(ctx, uid)
		cancel()
	}()

	t.Run("OK", func(t *testing.T) {
		test_id := applIDs[0]
		err := MockApplicationStore.DeleteApplication(ctx, test_id)
		// expecting no error
		if err != nil {
			t.Fail()
		}
	})
	t.Run("Invalid application id", func(t *testing.T) {
		err := MockApplicationStore.DeleteApplication(ctx, "NA")
		// expecting error
		if err == nil {
			log.Printf("Got no error\n")
			t.Fail()
		}
	})
}

func TestGetApplication(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	// generate a dummy creator and brand
	uid := "dummy_user_01"
	bid := "dummy_brand_01"

	generateCreator(ctx, uid)
	generateBrand(bid)

	// generate the campaigns
	campIDs := SeedCampaign(ctx, bid, ActiveStatus, 1)
	appl_id := SeedApplications(ctx, campIDs, uid)
	//clean up the entries
	defer func() {
		destroyApplications(ctx, appl_id)
		destroyBrand(bid)
		destroyCreator(ctx, uid)
		cancel()
	}()

	t.Run("OK", func(t *testing.T) {
		_, err := MockApplicationStore.GetApplicationByID(ctx, appl_id[0])
		// not expecting errors
		if err != nil {
			t.Fail()
		}
	})
	t.Run("invalid application id", func(t *testing.T) {
		// invalid id
		_, err := MockApplicationStore.GetApplicationByID(ctx, "")
		// expecting an error
		if err == nil {
			t.Fail()
		}
	})
}

func TestSetStatus(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	// generate a dummy creator and brand
	uid := "dummy_user_01"
	bid := "dummy_brand_01"

	generateCreator(ctx, uid)
	generateBrand(bid)

	// generate the campaigns
	campIDs := SeedCampaign(ctx, bid, ActiveStatus, 1)
	appl_id := SeedApplications(ctx, campIDs, uid)
	//clean up the entries
	defer func() {
		destroyApplications(ctx, appl_id)
		destroyBrand(bid)
		destroyCreator(ctx, uid)
		cancel()
	}()

	t.Run("OK", func(t *testing.T) {
		err := MockApplicationStore.SetApplicationStatus(ctx, appl_id[0], ApplicationApprove)
		// not expecting
		if err != nil {
			t.Fail()
		}
	})
	t.Run("invalid application id", func(t *testing.T) {
		err := MockApplicationStore.SetApplicationStatus(ctx, "", ApplicationApprove)
		// expecting error
		if err == nil {
			t.Fail()
		}
	})
	t.Run("invalid status", func(t *testing.T) {
		// 3 is an Invalid status code
		err := MockApplicationStore.SetApplicationStatus(ctx, appl_id[0], 3)
		// expecting error
		if err == nil {
			t.Fail()
		}
	})
}
