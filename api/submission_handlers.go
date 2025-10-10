package api

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/Alter-Sitanshu/campaignHub/internals/cache"
	"github.com/Alter-Sitanshu/campaignHub/internals/db"
	"github.com/Alter-Sitanshu/campaignHub/internals/platform"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type SubmissionPayload struct {
	CreatorId  string `json:"creator_id" binding:"required"`
	CampaignId string `json:"campaign_id" binding:"required"`
	Url        string `json:"url" binding:"required"`
	Status     int    `json:"status" binding:"required,oneof=1 0 3"`
}

type SubmissionResponse struct {
	Submission db.Submission       `json:"submission"`
	Meta       cache.VideoMetadata `json:"meta"`
}

func (app *Application) CreateSubmission(c *gin.Context) {
	ctx := c.Request.Context()
	LogInUser, ok := c.Get("user")
	if !ok {
		c.JSON(http.StatusUnauthorized, WriteError("unauthorised request"))
		return
	}
	Entity, ok := LogInUser.(db.AuthenticatedEntity)
	if !ok {
		c.JSON(http.StatusUnauthorized, WriteError("unauthorised request"))
		return
	}
	var payload SubmissionPayload
	// validating the struct
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, WriteError(err.Error()))
		return
	}
	if Entity.GetID() != payload.CreatorId {
		c.JSON(http.StatusUnauthorized, WriteError("unauthorised request"))
		return
	}
	// making the submission object
	unixTimestamp := time.Now().Unix()
	t := time.Unix(unixTimestamp, 0)                 // Convert Unix timestamp to time.Time object
	formattedTime := t.Format("2006-01-02 15:04:05") // Format using a reference time
	submission := db.Submission{
		Id:         uuid.New().String(),
		CreatorId:  payload.CreatorId,
		CampaignId: payload.CampaignId,
		Url:        payload.Url,
		Status:     payload.Status,
	}

	vid, err := platform.ParseVideoURL(payload.Url)
	if err != nil {
		log.Printf("error: %s\n", err.Error())
		c.JSON(http.StatusBadRequest, WriteError(err.Error()))
		return
	}
	// Fetch the Meta Data for the sumission

	data, err := app.factory.GetVideoDetails(ctx, vid.Name, vid.VideoID)
	if err != nil {
		log.Printf("error fetching meta data: %s\n", err.Error())
		c.JSON(http.StatusInternalServerError, WriteError("server error try again"))
		return
	}

	Data := data.(platform.VideoMetadata) // convert the meta data type into desired struct

	// populate the meta data
	metaData := cache.VideoMetadata{
		SubmissionID: submission.Id,
		VideoID:      Data.VideoID,
		Platform:     Data.Platform,
		Title:        Data.Title,
		ViewCount:    Data.ViewCount,
		LikeCount:    Data.LikeCount,
		Thumbnail:    Data.Thumbnails,
		UploadedAt:   Data.UploadedAt,
	}

	// Populating the submission with the meta data
	submission.VideoTitle = metaData.Title
	submission.VideoID = metaData.VideoID
	submission.ThumbnailURL = metaData.Thumbnail.URL
	submission.Views = metaData.ViewCount
	submission.LikeCount = metaData.LikeCount
	submission.SyncFrequency = DefaultSyncFrequency

	err = app.store.SubmissionInterface.MakeSubmission(ctx, submission)
	if err != nil {
		c.JSON(http.StatusInternalServerError, WriteError("server error"))
		return
	}
	submission.CreatedAt = formattedTime
	submission.LastSyncedAt = submission.CreatedAt // Just now synced

	// cache the submission
	app.cache.SetCreatorSubmissions(ctx, Entity.GetID(), []string{submission.Id})
	app.cache.SetSubmissionEarnings(ctx, submission.Id, submission.Earnings)
	app.cache.SetSubmissionStatus(ctx, submission.Id, submission.Status)
	app.cache.SetVideoMetadata(ctx, submission.Id, metaData)

	// successfully made the submission
	resp := SubmissionResponse{
		Submission: submission,
		Meta:       metaData,
	}
	c.JSON(http.StatusCreated, WriteResponse(resp))
}

