package utils

import (
	"crypto/rand"
	"fmt"
	"regexp"
	"strings"
	"time"
)

const (
	// defaultRandomSuffixLength controls the random suffix length when no custom suffix is provided.
	defaultRandomSuffixLength = 4
)

var (
	// shortSuffixAlphabet contains safe characters for username/email/path suffixes.
	shortSuffixAlphabet = []byte("abcdefghijklmnopqrstuvwxyz0123456789")
)

// GetVisibility returns the visibility value, defaulting to private.
func GetVisibility(v string) string {
	if v == "" {
		return "private"
	}
	return v
}

// GenerateTimestampSuffix returns a millisecond-level timestamp suffix in yyyyMMddHHmmssSSS format.
func GenerateTimestampSuffix() string {
	now := time.Now()
	millisecond := now.Nanosecond() / int(time.Millisecond)
	return fmt.Sprintf("%s%03d", now.Format("20060102150405"), millisecond)
}

// GenerateTemporalSuffix returns a unique suffix in the format timestamp-randomOrCustom.
func GenerateTemporalSuffix(customSuffix string) string {
	normalizedSuffix := normalizeCustomSuffix(customSuffix)
	if normalizedSuffix == "" {
		normalizedSuffix = generateShortRandomSuffix(defaultRandomSuffixLength)
	}
	return fmt.Sprintf("%s-%s", GenerateTimestampSuffix(), normalizedSuffix)
}

// GenerateUsernameWithTimestamp builds a GitLab-safe username as prefix-temporalSuffix.
func GenerateUsernameWithTimestamp(prefix, customSuffix string) string {
	temporalSuffix := GenerateTemporalSuffix(customSuffix)
	username := fmt.Sprintf("%s-%s", prefix, temporalSuffix)
	// Ensure username follows GitLab username rules.
	username = sanitizeUsername(username)
	// Enforce GitLab username max length.
	if len(username) > 255 {
		username = username[:255]
	}
	return username
}

// GenerateEmailWithTimestamp builds an email as localPart-temporalSuffix@domain.
func GenerateEmailWithTimestamp(emailPrefix, customSuffix string) string {
	// Split local part and domain.
	parts := strings.Split(emailPrefix, "@")
	if len(parts) != 2 {
		// Fall back to a default test domain when input is not an email.
		return fmt.Sprintf("%s-%s@test.example.com", emailPrefix, GenerateTemporalSuffix(customSuffix))
	}

	localPart := parts[0]
	domain := parts[1]
	temporalSuffix := GenerateTemporalSuffix(customSuffix)

	return fmt.Sprintf("%s-%s@%s", localPart, temporalSuffix, domain)
}

// GenerateGroupPathWithTimestamp builds a GitLab-safe group path as prefix-temporalSuffix.
func GenerateGroupPathWithTimestamp(prefix, customSuffix string) string {
	temporalSuffix := GenerateTemporalSuffix(customSuffix)
	path := fmt.Sprintf("%s-%s", prefix, temporalSuffix)
	// Ensure path follows GitLab group path rules.
	path = sanitizeGroupPath(path)
	// Enforce GitLab path max length.
	if len(path) > 255 {
		path = path[:255]
	}
	return path
}

// GenerateProjectPathWithTimestamp builds a GitLab-safe project path as prefix-temporalSuffix.
func GenerateProjectPathWithTimestamp(prefix, customSuffix string) string {
	return GenerateGroupPathWithTimestamp(prefix, customSuffix)
}

// normalizeCustomSuffix sanitizes a custom suffix to safe characters and lowercase.
func normalizeCustomSuffix(customSuffix string) string {
	trimmedSuffix := strings.TrimSpace(customSuffix)
	if trimmedSuffix == "" {
		return ""
	}

	safeSuffix := regexp.MustCompile(`[^a-zA-Z0-9_-]`).ReplaceAllString(trimmedSuffix, "")
	return strings.ToLower(safeSuffix)
}

// generateShortRandomSuffix returns a random lowercase alphanumeric suffix.
func generateShortRandomSuffix(length int) string {
	if length <= 0 {
		length = defaultRandomSuffixLength
	}

	rawRandomBytes := make([]byte, length)
	_, randomErr := rand.Read(rawRandomBytes)
	if randomErr != nil {
		fallbackFromTime := fmt.Sprintf("%d", time.Now().UnixNano())
		if len(fallbackFromTime) > length {
			return fallbackFromTime[len(fallbackFromTime)-length:]
		}
		return fallbackFromTime
	}

	randomSuffix := make([]byte, length)
	for i := range rawRandomBytes {
		randomSuffix[i] = shortSuffixAlphabet[int(rawRandomBytes[i])%len(shortSuffixAlphabet)]
	}

	return string(randomSuffix)
}

// sanitizeUsername sanitizes username to comply with GitLab username rules.
func sanitizeUsername(username string) string {
	// Remove unsupported characters.
	reg := regexp.MustCompile(`[^a-zA-Z0-9_.-]`)
	username = reg.ReplaceAllString(username, "")

	// Ensure username does not start/end with hyphen or dot.
	username = strings.Trim(username, "-.")

	return username
}

// sanitizeGroupPath sanitizes group/project path to comply with GitLab path rules.
func sanitizeGroupPath(path string) string {
	// Normalize to lowercase.
	path = strings.ToLower(path)

	// Remove unsupported characters.
	reg := regexp.MustCompile(`[^a-z0-9_-]`)
	path = reg.ReplaceAllString(path, "")

	// Ensure path does not start/end with hyphen.
	path = strings.Trim(path, "-")

	return path
}
