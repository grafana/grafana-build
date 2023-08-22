package pipeline

import (
	"context"
	"errors"

	"dagger.io/dagger"
	"github.com/grafana/grafana-build/cliutil"
	"github.com/urfave/cli/v2"
)

type ArgumentType int

const (
	ArgumentTypeString ArgumentType = iota
	ArgumentTypeInt64
	ArgumentTypeDirectory
	ArgumentTypeFile
	ArgumentTypeBool
)

// An Argument is an input to a artifact command.
// It wraps the concept of a general CLI "Flag" to allow it to
// All arguments are required.
type Argument struct {
	ArgumentType ArgumentType
	Name         string
	Description  string
	ValueFunc    func(ctx context.Context, c cliutil.CLIContext, d *dagger.Client) (any, error)
	Flags        []cli.Flag
	Required     bool
}

func (a *Argument) Directory(c *cli.Context, d *dagger.Client) (*dagger.Directory, error) {
	value, err := a.ValueFunc(c, d)
	if err != nil {
		return nil, err
	}
	dir, ok := value.(*dagger.Directory)
	if !ok {
		return nil, errors.New("value returned by valuefunc is not a *dagger.Directory")
	}

	return dir, nil
}

func (a *Argument) MustDirectory(c *cli.Context, d *dagger.Client) *dagger.Directory {
	v, err := a.Directory(c, d)
	if err != nil {
		panic(err)
	}

	return v
}

func (a *Argument) String(c *cli.Context, d *dagger.Client) (string, error) {
	value, err := a.ValueFunc(c, d)
	if err != nil {
		return "", err
	}
	v, ok := value.(string)
	if !ok {
		return "", errors.New("value returned by valuefunc is not a string")
	}

	return v, nil
}

func (a *Argument) MustString(c *cli.Context, d *dagger.Client) string {
	v, err := a.String(c, d)
	if err != nil {
		panic(err)
	}

	return v
}

func (a *Argument) Int64(c *cli.Context, d *dagger.Client) (int64, error) {
	value, err := a.ValueFunc(c, d)
	if err != nil {
		return 0, err
	}
	v, ok := value.(int64)
	if !ok {
		return 0, errors.New("value returned by valuefunc is not an int64")
	}

	return v, nil
}

func (a *Argument) MustInt64(c *cli.Context, d *dagger.Client) int64 {
	v, err := a.Int64(c, d)
	if err != nil {
		panic(err)
	}

	return v
}

func (a *Argument) Bool(c *cli.Context, d *dagger.Client) (bool, error) {
	value, err := a.ValueFunc(c, d)
	if err != nil {
		return false, err
	}
	v, ok := value.(bool)
	if !ok {
		return false, errors.New("value returned by valuefunc is not a bool")
	}

	return v, nil
}

func (a *Argument) MustBool(c *cli.Context, d *dagger.Client) bool {
	v, err := a.Bool(c, d)
	if err != nil {
		panic(err)
	}

	return v
}

func (a *Argument) File(c *cli.Context, d *dagger.Client) (*dagger.File, error) {
	value, err := a.ValueFunc(c, d)
	if err != nil {
		return nil, err
	}
	dir, ok := value.(*dagger.File)
	if !ok {
		return nil, errors.New("value returned by valuefunc is not a *dagger.File")
	}

	return dir, nil
}

func (a *Argument) MustFile(c *cli.Context, d *dagger.Client) *dagger.File {
	v, err := a.File(c, d)
	if err != nil {
		panic(err)
	}

	return v
}
