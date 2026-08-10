package core

import (
	testx "github.com/lcylpzls/testx"
	"reflect"
	"testing"

	"github.com/lcylpzls/errx"
)

// ---------- 条件必填 ----------

type PaymentForm struct {
	Method string `validate:"oneof=card wallet"`
	CardNo string `validate:"required_if=Method card"`
	Wallet string `validate:"required_unless=Method wallet"`
}

func TestRequiredIf(t *testing.T) {
	v := newTestValidator(t)
	// 条件满足且已填:通过
	ok := PaymentForm{Method: "card", CardNo: "1234", Wallet: "w"}
	if err := v.Validate(ok); err != nil {
		t.Fatalf("合法值应通过:%v", err)
	}
	// 条件满足但未填:失败
	err := v.Validate(PaymentForm{Method: "card"})
	testx.RequireError(t, err)

	if code, _ := errx.CodeOf(err); code != CodeValidationFailed {
		t.Errorf("错误码 = %s", code)
	}
	// 条件不满足:不要求
	if err := v.Validate(PaymentForm{Method: "wallet"}); err != nil {
		t.Fatalf("条件不满足不应要求:%v", err)
	}
}

func TestRequiredUnless(t *testing.T) {
	v := newTestValidator(t)
	// Method != wallet 时 Wallet 必填
	if err := v.Validate(PaymentForm{Method: "card", CardNo: "x"}); err == nil {
		t.Fatal("required_unless 条件成立且为空应失败")
	}
	ok := PaymentForm{Method: "card", CardNo: "x", Wallet: "w"}
	if err := v.Validate(ok); err != nil {
		t.Fatalf("合法值应通过:%v", err)
	}
	// Method == wallet 时不要求 Wallet
	if err := v.Validate(PaymentForm{Method: "wallet"}); err != nil {
		t.Fatalf("条件不成立不应要求:%v", err)
	}
}

func TestRequiredIfCompileErrors(t *testing.T) {
	type BadFormat struct {
		A string `validate:"required_if=Method"` // 缺值
	}
	type Missing struct {
		A string `validate:"required_if=NoSuch x"`
	}
	type Unexported struct {
		A    string `validate:"required_if=priv x"`
		priv string
	}
	v := newTestValidator(t)
	for _, c := range []any{BadFormat{}, Missing{}, Unexported{A: "x", priv: "y"}} {
		err := v.Validate(c)
		if err == nil {
			t.Errorf("%T 应报规则错误", c)
			continue
		}
		if code, _ := errx.CodeOf(err); code != CodeInvalidRule {
			t.Errorf("%T 错误码 = %s", c, code)
		}
	}
}

func TestRequiredIfInDiveRejected(t *testing.T) {
	type S struct {
		Items []string `validate:"dive,required_if=A x"`
	}
	v := newTestValidator(t)
	err := v.Validate(S{Items: []string{"a"}})
	testx.RequireError(t, err)

	if code, _ := errx.CodeOf(err); code != CodeInvalidRule {
		t.Errorf("错误码 = %s", code)
	}
}

// ---------- 新格式规则 ----------

func TestBase64Rule(t *testing.T) {
	type S struct {
		V string `validate:"base64"`
	}
	v := newTestValidator(t)
	for _, ok := range []string{"", "aGVsbG8=", "YWJj"} {
		if err := v.Validate(S{V: ok}); err != nil {
			t.Errorf("%q 应通过:%v", ok, err)
		}
	}
	for _, bad := range []string{"!!!", "aGVsbG8", "abc==="} {
		if err := v.Validate(S{V: bad}); err == nil {
			t.Errorf("%q 应失败", bad)
		}
	}
}

func TestJSONRule(t *testing.T) {
	type S struct {
		V string `validate:"json"`
	}
	v := newTestValidator(t)
	for _, ok := range []string{`{}`, `{"a":1}`, `[1,2]`, `"str"`, `null`} {
		if err := v.Validate(S{V: ok}); err != nil {
			t.Errorf("%q 应通过:%v", ok, err)
		}
	}
	for _, bad := range []string{"", "{", "abc", `{"a":}`} {
		if err := v.Validate(S{V: bad}); err == nil {
			t.Errorf("%q 应失败", bad)
		}
	}
}

