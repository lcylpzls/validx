package core

import (
	testx "github.com/lcylpzls/testx"
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
		testx.RequireNoError(t, err)

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
		testx.RequireNoError(t, err)

		_ = v.Validate(S{Name: s, Code: s, Role: s})
	})
}

// FuzzValidateBatch 保证切片/map 批量校验路径不 panic。
func FuzzValidateBatch(f *testing.F) {
	f.Add("", int64(0))
	f.Add("上海", int64(1))
	f.Add("x", int64(2))
	f.Fuzz(func(t *testing.T, city string, count int64) {
		type Item struct {
			City  string `validate:"required,min=1"`
			Count int64  `validate:"gte=0,lte=100"`
		}
		v, err := New()
		testx.RequireNoError(t, err)

		n := count % 4
		if n < 0 {
			n = -n
		}
		items := make([]Item, 0, n)
		for i := int64(0); i < n; i++ {
			items = append(items, Item{City: city, Count: count})
		}
		_ = v.Validate(items)
		_ = v.Validate(map[string]Item{"k": {City: city, Count: count}})
	})
}
