package api

import (
	"log"
	"net/http"
	"strconv"

	"github.com/Alter-Sitanshu/campaignHub/internals/db"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CampaignPayload struct {
	BrandId string  `json:"brand_id" binding:"required"`
	Title   string  `json:"title" binding:"required"`
	Budget  float64 `json:"budget" binding:"required"`
	CPM     float64 `json:"cpm" binding:"required"`
	Req     string  `json:"requirements" binding:"required"`
	// added this to segregate the campaigns on the basis of platform
	Platform string `json:"platform"`
	DocLink  string `json:"doc_link" binding:"required"`
	Status   int    `json:"status" binding:"required,oneof=0 1 3"`
}

func (app *Application) CreateCampaign(c *gin.Context) {
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
	if !ok {
		c.JSON(http.StatusUnauthorized, WriteError("unauthorised request"))
		return
	}
	var payload CampaignPayload
	if err := c.BindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, WriteError("invalid input"))
		return
	}
	if Entity.GetID() != payload.BrandId {
		c.JSON(http.StatusUnauthorized, WriteError("unauthorised request"))
		return
	}
	// making the payload
	campaign := db.Campaign{
		Id:       uuid.New().String(),
		BrandId:  payload.BrandId,
		Title:    payload.Title,
		Budget:   payload.Budget,
		CPM:      payload.CPM,
		Req:      payload.Req,
		Platform: payload.Platform,
		DocLink:  payload.DocLink,
		Status:   payload.Status,
	}
	err := app.store.CampaignInterace.LaunchCampaign(ctx, &campaign)
	if err != nil {
		log.Printf("error campaign: %v", err.Error())
		c.JSON(http.StatusInternalServerError, WriteError("server error"))
		return
	}
	// cache the campaign
	err = app.cache.SetCampaign(ctx, campaign.Id, campaign)
	if err != nil {
		log.Printf("error caching campaign: %s\n", err.Error())
	}
	// campaign launched
	c.JSON(http.StatusOK, WriteResponse(&campaign))
}

func (app *Application) GetBrandCampaigns(c *gin.Context) {
	ctx := c.Request.Context()
	BrandID := c.Query("brandid")
	if ok := uuid.Validate(BrandID); ok != nil {
		c.JSON(http.StatusBadRequest, WriteError("invalid request"))
		return
	}
	campaigns, err := app.store.CampaignInterace.GetBrandCampaigns(ctx, BrandID)
	if err != nil {
		// server error
		c.JSON(http.StatusInternalServerError, WriteError("internal server error"))
		return
	}
	// successfully retreived the brand campaign
	c.JSON(http.StatusOK, WriteResponse(campaigns))
}

func (app *Application) StopCampaign(c *gin.Context) {
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
	ID := c.Request.PathValue("campaign_id")
	if ok := uuid.Validate(ID); ok != nil {
		c.JSON(http.StatusBadRequest, WriteError("invalid request"))
		return
	}
	// fetch campaign details
	campaign, err := app.store.CampaignInterace.GetCampaign(ctx, ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, WriteError("internal server error"))
		return
	}
	if campaign.BrandId != Entity.GetID() && Entity.GetRole() != "admin" {
		c.JSON(http.StatusUnauthorized, WriteError("unauthorised request"))
		return
	}

	// delete the campaign
	if err := app.store.CampaignInterace.EndCampaign(ctx, ID); err != nil {
		c.JSON(http.StatusInternalServerError, WriteError("internal server error"))
		return
	}
	// invalidate the active campaign
	app.cache.RemoveActiveCampaign(ctx, campaign.Id)
	// successfully delted the campaign
	c.JSON(http.StatusNoContent, WriteResponse("campaign ended"))
}

func (app *Application) DeleteCampaign(c *gin.Context) {
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
	ID := c.Request.PathValue("campaign_id")
	if ok := uuid.Validate(ID); ok != nil {
		c.JSON(http.StatusBadRequest, WriteError("invalid request"))
		return
	}
	// fetch campaign details
	campaign, err := app.store.CampaignInterace.GetCampaign(ctx, ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, WriteError("internal server error"))
		return
	}
	if campaign.BrandId != Entity.GetID() && Entity.GetRole() != "admin" {
		c.JSON(http.StatusUnauthorized, WriteError("unauthorised request"))
		return
	}
	// delete the campaign
	if err := app.store.CampaignInterace.DeleteCampaign(ctx, ID); err != nil {
		c.JSON(http.StatusInternalServerError, WriteError("internal server error"))
		return
	}
	// invalidate the affected campaigns
	app.cache.InvalidateCampaign(ctx, campaign.Id)
	app.cache.InvalidateOneBrandCampaign(ctx, campaign.BrandId, campaign.Id)

	// successfully deleted the campaign
	c.JSON(http.StatusNoContent, WriteResponse("campaign ended"))
}

