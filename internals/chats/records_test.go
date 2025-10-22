package chats

import (
	"context"
	"fmt"
	"log"
	"sync"
	"testing"
	"time"
)

const (
	QueryTimeOut = time.Second * 5
)

func TestRun(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		newHub := NewHub(MockHubStore, MockCacheService)
		go newHub.Run()
		newHub.Stop()
	})
}

func TestLoadFollowedBrands(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), QueryTimeOut)
	numBrands := 10
	// create a user to follow brands
	creatorID := "mock_user_001"
	GenerateCreator(ctx, creatorID)

	// Seed brands to follow
	bids := SeedBrands(ctx, numBrands)

	// Clean up
	defer func() {
		ClearFollowList(ctx)
		DestroyBrands(ctx, bids)
		DestroyCreator(ctx, creatorID)
		cancel()
	}()

	t.Run("OK", func(t *testing.T) {
		for _, id := range bids {
			err := MockHub.store.FollowBrand(creatorID, id)
			if err != nil {
				log.Printf("failed at creation of follow\n")
				t.Fatal()
			}
		}
		res, err := MockHub.store.LoadFollowedBrands(ctx, creatorID)
		if err != nil {
			t.Fail()
		}
		if len(res) != numBrands {
			log.Printf("want: %d, got: %d", numBrands, len(res))
			t.Fail()
		}
		// Clear the following
		ClearFollowList(ctx)
	})
}

func TestGetConversation(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), QueryTimeOut)

	// Generating creators
	creatorOne := "mock_creator_001"
	creatorTwo := "mock_creator_002"

	GenerateCreator(ctx, creatorOne)
	GenerateCreator(ctx, creatorTwo)

	defer func() {
		DestroyCreator(ctx, creatorOne)
		DestroyCreator(ctx, creatorTwo)
		cancel()
	}()

	t.Run("OK", func(t *testing.T) {
		conv := &Conversation{
			ID:             "mock_conversation_direct",
			ParticipantOne: creatorOne,
			ParticipantTwo: creatorTwo,
			Type:           Direct,
		}
		err := MockHub.store.CreateConversation(ctx, conv)
		if err != nil {
			t.Fail()
		}
		got, err := MockHub.store.GetConversationByID(ctx, conv.ID)
		if err != nil {
			t.Fail()
		}
		if got.ID != conv.ID {
			log.Printf("got: %q, want: %q\n", got.ID, conv.ID)
			t.Fail()
		}
		MockHub.store.DeleteConversation(ctx, conv.ID)
	})
	t.Run("empty conversation id", func(t *testing.T) {
		_, err := MockHub.store.GetConversationByID(ctx, "")
		if err == nil {
			t.Fail()
		}
	})
	t.Run("invalid id", func(t *testing.T) {
		_, err := MockHub.store.GetConversationByID(ctx, "invalid-id")
		if err == nil {
			t.Fail()
		}
	})
}

func TestFollowUnfollow(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), QueryTimeOut)

	// creating a mock creator
	creator := "mock_user_001"
	GenerateCreator(ctx, creator)

	// brands to follow for the user
	bids := SeedBrands(ctx, 10)

	defer func() {
		ClearFollowList(ctx)
		DestroyBrands(ctx, bids)
		DestroyCreator(ctx, creator)
		cancel()
	}()

	t.Run("Follow 10 brands and then unfollow last 5", func(t *testing.T) {
		// follow all the brands first
		before, after := 10, 5
		for i := range bids {
			err := MockHub.store.FollowBrand(creator, bids[i])
			if err != nil {
				log.Printf("failed at creation of follow\n")
				t.Fail()
			}
		}
		listBefore, _ := MockHub.store.LoadFollowedBrands(ctx, creator)
		if len(listBefore) != before {
			log.Printf("followed all brands, got: %d, want: %d\n", len(listBefore), before)
			t.Fail()
		}
		// unfollow the last 5
		for i := 5; i < 10; i++ {
			err := MockHub.store.UnfollowBrand(creator, bids[i])
			if err != nil {
				log.Printf("failed at unfollow\n")
				t.Fail()
			}
		}
		listAfter, _ := MockHub.store.LoadFollowedBrands(ctx, creator)
		if len(listAfter) != after {
			log.Printf("unfollowed last 5 brands, got: %d, want: %d", len(listAfter), after)
			t.Fail()
		}
	})

}

