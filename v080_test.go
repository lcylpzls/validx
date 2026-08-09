package validx

import (
	"reflect"
	"testing"

	"github.com/lcylpzls/errx"
)

// ---------- excluded_if ----------

type BankForm struct {
	Method string `validate:"oneof=card wallet"`
	CardNo string `validate:"excluded_if=Method wallet"`
}

func TestExcludedIf(t *testing.T) {
	v := newTestValidator(t)
	// 条件成立且为空:通过
	if err := v.Validate(BankForm{Method: "wallet"}); err != nil {
		t.Fatalf("条件成立且为空应通过:%v", err)
	}
	// 条件成立但有值:失败
	err := v.Validate(BankForm{Method: "wallet", CardNo: "1234"})
	if err == nil {
		t.Fatal("条件成立但有值应失败")
	}
	if code, _ := errx.CodeOf(err); code != CodeValidationFailed {
		t.Errorf("错误码 = %s", code)
	}
	// 条件不成立有值:通过
	if err := v.Validate(BankForm{Method: "card", CardNo: "1234"}); err != nil {
		t.Fatalf("条件不成立有值应通过:%v", err)
	}
}

func TestExcludedIfCompileErrors(t *testing.T) {
	type BadFormat struct {
		A string `validate:"excluded_if=Method"`
	}
	type Missing struct {
		A string `validate:"excluded_if=NoSuch x"`
	}
	v := newTestValidator(t)
	for _, c := range []any{BadFormat{}, Missing{}} {
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

func TestExcludedIfInDiveRejected(t *testing.T) {
	type S struct {
		Items []string `validate:"dive,excluded_if=A x"`
	}
	v := newTestValidator(t)
	if err := v.Validate(S{Items: []string{"a"}}); err == nil {
		t.Fatal("dive 元素使用 excluded_if 应报错")
	}
}

func TestExcludedIfDirectDefense(t *testing.T) {
	v := newTestValidator(t)
	parent := reflect.ValueOf(struct{ A string }{A: "x"})
	if err := v.evalRule(Rule{name: "excluded_if", param: "A"},
		reflect.ValueOf(""), "F", parent); err == nil {
		t.Error("参数格式错误应报错")
	}
	if err := v.evalRule(Rule{name: "excluded_if", param: "NoSuch x"},
		reflect.ValueOf(""), "F", parent); err == nil {
		t.Error("引用不存在字段应报错")
	}
}

// ---------- 网络规则 ----------

func TestIPv4IPv6Rules(t *testing.T) {
	type S struct {
		V4 string `validate:"ipv4"`
		V6 string `validate:"ipv6"`
	}
	v := newTestValidator(t)
	ok := S{V4: "192.168.1.1", V6: "2001:db8::1"}
	if err := v.Validate(ok); err != nil {
		t.Fatalf("合法值应通过:%v", err)
	}
	// IPv6 过 ipv4 失败
	if err := v.Validate(S{V4: "::1", V6: "2001:db8::1"}); err == nil {
		t.Error("IPv6 不应通过 ipv4")
	}
	// IPv4 过 ipv6 失败
	if err := v.Validate(S{V4: "192.168.1.1", V6: "127.0.0.1"}); err == nil {
		t.Error("IPv4 不应通过 ipv6")
	}
	if err := v.Validate(S{V4: "999.1.1.1", V6: "2001:db8::1"}); err == nil {
		t.Error("非法 IPv4 应失败")
	}
}

func TestHostnameFqdnRules(t *testing.T) {
	type S struct {
		Host string `validate:"hostname"`
		FQDN string `validate:"fqdn"`
	}
	v := newTestValidator(t)
	ok := S{Host: "api-1", FQDN: "api.example.com"}
	if err := v.Validate(ok); err != nil {
		t.Fatalf("合法值应通过:%v", err)
	}
	cases := []S{
		{Host: "-api", FQDN: "api.example.com"},
		{Host: "api_1", FQDN: "api.example.com"},
		{Host: "api-1", FQDN: "localhost"},
		{Host: "api-1", FQDN: ".example.com"},
		{Host: "", FQDN: "api.example.com"},
	}
	for _, c := range cases {
		if err := v.Validate(c); err == nil {
			t.Errorf("非法值应失败:%+v", c)
		}
	}
}

func TestPortRule(t *testing.T) {
	type S struct {
		PortStr string `validate:"port"`
		PortInt int    `validate:"port"`
	}
	v := newTestValidator(t)
	for _, p := range []int{0, 80, 443, 65535} {
		if err := v.Validate(S{PortStr: "80", PortInt: p}); err != nil {
			t.Errorf("端口 %d 应通过:%v", p, err)
		}
	}
	if err := v.Validate(S{PortStr: "65536", PortInt: 80}); err == nil {
		t.Error("字符串端口超限应失败")
	}
	if err := v.Validate(S{PortStr: "abc", PortInt: 80}); err == nil {
		t.Error("非法字符串端口应失败")
	}
	if err := v.Validate(S{PortStr: "80", PortInt: -1}); err == nil {
		t.Error("负数端口应失败")
	}
	if err := v.Validate(S{PortStr: "80", PortInt: 65536}); err == nil {
		t.Error("整数端口超限应失败")
	}
}
