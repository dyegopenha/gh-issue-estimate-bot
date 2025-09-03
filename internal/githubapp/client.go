package githubapp

import (
	"net/http"
	"os"
	"strconv"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v66/github"
)

// NewInstallationClient returns a go-github client for the given installationID using the app's private key.
func NewInstallationClient(installationID int64) (*github.Client, error) {
	appIDStr := os.Getenv("APP_ID")
	privateKeyPath := os.Getenv("PRIVATE_KEY_PATH")
	if appIDStr == "" || privateKeyPath == "" {
		return nil, ErrMissingConfig
	}
	appID, err := strconv.ParseInt(appIDStr, 10, 64)
	if err != nil {
		return nil, err
	}

	tr := http.DefaultTransport
	itr, err := ghinstallation.NewKeyFromFile(tr, appID, installationID, privateKeyPath)
	if err != nil {
		return nil, err
	}

	client := github.NewClient(&http.Client{Transport: itr})
	return client, nil
}

var ErrMissingConfig = &ConfigError{Msg: "APP_ID and PRIVATE_KEY_PATH must be set"}

type ConfigError struct{ Msg string }
func (e *ConfigError) Error() string { return e.Msg }
