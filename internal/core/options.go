package core

import "github.com/lcylpzls/errx"

// defaultTagName 是默认校验 tag 名。
const defaultTagName = "validate"

// config 是 Validator 的配置。
type config struct {
	tagName string
}

// defaultConfig 返回默认配置。
func defaultConfig() config {
	return config{tagName: defaultTagName}
}

// Option 修改 Validator 配置,在 New 时按顺序应用。
type Option func(*config)

// WithTagName 设置校验 tag 名(默认 validate),空串非法。
func WithTagName(name string) Option {
	return func(c *config) { c.tagName = name }
}

// validateConfig 校验配置参数。
func validateConfig(cfg config) error {
	if cfg.tagName == "" {
		return errx.NewCode(CodeInvalidRule, "tag 名不能为空")
	}
	return nil
}
