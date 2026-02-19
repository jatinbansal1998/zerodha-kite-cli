package updater

type ApplyHelperRequest struct {
	TargetPath string
	SourcePath string
	CacheDir   string
	Version    string
}
