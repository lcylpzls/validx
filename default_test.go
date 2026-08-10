package validx

import (
	"errors"
	"reflect"
	"testing"

	"github.com/lcylpzls/errx"
	"github.com/lcylpzls/testx"
)

func TestDefaultValidateField(t *testing.T) {
	testx.RequireNoError(t, ValidateField("abc@example.com", "email"))
	testx.RequireError(t, ValidateField("not-an-email", "email"))
}

func TestDefaultValidateStruct(t *testing.T) {
	type S struct {
		Name string `validate:"required,min=2"`
	}
	testx.RequireNoError(t, Validate(S{Name: "张三"}))
	err := Validate(S{Name: "a"})
	testx.RequireError(t, err)
	testx.RequireErrCode(t, err, CodeValidationFailed)
}

func TestDefaultValidateNil(t *testing.T) {
	testx.RequireErrCode(t, Validate(nil), CodeValidationFailed)
	var p *struct{}
	testx.RequireErrCode(t, Validate(p), CodeValidationFailed)
}

func TestRegisterRuleGlobal(t *testing.T) {
	err := RegisterRule("even_length", func(value any, param, path string) error {
		s, ok := value.(string)
		if !ok {
			return errors.New("仅支持字符串")
		}
		if len(s)%2 != 0 {
			return errx.NewCode(CodeValidationFailed, "长度必须为偶数")
		}
		return nil
	})
	testx.RequireNoError(t, err)
	testx.RequireNoError(t, ValidateField("abcd", "even_length"))
	testx.RequireError(t, ValidateField("abc", "even_length"))
	// 重复注册以最后一次为准。
	testx.RequireNoError(t, RegisterRule("even_length", func(value any, param, path string) error {
		return nil
	}))
	testx.RequireNoError(t, ValidateField("abc", "even_length"))
}

func TestRegisterRuleErrors(t *testing.T) {
	testx.RequireErrCode(t, RegisterRule("", func(any, string, string) error { return nil }), CodeInvalidRule)
	testx.RequireErrCode(t, RegisterRule("required", func(any, string, string) error { return nil }), CodeInvalidRule)
	testx.RequireErrCode(t, RegisterRule("ok_rule", nil), CodeInvalidRule)
}

func TestDefaultInstanceConsistent(t *testing.T) {
	if defaultValidator == nil {
		t.Fatal("默认校验器不能为 nil")
	}
	if err := RegisterRule("tmp_rule", func(any, string, string) error { return nil }); err != nil {
		t.Fatalf("注册失败：%v", err)
	}
	if _, ok := defaultValidator.customFn("tmp_rule"); !ok {
		t.Fatal("全局注册应写入默认实例")
	}
}

func TestValidateFieldRaw(t *testing.T) {
	testx.RequireNoError(t, ValidateFieldRaw(&struct{ A int }{1}, "required"))
	testx.RequireError(t, ValidateFieldRaw((*struct{ A int })(nil), "required"))
	testx.RequireNoError(t, ValidateFieldRaw("abc", "required"))
	testx.RequireErrCode(t, ValidateFieldRaw(1, "unknown_rule"), CodeInvalidRule)
	testx.RequireError(t, ValidateFieldRaw(nil, "required"))
	testx.RequireError(t, ValidateField(nil, "required"))

	// 自定义规则能拿到原始指针（不被 validx 解引用）。
	testx.RequireNoError(t, RegisterRule("raw_ptr", func(value any, param, path string) error {
		rv := reflect.ValueOf(value)
		if !rv.IsValid() || rv.Kind() != reflect.Ptr || rv.IsNil() {
			return errx.NewCode(CodeValidationFailed, "必须为非空指针")
		}
		return nil
	}))
	testx.RequireNoError(t, ValidateFieldRaw(&struct{ A int }{1}, "raw_ptr"))
	testx.RequireError(t, ValidateFieldRaw((*struct{ A int })(nil), "raw_ptr"))
	testx.RequireError(t, ValidateFieldRaw(42, "raw_ptr"))
}
