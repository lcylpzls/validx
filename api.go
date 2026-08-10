package validx

import "github.com/lcylpzls/validx/internal/core"

const Version = core.Version

const (
	CodeInvalidRule      = core.CodeInvalidRule
	CodeValidationFailed = core.CodeValidationFailed
)

type (
	Option         = core.Option
	Rule           = core.Rule
	Validator      = core.Validator
	ValidationFunc = core.ValidationFunc
)

func Validate(value any) error                    { return core.Validate(value) }
func ValidateField(value any, rules string) error { return core.ValidateField(value, rules) }
func ValidateFieldRaw(value any, rules string) error {
	return core.ValidateFieldRaw(value, rules)
}
func RegisterRule(name string, fn ValidationFunc) error {
	return core.RegisterRule(name, fn)
}
func WithTagName(name string) Option { return core.WithTagName(name) }
func New(opts ...Option) (*Validator, error) {
	return core.New(opts...)
}
