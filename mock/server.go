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

	// Connections
	mux.HandleFunc("/voyager/api/relationships/connections", s.handleConnections)
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
		"elements": []interface{}{
			map[string]interface{}{
				"entityUrn":     "urn:li:company:9999",
				"universalName": "anthropic",
				"name":          "Anthropic",
				"tagline":       "AI safety company",
				"description":   "Anthropic is an AI safety company.",
				"industry":      map[string]interface{}{"localizedName": "Software Development"},
				"websiteUrl":    "https://anthropic.com",
				"staffCount":    float64(500),
				"headquarter":   map[string]interface{}{"city": "San Francisco", "country": "US"},
			},
		},
	})
}

func (s *Server) handleFeed(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]interface{}{
		"elements": []interface{}{
			map[string]interface{}{
				"com.linkedin.voyager.feed.render.UpdateV2": map[string]interface{}{
					"entityUrn": "urn:li:activity:aaa001",
					"commentary": map[string]interface{}{
						"text": map[string]interface{}{"text": "Excited to announce our new product launch!"},
					},
					"socialDetail": map[string]interface{}{
						"likeCount": 42, "commentCount": 7, "shareCount": 3,
					},
					"createdAt": 1741305600000,
					"actor": map[string]interface{}{
						"urn":  "urn:li:member:123456789",
						"name": map[string]interface{}{"text": "Jane Doe"},
					},
				},
			},
		},
		"paging": map[string]interface{}{"start": 0, "count": 20, "total": 1},
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

// handleGraphQL routes GraphQL requests by queryId parameter.
func (s *Server) handleGraphQL(w http.ResponseWriter, r *http.Request) {
	queryID := r.URL.Query().Get("queryId")
	switch {
	case strings.HasPrefix(queryID, "voyagerFeedDashIdentityModule."):
		writeFixture(w, "who-viewed-count.json")
	case strings.HasPrefix(queryID, "voyagerPremiumDashAnalyticsObject."):
		writeFixture(w, "who-viewed.json")
	default:
		http.Error(w, `{"status":404,"message":"Unknown queryId"}`, http.StatusNotFound)
	}
}
