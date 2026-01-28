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
	Id         string `json:"id"`
	CreatorId  string `json:"creator_id"`
	CampaignId string `json:"campaign_id"`
	Url        string `json:"url"`
	Status     int    `json:"status"`
	// Meta data for the Video submission
	VideoTitle    string  `json:"video_title"`
	VideoPlatform string  `json:"video_platform"`
	VideoID       string  `json:"platform_video_id"`
	ThumbnailURL  string  `json:"thumbnail_url"`
	Views         int     `json:"views"`
	LikeCount     int     `json:"like_count"`
	VideoStatus   string  `json:"video_status"`
	Earnings      float64 `json:"earnings"`
	LastSyncedAt  string  `json:"last_synced_at"`
	// -------- x ----------
	SyncFrequency int    `json:"sync_frequency,omitempty"`
	CreatedAt     string `json:"created_at"`
}

// custom struct for the polling worker
// reduces the overhead by 50% changing from Submission -> PollingSubmission
type PollingSubmission struct {
	Id            string `json:"id"`
	CreatorId     string `json:"creator_id"`
	CampaignId    string `json:"campaign_id"`
	Url           string `json:"url"`
	Views         int    `json:"views"`
	LastSyncedAt  string `json:"last_synced_at"`
	SyncFrequency int    `json:"sync_frequency,omitempty"`
	CreatedAt     string `json:"created_at"`
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
		INSERT INTO submissions
		(
			id, creator_id, campaign_id, url, status, video_title, video_platform,
			platform_video_id, thumbnail_url, views, like_count, video_status, earnings,
			sync_frequency
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`
	_, err := s.db.ExecContext(ctx, query,
		sub.Id, sub.CreatorId, sub.CampaignId, sub.Url, sub.Status,
		sub.VideoTitle, sub.VideoPlatform, sub.VideoID, sub.ThumbnailURL,
		sub.Views, sub.LikeCount, sub.VideoStatus, sub.Earnings, sub.SyncFrequency,
	)
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
		SELECT id, creator_id, campaign_id, url, status, video_title, video_platform,
			platform_video_id, thumbnail_url, views, like_count, video_status, earnings,
			sync_frequency, created_at, last_synced_at
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
		&sub.VideoTitle, &sub.VideoPlatform, &sub.VideoID,
		&sub.ThumbnailURL, &sub.Views, &sub.LikeCount, &sub.VideoStatus,
		&sub.Earnings,
		&sub.SyncFrequency,
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
	ctx context.Context, filter Filter, limit, offset int) ([]Submission, bool, error,
) {
	var output []Submission
	query := `
		SELECT id, creator_id, campaign_id, url, status, video_title, video_platform,
			platform_video_id, thumbnail_url, views, like_count, video_status, earnings,
			sync_frequency, created_at, last_synced_at
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
	if filter.Time != nil && *filter.Time != "" {
		t, err := time.Parse("01-2006", *filter.Time) // "MM-YYYY"
		if err != nil {
			return nil, false, fmt.Errorf("invalid time format: %v", err)
		}

		month := int(t.Month())
		year := t.Year()
		conditions = append(conditions,
			fmt.Sprintf("EXTRACT(MONTH FROM created_at) >= $%d AND EXTRACT(YEAR FROM created_at) >= $%d", i, i+1))
		args = append(args, month, year)
		i += 2
	}
	// building the final query
	queryBuilder.WriteString(strings.Join(conditions, " AND "))
	// add the limit and offset constraints
	queryBuilder.WriteString(
		fmt.Sprintf(" LIMIT $%d OFFSET $%d", i, i+1),
	)
	args = append(args, limit+1, offset)
	i += 2
	query = queryBuilder.String()

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		// internal server error
		log.Printf("error filtering submissions: %v, %v, %v", query, args, err.Error())
		return nil, false, err
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
			&sub.VideoTitle, &sub.VideoPlatform, &sub.VideoID,
			&sub.ThumbnailURL, &sub.Views, &sub.LikeCount, &sub.VideoStatus,
			&sub.Earnings,
			&sub.SyncFrequency,
			&sub.CreatedAt,
			&sub.LastSyncedAt,
		)
		// append to the output
		if err != nil {
			log.Printf("error scanning submission: %v", err.Error())
			return nil, false, err
		}
		output = append(output, sub)
	}
	hasMore := len(output) > limit
	n := min(limit, len(output)) // the minimum boundary of the array output
	// successful filtering
	return output[:n], hasMore, nil
}

func (s *SubmissionStore) FindMySubmissions(ctx context.Context, time_ string,
	subids []string, limit, offset int,
) ([]Submission, error) {
	query := `
		SELECT id, creator_id, campaign_id, url, status, video_title, video_platform,
			platform_video_id, thumbnail_url, views, like_count, video_status, earnings,
			sync_frequency, created_at, last_synced_at
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
			fmt.Sprintf("EXTRACT(MONTH FROM created_at) >= $%d AND EXTRACT(YEAR FROM created_at) >= $%d", i, i+1))
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
			&sub.VideoTitle, &sub.VideoPlatform, &sub.VideoID,
			&sub.ThumbnailURL, &sub.Views, &sub.LikeCount, &sub.VideoStatus,
			&sub.Earnings,
			&sub.SyncFrequency,
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

func (s *SubmissionStore) GetSubmissionsForSync(ctx context.Context) ([]PollingSubmission, error) {
	// filter out the active submissions
	// select the submissions which have there sync frequency
	// less than the interval passed from last_sync
	query := `
        SELECT 
            s.id, 
            s.url, 
            s.views,
            s.sync_frequency,
            s.last_synced_at,
            s.creator_id,
            s.campaign_id,
			s.created_at
        FROM submissions s
        JOIN status st ON s.status = st.id
        WHERE st.name = 'active'
		AND s.last_synced_at <= NOW() - (s.sync_frequency::text || ' minutes')::INTERVAL
        ORDER BY s.last_synced_at ASC
        LIMIT 500;
    `
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		log.Printf("error while fetching submissions to sync: %s\n", err.Error())
		return nil, ErrServer
	}
	defer rows.Close()

	// output slice for all the submissions
	var output []PollingSubmission

	for rows.Next() {
		var sub PollingSubmission
		err = rows.Scan(
			&sub.Id,
			&sub.Url,
			&sub.Views,
			&sub.SyncFrequency,
			&sub.LastSyncedAt,
			&sub.CreatorId,
			&sub.CampaignId,
			&sub.CreatedAt,
		)
		if err != nil {
			log.Printf("error scanning polling submissions: %s\n", err.Error())
			return nil, err
		}
		output = append(output, sub)
	}

	// no error in submission fetching
	// success
	return output, nil
}

func (s *SubmissionStore) UpdateSyncFrequency(ctx context.Context, id string, freq int) error {
	query := `
		UPDATE submissions
		SET sync_frequency = $1
		WHERE id = $2
	`
	res, err := s.db.ExecContext(ctx, query, freq, id)
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
