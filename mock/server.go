// Package mock provides a fake LinkedIn Voyager API server for testing.
// It serves realistic JSON fixtures and validates request auth headers,
// allowing full integration tests with no network calls.
package mock

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

const (
	testLiAt       = "test-li-at-token"
	testJSESSIONID = "ajax:test-jsessionid"
)

// TestCredentials returns the credentials the mock server accepts.
func TestCredentials() (liAt, jsessionid string) {
	return testLiAt, testJSESSIONID
}

// sentMessage records a message sent in a test.
type sentMessage struct {
	Path    string
	Payload map[string]interface{}
}

// Server is the mock LinkedIn Voyager API.
type Server struct {
	HTTPServer   *httptest.Server
	mu           sync.Mutex
	sentMessages []sentMessage
	createdPosts []map[string]interface{}
	savedJobs    map[string]bool
	likes        map[string]bool
	following    map[string]bool
}

// New creates and starts a mock LinkedIn server.
func New() *Server {
	s := &Server{
		savedJobs: make(map[string]bool),
		likes:     make(map[string]bool),
		following: make(map[string]bool),
	}
	mux := http.NewServeMux()
	s.registerRoutes(mux)
	s.HTTPServer = httptest.NewServer(s.authMiddleware(mux))
	return s
}

// URL returns the base URL of the mock server.
func (s *Server) URL() string {
	return s.HTTPServer.URL
}

// Close shuts down the mock server.
func (s *Server) Close() {
	s.HTTPServer.Close()
}

// SentMessages returns messages sent during the test.
func (s *Server) SentMessages() []sentMessage {
	s.mu.Lock()
	defer s.mu.Unlock()
	cp := make([]sentMessage, len(s.sentMessages))
	copy(cp, s.sentMessages)
	return cp
}

// CreatedPosts returns posts created during the test.
func (s *Server) CreatedPosts() []map[string]interface{} {
	s.mu.Lock()
	defer s.mu.Unlock()
	cp := make([]map[string]interface{}, len(s.createdPosts))
	copy(cp, s.createdPosts)
	return cp
}

// authMiddleware validates that requests carry the expected CSRF token.
func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		csrf := r.Header.Get("csrf-token")
		if csrf != testJSESSIONID {
			http.Error(w, `{"status":401,"message":"Unauthorized"}`, http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// registerRoutes wires all mock endpoint handlers.
// Note: Go's default ServeMux uses longest-prefix matching so more-specific
// paths must be registered first.
func (s *Server) registerRoutes(mux *http.ServeMux) {
	// Profile
	mux.HandleFunc("/voyager/api/me", s.handleMe)
	mux.HandleFunc("/voyager/api/identity/profiles/", s.handleProfiles)

	// Recommendations — specific path before generic
	mux.HandleFunc("/voyager/api/identity/recommendations/", s.handleRecommendationByID)
	mux.HandleFunc("/voyager/api/identity/recommendations", s.handleRecommendations)

	// Messaging
	mux.HandleFunc("/voyager/api/messaging/conversations/", s.handleConversationSub)
	mux.HandleFunc("/voyager/api/messaging/conversations", s.handleConversations)
	mux.HandleFunc("/voyager/api/voyagerMessagingGraphQL/graphql", s.handleMessagingGraphQL)

	// Connections
	mux.HandleFunc("/voyager/api/relationships/connections", s.handleConnections)
	mux.HandleFunc("/voyager/api/relationships/dash/connections", s.handleDashConnections)

	// Profiles (dash)
	mux.HandleFunc("/voyager/api/identity/dash/profiles/", s.handleDashProfiles)
	mux.HandleFunc("/voyager/api/identity/dash/profiles", s.handleDashProfiles)
	mux.HandleFunc("/voyager/api/relationships/invitationViews", s.handleInvitations)
	mux.HandleFunc("/voyager/api/relationships/sentInvitationViewsV2", s.handleSentInvitations)
	mux.HandleFunc("/voyager/api/relationships/invitations/", s.handleInvitationAction)

	// Follows — with and without trailing ID
	mux.HandleFunc("/voyager/api/feed/follows/", s.handleFollowByID)
	mux.HandleFunc("/voyager/api/feed/follows", s.handleFollows)

	// Jobs
	mux.HandleFunc("/voyager/api/jobs/jobPostings/", s.handleJobPosting)
	mux.HandleFunc("/voyager/api/jobs/jobSaves/", s.handleJobSaveByID)
	mux.HandleFunc("/voyager/api/jobs/jobSaves", s.handleJobSaves)
	mux.HandleFunc("/voyager/api/jobs/search", s.handleJobSearch)
	mux.HandleFunc("/voyager/api/jobs/appliedJobs", s.handleAppliedJobs)

	// Companies
	mux.HandleFunc("/voyager/api/organization/companies", s.handleCompanies)

	// Posts / Feed
	mux.HandleFunc("/voyager/api/feed/updatesV2", s.handleFeed)
	mux.HandleFunc("/voyager/api/feed/notifications/", s.handleNotificationByID)
	mux.HandleFunc("/voyager/api/feed/notifications", s.handleNotifications)
	mux.HandleFunc("/voyager/api/voyagerIdentityDashNotificationCards", s.handleDashNotificationCards)
	mux.HandleFunc("/voyager/api/voyagerNotificationsDashBadgingItemCounts", s.handleDashBadgingCounts)
	mux.HandleFunc("/voyager/api/voyagerMessagingDashMessengerMessages", s.handleDashMessengerMessages)
	mux.HandleFunc("/voyager/api/ugcPosts", s.handleUGCPosts)
	mux.HandleFunc("/voyager/api/socialActions/", s.handleSocialActions)

	// Search
	mux.HandleFunc("/voyager/api/search/hits", s.handleSearchHits)
	mux.HandleFunc("/voyager/api/search/blended", s.handleSearchBlended)

	// GraphQL (used by newer endpoints)
	mux.HandleFunc("/voyager/api/graphql", s.handleGraphQL)
}

// ---- JSON helpers ----

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func writeFixture(w http.ResponseWriter, name string) {
	data := loadFixture(name)
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(data)
}

func writeEmpty(w http.ResponseWriter) {
	writeJSON(w, map[string]interface{}{
		"elements": []interface{}{},
		"paging":   map[string]interface{}{"start": 0, "count": 0, "total": 0},
	})
}

func writeOK(w http.ResponseWriter) {
	w.WriteHeader(http.StatusCreated)
}

func readBody(r *http.Request) map[string]interface{} {
	body, _ := io.ReadAll(r.Body)
	var m map[string]interface{}
	_ = json.Unmarshal(body, &m)
	return m
}

// loadFixture reads a fixture file relative to the mock/fixtures directory.
func loadFixture(name string) []byte {
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)
	path := filepath.Join(dir, "fixtures", name)
	data, err := os.ReadFile(path)
	if err != nil {
		panic(fmt.Sprintf("mock: missing fixture %q: %v", name, err))
	}
	return data
}

// loadFixtureAsInterface parses a fixture into a generic map.
func loadFixtureAsInterface(name string) interface{} {
	data := loadFixture(name)
	var v interface{}
	_ = json.Unmarshal(data, &v)
	return v
}

// ---- Handlers ----

func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	writeFixture(w, "profile.json")
}

