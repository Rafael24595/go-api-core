package utils

import (
	"fmt"
	"regexp"
	"strconv"
)

const SEMVER = `^([a-zA-Z.]*?)?(\d+)(?:\.(\d+))?(?:\.(\d+))?(?:-([0-9A-Za-z.-]+))?$`

type Version struct {
	Prefix     string
	Major      int
	Minor      int
	Patch      int
	PreRelease string
}

func ParseVersion(v string) (*Version, error) {
	re := regexp.MustCompile(SEMVER)
	matches := re.FindStringSubmatch(v)
	if matches == nil {
		return nil, fmt.Errorf("invalid version format: %s", v)
	}

	prefix := matches[1]

	major, err := strconv.Atoi(matches[2])
	if err != nil {
		return nil, fmt.Errorf("invalid major: %w", err)
	}

	minor := 0
	if matches[3] != "" {
		minor, err = strconv.Atoi(matches[3])
		if err != nil {
			return nil, fmt.Errorf("invalid minor: %w", err)
		}
	}

	patch := 0
	if matches[4] != "" {
		patch, err = strconv.Atoi(matches[4])
		if err != nil {
			return nil, fmt.Errorf("invalid patch: %w", err)
		}
	}

	preRelease := matches[5]

	return &Version{
		Prefix:     prefix,
		Major:      major,
		Minor:      minor,
		Patch:      patch,
		PreRelease: preRelease,
	}, nil
}
