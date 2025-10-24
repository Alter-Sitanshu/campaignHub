package cache

import "fmt"

// Key patterns for different data types
const (
	keySubmissionEarnings  = "earnings:%s"
	keyCampaignBudget      = "budget:%s"
	keyCampaign            = "campaign:%s"
	keyActiveCampaigns     = "campaigns:active"
	keyCreatorSubmissions  = "submissions:creator:%s"
	keyBrandCampaigns      = "campaigns:brand:%s"
	keyUserCampaigns       = "campaigns:user:%s"
	keyUserBalance         = "balance:%s"
	keyUserProfile         = "user:%s"
	keySubmissionStatus    = "status:%s"
	keyPendingApplications = "applications:pending:%s"
	keyVideoMetaData       = "video:metadata:%s"
	batchQueueKey          = "queue:batch:updates"
)

// Key builders
func VideoMetadataKey(submissionID string) string {
	return fmt.Sprintf(keyVideoMetaData, submissionID)
}

func SubmissionEarningsKey(submissionID string) string {
	return fmt.Sprintf(keySubmissionEarnings, submissionID)
}

func CampaignBudgetKey(campaignID string) string {
	return fmt.Sprintf(keyCampaignBudget, campaignID)
}

func CampaignKey(campaignID string) string {
	return fmt.Sprintf(keyCampaign, campaignID)
}

func CreatorSubmissionsKey(creatorID string) string {
	return fmt.Sprintf(keyCreatorSubmissions, creatorID)
}

func CompanyCampaignsKey(companyID string) string {
	return fmt.Sprintf(keyBrandCampaigns, companyID)
}

func UserCampaignsKey(userID string) string {
	return fmt.Sprintf(keyUserCampaigns, userID)
}

func UserBalanceKey(userID string) string {
	return fmt.Sprintf(keyUserBalance, userID)
}

func UserProfileKey(userID string) string {
	return fmt.Sprintf(keyUserProfile, userID)
}

func SubmissionStatusKey(submissionID string) string {
	return fmt.Sprintf(keySubmissionStatus, submissionID)
}

func PendingApplicationsKey(companyID string) string {
	return fmt.Sprintf(keyPendingApplications, companyID)
}

func ActiveCampaignsKey() string {
	return keyActiveCampaigns
}