func (s *Server) handleProfiles(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		if strings.HasSuffix(r.URL.Path, "/profileContactInfo") {
			writeJSON(w, map[string]interface{}{
				"emailAddress":   "jane@example.com",
				"phoneNumbers":   []interface{}{},
				"twitterHandles": []interface{}{},
				"websites":       []interface{}{},
			})
			return
		}
		writeFixture(w, "profile.json")
	case http.MethodPut:
		writeOK(w)
	}
}

func (s *Server) handleConversations(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// Filter to unread only when q=unread is set.
		if r.URL.Query().Get("q") == "unread" {
			s.serveUnreadConversations(w)
			return
		}
		writeFixture(w, "conversations.json")
	case http.MethodPost:
		body := readBody(r)
		s.mu.Lock()
		s.sentMessages = append(s.sentMessages, sentMessage{Path: r.URL.Path, Payload: body})
		s.mu.Unlock()
		writeOK(w)
	}
}

// handleConversationSub handles paths like /voyager/api/messaging/conversations/<id>/events
// and /voyager/api/messaging/conversations/<id> (PUT for mark-read).
func (s *Server) handleConversationSub(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, map[string]interface{}{
			"elements": []interface{}{
				map[string]interface{}{
					"entityUrn": "urn:li:msg_event:evt001",
					"createdAt": 1741305600000,
					"eventContent": map[string]interface{}{
						"com.linkedin.voyager.messaging.event.MessageEvent": map[string]interface{}{
							"attributedBody": map[string]interface{}{
								"text": "Hey, would you be interested in discussing an opportunity?",
							},
						},
					},
					"from": map[string]interface{}{
						"com.linkedin.voyager.messaging.MessagingMember": map[string]interface{}{
							"entityUrn":        "urn:li:member:987654321",
							"publicIdentifier": "john-smith",
							"firstName":        "John",
							"lastName":         "Smith",
						},
					},
				},
			},
			"paging": map[string]interface{}{"start": 0, "count": 20, "total": 1},
		})
	case http.MethodPost:
		body := readBody(r)
		s.mu.Lock()
		s.sentMessages = append(s.sentMessages, sentMessage{Path: r.URL.Path, Payload: body})
		s.mu.Unlock()
		writeOK(w)
	case http.MethodPut:
		writeOK(w) // mark-read
	}
}

// serveUnreadConversations returns only the unread conversations from the fixture.
func (s *Server) serveUnreadConversations(w http.ResponseWriter) {
	data := loadFixture("conversations.json")
	var raw map[string]interface{}
	_ = json.Unmarshal(data, &raw)

	allElements, _ := raw["elements"].([]interface{})
	var unread []interface{}
	for _, el := range allElements {
		m, ok := el.(map[string]interface{})
		if !ok {
			continue
		}
		if read, _ := m["read"].(bool); !read {
			unread = append(unread, el)
		}
	}

	writeJSON(w, map[string]interface{}{
		"elements": unread,
		"paging":   map[string]interface{}{"start": 0, "count": 20, "total": len(unread)},
	})
}

