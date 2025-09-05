package db

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func init() {
	MockUserStore = UserStore{
		db: MockDB,
	}
	MockLinkStore.db = MockDB
}

func TestGetUsers(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	user := User{
		Id:            "0001",
		FirstName:     "Croc",
		LastName:      "Singh",
		Email:         "sitanshu5@gmail.com",
		Gender:        "M",
		Age:           20,
		Role:          "LVL1",
		PlatformLinks: []Links{},
	}
	if err := user.Password.Hash("random_pass"); err != nil {
		log.Printf("Hashing error: %v\n", err.Error())
		t.Fail()
	}
	if err := MockUserStore.CreateUser(ctx, &user); err != nil {
		t.Fail()
	}
	t.Run("Getting user by id", func(t *testing.T) {
		_, err := MockUserStore.GetUserById(ctx, "0001") // This is a given id that will always exist
		if err != nil {
			t.Fail()
		}
	})
	t.Run("Getting user by mail", func(t *testing.T) {
		_, err := MockUserStore.GetUserByEmail(ctx, "sitanshu5@gmail.com") // This is a given mail that will always exist
		if err != nil {
			t.Fail()
		}
	})
	t.Run("Fetching Invalid user id", func(t *testing.T) {
		_, err := MockUserStore.GetUserById(ctx, "NA") // Some random id
		if err == nil {
			t.Fail()
		}
		_, err = MockUserStore.GetUserById(ctx, "")
		if err == nil {
			// We expect an occur to occur
			t.Fail()
		}
		_, err = MockUserStore.GetUserByEmail(ctx, "")
		if err == nil {
			// We expect an occur to occur
			t.Fail()
		}
	})
	MockUserStore.DeleteUser(ctx, user.Id)
}

func TestCreateUser(t *testing.T) {
	ctx := t.Context()
	t.Run("Creating a valid user", func(t *testing.T) {
		user := User{
			Id:            "0002",
			FirstName:     "Croc",
			LastName:      "Singh",
			Email:         "crocs2@gmail.com",
			Gender:        "M",
			Age:           20,
			Role:          "LVL1",
			PlatformLinks: []Links{},
		}
		if err := user.Password.Hash("random_pass"); err != nil {
			log.Printf("Hashing error: %v\n", err.Error())
			t.Fail()
		}
		if err := MockUserStore.CreateUser(ctx, &user); err != nil {
			t.Fail()
		}
		MockUserStore.DeleteUser(ctx, user.Id)
	})
	t.Run("Checking age range", func(t *testing.T) {
		user := User{
			Id:            "0003",
			FirstName:     "Croc",
			LastName:      "Singh",
			Email:         "crocs3@gmail.com",
			Gender:        "M",
			Age:           0,
			Role:          "LVL1",
			PlatformLinks: []Links{},
		}
		if err := user.Password.Hash("random_pass"); err != nil {
			t.Fail()
		}
		if err := MockUserStore.CreateUser(ctx, &user); err == nil {
			// We expect an occur to occur
			t.Fail()
		}
	})
	t.Run("Checking age range", func(t *testing.T) {
		user := User{
			Id:            "0004",
			FirstName:     "Croc",
			LastName:      "Singh",
			Email:         "crocs4@gmail.com",
			Gender:        "M",
			Age:           100,
			Role:          "LVL1",
			PlatformLinks: []Links{},
		}
		if err := user.Password.Hash("random_pass"); err != nil {
			log.Printf("Hashing error: %v\n", err.Error())
			t.Fail()
		}
		if err := MockUserStore.CreateUser(ctx, &user); err == nil {
			// We expect an occur to occur
			t.Fail()
		}
	})
}