func (app *Application) GetUserCampaigns(c *gin.Context) {
	ctx := c.Request.Context()
	LogInUser, ok := c.Get("user")
	if !ok {
		c.JSON(http.StatusUnauthorized, WriteError("unauthorised request"))
		return
	}
	_, ok = LogInUser.(db.AuthenticatedEntity)
	if !ok {
		c.JSON(http.StatusUnauthorized, WriteError("unauthorised request"))
		return
	}
	UserID := c.Query("userid")
	if ok := uuid.Validate(UserID); ok != nil {
		c.JSON(http.StatusBadRequest, WriteError("invalid request"))
		return
	}

	// check the cache first
	res, err := app.cache.GetUserCampaigns(ctx, UserID)
	if err == nil {
		// cache hit
		campaigns, err := app.store.CampaignInterace.GetMultipleCampaigns(ctx, res)
		if err != nil {
			// server error
			c.JSON(http.StatusInternalServerError, WriteError("internal server error"))
			return
		}
		c.JSON(http.StatusOK, WriteResponse(campaigns))
		return
	}

	campaigns, err := app.store.CampaignInterace.GetUserCampaigns(ctx, UserID)
	if err != nil {
		// server error
		c.JSON(http.StatusInternalServerError, WriteError("internal server error"))
		return
	}

	// cache the user campaigns
	campaignIDs := make([]string, len(campaigns))
	for _, v := range campaigns {
		campaignIDs = append(campaignIDs, v.Id)
	}
	err = app.cache.SetUserCampaigns(ctx, UserID, campaignIDs)
	if err != nil {
		log.Printf("error caching the user campaigns: %s", err.Error())
	}
	// successfully retreived the brand campaign
	c.JSON(http.StatusOK, WriteResponse(campaigns))
}

func (app *Application) UpdateCampaign(c *gin.Context) {
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
	campaign_id := c.Request.PathValue("campaign_id")
	var payload db.UpdateCampaign
	// checking the parameters entered
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, WriteError("invalid parameters"))
		return
	}

	// fetch the campaign details
	campaign, err := app.store.CampaignInterace.GetCampaign(ctx, campaign_id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, WriteError("internal server error"))
		return
	}
	// Authorise user
	if campaign.BrandId != Entity.GetID() && Entity.GetRole() != "admin" {
		c.JSON(http.StatusUnauthorized, WriteError("unauthorised request"))
		return
	}

	// update request
	err = app.store.CampaignInterace.UpdateCampaign(ctx, campaign_id, payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, WriteError("server error"))
		return
	}
	campaign_response, _ := app.store.CampaignInterace.GetCampaign(ctx, campaign_id)
	// cache the updated campaigns
	if err = app.cache.SetCampaign(ctx, campaign_id, campaign_response); err != nil {
		log.Printf("error caching the campaign: %s", err.Error())
		app.cache.InvalidateCampaign(ctx, campaign_id)
	}
	// successfully updated the campaign
	c.JSON(http.StatusNoContent, WriteResponse(campaign_response))
}

func (app *Application) GetCampaignFeed(c *gin.Context) {
	ctx := c.Request.Context()
	limit, err := strconv.Atoi(c.Query("limit"))
	if err != nil {
		c.JSON(http.StatusBadRequest, WriteError("invalid query"))
		return
	}
	offset, err := strconv.Atoi(c.Query("offset"))
	if err != nil {
		c.JSON(http.StatusBadRequest, WriteError("invalid credentials"))
		return
	}
	output, err := app.store.CampaignInterace.GetRecentCampaigns(ctx, offset, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, WriteError("server error"))
		return
	}

	// return the campaign feed
	c.JSON(http.StatusOK, WriteResponse(output))
}

func (app *Application) GetCampaign(c *gin.Context) {
	ctx := c.Request.Context()
	campaign_id := c.Request.PathValue("campaign_id")
	if ok := uuid.Validate(campaign_id); ok != nil {
		c.JSON(http.StatusBadRequest, WriteError("invalid credentials"))
		return
	}

	// check cache
	var campaignResponse *db.Campaign
	err := app.cache.GetCampaign(ctx, campaign_id, campaignResponse)
	if err == nil {
		c.JSON(http.StatusOK, WriteResponse(campaignResponse))
		return
	}
	// cache miss
	campaignResponse, err = app.store.CampaignInterace.GetCampaign(ctx, campaign_id)
	if err != nil {
		if err == db.ErrNotFound {
			// bad request error
			c.JSON(http.StatusBadRequest, WriteError("invalid credentials"))
			return
		}
		// server error
		log.Printf("could not find campaign: %v", err.Error()) // log to fix the error
		c.JSON(http.StatusInternalServerError, WriteError("server error"))
		return
	}
	// cache the campaign
	if err = app.cache.SetCampaign(ctx, campaign_id, *campaignResponse); err != nil {
		log.Printf("error caching campaign: %s", err.Error())
	}
	// successfully fetched the campaign
	c.JSON(http.StatusOK, WriteResponse(campaignResponse))
}
