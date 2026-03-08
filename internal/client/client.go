package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/russ-blaisdell/linked/internal/models"
)

const (
	userAgent      = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/145.0.0.0 Safari/537.36"
	clientVersion  = "1.13.42665"
	defaultTimeout = 30 * time.Second
)

// Client is an authenticated LinkedIn Voyager API client.
type Client struct {
	httpClient *http.Client
	baseURL    string
	creds      *models.Credentials
}

// cookieTransport ensures auth cookies are sent with their exact stored values
// (including surrounding quotes that Go's cookie jar would strip). It runs after
// the jar has populated the Cookie header, strips any jar-set versions of the
// auth cookies, then appends the correctly-formatted values from credentials.
type cookieTransport struct {
	base       http.RoundTripper
	liAt       string
	jsessionid string
	bcookie    string // browser fingerprint — must match the session li_at was issued with
	bscookie   string // secure browser fingerprint — also session-bound
}

func (t *cookieTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())

	// Rebuild Cookie header: keep jar cookies except the ones we control.
	var kept []string
	for _, part := range strings.Split(req.Header.Get("Cookie"), "; ") {
		if part == "" ||
			strings.HasPrefix(part, "li_at=") ||
			strings.HasPrefix(part, "JSESSIONID=") ||
			(t.bcookie != "" && strings.HasPrefix(part, "bcookie=")) ||
			(t.bscookie != "" && strings.HasPrefix(part, "bscookie=")) {
			continue
		}
		kept = append(kept, part)
	}
	kept = append(kept, "li_at="+t.liAt, "JSESSIONID="+t.jsessionid)
	if t.bcookie != "" {
		kept = append(kept, "bcookie="+t.bcookie)
	}
	if t.bscookie != "" {
		kept = append(kept, "bscookie="+t.bscookie)
	}
	req.Header.Set("Cookie", strings.Join(kept, "; "))

	resp, err := t.base.RoundTrip(req)
	if err == nil && os.Getenv("LINKED_DEBUG") != "" && (resp.StatusCode >= 300 && resp.StatusCode < 400) {
		fmt.Fprintf(os.Stderr, "[debug] transport: %d Location=%s Set-Cookie=%v\n",
			resp.StatusCode, resp.Header.Get("Location"), resp.Header["Set-Cookie"])
	}
	return resp, err
}

// New creates a Client using the given credentials against the real LinkedIn API.
func New(creds *models.Credentials) (*Client, error) {
	return NewWithBaseURL(creds, BaseURL)
}

// NewWithBaseURL creates a Client with a custom base URL (used for mock server in tests).
func NewWithBaseURL(creds *models.Credentials, baseURL string) (*Client, error) {
	if _, err := url.Parse(baseURL); err != nil {
		return nil, fmt.Errorf("parsing base URL: %w", err)
	}
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("creating cookie jar: %w", err)
	}
	// Seed the jar with the lang cookie (no quote issues here).
	parsed, _ := url.Parse(baseURL)
	jar.SetCookies(parsed, []*http.Cookie{
		{Name: "lang", Value: "v=2&lang=en-us"},
	})

	transport := &cookieTransport{
		base:       http.DefaultTransport,
		liAt:       creds.LiAt,
		jsessionid: creds.JSESSIONID,
		bcookie:    creds.Bcookie,
		bscookie:   creds.Bscookie,
	}
	return &Client{
		httpClient: &http.Client{
			Timeout:   defaultTimeout,
			Jar:       jar,
			Transport: transport,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if os.Getenv("LINKED_DEBUG") != "" {
					fmt.Fprintf(os.Stderr, "[debug] redirect → %s\n", req.URL)
				}
				if len(via) >= 10 {
					return fmt.Errorf("stopped after 10 redirects")
				}
				return nil
			},
		},
		baseURL: baseURL,
		creds:   creds,
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
		`{"clientVersion":"%s","mpVersion":"%s","osName":"web","timezoneOffset":-5,"timezone":"America/New_York","deviceFormFactor":"DESKTOP","mpName":"voyager-web","displayDensity":2,"displayWidth":1920,"displayHeight":1080}`,
		clientVersion, clientVersion,
	))
	req.Header.Set("Accept", "application/vnd.linkedin.normalized+json+2.1")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Origin", BaseURL)
	req.Header.Set("Referer", BaseURL+"/feed/")
	req.Header.Set("sec-ch-ua", `"Not:A-Brand";v="99", "Google Chrome";v="145", "Chromium";v="145"`)
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", `"macOS"`)
	req.Header.Set("sec-fetch-dest", "empty")
	req.Header.Set("sec-fetch-mode", "cors")
	req.Header.Set("sec-fetch-site", "same-origin")
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

	if os.Getenv("LINKED_DEBUG") != "" {
		fmt.Fprintf(os.Stderr, "[debug] %s %s → %d Set-Cookie=%v\n%s\n",
			req.Method, req.URL, resp.StatusCode, resp.Header["Set-Cookie"], body)
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

// PutBinary performs a PUT with raw binary data to an absolute URL.
// Used for media uploads where LinkedIn returns an external upload URL.
func (c *Client) PutBinary(uploadURL string, data []byte, contentType string) error {
	req, err := http.NewRequest(http.MethodPut, uploadURL, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("creating upload request: %w", err)
	}
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("User-Agent", userAgent)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("executing upload request: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return apiError(resp.StatusCode, body)
	}
	return nil
}