func TestHexadecimalRule(t *testing.T) {
	type S struct {
		V string `validate:"hexadecimal"`
	}
	v := newTestValidator(t)
	for _, ok := range []string{"0", "deadBEEF", "0123456789abcdefABCDEF"} {
		if err := v.Validate(S{V: ok}); err != nil {
			t.Errorf("%q 应通过:%v", ok, err)
		}
	}
	for _, bad := range []string{"", "0x1F", "ghij"} {
		if err := v.Validate(S{V: bad}); err == nil {
			t.Errorf("%q 应失败", bad)
		}
	}
}

func TestMACRule(t *testing.T) {
	type S struct {
		V string `validate:"mac"`
	}
	v := newTestValidator(t)
	for _, ok := range []string{"00:1A:2B:3C:4D:5E", "00-1A-2B-3C-4D-5E"} {
		if err := v.Validate(S{V: ok}); err != nil {
			t.Errorf("%q 应通过:%v", ok, err)
		}
	}
	for _, bad := range []string{"", "001A2B3C4D5E", "00:1A:2B:3C:4D"} {
		if err := v.Validate(S{V: bad}); err == nil {
			t.Errorf("%q 应失败", bad)
		}
	}
}

func TestSemverRule(t *testing.T) {
	type S struct {
		V string `validate:"semver"`
	}
	v := newTestValidator(t)
	for _, ok := range []string{"1.2.3", "v1.2.3", "1.2.3-alpha.1", "1.2.3+build.5"} {
		if err := v.Validate(S{V: ok}); err != nil {
			t.Errorf("%q 应通过:%v", ok, err)
		}
	}
	for _, bad := range []string{"", "1.2", "1.2.x", "v1.2.3.4"} {
		if err := v.Validate(S{V: bad}); err == nil {
			t.Errorf("%q 应失败", bad)
		}
	}
}

// ---------- 字符串包含规则 ----------

func TestStringContainRules(t *testing.T) {
	type S struct {
		A string `validate:"contains=abc"`
		B string `validate:"excludes=x"`
		C string `validate:"startswith=go"`
		D string `validate:"endswith=.go"`
	}
	v := newTestValidator(t)
	ok := S{A: "xabcy", B: "hello", C: "golang", D: "main.go"}
	if err := v.Validate(ok); err != nil {
		t.Fatalf("合法值应通过:%v", err)
	}
	cases := []S{
		{A: "xyz", B: "hello", C: "golang", D: "main.go"},
		{A: "xabcy", B: "x", C: "golang", D: "main.go"},
		{A: "xabcy", B: "hello", C: "ang", D: "main.go"},
		{A: "xabcy", B: "hello", C: "golang", D: "go"},
	}
	for _, c := range cases {
		if err := v.Validate(c); err == nil {
			t.Errorf("非法值应失败:%+v", c)
		}
	}
	type Wrong struct {
		A int `validate:"contains=x"`
	}
	if err := v.Validate(Wrong{A: 1}); err == nil {
		t.Error("contains 用于非字符串应失败")
	}
}

func TestRequiredIfDirectDefense(t *testing.T) {
	// 运行时防御分支:编译期已拦截,直接调用验证不 panic。
	v := newTestValidator(t)
	parent := reflect.ValueOf(struct{ A string }{A: "x"})
	if err := v.evalRule(Rule{name: "required_if", param: "A"},
		reflect.ValueOf(""), "F", parent); err == nil {
		t.Error("参数格式错误应报错")
	}
	if err := v.evalRule(Rule{name: "required_if", param: "NoSuch x"},
		reflect.ValueOf(""), "F", parent); err == nil {
		t.Error("引用不存在字段应报错")
	}
}
