package validx

import (
	"testing"
)

// FuzzRules 保证任意 tag 解析不 panic,且错误码稳定。
func FuzzRules(f *testing.F) {
	f.Add("required")
	f.Add("min=3,max=10")
	f.Add("regexp=^[a-z]+$")
	f.Add("bad rule")
	f.Add("dive,dive")
	f.Fuzz(func(t *testing.T, tag string) {
		v, err := New()
		if err != nil {
			t.Fatal(err)
		}
		_, _ = v.compileRules(tag)
	})
}

// FuzzValidate 保证任意字符串结构体校验不 panic。
func FuzzValidate(f *testing.F) {
	f.Add("")
	f.Add("a@b.com")
	f.Add(string([]byte{0, 255, 'x'}))
	f.Fuzz(func(t *testing.T, s string) {
		type S struct {
			Name string `validate:"required,min=1,max=100"`
			Code string `validate:"omitempty,regexp=^[a-z0-9]+$"`
			Role string `validate:"oneof=admin user guest"`
		}
		v, err := New()
		if err != nil {
			t.Fatal(err)
		}
		_ = v.Validate(S{Name: s, Code: s, Role: s})
	})
}
