package chats

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/Alter-Sitanshu/campaignHub/env"
	"github.com/Alter-Sitanshu/campaignHub/internals/cache"
)

var MockHub *Hub
var MockHubStore *sql.DB
var MockCacheService *cache.Service

func Init() {
	envCfg := env.New()
	dsn := envCfg.DBADDR
	log.Println(dsn)
	var err error
	MockHubStore, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}

	if err := MockHubStore.Ping(); err != nil {
		log.Fatalf("failed to ping db: %v", err)
	}

	MockHub = NewHub(MockHubStore, MockCacheService)
}

func GenerateCreator(ctx context.Context, mockUserId string) (string, error) {
	query := `
		INSERT INTO users (id, first_name, last_name, email, password, gender, age, role)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	args := []any{
		mockUserId, fmt.Sprintf("mock_name_%s", mockUserId), "mock_last_name",
		fmt.Sprintf("%s_email@gmail.com", mockUserId),
		"password", "O", 20, "LVL1",
	}
	if _, err := MockHubStore.ExecContext(ctx, query, args...); err != nil {
		return "", err
	}
	return mockUserId, nil
}

func DestroyCreator(ctx context.Context, mockUserId string) {
	query := `
		DELETE FROM users
		WHERE id = $1
	`
	MockHubStore.ExecContext(ctx, query, mockUserId)
}

func SeedBrands(ctx context.Context, num int) []string {
	i := 0
	var ids []string
	tx, _ := MockHubStore.BeginTx(ctx, nil)
	defer tx.Rollback()
	for i < num {
		id := fmt.Sprintf("010%d", i)
		query := `
			INSERT INTO brands (id, name, email, sector, password, website, address, campaigns)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`
		args := []any{
			id, "MockBrand", fmt.Sprintf("mockbrand%d@gmail.com", i), "skin_care", "random_pass", "mockbrand.com",
			"Guwahati", 0,
		}
		_, err := tx.ExecContext(ctx, query, args...)
		if err != nil {
			log.Printf("error seeding brands for following: %s", err.Error())
			return nil
		}
		ids = append(ids, id)
		i++
	}
	tx.Commit()
	return ids
}

func DestroyBrands(ctx context.Context, ids []string) {
	query := `
		DELETE FROM brands
		WHERE id = $1
	`
	for i := range ids {
		_, err := MockHubStore.ExecContext(ctx, query, ids[i])
		if err != nil {
			log.Printf("error destroying brands: %s\n", err.Error())
		}
	}
}

func ClearFollowList(ctx context.Context) {
	query := `
		DELETE FROM following_list
	`
	MockHubStore.ExecContext(ctx, query)
}

func ClearMessages(ctx context.Context) {
	query := `
		DELETE FROM messages
	`
	MockHubStore.ExecContext(ctx, query)
}
