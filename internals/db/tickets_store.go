package db

import (
	"context"
	"database/sql"
	"log"
)

type TicketStore struct {
	db *sql.DB
}

// Support Tickets model
type Ticket struct {
	Id         string `json:"id"`
	CustomerId string `json:"customer_id"` // id of the entity
	Type       string `json:"type"`        // creator or brand whoever raised a ticket
	Subject    string `json:"subject"`
	Message    string `json:"message"`
	Status     int    `json:"status"` // open or resolved ticket
	CreatedAt  string `json:"created_at"`
}

// This function helps locate a support ticket based on id
func (ts *TicketStore) FindTicket(ctx context.Context, id string) (*Ticket, error) {
	query := `
		SELECT id, customer_id, type, subject, message, status, created_at
		FROM support_tickets
		WHERE id = $1
	`
	var output Ticket
	// Scanning the ticket
	err := ts.db.QueryRowContext(ctx, query, id).Scan(
		&output.Id,
		&output.CustomerId,
		&output.Type,
		&output.Subject,
		&output.Message,
		&output.Status,
		&output.CreatedAt,
	)
	if err != nil {
		log.Printf("Error scaning the ticket: %v\n", err.Error())
		return nil, err
	}
	// succesfully retrieved the support ticket
	return &output, nil
}

// This function deletes the support ticket
func (ts *TicketStore) DeleteTicket(ctx context.Context, id string) error {
	query := `
		DELETE FROM support_tickets
		WHERE id = $1
	`
	v, err := ts.db.ExecContext(ctx, query, id)
	if err != nil {
		// internal server error
		log.Printf("error deleting the support ticket: %v\n", err.Error())
		return err
	}
	cnt, _ := v.RowsAffected()
	if cnt == 0 {
		// invalid user input into id
		log.Printf("Invalid ticket id: %s", id)
		return sql.ErrNoRows
	}
	// successfully deleted the support ticket
	return nil
}

func (ts *TicketStore) OpenTicket(ctx context.Context, ticket *Ticket) error {
	query := `
		INSERT INTO support_tickets (id, customer_id, type, subject, message, status)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := ts.db.ExecContext(ctx, query,
		ticket.Id, ticket.CustomerId, ticket.Type, ticket.Subject,
		ticket.Message, OpenTicket,
	)

	if err != nil {
		// error while issuing a new ticket
		log.Printf("Error while issuing new ticket: %v", err.Error())
		return err
	}

	// successfully inserted new ticket
	return nil
}

// This function an opened ticket
func (ts *TicketStore) ResolveTicket(ctx context.Context, id string) error {
	query := `
		UPDATE support_tickets
		SET status = $1
		WHERE id = $2 AND status = $3
	`

	v, err := ts.db.ExecContext(ctx, query, CloseTicket, id, OpenTicket)
	if err != nil {
		// internal server error
		log.Printf("error closing the support ticket: %v\n", err.Error())
		return err
	}
	cnt, _ := v.RowsAffected()
	if cnt == 0 {
		// invalid user input
		log.Printf("ticket closed/invalid: %s", id)
		return sql.ErrNoRows
	}
	// successfully closed the support ticket
	return nil
}

func (ts *TicketStore) FetchRecentTickets(ctx context.Context, status, limit, offset int) ([]Ticket, error) {
	query := `
		SELECT id, customer_id, type, subject, message, status, created_at
		FROM support_tickets
		WHERE status = $1
		ORDER BY created_at DESC, id DESC
		LIMIT $2 OFFSET $3
	`
	var output []Ticket
	rows, err := ts.db.QueryContext(ctx, query, status, limit, offset)
	if err != nil {
		log.Printf("error fetching tickets: %v", err.Error())
		return nil, err
	}
	// do not forget to close the rows object
	defer rows.Close()
	for rows.Next() {
		var t Ticket
		err = rows.Scan(
			&t.Id,
			&t.CustomerId,
			&t.Type,
			&t.Subject,
			&t.Message,
			&t.Status,
			&t.CreatedAt,
		)
		// error while scanning the tickets
		if err != nil {
			log.Printf("error reading tickets: %v", err.Error())
			return nil, err
		}
		output = append(output, t)
	}

	// successfully fetched tickets
	return output, nil
}
