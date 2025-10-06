package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"
)

type SubmissionStore struct {
	db *sql.DB
}

type Submission struct {
	Id            string  `json:"id"`
	CreatorId     string  `json:"creator_id"`
	CampaignId    string  `json:"campaign_id"`
	Url           string  `json:"url"`
	Status        int     `json:"status"`
	Views         int     `json:"views"`
	SyncFrequency int     `json:"sync_frequency,omitempty"`
	Earnings      float64 `json:"earnings"`
	CreatedAt     string  `json:"created_at"`
	LastSyncedAt  string  `json:"last_synced_at"`
}

type UpdateSubmission struct {
	Id       string   `json:"id"`
	Status   *int     `json:"status"`
	Url      *string  `json:"url"`
	Views    *int     `json:"views"`
	Earnings *float64 `json:"earnings"`
}

type Filter struct {
	CreatorId  *string `json:"creator_id"`
	CampaignId *string `json:"campaign_id"`
	Time       *string `json:"time"`
}

func (s *SubmissionStore) MakeSubmission(ctx context.Context, sub Submission) error {
	query := `
		INSERT INTO submissions (id, creator_id, campaign_id, url, status)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := s.db.ExecContext(ctx, query,
		sub.Id, sub.CreatorId, sub.CampaignId, sub.Url, sub.Status)
	if err != nil {
		// internal server error
		log.Printf("error making submission: %v\n", err.Error())
		return err
	}
	// successfully submitted
	return nil
}

func (s *SubmissionStore) DeleteSubmission(ctx context.Context, id string) error {
	query := `
		DELETE FROM submissions
		WHERE id = $1
	`
	v, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		// internal server error
		log.Printf("error deleting the submission: %v", err.Error())
		return err
	}
	cnt, _ := v.RowsAffected()
	if cnt == 0 {
		// invalid id error
		log.Printf("invalid id: %v", id)
		return sql.ErrNoRows
	}

	// successfully deleted the submission
	return nil
}

func (s *SubmissionStore) FindSubmissionById(ctx context.Context, id string) (*Submission, error) {
	query := `
		SELECT id, creator_id, campaign_id, url, status, views,
		earnings, created_at, last_synced_at
		FROM submissions
		WHERE id = $1
	`
	var sub Submission
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&sub.Id,
		&sub.CreatorId,
		&sub.CampaignId,
		&sub.Url,
		&sub.Status,
		&sub.Views,
		&sub.Earnings,
		&sub.CreatedAt,
		&sub.LastSyncedAt,
	)
	if err != nil {
		// internal server error/ invalid query
		log.Printf("error scanning submission: %v", err.Error())
		return nil, err
	}

	// successfully fetched submission
	return &sub, nil
}

// This function filters submissions using filter struct
// It accepts the Time in format "MM-YYYY"
func (s *SubmissionStore) FindSubmissionsByFilters(
	ctx context.Context, filter Filter) ([]Submission, error,
) {
	var output []Submission
	query := `
		SELECT id, creator_id, campaign_id, url, status, views, earnings, created_at,
		last_synced_at
		FROM submissions
		WHERE 
	`
	// dynamically build the query
	queryBuilder := strings.Builder{}
	queryBuilder.WriteString(query)

	conditions := []string{} // filters
	args := []any{}          // arguments
	var i int = 1
	if filter.CampaignId != nil {
		conditions = append(conditions, fmt.Sprintf("campaign_id = $%d", i))
		args = append(args, *filter.CampaignId)
		i++
	}
	if filter.CreatorId != nil {
		conditions = append(conditions, fmt.Sprintf("creator_id = $%d", i))
		args = append(args, *filter.CreatorId)
		i++
	}
	if filter.Time != nil {
		t, err := time.Parse("01-2006", *filter.Time) // "MM-YYYY"
		if err != nil {
			return nil, fmt.Errorf("invalid time format: %v", err)
		}

		month := int(t.Month())
		year := t.Year()
		conditions = append(conditions,
			fmt.Sprintf("EXTRACT(MONTH FROM created_at) > $%d AND EXTRACT(YEAR FROM created_at) > $%d", i, i+1))
		args = append(args, month, year)
		i += 2
	}

	// building the final query
	queryBuilder.WriteString(strings.Join(conditions, " AND "))
	query = queryBuilder.String()

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		// internal server error
		log.Printf("error filtering submissions: %v", err.Error())
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var sub Submission
		err = rows.Scan(
			&sub.Id,
			&sub.CreatorId,
			&sub.CampaignId,
			&sub.Url,
			&sub.Status,
			&sub.Views,
			&sub.Earnings,
			&sub.CreatedAt,
			&sub.LastSyncedAt,
		)
		// append to the output
		if err != nil {
			log.Printf("error scanning submission: %v", err.Error())
			return nil, err
		}
		output = append(output, sub)
	}

	// successful filtering
	return output, nil
}

func (s *SubmissionStore) FindMySubmissions(ctx context.Context, time_ string,
	subids []string,
) ([]Submission, error) {
	query := `
		SELECT id, creator_id, campaign_id, url, status, views, earnings, created_at,
		last_synced_at
		FROM submissions
		WHERE
	`

	queryBuilder := strings.Builder{}
	queryBuilder.WriteString(query)

	var conditions []string
	var args []any
	i := 1
	if time_ != "" {
		t, err := time.Parse("01-2006", time_) // "MM-YYYY"
		if err != nil {
			return nil, fmt.Errorf("invalid time format: %v", err)
		}

		month := int(t.Month())
		year := t.Year()
		conditions = append(conditions,
			fmt.Sprintf("EXTRACT(MONTH FROM created_at) > $%d AND EXTRACT(YEAR FROM created_at) > $%d", i, i+1))
		args = append(args, month, year)
		i += 2
	}
	conditions = append(conditions, fmt.Sprintf("id IN $%d", i))
	args = append(args, subids)

	// building the final query
	queryBuilder.WriteString(strings.Join(conditions, " AND "))
	query = queryBuilder.String()

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		// internal server error
		log.Printf("error filtering submissions: %v", err.Error())
		return nil, err
	}
	defer rows.Close()
	var output []Submission
	for rows.Next() {
		var sub Submission
		err = rows.Scan(
			&sub.Id,
			&sub.CreatorId,
			&sub.CampaignId,
			&sub.Url,
			&sub.Status,
			&sub.Views,
			&sub.Earnings,
			&sub.CreatedAt,
			&sub.LastSyncedAt,
		)
		// append to the output
		if err != nil {
			log.Printf("error scanning submission: %v", err.Error())
			return nil, err
		}
		output = append(output, sub)
	}

	// succesfully fetched the submissions
	return output, nil
}

// This function updates a submission entity
func (s *SubmissionStore) UpdateSubmission(ctx context.Context, payload UpdateSubmission) error {
	var queryBuilder strings.Builder
	queryBuilder.WriteString("UPDATE submissions SET ")
	var args []any
	var conditions []string
	var i int = 1
	if payload.Status != nil {
		conditions = append(conditions, fmt.Sprintf("status = $%d", i))
		args = append(args, *payload.Status)
		i++
	}
	if payload.Url != nil {
		conditions = append(conditions, fmt.Sprintf("url = $%d", i))
		args = append(args, *payload.Url)
		i++
	}
	if payload.Views != nil {
		conditions = append(conditions, fmt.Sprintf("views = $%d", i))
		args = append(args, *payload.Views)
		i++
	}
	if payload.Earnings != nil {
		conditions = append(conditions, fmt.Sprintf("earnings = $%d", i))
		args = append(args, *payload.Earnings)
		i++
	}
	// building query
	queryBuilder.WriteString(strings.Join(conditions, ", "))
	queryBuilder.WriteString(fmt.Sprintf(" WHERE id = $%d", i))
	args = append(args, payload.Id)
	query := queryBuilder.String()

	if len(conditions) == 0 {
		return errors.New("invalid field to update")
	}
	_, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		log.Printf("Error updating brand: %v\n", err.Error())
		return err
	}

	// successfully updated submission details
	return nil
}

func (s *SubmissionStore) ChangeViews(ctx context.Context, delta int, id string) error {
	// update the views count and the last_synced_at
	query := `
		UPDATE submissions
		SET views = views + $1, last_synced_at = now()
		WHERE id = $2
	`
	res, err := s.db.ExecContext(ctx, query, delta, id)
	if err != nil {
		log.Printf("error updating views for id = %s: %v\n", id, err)
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		// This can fail with drivers that don’t support it
		log.Printf("could not determine rows affected: %v\n", err)
		return fmt.Errorf("rows affected: %w", err)
	}

	if count == 0 {
		log.Printf("no submission found with id = %q\n", id)
		return sql.ErrNoRows
	}
	return nil
}
