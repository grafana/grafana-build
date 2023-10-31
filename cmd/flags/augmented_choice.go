package flags

import (
	"flag"
	"fmt"
	"strings"

	"github.com/grafana/grafana-build/stringutil"
	"github.com/urfave/cli/v2"
)

type AugmentedChoiceValue struct {
	value         string
	agumentations []string
	choices       []string
}

func (c *AugmentedChoiceValue) String() string {
	return c.value
}

func (c *AugmentedChoiceValue) Set(val string) error {
	if !stringutil.ContainsPrefix(c.choices, val) {
		return fmt.Errorf("'%s' needs to be one of [ %s ] and with the optional augmentations", val, strings.Join(c.choices, ","))
	}
	opt := strings.Split(val, ":")

	c.value = opt[0]
	if len(opt) > 1 {
		c.agumentations = opt[1:]
	}

	return nil
}

// AugmentedChoiceFlag is a cli.Flag whose value is populated from preconfigured choices.
// Those choices are formatted as so:
// * {value}[:...augmentation]
// For example: "deb:amd64:static"
type AugmentedChoiceFlag struct {
	Name         string
	Usage        string
	Value        string
	choiceValue  *AugmentedChoiceValue
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

func (f *AugmentedChoiceFlag) String() string {
	return cli.FlagStringer(f)
}

func (f *AugmentedChoiceFlag) Apply(set *flag.FlagSet) error {
	if !stringutil.ContainsPrefix(f.Choices, f.Value) {
		return fmt.Errorf("'%s' needs to be one of %s", f.Value, strings.Join(f.Choices, ","))
	}

	f.choiceValue = &AugmentedChoiceValue{
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
func (f *AugmentedChoiceFlag) TakesValue() bool {
	return false
}

// Names returns the names of the flag
func (f *AugmentedChoiceFlag) Names() []string {
	return cli.FlagNames(f.Name, f.Aliases)
}

func (f *AugmentedChoiceFlag) IsSet() bool {
	return f.HasBeenSet
}

// IsRequired returns whether or not the flag is required
func (f *AugmentedChoiceFlag) IsRequired() bool {
	return f.Required
}

// IsVisible returns true if the flag is not hidden, otherwise false
func (f *AugmentedChoiceFlag) IsVisible() bool {
	return !f.Hidden // this function is required for help text to show
}

// GetDefaultText returns the default text for this flag
func (f *AugmentedChoiceFlag) GetDefaultText() string {
	if f.DefaultText != "" {
		return f.DefaultText
	}
	return fmt.Sprintf("%v", f.defaultValue)
}

// GetUsage returns the usage string for the flag
func (f *AugmentedChoiceFlag) GetUsage() string {
	return fmt.Sprintf("%s (choices: [%s])", f.Usage, strings.Join(f.Choices, ","))
}

// GetCategory returns the category for the flag
func (f *AugmentedChoiceFlag) GetCategory() string {
	return f.Category
}

// GetValue returns the flags value as string representation and an empty
// string if the flag takes no value at all.
func (f *AugmentedChoiceFlag) GetValue() string {
	return f.choiceValue.value
}

// GetEnvVars returns the env vars for this flag
func (f *AugmentedChoiceFlag) GetEnvVars() []string {
	return f.EnvVars
}
