package validx

import (
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/lcylpzls/errx"
)

func newTestValidator(t *testing.T, opts ...Option) *Validator {
	t.Helper()
	v, err := New(opts...)
	if err != nil {
		t.Fatalf("New 失败:%v", err)
	}
	return v
}

// ---------- 配置 ----------

func TestOptions(t *testing.T) {
	if _, err := New(WithTagName("")); err == nil {
		t.Error("空 tag 名应非法")
	}
	v := newTestValidator(t, WithTagName("check"))
	if v.cfg.tagName != "check" {
		t.Error("自定义 tag 名未生效")
	}
	_ = newTestValidator(t, nil)
}

type customTag struct {
	Name string `check:"required"`
}

func TestWithTagName(t *testing.T) {
	v := newTestValidator(t, WithTagName("check"))
	if err := v.Validate(customTag{}); err == nil {
		t.Error("自定义 tag 应生效")
	}
	// 默认 tag 名不识别 check
	v2 := newTestValidator(t)
	if err := v2.Validate(customTag{}); err != nil {
		t.Errorf("默认 tag 不应识别 check:%v", err)
	}
}

// ---------- required / omitempty ----------

func TestRequired(t *testing.T) {
	v := newTestValidator(t)
	type Req struct {
		Str   string         `validate:"required"`
		Int   int            `validate:"required"`
		Bool  bool           `validate:"required"`
		Ptr   *int           `validate:"required"`
		Slice []int          `validate:"required"`
		Map   map[string]int `validate:"required"`
	}
	err := v.Validate(Req{})
	if err == nil {
		t.Fatal("全零值应校验失败")
	}
	if code, _ := errx.CodeOf(err); code != CodeValidationFailed {
		t.Errorf("错误码 = %s,want %s", code, CodeValidationFailed)
	}
	// 非零值通过
	n := 1
	ok := Req{Str: "a", Int: 1, Bool: true, Ptr: &n, Slice: []int{1}, Map: map[string]int{"a": 1}}
	if err := v.Validate(ok); err != nil {
		t.Fatalf("非零值应通过:%v", err)
	}
}

func TestOmitEmpty(t *testing.T) {
	type S struct {
		Name string `validate:"omitempty,min=3"`
		Age  int    `validate:"omitempty,min=18"`
	}
	v := newTestValidator(t)
	if err := v.Validate(S{}); err != nil {
		t.Fatalf("空值应跳过后续规则:%v", err)
	}
	if err := v.Validate(S{Name: "ab", Age: 10}); err == nil {
		t.Fatal("非空值应继续校验")
	}
	if err := v.Validate(S{Name: "abc", Age: 20}); err != nil {
		t.Fatalf("合法值应通过:%v", err)
	}
}

func TestRequiredExtraKinds(t *testing.T) {
	type S struct {
		U  uint     `validate:"required"`
		F  float64  `validate:"required"`
		Ch chan int `validate:"required"`
		I  any      `validate:"required"`
		Fn func()   `validate:"required"`
	}
	v := newTestValidator(t)
	if err := v.Validate(S{}); err == nil {
		t.Fatal("零值 uint/float/chan/interface/func 应失败")
	}
	ok := S{U: 1, F: 0.5, Ch: make(chan int), I: "x", Fn: func() {}}
	if err := v.Validate(ok); err != nil {
		t.Fatalf("非零值应通过:%v", err)
	}
}

// ---------- min / max / len ----------

func TestLengthRules(t *testing.T) {
	type S struct {
		Int   int            `validate:"min=5,max=10"`
		Str   string         `validate:"len=3"`
		Slice []int          `validate:"min=2"`
		Map   map[string]int `validate:"max=2"`
		Runes string         `validate:"min=2"`
	}
	v := newTestValidator(t)
	ok := S{Int: 7, Str: "abc", Slice: []int{1, 2}, Map: map[string]int{"a": 1}, Runes: "你好"}
	if err := v.Validate(ok); err != nil {
		t.Fatalf("合法值应通过:%v", err)
	}
	bad := S{Int: 3, Str: "ab", Slice: []int{1}, Map: map[string]int{"a": 1, "b": 2, "c": 3}}
	if err := v.Validate(bad); err == nil {
		t.Fatal("非法值应失败")
	}
	// rune 计数:中文算一个字符
	if err := v.Validate(S{Int: 5, Str: "你好啊", Slice: []int{1, 2}, Map: map[string]int{}, Runes: "你好"}); err != nil {
		t.Fatalf("rune 长度应通过:%v", err)
	}
}

