package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Alter-Sitanshu/campaignHub/internals/db"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type SubmissionPayload struct {
	CreatorId  string `json:"creator_id" binding:"required"`
	CampaignId string `json:"campaign_id" binding:"required"`
	Url        string `json:"url" binding:"required"`
	Status     int    `json:"status" binding:"required,oneof=1 0 3"`
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

	err := app.store.SubmissionInterface.MakeSubmission(ctx, submission)
	if err != nil {
		c.JSON(http.StatusInternalServerError, WriteError("server error"))
		return
	}
	submission.CreatedAt = formattedTime
	// cache the submission
	app.cache.SetCreatorSubmissions(ctx, Entity.GetID(), []string{submission.Id})
	app.cache.SetSubmissionEarnings(ctx, submission.Id, submission.Earnings)
	app.cache.SetSubmissionStatus(ctx, submission.Id, submission.Status)
	app.cache.SetViewCount(ctx, submission.Id, submission.Views)

	// successfully made the submission
	c.JSON(http.StatusCreated, WriteResponse(submission))
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
	output, err := app.store.SubmissionInterface.FindSubmissionsByFilters(ctx, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, WriteError("server error"))
		return
	}

	// successfully filtered
	c.JSON(http.StatusOK, WriteResponse(output))
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

	sub_response, _ := app.store.SubmissionInterface.FindSubmissionById(ctx, sub_id)
	// cache the submission
	app.cache.SetSubmissionEarnings(ctx, sub.Id, sub_response.Earnings)
	app.cache.SetSubmissionStatus(ctx, sub.Id, sub_response.Status)

	// update successful
	c.JSON(http.StatusOK, WriteResponse(sub_response))
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
	views, err := app.cache.GetViewCount(ctx, sub.Id)
	if err == nil {
		sub.Views = views
	}

	// successful fetching
	c.JSON(http.StatusOK, WriteResponse(sub))
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
		output, err := app.store.SubmissionInterface.FindMySubmissions(ctx, time_, subids)
		if err != nil {
			c.JSON(http.StatusInternalServerError, WriteError("server error"))
			return
		}
		for i := range output {
			amt, err := app.cache.GetSubmissionEarnings(ctx, output[i].Id)
			if err == nil {
				output[i].Earnings = amt
			}
			views, err := app.cache.GetViewCount(ctx, output[i].Id)
			if err == nil {
				output[i].Views = views
			}
			status, err := app.cache.GetSubmissionStatus(ctx, output[i].Id)
			if err == nil {
				output[i].Status = status
			}
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
	output, err := app.store.SubmissionInterface.FindSubmissionsByFilters(ctx, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, WriteError("server error"))
		return
	}

	// successfully filtered
	c.JSON(http.StatusOK, WriteResponse(output))
}