func (app *Application) DeleteSubmission(c *gin.Context) {
	ctx := c.Request.Context()
	LogInUser, ok := c.Get("user")
	if !ok {
		c.JSON(http.StatusUnauthorized, WriteError("unauthorised request"))
		return
	}
	Entity, ok := LogInUser.(db.AuthenticatedEntity)
	if !ok {
		c.JSON(http.StatusUnauthorized, WriteError("unauthorised request"))
		return
	}
	sub_id := c.Request.PathValue("sub_id")
	// validate the submission id
	if ok := uuid.Validate(sub_id); ok != nil {
		c.JSON(http.StatusInternalServerError, WriteError("invalid credentials"))
		return
	}

	// check if the submission belongs to the user
	submission, err := app.store.SubmissionInterface.FindSubmissionById(ctx, sub_id)
	if err != nil {
		if err == db.ErrNotFound {
			c.JSON(http.StatusInternalServerError, WriteError("invalid credentials"))
			return
		}
		c.JSON(http.StatusInternalServerError, WriteError("server error"))
		return
	}
	if submission.CreatorId != Entity.GetID() && Entity.GetRole() != "admin" {
		c.JSON(http.StatusUnauthorized, WriteError("unauthorised request"))
		return
	}

	// try deleting the submission
	err = app.store.SubmissionInterface.DeleteSubmission(ctx, sub_id)
	if err != nil {
		if err == db.ErrNotFound {
			c.JSON(http.StatusInternalServerError, WriteError("invalid credentials"))
			return
		}
		c.JSON(http.StatusInternalServerError, WriteError("server error"))
		return
	}

	// invalidate the affected caches
	app.cache.InvalidateSubmissionStatus(ctx, sub_id)
	app.cache.InvalidateSubmissionEarnings(ctx, sub_id)
	app.cache.InvalidateOneCreatorSubmissions(ctx, submission.CreatorId, sub_id)
	app.cache.InvalidateVideoMetadata(ctx, sub_id)

	// delete was successful
	c.JSON(http.StatusNoContent, WriteResponse("delete was successful"))
}

func (app *Application) FilterSubmissions(c *gin.Context) {
	ctx := c.Request.Context()
	LogInUser, ok := c.Get("user")
	if !ok {
		c.JSON(http.StatusUnauthorized, WriteError("unauthorised request"))
		return
	}
	Entity, ok := LogInUser.(db.AuthenticatedEntity)
	if !ok {
		c.JSON(http.StatusUnauthorized, WriteError("unauthorised request"))
		return
	}
	if Entity.GetRole() != "admin" {
		c.JSON(http.StatusUnauthorized, WriteError("unauthorised request"))
		return
	}
	// only the admin is allowed
	creator_id := c.Query("creator_id")
	campaign_id := c.Query("campaign_id")
	time_ := c.Query("time")
	limit, err := strconv.Atoi(c.Query("limit"))
	if err != nil {
		c.JSON(http.StatusBadRequest, WriteError("bad request on query limit"))
		return
	}
	offset, err := strconv.Atoi(c.Query("offset"))
	if err != nil {
		c.JSON(http.StatusBadRequest, WriteError("bad request on offset"))
		return
	}
	var filter db.Filter
	if creator_id != "" {
		filter.CreatorId = &creator_id
	}
	if campaign_id != "" {
		filter.CampaignId = &campaign_id
	}
	if time_ != "" {
		filter.Time = &time_
	}
	// check the time format
	if filter.Time != nil {
		_, err := time.Parse("01-2006", *filter.Time) // "MM-YYYY"
		if err != nil {
			c.JSON(http.StatusBadRequest, WriteError(fmt.Sprintf("invalid time format: %v", err)))
			return
		}
	}
	// get the filtered submissions
	output, err := app.store.SubmissionInterface.FindSubmissionsByFilters(ctx, filter, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, WriteError("server error"))
		return
	}
	resp := make([]SubmissionResponse, len(output))

	// Attaching the meta data from the cache
	// If i don't get the meta data in cache i ignore that for now
	// Later i can add that by API calls i will figure something out

	for i := range len(output) {
		resp[i].Submission = output[i]
		VideoMeta, err := app.cache.GetVideoMetadata(ctx, output[i].Id)
		if err == nil {
			resp[i].Meta = *VideoMeta
			// i do not need to throw an error
			// i will ignore the failure and move forward
		}
	}

	// successfully filtered
	c.JSON(http.StatusOK, WriteResponse(resp))
}

