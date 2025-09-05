package db

import (
	"crypto/rand"
	"database/sql"
	"log"
	"math/big"

	"github.com/Alter-Sitanshu/campaignHub/env"
)

var (
	MockDB            *sql.DB
	MockUserStore     UserStore
	MockLinkStore     LinkStore
	MockTsStore       TransactionStore
	MockTicketStore   TicketStore
	MockSubStore      SubmissionStore
	MockBrandStore    BrandStore
	MockCampaignStore CampaignStore
)

func Init() {
	envCfg := env.New()
	dsn := envCfg.DBADDR
	log.Println(dsn)
	var err error
	MockDB, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}

	if err := MockDB.Ping(); err != nil {
		log.Fatalf("failed to ping db: %v", err)
	}

	MockBrandStore.db = MockDB
	MockUserStore.db = MockDB
	MockLinkStore.db = MockDB
	MockTsStore.db = MockDB
	MockTicketStore.db = MockDB
	MockSubStore.db = MockDB
	MockCampaignStore.db = MockDB
}

func RandString(size int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	var output string

	for range size {
		// pick a random index
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			log.Printf("error generating random int: %v", err)
			return ""
		}
		output += string(letters[n.Int64()])
	}
	return output
}
