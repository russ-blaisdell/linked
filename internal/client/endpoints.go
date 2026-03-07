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

	// Who Viewed Profile
	EndpointWVMPCards = VoyagerAPI + "/identity/wvmpCards"

	// Connections & Relationships
	EndpointConnections       = VoyagerAPI + "/relationships/connections"
	EndpointConnection        = VoyagerAPI + "/relationships/connections/%s"
	EndpointInvitations       = VoyagerAPI + "/relationships/invitationViews"
	EndpointSentInvitations   = VoyagerAPI + "/relationships/sentInvitationViewsV2"
	EndpointInvitationHandle  = VoyagerAPI + "/relationships/invitations/%s"
	EndpointFollowEntity      = VoyagerAPI + "/feed/follows"
	EndpointMutualConnections = VoyagerAPI + "/relationships/connectionOf"

	// Messaging
	EndpointConversations      = VoyagerAPI + "/messaging/conversations"
	EndpointConversationByID   = VoyagerAPI + "/messaging/conversations/%s"
	EndpointConversationEvents = VoyagerAPI + "/messaging/conversations/%s/events"
	EndpointMessageEventByID   = VoyagerAPI + "/messaging/conversations/%s/events/%s"
	EndpointMessageCreate      = VoyagerAPI + "/messaging/conversations"

	// Jobs
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

	// Search
	EndpointSearch        = VoyagerAPI + "/search/hits"
	EndpointSearchBlended = VoyagerAPI + "/search/blended"
	EndpointJobSearchDash = VoyagerAPI + "/jobs/search"

	// Recommendations
	EndpointRecommendations     = VoyagerAPI + "/identity/recommendations"
	EndpointRecommendationGiven = VoyagerAPI + "/identity/recommendations?q=given"
	EndpointRecommendationByID  = VoyagerAPI + "/identity/recommendations/%s"

	// Notifications
	EndpointNotifications     = VoyagerAPI + "/feed/notifications"
	EndpointNotificationBadge = VoyagerAPI + "/feed/notificationBadge"
)

// Default pagination values.
const (
	DefaultStart = 0
	DefaultCount = 20
	MaxCount     = 100
)
