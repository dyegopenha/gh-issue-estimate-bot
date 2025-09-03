package estimate

import (
	"regexp"
	"strings"
)

// HasEstimate returns true if the issue body contains an estimate in the format:
// "Estimate: X days" (case-insensitive, allows "day"/"days" and decimal numbers).
func HasEstimate(body string) bool {
	if body == "" {
		return false
	}
	// Normalize to NFC-ish simple handling and trim
	s := strings.TrimSpace(body)

	// \bEstimate:\s*\d+(\.\d+)?\s*days?\b
	re := regexp.MustCompile(`(?i)\bEstimate:\s*\d+(\.\d+)?\s*days?\b`)
	return re.FindStringIndex(s) != nil
}
