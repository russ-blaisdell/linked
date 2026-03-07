package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"github.com/russ-blaisdell/linked/internal/models"
)

const (
	userAgent      = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36"
	clientVersion  = "1.13.14099"
	defaultTimeout = 30 * time.Second
)

// Client is an authenticated LinkedIn Voyager API client.
type Client struct {
	httpClient *http.Client
	baseURL    string
	creds      *models.Credentials
}

// New creates a Client using the given credentials against the real LinkedIn API.
func New(creds *models.Credentials) (*Client, error) {
	return NewWithBaseURL(creds, BaseURL)
}

// NewWithBaseURL creates a Client with a custom base URL (used for mock server in tests).
func NewWithBaseURL(creds *models.Credentials, baseURL string) (*Client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("creating cookie jar: %w", err)
	}

	parsed, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("parsing base URL: %w", err)
	}

	jar.SetCookies(parsed, []*http.Cookie{
		{Name: "li_at", Value: creds.LiAt},
		{Name: "JSESSIONID", Value: creds.JSESSIONID},
		{Name: "lang", Value: "v=2&lang=en-us"},
	})

	return &Client{
		httpClient: &http.Client{Jar: jar, Timeout: defaultTimeout},
		baseURL:    baseURL,
		creds:      creds,
	}, nil
}

// csrfToken returns the CSRF token derived from JSESSIONID.
func (c *Client) csrfToken() string {
	return strings.Trim(c.creds.JSESSIONID, `"`)
}

// newRequest builds an HTTP request with all LinkedIn-required headers.
func (c *Client) newRequest(method, path string, body interface{}) (*http.Request, error) {
	fullURL := c.baseURL + path

	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshaling request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, fullURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("csrf-token", c.csrfToken())
	req.Header.Set("x-restli-protocol-version", "2.0.0")
	req.Header.Set("x-li-lang", "en_US")
	req.Header.Set("x-li-track", fmt.Sprintf(
		`{"clientVersion":"%s","osName":"web","timezoneOffset":0,"timezone":"UTC","deviceFormFactor":"DESKTOP","mpName":"voyager-web"}`,
		clientVersion,
	))
	req.Header.Set("Accept", "application/vnd.linkedin.normalized+json+2.1")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return req, nil
}

// do executes a request and decodes the JSON response into dest.
// dest may be nil to discard the body.
func (c *Client) do(req *http.Request, dest interface{}) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return apiError(resp.StatusCode, body)
	}

	if dest != nil && len(body) > 0 {
		if err := json.Unmarshal(body, dest); err != nil {
			return fmt.Errorf("decoding response: %w", err)
		}
	}

	return nil
}

// Get performs a GET request with optional query params.
func (c *Client) Get(path string, params map[string]string, dest interface{}) error {
	if len(params) > 0 {
		q := url.Values{}
		for k, v := range params {
			q.Set(k, v)
		}
		path = path + "?" + q.Encode()
	}
	req, err := c.newRequest(http.MethodGet, path, nil)
	if err != nil {
		return err
	}
	return c.do(req, dest)
}

// Post performs a POST request with a JSON body.
func (c *Client) Post(path string, body interface{}, dest interface{}) error {
	req, err := c.newRequest(http.MethodPost, path, body)
	if err != nil {
		return err
	}
	return c.do(req, dest)
}

// Put performs a PUT request with a JSON body.
func (c *Client) Put(path string, body interface{}, dest interface{}) error {
	req, err := c.newRequest(http.MethodPut, path, body)
	if err != nil {
		return err
	}
	return c.do(req, dest)
}

// Delete performs a DELETE request.
func (c *Client) Delete(path string) error {
	req, err := c.newRequest(http.MethodDelete, path, nil)
	if err != nil {
		return err
	}
	return c.do(req, nil)
}
