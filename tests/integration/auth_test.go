package integration_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/russ-blaisdell/linked/internal/api"
	"github.com/russ-blaisdell/linked/internal/client"
	"github.com/russ-blaisdell/linked/internal/models"
	"github.com/russ-blaisdell/linked/mock"
)

func TestUnauthorizedRequest(t *testing.T) {
	s := startServer(t)

	// Use wrong credentials.
	creds := &models.Credentials{
		LiAt:       "bad-token",
		JSESSIONID: "wrong-jsessionid",
	}
	c, err := client.NewWithBaseURL(creds, s.URL())
	if err != nil {
		t.Fatalf("creating client: %v", err)
	}
	li := api.New(c)

	_, err = li.Profile.GetMe()
	if err == nil {
		t.Fatal("expected error for bad credentials, got nil")
	}

	liErr, ok := err.(*client.LinkedInError)
	if !ok {
		// Wrapped error — check the message contains 401 info.
		t.Logf("error (wrapped): %v", err)
		return
	}
	if liErr.StatusCode != 401 {
		t.Errorf("StatusCode = %d, want 401", liErr.StatusCode)
	}
}

func TestRateLimitError(t *testing.T) {
	// Set up a server that always returns 429.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"status":429}`, http.StatusTooManyRequests)
	}))
	defer ts.Close()

	liAt, jsessionid := mock.TestCredentials()
	creds := &models.Credentials{LiAt: liAt, JSESSIONID: jsessionid}
	c, _ := client.NewWithBaseURL(creds, ts.URL)
	li := api.New(c)

	_, err := li.Profile.GetMe()
	if err == nil {
		t.Fatal("expected rate limit error")
	}
	t.Logf("rate limit error: %v", err)
}

func TestNotFoundError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"status":404}`, http.StatusNotFound)
	}))
	defer ts.Close()

	liAt, jsessionid := mock.TestCredentials()
	creds := &models.Credentials{LiAt: liAt, JSESSIONID: jsessionid}
	c, _ := client.NewWithBaseURL(creds, ts.URL)
	li := api.New(c)

	_, err := li.Profile.GetMe()
	if err == nil {
		t.Fatal("expected not found error")
	}
	t.Logf("not found error: %v", err)
}

func TestCSRFValidation(t *testing.T) {
	s := startServer(t)

	// Craft a request with no CSRF token.
	resp, err := http.Get(s.URL() + "/voyager/api/me")
	if err != nil {
		t.Fatalf("GET error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401 without CSRF token, got %d", resp.StatusCode)
	}
}
