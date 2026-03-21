package client

// Base URLs and path prefixes for the LinkedIn Voyager API.
const (
	BaseURL    = "https://www.linkedin.com"
	VoyagerAPI = "/voyager/api"

	// Auth
	EndpointAuthenticate = "/uas/authenticate"

	// Identity / Profiles
	EndpointProfiles       = VoyagerAPI + "/identity/profiles"
	EndpointDashProfiles   = VoyagerAPI + "/identity/dash/profiles"
	EndpointProfileContact = VoyagerAPI + "/identity/profiles/%s/profileContactInfo"
	EndpointProfileView    = VoyagerAPI + "/identity/profiles/%s/profileView"
	EndpointMe             = VoyagerAPI + "/me"

	// Profile sections (write)
	EndpointProfilePositions      = VoyagerAPI + "/identity/profiles/%s/positions"
	EndpointProfilePosition       = VoyagerAPI + "/identity/profiles/%s/positions/%s"
	EndpointProfileEducations     = VoyagerAPI + "/identity/profiles/%s/educations"
	EndpointProfileEducation      = VoyagerAPI + "/identity/profiles/%s/educations/%s"
	EndpointProfileSkillsPath     = VoyagerAPI + "/identity/profiles/%s/skills"
	EndpointProfileSkillPath      = VoyagerAPI + "/identity/profiles/%s/skills/%s"
	EndpointProfileCertifications = VoyagerAPI + "/identity/profiles/%s/certifications"
	EndpointProfileCertification  = VoyagerAPI + "/identity/profiles/%s/certifications/%s"
	EndpointProfileLanguages      = VoyagerAPI + "/identity/profiles/%s/languages"
	EndpointProfileLanguage       = VoyagerAPI + "/identity/profiles/%s/languages/%s"
	EndpointProfileVolunteer      = VoyagerAPI + "/identity/profiles/%s/volunteerExperiences"
	EndpointProfileVolunteerItem  = VoyagerAPI + "/identity/profiles/%s/volunteerExperiences/%s"
	EndpointProfileProjects       = VoyagerAPI + "/identity/profiles/%s/projects"
	EndpointProfileProject        = VoyagerAPI + "/identity/profiles/%s/projects/%s"
	EndpointProfilePublications   = VoyagerAPI + "/identity/profiles/%s/publications"
	EndpointProfilePublication    = VoyagerAPI + "/identity/profiles/%s/publications/%s"
	EndpointProfileHonors         = VoyagerAPI + "/identity/profiles/%s/honors"
	EndpointProfileHonor          = VoyagerAPI + "/identity/profiles/%s/honors/%s"
	EndpointProfileCourses        = VoyagerAPI + "/identity/profiles/%s/courses"
	EndpointProfileCourse         = VoyagerAPI + "/identity/profiles/%s/courses/%s"
	EndpointProfilePhotoRegister  = VoyagerAPI + "/assets?action=registerUpload"
	EndpointProfilePhoto          = VoyagerAPI + "/identity/profiles/%s/profilePictures"

	// Open to Work / Job Seeking
	EndpointJobSeekingProfiles = VoyagerAPI + "/identity/jobSeekingProfiles"
	EndpointJobSeekingProfile  = VoyagerAPI + "/identity/jobSeekingProfiles/%s"

	// GraphQL (used by newer endpoints)
	EndpointGraphQL = VoyagerAPI + "/graphql"

	// Who Viewed Profile
	// The total viewer count comes from the feed identity module widget (free accounts).
	// queryId for voyagerFeedDashIdentityModule — look for widgetType "WHO_VIEWED_MY_PROFILE".
	EndpointFeedIdentityModuleQueryID = "voyagerFeedDashIdentityModule.803fe19f843a4d461478049f70d7babd"
	// Individual viewer names require LinkedIn Premium (voyagerPremiumDashAnalyticsObject).
	EndpointWVMPCards = VoyagerAPI + "/identity/wvmpCards" // deprecated — kept for mock compatibility

	// Connections & Relationships
	EndpointInvitationViewsQueryID     = "voyagerRelationshipsDashInvitationViews.57e1286f887065b96393b947e09ef04c"
	EndpointInvitationsSummaryQueryID  = "voyagerRelationshipsDashInvitationsSummary.26002c38d857d2d5cd4503df1a43a0ab"
	EndpointSentInvitationViewsQueryID = "voyagerRelationshipsDashSentInvitationViews.1901307baa315a33bf17bb743daf1250"
	EndpointConnections                = VoyagerAPI + "/relationships/connections"
	EndpointDashConnections            = VoyagerAPI + "/relationships/dash/connections"
	EndpointConnection        = VoyagerAPI + "/relationships/connections/%s"
	EndpointInvitations       = VoyagerAPI + "/relationships/invitationViews"
	EndpointSentInvitations   = VoyagerAPI + "/relationships/sentInvitationViewsV2"
	EndpointInvitationHandle  = VoyagerAPI + "/relationships/invitations/%s"
	EndpointFollowEntity      = VoyagerAPI + "/feed/follows"
	EndpointMutualConnections = VoyagerAPI + "/relationships/connectionOf"

	// Messaging (GraphQL)
	EndpointMessengerConversationsQueryID = "voyagerMessagingDashMessengerConversations.ccc086e11ebcecef63b31ac465ccfebd"
	EndpointMessengerMessagesQueryID      = "voyagerMessagingDashMessengerMessages.073958b6fdfe5f5ceeb4d0416523317e"
	EndpointMessengerMailboxCountsQueryID = "voyagerMessagingDashMessengerMailboxCounts.15769ef365ec721fc539d76dbef5f813"

	// Messaging (legacy — deprecated, kept for reference)
	EndpointConversations      = VoyagerAPI + "/messaging/conversations"
	EndpointConversationByID   = VoyagerAPI + "/messaging/conversations/%s"
	EndpointConversationEvents = VoyagerAPI + "/messaging/conversations/%s/events"
	EndpointMessageEventByID   = VoyagerAPI + "/messaging/conversations/%s/events/%s"
	EndpointMessageCreate      = VoyagerAPI + "/messaging/conversations"

	// Jobs (GraphQL)
	EndpointJobsFeedQueryID     = "voyagerJobsDashJobsFeed.40bc6ea7c5b88757481d40f6e4527f17"
	EndpointJobCardsQueryID     = "voyagerJobsDashJobCards.7fb7b035d6233f835789e4088cdbf44b"
	EndpointJobPostingsQueryID  = "voyagerJobsDashJobPostings.891aed7916d7453a37e4bbf5f1f60de4"

	// Jobs (legacy — deprecated, kept for save/unsave which still use REST)
	EndpointJobSearch          = VoyagerAPI + "/jobs/jobPostings"
	EndpointJobPosting         = VoyagerAPI + "/jobs/jobPostings/%s"
	EndpointSavedJobs          = VoyagerAPI + "/jobs/jobSaves"
	EndpointAppliedJobs        = VoyagerAPI + "/jobs/appliedJobs"
	EndpointJobRecommendations = VoyagerAPI + "/jobs/jobRecommendations"

	// Companies / Organizations
	EndpointCompanies      = VoyagerAPI + "/organization/companies"
	EndpointCompany        = VoyagerAPI + "/organization/companies/%s"
	EndpointCompanyUpdates = VoyagerAPI + "/feed/updates"

	// Posts / Feed
	EndpointFeed          = VoyagerAPI + "/feed/updatesV2"
	EndpointPostCreate    = VoyagerAPI + "/ugcPosts"
	EndpointUGCPost       = VoyagerAPI + "/ugcPosts/%s"
	EndpointSocialActions = VoyagerAPI + "/socialActions/%s"
	EndpointLike          = VoyagerAPI + "/socialActions/%s/likes"
	EndpointComments      = VoyagerAPI + "/socialActions/%s/comments"
	EndpointCommentByID   = VoyagerAPI + "/socialActions/%s/comments/%s"
	EndpointCommentLike   = VoyagerAPI + "/socialActions/%s/comments/%s/likes"
	EndpointShares        = VoyagerAPI + "/shares"
	EndpointMediaUpload   = VoyagerAPI + "/assets"

	// Search (GraphQL)
	EndpointSearchClustersQueryID = "voyagerSearchDashClusters.05111e1b90ee7fea15bebe9f9410ced9"

	// Search (legacy — deprecated)
	EndpointSearch        = VoyagerAPI + "/search/hits"
	EndpointSearchBlended = VoyagerAPI + "/search/blended"
	EndpointJobSearchDash = VoyagerAPI + "/jobs/search"

	// Recommendations
	EndpointRecommendations     = VoyagerAPI + "/identity/recommendations"
	EndpointRecommendationGiven = VoyagerAPI + "/identity/recommendations?q=given"
	EndpointRecommendationByID  = VoyagerAPI + "/identity/recommendations/%s"

	// Notifications (dash)
	EndpointDashNotificationCards = VoyagerAPI + "/voyagerIdentityDashNotificationCards"
	EndpointDashBadgingCounts     = VoyagerAPI + "/voyagerNotificationsDashBadgingItemCounts"

	// Notifications (legacy — deprecated)
	EndpointNotifications     = VoyagerAPI + "/feed/notifications"
	EndpointNotificationBadge = VoyagerAPI + "/feed/notificationBadge"
)

// Default pagination values.
const (
	DefaultStart = 0
	DefaultCount = 20
	MaxCount     = 100
)
