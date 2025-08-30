package db

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"time"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

const QueryTimeOut = time.Minute * 3

var (
	ErrNotFound     = errors.New("not found")
	ErrTokenExpired = errors.New("invalid or expired token")
	ErrDupliMail    = errors.New("email already exists")
	ErrDupliName    = errors.New("name taken")
)

// Differentiating the password struct to handle the hashing of plain_pass
type PassW struct {
	text        *string
	hashed_pass []byte
}

// hashing function of password to assign the hashed password
func (p *PassW) Hash(plain_pass string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plain_pass), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	p.hashed_pass = hash
	p.text = &plain_pass

	return nil
}

type Store struct {
	UserInterface interface {
		GetUserById(context.Context, string) (*User, error)
		GetUserByEmail(context.Context, string) (*User, error)
		CreateUser(context.Context, *User) error
		DeleteUser(context.Context, string) error
		UpdateUser(context.Context, UpdatePayload) error

		// TODO: Implement the follow/unfollow brand option(AT LAST)
		// FollowBrand(context.Context, string, string) error
		// UnFollowBrand(context.Context, string, string) error
		// ctx, from_id, to_id, type(withdraw/deposit), amount, tx
		// ExecTransaction(context.Context, string, string, string, float32, sql.Tx) error
	}
	BrandInterface interface {
		GetBrandById(context.Context, string) (*Brand, error)
		GetBrandsByFilter(context.Context, string, any) ([]Brand, error)
		RegisterBrand(context.Context, *Brand) error
		DeregisterBrand(context.Context, string) error
		UpdateBrand(context.Context, BrandUpdatePayload) error
		// ctx, from_id, to_id, type(withdraw/deposit), amount, tx
		// ExecTransaction(context.Context, string, string, string, float32, sql.Tx) error
	}
	CampaignInterace interface {
		LaunchCampaign(context.Context, *Campaign) error
		EndCampaign(context.Context, string) error
		UpdateCampaign(context.Context, UpdateCampaign) error
		GetRecentCampaigns(context.Context, int, int) ([]Campaign, error)
		GetCampaign(context.Context, string) (*Campaign, error)
		//TODO
	}
}

func NewStore(db *sql.DB) *Store {
	return &Store{
		UserInterface: &UserStore{
			db: db,
		},
		BrandInterface: &BrandStore{
			db: db,
		},
		CampaignInterace: &CampaignStore{
			db: db,
		},
	}
}

func Mount(addr string, MaxConns, MaxIdleConn, MaxIdleTime int) (*sql.DB, error) {
	// Open a new Databse Session
	// Using POSTGRES Engine
	db, err := sql.Open("postgres", addr)
	if err != nil {
		// For Debugging
		log.Printf("Could not Open DB: %v\n", err)
		return nil, err
	}
	// Set the Databse Constraints
	db.SetConnMaxIdleTime(time.Minute * time.Duration(MaxIdleTime))
	db.SetMaxOpenConns(MaxConns)
	db.SetMaxIdleConns(MaxIdleConn)
	// Initialise the context
	ctx, cancel := context.WithTimeout(context.Background(), QueryTimeOut)
	defer cancel()

	if err = db.PingContext(ctx); err != nil {
		// For Debugging
		log.Printf("Error Pinging DB: %v\n", err.Error())
		return nil, err
	}

	return db, nil
}

// Implement this ATLAST To modularise
// func Transaction(db *sql.DB, ctx context.Context, f func(*sql.Tx) error) error {
// 	tx, err := db.BeginTx(ctx, nil)
// 	if err != nil {
// 		log.Printf("Error Beginning transaction: %v\n", err.Error())
// 		return err
// 	}
// 	err = f(tx)
// 	if err != nil {
// 		log.Printf("Error during transaction: %v\n", err.Error())
// 		return err
// 	}

// 	return nil
// }