func (s *Server) handleConnections(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeFixture(w, "connections.json")
	case http.MethodPost:
		writeOK(w)
	}
}

func (s *Server) handleDashConnections(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]interface{}{
		"data": map[string]interface{}{
			"*elements": []interface{}{
				"urn:li:fsd_connection:test-member-id-001",
				"urn:li:fsd_connection:test-member-id-002",
			},
			"paging": map[string]interface{}{"start": 0, "count": 20},
		},
		"included": []interface{}{
			map[string]interface{}{
				"entityUrn":       "urn:li:fsd_connection:test-member-id-001",
				"$type":           "com.linkedin.voyager.dash.relationships.Connection",
				"connectedMember": "urn:li:fsd_profile:test-member-id-001",
				"createdAt":       1741305600000,
			},
			map[string]interface{}{
				"entityUrn":       "urn:li:fsd_connection:test-member-id-002",
				"$type":           "com.linkedin.voyager.dash.relationships.Connection",
				"connectedMember": "urn:li:fsd_profile:test-member-id-002",
				"createdAt":       1741305600000,
			},
		},
	})
}

func (s *Server) handleDashProfiles(w http.ResponseWriter, r *http.Request) {
	// Full profile with decoration: /identity/dash/profiles/urn:li:fsd_profile:...?decorationId=...
	if r.URL.Query().Get("decorationId") != "" {
		s.handleFullProfile(w, r)
		return
	}

	memberID := r.URL.Query().Get("memberIdentity")
	// Return a mock profile for any memberIdentity (used by connections resolution)
	firstName := "Jane"
	lastName := "Doe"
	headline := "Engineering Manager"
	publicID := "jane-doe"
	if memberID == "test-member-id-001" {
		firstName = "Test"
		lastName = "Connection"
		headline = "Software Engineer"
		publicID = "test-connection"
	}
	writeJSON(w, map[string]interface{}{
		"data": map[string]interface{}{
			"*elements": []interface{}{
				"urn:li:fsd_profile:" + memberID,
			},
			"paging": map[string]interface{}{"start": 0, "count": 10},
		},
		"included": []interface{}{
			map[string]interface{}{
				"entityUrn":        "urn:li:fsd_profile:" + memberID,
				"firstName":        firstName,
				"lastName":         lastName,
				"headline":         headline,
				"publicIdentifier": publicID,
				"$type":            "com.linkedin.voyager.dash.identity.profile.Profile",
			},
		},
	})
}

func (s *Server) handleFullProfile(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]interface{}{
		"data": map[string]interface{}{
			"data": map[string]interface{}{
				"entityUrn":        "urn:li:fsd_profile:test-user-encoded-id",
				"firstName":        "Jane",
				"lastName":         "Doe",
				"headline":         "Senior Software Engineer at Acme Corp",
				"summary":          "Passionate software engineer with 10+ years of experience.",
				"publicIdentifier": "jane-doe",
				"objectUrn":        "urn:li:member:123456789",
			},
		},
		"included": []interface{}{
			map[string]interface{}{
				"$type":       "com.linkedin.voyager.dash.identity.profile.Position",
				"title":       "Senior Software Engineer",
				"companyName": "Acme Corp",
				"description": "Lead engineer on the platform team.",
				"locationName": "San Francisco, CA",
				"dateRange": map[string]interface{}{
					"start": map[string]interface{}{"month": 1, "year": 2020},
				},
			},
			map[string]interface{}{
				"$type":       "com.linkedin.voyager.dash.identity.profile.Position",
				"title":       "Software Engineer",
				"companyName": "StartupCo",
				"description": "Full-stack development.",
				"dateRange": map[string]interface{}{
					"start": map[string]interface{}{"month": 6, "year": 2015},
					"end":   map[string]interface{}{"month": 12, "year": 2019},
				},
			},
			map[string]interface{}{
				"$type":      "com.linkedin.voyager.dash.identity.profile.Education",
				"schoolName": "MIT",
				"degreeName": "BS",
				"fieldOfStudy": "Computer Science",
			},
			map[string]interface{}{
				"$type": "com.linkedin.voyager.dash.identity.profile.Skill",
				"name":  "Go",
			},
			map[string]interface{}{
				"$type": "com.linkedin.voyager.dash.identity.profile.Skill",
				"name":  "Python",
			},
			map[string]interface{}{
				"$type": "com.linkedin.voyager.dash.identity.profile.Skill",
				"name":  "Distributed Systems",
			},
			map[string]interface{}{
				"$type": "com.linkedin.voyager.dash.identity.profile.Language",
				"name":  "English",
				"proficiency": "NATIVE_OR_BILINGUAL",
			},
			map[string]interface{}{
				"$type":     "com.linkedin.voyager.dash.identity.profile.Certification",
				"name":      "AWS Solutions Architect",
				"authority":  "Amazon Web Services",
			},
			map[string]interface{}{
				"$type": "com.linkedin.voyager.dash.identity.profile.Patent",
				"title": "Method for distributed cache invalidation",
			},
			map[string]interface{}{
				"$type":                  "com.linkedin.voyager.dash.common.Geo",
				"defaultLocalizedName":   "San Francisco, California, United States",
			},
			map[string]interface{}{
				"$type": "com.linkedin.voyager.dash.common.Industry",
				"name":  "Software Development",
			},
		},
	})
}