func TestLengthRulesMoreKinds(t *testing.T) {
	type S struct {
		F   float64 `validate:"min=1"`
		U   uint    `validate:"min=3"`
		Arr [3]int  `validate:"len=3"`
	}
	v := newTestValidator(t)
	if err := v.Validate(S{F: 2.9, U: 4, Arr: [3]int{1, 2, 3}}); err != nil {
		t.Fatalf("合法值应通过:%v", err)
	}
	if err := v.Validate(S{F: 0.5, U: 1, Arr: [3]int{}}); err == nil {
		t.Fatal("非法值应失败")
	}
}

func TestLengthRulesInvalidParam(t *testing.T) {
	type S struct {
		A int `validate:"min=abc"`
	}
	v := newTestValidator(t)
	err := v.Validate(S{})
	if err == nil {
		t.Fatal("非法参数应报错")
	}
	if code, _ := errx.CodeOf(err); code != CodeInvalidRule {
		t.Errorf("错误码 = %s,want %s", code, CodeInvalidRule)
	}
}

func TestLengthRulesUnsupportedType(t *testing.T) {
	type S struct {
		B bool `validate:"min=1"`
	}
	v := newTestValidator(t)
	if err := v.Validate(S{B: true}); err == nil {
		t.Fatal("不适用类型应报错")
	} else if code, _ := errx.CodeOf(err); code != CodeValidationFailed {
		t.Errorf("错误码 = %s,want %s", code, CodeValidationFailed)
	}
}

func TestUintSaturation(t *testing.T) {
	type S struct {
		N uint64 `validate:"min=10"`
	}
	v := newTestValidator(t)
	if err := v.Validate(S{N: ^uint64(0)}); err != nil {
		t.Fatalf("超大 uint 应通过 min:%v", err)
	}
}

// ---------- email / regexp / oneof ----------

func TestEmail(t *testing.T) {
	type S struct {
		Email string `validate:"email"`
	}
	v := newTestValidator(t)
	for _, ok := range []string{"a@b.com", "user.name+tag@example.co.uk"} {
		if err := v.Validate(S{Email: ok}); err != nil {
			t.Errorf("%q 应通过:%v", ok, err)
		}
	}
	for _, bad := range []string{"", "not-an-email", "a@", "@b.com"} {
		if err := v.Validate(S{Email: bad}); err == nil {
			t.Errorf("%q 应失败", bad)
		}
	}
	type Wrong struct {
		Email int `validate:"email"`
	}
	if err := v.Validate(Wrong{Email: 1}); err == nil {
		t.Error("email 用于非字符串应失败")
	}
}

func TestRegexp(t *testing.T) {
	type S struct {
		Code string `validate:"regexp=^[A-Z]{2}[0-9]{3}$"`
	}
	v := newTestValidator(t)
	if err := v.Validate(S{Code: "AB123"}); err != nil {
		t.Fatalf("合法正则应通过:%v", err)
	}
	if err := v.Validate(S{Code: "ab123"}); err == nil {
		t.Error("不匹配应失败")
	}
	type Bad struct {
		Code string `validate:"regexp=[unclosed"`
	}
	if err := v.Validate(Bad{}); err == nil {
		t.Fatal("非法正则应报错")
	} else if code, _ := errx.CodeOf(err); code != CodeInvalidRule {
		t.Errorf("错误码 = %s,want %s", code, CodeInvalidRule)
	}
	type Wrong struct {
		Code int `validate:"regexp=^x$"`
	}
	if err := v.Validate(Wrong{Code: 1}); err == nil {
		t.Error("regexp 用于非字符串应失败")
	}
}

func TestOneOf(t *testing.T) {
	type S struct {
		Role  string `validate:"oneof=admin user guest"`
		Level int    `validate:"oneof=1 2 3"`
	}
	v := newTestValidator(t)
	if err := v.Validate(S{Role: "admin", Level: 2}); err != nil {
		t.Fatalf("合法枚举应通过:%v", err)
	}
	if err := v.Validate(S{Role: "root", Level: 2}); err == nil {
		t.Error("非法枚举应失败")
	}
	if err := v.Validate(S{Role: "admin", Level: 9}); err == nil {
		t.Error("数字枚举应失败")
	}
	type More struct {
		U  uint    `validate:"oneof=1 2"`
		F  float64 `validate:"oneof=1.5 2.5"`
		B  bool    `validate:"oneof=true false"`
		St any     `validate:"oneof=hello"`
	}
	v2 := newTestValidator(t)
	if err := v2.Validate(More{U: 2, F: 2.5, B: false, St: "hello"}); err != nil {
		t.Fatalf("多类型枚举应通过:%v", err)
	}
	if err := v2.Validate(More{U: 3, F: 2.5, B: false, St: "hello"}); err == nil {
		t.Error("uint 枚举应失败")
	}
}

