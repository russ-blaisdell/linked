package client

// Base URLs and path prefixes for the LinkedIn Voyager API.
const (
	BaseURL    = "https://www.linkedin.com"
	VoyagerAPI = "/voyager/api"

	// Auth
	EndpointAuthenticate = "/uas/authenticate"

	// Identity / Profiles
	EndpointProfiles       = VoyagerAPI + "/identity/profiles"
	EndpointProfileContact = VoyagerAPI + "/identity/profiles/%s/profileContactInfo"
	EndpointProfileView    = VoyagerAPI + "/identity/profiles/%s/profileView"
	EndpointMe             = VoyagerAPI + "/me"

	// Connections & Relationships
	EndpointConnections         = VoyagerAPI + "/relationships/connections"
	EndpointInvitations         = VoyagerAPI + "/relationships/invitationViews"
	EndpointSentInvitations     = VoyagerAPI + "/relationships/sentInvitationViewsV2"
	EndpointInvitationHandle    = VoyagerAPI + "/relationships/invitations/%s"
	EndpointFollowEntity        = VoyagerAPI + "/feed/follows"

	// Messaging
	EndpointConversations       = VoyagerAPI + "/messaging/conversations"
	EndpointConversationEvents  = VoyagerAPI + "/messaging/conversations/%s/events"
	EndpointMessageCreate       = VoyagerAPI + "/messaging/conversations"

	// Jobs
	EndpointJobSearch           = VoyagerAPI + "/jobs/jobPostings"
	EndpointJobPosting          = VoyagerAPI + "/jobs/jobPostings/%s"
	EndpointSavedJobs           = VoyagerAPI + "/jobs/jobSaves"
	EndpointAppliedJobs         = VoyagerAPI + "/jobs/appliedJobs"

	// Companies / Organizations
	EndpointCompanies           = VoyagerAPI + "/organization/companies"
	EndpointCompany             = VoyagerAPI + "/organization/companies/%s"
	EndpointCompanyUpdates      = VoyagerAPI + "/feed/updates"

	// Posts / Feed
	EndpointFeed                = VoyagerAPI + "/feed/updatesV2"
	EndpointPostCreate          = VoyagerAPI + "/ugcPosts"
	EndpointSocialActions       = VoyagerAPI + "/socialActions/%s"
	EndpointLike                = VoyagerAPI + "/socialActions/%s/likes"
	EndpointComments            = VoyagerAPI + "/socialActions/%s/comments"
	EndpointShares              = VoyagerAPI + "/shares"

	// Search
	EndpointSearch              = VoyagerAPI + "/search/hits"
	EndpointSearchBlended       = VoyagerAPI + "/search/blended"
	EndpointJobSearchDash       = VoyagerAPI + "/jobs/search"

	// Recommendations
	EndpointRecommendations     = VoyagerAPI + "/identity/recommendations"
	EndpointRecommendationGiven = VoyagerAPI + "/identity/recommendations?q=given"

	// Notifications
	EndpointNotifications       = VoyagerAPI + "/feed/notifications"
)

// Default pagination values.
const (
	DefaultStart = 0
	DefaultCount = 20
	MaxCount     = 100
)
