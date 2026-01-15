package db

import (
	"context"
	"database/sql"
	"log"
)

const (
	ApplicationApprove = 1 // approve application status
	ApplicationReject  = 0 // reject application status
	ApplicationPending = 2 // pending application status
)

type ApplicationStore struct {
	db *sql.DB
}

type CampaignApplication struct {
	Id         string `json:"id" binding:"required"`
	CampaignId string `json:"campaign_id" binding:"required"`
	CreatorId  string `json:"creator_id" binding:"required"`
	Status     int    `json:"status"`
}

type ApplicationResponse struct {
	Id         string `json:"id"`
	CampaignId string `json:"campaign_id"`
	CreatorId  string `json:"creator_id"`
	Status     string `json:"status"`
	CreatedAt  string `json:"created_at"`
}

type ApplicationFeedResponse struct {
	Id            string `json:"id"`
	CampaignTitle string `json:"campaign_name"`
	Brand         string `json:"brand_name"`
	Status        int    `json:"status"`
	CreatedAt     string `json:"created_at"`
}

func NewApplicationStore(db *sql.DB) *ApplicationStore {
	return &ApplicationStore{db: db}
}

func (s *ApplicationStore) GetApplicationByID(
	ctx context.Context, appl_id string,
) (ApplicationResponse, error) {
	var appl ApplicationResponse
	query := `
		SELECT id, campaign_id, creator_id, status, created_at
		FROM applications
		WHERE id = $1
	`
	// trying to get the application by id
	err := s.db.QueryRowContext(ctx, query, appl_id).Scan(
		&appl.Id,
		&appl.CampaignId,
		&appl.CreatorId,
		&appl.Status,
		&appl.CreatedAt,
	)
	// error occured
	if err != nil {
		log.Printf("error while getting application: %v", err.Error())
		return ApplicationResponse{}, err
	}
	// successfully got the application
	return appl, nil
}

func (s *ApplicationStore) GetCreatorApplications(
	ctx context.Context, creator_id string, offset, limit int,
) ([]ApplicationFeedResponse, bool, error) {
	if creator_id == "" {
		return nil, false, ErrInvalidId
	}
	if offset < 0 || limit < 0 {
		return nil, false, ErrInvalidArgs
	}
	var output []ApplicationFeedResponse
	var hasMore = false
	// latest first search
	query := `
		SELECT a.id, c.title, b.name, a.status, a.created_at
		FROM applications a
		LEFT JOIN campaigns c ON c.id = a.campaign_id
		LEFT JOIN brands b ON b.id = c.brand_id
		WHERE creator_id = $1
		ORDER BY created_at DESC
		OFFSET $2 LIMIT $3
	`
	// trying to get the applications by creator_id
	rows, err := s.db.QueryContext(ctx, query, creator_id, offset, limit+1)
	// error occured
	if err != nil {
		log.Printf("error while getting %s applications: %v", creator_id, err.Error())
		return nil, false, err
	}
	defer rows.Close()
	for rows.Next() {
		var appl ApplicationFeedResponse
		// trying to get the application
		err := rows.Scan(
			&appl.Id,
			&appl.CampaignTitle,
			&appl.Brand,
			&appl.Status,
			&appl.CreatedAt,
		)
		// error occured
		if err != nil {
			log.Printf("error while getting application: %v", err.Error())
			return nil, false, err
		}
		output = append(output, appl)
	}
	if len(output) > limit {
		hasMore = true
		output = output[:limit]
	}

	// successfully got user applications
	return output, hasMore, nil
}

func (s *ApplicationStore) GetCampaignApplications(
	ctx context.Context, campaign_id string, offset, limit int,
) ([]ApplicationResponse, error) {
	if campaign_id == "" {
		return nil, ErrInvalidId
	}
	var output []ApplicationResponse
	// latest first search
	query := `
		SELECT id, campaign_id, creator_id, status, created_at
		FROM applications
		WHERE campaign_id = $1
		ORDER BY created_at DESC
		OFFSET $2 LIMIT $3
	`
	// trying to get the applications by creator_id
	rows, err := s.db.QueryContext(ctx, query, campaign_id, offset, limit)
	// error occured
	if err != nil {
		log.Printf("error while getting %s applications: %v", campaign_id, err.Error())
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var appl ApplicationResponse
		// trying to get the application
		err := rows.Scan(
			&appl.Id,
			&appl.CampaignId,
			&appl.CreatorId,
			&appl.Status,
			&appl.CreatedAt,
		)
		// error occured
		if err != nil {
			log.Printf("error while getting application: %v", err.Error())
			return nil, err
		}
		output = append(output, appl)
	}

	// successfully got user applications
	return output, nil
}

func (s *ApplicationStore) CreateApplication(
	ctx context.Context, appl CampaignApplication,
) error {
	query := `
		INSERT INTO applications (id, campaign_id, creator_id)
		VALUES ($1, $2, $3)
	`
	_, err := s.db.ExecContext(ctx, query,
		appl.Id, appl.CampaignId, appl.CreatorId,
	)
	/// error occured while creating application
	if err != nil {
		log.Printf("error while creating appication\ncampaign: %s, creator: %s: %q\n",
			appl.CampaignId, appl.CreatorId, err.Error())
		return err
	}
	// successfully submitted application
	return nil
}

// Validate the application id before calling the function
func (s *ApplicationStore) SetApplicationStatus(
	ctx context.Context, appl_id string, status int,
) error {
	// validate the status first
	if status != ApplicationApprove && status != ApplicationReject && status != ApplicationPending {
		return ErrInvalidStatus
	}
	query := `
		UPDATE applications
		SET status = $1
		WHERE id = $2
	`
	res, err := s.db.ExecContext(ctx, query, status, appl_id)

	if err != nil {
		log.Printf("server error: %v", err.Error())
		return ErrServer
	}
	// invalid id was requested
	if count, _ := res.RowsAffected(); count == 0 {
		log.Printf("invalid application id received\n")
		return sql.ErrNoRows
	}

	// successfully updated the application status
	return nil
}

func (s *ApplicationStore) DeleteApplication(
	ctx context.Context, appl_id string,
) error {
	query := `
		DELETE FROM applications
		WHERE id = $1
	`
	res, err := s.db.ExecContext(ctx, query, appl_id)
	if err != nil {
		log.Printf("server error: %v", err.Error())
		return ErrServer
	}
	if count, _ := res.RowsAffected(); count == 0 {
		log.Printf("invalid application id received\n")
		return sql.ErrNoRows
	}
	// successfully deleted the application status
	return nil
}