func (s *Server) handleInvitations(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]interface{}{
		"elements": []interface{}{
			map[string]interface{}{
				"entityUrn": "urn:li:invitation:inv001",
				"id":        "inv001",
				"fromMember": map[string]interface{}{
					"entityUrn":        "urn:li:member:987654321",
					"publicIdentifier": "john-smith",
					"firstName":        "John",
					"lastName":         "Smith",
					"occupation":       "Engineering Manager",
				},
				"message":  "I'd love to connect!",
				"sentTime": 1741305600000,
				"status":   "PENDING",
			},
		},
		"paging": map[string]interface{}{"start": 0, "count": 20, "total": 1},
	})
}

func (s *Server) handleSentInvitations(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]interface{}{
		"elements": []interface{}{},
		"paging":   map[string]interface{}{"start": 0, "count": 20, "total": 0},
	})
}

func (s *Server) handleInvitationAction(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusCreated)
}

func (s *Server) handleFollows(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		body := readBody(r)
		urn, _ := body["followedEntityUrn"].(string)
		s.mu.Lock()
		s.following[urn] = true
		s.mu.Unlock()
		writeOK(w)
	}
}

func (s *Server) handleFollowByID(w http.ResponseWriter, r *http.Request) {
	// DELETE /voyager/api/feed/follows/<id>
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleJobPosting(w http.ResponseWriter, r *http.Request) {
	writeFixture(w, "jobs.json")
}

func (s *Server) handleJobSaves(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, map[string]interface{}{
			"elements": []interface{}{
				map[string]interface{}{
					"job": loadFixtureAsInterface("jobs.json"),
				},
			},
			"paging": map[string]interface{}{"start": 0, "count": 20, "total": 1},
		})
	case http.MethodPost:
		body := readBody(r)
		jobID, _ := body["jobId"].(string)
		s.mu.Lock()
		s.savedJobs[jobID] = true
		s.mu.Unlock()
		writeOK(w)
	}
}

func (s *Server) handleJobSaveByID(w http.ResponseWriter, r *http.Request) {
	// DELETE /voyager/api/jobs/jobSaves/<id>
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleJobSearch(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]interface{}{
		"elements": []interface{}{
			map[string]interface{}{
				"jobCardUnion": map[string]interface{}{
					"com.linkedin.voyager.jobs.JobPostingCard": map[string]interface{}{
						"entityUrn":         "urn:li:jobPosting:987654321",
						"title":             "Senior Go Engineer",
						"company":           map[string]interface{}{"name": "TechCorp", "universalName": "techcorp"},
						"formattedLocation": "San Francisco, CA",
						"workRemoteAllowed": true,
						"listedAt":          1741219200000,
					},
				},
			},
		},
		"paging": map[string]interface{}{"start": 0, "count": 20, "total": 1},
	})
}

func (s *Server) handleAppliedJobs(w http.ResponseWriter, r *http.Request) {
	writeEmpty(w)
}

func (s *Server) handleCompanies(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]interface{}{
		"data": map[string]interface{}{
			"*elements": []interface{}{"urn:li:fs_normalized_company:9999"},
			"paging":    map[string]interface{}{"start": 0, "count": 10},
		},
		"included": []interface{}{
			map[string]interface{}{
				"entityUrn":     "urn:li:fs_normalized_company:9999",
				"universalName": "anthropic",
				"name":          "Anthropic",
				"tagline":       "AI safety company",
				"description":   "Anthropic is an AI safety company.",
				"industries":    []interface{}{"Software Development"},
				"companyPageUrl": "https://anthropic.com",
				"staffCount":    500,
				"headquarter":   map[string]interface{}{"city": "San Francisco", "country": "US"},
				"$type":         "com.linkedin.voyager.organization.Company",
			},
		},
	})
}

