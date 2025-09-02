package db

import (
	"context"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"strings"

	"github.com/google/uuid"
)

type TransactionStore struct {
	db *sql.DB
}

type Account struct {
	Id        string  `json:"id"`
	HolderId  string  `json:"holder_id"`
	Type      string  `json:"type"` // either creator or brand
	Amount    float64 `json:"amount"`
	Active    bool    `json:"active"`
	CreatedAt string  `json:"created_at"`
}

type Transaction struct {
	Id        string  `json:"id"`
	FromId    string  `json:"from_id"`
	ToId      string  `json:"to_id"`
	Amount    float64 `json:"amount"`
	Status    int     `json:"status"`
	Type      string  `json:"type"`
	CretaedAt string  `json:"created_at"`
}

func DecideLock(a, b uuid.UUID) bool {
	for i := range 16 {
		if a[i] < b[i] {
			return true
		}
		if a[i] > b[i] {
			return false
		}
	}
	return false
}

func DecodeUUID(token string) [16]byte {
	// Remove hyphens from the UUID string for easier hex decoding
	// A standard UUID string has 36 characters including hyphens.
	// The raw hex string without hyphens is 32 characters.
	uuidHex := token
	for _, r := range []rune{'-'} {
		uuidHex = strings.ReplaceAll(uuidHex, string(r), "")
	}

	// Decode the hexadecimal string into a byte slice
	byteSlice, err := hex.DecodeString(uuidHex)
	if err != nil {
		log.Fatalf("Error decoding hex string: %v", err)
	}

	// Convert the byte slice to a [16]byte array
	var byteArray [16]byte
	copy(byteArray[:], byteSlice)
	return byteArray
}

// This function executes a transaction and logs the transaction into transactions table
// type is payout(case sensitive)
func (txs *TransactionStore) Payout(ctx context.Context, ts *Transaction) error {
	tx, err := txs.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()
	var debit_balance float64
	creditQuery := `UPDATE accounts SET amount = amount + $1 WHERE id = $2 AND active = $3`
	debitQuery := `
		UPDATE accounts SET amount = amount - $1 WHERE id = $2 AND active = $3
		RETURNING amount
	`
	logQuery := `
		INSERT INTO transactions (id, from_id, to_id, amount, status, type)
	    VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (id) DO NOTHING
	`
	FromAcc := DecodeUUID(ts.FromId)
	ToAcc := DecodeUUID(ts.ToId)
	if DecideLock(FromAcc, ToAcc) {
		// credit
		if _, err := tx.ExecContext(ctx, creditQuery, ts.Amount, ts.ToId, true); err != nil {
			ts.Status = FailedTxStatus
			_, _ = tx.ExecContext(ctx, logQuery, ts.Id, ts.FromId, ts.ToId, ts.Amount, ts.Status, ts.Type)
			return fmt.Errorf("credit failed: %w", err)
		}

		// debit
		if err := tx.QueryRowContext(ctx, debitQuery, ts.Amount, ts.FromId, true).Scan(&debit_balance); err != nil {
			ts.Status = FailedTxStatus
			_, _ = tx.ExecContext(ctx, logQuery, ts.Id, ts.FromId, ts.ToId, ts.Amount, ts.Status, ts.Type)
			return fmt.Errorf("debit failed: %w", err)
		}
	} else {
		// debit
		if err := tx.QueryRowContext(ctx, debitQuery, ts.Amount, ts.FromId, true).Scan(&debit_balance); err != nil {
			ts.Status = FailedTxStatus
			_, _ = tx.ExecContext(ctx, logQuery, ts.Id, ts.FromId, ts.ToId, ts.Amount, ts.Status, ts.Type)
			return fmt.Errorf("debit failed: %w", err)
		}
		// credit
		if _, err := tx.ExecContext(ctx, creditQuery, ts.Amount, ts.ToId, true); err != nil {
			ts.Status = FailedTxStatus
			_, _ = tx.ExecContext(ctx, logQuery, ts.Id, ts.FromId, ts.ToId, ts.Amount, ts.Status, ts.Type)
			return fmt.Errorf("credit failed: %w", err)
		}
	}

	// success
	ts.Status = SuccessTxStatus
	if _, err := tx.ExecContext(ctx, logQuery, ts.Id, ts.FromId, ts.ToId, ts.Amount, ts.Status, ts.Type); err != nil {
		return fmt.Errorf("log transaction failed: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit failed: %w", err)
	}
	return nil
}

