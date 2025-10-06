package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"
)

type BrandStore struct {
	db *sql.DB
}

// Brand Model
type Brand struct {
	Id        string `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Sector    string `json:"sector"`
	Password  PassW  `json:"-"`
	Website   string `json:"website"`
	Address   string `json:"address"`
	Campaigns int    `json:"campaign_count"`
}

type BrandUpdatePayload struct {
	// I wont allow brands to change their names/sector (Prevention measure for frauds)
	Email   *string `json:"email"`
	Website *string `json:"website"`
	Address *string `json:"address"`
}

// Function to change the password of a brand
// Validate the brandid before calling the function
func (b *BrandStore) ChangePassword(ctx context.Context, id, new_pass string) error {
	var pw PassW
	if len(new_pass) < 8 {
		log.Printf("error: want pass len: %d, got: %d", MinPassLen, len(new_pass))
		return ErrPasswordTooShort
	}
	if err := pw.Hash(new_pass); err != nil {
		// logging for debugging
		log.Printf("error hashing password: %v\n", err.Error())
		return err
	}
	query := `
		UPDATE brands
		SET password = $1
		WHERE id = $2
	`
	res, err := b.db.ExecContext(ctx, query, pw.hashed_pass, id)
	if err != nil {
		log.Printf("error changing password: %v\n", err.Error())
		return err
	}
	if count, _ := res.RowsAffected(); count == 0 {
		log.Printf("invalid id: %v\n", id)
		return ErrInvalidId
	}
	return nil
}

func (b *BrandStore) GetBrandById(ctx context.Context, id string) (*Brand, error) {
	query := `
		SELECT id, name, email, password, sector, website, address, campaigns
		FROM brands
		WHERE id = $1
	`
	var brand Brand
	// Scan the brand fetched
	err := b.db.QueryRowContext(ctx, query, id).Scan(
		&brand.Id,
		&brand.Name,
		&brand.Email,
		&brand.Password.hashed_pass,
		&brand.Sector,
		&brand.Website,
		&brand.Address,
		&brand.Campaigns,
	)
	if err != nil {
		// Loggging error
		log.Printf("Brand fetch error: %v\n", err.Error())
		return nil, err
	}

	return &brand, nil
}

// Filter brands by category [sector, name, campaign_counts] pass "%name%" to filter by name
func (b *BrandStore) GetBrandsByFilter(ctx context.Context, filter_type string, arg any) ([]Brand, error) {
	var result []Brand
	// Filters = [Sector, Campaigns, Name]
	filters := map[string]string{
		"sector":    "WHERE sector = $1",
		"campaigns": "WHERE campaigns >= $1",
		"name":      "WHERE name LIKE $1",
	}
	// check if filter allowed ?
	condition, ok := filters[filter_type]
	if !ok {
		return nil, errors.New("filter not allowed")
	}

	query := "SELECT id, name, email, sector, website, address, campaigns FROM brands " + condition
	rows, err := b.db.QueryContext(ctx, query, arg)
	if err != nil {
		log.Printf("Filtering brand error: %v\n", err.Error())
		return nil, err
	}
	// Safely close the opened rows
	defer rows.Close()
	for rows.Next() {
		var brand Brand
		err = rows.Scan(
			&brand.Id,
			&brand.Name,
			&brand.Email,
			&brand.Sector,
			&brand.Website,
			&brand.Address,
			&brand.Campaigns,
		)
		if err != nil {
			log.Printf("Error while scanning filered brands: %v\n", err.Error())
			return nil, err
		}
		result = append(result, brand)
	}

	// Return the requested brands
	return result, nil
}

// Onboard a brand onto the platform
func (b *BrandStore) RegisterBrand(ctx context.Context, new_brand *Brand) error {
	tx, err := b.db.BeginTx(ctx, nil)
	if err != nil {
		log.Printf("Error Beginning transaction: %v\n", err.Error())
		return err
	}
	query := `
		INSERT INTO brands (id, name, email, sector, password, website, address, campaigns)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err = tx.ExecContext(ctx, query,
		new_brand.Id,
		new_brand.Name,
		new_brand.Email,
		new_brand.Sector,
		new_brand.Password.hashed_pass,
		new_brand.Website,
		new_brand.Address,
		new_brand.Campaigns,
	)
	if err != nil {
		// Logging to debug and check
		log.Printf("Error onboarding brand: %v\n", err.Error())
		return err
	}

	// CREATE THE BRAND'S ACCOUNT
	// Prompt the brand to create account if it tries to launch a campaign
	// Successfully onboarded brand
	tx.Commit()
	return nil
}

func (b *BrandStore) DeregisterBrand(ctx context.Context, id string) error {
	tx, err := b.db.BeginTx(ctx, nil)
	if err != nil {
		log.Printf("Error Beginning transaction: %v\n", err.Error())
		return err
	}
	defer tx.Rollback()
	// Restrict Brands if they have atleast one active campaign
	var active_campaigns int
	checkQuery := `
		SELECT COUNT(*) FROM campaigns WHERE brand_id = $1 AND status = $2
	`
	err = tx.QueryRowContext(ctx, checkQuery, id, ActiveStatus).Scan(&active_campaigns)
	if err != nil {
		log.Printf("error deleting brand account(%s): %v\n", id, err.Error())
		return err
	}
	if active_campaigns > 0 {
		log.Printf("restricted deactivation, open campaigns found")
		return fmt.Errorf("restricted deactivation, open campaigns found")
	}
	query := `
		DELETE FROM brands
		WHERE id = $1
	`
	_, err = tx.ExecContext(ctx, query, id)
	if err != nil {
		log.Printf("error deleting brand account(%s): %v\n", id, err.Error())
		return err
	}

	// Delete the brands bank account
	DeleteAccQuery := `
		DELETE FROM accounts
		WHERE holder_id = $1
	`
	_, err = tx.ExecContext(ctx, DeleteAccQuery, id)
	if err != nil {
		log.Printf("error deleting account details of %s: %v\n", id, err.Error())
		return err
	}
	// The campaigns connected to the brands also disappear

	tx.Commit()
	return nil
}

func (b *BrandStore) UpdateBrand(ctx context.Context, brand_id string, payload BrandUpdatePayload) error {
	var queryBuilder strings.Builder
	queryBuilder.WriteString("UPDATE brands SET ")
	var args []any
	var cols []string
	i := 1
	if payload.Email != nil {
		cols = append(cols, fmt.Sprintf("email = $%d", i))
		args = append(args, *payload.Email)
		i++
	}
	if payload.Website != nil {
		cols = append(cols, fmt.Sprintf("website = $%d", i))
		args = append(args, *payload.Website)
		i++
	}
	if payload.Address != nil {
		cols = append(cols, fmt.Sprintf("address = $%d", i))
		args = append(args, *payload.Address)
		i++
	}
	queryBuilder.WriteString(strings.Join(cols, ", "))
	queryBuilder.WriteString(fmt.Sprintf(" WHERE id = $%d", i))
	args = append(args, brand_id)
	query := queryBuilder.String()

	if len(cols) == 0 {
		return errors.New("no fields to update")
	}
	_, err := b.db.ExecContext(ctx, query, args...)
	if err != nil {
		log.Printf("Error updating brand: %v\n", err.Error())
		return err
	}

	// successfully updated brand details
	return nil
}
