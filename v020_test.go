package validx

import (
	"errors"
	"strings"
	"sync"
	"testing"

	"github.com/lcylpzls/errx"
)

// ---------- 新规则 ----------

func TestAlphaRules(t *testing.T) {
	type S struct {
		A string `validate:"alpha"`
		B string `validate:"alphanum"`
		C string `validate:"numeric"`
	}
	v := newTestValidator(t)
	ok := S{A: "Abc", B: "Abc123", C: "12345"}
	if err := v.Validate(ok); err != nil {
		t.Fatalf("合法值应通过:%v", err)
	}
	cases := []S{
		{A: "abc1", B: "Abc123", C: "12345"},
		{A: "Abc", B: "Abc-123", C: "12345"},
		{A: "Abc", B: "Abc123", C: "12a45"},
		{A: "", B: "", C: ""},
	}
	for _, c := range cases {
		if err := v.Validate(c); err == nil {
			t.Errorf("非法值应失败:%+v", c)
		}
	}
	type Wrong struct {
		A int `validate:"alpha"`
	}
	if err := v.Validate(Wrong{A: 1}); err == nil {
		t.Error("alpha 用于非字符串应失败")
	}
}

func TestBooleanRule(t *testing.T) {
	type S struct {
		B string `validate:"boolean"`
	}
	v := newTestValidator(t)
	for _, ok := range []string{"true", "false", "1", "0", "TRUE"} {
		if err := v.Validate(S{B: ok}); err != nil {
			t.Errorf("%q 应通过:%v", ok, err)
		}
	}
	for _, bad := range []string{"", "yes", "maybe"} {
		if err := v.Validate(S{B: bad}); err == nil {
			t.Errorf("%q 应失败", bad)
		}
	}
	type Wrong struct {
		B bool `validate:"boolean"`
	}
	if err := v.Validate(Wrong{B: true}); err == nil {
		t.Error("boolean 用于非字符串应失败")
	}
}

func TestUUIDRule(t *testing.T) {
	type S struct {
		ID string `validate:"uuid"`
	}
	v := newTestValidator(t)
	ok := "123e4567-e89b-12d3-a456-426614174000"
	if err := v.Validate(S{ID: ok}); err != nil {
		t.Fatalf("合法 UUID 应通过:%v", err)
	}
	for _, bad := range []string{"", "123e4567e89b12d3a456426614174000", "123e4567-e89b-12d3-a456-42661417400Z"} {
		if err := v.Validate(S{ID: bad}); err == nil {
			t.Errorf("%q 应失败", bad)
		}
	}
}

func TestURLRule(t *testing.T) {
	type S struct {
		URL string `validate:"url"`
	}
	v := newTestValidator(t)
	for _, ok := range []string{"https://example.com", "http://localhost:8080/path?q=1", "ftp://host"} {
		if err := v.Validate(S{URL: ok}); err != nil {
			t.Errorf("%q 应通过:%v", ok, err)
		}
	}
	for _, bad := range []string{"", "example.com", "/relative", "http://", "://x"} {
		if err := v.Validate(S{URL: bad}); err == nil {
			t.Errorf("%q 应失败", bad)
		}
	}
}

func TestIPRule(t *testing.T) {
	type S struct {
		IP string `validate:"ip"`
	}
	v := newTestValidator(t)
	for _, ok := range []string{"127.0.0.1", "::1", "2001:db8::1"} {
		if err := v.Validate(S{IP: ok}); err != nil {
			t.Errorf("%q 应通过:%v", ok, err)
		}
	}
	for _, bad := range []string{"", "999.1.1.1", "not-an-ip"} {
		if err := v.Validate(S{IP: bad}); err == nil {
			t.Errorf("%q 应失败", bad)
		}
	}
}

func TestDatetimeRule(t *testing.T) {
	type S struct {
		Date string `validate:"datetime=2006-01-02"`
	}
	v := newTestValidator(t)
	if err := v.Validate(S{Date: "2026-08-09"}); err != nil {
		t.Fatalf("合法时间应通过:%v", err)
	}
	for _, bad := range []string{"", "2026/08/09", "2026-13-01"} {
		if err := v.Validate(S{Date: bad}); err == nil {
			t.Errorf("%q 应失败", bad)
		}
	}
	type Wrong struct {
		Date int `validate:"datetime=2006-01-02"`
	}
	if err := v.Validate(Wrong{Date: 1}); err == nil {
		t.Error("datetime 用于非字符串应失败")
	}
}

func TestCompareRules(t *testing.T) {
	type S struct {
		Gt  int    `validate:"gt=10"`
		Lt  int    `validate:"lt=20"`
		Gte int    `validate:"gte=10"`
		Lte int    `validate:"lte=20"`
		Str string `validate:"gte=3"` // 字符串长度比较
	}
	v := newTestValidator(t)
	if err := v.Validate(S{Gt: 11, Lt: 19, Gte: 10, Lte: 20, Str: "abc"}); err != nil {
		t.Fatalf("合法值应通过:%v", err)
	}
	cases := []S{
		{Gt: 10, Lt: 19, Gte: 10, Lte: 20, Str: "abc"},
		{Gt: 11, Lt: 20, Gte: 10, Lte: 20, Str: "abc"},
		{Gt: 11, Lt: 19, Gte: 9, Lte: 20, Str: "abc"},
		{Gt: 11, Lt: 19, Gte: 10, Lte: 21, Str: "abc"},
		{Gt: 11, Lt: 19, Gte: 10, Lte: 20, Str: "ab"},
	}
	for _, c := range cases {
		if err := v.Validate(c); err == nil {
			t.Errorf("非法值应失败:%+v", c)
		}
	}
	type Bad struct {
		A int `validate:"gt=abc"`
	}
	if err := v.Validate(Bad{}); err == nil {
		t.Fatal("非法参数应报错")
	} else if code, _ := errx.CodeOf(err); code != CodeInvalidRule {
		t.Errorf("错误码 = %s,want %s", code, CodeInvalidRule)
	}
	type Wrong struct {
		B bool `validate:"gt=1"`
	}
	if err := v.Validate(Wrong{B: true}); err == nil {
		t.Error("gt 用于不适用类型应失败")
	}
}

