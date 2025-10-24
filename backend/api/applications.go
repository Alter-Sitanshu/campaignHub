package api

// Here the Application means the creator's application
// A creator can apply for a campaign on the portal. Just like LinkedIn

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"

	"github.com/Alter-Sitanshu/campaignHub/internals/db"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (app *Application) GetApplication(c *gin.Context) {
	ctx := c.Request.Context()
	// Get the logged in user and verify their status
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
	ID := c.Param("application_id")
	if ok := uuid.Validate(ID); ok != nil {
		log.Printf("invalid application id requested\n")
		c.JSON(http.StatusBadRequest, WriteError("invalid application"))
		return
	}
	// Get the application response from Database
	appl, err := app.store.ApplicationInterface.GetApplicationByID(ctx, ID)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			log.Printf("invalid application id requested\n")
			c.JSON(http.StatusBadRequest, WriteError("invalid application"))
			return
		default:
			log.Printf("server error\n")
			c.JSON(http.StatusInternalServerError, WriteError("internal server error"))
			return
		}
	}
	// User is allowed to get only their own applications
	if appl.CreatorId != Entity.GetID() {
		log.Printf("unauthorised application request: %v\n", LogInUser)
		c.JSON(http.StatusUnauthorized, WriteError("unauthorised. access denied"))
		return
	}

	// authorised application access granted
	c.JSON(http.StatusOK, WriteResponse(appl))
}

func (app *Application) CreateApplication(c *gin.Context) {
	ctx := c.Request.Context()
	// Get the logged in user and verify their status
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

	// Get the campaign_id from the Path
	CampaignId := c.Param("campaign_id")

	appl := db.CampaignApplication{
		Id:         uuid.New().String(),
		CreatorId:  Entity.GetID(),
		CampaignId: CampaignId,
		Status:     db.ApplicationPending,
	}

	// Create a Database entry for the application made
	err := app.store.ApplicationInterface.CreateApplication(ctx, appl)
	if err != nil {
		log.Printf("error creating application: %v", err.Error())
		c.JSON(http.StatusInternalServerError, WriteError("server error"))
		return
	}

	// Choice :- Should i mail the brand or not ?
	// Mailing the brand makes it cluttered ?
	// The brnads can check their dashboards for the campaign applications

	// Applicatioin submitted to the brand
	c.JSON(http.StatusCreated, WriteResponse("application submitted"))

}

func (app *Application) GetCampaignApplications(c *gin.Context) {
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
	campaignID := c.Param("campaign_id")
	// verify this campaign belongs to this brand
	camp, err := app.store.CampaignInterace.GetCampaign(ctx, campaignID)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			log.Printf("invalid campaign id requested\n")
			c.JSON(http.StatusBadRequest, WriteError("invalid application"))
			return
		default:
			log.Printf("server error\n")
			c.JSON(http.StatusInternalServerError, WriteError("internal server error"))
			return
		}
	}
	// the entity is not authotised to view the campaign applications
	if camp.BrandId != Entity.GetID() && Entity.GetRole() != "admin" {
		c.JSON(http.StatusUnauthorized, WriteError("unauthorised request"))
		return
	}
	limit, err := strconv.Atoi(c.Query("limit"))
	if err != nil {
		c.JSON(http.StatusBadRequest, WriteError("invalid query"))
		return
	}
	offset, err := strconv.Atoi(c.Query("offset"))
	if err != nil {
		c.JSON(http.StatusBadRequest, WriteError("invalid query"))
		return
	}
	appls, err := app.store.ApplicationInterface.GetCampaignApplications(
		ctx, campaignID, offset, limit,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, WriteError("internal server error"))
		return
	}

	// successfully got the campaign applications
	c.JSON(http.StatusOK, WriteResponse(appls))
}

func (app *Application) GetCreatorApplications(c *gin.Context) {
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
	CreatorID := Entity.GetID()
	limit, err := strconv.Atoi(c.Query("limit"))
	if err != nil {
		c.JSON(http.StatusBadRequest, WriteError("invalid query"))
		return
	}
	offset, err := strconv.Atoi(c.Query("offset"))
	if err != nil {
		c.JSON(http.StatusBadRequest, WriteError("invalid query"))
		return
	}
	appls, err := app.store.ApplicationInterface.GetCreatorApplications(
		ctx, CreatorID, offset, limit,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, WriteError("internal server error"))
		return
	}

	// successfully got the creator applications
	c.JSON(http.StatusOK, WriteResponse(appls))
}

func (app *Application) SetApplicationStatus(c *gin.Context) {
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

	applID := c.Param("application_id")
	if ok := uuid.Validate(applID); ok != nil {
		log.Printf("error: invalid application access: %v", applID)
		c.JSON(http.StatusBadRequest, WriteError("application does not exist"))
		return
	}

	status := c.Query("status")
	stat, err := strconv.Atoi(status)
	if err != nil {
		c.JSON(http.StatusBadRequest, WriteError("status must be integer"))
		return
	}
	// check if the application for the campaign belongs the brand
	appl, err := app.store.ApplicationInterface.GetApplicationByID(ctx, applID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, WriteError("server error. try again"))
		return
	}
	// get the brand of the campaign
	camp, err := app.store.CampaignInterace.GetCampaign(ctx, appl.CampaignId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, WriteError("server error. try again"))
		return
	}
	if camp.BrandId != Entity.GetID() && Entity.GetRole() != "admin" {
		c.JSON(http.StatusUnauthorized, WriteError("unauthorised access denied"))
		return
	}
	// try updating the status
	err = app.store.ApplicationInterface.SetApplicationStatus(ctx, applID, stat)
	if err != nil {
		c.JSON(http.StatusInternalServerError, WriteError(err.Error()))
		return
	}

	// success on updating the status code
	c.JSON(http.StatusNoContent, WriteResponse("status set successful"))
}

func (app *Application) DeleteApplication(c *gin.Context) {
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
		c.JSON(http.StatusUnauthorized, WriteError("unauthorised access denied"))
		return
	}
	// get the application id
	applID := c.Param("application_id")
	if ok := uuid.Validate(applID); ok != nil {
		log.Printf("error: invalid application access: %v", applID)
		c.JSON(http.StatusBadRequest, WriteError("application does not exist"))
		return
	}
	// try the deletion
	err := app.store.ApplicationInterface.DeleteApplication(ctx, applID)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			c.JSON(http.StatusBadRequest, WriteError("invalid application requested"))
			return
		default:
			c.JSON(http.StatusInternalServerError, WriteError("server error. try again"))
			return
		}
	}

	// successfully deleted the application foem from the database
	c.JSON(http.StatusNoContent, WriteResponse("applicaion deleted"))
}
