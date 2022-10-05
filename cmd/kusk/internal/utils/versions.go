package utils

import (
	"github.com/hashicorp/go-version"
)

func IsUptodate(latest, current string) bool {
	latestVersion, err := version.NewVersion(latest)
	if err != nil {
		return false
	}

	currentVersion, err := version.NewVersion(current)
	if err != nil {
		return false
	}

	return latestVersion.LessThanOrEqual(currentVersion)
}
