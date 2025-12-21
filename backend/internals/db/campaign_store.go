package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strconv"
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

type CampaignResp struct {
	Id      string  `json:"id"`
	BrandId string  `json:"brand_id,omitempty"`
	Brand   string  `json:"brand,omitempty"`
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
	return tx.Commit()
}

// This function Ends a campaign
func (c *CampaignStore) ActivateCampaign(ctx context.Context, id string) error {
	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		log.Printf("Error initialising transaction: %v\n", err.Error())
		return err
	}
	defer tx.Rollback()

	query := `
		UPDATE campaigns
		SET status = $1
		WHERE id = $2
	`
	_, err = tx.ExecContext(ctx, query, ActiveStatus, id)
	if err != nil {
		log.Printf("Error occured while activating campaign(%s): %v\n", id, err.Error())
		return err
	}

	// successfully activated the campaign
	return tx.Commit()
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

func (c *CampaignStore) GetRecentCampaigns(ctx context.Context, limit int, cursorSeq string,
) ([]CampaignResp, int64, bool, error) {
	var output []CampaignResp
	var nextCursor, prevCursor int64
	// 1. Base Query
	// Using a tuple comparison (row constructor) for speed and correctness: (created_at, seq) < ($2, $3)
	// I added "status = 1" because "Active" campaigns should be on Feed
	query := `
        SELECT c.id, c.brand_id, b.name AS brand_name, c.title, c.budget, c.cpm, 
        c.requirements, c.platform, c.doc_link, c.status, c.created_at, c.seq
        FROM campaigns c
        LEFT JOIN brands b ON c.brand_id = b.id
        WHERE c.status = 1
    `

	var rows *sql.Rows
	var err error

	// 2. Dynamic WHERE Clause
	// If this is the FIRST page (cursor is zero/empty), we don't filter by time.
	if cursorSeq == "" {
		query += ` ORDER BY c.seq DESC LIMIT $1`
		rows, err = c.db.QueryContext(ctx, query, limit+1)
		if err != nil {
			log.Printf("Error fetching campaigns: %v\n", err.Error())
			return nil, 0, false, err
		}
	} else {
		// If we have a cursor, fetch items OLDER than that cursor
		cursor, err := strconv.ParseInt(cursorSeq, 10, 64)
		if err != nil {
			return nil, 0, false, err
		}
		query += `
			AND c.seq < $1 
			ORDER BY  c.seq DESC 
			LIMIT $2
		`
		rows, err = c.db.QueryContext(ctx, query, cursor, limit+1)
		if err != nil {
			log.Printf("Error fetching campaigns: %v\n", err.Error())
			return nil, 0, false, err
		}
	}

	defer rows.Close()

	for rows.Next() {
		var row CampaignResp
		prevCursor = nextCursor
		err = rows.Scan(
			&row.Id,
			&row.BrandId,
			&row.Brand,
			&row.Title,
			&row.Budget,
			&row.CPM,
			&row.Req,
			&row.Platform,
			&row.DocLink,
			&row.Status,
			&row.CreatedAt,
			&nextCursor,
		)
		if err != nil {
			log.Printf("Error scanning campaign: %v\n", err.Error())
			return nil, 0, false, err
		}
		output = append(output, row)
	}
	HasMore := len(output) > limit
	n := min(limit, len(output))
	if HasMore {
		nextCursor = prevCursor
	}
	// return the fetched the feed
	return output[:n], nextCursor, HasMore, nil
}

func (c *CampaignStore) GetBrandCampaigns(ctx context.Context, brandid string, limit int, cursorSeq string,
) ([]CampaignResp, int64, bool, error) {

	var (
		output                 []CampaignResp
		err                    error
		rows                   *sql.Rows
		nextCursor, prevCursor int64
	)
	query := `
		SELECT id, title, budget, cpm, requirements, platform, doc_link, 
		status, created_at, seq
		FROM campaigns
		WHERE brand_id = $1
	`
	// 2. Dynamic WHERE Clause
	// If this is the FIRST page (cursor is zero/empty), we don't filter by time.
	if cursorSeq == "" {
		query += ` ORDER BY seq DESC LIMIT $2`
		rows, err = c.db.QueryContext(ctx, query, brandid, limit+1)
		if err != nil {
			log.Printf("Error fetching campaigns: %v\n", err.Error())
			return nil, 0, false, err
		}
	} else {
		// If we have a cursor, fetch items OLDER than that cursor
		cursor, err := strconv.ParseInt(cursorSeq, 10, 64)
		if err != nil {
			return nil, 0, false, err
		}
		query += `
			AND seq < $2 
			ORDER BY seq DESC 
			LIMIT $3
		`
		rows, err = c.db.QueryContext(ctx, query, brandid, cursor, limit+1)
		if err != nil {
			log.Printf("Error fetching campaigns: %v\n", err.Error())
			return nil, 0, false, err
		}
	}

	defer rows.Close()
	for rows.Next() {
		var row CampaignResp
		prevCursor = nextCursor
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
			&nextCursor,
		)
		if err != nil {
			log.Printf("Error scanning campaign: %v\n", err.Error())
			return nil, 0, false, err
		}
		output = append(output, row)
	}
	// return the fetched the feed
	HasMore := len(output) > limit
	n := min(limit, len(output))
	if HasMore {
		nextCursor = prevCursor
	}
	return output[:n], nextCursor, HasMore, nil
}

