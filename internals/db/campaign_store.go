package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
)

type CampaignStore struct {
	db *sql.DB
}

type Campaign struct {
	Id      string  `json:"id"`
	BrandId string  `json:"brand_id"`
	Title   string  `json:"title"`
	Budget  float64 `json:"budget"`
	CPM     float64 `json:"cpm"`
	Req     string  `json:"requirements"`
	// added this to segregate the campaigns on the basis of platform
	Platform  string `json:"platform"`
	DocLink   string `json:"doc_link"`
	Status    int    `json:"status"`
	CreatedAt string `json:"created_at"`
}

// Update Campaign payload
type UpdateCampaign struct {
	// No option to update CPM to avoid frauds
	Title   *string  `json:"title"`
	Budget  *float64 `json:"budget"`
	Req     *string  `json:"requirements"`
	DocLink *string  `json:"doc_link"`
}

// This function adds a new campaign record
func (c *CampaignStore) LaunchCampaign(ctx context.Context, campaign *Campaign) error {
	query := `
		INSERT INTO campaigns (id, brand_id, title, budget, cpm, requirements, platform, doc_link, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		log.Printf("Error Beginning transaction: %v\n", err.Error())
		return err
	}
	defer tx.Rollback()
	_, err = tx.ExecContext(ctx, query,
		campaign.Id,
		campaign.BrandId,
		campaign.Title,
		campaign.Budget,
		campaign.CPM,
		campaign.Req,
		campaign.Platform,
		campaign.DocLink,
		campaign.Status,
	)
	if err != nil {
		log.Printf("Error launching new campaign: %v\n", err.Error())
		return err
	}
	// successfully launched a new campaign
	update_brand_query := `
		UPDATE brands
		SET campaigns = campaigns + 1
		WHERE id = $1
	`
	_, err = tx.ExecContext(ctx, update_brand_query, campaign.BrandId)
	if err != nil {
		log.Printf("Error updating brand's campaign count: %v\n", err.Error())
		return err
	}
	tx.Commit()
	return nil
}

// This function Ends a campaign
func (c *CampaignStore) EndCampaign(ctx context.Context, id string) error {
	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		log.Printf("Error initialising transaction: %v\n", err.Error())
		return err
	}
	defer tx.Rollback()
	// Iniate a payout bill to the creators associated to the campaign
	// TODO: API layer

	query := `
		UPDATE campaigns
		SET status = $1
		WHERE id = $2
	`
	_, err = tx.ExecContext(ctx, query, ExpiredStatus, id)
	if err != nil {
		log.Printf("Error occured while closing campaign(%s): %v\n", id, err.Error())
		return err
	}

	// successfully ended the campaign
	return nil
}

// This functions updates a specific campaign details
func (c *CampaignStore) UpdateCampaign(ctx context.Context, campaign_id string, payload UpdateCampaign) error {
	var queryBuilder strings.Builder
	queryBuilder.WriteString("UPDATE campaigns SET ")
	var expressions []string // query parameters
	var args []any           // arguments

	i := 1
	if payload.Title != nil {
		expressions = append(expressions, fmt.Sprintf("title = $%d", i))
		args = append(args, *payload.Title)
		i++
	}
	if payload.Budget != nil {
		expressions = append(expressions, fmt.Sprintf("budget = $%d", i))
		args = append(args, *payload.Budget)
		i++
	}
	if payload.Req != nil {
		expressions = append(expressions, fmt.Sprintf("requirements = $%d", i))
		args = append(args, *payload.Req)
		i++
	}
	if payload.DocLink != nil {
		expressions = append(expressions, fmt.Sprintf("doc_link = $%d", i))
		args = append(args, *payload.DocLink)
		i++
	}
	queryBuilder.WriteString(strings.Join(expressions, ", "))
	queryBuilder.WriteString(fmt.Sprintf(" WHERE id = $%d", i))
	args = append(args, campaign_id)
	query := queryBuilder.String()
	_, err := c.db.ExecContext(ctx, query, args...)

	if err != nil {
		log.Printf("Error updating campaign details: %v", err.Error())
		return err
	}

	// successfully updated details
	return nil
}

func (c *CampaignStore) GetRecentCampaigns(ctx context.Context, offset int, limit int) ([]Campaign, error) {
	var output []Campaign
	query := `
		SELECT id, brand_id, title, budget, cpm, requirements, platform, doc_link, status, created_at
		FROM campaigns
		ORDER BY created_at DESC, id DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := c.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		log.Printf("Error fetching campaigns: %v\n", err.Error())
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var row Campaign
		err = rows.Scan(
			&row.Id,
			&row.BrandId,
			&row.Title,
			&row.Budget,
			&row.CPM,
			&row.Req,
			&row.Platform,
			&row.DocLink,
			&row.Status,
			&row.CreatedAt,
		)
		if err != nil {
			log.Printf("Error scanning campaign: %v\n", err.Error())
			return nil, err
		}
		output = append(output, row)
	}
	// return the fetched the feed
	return output, nil
}

