package models

// Profile represents a LinkedIn member profile.
type Profile struct {
	URN            string          `json:"urn"`
	ProfileID      string          `json:"profileId"`
	FirstName      string          `json:"firstName"`
	LastName       string          `json:"lastName"`
	Headline       string          `json:"headline"`
	Summary        string          `json:"summary"`
	Location       string          `json:"location"`
	Industry       string          `json:"industry"`
	PhotoURL       string          `json:"photoUrl,omitempty"`
	PublicURL      string          `json:"publicUrl,omitempty"`
	Connection     string          `json:"connection,omitempty"` // FIRST_DEGREE, SECOND_DEGREE, etc.
	Experience     []Experience    `json:"experience,omitempty"`
	Education      []Education     `json:"education,omitempty"`
	Skills         []Skill         `json:"skills,omitempty"`
	Languages      []Language      `json:"languages,omitempty"`
	Certifications []Certification `json:"certifications,omitempty"`
	Publications   []Publication   `json:"publications,omitempty"`
	Patents        []Patent        `json:"patents,omitempty"`
}

// ContactInfo holds profile contact details.
type ContactInfo struct {
	ProfileID      string   `json:"profileId"`
	Emails         []string `json:"emails,omitempty"`
	PhoneNumbers   []string `json:"phoneNumbers,omitempty"`
	TwitterHandles []string `json:"twitterHandles,omitempty"`
	Websites       []string `json:"websites,omitempty"`
	Address        string   `json:"address,omitempty"`
}

// Experience is a position on a profile.
type Experience struct {
	ID          string `json:"id,omitempty"`
	Title       string `json:"title"`
	CompanyName string `json:"companyName"`
	CompanyURN  string `json:"companyUrn,omitempty"`
	StartDate   string `json:"startDate,omitempty"`
	EndDate     string `json:"endDate,omitempty"`
	Current     bool   `json:"current"`
	Description string `json:"description,omitempty"`
	Location    string `json:"location,omitempty"`
}

// Education is a school/degree on a profile.
type Education struct {
	ID           string `json:"id,omitempty"`
	SchoolName   string `json:"schoolName"`
	SchoolURN    string `json:"schoolUrn,omitempty"`
	Degree       string `json:"degree,omitempty"`
	FieldOfStudy string `json:"fieldOfStudy,omitempty"`
	StartDate    string `json:"startDate,omitempty"`
	EndDate      string `json:"endDate,omitempty"`
	Description  string `json:"description,omitempty"`
}

// Skill is a skill listed on a profile.
type Skill struct {
	ID           string `json:"id,omitempty"`
	Name         string `json:"name"`
	Endorsements int    `json:"endorsements"`
}

// Language is a language listed on a profile.
type Language struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name"`
	Proficiency string `json:"proficiency,omitempty"`
}

// Certification is a professional certification on a profile.
type Certification struct {
	ID         string `json:"id,omitempty"`
	Name       string `json:"name"`
	Authority  string `json:"authority,omitempty"`
	LicenseNum string `json:"licenseNumber,omitempty"`
	URL        string `json:"url,omitempty"`
	StartDate  string `json:"startDate,omitempty"`
	EndDate    string `json:"endDate,omitempty"`
}

// VolunteerExperience is a volunteer position on a profile.
type VolunteerExperience struct {
	ID           string `json:"id,omitempty"`
	Role         string `json:"role"`
	Organization string `json:"organization"`
	Cause        string `json:"cause,omitempty"`
	StartDate    string `json:"startDate,omitempty"`
	EndDate      string `json:"endDate,omitempty"`
	Current      bool   `json:"current"`
	Description  string `json:"description,omitempty"`
}

// Project is a project on a profile.
type Project struct {
	ID          string `json:"id,omitempty"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	URL         string `json:"url,omitempty"`
	StartDate   string `json:"startDate,omitempty"`
	EndDate     string `json:"endDate,omitempty"`
	Current     bool   `json:"current"`
}

// Publication is a publication on a profile.
type Publication struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name"`
	Publisher   string `json:"publisher,omitempty"`
	PublishedOn string `json:"publishedOn,omitempty"`
	URL         string `json:"url,omitempty"`
	Description string `json:"description,omitempty"`
}

// Patent is a patent on a profile.
type Patent struct {
	ID          string `json:"id,omitempty"`
	Title       string `json:"title"`
	Issuer      string `json:"issuer,omitempty"`
	Number      string `json:"number,omitempty"`
	URL         string `json:"url,omitempty"`
	Description string `json:"description,omitempty"`
}

// Honor is an award or honor on a profile.
type Honor struct {
	ID          string `json:"id,omitempty"`
	Title       string `json:"title"`
	Issuer      string `json:"issuer,omitempty"`
	IssuedOn    string `json:"issuedOn,omitempty"`
	Description string `json:"description,omitempty"`
}

// Course is a course on a profile.
type Course struct {
	ID         string `json:"id,omitempty"`
	Name       string `json:"name"`
	Number     string `json:"number,omitempty"`
	Occupation string `json:"occupation,omitempty"`
}

