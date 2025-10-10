package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

const QueryTimeOut time.Duration = time.Minute * 3

// macros for transactions
const (
	FailedTxStatus  int = 0
	SuccessTxStatus int = 1
)

// macros for ticket_status
const (
	OpenTicket  int = 1
	CloseTicket int = 0
)

// macros for status
const (
	ActiveStatus  int = 1
	DraftStatus   int = 0
	ExpiredStatus int = 3
)

// macros for db errors
var (
	ErrServer           = errors.New("internal server error")
	ErrNotFound         = errors.New("not found")
	ErrTokenExpired     = errors.New("invalid or expired token")
	ErrDupliMail        = errors.New("email already exists")
	ErrDupliName        = errors.New("name taken")
	ErrInvalidPass      = errors.New("invalid password")
	ErrInvalidId        = errors.New("invalid id")
	ErrInvalidStatus    = errors.New("invalid application status")
	ErrPasswordTooShort = fmt.Errorf("password should be minimum of length  %d", MinPassLen)
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

func (p *PassW) Compare(plain_pass string) error {
	return bcrypt.CompareHashAndPassword(p.hashed_pass, []byte(plain_pass))
}

type AuthenticatedEntity interface {
	GetID() string
	GetEntityType() EntityType
	GetEmail() string
	GetRole() string
}

type EntityType string

const (
	EntityTypeUser  EntityType = "user"
	EntityTypeBrand EntityType = "brand"
)

type Store struct {
	UserInterface interface {
		GetUserById(context.Context, string) (*User, error)
		GetUserByEmail(context.Context, string) (*User, error)
		CreateUser(context.Context, *User) error
		DeleteUser(context.Context, string) error
		UpdateUser(context.Context, string, UpdatePayload) error
		VerifyUser(ctx context.Context, entity, id string) error
		ChangePassword(ctx context.Context, id, new_pass string) error

		// TODO: Implement the follow/unfollow brand option(AT LAST)
		// FollowBrand(context.Context, string, string) error
		// UnFollowBrand(context.Context, string, string) error
		// ctx, from_id, to_id, type(withdraw/deposit), amount, tx
	}
	BrandInterface interface {
		GetBrandById(context.Context, string) (*Brand, error)
		GetBrandsByFilter(context.Context, string, any) ([]Brand, error)
		RegisterBrand(context.Context, *Brand) error
		DeregisterBrand(context.Context, string) error
		UpdateBrand(context.Context, string, BrandUpdatePayload) error
		ChangePassword(ctx context.Context, id, new_pass string) error
		// ctx, from_id, to_id, type(withdraw/deposit), amount, tx
		// ExecTransaction(context.Context, string, string, string, float32, sql.Tx) error
	}
	CampaignInterace interface {
		LaunchCampaign(context.Context, *Campaign) error
		EndCampaign(context.Context, string) error
		UpdateCampaign(context.Context, string, UpdateCampaign) error
		GetRecentCampaigns(context.Context, int, int) ([]Campaign, error)
		GetBrandCampaigns(context.Context, string) ([]Campaign, error)
		GetUserCampaigns(context.Context, string) ([]Campaign, error)
		GetCampaign(context.Context, string) (*Campaign, error)
		DeleteCampaign(context.Context, string) error
		GetMultipleCampaigns(ctx context.Context, campaignIDs []string) ([]Campaign, error)
	}
	TicketInterface interface {
		OpenTicket(context.Context, *Ticket) error
		ResolveTicket(context.Context, string) error
		FindTicket(context.Context, string) (*Ticket, error)
		DeleteTicket(context.Context, string) error
		GetRecentTickets(context.Context, int, int, int) ([]Ticket, error)
	}
	SubmissionInterface interface {
		MakeSubmission(context.Context, Submission) error
		DeleteSubmission(context.Context, string) error
		FindSubmissionById(context.Context, string) (*Submission, error)
		FindSubmissionsByFilters(context.Context, Filter, int, int) ([]Submission, error)
		FindMySubmissions(ctx context.Context, time_ string, subids []string, limit, offset int) ([]Submission, error)
		UpdateSubmission(context.Context, UpdateSubmission) error
		ChangeViews(ctx context.Context, delta int, id string) error
	}
	LinkInterface interface {
		AddLinks(context.Context, string, []Links) error
		DeleteLinks(context.Context, string, string) error
		GetLinks(context.Context, string) []Links
	}
	TransactionInterface interface {
		Payout(context.Context, *Transaction) error
		Deposit(context.Context, *Transaction) error
		Withdraw(context.Context, *Transaction) error
		OpenAccount(context.Context, *Account) error
		DisableAccount(context.Context, string) error
		DeleteAccount(context.Context, string) error
		GetAccount(context.Context, string) (*Account, error)
		GetAllAccounts(context.Context, int, int) ([]Account, error)
	}
	ApplicationInterface interface {
		GetApplicationByID(ctx context.Context, appl_id string) (ApplicationResponse, error)
		GetCreatorApplications(ctx context.Context, creator_id string, offset, limit int) ([]ApplicationResponse, error)
		GetCampaignApplications(ctx context.Context, campaign_id string, offset, limit int) ([]ApplicationResponse, error)
		CreateApplication(ctx context.Context, appl CampaignApplication) error
		SetApplicationStatus(ctx context.Context, appl_id string, status int) error
		DeleteApplication(ctx context.Context, appl_id string) error
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
		TicketInterface: &TicketStore{
			db: db,
		},
		SubmissionInterface: &SubmissionStore{
			db: db,
		},
		LinkInterface: &LinkStore{
			db: db,
		},
		TransactionInterface: &TransactionStore{
			db: db,
		},
		ApplicationInterface: &ApplicationStore{
			db: db,
		},
	}
}

func (s *Store) GetEntity(ctx context.Context, id string) (AuthenticatedEntity, error) {
	u, err := s.UserInterface.GetUserById(ctx, id)
	if err == nil {
		return u, nil
	}
	b, err := s.BrandInterface.GetBrandById(ctx, id)
	if err == nil {
		return b, nil
	}
	return nil, ErrInvalidId
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

// Authorised Entity implementation
func (user User) GetID() string {
	return user.Id
}

func (user User) GetEntityType() EntityType {
	return EntityTypeUser
}

func (user User) GetEmail() string {
	return user.Email
}

func (user User) GetRole() string {
	return user.Role
}

func (brand Brand) GetID() string {
	return brand.Id
}

func (brand Brand) GetEntityType() EntityType {
	return EntityTypeBrand
}

func (brand Brand) GetEmail() string {
	return brand.Email
}

func (brand Brand) GetRole() string {
	return "brand"
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