// ---------- 规则语法 ----------

func TestInvalidRules(t *testing.T) {
	type Unknown struct {
		A string `validate:"nosuchrule"`
	}
	type BadName struct {
		A string `validate:"bad rule"`
	}
	type MissingParam struct {
		A string `validate:"min"`
	}
	type ExtraParam struct {
		A string `validate:"required=1"`
	}
	type DoubleDive struct {
		A []string `validate:"dive,dive"`
	}
	type EmptyName struct {
		A string `validate:"=1"`
	}
	v := newTestValidator(t)
	cases := []any{
		Unknown{}, BadName{}, MissingParam{}, ExtraParam{}, DoubleDive{}, EmptyName{},
	}
	for _, c := range cases {
		err := v.Validate(c)
		if err == nil {
			t.Errorf("%T 应返回规则错误", c)
			continue
		}
		if code, _ := errx.CodeOf(err); code != CodeInvalidRule {
			t.Errorf("%T 错误码 = %s,want %s", c, code, CodeInvalidRule)
		}
	}
}

func TestEvalRuleUnknownDirect(t *testing.T) {
	// evalRule 防御分支:未知规则(正常路径已被编译层拦截)。
	v := newTestValidator(t)
	err := v.evalRule(Rule{name: "nosuchrule"}, reflect.ValueOf("x"), "F")
	if err == nil {
		t.Fatal("直接调用未知规则应报错")
	}
}

func TestEmptyRuleSegmentIgnored(t *testing.T) {
	type S struct {
		A string `validate:"required,,min=2"`
	}
	v := newTestValidator(t)
	if err := v.Validate(S{A: "ab"}); err != nil {
		t.Fatalf("空规则段应忽略:%v", err)
	}
	if err := v.Validate(S{}); err == nil {
		t.Error("非空规则仍应生效")
	}
}

// ---------- 嵌套与 dive ----------

type Address struct {
	City    string `validate:"required"`
	ZipCode string `validate:"len=6"`
}

type Profile struct {
	Address Address `validate:"required"`
	Memo    *Address
}

func TestNestedStruct(t *testing.T) {
	v := newTestValidator(t)
	okProfile := Profile{Address: Address{City: "上海", ZipCode: "200000"}}
	if err := v.Validate(okProfile); err != nil {
		t.Fatalf("合法嵌套应通过:%v", err)
	}
	err := v.Validate(Profile{Address: Address{City: "上海"}})
	if err == nil {
		t.Fatal("嵌套字段失败应报错")
	}
	e, isErr := errx.As(err)
	if !isErr {
		t.Fatalf("应为结构化错误:%v", err)
	}
	hasPath := false
	for _, f := range e.Fields() {
		if f.Key == "field" && f.Value == "Address.ZipCode" {
			hasPath = true
		}
	}
	if !hasPath {
		t.Errorf("嵌套路径不符:%v", e.Fields())
	}
	// 指针嵌套
	memo := &Address{City: "北京", ZipCode: "100000"}
	if err := v.Validate(Profile{Address: okProfile.Address, Memo: memo}); err != nil {
		t.Fatalf("指针嵌套应通过:%v", err)
	}
	// 空指针嵌套:无 required 时跳过
	if err := v.Validate(Profile{Address: okProfile.Address}); err != nil {
		t.Fatalf("空指针应跳过:%v", err)
	}
}

