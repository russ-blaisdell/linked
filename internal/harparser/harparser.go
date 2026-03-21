package harparser

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/russ-blaisdell/linked/internal/models"
)

// HAR 1.2 types — only the fields we need.
type harFile struct {
	Log struct {
		Entries []entry `json:"entries"`
	} `json:"log"`
}

type entry struct {
	Request struct {
		URL     string   `json:"url"`
		Headers []header `json:"headers"`
		Cookies []cookie `json:"cookies"`
	} `json:"request"`
}

type header struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type cookie struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// Result holds data extracted from a HAR file.
type Result struct {
	Fingerprint *models.BrowserFingerprint
	LiAt        string
	JSESSIONID  string
	Bcookie     string
	Bscookie    string
	HasCookies  bool // true if at least li_at was found
}

// Parse reads a HAR file and extracts browser fingerprint headers and
// (if present) authentication cookies from LinkedIn Voyager API requests.
func Parse(path string) (*Result, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading HAR file: %w", err)
	}

	var har harFile
	if err := json.Unmarshal(data, &har); err != nil {
		// HAR files can be truncated (e.g. large captures). Try to extract
		// what we can by finding the first voyager request in the raw JSON.
		return parsePartial(data)
	}

	return extract(har)
}

func extract(har harFile) (*Result, error) {
	result := &Result{
		Fingerprint: &models.BrowserFingerprint{},
	}

	// Find the first Voyager API request for fingerprint headers.
	var found bool
	for _, e := range har.Log.Entries {
		if !strings.Contains(e.Request.URL, "/voyager/api/") {
			continue
		}
		found = true
		for _, h := range e.Request.Headers {
			switch strings.ToLower(h.Name) {
			case "user-agent":
				result.Fingerprint.UserAgent = h.Value
			case "sec-ch-ua":
				result.Fingerprint.SecChUA = h.Value
			case "sec-ch-ua-mobile":
				result.Fingerprint.SecChUAMobile = h.Value
			case "sec-ch-ua-platform":
				result.Fingerprint.SecChUAPlatform = h.Value
			case "x-li-track":
				result.Fingerprint.XLiTrack = h.Value
			}
		}
		break
	}

	if !found {
		return nil, fmt.Errorf("no LinkedIn Voyager API requests found in HAR file — ensure the HAR was captured while using linkedin.com")
	}

	// Scan all entries for cookies (Chrome often strips them, but some
	// exporters include them in the cookies array or the Cookie header).
	for _, e := range har.Log.Entries {
		if !strings.Contains(e.Request.URL, "linkedin.com") {
			continue
		}
		// Check parsed cookies array.
		for _, c := range e.Request.Cookies {
			setCookie(result, c.Name, c.Value)
		}
		// Check raw Cookie header (some exporters put cookies there instead).
		for _, h := range e.Request.Headers {
			if strings.EqualFold(h.Name, "cookie") {
				parseCookieHeader(result, h.Value)
			}
		}
		if result.HasCookies {
			break
		}
	}

	return result, nil
}

func setCookie(r *Result, name, value string) {
	switch name {
	case "li_at":
		r.LiAt = value
		r.HasCookies = true
	case "JSESSIONID":
		r.JSESSIONID = value
	case "bcookie":
		r.Bcookie = value
	case "bscookie":
		r.Bscookie = value
	}
}

func parseCookieHeader(r *Result, raw string) {
	for _, part := range strings.Split(raw, "; ") {
		name, value, ok := strings.Cut(part, "=")
		if !ok {
			continue
		}
		setCookie(r, strings.TrimSpace(name), value)
	}
}

// parsePartial handles truncated HAR files by scanning the raw JSON for
// the first voyager request headers using a simple state machine.
func parsePartial(data []byte) (*Result, error) {
	s := string(data)

	// Find first voyager API request
	idx := strings.Index(s, "/voyager/api/")
	if idx < 0 {
		return nil, fmt.Errorf("no LinkedIn Voyager API requests found in HAR file")
	}

	result := &Result{
		Fingerprint: &models.BrowserFingerprint{},
	}

	// Extract headers near that request using simple string matching.
	// Look backwards and forwards from the voyager URL for the entry's headers.
	// We search a generous window around the URL occurrence.
	start := max(0, idx-5000)
	end := min(len(s), idx+20000)
	window := s[start:end]

	extractHeader := func(name string) string {
		// Match: "name": "<name>", "value": "<value>"
		// Also handles: "name":"<name>","value":"<value>"
		lower := strings.ToLower(window)
		target := strings.ToLower(name)

		// Find the header name
		searchIdx := 0
		for {
			pos := strings.Index(lower[searchIdx:], `"`+target+`"`)
			if pos < 0 {
				return ""
			}
			pos += searchIdx
			// Find the value field after this
			after := window[pos:]
			valIdx := strings.Index(after, `"value"`)
			if valIdx < 0 || valIdx > 200 {
				searchIdx = pos + 1
				continue
			}
			// Find the value string
			valStart := strings.Index(after[valIdx+7:], `"`)
			if valStart < 0 {
				return ""
			}
			valContent := after[valIdx+7+valStart+1:]
			valEnd := strings.Index(valContent, `"`)
			if valEnd < 0 {
				return ""
			}
			val := valContent[:valEnd]
			// Unescape JSON string escapes
			val = strings.ReplaceAll(val, `\"`, `"`)
			val = strings.ReplaceAll(val, `\\`, `\`)
			return val
		}
	}

	result.Fingerprint.UserAgent = extractHeader("user-agent")
	result.Fingerprint.SecChUA = extractHeader("sec-ch-ua")
	result.Fingerprint.SecChUAMobile = extractHeader("sec-ch-ua-mobile")
	result.Fingerprint.SecChUAPlatform = extractHeader("sec-ch-ua-platform")
	result.Fingerprint.XLiTrack = extractHeader("x-li-track")

	if result.Fingerprint.UserAgent == "" {
		return nil, fmt.Errorf("could not extract browser fingerprint from HAR file")
	}

	return result, nil
}
