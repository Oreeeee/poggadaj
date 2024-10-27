package utils

import "strings"

func IsGG77(version uint32) bool {
	fullVersion, exists := GGVersions[version]
	if !exists || !strings.Contains(fullVersion, "Gadu-Gadu 7.7") {
		return false
	}
	return true
}
