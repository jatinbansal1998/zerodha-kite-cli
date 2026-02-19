package updater

import (
	"fmt"
	"strings"
)

func SummarizeState(currentVersion string, state State) string {
	if state.ApplyPending && strings.TrimSpace(state.DownloadedVersion) != "" {
		return fmt.Sprintf("update: %s downloaded; apply pending", state.DownloadedVersion)
	}
	if state.LastCheckedAt.IsZero() {
		return "update: not checked yet"
	}
	if strings.TrimSpace(state.LastError) != "" {
		return fmt.Sprintf("update: last check failed: %s", truncate(state.LastError, 140))
	}
	if strings.TrimSpace(state.LatestVersionSeen) != "" && IsNewerVersion(state.LatestVersionSeen, currentVersion) {
		return fmt.Sprintf("update: %s available", state.LatestVersionSeen)
	}
	return "update: up to date"
}

func truncate(value string, max int) string {
	if len(value) <= max {
		return value
	}
	if max <= 3 {
		return value[:max]
	}
	return value[:max-3] + "..."
}
