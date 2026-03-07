package api

import (
	"crypto/rand"
	"encoding/base64"
	"strings"
)

// urnToID extracts the trailing ID segment from a LinkedIn URN.
// e.g. "urn:li:member:12345678" → "12345678"
func urnToID(urn string) string {
	if urn == "" {
		return ""
	}
	parts := strings.Split(urn, ":")
	return parts[len(parts)-1]
}

// generateTrackingID creates a random base64 tracking token used in some API calls.
func generateTrackingID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}
