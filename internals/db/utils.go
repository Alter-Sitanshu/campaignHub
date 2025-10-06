package db

import (
	"database/sql"
	"log"

	"github.com/Alter-Sitanshu/campaignHub/env"
)

var (
	MockDB               *sql.DB
	MockUserStore        UserStore
	MockLinkStore        LinkStore
	MockTsStore          TransactionStore
	MockTicketStore      TicketStore
	MockSubStore         SubmissionStore
	MockBrandStore       BrandStore
	MockCampaignStore    CampaignStore
	MockApplicationStore ApplicationStore
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
	MockApplicationStore.db = MockDB
}