func (c *CampaignStore) GetBrandCampaigns(ctx context.Context, brandid string) ([]Campaign, error) {
	var output []Campaign
	query := `
		SELECT id, title, budget, cpm, requirements, platform, doc_link, status, created_at
		FROM campaigns
		WHERE brand_id = $1
		ORDER BY created_at DESC, id DESC
	`
	rows, err := c.db.QueryContext(ctx, query, brandid)
	if err != nil {
		log.Printf("Error fetching campaigns: %v\n", err.Error())
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var row Campaign
		err = rows.Scan(
			&row.Id,
			&row.Title,
			&row.Budget,
			&row.CPM,
			&row.Req,
			&row.Platform,
			&row.DocLink,
			&row.Status,
			&row.CreatedAt,
		)
		if err != nil {
			log.Printf("Error scanning campaign: %v\n", err.Error())
			return nil, err
		}
		output = append(output, row)
	}
	// return the fetched the feed
	return output, nil
}

func (c *CampaignStore) GetUserCampaigns(ctx context.Context, userid string) ([]Campaign, error) {
	var output []Campaign
	query := `
		SELECT id, brand_id, title, budget, cpm, requirements, platform, doc_link, status, created_at
		FROM campaigns
		WHERE id = (SELECT campaign_id
			FROM submissions
			WHERE creator_id = $1
		)
		ORDER BY created_at DESC, id DESC
	`
	rows, err := c.db.QueryContext(ctx, query, userid)
	if err != nil {
		log.Printf("Error fetching campaigns: %v\n", err.Error())
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var row Campaign
		err = rows.Scan(
			&row.Id,
			&row.BrandId,
			&row.Title,
			&row.Budget,
			&row.CPM,
			&row.Req,
			&row.Platform,
			&row.DocLink,
			&row.Status,
			&row.CreatedAt,
		)
		if err != nil {
			log.Printf("Error scanning campaign: %v\n", err.Error())
			return nil, err
		}
		output = append(output, row)
	}
	// return the fetched the feed
	return output, nil
}

func (c *CampaignStore) GetMultipleCampaigns(ctx context.Context, campaignIDs []string,
) ([]Campaign, error) {
	query := `
		SELECT id, brand_id, title, budget, cpm, requirements, platform, doc_link, status, created_at
		FROM campaigns
		WHERE id IN $1
	`
	rows, err := c.db.QueryContext(ctx, query, campaignIDs)
	if err != nil {
		log.Printf("error getting the campaigns: %s\n", err.Error())
		return nil, err
	}
	defer rows.Close()
	var output []Campaign
	for rows.Next() {
		var row Campaign
		err = rows.Scan(
			&row.Id,
			&row.BrandId,
			&row.Title,
			&row.Budget,
			&row.CPM,
			&row.Req,
			&row.Platform,
			&row.DocLink,
			&row.Status,
			&row.CreatedAt,
		)
		if err != nil {
			log.Printf("Error scanning campaign: %v\n", err.Error())
			return nil, err
		}
		output = append(output, row)
	}

	// success
	return output, nil
}

func (c *CampaignStore) GetCampaign(ctx context.Context, id string) (*Campaign, error) {
	query := `
		SELECT id, brand_id, title, budget, cpm, requirements, platform, doc_link, status, created_at
		FROM campaigns
		WHERE id = $1
	`
	var row Campaign
	err := c.db.QueryRowContext(ctx, query, id).Scan(
		&row.Id,
		&row.BrandId,
		&row.Title,
		&row.Budget,
		&row.CPM,
		&row.Req,
		&row.Platform,
		&row.DocLink,
		&row.Status,
		&row.CreatedAt,
	)
	if err != nil {
		// Error while fetching
		log.Printf("Error fetching camapaign: %v\n", id)
		return nil, err
	}

	// Fetched campaign successfully
	return &row, nil
}

func (c *CampaignStore) DeleteCampaign(ctx context.Context, id string) error {
	query := `
		DELETE FROM campaigns
		WHERE id = $1
	`
	_, err := c.db.ExecContext(ctx, query, id)
	if err != nil {
		log.Printf("Error deleting campaign")
		return err
	}

	// successfully deleted campaign
	return nil
}
