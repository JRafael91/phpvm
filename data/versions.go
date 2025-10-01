package data

import (
	"strings"
	"time"
)

// PHPVersion represents a PHP version with its metadata
type PHPVersion struct {
	Version        string
	Released       time.Time
	BinaryURLx64   string
	BinaryURLarm64 string
}

type ComposerVersion struct {
	Version        string
	Released       time.Time
	URL            string
	MinPHPVersion  string
	MaxPHPVersion  string
	CompatiblePHP  []string // Specific PHP versions tested/confirmed compatible
}

// AvailableVersions contains all available PHP versions
var AvailableVersions = []PHPVersion{
	{
		Version:        "8.4.1",
		Released:       time.Date(2024, 11, 21, 0, 0, 0, 0, time.UTC),
		BinaryURLx64:   "https://download.herdphp.com/herd-lite/linux/x64/8.4/php",
		BinaryURLarm64: "https://download.herdphp.com/herd-lite/linux/arm64/8.4/php",
	},
}

var AvailableComposerVersions = []ComposerVersion{
	{
		Version:        "2.8.11",
		Released:       time.Date(2024, 8, 21, 0, 0, 0, 0, time.UTC),
		URL:            "https://getcomposer.org/download/2.8.11/composer.phar",
		MinPHPVersion:  "7.2.5",
		MaxPHPVersion:  "8.4.99",
		CompatiblePHP:  []string{"8.0", "8.1", "8.2", "8.3", "8.4"},
	},
	{
		Version:        "2.7.9",
		Released:       time.Date(2024, 6, 4, 0, 0, 0, 0, time.UTC),
		URL:            "https://getcomposer.org/download/2.7.9/composer.phar",
		MinPHPVersion:  "7.2.5",
		MaxPHPVersion:  "8.3.99",
		CompatiblePHP:  []string{"8.0", "8.1", "8.2", "8.3"},
	},
}

// GetCompatibleComposerVersion returns the best Composer version for a given PHP version
func GetCompatibleComposerVersion(phpVersion string) *ComposerVersion {
	// Extract major.minor from PHP version (e.g., "8.4.1" -> "8.4")
	phpMajorMinor := extractMajorMinor(phpVersion)
	
	// Find the latest compatible Composer version
	var bestComposer *ComposerVersion
	for i := range AvailableComposerVersions {
		composer := &AvailableComposerVersions[i]
		
		// Check if PHP version is in compatible list
		for _, compatiblePHP := range composer.CompatiblePHP {
			if compatiblePHP == phpMajorMinor {
				if bestComposer == nil || composer.Released.After(bestComposer.Released) {
					bestComposer = composer
				}
				break
			}
		}
	}
	
	return bestComposer
}

// extractMajorMinor extracts major.minor version from a full version string
func extractMajorMinor(version string) string {
	parts := strings.Split(version, ".")
	if len(parts) >= 2 {
		return parts[0] + "." + parts[1]
	}
	return version
}
