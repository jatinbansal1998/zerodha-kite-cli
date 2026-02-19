package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	kiteconnect "github.com/zerodha/gokiteconnect/v4"

	"github.com/jatinbansal1998/zerodha-kite-cli/internal/config"
	"github.com/jatinbansal1998/zerodha-kite-cli/internal/exitcode"
	"github.com/jatinbansal1998/zerodha-kite-cli/internal/output"
	"github.com/jatinbansal1998/zerodha-kite-cli/internal/paths"
)

type commandContext struct {
	opts  *rootOptions
	store *config.FileStore
	cfg   config.Config
}

func newCommandContext(opts *rootOptions) (*commandContext, error) {
	configPath := strings.TrimSpace(opts.configPath)
	if configPath == "" {
		path, err := paths.DefaultConfigPath()
		if err != nil {
			return nil, exitcode.Wrap(exitcode.Config, "resolve config path", err)
		}
		configPath = path
	}

	store := config.NewFileStore(configPath)
	cfg, err := store.Load()
	if err != nil {
		return nil, exitcode.Wrap(exitcode.Config, "load config", err)
	}
	if cacheDir, err := paths.DefaultCacheDir(); err == nil {
		_ = os.MkdirAll(cacheDir, 0o700)
	}

	return &commandContext{
		opts:  opts,
		store: store,
		cfg:   cfg,
	}, nil
}

func (c *commandContext) save() error {
	if err := c.store.Save(c.cfg); err != nil {
		return exitcode.Wrap(exitcode.Config, "save config", err)
	}
	return nil
}

func (c *commandContext) printer(w io.Writer) output.Printer {
	return output.New(w, c.opts.outputJSON)
}

func (c *commandContext) resolveProfile(require bool) (string, *config.Profile, error) {
	name := strings.TrimSpace(c.opts.profile)
	if name == "" {
		name = strings.TrimSpace(c.cfg.ActiveProfile)
	}
	if name == "" && len(c.cfg.Profiles) == 1 {
		for n := range c.cfg.Profiles {
			name = n
		}
	}
	if name == "" {
		if require {
			return "", nil, exitcode.New(exitcode.Config, "no profile selected; set one with `zerodha config profile use <name>` or pass --profile")
		}
		return "", nil, nil
	}

	profile, ok := c.cfg.Profiles[name]
	if !ok {
		return "", nil, exitcode.New(exitcode.Config, fmt.Sprintf("profile %q not found", name))
	}

	return name, &profile, nil
}

func (c *commandContext) setProfile(name string, profile config.Profile) {
	if c.cfg.Profiles == nil {
		c.cfg.Profiles = make(map[string]config.Profile)
	}
	c.cfg.Profiles[name] = profile
}

func (c *commandContext) deleteProfile(name string) {
	delete(c.cfg.Profiles, name)
}

