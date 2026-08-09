package validx

import (
	"fmt"
	"net/mail"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/lcylpzls/errx"
)

// ruleMeta 描述规则的参数形态。
type ruleMeta struct {
	needsParam bool
}

// builtinRules 是 v0.1.0 内置规则集。
var builtinRules = map[string]ruleMeta{
	"required":  {},
	"omitempty": {},
	"min":       {needsParam: true},
	"max":       {needsParam: true},
	"len":       {needsParam: true},
	"email":     {},
	"regexp":    {needsParam: true},
	"oneof":     {needsParam: true},
	"dive":      {},
}

// Rule 是编译后的单条规则。
type Rule struct {
	name   string
	param  string
	regexp *regexp.Regexp
}

// compileRules 解析 tag 字符串为规则列表,并预编译正则。
func (v *Validator) compileRules(tag string) ([]Rule, error) {
	parts := strings.Split(tag, ",")
	rules := make([]Rule, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		name, param, hasParam := strings.Cut(part, "=")
		name = strings.TrimSpace(name)
		param = strings.TrimSpace(param)
		if !validRuleName(name) {
			return nil, errx.Newf(errx.KindInvalid, CodeInvalidRule, "非法规则名 %q", name)
		}
		meta, ok := v.ruleMeta(name)
		if !ok {
			return nil, errx.Newf(errx.KindInvalid, CodeInvalidRule, "未知规则 %q", name)
		}
		if meta.needsParam && !hasParam {
			return nil, errx.Newf(errx.KindInvalid, CodeInvalidRule, "规则 %s 缺少参数", name)
		}
		if !meta.needsParam && hasParam {
			return nil, errx.Newf(errx.KindInvalid, CodeInvalidRule, "规则 %s 不接受参数", name)
		}
		rule := Rule{name: name, param: param}
		if name == "regexp" {
			re, err := regexp.Compile(param)
			if err != nil {
				return nil, errx.Wrap(err, errx.KindInvalid, CodeInvalidRule, "正则表达式非法")
			}
			rule.regexp = re
		}
		rules = append(rules, rule)
	}
	return rules, nil
}

// validRuleName 校验规则名仅含字母、数字、下划线。
func validRuleName(name string) bool {
	if name == "" {
		return false
	}
	for i := 0; i < len(name); i++ {
		c := name[i]
		if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '_') {
			return false
		}
	}
	return true
}

// isEmpty 判断值是否为零值(required / omitempty 语义)。
// 结构体永不视为空;time.Time 等由具体规则处理。
func isEmpty(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Ptr, reflect.Interface, reflect.Slice, reflect.Map, reflect.Chan, reflect.Func:
		return v.IsNil()
	default:
		return false
	}
}

// evalRule 对值执行单条规则,返回 nil 或字段错误。
// 调用方保证 rv 已解引用且非 nil。
func (v *Validator) evalRule(rule Rule, rv reflect.Value, path string) error {
	switch rule.name {
	case "required":
		if isEmpty(rv) {
			return v.fieldErr(path, rule.name, "字段为必填项")
		}
	case "omitempty":
		// applyRules 已在空值时跳过后续规则;此处仅保证直接调用安全。
	case "min", "max", "len":
		return v.evalLengthRule(rule, rv, path)
	case "email":
		if rv.Kind() != reflect.String {
			return v.fieldErr(path, rule.name, "email 仅适用于字符串,当前类型 %s", rv.Kind())
		}
		addr, err := mail.ParseAddress(rv.String())
		if err != nil || addr.Address != rv.String() {
			return v.fieldErr(path, rule.name, "邮箱格式非法")
		}
	case "regexp":
		if rv.Kind() != reflect.String {
			return v.fieldErr(path, rule.name, "regexp 仅适用于字符串,当前类型 %s", rv.Kind())
		}
		if !rule.regexp.MatchString(rv.String()) {
			return v.fieldErr(path, rule.name, "不匹配正则 %q", rule.param)
		}
	case "oneof":
		val := stringify(rv)
		for _, opt := range strings.Fields(rule.param) {
			if opt == val {
				return nil
			}
		}
		return v.fieldErr(path, rule.name, "取值必须在 %q 中", rule.param)
	default:
		return errx.Newf(errx.KindInvalid, CodeInvalidRule, "未知规则 %q", rule.name)
	}
	return nil
}

// evalLengthRule 执行 min / max / len 规则:
// 数值比较数值大小,字符串比较字符数,容器比较元素数。
func (v *Validator) evalLengthRule(rule Rule, rv reflect.Value, path string) error {
	limit, err := strconv.ParseInt(rule.param, 10, 64)
	if err != nil {
		return errx.Newf(errx.KindInvalid, CodeInvalidRule,
			"规则 %s 参数必须是整数:%q", rule.name, rule.param)
	}
	var n int64
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n = rv.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		u := rv.Uint()
		if u > uint64(^uint64(0)>>1) {
			n = int64(^uint64(0) >> 1)
		} else {
			n = int64(u)
		}
	case reflect.Float32, reflect.Float64:
		n = int64(rv.Float())
	case reflect.String:
		n = int64(utf8.RuneCountInString(rv.String()))
	case reflect.Slice, reflect.Array, reflect.Map:
		n = int64(rv.Len())
	default:
		return v.fieldErr(path, rule.name, "规则不适用于类型 %s", rv.Kind())
	}
	switch rule.name {
	case "min":
		if n < limit {
			return v.fieldErr(path, rule.name, "值 %d 小于下限 %d", n, limit)
		}
	case "max":
		if n > limit {
			return v.fieldErr(path, rule.name, "值 %d 超过上限 %d", n, limit)
		}
	case "len":
		if n != limit {
			return v.fieldErr(path, rule.name, "长度 %d 不等于 %d", n, limit)
		}
	}
	return nil
}

// stringify 将值转换为用于 oneof 比较的字符串。
func stringify(rv reflect.Value) string {
	switch rv.Kind() {
	case reflect.String:
		return rv.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(rv.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return strconv.FormatUint(rv.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(rv.Float(), 'f', -1, 64)
	case reflect.Bool:
		return strconv.FormatBool(rv.Bool())
	default:
		return fmt.Sprint(rv.Interface())
	}
}

// fieldErr 构造字段校验失败错误,携带 field 与 rule 字段。
func (v *Validator) fieldErr(path, rule, format string, args ...any) error {
	return errx.Newf(errx.KindInvalid, CodeValidationFailed, format, args...).
		WithField("field", path).
		WithField("rule", rule)
}