func (c *CampaignStore) GetUserCampaigns(ctx context.Context, userid string, limit int, cursorSeq string,
) ([]CampaignResp, int64, bool, error) {

	var (
		output     []CampaignResp
		err        error
		rows       *sql.Rows
		nextCursor int64
		prevCursor int64
	)
	query := `
		SELECT c.id, c.brand_id, b.name AS brand, c.title, c.budget, 
		c.cpm, c.requirements, c.platform, c.doc_link, c.status, c.created_at, c.seq
		FROM campaigns c
		LEFT JOIN brands b ON c.brand_id = b.id
		WHERE c.id IN (
			SELECT campaign_id
			FROM submissions
			WHERE creator_id = $1
		)
	`
	if cursorSeq == "" {
		query += ` ORDER BY c.seq DESC LIMIT $2`
		rows, err = c.db.QueryContext(ctx, query, userid, limit+1)
		if err != nil {
			log.Printf("Error fetching campaigns: %v\n", err.Error())
			return nil, 0, false, err
		}
	} else {
		// If we have a cursor, fetch items OLDER than that cursor
		cursor, err := strconv.ParseInt(cursorSeq, 10, 64)
		if err != nil {
			return nil, 0, false, err
		}
		query += `
			AND c.seq < $2 
			ORDER BY c.seq DESC 
			LIMIT $3
		`
		rows, err = c.db.QueryContext(ctx, query, userid, cursor, limit+1)
		if err != nil {
			log.Printf("Error fetching campaigns: %v\n", err.Error())
			return nil, 0, false, err
		}
	}

	defer rows.Close()
	for rows.Next() {
		var row CampaignResp
		prevCursor = nextCursor
		err = rows.Scan(
			&row.Id,
			&row.BrandId,
			&row.Brand,
			&row.Title,
			&row.Budget,
			&row.CPM,
			&row.Req,
			&row.Platform,
			&row.DocLink,
			&row.Status,
			&row.CreatedAt,
			&nextCursor,
		)
		if err != nil {
			log.Printf("Error scanning campaign: %v\n", err.Error())
			return nil, 0, false, err
		}
		output = append(output, row)
	}
	// return the fetched the feed
	HasMore := len(output) > limit
	n := min(limit, len(output))
	if HasMore {
		nextCursor = prevCursor
	}
	return output[:n], nextCursor, HasMore, nil
}

func (c *CampaignStore) GetMultipleCampaigns(ctx context.Context, campaignIDs []string,
) ([]CampaignResp, error) {
	query := `
		SELECT c.id, c.brand_id, b.name AS brand, c.title, c.budget, c.cpm, 
		c.requirements, c.platform, c.doc_link, c.status, c.created_at
		FROM campaigns c
		LEFT JOIN brands b ON c.brand_id = b.id
		WHERE c.id IN $1
	`
	rows, err := c.db.QueryContext(ctx, query, campaignIDs)
	if err != nil {
		log.Printf("error getting the campaigns: %s\n", err.Error())
		return nil, err
	}
	defer rows.Close()
	var output []CampaignResp
	for rows.Next() {
		var row CampaignResp
		err = rows.Scan(
			&row.Id,
			&row.BrandId,
			&row.Brand,
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

func (c *CampaignStore) GetCampaign(ctx context.Context, id string) (*CampaignResp, error) {
	query := `
		SELECT c.id, c.brand_id, b.name AS brand, c.title, c.budget, c.cpm, 
		c.requirements, c.platform, c.doc_link, c.status, c.created_at
		FROM campaigns c
		LEFT JOIN brands b ON c.brand_id = b.id
		WHERE c.id = $1
	`
	var row CampaignResp
	err := c.db.QueryRowContext(ctx, query, id).Scan(
		&row.Id,
		&row.BrandId,
		&row.Brand,
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
		log.Printf("Error deleting campaign: %s\n", err.Error())
		return err
	}

	// successfully deleted campaign
	return nil
}
