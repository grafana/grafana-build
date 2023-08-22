package git

import (
	"context"
	"fmt"
	"net/url"

	"dagger.io/dagger"
)

type GitCloneOptions struct {
	Ref string
	URL string

	SSHKeyPath string

	// Username is injected into the final URL used for cloning
	Username string
	// Password is injected into the final URL used for cloning
	Password string
}

// Clone returns the directory with the cloned repository ('url') and checked out ref ('ref').
func Clone(d *dagger.Client, url, ref string) (*dagger.Directory, error) {
	container, err := CloneContainer(d, &GitCloneOptions{
		URL: url,
		Ref: ref,
	})

	if err != nil {
		return nil, err
	}

	if code, err := container.ExitCode(context.Background()); err != nil || code != 0 {
		return nil, fmt.Errorf("%d: %w", code, err)
	}

	return container.Directory("src"), nil
}

func CloneWithGitHubToken(d *dagger.Client, token, url, ref string) (*dagger.Directory, error) {
	container, err := CloneContainer(d, &GitCloneOptions{
		URL:      url,
		Ref:      ref,
		Username: "x-oauth-token",
		Password: token,
	})
	if err != nil {
		return nil, err
	}

	if code, err := container.ExitCode(context.Background()); err != nil || code != 0 {
		return nil, fmt.Errorf("%d: %w", code, err)
	}

	return container.Directory("src"), nil
}

// CloneWithSSHAuth returns the directory with the cloned repository ('url') and checked out ref ('ref').
func CloneWithSSHAuth(d *dagger.Client, sshKeyPath, url, ref string) (*dagger.Directory, error) {
	container, err := CloneContainer(d, &GitCloneOptions{
		URL:        url,
		Ref:        ref,
		SSHKeyPath: sshKeyPath,
	})

	if err != nil {
		return nil, err
	}

	if code, err := container.ExitCode(context.Background()); err != nil || code != 0 {
		return nil, fmt.Errorf("%d: %w", code, err)
	}

	return container.Directory("src"), nil
}

// injectURLCredentials modifies as provided URL to set the given username and password in it.
func injectURLCredentials(u string, username string, password string) (string, error) {
	rawURL, err := url.Parse(u)
	if err != nil {
		return "", err
	}
	ui := url.UserPassword(username, password)
	rawURL.User = ui
	return rawURL.String(), nil
}
