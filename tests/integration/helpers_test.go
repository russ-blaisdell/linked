package integration_test

import (
	"testing"

	"github.com/russ-blaisdell/linked/internal/api"
	"github.com/russ-blaisdell/linked/internal/client"
	"github.com/russ-blaisdell/linked/internal/models"
	"github.com/russ-blaisdell/linked/mock"
)

// newTestLinkedIn creates a LinkedIn API client wired to the mock server.
func newTestLinkedIn(t *testing.T, s *mock.Server) *api.LinkedIn {
	t.Helper()
	liAt, jsessionid := mock.TestCredentials()
	creds := &models.Credentials{
		LiAt:       liAt,
		JSESSIONID: jsessionid,
	}
	c, err := client.NewWithBaseURL(creds, s.URL())
	if err != nil {
		t.Fatalf("creating test client: %v", err)
	}
	return api.New(c)
}

// startServer starts the mock server and registers cleanup.
func startServer(t *testing.T) *mock.Server {
	t.Helper()
	s := mock.New()
	t.Cleanup(s.Close)
	return s
}