func (s *Server) handleFeed(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]interface{}{
		"data": map[string]interface{}{
			"*elements": []interface{}{
				"urn:li:fs_updateV2:(urn:li:activity:aaa001,MAIN_FEED,EMPTY,DEFAULT,false)",
			},
			"paging": map[string]interface{}{"start": 0, "count": 20},
		},
		"included": []interface{}{
			map[string]interface{}{
				"entityUrn": "urn:li:fs_updateV2:(urn:li:activity:aaa001,MAIN_FEED,EMPTY,DEFAULT,false)",
				"$type":     "com.linkedin.voyager.feed.render.UpdateV2",
				"updateMetadata": map[string]interface{}{
					"urn": "urn:li:activity:aaa001",
				},
				"header": map[string]interface{}{
					"text": map[string]interface{}{
						"text": "Jane Doe",
						"attributes": []interface{}{
							map[string]interface{}{
								"*miniProfile": "urn:li:fs_miniProfile:jane-doe-id",
							},
						},
					},
				},
				"commentary": map[string]interface{}{
					"text": map[string]interface{}{"text": "Excited to announce our new product launch!"},
				},
				"*socialDetail": "urn:li:fs_socialDetail:urn:li:activity:aaa001",
			},
			map[string]interface{}{
				"entityUrn": "urn:li:fs_miniProfile:jane-doe-id",
				"$type":     "com.linkedin.voyager.identity.shared.MiniProfile",
				"firstName": "Jane",
				"lastName":  "Doe",
				"occupation": "Product Manager",
				"publicIdentifier": "jane-doe",
			},
			map[string]interface{}{
				"entityUrn":   "urn:li:fs_socialActivityCounts:urn:li:activity:aaa001",
				"$type":       "com.linkedin.voyager.feed.shared.SocialActivityCounts",
				"numLikes":    42,
				"numComments": 7,
				"numShares":   3,
			},
		},
	})
}

func (s *Server) handleUGCPosts(w http.ResponseWriter, r *http.Request) {
	body := readBody(r)
	s.mu.Lock()
	s.createdPosts = append(s.createdPosts, body)
	s.mu.Unlock()
	writeOK(w)
}

func (s *Server) handleSocialActions(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	switch {
	case strings.HasSuffix(path, "/likes"):
		w.WriteHeader(http.StatusCreated)
	case strings.HasSuffix(path, "/comments"):
		switch r.Method {
		case http.MethodGet:
			writeJSON(w, map[string]interface{}{
				"elements": []interface{}{
					map[string]interface{}{
						"entityUrn": "urn:li:comment:c001",
						"actor":     "urn:li:member:111000001",
						"message":   map[string]interface{}{"text": "Great post!"},
						"likeCount": 5,
						"createdAt": 1741219200000,
					},
				},
			})
		case http.MethodPost:
			writeOK(w)
		}
	default:
		writeOK(w)
	}
}

func (s *Server) handleSearchHits(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	switch q {
	case "people":
		writeJSON(w, map[string]interface{}{
			"elements": []interface{}{
				map[string]interface{}{
					"hitInfo": map[string]interface{}{
						"com.linkedin.voyager.search.SearchProfile": map[string]interface{}{
							"entityUrn":        "urn:li:member:111000010",
							"publicIdentifier": "alex-engineer",
							"firstName":        "Alex",
							"lastName":         "Engineer",
							"occupation":       "Software Engineer at FutureCo",
						},
					},
					"distance": map[string]interface{}{"value": "DISTANCE_2"},
				},
			},
			"paging": map[string]interface{}{"start": 0, "count": 20, "total": 1},
		})
	case "company":
		writeJSON(w, map[string]interface{}{
			"elements": []interface{}{
				map[string]interface{}{
					"hitInfo": map[string]interface{}{
						"com.linkedin.voyager.search.SearchCompany": map[string]interface{}{
							"entityUrn":     "urn:li:company:9999",
							"universalName": "anthropic",
							"name":          "Anthropic",
							"industry":      map[string]interface{}{"localizedName": "AI Safety"},
						},
					},
				},
			},
			"paging": map[string]interface{}{"start": 0, "count": 20, "total": 1},
		})
	default:
		writeEmpty(w)
	}
}

func (s *Server) handleSearchBlended(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]interface{}{
		"elements": []interface{}{},
		"paging":   map[string]interface{}{"start": 0, "count": 20, "total": 0},
	})
}

func (s *Server) handleRecommendations(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeFixture(w, "recommendations.json")
	case http.MethodPost:
		writeOK(w)
	}
}

func (s *Server) handleRecommendationByID(w http.ResponseWriter, r *http.Request) {
	// PUT /voyager/api/identity/recommendations/<id> — hide or show
	w.WriteHeader(http.StatusCreated)
}

func (s *Server) handleNotifications(w http.ResponseWriter, r *http.Request) {
	writeFixture(w, "notifications.json")
}

func (s *Server) handleNotificationByID(w http.ResponseWriter, r *http.Request) {
	// PUT /voyager/api/feed/notifications/<id> — mark read
	w.WriteHeader(http.StatusCreated)
}

func (s *Server) handleDashNotificationCards(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]interface{}{
		"data": map[string]interface{}{
			"*elements": []interface{}{
				"urn:li:fsd_notificationCard:notif001",
				"urn:li:fsd_notificationCard:notif002",
			},
			"paging": map[string]interface{}{"start": 0, "count": 10},
		},
		"included": []interface{}{
			map[string]interface{}{
				"entityUrn":   "urn:li:fsd_notificationCard:notif001",
				"headline":    map[string]interface{}{"text": "John Smith liked your post"},
				"read":        false,
				"publishedAt": 1741305600000,
			},
			map[string]interface{}{
				"entityUrn":   "urn:li:fsd_notificationCard:notif002",
				"headline":    map[string]interface{}{"text": "You appeared in 5 searches this week"},
				"read":        true,
				"publishedAt": 1741219200000,
			},
		},
	})
}