// (only allow brands to deposit)
// deposit function should be given the same from_id and to_id
// type is deposit (case sensitive)
func (txs *TransactionStore) Deposit(ctx context.Context, ts *Transaction) error {
	tx, err := txs.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	creditQuery := `UPDATE accounts SET amount = amount + $1 WHERE id = $2 AND active = $3`
	logQuery := `
		INSERT INTO transactions (id, from_id, to_id, amount, status, type)
	    VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (id) DO NOTHING
	`

	// credit
	if _, err := tx.ExecContext(ctx, creditQuery, ts.Amount, ts.ToId, true); err != nil {
		ts.Status = FailedTxStatus
		_, _ = tx.ExecContext(ctx, logQuery, ts.Id, ts.FromId, ts.ToId, ts.Amount, ts.Status, ts.Type)
		return fmt.Errorf("credit failed: %w", err)
	}
	// success
	log.Printf("credit done\n")
	ts.Status = SuccessTxStatus
	if _, err := tx.ExecContext(ctx, logQuery, ts.Id, ts.FromId, ts.ToId, ts.Amount, ts.Status, ts.Type); err != nil {
		return fmt.Errorf("log transaction failed: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit failed: %w", err)
	}
	return nil
}

// withdraw function should be given the same from_id and to_id
// type is withdraw (case sensitive)
func (txs *TransactionStore) Withdraw(ctx context.Context, ts *Transaction) error {
	tx, err := txs.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()
	var debit_balance float64
	debitQuery := `
		UPDATE accounts SET amount = amount - $1 WHERE id = $2 AND active = $3
		RETURNING amount
	`
	logQuery := `
		INSERT INTO transactions (id, from_id, to_id, amount, status, type)
	    VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (id) DO NOTHING
	`
	// debit
	if err := tx.QueryRowContext(ctx, debitQuery, ts.Amount, ts.FromId, true).Scan(&debit_balance); err != nil {
		ts.Status = FailedTxStatus
		_, _ = tx.ExecContext(ctx, logQuery, ts.Id, ts.FromId, ts.ToId, ts.Amount, ts.Status, ts.Type)
		return fmt.Errorf("debit failed: %w", err)
	}
	// success
	ts.Status = SuccessTxStatus
	if _, err := tx.ExecContext(ctx, logQuery, ts.Id, ts.FromId, ts.ToId, ts.Amount, ts.Status, ts.Type); err != nil {
		return fmt.Errorf("log transaction failed: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit failed: %w", err)
	}
	return nil
}

// ------- Accounts Functions -----------

func (ts *TransactionStore) OpenAccount(ctx context.Context, acc *Account) error {
	// Only a single request to open an account will be accepted
	// so row level isolation is sufficient
	// Check If the person exists
	creatorCheck := `
		SELECT COUNT(*) FROM users
		WHERE id = $1
	`
	brandCheck := `
		SELECT COUNT(*) FROM brands
		WHERE id = $1
	`
	var (
		ucount int
		bcount int
	)
	err := ts.db.QueryRowContext(ctx, creatorCheck, acc.HolderId).Scan(&ucount)
	if err != nil {
		log.Println(err.Error())
		return fmt.Errorf("user check failed. try again")
	}
	err = ts.db.QueryRowContext(ctx, brandCheck, acc.HolderId).Scan(&bcount)
	if err != nil {
		log.Println(err.Error())
		return fmt.Errorf("brand check failed. try again")
	}
	if ucount != 1 && bcount != 1 {
		return fmt.Errorf("invalid credentials entered")
	}

	query := `
		INSERT INTO accounts (id, holder_id, holder_type, amount)
		VALUES ($1, $2, $3, $4)
	`
	_, err = ts.db.ExecContext(ctx, query,
		acc.Id,
		acc.HolderId,
		acc.Type,
		acc.Amount,
	)
	if err != nil {
		// user exists but account creation error
		log.Printf("error opening account for: %s\n", acc.HolderId)
		log.Println(err.Error())
		return err
	}

	// successfully created account
	return nil
}

func (ts *TransactionStore) DisableAccount(ctx context.Context, id string) error {
	query := `
		UPDATE accounts SET active = $1
		WHERE id = $2
	`
	_, err := ts.db.ExecContext(ctx, query, false, id)
	if err != nil {
		log.Printf("error disabling account: %s", id)
		return err
	}

	return nil
}

func (ts *TransactionStore) DeleteAccount(ctx context.Context, id string) error {
	query := `
		DELETE FROM accounts WHERE id = $1
	`
	_, err := ts.db.ExecContext(ctx, query, id)
	if err != nil {
		log.Printf("error deleting account: %s", id)
		return err
	}

	return nil
}

func (ts *TransactionStore) GetAccount(ctx context.Context, id string) (*Account, error) {
	query := `
		SELECT id, holder_id, holder_type, amount, created_at 
		FROM accounts
		WHERE id = $1
	`
	// fetching the account details
	var acc Account
	err := ts.db.QueryRowContext(ctx, query, id).Scan(
		&acc.Id,
		&acc.HolderId,
		&acc.Type,
		&acc.Amount,
		&acc.CreatedAt,
	)
	if err != nil {
		// invalid account credentials
		log.Printf("invalid credentials for account: %v", err.Error())
		return nil, err
	}

	// fetched account
	return &acc, nil
}
