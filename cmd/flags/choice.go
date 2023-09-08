package flags

import (
	"flag"
	"fmt"
	"strings"

	"github.com/grafana/grafana-build/stringutil"
	"github.com/urfave/cli/v2"
)

type ChoiceValue struct {
	value   string
	choices []string
}

func (c *ChoiceValue) String() string {
	return c.value
}

func (c *ChoiceValue) Set(val string) error {
	if !stringutil.Contains(c.choices, val) {
		return fmt.Errorf("'%s' needs to be one of %s", val, strings.Join(c.choices, ","))
	}

	c.value = val

	return nil
}

// ChoiceFlag is a cli.Flag whose value is populated from preconfigured choices
type ChoiceFlag struct {
	Name         string
	Usage        string
	Value        string
	choiceValue  *ChoiceValue
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
	if !stringutil.Contains(f.Choices, f.Value) {
		return fmt.Errorf("'%s' needs to be one of %s", f.Value, strings.Join(f.Choices, ","))
	}

	f.choiceValue = &ChoiceValue{
		// set default value so that flagset wont be able to overwrite it
		value:   f.Value,
		choices: f.Choices,
	}

	for _, name := range f.Names() {
		set.Var(f.choiceValue, name, f.Usage)
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
	return f.choiceValue.value
}

// GetEnvVars returns the env vars for this flag
func (f *ChoiceFlag) GetEnvVars() []string {
	return f.EnvVars
}