func (app *Application) UpdateSubmission(c *gin.Context) {
	ctx := c.Request.Context()
	LogInUser, ok := c.Get("user")
	if !ok {
		c.JSON(http.StatusUnauthorized, WriteError("unauthorised request"))
		return
	}
	Entity, ok := LogInUser.(db.AuthenticatedEntity)
	if !ok {
		c.JSON(http.StatusUnauthorized, WriteError("unauthorised request"))
		return
	}
	sub_id := c.Request.PathValue("sub_id")
	if ok := uuid.Validate(sub_id); ok != nil {
		c.JSON(http.StatusInternalServerError, WriteError("invalid credentials"))
		return
	}
	// get the submission
	sub, err := app.store.SubmissionInterface.FindSubmissionById(ctx, sub_id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, WriteError("invalid filter request"))
		return
	}
	if sub.CreatorId != Entity.GetID() && Entity.GetRole() != "admin" {
		c.JSON(http.StatusUnauthorized, WriteError("unauthorised request"))
		return
	}
	// allowed only if the user is the owner or the admin
	// get the update payload
	var payload db.UpdateSubmission
	if err := c.BindJSON(&payload); err != nil {
		c.JSON(http.StatusInternalServerError, WriteError("invalid filter request"))
		return
	}

	// try the update
	if err := app.store.SubmissionInterface.UpdateSubmission(ctx, payload); err != nil {
		c.JSON(http.StatusInternalServerError, WriteError("server error"))
		return
	}
	var resp SubmissionResponse
	sub_response, _ := app.store.SubmissionInterface.FindSubmissionById(ctx, sub_id)
	resp.Submission = *sub_response
	// cache the submission
	app.cache.SetSubmissionEarnings(ctx, sub.Id, sub_response.Earnings)
	app.cache.SetSubmissionStatus(ctx, sub.Id, sub_response.Status)

	// Get the meta data for the updated submission from the cache
	meta, err := app.cache.GetVideoMetadata(ctx, sub_id)
	if err == nil {
		// cache miss
		resp.Meta = *meta
	}

	// update successful
	c.JSON(http.StatusOK, WriteResponse(resp))
}

func (app *Application) GetSubmission(c *gin.Context) {
	ctx := c.Request.Context()
	LogInUser, ok := c.Get("user")
	if !ok {
		c.JSON(http.StatusUnauthorized, WriteError("unauthorised request"))
		return
	}
	Entity, ok := LogInUser.(db.AuthenticatedEntity)
	if !ok {
		c.JSON(http.StatusUnauthorized, WriteError("unauthorised request"))
		return
	}

	sub_id := c.Request.PathValue("sub_id")
	if ok := uuid.Validate(sub_id); ok != nil {
		c.JSON(http.StatusInternalServerError, WriteError("invalid credentials"))
		return
	}

	// fetch from the cache
	var output SubmissionResponse
	VideoMetaData, err := app.cache.GetVideoMetadata(ctx, sub_id)
	if err == nil {
		// cache hit
		output.Meta = *VideoMetaData
	}

	// try fetching the submission
	sub, err := app.store.SubmissionInterface.FindSubmissionById(ctx, sub_id)
	if err != nil {
		if err == db.ErrNotFound {
			c.JSON(http.StatusBadRequest, WriteError("invalid credentials"))
			return
		}
		c.JSON(http.StatusInternalServerError, WriteError("server error"))
	}
	if sub.CreatorId != Entity.GetID() {
		c.JSON(http.StatusUnauthorized, WriteError("unauthorised request"))
		return
	}
	// get the updated cahed values
	amt, err := app.cache.GetSubmissionEarnings(ctx, sub.Id)
	if err == nil {
		sub.Earnings = amt
	}
	status, err := app.cache.GetSubmissionStatus(ctx, sub.Id)
	if err == nil {
		sub.Status = status
	}

	// Populate the output submission
	output.Submission = *sub

	// successful fetching
	c.JSON(http.StatusOK, WriteResponse(output))
}

