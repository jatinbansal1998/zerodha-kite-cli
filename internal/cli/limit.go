package cli

import (
	"fmt"

	"github.com/jatinbansal1998/zerodha-kite-cli/internal/exitcode"
)

func validateLimit(limit int) error {
	if limit < 0 {
		return exitcode.New(exitcode.Validation, fmt.Sprintf("invalid --limit %d: must be >= 0", limit))
	}
	return nil
}

func applyLimit[S ~[]E, E any](items S, limit int) S {
	if limit <= 0 || limit >= len(items) {
		return items
	}
	return items[:limit]
}
