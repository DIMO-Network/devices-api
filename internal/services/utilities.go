package services

import (
	"regexp"
	"strings"

	"github.com/ethereum/go-ethereum/common"
)

// don't let this become a dumping ground!

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

var zeroAddr common.Address

// IsZeroAddress validate if it's a 0 address
func IsZeroAddress(addr common.Address) bool {
	return addr == zeroAddr
}
