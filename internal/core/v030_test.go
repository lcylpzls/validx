package core

import (
	testx "github.com/lcylpzls/testx"
	"reflect"
	"testing"

	"github.com/lcylpzls/errx"
)

// ---------- 跨字段规则 ----------

type RegisterForm struct {
	Password string `validate:"required,min=6"`
	Confirm  string `validate:"required,eqfield=Password"`
	Username string `validate:"required,nefield=Password"`
	Role     string `validate:"eq=admin"`
	Level    int    `validate:"ne=0"`
}

func TestEqField(t *testing.T) {
	v := newTestValidator(t)
	ok := RegisterForm{Password: "secret123", Confirm: "secret123", Username: "tom", Role: "admin", Level: 1}
	if err := v.Validate(ok); err != nil {
		t.Fatalf("合法跨字段应通过:%v", err)
	}
	err := v.Validate(RegisterForm{Password: "secret123", Confirm: "different", Username: "tom", Role: "admin", Level: 1})
	testx.RequireError(t, err)

	if e, isErr := errx.As(err); isErr {
		for _, f := range e.Fields() {
			if f.Key == "field" && f.Value != "Confirm" {
				t.Errorf("跨字段错误路径应为 Confirm:%v", f.Value)
			}
		}
	}
}

func TestNeField(t *testing.T) {
	v := newTestValidator(t)
	if err := v.Validate(RegisterForm{Password: "secret123", Confirm: "secret123", Username: "secret123", Role: "admin", Level: 1}); err == nil {
		t.Error("Username 与 Password 相同应失败")
	}
}

func TestEqNe(t *testing.T) {
	v := newTestValidator(t)
	if err := v.Validate(RegisterForm{Password: "secret123", Confirm: "secret123", Username: "tom", Role: "staff", Level: 1}); err == nil {
		t.Error("Role 不等于 admin 应失败")
	}
	if err := v.Validate(RegisterForm{Password: "secret123", Confirm: "secret123", Username: "tom", Role: "admin", Level: 0}); err == nil {
		t.Error("Level 等于 0 应失败")
	}
}

func TestEqFieldPointer(t *testing.T) {
	type S struct {
		A *int `validate:"required"`
		B *int `validate:"eqfield=A"`
	}
	v := newTestValidator(t)
	a, b := 1, 1
	if err := v.Validate(S{A: &a, B: &b}); err != nil {
		t.Fatalf("指针跨字段应解引用比较:%v", err)
	}
	c := 2
	if err := v.Validate(S{A: &a, B: &c}); err == nil {
		t.Error("指针值不同应失败")
	}
}

func TestCrossFieldCompileErrors(t *testing.T) {
	type Missing struct {
		A string `validate:"eqfield=NoSuch"`
	}
	type Unexported struct {
		A    string `validate:"nefield=priv"`
		priv string
	}
	v := newTestValidator(t)
	if err := v.Validate(Missing{}); err == nil {
		t.Error("引用不存在字段应报错")
	} else if code, _ := errx.CodeOf(err); code != CodeInvalidRule {
		t.Errorf("错误码 = %s,want %s", code, CodeInvalidRule)
	}
	if err := v.Validate(Unexported{A: "x", priv: "y"}); err == nil {
		t.Error("引用未导出字段应报错")
	}
}

func TestCrossFieldInDiveRejected(t *testing.T) {
	type S struct {
		Items []string `validate:"dive,eqfield=X"`
	}
	v := newTestValidator(t)
	err := v.Validate(S{Items: []string{"a"}})
	testx.RequireError(t, err)

	if code, _ := errx.CodeOf(err); code != CodeInvalidRule {
		t.Errorf("错误码 = %s,want %s", code, CodeInvalidRule)
	}
}

func TestDerefValue(t *testing.T) {
	if got := derefValue(reflect.ValueOf(&nilTest)); got.Kind() != reflect.Int {
		t.Errorf("指针解引用失败:%v", got.Kind())
	}
	var p *int
	if got := derefValue(reflect.ValueOf(p)); !got.IsNil() {
		t.Error("空指针应原样返回")
	}
}

var nilTest = 1

func TestEvalRuleMissingFieldDirect(t *testing.T) {
	// 运行时防御分支:编译期已拦截,直接调用时验证不 panic。
	v := newTestValidator(t)
	err := v.evalRule(Rule{name: "eqfield", param: "NoSuch"},
		reflect.ValueOf("x"), "F", reflect.ValueOf(struct{ A string }{}))
	testx.RequireError(t, err)

}
