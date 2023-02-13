package services

import (
	"regexp"
	"strings"
)

// don't let this become a dumping ground!

// Contains returns true if string exist in slice
func Contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

// ValidateAndCleanUUID returns false if uuid is not RFC 4122 - UUIDv2, which is what AutoPi uses. if valid, returns a lowercased and empty space trimmed of input uuid.
func ValidateAndCleanUUID(uuid string) (bool, string) {
	uuid = strings.TrimSpace(strings.ToLower(uuid))

	pattern := "^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}"
	res, _ := regexp.MatchString(pattern, uuid)
	if res {
		return true, uuid
	}
	return false, ""
}
