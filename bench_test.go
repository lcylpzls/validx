package validx

import (
	"testing"
)

// benchUser 是基准结构体。
type benchUser struct {
	Name  string   `validate:"required,min=2,max=32"`
	Email string   `validate:"required,email"`
	Age   int      `validate:"min=0,max=150"`
	Roles []string `validate:"dive,oneof=admin user guest"`
}

// BenchmarkValidate 基准:5 字段结构体验证。
func BenchmarkValidate(b *testing.B) {
	v, err := New()
	if err != nil {
		b.Fatal(err)
	}
	u := benchUser{
		Name:  "张三",
		Email: "user@example.com",
		Age:   30,
		Roles: []string{"admin", "user"},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := v.Validate(u); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkValidateInvalid 基准:含失败字段的校验。
func BenchmarkValidateInvalid(b *testing.B) {
	v, err := New()
	if err != nil {
		b.Fatal(err)
	}
	u := benchUser{Name: "x", Email: "bad", Age: -1}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = v.Validate(u)
	}
}

// BenchmarkCompileOnce 基准:首次编译(缓存未命中)。
func BenchmarkCompileOnce(b *testing.B) {
	v, err := New()
	if err != nil {
		b.Fatal(err)
	}
	u := benchUser{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = v.Validate(u)
	}
}