// ---------- 自定义规则 ----------

func TestRegisterValidation(t *testing.T) {
	v := newTestValidator(t)
	if err := v.RegisterValidation("even", func(value any, _ string, path string) error {
		n, ok := value.(int)
		if !ok || n%2 != 0 {
			return errx.NewCodef(CodeValidationFailed, "必须为偶数").
				WithField("field", path).WithField("rule", "even")
		}
		return nil
	}); err != nil {
		t.Fatalf("注册失败:%v", err)
	}
	type S struct {
		N int `validate:"even"`
	}
	if err := v.Validate(S{N: 4}); err != nil {
		t.Fatalf("自定义规则应通过:%v", err)
	}
	err := v.Validate(S{N: 3})
	if err == nil {
		t.Fatal("自定义规则应失败")
	}
	e, ok := errx.As(err)
	if !ok {
		t.Fatalf("应为结构化错误:%v", err)
	}
	var hasField, hasRule bool
	for _, f := range e.Fields() {
		if f.Key == "field" && f.Value == "N" {
			hasField = true
		}
		if f.Key == "rule" && f.Value == "even" {
			hasRule = true
		}
	}
	if !hasField || !hasRule {
		t.Errorf("自定义规则错误字段不符:%v", e.Fields())
	}
}

func TestRegisterValidationWithParam(t *testing.T) {
	v := newTestValidator(t)
	if err := v.RegisterValidation("prefix", func(value any, param string, path string) error {
		s, ok := value.(string)
		if !ok || !strings.HasPrefix(s, param) {
			return errors.New("前缀不匹配")
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	type S struct {
		Code string `validate:"prefix=ABC-"`
	}
	if err := v.Validate(S{Code: "ABC-123"}); err != nil {
		t.Fatalf("带参自定义规则应通过:%v", err)
	}
	err := v.Validate(S{Code: "XYZ-123"})
	if err == nil {
		t.Fatal("带参自定义规则应失败")
	}
	if code, _ := errx.CodeOf(err); code != CodeValidationFailed {
		t.Errorf("普通错误应包装为校验失败:%s", code)
	}
}

func TestRegisterValidationErrors(t *testing.T) {
	v := newTestValidator(t)
	fn := func(any, string, string) error { return nil }
	if err := v.RegisterValidation("bad name", fn); err == nil {
		t.Error("非法规则名应报错")
	}
	if err := v.RegisterValidation("required", fn); err == nil {
		t.Error("内置规则冲突应报错")
	}
	if err := v.RegisterValidation("myrule", nil); err == nil {
		t.Error("nil 函数应报错")
	}
	// 覆盖注册
	if err := v.RegisterValidation("myrule", fn); err != nil {
		t.Fatal(err)
	}
	if err := v.RegisterValidation("myrule", func(any, string, string) error {
		return errors.New("新版")
	}); err != nil {
		t.Fatal(err)
	}
}

func TestCustomRuleConcurrent(t *testing.T) {
	v := newTestValidator(t)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 200; i++ {
			_ = v.RegisterValidation("r"+string(rune('a'+i%10)), func(any, string, string) error {
				return nil
			})
		}
	}()
	for i := 0; i < 200; i++ {
		_ = v.Validate(struct {
			A string `validate:"r0"`
		}{})
	}
	wg.Wait()
}

// ---------- 单字段校验 ----------

func TestValidateField(t *testing.T) {
	v := newTestValidator(t)
	if err := v.ValidateField("user@example.com", "required,email"); err != nil {
		t.Fatalf("合法单字段应通过:%v", err)
	}
	err := v.ValidateField("", "required,email")
	if err == nil {
		t.Fatal("空值应失败")
	}
	if err := v.ValidateField(nil, "required"); err == nil {
		t.Error("nil 应触发 required 失败")
	}
	var nilPtr *string
	if err := v.ValidateField(nilPtr, "required"); err == nil {
		t.Error("nil 指针应触发 required 失败")
	}
	if err := v.ValidateField("abc", "min=3"); err != nil {
		t.Fatalf("指针值应解引用:%v", err)
	}
	s := "abcd"
	if err := v.ValidateField(&s, "min=3"); err != nil {
		t.Fatalf("非 nil 指针应解引用:%v", err)
	}
	if err := v.ValidateField(42, "min=10"); err != nil {
		t.Fatalf("数值单字段应通过:%v", err)
	}
	// 失败错误路径为 value
	err = v.ValidateField(1, "min=10")
	if e, ok := errx.As(err); ok {
		for _, f := range e.Fields() {
			if f.Key == "field" && f.Value != "value" {
				t.Errorf("单字段路径应为 value:%v", f)
			}
		}
	}
	// 非法规则
	if err := v.ValidateField("x", "nosuch"); err == nil {
		t.Error("未知规则应报错")
	}
	// dive 拒绝
	if err := v.ValidateField([]int{1}, "dive"); err == nil {
		t.Error("单字段不支持 dive")
	}
}

func TestValidateFieldCache(t *testing.T) {
	v := newTestValidator(t)
	for i := 0; i < 3; i++ {
		if err := v.ValidateField("abc", "required,min=3"); err != nil {
			t.Fatal(err)
		}
	}
	if _, ok := v.fieldCache.Load("required,min=3"); !ok {
		t.Error("单字段规则应缓存")
	}
}
