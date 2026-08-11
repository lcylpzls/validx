package validx_test

import (
	"testing"

	"github.com/lcylpzls/validx"
)

// TestPublicAPI 黑盒冒烟测试：覆盖根包全部转发函数、类型别名与常量。
func TestPublicAPI(t *testing.T) {
	if validx.Version != "v1.3.1" {
		t.Fatalf("Version = %s", validx.Version)
	}

	type smokeStruct struct {
		Name string `validate:"required"`
	}
	if err := validx.Validate(smokeStruct{Name: "ok"}); err != nil {
		t.Fatalf("Validate 失败：%v", err)
	}
	if err := validx.ValidateField("abc", "required"); err != nil {
		t.Fatalf("ValidateField 失败：%v", err)
	}
	if err := validx.ValidateFieldRaw("abc", "required"); err != nil {
		t.Fatalf("ValidateFieldRaw 失败：%v", err)
	}
	if err := validx.RegisterRule("smoke_rule", func(value any, param, path string) error {
		return nil
	}); err != nil {
		t.Fatalf("RegisterRule 失败：%v", err)
	}

	v, err := validx.New(validx.WithTagName("validate"))
	if err != nil || v == nil {
		t.Fatalf("New 失败：%v", err)
	}

	var _ validx.Option
	var _ validx.Rule
	var _ validx.ValidationFunc
	_ = validx.CodeInvalidRule
	_ = validx.CodeValidationFailed
}