func TestDive(t *testing.T) {
	type Item struct {
		Name string `validate:"required"`
	}
	type S struct {
		Nums  []int          `validate:"dive,min=1"`
		Names []string       `validate:"dive,oneof=admin user"`
		Meta  map[string]int `validate:"dive,min=5"`
		Items []Item         `validate:"dive"`
		Plain []int          `validate:"min=1"` // 无 dive:容器自身长度
	}
	v := newTestValidator(t)
	ok := S{
		Nums:  []int{1, 2},
		Names: []string{"admin", "user"},
		Meta:  map[string]int{"a": 5},
		Items: []Item{{Name: "x"}},
		Plain: []int{1},
	}
	if err := v.Validate(ok); err != nil {
		t.Fatalf("合法 dive 应通过:%v", err)
	}
	// 元素规则失败
	if err := v.Validate(S{Nums: []int{0, 2}, Plain: []int{1}}); err == nil {
		t.Error("元素 min 应失败")
	}
	// map 值规则失败,路径含 key
	err := v.Validate(S{Meta: map[string]int{"a": 1}, Plain: []int{1}})
	if err == nil {
		t.Fatal("map 值应失败")
	}
	// 元素结构体递归
	err = v.Validate(S{Items: []Item{{}}, Plain: []int{1}})
	if err == nil {
		t.Fatal("元素结构体应递归")
	}
	var hasIdx bool
	if e, ok := errx.As(err); ok {
		for _, f := range e.Fields() {
			if f.Key == "field" && strings.HasPrefix(f.Value.(string), "Items[") {
				hasIdx = true
			}
		}
	}
	if !hasIdx {
		t.Errorf("元素路径应含索引:%v", err)
	}
	// 指针元素:nil 跳过、非 nil 递归
	type P struct {
		Items []*Item `validate:"dive"`
	}
	if err := v.Validate(P{Items: []*Item{nil, &Item{Name: "ok"}}}); err != nil {
		t.Fatalf("指针元素应通过:%v", err)
	}
}

func TestDiveOnNonContainer(t *testing.T) {
	type S struct {
		A int `validate:"dive"`
	}
	v := newTestValidator(t)
	if err := v.Validate(S{A: 1}); err == nil {
		t.Fatal("dive 用于非容器应报错")
	}
}

// ---------- 跳过与未导出 ----------

type Skip struct {
	Name string `validate:"required"`
	Skip string `validate:"-"`
	priv string `validate:"required"`
}

func TestSkipAndUnexported(t *testing.T) {
	v := newTestValidator(t)
	err := v.Validate(Skip{Name: "ok", Skip: "", priv: ""})
	if err != nil {
		t.Fatalf("跳过字段不应参与校验:%v", err)
	}
	if err := v.Validate(Skip{}); err == nil {
		t.Error("未跳过字段仍应校验")
	}
}

// ---------- 聚合错误 ----------

func TestAggregateErrors(t *testing.T) {
	type S struct {
		A string `validate:"required"`
		B string `validate:"min=3"`
		C string `validate:"email"`
	}
	v := newTestValidator(t)
	err := v.Validate(S{})
	if err == nil {
		t.Fatal("多字段失败应报错")
	}
	// 聚合:通过 errx.Is 检查任一字段错误码,并验证字段路径存在
	if !errx.Is(err, CodeValidationFailed) {
		t.Error("聚合应包含校验失败码")
	}
	agg, ok := err.(*errx.Aggregate)
	if !ok {
		t.Fatalf("多字段失败应为聚合错误:%T", err)
	}
	var gotA, gotB, gotC bool
	for _, e := range agg.Errors() {
		if ee, ok := errx.As(e); ok {
			for _, f := range ee.Fields() {
				switch f.Key {
				case "field":
					switch f.Value {
					case "A":
						gotA = true
					case "B":
						gotB = true
					case "C":
						gotC = true
					}
				}
			}
		}
	}
	if !gotA || !gotB || !gotC {
		t.Errorf("聚合字段不完整:A=%v B=%v C=%v", gotA, gotB, gotC)
	}
}

// ---------- Validate 入口 ----------

func TestValidateNonStruct(t *testing.T) {
	v := newTestValidator(t)
	for _, val := range []any{nil, (*Profile)(nil), "string", 42, []int{1}} {
		if err := v.Validate(val); err == nil {
			t.Errorf("%T 应报错", val)
		}
	}
	// 非 nil 指针应解引用通过
	okProfile := &Profile{Address: Address{City: "x", ZipCode: "123456"}}
	if err := v.Validate(okProfile); err != nil {
		t.Fatalf("指针应解引用通过:%v", err)
	}
}

// ---------- 并发 ----------

func TestConcurrentValidate(t *testing.T) {
	v := newTestValidator(t)
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 200; j++ {
				_ = v.Validate(Profile{Address: Address{City: "x", ZipCode: "123456"}})
				_ = v.Validate(Profile{})
			}
		}()
	}
	wg.Wait()
}

func TestErrorCodesRegistered(t *testing.T) {
	if errx.Describe(CodeInvalidRule) == "" || errx.Describe(CodeValidationFailed) == "" {
		t.Error("错误码未注册")
	}
}
