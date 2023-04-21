package containers

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"dagger.io/dagger"
)

type GitCloneOptions struct {
	Ref   string
	URL   string
	Depth int

	SSHKeyPath string

	// Username is injected into the final URL used for cloning
	Username string
	// Password is injected into the final URL used for cloning
	Password string
}

// CloneContainer returns the container definition that uses git clone to clone the 'git url' and checks out the ref provided at 'ref'.
// Multiple refs can be provided via a space character (' '). If multiple refs are provided, then the container will attempt to checkout
// each ref at a time, stopping at the first one that is successful.
// This can be useful in PRs which have a coupled association with another codebase.
// A practical example (and why this exists): "${pr_source_branch} ${pr_target_branch} ${main}" will first attempt to checkout the PR source branch, then the PR target branch, then "main"; whichever is successul first.
func CloneContainer(d *dagger.Client, opts *GitCloneOptions) (*dagger.Container, error) {
	var err error
	if opts.URL == "" {
		return nil, errors.New("URL can not be empty")
	}

	if opts.SSHKeyPath != "" && (opts.Username != "" || opts.Password != "") {
		return nil, fmt.Errorf("conflicting options: use either username/password or an SSH key")
	}

	cloneURL := opts.URL
	if opts.Username != "" && opts.Password != "" {
		cloneURL, err = injectURLCredentials(cloneURL, opts.Username, opts.Password)
		if err != nil {
			return nil, fmt.Errorf("failed to inject credentials into cloning URL: %w", err)
		}
	}

	cloneArgs := []string{"git", "clone"}
	if opts.Depth != 0 {
		cloneArgs = append(cloneArgs, "--depth", strconv.Itoa(opts.Depth))
	}

	cloneArgs = append(cloneArgs, "${GIT_CLONE_URL}", "src")

	container := d.Container().From(GitImage).
		WithEntrypoint([]string{})

	if opts.SSHKeyPath != "" {
		if !strings.Contains(opts.URL, "@") {
			return nil, errors.New("git URL with SSH needs an '@'")
		}
		if !strings.Contains(opts.URL, ":") {
			return nil, errors.New("git URL with SSH needs a ':'")
		}

		host := opts.URL[strings.Index(opts.URL, "@")+1 : strings.Index(opts.URL, ":")]

		container = container.
			WithExec([]string{"mkdir", "-p", "/root/.ssh"}).
			WithMountedFile("/root/.ssh/id_rsa", d.Host().Directory(filepath.Dir(opts.SSHKeyPath)).File(filepath.Base(opts.SSHKeyPath))).
			WithExec([]string{"/bin/sh", "-c", fmt.Sprintf(`ssh-keyscan %s > /root/.ssh/known_hosts`, host)})
	}

	cloneURLSecret := d.SetSecret("git-clone-url", cloneURL)
	container = container.
		WithSecretVariable("GIT_CLONE_URL", cloneURLSecret).
		WithExec([]string{"/bin/sh", "-c", strings.Join(cloneArgs, " ")})

	ref := "main"
	if opts.Ref != "" {
		ref = opts.Ref
	}

	// TODO: this section really needs to be its own function with unit tests, or an interface or something.
	var (
		checkouts    = strings.Split(ref, " ")
		checkoutArgs = []string{fmt.Sprintf(`if git -C src checkout %[1]s; then echo "checked out %[1]s";`, checkouts[0])}
	)

	for _, v := range checkouts[1:] {
		checkoutArgs = append(checkoutArgs, fmt.Sprintf(`elif git -C src checkout %[1]s; then echo "checked out %[1]s";`, v))
	}

	checkoutArgs = append(checkoutArgs, "else exit 3; fi")

	container = container.WithExec([]string{"/bin/sh", "-c", strings.Join(checkoutArgs, " ")})
	log.Println(strings.Join(checkoutArgs, " "))
	return container, nil
}

// Clone returns the directory with the cloned repository ('url') and checked out ref ('ref').
func Clone(d *dagger.Client, url, ref string) (*dagger.Directory, error) {
	container, err := CloneContainer(d, &GitCloneOptions{
		URL: url,
		Ref: ref,
		//Depth: 1,
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
		Depth:    1,
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

func MountLocalDir(d *dagger.Client, localPath string) (*dagger.Directory, error) {
	_, err := os.Stat(localPath)
	if err != nil {
		return nil, err
	}
	src := d.Host().Directory(localPath)
	return src, nil
}

// CloneWithSSHAuth returns the directory with the cloned repository ('url') and checked out ref ('ref').
func CloneWithSSHAuth(d *dagger.Client, sshKeyPath, url, ref string) (*dagger.Directory, error) {
	container, err := CloneContainer(d, &GitCloneOptions{
		URL:        url,
		Ref:        ref,
		Depth:      1,
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
