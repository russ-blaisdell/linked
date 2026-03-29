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
	userAgent      = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/146.0.0.0 Safari/537.36"
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

	// Use Chrome TLS fingerprint for real LinkedIn to avoid bot detection.
	// Mock test servers run on localhost without TLS, so use the default transport.
	var base http.RoundTripper
	if isLocalhostURL(baseURL) {
		base = http.DefaultTransport
	} else {
		base = chromeTransport()
	}

	transport := &cookieTransport{
		base:       base,
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

// fp returns the stored fingerprint value if non-empty, otherwise the fallback.
func (c *Client) fp(fpVal, fallback string) string {
	if fpVal != "" {
		return fpVal
	}
	return fallback
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

	// Use stored browser fingerprint when available, fall back to defaults.
	var fpUA, fpSecUA, fpSecMobile, fpSecPlatform, fpTrack string
	if f := c.creds.Fingerprint; f != nil {
		fpUA = f.UserAgent
		fpSecUA = f.SecChUA
		fpSecMobile = f.SecChUAMobile
		fpSecPlatform = f.SecChUAPlatform
		fpTrack = f.XLiTrack
	}

	defaultTrack := fmt.Sprintf(
		`{"clientVersion":"%s","mpVersion":"%s","osName":"web","timezoneOffset":-5,"timezone":"America/Chicago","deviceFormFactor":"DESKTOP","mpName":"voyager-web","displayDensity":2,"displayWidth":3456,"displayHeight":2234}`,
		clientVersion, clientVersion,
	)

	req.Header.Set("User-Agent", c.fp(fpUA, userAgent))
	req.Header.Set("csrf-token", c.csrfToken())
	req.Header.Set("x-restli-protocol-version", "2.0.0")
	req.Header.Set("x-li-lang", "en_US")
	req.Header.Set("x-li-track", c.fp(fpTrack, defaultTrack))
	req.Header.Set("Accept", "application/vnd.linkedin.normalized+json+2.1")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	// Browsers only send Origin on non-GET requests (POST, PUT, DELETE).
	// Sending Origin on GETs is a bot signal that triggers session revocation.
	if method != http.MethodGet {
		req.Header.Set("Origin", BaseURL)
	}
	req.Header.Set("Referer", BaseURL+"/feed/")
	req.Header.Set("sec-ch-ua", c.fp(fpSecUA, `"Chromium";v="146", "Not-A.Brand";v="24", "Google Chrome";v="146"`))
	req.Header.Set("sec-ch-ua-mobile", c.fp(fpSecMobile, "?0"))
	req.Header.Set("sec-ch-ua-platform", c.fp(fpSecPlatform, `"Linux"`))
	req.Header.Set("sec-ch-prefers-color-scheme", "light")
	req.Header.Set("sec-fetch-dest", "empty")
	req.Header.Set("sec-fetch-mode", "cors")
	req.Header.Set("sec-fetch-site", "same-origin")
	req.Header.Set("priority", "u=1, i")
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

// GetGraphQL performs a GET to the GraphQL endpoint with a pre-formed query string.
// It overrides the Accept header to application/json, which the graphql endpoint requires.
func (c *Client) GetGraphQL(rawPath string, dest interface{}) error {
	req, err := c.newRequest(http.MethodGet, rawPath, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	return c.do(req, dest)
}

// GetMessagingGraphQL performs a GET to the messaging-specific GraphQL endpoint.
// Uses application/graphql Accept header (not application/json).
func (c *Client) GetMessagingGraphQL(rawPath string, dest interface{}) error {
	req, err := c.newRequest(http.MethodGet, rawPath, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/graphql")
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

// RawGet performs a GET request and returns the raw response body as bytes.
// Used for exploring API endpoints without pre-defined response structs.
func (c *Client) RawGet(path string) ([]byte, int, error) {
	req, err := c.newRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, 0, err
	}
	// Use application/json Accept for graphql paths, otherwise the default.
	if strings.Contains(path, "graphql") {
		req.Header.Set("Accept", "application/json")
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("reading response: %w", err)
	}
	return body, resp.StatusCode, nil
}

// PostMessenger performs a POST to a messenger endpoint with the correct headers.
// Messenger endpoints use text/plain content type and application/json accept,
// unlike the standard Voyager endpoints.
func (c *Client) PostMessenger(path string, body interface{}, dest interface{}) error {
	data, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshaling request body: %w", err)
	}

	req, err := c.newRequest(http.MethodPost, path, nil)
	if err != nil {
		return err
	}
	req.Body = io.NopCloser(bytes.NewReader(data))
	req.ContentLength = int64(len(data))
	req.Header.Set("Content-Type", "text/plain;charset=UTF-8")
	req.Header.Set("Accept", "application/json")

	return c.do(req, dest)
}

// PostMessengerRaw performs a POST to a messenger endpoint with pre-built JSON bytes.
// Messenger endpoints are sensitive to the Chrome utls TLS fingerprint and extra
// headers — using them causes immediate session revocation. Instead, this method
// uses Go's default TLS transport and sends only the essential headers.
func (c *Client) PostMessengerRaw(path string, rawJSON []byte, dest interface{}) error {
	fullURL := c.baseURL + path
	req, err := http.NewRequest(http.MethodPost, fullURL, bytes.NewReader(rawJSON))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("accept", "application/json")
	req.Header.Set("content-type", "text/plain;charset=UTF-8")
	req.Header.Set("csrf-token", c.csrfToken())
	req.Header.Set("x-restli-protocol-version", "2.0.0")

	// Use a separate HTTP client with Go's default TLS (not utls Chrome fingerprint).
	// The messenger POST endpoint rejects requests that come through Chrome-fingerprinted
	// TLS, causing session revocation. Standard Go TLS works (verified via Python http.client).
	messengerClient := &http.Client{
		Timeout: defaultTimeout,
		Transport: &cookieTransport{
			base:       http.DefaultTransport,
			liAt:       c.creds.LiAt,
			jsessionid: c.creds.JSESSIONID,
			bcookie:    c.creds.Bcookie,
			bscookie:   c.creds.Bscookie,
		},
	}

	resp, err := messengerClient.Do(req)
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

// RawPost performs a POST request with a body and returns the raw response.
func (c *Client) RawPost(path string, jsonBody []byte) ([]byte, int, error) {
	req, err := c.newRequest(http.MethodPost, path, nil)
	if err != nil {
		return nil, 0, err
	}
	req.Body = io.NopCloser(bytes.NewReader(jsonBody))
	req.ContentLength = int64(len(jsonBody))
	// Messenger endpoints use text/plain content type and application/json accept.
	if strings.Contains(path, "Messenger") || strings.Contains(path, "messenger") {
		req.Header.Set("Content-Type", "text/plain;charset=UTF-8")
		req.Header.Set("Accept", "application/json")
	} else {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("reading response: %w", err)
	}
	return body, resp.StatusCode, nil
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