func (c *commandContext) profileNames() []string {
	names := make([]string, 0, len(c.cfg.Profiles))
	for name := range c.cfg.Profiles {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func newKiteClient(profile config.Profile, debug bool) *kiteconnect.Client {
	client := kiteconnect.New(profile.APIKey)
	client.SetDebug(debug)
	if profile.AccessToken != "" {
		client.SetAccessToken(profile.AccessToken)
	}
	return client
}

func callWithAuthRetry[T any](
	ctx *commandContext,
	profileName string,
	profile *config.Profile,
	fn func(*kiteconnect.Client) (T, error),
) (T, error) {
	var zero T

	client := newKiteClient(*profile, ctx.opts.debug)
	value, err := fn(client)
	if err == nil {
		return value, nil
	}

	if !isTokenError(err) {
		return zero, wrapKiteError("kite api call failed", err)
	}

	if profile.RefreshToken == "" || profile.APISecret == "" {
		return zero, exitcode.Wrap(exitcode.Auth, "access token expired and refresh token is unavailable; run `zerodha auth login`", err)
	}

	renewed, renewErr := client.RenewAccessToken(profile.RefreshToken, profile.APISecret)
	if renewErr != nil {
		return zero, exitcode.Wrap(exitcode.Auth, "failed to refresh access token; run `zerodha auth login`", renewErr)
	}
	if renewed.AccessToken == "" {
		return zero, exitcode.New(exitcode.Auth, "token refresh returned empty access token")
	}

	profile.AccessToken = renewed.AccessToken
	if renewed.RefreshToken != "" {
		profile.RefreshToken = renewed.RefreshToken
	}
	profile.LastLoginAt = time.Now().UTC()
	ctx.setProfile(profileName, *profile)
	if err := ctx.save(); err != nil {
		return zero, err
	}

	client.SetAccessToken(profile.AccessToken)
	value, err = fn(client)
	if err != nil {
		return zero, wrapKiteError("kite api call failed after token refresh", err)
	}

	return value, nil
}

func isTokenError(err error) bool {
	if kErr, ok := errors.AsType[kiteconnect.Error](err); ok {
		return kErr.ErrorType == kiteconnect.TokenError ||
			kErr.ErrorType == kiteconnect.PermissionError ||
			kErr.ErrorType == kiteconnect.TwoFAError
	}
	return false
}

func wrapKiteError(msg string, err error) error {
	return exitcode.Wrap(exitcode.Code(err), msg, err)
}

func ensureProfileCredentials(profile *config.Profile) error {
	if strings.TrimSpace(profile.APIKey) == "" {
		return exitcode.New(exitcode.Config, "profile missing api_key; set via `zerodha config profile add` or `zerodha config profile set-api-key`")
	}
	if strings.TrimSpace(profile.APISecret) == "" {
		return exitcode.New(exitcode.Config, "profile missing api_secret; set via `zerodha config profile add` or `zerodha config profile set-api-secret`")
	}
	return nil
}

func ensureAccessToken(profile *config.Profile) error {
	if strings.TrimSpace(profile.AccessToken) == "" {
		return exitcode.New(exitcode.Auth, "profile has no access token; run `zerodha auth login`")
	}
	return nil
}

func validateAuthLoginFlags(
	useCallback bool,
	rawRequestToken string,
	callbackPort int,
	callbackPortChanged bool,
) (string, error) {
	if callbackPort < 1 || callbackPort > 65535 {
		return "", exitcode.New(exitcode.Validation, "--callback-port must be between 1 and 65535")
	}

	if callbackPortChanged && !useCallback {
		return "", exitcode.New(exitcode.Validation, "--callback-port can only be used with --callback")
	}

	requestToken := extractRequestToken(rawRequestToken)
	if useCallback {
		if strings.TrimSpace(requestToken) != "" {
			return "", exitcode.New(exitcode.Validation, "--request-token cannot be used with --callback")
		}
		return "", nil
	}

	if strings.TrimSpace(requestToken) == "" {
		return "", exitcode.New(exitcode.Validation, "exactly one login mode is required: pass --request-token or --callback")
	}

	return requestToken, nil
}

func extractRequestToken(input string) string {
	input = strings.TrimSpace(input)
	if input == "" {
		return ""
	}

	if strings.Contains(input, "request_token=") {
		if parsedURL, err := url.Parse(input); err == nil {
			if token := parsedURL.Query().Get("request_token"); token != "" {
				return token
			}
		}

		if values, err := url.ParseQuery(input); err == nil {
			if token := values.Get("request_token"); token != "" {
				return token
			}
		}
	}

	return input
}

func captureRequestToken(port int, timeout time.Duration) (string, error) {
	tokenCh := make(chan string, 1)
	errCh := make(chan error, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		token := strings.TrimSpace(r.URL.Query().Get("request_token"))
		if token == "" {
			http.Error(w, "missing request_token", http.StatusBadRequest)
			return
		}
		_, _ = fmt.Fprintln(w, "Login complete. You can return to the terminal.")
		select {
		case tokenCh <- token:
		default:
		}
	})

	server := &http.Server{
		Addr:              fmt.Sprintf("127.0.0.1:%d", port),
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	defer func() {
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer shutdownCancel()
		_ = server.Shutdown(shutdownCtx)
	}()

	select {
	case token := <-tokenCh:
		return token, nil
	case err := <-errCh:
		return "", err
	case <-ctx.Done():
		return "", errors.New("timed out waiting for request_token callback")
	}
}

func firstRemainingProfile(names []string, excluded string) string {
	for _, name := range names {
		if name != excluded {
			return name
		}
	}
	return ""
}

func nowUTC() time.Time {
	return time.Now().UTC()
}

func intToString(v int) string {
	return strconv.Itoa(v)
}

func formatModelTime(t time.Time) string {
	if t.IsZero() {
		return "-"
	}
	return t.Format("2006-01-02 15:04:05")
}