func TestUpdateUser(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	user := User{
		Id:            "0002",
		FirstName:     "Croc",
		LastName:      "Singh",
		Email:         "crocs2@gmail.com",
		Gender:        "M",
		Age:           20,
		Role:          "LVL1",
		PlatformLinks: []Links{},
	}
	if err := user.Password.Hash("random_pass"); err != nil {
		log.Printf("Hashing error: %v\n", err.Error())
		t.Fail()
	}
	if err := MockUserStore.CreateUser(ctx, &user); err != nil {
		t.Fail()
	}
	NewFName := "NewFName"
	NewLName := "NewLName"
	NewEmail := "NewMail@gamil.com"
	NewGender := "O"
	NewPass := "newpassword"
	t.Run("updating user", func(t *testing.T) {
		payload := UpdatePayload{
			Id:        user.Id,
			FirstName: &NewFName,
		}
		// Only updating FirstName
		err := MockUserStore.UpdateUser(ctx, payload)
		if err != nil {
			t.Fail()
		}
		// Only updating LastName
		payload.FirstName = nil
		payload.LastName = &NewLName
		err = MockUserStore.UpdateUser(ctx, payload)
		if err != nil {
			t.Fail()
		}
		// Only updating Email
		payload.LastName = nil
		payload.Email = &NewEmail
		err = MockUserStore.UpdateUser(ctx, payload)
		if err != nil {
			t.Fail()
		}
		// Only updating Gender
		payload.Email = nil
		payload.Gender = &NewGender
		err = MockUserStore.UpdateUser(ctx, payload)
		if err != nil {
			t.Fail()
		}
		// Only updating Password
		payload.NewPass = &NewPass
		err = MockUserStore.UpdateUser(ctx, payload)
		if err != nil {
			t.Fail()
		}
		updatedUser, _ := MockUserStore.GetUserById(ctx, user.Id)
		if updatedUser.FirstName != NewFName ||
			updatedUser.LastName != NewLName ||
			updatedUser.Gender != NewGender ||
			updatedUser.Email != NewEmail {
			t.Fail()
		}
		gotPass := bcrypt.CompareHashAndPassword(updatedUser.Password.hashed_pass, []byte(NewPass))
		if gotPass != nil {
			t.Fail()
		}
	})
	NewGender = "N" // Not Possible
	t.Run("Invalid Gender type", func(t *testing.T) {
		payload := UpdatePayload{
			Id:     user.Id,
			Gender: &NewGender,
		}

		err := MockUserStore.UpdateUser(ctx, payload)
		if err == nil {
			t.Fail()
		}
	})
	t.Run("No fields to update", func(t *testing.T) {
		payload := UpdatePayload{
			Id: user.Id,
		}

		err := MockUserStore.UpdateUser(ctx, payload)
		if err == nil {
			t.Fail()
		}
	})
	MockUserStore.DeleteUser(ctx, user.Id)
}

func TestDeleteUsers(t *testing.T) {
	ctx := t.Context()
	t.Run("Delete a valid user", func(t *testing.T) {
		// Creating the user to be deleted
		user := User{
			Id:            "0002",
			FirstName:     "Croc",
			LastName:      "Singh",
			Email:         "crocs2@gmail.com",
			Gender:        "M",
			Age:           20,
			Role:          "LVL1",
			PlatformLinks: []Links{},
		}
		if err := user.Password.Hash("random_pass"); err != nil {
			log.Printf("Hashing error: %v\n", err.Error())
			t.Fail()
		}
		if err := MockUserStore.CreateUser(ctx, &user); err != nil {
			t.Fail()
		}
		// Deleting the newly created user
		err := MockUserStore.DeleteUser(ctx, "0002")
		if err != nil {
			t.Fail()
		}
	})
	t.Run("Deleting an invalid user", func(t *testing.T) {
		err := MockUserStore.DeleteUser(ctx, "NA") // Impossible to occur in DB
		if err == nil {
			// We expect an error to occur
			t.Fail()
		}
	})
}

// ----------- LinkStore Tests ---------------

func seedLinks(num int) []Links {
	var output []Links
	for i := range num {
		link := Links{
			Platform: RandString(7),
			Url:      fmt.Sprintf("example_%d.com", i),
		}
		output = append(output, link)
	}
	return output
}

func destroyLinks(ctx context.Context, uid string, links []Links) {
	for _, v := range links {
		MockLinkStore.DeleteLinks(ctx, uid, v.Platform)
	}
}

func TestAddLink(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	// make a mock creator
	uid := generateCreator(ctx, "0001")
	defer destroyCreator(ctx, uid)

	links := seedLinks(10)
	defer destroyLinks(ctx, uid, links)
	t.Run("adding valid links", func(t *testing.T) {
		err := MockLinkStore.AddLinks(ctx, uid, links)
		if err != nil {
			t.Fail()
		}
	})
}

func TestDeleteLinks(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	uid := generateCreator(ctx, "0001")
	defer destroyCreator(ctx, uid)

	links := seedLinks(1)
	defer destroyLinks(ctx, uid, links)

	t.Run("blank input for uid and platform", func(t *testing.T) {
		err := MockLinkStore.DeleteLinks(ctx, "", "")
		if err == nil {
			t.Fail()
		}
	})
	t.Run("with blank uid and valid platform", func(t *testing.T) {
		err := MockLinkStore.DeleteLinks(ctx, "", links[0].Platform)
		if err == nil {
			t.Fail()
		}
	})
	t.Run("empty invalid input for platform parameter", func(t *testing.T) {
		err := MockLinkStore.DeleteLinks(ctx, uid, "")
		if err == nil {
			t.Fail()
		}
	})
	t.Run("deleting valid link", func(t *testing.T) {
		err := MockLinkStore.DeleteLinks(ctx, uid, links[0].Platform)
		if err != nil {
			t.Fail()
		}
	})
}
