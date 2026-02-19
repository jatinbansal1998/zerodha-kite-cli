//go:build !windows

package updater

import "errors"

func SpawnApplyHelper(_ string, _ ApplyHelperRequest) error {
	return errors.New("windows helper is not supported on this platform")
}

func RunApplyHelper(_ ApplyHelperRequest) error {
	return errors.New("windows helper is not supported on this platform")
}