func TestConcurrentMessaging(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), QueryTimeOut)

	// create the creators/users
	creatorOne := "mock_user_001"
	creatorTwo := "mock_user_002"
	GenerateCreator(ctx, creatorOne)
	GenerateCreator(ctx, creatorTwo)

	defer func() {
		DestroyCreator(ctx, creatorOne)
		DestroyCreator(ctx, creatorTwo)
		cancel()
	}()

	t.Run("5 concurrent messages to a single conversation", func(t *testing.T) {
		iters := 5
		conv := &Conversation{
			ID:             "mock_conversation_direct",
			ParticipantOne: creatorOne,
			ParticipantTwo: creatorTwo,
			Type:           Direct,
		}
		// open a conversation
		MockHub.store.CreateConversation(ctx, conv)
		defer func() {
			ClearMessages(ctx)
			MockHub.store.DeleteConversation(ctx, conv.ID)
		}()
		var wg sync.WaitGroup
		for i := range iters {
			wg.Add(1)
			var sender string
			if i%2 == 1 {
				sender = creatorOne
			} else {
				sender = creatorTwo
			}
			go func(sender string) {
				msg := &Message{
					ID:             fmt.Sprintf("%d", i),
					ConversationID: conv.ID,
					SenderID:       sender,
					MessageType:    "txt",
					Content:        []byte("Hello world"),
				}
				MockHub.store.SaveMessage(ctx, msg)
				wg.Done()
			}(sender)
		}
		wg.Wait()
		msg, err := MockHub.store.GetConversationMessages(ctx, conv.ID, 0, 5)
		if err != nil {
			t.Fail()
		}
		if len(msg) != iters {
			log.Printf("message fetch error: got: %d, want:%d\n", len(msg), iters)
			t.Fail()
		}
	})

	t.Run("messaging to invalid conversation id", func(t *testing.T) {
		msg := &Message{
			ID:             "mock_message_001",
			ConversationID: "NA",
			SenderID:       creatorOne,
			MessageType:    "txt",
			Content:        []byte("Hello world"),
		}
		err := MockHub.store.SaveMessage(ctx, msg)
		if err == nil {
			t.Fail()
		}
	})

	t.Run("invalid message type", func(t *testing.T) {
		conv := &Conversation{
			ID:             "mock_conversation_direct",
			ParticipantOne: creatorOne,
			ParticipantTwo: creatorTwo,
			Type:           Direct,
		}
		// open a conversation
		MockHub.store.CreateConversation(ctx, conv)
		defer func() {
			MockHub.store.DeleteConversation(ctx, conv.ID)
		}()
		msg := &Message{
			ID:             "mock_message_001",
			ConversationID: conv.ID,
			SenderID:       creatorOne,
			MessageType:    "nan type",
			Content:        []byte("Hello world"),
		}
		err := MockHub.store.SaveMessage(ctx, msg)
		if err == nil {
			log.Printf("message should have failed\n")
			t.Fail()
		}
	})

	t.Run("invalid content (NULL)", func(t *testing.T) {
		conv := &Conversation{
			ID:             "mock_conversation_direct",
			ParticipantOne: creatorOne,
			ParticipantTwo: creatorTwo,
			Type:           Direct,
		}
		// open a conversation
		MockHub.store.CreateConversation(ctx, conv)
		defer func() {
			MockHub.store.DeleteConversation(ctx, conv.ID)
		}()
		msg := &Message{
			ID:             "mock_message_001",
			ConversationID: conv.ID,
			SenderID:       creatorOne,
			MessageType:    "nan type",
			Content:        nil,
		}
		err := MockHub.store.SaveMessage(ctx, msg)
		if err == nil {
			log.Printf("message should have failed\n")
			t.Fail()
		}
	})
}
