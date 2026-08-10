package core

import (
	testx "github.com/lcylpzls/testx"
	"strings"
	"testing"

	"github.com/lcylpzls/errx"
)

func TestVersion(t *testing.T) {
	testx.Equal(t, Version, "v1.3.0")

}

// ---------- dive 元素 nil 指针 ----------

func TestDiveNilPointerRequired(t *testing.T) {
	type Item struct {
		Name string `validate:"required"`
	}
	type S struct {
		Items []*Item `validate:"dive,required"`
	}
	v := newTestValidator(t)
	if err := v.Validate(S{Items: []*Item{&Item{Name: "ok"}}}); err != nil {
		t.Fatalf("合法元素应通过:%v", err)
	}
	err := v.Validate(S{Items: []*Item{nil}})
	testx.RequireError(t, err)

	if code, _ := errx.CodeOf(err); code != CodeValidationFailed {
		t.Errorf("错误码 = %s,want %s", code, CodeValidationFailed)
	}
}

// ---------- float 比较正确性 ----------

func TestFloatComparisonCorrectness(t *testing.T) {
	type S struct {
		Max float64 `validate:"max=1"`
		Min float64 `validate:"min=1.5"`
		Gt  float64 `validate:"gt=1.5"`
		Lt  float64 `validate:"lt=2.5"`
		Gte float64 `validate:"gte=1.5"`
		Lte float64 `validate:"lte=2.5"`
	}
	v := newTestValidator(t)
	// 1.9 必须被 max=1 拦截(修复前截断为 1 会误通过)。
	if err := v.Validate(S{Max: 1.9, Min: 2, Gt: 2, Lt: 2, Gte: 2, Lte: 2}); err == nil {
		t.Fatal("1.9 应被 max=1 拦截")
	}
	// 小数参数
	ok := S{Max: 1.0, Min: 1.5, Gt: 1.6, Lt: 2.4, Gte: 1.5, Lte: 2.5}
	if err := v.Validate(ok); err != nil {
		t.Fatalf("合法小数边界应通过:%v", err)
	}
	if err := v.Validate(S{Max: 1.0, Min: 1.4, Gt: 1.6, Lt: 2.4, Gte: 1.5, Lte: 2.5}); err == nil {
		t.Error("min=1.5 对 1.4 应失败")
	}
	if err := v.Validate(S{Max: 1.0, Min: 1.5, Gt: 1.5, Lt: 2.4, Gte: 1.5, Lte: 2.5}); err == nil {
		t.Error("gt=1.5 对 1.5 应失败(严格大于)")
	}
	if err := v.Validate(S{Max: 1.0, Min: 1.5, Gt: 1.6, Lt: 2.5, Gte: 1.5, Lte: 2.5}); err == nil {
		t.Error("lt=2.5 对 2.5 应失败(严格小于)")
	}
}

// ---------- 批量校验 ----------

func TestValidateSlice(t *testing.T) {
	v := newTestValidator(t)
	users := []Address{{City: "上海", ZipCode: "200000"}, {City: "北京", ZipCode: "100000"}}
	if err := v.Validate(users); err != nil {
		t.Fatalf("合法切片应通过:%v", err)
	}
	bad := []Address{{City: "上海"}, {City: "北京", ZipCode: "100000"}}
	err := v.Validate(bad)
	testx.RequireError(t, err)

	var hasPath bool
	if e, ok := errx.As(err); ok {
		for _, f := range e.Fields() {
			if f.Key == "field" && f.Value == "[0].ZipCode" {
				hasPath = true
			}
		}
	}
	testx.True(t, hasPath)

	// 空切片
	if err := v.Validate([]Address{}); err != nil {
		t.Fatalf("空切片应通过:%v", err)
	}
	// 元素非结构体
	if err := v.Validate([]int{1, 2}); err == nil {
		t.Error("非结构体元素应报错")
	}
}

func TestValidateMap(t *testing.T) {
	v := newTestValidator(t)
	items := map[string]Address{
		"home": {City: "上海", ZipCode: "200000"},
	}
	if err := v.Validate(items); err != nil {
		t.Fatalf("合法 map 应通过:%v", err)
	}
	err := v.Validate(map[string]Address{"home": {City: "上海"}})
	testx.RequireError(t, err)

	var hasPath bool
	if e, ok := errx.As(err); ok {
		for _, f := range e.Fields() {
			if f.Key == "field" && f.Value == "[home].ZipCode" {
				hasPath = true
			}
		}
	}
	testx.True(t, hasPath)

	// 空 map
	if err := v.Validate(map[string]Address{}); err != nil {
		t.Fatalf("空 map 应通过:%v", err)
	}
	// 值非结构体
	if err := v.Validate(map[string]int{"a": 1}); err == nil {
		t.Error("非结构体值应报错")
	}
}

func TestValidatePointerSlice(t *testing.T) {
	v := newTestValidator(t)
	users := []*Address{{City: "上海", ZipCode: "200000"}}
	if err := v.Validate(users); err != nil {
		t.Fatalf("指针元素切片应通过:%v", err)
	}
	if err := v.Validate([]*Address{nil}); err == nil {
		t.Fatal("nil 指针元素应报错(解引用失败)")
	}
}

func TestValidateSliceAggregate(t *testing.T) {
	v := newTestValidator(t)
	bad := []Address{{}, {}}
	err := v.Validate(bad)
	testx.RequireError(t, err)

	if _, ok := err.(*errx.Aggregate); !ok {
		t.Errorf("多元素失败应为聚合错误:%T", err)
	}
}

// ---------- 错误消息 ----------

func TestErrorMessageContainsPath(t *testing.T) {
	type S struct {
		Name string `validate:"required"`
	}
	v := newTestValidator(t)
	err := v.Validate(S{})
	testx.RequireError(t, err)

	if !strings.Contains(err.Error(), "Name") {
		t.Errorf("错误消息应含字段路径:%v", err)
	}
}