func (s *Server) handleDashBadgingCounts(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]interface{}{
		"data": map[string]interface{}{
			"elements": []interface{}{
				map[string]interface{}{"badgingItem": "NOTIFICATIONS", "count": 3},
				map[string]interface{}{"badgingItem": "MESSAGING", "count": 1},
				map[string]interface{}{"badgingItem": "MY_NETWORK", "count": 0},
			},
		},
	})
}

func (s *Server) handleDashMessengerMessages(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Get("action") == "createMessage" && r.Method == "POST" {
		writeJSON(w, map[string]interface{}{
			"value": map[string]interface{}{
				"entityUrn":              "urn:li:msg_message:test-msg-001",
				"backendConversationUrn": "urn:li:messagingThread:test-thread-001",
			},
		})
		return
	}
	http.Error(w, `{"status":404}`, http.StatusNotFound)
}

// handleMessagingGraphQL handles requests to the dedicated messaging GraphQL gateway.
func (s *Server) handleMessagingGraphQL(w http.ResponseWriter, r *http.Request) {
	queryID := r.URL.Query().Get("queryId")
	switch {
	case strings.HasPrefix(queryID, "messengerConversations."):
		writeJSON(w, map[string]interface{}{
			"data": map[string]interface{}{
				"messengerConversationsByCategoryQuery": map[string]interface{}{
					"elements": []interface{}{
						map[string]interface{}{
							"entityUrn":      "urn:li:msg_conversation:(urn:li:fsd_profile:test-user-encoded-id,thread001)",
							"backendUrn":     "urn:li:messagingThread:thread001",
							"unreadCount":    0,
							"read":           true,
							"lastActivityAt": 1741305600000,
							"conversationParticipants": []interface{}{
								map[string]interface{}{
									"hostIdentityUrn": "urn:li:fsd_profile:other-user-id",
									"participantType": map[string]interface{}{
										"member": map[string]interface{}{
											"firstName": map[string]interface{}{"text": "Alice"},
											"lastName":  map[string]interface{}{"text": "Smith"},
										},
									},
								},
								map[string]interface{}{
									"hostIdentityUrn": "urn:li:fsd_profile:test-user-encoded-id",
									"participantType": map[string]interface{}{
										"member": map[string]interface{}{
											"firstName": map[string]interface{}{"text": "Test"},
											"lastName":  map[string]interface{}{"text": "User"},
										},
									},
								},
							},
							"messages": map[string]interface{}{
								"elements": []interface{}{
									map[string]interface{}{
										"entityUrn":  "urn:li:msg_message:msg001",
										"deliveredAt": 1741305600000,
										"body":        map[string]interface{}{"text": "Hi there, how are you?"},
										"sender": map[string]interface{}{
											"hostIdentityUrn": "urn:li:fsd_profile:other-user-id",
											"participantType": map[string]interface{}{
												"member": map[string]interface{}{
													"firstName": map[string]interface{}{"text": "Alice"},
													"lastName":  map[string]interface{}{"text": "Smith"},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		})
	case strings.HasPrefix(queryID, "messengerMessages."):
		writeJSON(w, map[string]interface{}{
			"data": map[string]interface{}{
				"messengerMessagesBySyncToken": map[string]interface{}{
					"elements": []interface{}{
						map[string]interface{}{
							"entityUrn":  "urn:li:msg_message:msg001",
							"deliveredAt": 1741305600000,
							"body":        map[string]interface{}{"text": "Hi there, how are you?"},
							"sender": map[string]interface{}{
								"hostIdentityUrn": "urn:li:fsd_profile:other-user-id",
								"participantType": map[string]interface{}{
									"member": map[string]interface{}{
										"firstName": map[string]interface{}{"text": "Alice"},
										"lastName":  map[string]interface{}{"text": "Smith"},
									},
								},
							},
						},
					},
				},
			},
		})
	default:
		http.Error(w, `{"status":404,"message":"Unknown messaging queryId"}`, http.StatusNotFound)
	}
}

// handleGraphQL routes GraphQL requests by queryId parameter.
func (s *Server) handleGraphQL(w http.ResponseWriter, r *http.Request) {
	queryID := r.URL.Query().Get("queryId")
	switch {
	case strings.HasPrefix(queryID, "voyagerFeedDashIdentityModule."):
		writeFixture(w, "who-viewed-count.json")
	case strings.HasPrefix(queryID, "voyagerPremiumDashAnalyticsObject."):
		writeFixture(w, "who-viewed.json")
	case strings.HasPrefix(queryID, "voyagerRelationshipsDashInvitationViews."):
		writeFixture(w, "invitation-views.json")
	case strings.HasPrefix(queryID, "voyagerRelationshipsDashSentInvitationViews."):
		writeFixture(w, "sent-invitation-views.json")
	case strings.HasPrefix(queryID, "voyagerRelationshipsDashInvitationsSummary."):
		writeJSON(w, map[string]interface{}{
			"data": map[string]interface{}{
				"relationshipsDashInvitationsSummaryByInvitationSummaryTypes": map[string]interface{}{
					"elements": []interface{}{
						map[string]interface{}{"numPendingInvitations": 1, "numNewInvitations": 0},
					},
				},
			},
		})
	case strings.HasPrefix(queryID, "voyagerMessagingDashMessengerConversations."):
		writeJSON(w, map[string]interface{}{
			"data": map[string]interface{}{
				"messengerConversationsByCategory": map[string]interface{}{
					"elements": []interface{}{
						map[string]interface{}{
							"entityUrn":      "urn:li:msg_conversation:(urn:li:fsd_profile:test-user-encoded-id,thread001)",
							"backendUrn":     "urn:li:messagingThread:thread001",
							"unreadCount":    0,
							"read":           true,
							"lastActivityAt":  1741305600000,
							"conversationParticipants": []interface{}{
								map[string]interface{}{
									"hostIdentityUrn": "urn:li:fsd_profile:other-user-id",
									"participantType": map[string]interface{}{
										"member": map[string]interface{}{
											"firstName": map[string]interface{}{"text": "Alice"},
											"lastName":  map[string]interface{}{"text": "Smith"},
										},
									},
								},
								map[string]interface{}{
									"hostIdentityUrn": "urn:li:fsd_profile:test-user-encoded-id",
									"participantType": map[string]interface{}{
										"member": map[string]interface{}{
											"firstName": map[string]interface{}{"text": "Test"},
											"lastName":  map[string]interface{}{"text": "User"},
										},
									},
								},
							},
							"messages": map[string]interface{}{
								"elements": []interface{}{
									map[string]interface{}{
										"entityUrn":   "urn:li:msg_message:msg001",
										"deliveredAt":  1741305600000,
										"body":         map[string]interface{}{"text": "Hi there, how are you?"},
										"sender": map[string]interface{}{
											"hostIdentityUrn": "urn:li:fsd_profile:other-user-id",
											"participantType": map[string]interface{}{
												"member": map[string]interface{}{
													"firstName": map[string]interface{}{"text": "Alice"},
													"lastName":  map[string]interface{}{"text": "Smith"},
												},
											},
										},
									},
								},
							},
						},
					},
					"paging": map[string]interface{}{"start": 0, "count": 20, "total": 1},
				},
			},
		})
	case strings.HasPrefix(queryID, "voyagerMessagingDashMessengerMessages."):
		writeJSON(w, map[string]interface{}{
			"data": map[string]interface{}{
				"messengerMessagesByConversation": map[string]interface{}{
					"elements": []interface{}{
						map[string]interface{}{
							"entityUrn":   "urn:li:msg_message:msg001",
							"deliveredAt":  1741305600000,
							"body":         map[string]interface{}{"text": "Hi there, how are you?"},
							"sender": map[string]interface{}{
								"hostIdentityUrn": "urn:li:fsd_profile:other-user-id",
								"participantType": map[string]interface{}{
									"member": map[string]interface{}{
										"firstName": map[string]interface{}{"text": "Alice"},
										"lastName":  map[string]interface{}{"text": "Smith"},
									},
								},
							},
						},
					},
					"paging": map[string]interface{}{"start": 0, "count": 20, "total": 1},
				},
			},
		})
	case strings.HasPrefix(queryID, "voyagerJobsDashJobsFeed."):
		s.handleGraphQLJobsFeed(w, r)
	case strings.HasPrefix(queryID, "voyagerJobsDashJobCards."):
		s.handleGraphQLJobCards(w, r)
	case strings.HasPrefix(queryID, "voyagerJobsDashJobPostings."):
		s.handleGraphQLJobPosting(w, r)
	case strings.HasPrefix(queryID, "voyagerSearchDashClusters."):
		s.handleGraphQLSearch(w, r)
	default:
		http.Error(w, `{"status":404,"message":"Unknown queryId"}`, http.StatusNotFound)
	}
}

// handleGraphQLJobsFeed returns recommended jobs.
func (s *Server) handleGraphQLJobsFeed(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, map[string]interface{}{
		"data": map[string]interface{}{
			"jobsDashJobsFeedAll": map[string]interface{}{
				"elements": []interface{}{
					map[string]interface{}{
						"entitiesResolutionResults": []interface{}{
							map[string]interface{}{
								"jobPostingCard": map[string]interface{}{
									"jobPostingTitle": "Senior Go Engineer",
									"primaryDescription": map[string]interface{}{
										"text": "TechCorp · San Francisco, CA",
									},
									"jobPosting": map[string]interface{}{
										"entityUrn": "urn:li:fsd_jobPosting:987654321",
										"title":     "Senior Go Engineer",
									},
								},
							},
							map[string]interface{}{
								"jobPostingCard": map[string]interface{}{
									"jobPostingTitle": "Backend Developer",
									"primaryDescription": map[string]interface{}{
										"text": "StartupCo · New York, NY",
									},
									"jobPosting": map[string]interface{}{
										"entityUrn": "urn:li:fsd_jobPosting:111222333",
										"title":     "Backend Developer",
									},
								},
							},
						},
					},
				},
				"paging": map[string]interface{}{"start": 0, "count": 20, "total": 2},
			},
		},
	})
}

// handleGraphQLJobCards returns saved or applied job cards based on jobCollectionSlug.
func (s *Server) handleGraphQLJobCards(w http.ResponseWriter, r *http.Request) {
	vars := r.URL.Query().Get("variables")
	isApplied := strings.Contains(vars, "jobCollectionSlug:applied")

	if isApplied {
		writeJSON(w, map[string]interface{}{
			"data": map[string]interface{}{
				"jobsDashJobCardsByJobSearchV2": map[string]interface{}{
					"elements": []interface{}{},
					"paging":   map[string]interface{}{"start": 0, "count": 20, "total": 0},
				},
			},
		})
		return
	}

	// saved jobs
	writeJSON(w, map[string]interface{}{
		"data": map[string]interface{}{
			"jobsDashJobCardsByJobSearchV2": map[string]interface{}{
				"elements": []interface{}{
					map[string]interface{}{
						"jobCard": map[string]interface{}{
							"jobPostingCardWrapper": map[string]interface{}{
								"jobPostingCard": map[string]interface{}{
									"jobPostingTitle": "Senior Go Engineer",
									"primaryDescription": map[string]interface{}{
										"text": "TechCorp · San Francisco, CA",
									},
									"jobPosting": map[string]interface{}{
										"entityUrn": "urn:li:fsd_jobPosting:987654321",
										"title":     "Senior Go Engineer",
									},
								},
							},
						},
					},
				},
				"paging": map[string]interface{}{"start": 0, "count": 20, "total": 1},
			},
		},
	})
}

// handleGraphQLJobPosting returns a single job posting detail.
func (s *Server) handleGraphQLJobPosting(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, map[string]interface{}{
		"data": map[string]interface{}{
			"jobsDashJobPostingsById": map[string]interface{}{
				"entityUrn": "urn:li:fsd_jobPosting:987654321",
				"title":     "Senior Go Engineer",
				"description": map[string]interface{}{
					"text": "We are looking for a Senior Go Engineer to join our platform team.",
				},
				"companyDetails": map[string]interface{}{
					"name": "TechCorp",
				},
				"location": map[string]interface{}{
					"defaultLocalizedName": "San Francisco, CA",
				},
				"employmentStatus": map[string]interface{}{
					"localizedName": "Full-time",
				},
				"jobState":          "LISTED",
				"workRemoteAllowed": true,
				"companyApplyUrl":   "https://techcorp.com/apply/987654321",
			},
		},
	})
}

// handleGraphQLSearch returns mock search results.
func (s *Server) handleGraphQLSearch(w http.ResponseWriter, r *http.Request) {
	vars := r.URL.Query().Get("variables")

	if strings.Contains(vars, "CONTENT") {
		// Posts search
		writeJSON(w, map[string]interface{}{
			"data": map[string]interface{}{
				"searchDashClustersByAll": map[string]interface{}{
					"elements": []interface{}{
						map[string]interface{}{
							"items": []interface{}{
								map[string]interface{}{
									"item": map[string]interface{}{
										"searchFeedUpdate": map[string]interface{}{
											"update": map[string]interface{}{
												"actor":      map[string]interface{}{"name": map[string]interface{}{"text": "Jane Doe"}},
												"commentary": map[string]interface{}{"text": map[string]interface{}{"text": "Great article about cloud architecture!"}},
												"metadata":   map[string]interface{}{"shareUrn": "urn:li:ugcPost:12345"},
											},
										},
									},
								},
							},
						},
					},
					"paging": map[string]interface{}{"start": 0, "count": 20, "total": 1},
				},
			},
		})
		return
	}

	// People or Companies search (entityResult format)
	writeJSON(w, map[string]interface{}{
		"data": map[string]interface{}{
			"searchDashClustersByAll": map[string]interface{}{
				"elements": []interface{}{
					map[string]interface{}{
						"items": []interface{}{
							map[string]interface{}{
								"item": map[string]interface{}{
									"entityResult": map[string]interface{}{
										"entityUrn":       "urn:li:fsd_entityResultViewModel:(urn:li:fsd_profile:test-id,SEARCH_SRP,DEFAULT)",
										"title":           map[string]interface{}{"text": "Alex Engineer"},
										"primarySubtitle": map[string]interface{}{"text": "Software Engineer at FutureCo"},
									},
								},
							},
						},
					},
				},
				"paging": map[string]interface{}{"start": 0, "count": 20, "total": 1},
			},
		},
	})
}
