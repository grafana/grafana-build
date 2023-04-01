package main

import (
	"flag"
	"fmt"
	"strings"

	"github.com/urfave/cli/v2"
)

// ChoiceFlag is a cli.Flag whose value is populated from preconfigured choices
type ChoiceFlag struct {
	Name         string
	Usage        string
	Value        string
	HasBeenSet   bool
	Choices      []string
	Aliases      []string
	defaultValue string
	Category     string
	DefaultText  string
	Required     bool
	Hidden       bool
	EnvVars      []string
}

func (f *ChoiceFlag) String() string {
	return cli.FlagStringer(f)
}

func (f *ChoiceFlag) Apply(set *flag.FlagSet) error {
	// set default value so that environment wont be able to overwrite it
	f.defaultValue = f.Value

	for _, name := range f.Names() {
		set.String(name, f.Value, f.Usage)
	}

	return nil
}

// TakesValue returns true of the flag takes a value, otherwise false
func (f *ChoiceFlag) TakesValue() bool {
	return false
}

// Names returns the names of the flag
func (f *ChoiceFlag) Names() []string {
	return cli.FlagNames(f.Name, f.Aliases)
}

func (f *ChoiceFlag) IsSet() bool {
	return f.HasBeenSet
}

// IsRequired returns whether or not the flag is required
func (f *ChoiceFlag) IsRequired() bool {
	return f.Required
}

// IsVisible returns true if the flag is not hidden, otherwise false
func (f *ChoiceFlag) IsVisible() bool {
	return !f.Hidden // this function is required for help text to show
}

// GetDefaultText returns the default text for this flag
func (f *ChoiceFlag) GetDefaultText() string {
	if f.DefaultText != "" {
		return f.DefaultText
	}
	return fmt.Sprintf("%v", f.defaultValue)
}

// GetUsage returns the usage string for the flag
func (f *ChoiceFlag) GetUsage() string {
	return fmt.Sprintf("%s (choices: [%s])", f.Usage, strings.Join(f.Choices, ","))
}

// GetCategory returns the category for the flag
func (f *ChoiceFlag) GetCategory() string {
	return f.Category
}

// GetValue returns the flags value as string representation and an empty
// string if the flag takes no value at all.
func (f *ChoiceFlag) GetValue() string {
	return f.Value
}

// GetEnvVars returns the env vars for this flag
func (f *ChoiceFlag) GetEnvVars() []string {
	return f.EnvVars
}