// OpenToWork holds the user's job-seeking preferences.
type OpenToWork struct {
	ID              string   `json:"id,omitempty"`
	Active          bool     `json:"active"`
	JobTypes        []string `json:"jobTypes,omitempty"`
	Locations       []string `json:"locations,omitempty"`
	Title           string   `json:"title,omitempty"`
	PreferenceTypes []string `json:"preferenceTypes,omitempty"`
}

// ProfileViewer is someone who viewed your profile.
type ProfileViewer struct {
	Profile   Profile `json:"profile"`
	ViewedAt  string  `json:"viewedAt"`
	ViewCount int     `json:"viewCount,omitempty"`
}

// NotificationBadge holds the unread notification count.
type NotificationBadge struct {
	UnreadCount int `json:"unreadCount"`
}

// MutualConnection is a shared connection between two members.
type MutualConnection struct {
	Profile Profile `json:"profile"`
	Count   int     `json:"count"`
}

// Connection is a 1st-degree LinkedIn connection.
type Connection struct {
	ProfileID string `json:"profileId"`
	URN       string `json:"urn"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Headline  string `json:"headline"`
	PhotoURL  string `json:"photoUrl,omitempty"`
}

// Invitation is a pending connection invite (sent or received).
type Invitation struct {
	ID          string  `json:"id"`
	URN         string  `json:"urn"`
	FromProfile Profile `json:"fromProfile,omitempty"`
	ToProfile   Profile `json:"toProfile,omitempty"`
	Message     string  `json:"message,omitempty"`
	Insight     string  `json:"insight,omitempty"` // e.g. "1 mutual connection"
	SentAt      string  `json:"sentAt"`
	Status      string  `json:"status"`    // PENDING, ACCEPTED, IGNORED
	Direction   string  `json:"direction"` // INBOUND, OUTBOUND
}

// Conversation is a LinkedIn messaging thread.
type Conversation struct {
	ID           string    `json:"id"`
	URN          string    `json:"urn"`
	Participants []Profile `json:"participants"`
	LastMessage  *Message  `json:"lastMessage,omitempty"`
	Unread       bool      `json:"unread"`
	Starred      bool      `json:"starred"`
	Archived     bool      `json:"archived"`
	UpdatedAt    string    `json:"updatedAt"`
}

// Message is a single message in a conversation.
type Message struct {
	ID            string  `json:"id"`
	URN           string  `json:"urn"`
	Body          string  `json:"body"`
	SenderProfile Profile `json:"senderProfile"`
	SentAt        string  `json:"sentAt"`
	DeliveredAt   string  `json:"deliveredAt,omitempty"`
}

// Job is a LinkedIn job posting.
type Job struct {
	ID              string  `json:"id"`
	URN             string  `json:"urn"`
	Title           string  `json:"title"`
	Company         Company `json:"company"`
	Location        string  `json:"location"`
	Remote          bool    `json:"remote"`
	PostedAt        string  `json:"postedAt"`
	ExpiresAt       string  `json:"expiresAt,omitempty"`
	Description     string  `json:"description,omitempty"`
	ApplyURL        string  `json:"applyUrl,omitempty"`
	Saved           bool    `json:"saved"`
	Applied         bool    `json:"applied"`
	WorkplaceType   string  `json:"workplaceType,omitempty"` // ONSITE, REMOTE, HYBRID
	EmploymentType  string  `json:"employmentType,omitempty"`
	ExperienceLevel string  `json:"experienceLevel,omitempty"`
}

// Company is a LinkedIn organization.
type Company struct {
	ID            string `json:"id"`
	URN           string `json:"urn"`
	Name          string `json:"name"`
	Headline      string `json:"headline,omitempty"`
	Description   string `json:"description,omitempty"`
	Industry      string `json:"industry,omitempty"`
	Website       string `json:"website,omitempty"`
	Headquarters  string `json:"headquarters,omitempty"`
	EmployeeCount string `json:"employeeCount,omitempty"`
	LogoURL       string `json:"logoUrl,omitempty"`
	Following     bool   `json:"following"`
}

// Post is a LinkedIn content post/share.
type Post struct {
	URN           string   `json:"urn"`
	AuthorProfile Profile  `json:"authorProfile"`
	Body          string   `json:"body"`
	LikeCount     int      `json:"likeCount"`
	CommentCount  int      `json:"commentCount"`
	ShareCount    int      `json:"shareCount"`
	PostedAt      string   `json:"postedAt"`
	Liked         bool     `json:"liked"`
	ImageURLs     []string `json:"imageUrls,omitempty"`
}

// Comment is a comment on a LinkedIn post.
type Comment struct {
	URN           string  `json:"urn"`
	AuthorProfile Profile `json:"authorProfile"`
	Body          string  `json:"body"`
	LikeCount     int     `json:"likeCount"`
	PostedAt      string  `json:"postedAt"`
}

// Recommendation is a recommendation given or received.
type Recommendation struct {
	ID                 string  `json:"id"`
	URN                string  `json:"urn"`
	RecommenderProfile Profile `json:"recommenderProfile"`
	RecommendeeProfile Profile `json:"recommendeeProfile"`
	Body               string  `json:"body"`
	Relationship       string  `json:"relationship,omitempty"`
	CreatedAt          string  `json:"createdAt"`
	Status             string  `json:"status"` // VISIBLE, HIDDEN, PENDING
}

// Notification is a LinkedIn notification.
type Notification struct {
	ID        string `json:"id"`
	URN       string `json:"urn"`
	Type      string `json:"type"`
	Body      string `json:"body"`
	Read      bool   `json:"read"`
	CreatedAt string `json:"createdAt"`
	EntityURN string `json:"entityUrn,omitempty"`
}

// SearchPeopleResult is a single hit from a people search.
type SearchPeopleResult struct {
	Profile           Profile `json:"profile"`
	Distance          string  `json:"distance"` // DISTANCE_1, DISTANCE_2, DISTANCE_3
	SharedConnections int     `json:"sharedConnections"`
}

// Pagination holds paging state for list responses.
type Pagination struct {
	Start   int  `json:"start"`
	Count   int  `json:"count"`
	Total   int  `json:"total"`
	HasMore bool `json:"hasMore"`
}

// Paged wrappers for each domain type.

type PagedProfiles struct {
	Items      []Profile  `json:"items"`
	Pagination Pagination `json:"pagination"`
}

type PagedConnections struct {
	Items      []Connection `json:"items"`
	Pagination Pagination   `json:"pagination"`
}

type PagedInvitations struct {
	Items      []Invitation `json:"items"`
	Pagination Pagination   `json:"pagination"`
}

type PagedConversations struct {
	Items      []Conversation `json:"items"`
	Pagination Pagination     `json:"pagination"`
}

type PagedMessages struct {
	Items      []Message  `json:"items"`
	Pagination Pagination `json:"pagination"`
}

type PagedJobs struct {
	Items      []Job      `json:"items"`
	Pagination Pagination `json:"pagination"`
}

type PagedCompanies struct {
	Items      []Company  `json:"items"`
	Pagination Pagination `json:"pagination"`
}

type PagedPosts struct {
	Items      []Post     `json:"items"`
	Pagination Pagination `json:"pagination"`
}

type PagedSearchPeople struct {
	Items      []SearchPeopleResult `json:"items"`
	Pagination Pagination           `json:"pagination"`
}

type PagedRecommendations struct {
	Items      []Recommendation `json:"items"`
	Pagination Pagination       `json:"pagination"`
}

type PagedNotifications struct {
	Items      []Notification `json:"items"`
	Pagination Pagination     `json:"pagination"`
}

type PagedExperience struct {
	Items      []Experience `json:"items"`
	Pagination Pagination   `json:"pagination"`
}

type PagedEducation struct {
	Items      []Education `json:"items"`
	Pagination Pagination  `json:"pagination"`
}

type PagedSkills struct {
	Items      []Skill    `json:"items"`
	Pagination Pagination `json:"pagination"`
}

type PagedCertifications struct {
	Items      []Certification `json:"items"`
	Pagination Pagination      `json:"pagination"`
}

type PagedLanguages struct {
	Items      []Language `json:"items"`
	Pagination Pagination `json:"pagination"`
}

type PagedVolunteer struct {
	Items      []VolunteerExperience `json:"items"`
	Pagination Pagination            `json:"pagination"`
}

type PagedProjects struct {
	Items      []Project  `json:"items"`
	Pagination Pagination `json:"pagination"`
}

type PagedPublications struct {
	Items      []Publication `json:"items"`
	Pagination Pagination    `json:"pagination"`
}

type PagedHonors struct {
	Items      []Honor    `json:"items"`
	Pagination Pagination `json:"pagination"`
}

type PagedCourses struct {
	Items      []Course   `json:"items"`
	Pagination Pagination `json:"pagination"`
}

type PagedProfileViewers struct {
	Items      []ProfileViewer `json:"items"`
	Pagination Pagination      `json:"pagination"`
}

type PagedMutualConnections struct {
	Items      []MutualConnection `json:"items"`
	Pagination Pagination         `json:"pagination"`
}

// BrowserFingerprint holds browser identity headers extracted from a HAR file.
// When present, these override the hardcoded defaults in the HTTP client.
type BrowserFingerprint struct {
	UserAgent       string `json:"userAgent,omitempty"`
	SecChUA         string `json:"secChUa,omitempty"`
	SecChUAMobile   string `json:"secChUaMobile,omitempty"`
	SecChUAPlatform string `json:"secChUaPlatform,omitempty"`
	XLiTrack        string `json:"xLiTrack,omitempty"`
}

// Credentials holds the stored LinkedIn session tokens.
type Credentials struct {
	LiAt        string              `json:"li_at"`
	JSESSIONID  string              `json:"jsessionid"`
	Bcookie     string              `json:"bcookie,omitempty"`
	Bscookie    string              `json:"bscookie,omitempty"`
	ProfileID   string              `json:"profileId,omitempty"`
	CreatedAt   string              `json:"createdAt"`
	Fingerprint *BrowserFingerprint `json:"fingerprint,omitempty"`
}
