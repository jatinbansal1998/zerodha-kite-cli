package updater

import (
	"fmt"
	"strconv"
	"strings"
)

type semVersion struct {
	major int
	minor int
	patch int
	pre   []string
}

func CompareVersions(a, b string) (int, error) {
	va, err := parseSemVersion(a)
	if err != nil {
		return 0, err
	}
	vb, err := parseSemVersion(b)
	if err != nil {
		return 0, err
	}

	if va.major != vb.major {
		if va.major > vb.major {
			return 1, nil
		}
		return -1, nil
	}
	if va.minor != vb.minor {
		if va.minor > vb.minor {
			return 1, nil
		}
		return -1, nil
	}
	if va.patch != vb.patch {
		if va.patch > vb.patch {
			return 1, nil
		}
		return -1, nil
	}

	return comparePrerelease(va.pre, vb.pre), nil
}

func IsNewerVersion(latest, current string) bool {
	comp, err := CompareVersions(latest, current)
	if err != nil {
		return false
	}
	return comp > 0
}

func parseSemVersion(raw string) (semVersion, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return semVersion{}, fmt.Errorf("invalid version %q", raw)
	}

	value = strings.TrimPrefix(value, "v")
	value = strings.TrimPrefix(value, "V")

	if idx := strings.IndexByte(value, '+'); idx >= 0 {
		value = value[:idx]
	}

	parts := strings.SplitN(value, "-", 2)
	core := parts[0]
	var pre []string
	if len(parts) == 2 && strings.TrimSpace(parts[1]) != "" {
		pre = strings.Split(parts[1], ".")
	}

	coreParts := strings.Split(core, ".")
	if len(coreParts) != 3 {
		return semVersion{}, fmt.Errorf("invalid semantic version %q", raw)
	}

	major, err := parseNumericSemPart(coreParts[0], raw)
	if err != nil {
		return semVersion{}, err
	}
	minor, err := parseNumericSemPart(coreParts[1], raw)
	if err != nil {
		return semVersion{}, err
	}
	patch, err := parseNumericSemPart(coreParts[2], raw)
	if err != nil {
		return semVersion{}, err
	}

	return semVersion{
		major: major,
		minor: minor,
		patch: patch,
		pre:   pre,
	}, nil
}

func parseNumericSemPart(value, raw string) (int, error) {
	if strings.TrimSpace(value) == "" {
		return 0, fmt.Errorf("invalid semantic version %q", raw)
	}
	if len(value) > 1 && value[0] == '0' {
		return 0, fmt.Errorf("invalid semantic version %q", raw)
	}
	n, err := strconv.Atoi(value)
	if err != nil || n < 0 {
		return 0, fmt.Errorf("invalid semantic version %q", raw)
	}
	return n, nil
}

func comparePrerelease(a, b []string) int {
	if len(a) == 0 && len(b) == 0 {
		return 0
	}
	if len(a) == 0 {
		return 1
	}
	if len(b) == 0 {
		return -1
	}

	max := len(a)
	if len(b) > max {
		max = len(b)
	}

	for i := 0; i < max; i++ {
		if i >= len(a) {
			return -1
		}
		if i >= len(b) {
			return 1
		}
		if a[i] == b[i] {
			continue
		}
		return comparePrereleasePart(a[i], b[i])
	}

	return 0
}

func comparePrereleasePart(a, b string) int {
	aNum, aErr := strconv.Atoi(a)
	bNum, bErr := strconv.Atoi(b)

	if aErr == nil && bErr == nil {
		switch {
		case aNum > bNum:
			return 1
		case aNum < bNum:
			return -1
		default:
			return 0
		}
	}
	if aErr == nil {
		return -1
	}
	if bErr == nil {
		return 1
	}

	if a > b {
		return 1
	}
	return -1
}