func (app *Application) GetMySubmissions(c *gin.Context) {
	ctx := c.Request.Context()
	LogInUser, ok := c.Get("user")
	if !ok {
		c.JSON(http.StatusUnauthorized, WriteError("unauthorised request"))
		return
	}
	Entity, ok := LogInUser.(db.AuthenticatedEntity)
	if !ok {
		c.JSON(http.StatusUnauthorized, WriteError("unauthorised request"))
		return
	}
	UserID := Entity.GetID()
	limit, err := strconv.Atoi(c.Query("limit"))
	if err != nil {
		c.JSON(http.StatusBadRequest, WriteError("bad request on query limit"))
		return
	}
	offset, err := strconv.Atoi(c.Query("offset"))
	if err != nil {
		c.JSON(http.StatusBadRequest, WriteError("bad request on query limit"))
		return
	}
	time_ := c.Query("time")
	// check the time format
	if time_ != "" {
		_, err := time.Parse("01-2006", time_) // "MM-YYYY"
		if err != nil {
			c.JSON(http.StatusBadRequest, WriteError(fmt.Sprintf("invalid time format: %v", err)))
			return
		}
	}
	// get the user submissions submissions
	// check cache for the user submissions
	subids, err := app.cache.GetCreatorSubmissions(ctx, UserID)
	// Cache Hit
	if err == nil {
		var output []SubmissionResponse
		submissions, err := app.store.SubmissionInterface.FindMySubmissions(ctx, time_, subids, limit, offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, WriteError("server error"))
			return
		}
		for i := range submissions {
			var resp SubmissionResponse
			amt, err := app.cache.GetSubmissionEarnings(ctx, submissions[i].Id)
			if err == nil {
				submissions[i].Earnings = amt
			}
			meta, err := app.cache.GetVideoMetadata(ctx, submissions[i].Id)
			if err == nil {
				submissions[i].Views = meta.ViewCount
				resp.Meta = *meta
			}
			status, err := app.cache.GetSubmissionStatus(ctx, submissions[i].Id)
			if err == nil {
				submissions[i].Status = status
			}
			resp.Submission = submissions[i]
			output = append(output, resp)
		}
		// successfully filtered
		c.JSON(http.StatusOK, WriteResponse(output))
		return
	}

	// in case cache miss
	filter := db.Filter{
		CreatorId: &UserID,
		Time:      &time_,
	}
	// Database scan for user submissions
	output, err := app.store.SubmissionInterface.FindSubmissionsByFilters(ctx, filter, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, WriteError("server error"))
		return
	}
	resp := make([]SubmissionResponse, len(output))

	// Attaching the meta data from the cache
	// If i don't get the meta data in cache i ignore that for now
	// Later i can add that by API calls i will figure something out

	for i := range len(output) {
		resp[i].Submission = output[i]
		VideoMeta, err := app.cache.GetVideoMetadata(ctx, output[i].Id)
		if err == nil {
			resp[i].Meta = *VideoMeta
			// i do not need to throw an error
			// i will ignore the failure and move forward
		}
	}

	// successfully filtered
	c.JSON(http.StatusOK, WriteResponse(resp))
}
