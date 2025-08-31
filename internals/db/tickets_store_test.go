package db

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"
)

var MockTicketStore TicketStore

func init() {
	MockTicketStore = TicketStore{
		db: MockDB,
	}
}

func seedTickets(ctx context.Context, num int) []string {
	i := 0
	var ids []string
	tx, _ := MockTicketStore.db.BeginTx(ctx, nil)
	for i < num {
		id := fmt.Sprintf("010%d", i)
		query := `
			INSERT INTO support_tickets (id, customer_id, type, subject, message, status)
			VALUES ($1, $2, $3, $4, $5, $6)
		`
		args := []any{
			id, "0001", "creator", "test_subj", "test_message", OpenTicket,
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

func TestOpenTicket(t *testing.T) {
	ctx := t.Context()
	t.Run("opening a demo ticket", func(t *testing.T) {
		ticket := Ticket{
			Id:         "0001",
			CustomerId: "0001",
			Type:       "brand",
			Subject:    "New Ticket",
			Message:    "mock problem",
		}
		err := MockTicketStore.OpenTicket(ctx, &ticket)
		if err != nil {
			t.Fail()
		}
		MockTicketStore.DeleteTicket(ctx, ticket.Id)
	})
}

func TestCloseTicket(t *testing.T) {
	ctx := t.Context()
	t.Run("resolving a demo ticket", func(t *testing.T) {
		ticket := Ticket{
			Id:         "0001",
			CustomerId: "0001",
			Type:       "brand",
			Subject:    "New Ticket",
			Message:    "mock problem",
		}
		MockTicketStore.OpenTicket(ctx, &ticket)
		err := MockTicketStore.ResolveTicket(ctx, ticket.Id)
		if err != nil {
			t.Fail()
		}
		MockTicketStore.DeleteTicket(ctx, ticket.Id)
	})
	t.Run("closing an invalid ticket", func(t *testing.T) {
		err := MockTicketStore.ResolveTicket(ctx, "NA")
		if err == nil {
			t.Fail()
		}
	})
}

func TestFetchTicket(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	GenIds := seedTickets(ctx, 10)
	// limit = 10, offset = 0
	got, err := MockTicketStore.FetchRecentTickets(ctx, OpenTicket, 10, 0)
	if err != nil {
		log.Printf("Fetching error in tickets feed: %v", err.Error())
		t.Fail()
	}
	if len(got) != 10 {
		log.Printf("got: %d, want %d", len(got), 10)
		t.Fail()
	}

	if len(got) > 0 && got[0].Id != "0109" {
		log.Printf("sorting failed in tickets feed: top: %s", got[0].Id)
		t.Fail()
	}
	for _, v := range GenIds {
		// Clean up the seeding
		MockTicketStore.DeleteTicket(ctx, v)
	}
}

func TestFindTicket(t *testing.T) {
	ctx := t.Context()
	ticket := Ticket{
		Id:         "0001",
		CustomerId: "0001",
		Type:       "brand",
		Subject:    "New Ticket",
		Message:    "mock problem",
	}
	MockTicketStore.OpenTicket(ctx, &ticket)
	t.Run("finding a valid ticket", func(t *testing.T) {
		_, err := MockTicketStore.FindTicket(ctx, ticket.Id)
		if err != nil {
			t.Fail()
		}
	})
	t.Run("finding invalid ticket", func(t *testing.T) {
		_, err := MockTicketStore.FindTicket(ctx, "NA")
		if err == nil {
			t.Fail()
		}
	})
	MockTicketStore.DeleteTicket(ctx, ticket.Id)
}
