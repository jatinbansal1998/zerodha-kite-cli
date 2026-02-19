package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/jatinbansal1998/zerodha-kite-cli/internal/buildinfo"
	"github.com/jatinbansal1998/zerodha-kite-cli/internal/exitcode"
	"github.com/jatinbansal1998/zerodha-kite-cli/internal/output"
	"github.com/jatinbansal1998/zerodha-kite-cli/internal/paths"
	"github.com/jatinbansal1998/zerodha-kite-cli/internal/updater"
	"github.com/spf13/cobra"
)

type updateResult struct {
	Status         string `json:"status"`
	Message        string `json:"message"`
	CurrentVersion string `json:"current_version"`
	LatestVersion  string `json:"latest_version,omitempty"`
	UpdatedVersion string `json:"updated_version,omitempty"`
	PendingVersion string `json:"pending_version,omitempty"`
	ApplyPending   bool   `json:"apply_pending"`
}

func newUpdateCmd(opts *rootOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "Check for updates and apply the latest release",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cacheDir, err := paths.DefaultCacheDir()
			if err != nil {
				return exitcode.Wrap(exitcode.Internal, "resolve cache directory", err)
			}

			executablePath, err := os.Executable()
			if err != nil {
				return exitcode.Wrap(exitcode.Internal, "resolve executable path", err)
			}

			currentVersion := strings.TrimSpace(buildinfo.Version)
			state, err := updater.RunManual(updater.Options{
				CurrentVersion: currentVersion,
				ExecutablePath: executablePath,
				CacheDir:       cacheDir,
				RepoOwner:      updater.DefaultRepoOwner,
				RepoName:       updater.DefaultRepoName,
				Cooldown:       updater.DefaultCooldown,
			})
			if err != nil {
				return exitcode.Wrap(exitcode.Code(err), "manual update failed", err)
			}

			result := summarizeManualUpdate(currentVersion, state)
			printer := output.New(cmd.OutOrStdout(), opts.outputJSON)
			if printer.IsJSON() {
				return printer.JSON(result)
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), result.Message)
			return err
		},
	}
}

func summarizeManualUpdate(currentVersion string, state updater.State) updateResult {
	latestVersion := strings.TrimSpace(state.LatestVersionSeen)
	result := updateResult{
		Status:         "up_to_date",
		Message:        fmt.Sprintf("already up to date (%s)", currentVersion),
		CurrentVersion: currentVersion,
		LatestVersion:  latestVersion,
		ApplyPending:   state.ApplyPending,
	}

	pendingVersion := strings.TrimSpace(state.DownloadedVersion)
	if state.ApplyPending && pendingVersion != "" {
		result.Status = "apply_pending"
		result.PendingVersion = pendingVersion
		result.Message = fmt.Sprintf("update downloaded: %s (apply pending)", pendingVersion)
		return result
	}

	updatedVersion := strings.TrimSpace(state.CurrentVersion)
	if updatedVersion != "" && updater.IsNewerVersion(updatedVersion, currentVersion) {
		result.Status = "updated"
		result.UpdatedVersion = updatedVersion
		result.Message = fmt.Sprintf("updated from %s to %s", currentVersion, updatedVersion)
	}

	return result
}
