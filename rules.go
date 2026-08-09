package validx

import (
	"fmt"
	"net"
	"net/mail"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
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
	// v0.2.0 扩充
	"alpha":    {},
	"alphanum": {},
	"numeric":  {},
	"boolean":  {},
	"uuid":     {},
	"url":      {},
	"ip":       {},
	"datetime": {needsParam: true},
	"gt":       {needsParam: true},
	"lt":       {needsParam: true},
	"gte":      {needsParam: true},
	"lte":      {needsParam: true},
}

// uuidPattern 是标准 UUID 格式(8-4-4-4-12)。
var uuidPattern = regexp.MustCompile(
	`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

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
		meta, ok, isCustom := v.ruleMeta(name)
		if !ok {
			return nil, errx.Newf(errx.KindInvalid, CodeInvalidRule, "未知规则 %q", name)
		}
		if !isCustom {
			if meta.needsParam && !hasParam {
				return nil, errx.Newf(errx.KindInvalid, CodeInvalidRule, "规则 %s 缺少参数", name)
			}
			if !meta.needsParam && hasParam {
				return nil, errx.Newf(errx.KindInvalid, CodeInvalidRule, "规则 %s 不接受参数", name)
			}
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
	if !v.IsValid() {
		return true
	}
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
	case "alpha":
		return v.checkStringRule(rule, rv, path, func(s string) bool {
			for _, r := range s {
				if !unicode.IsLetter(r) {
					return false
				}
			}
			return s != ""
		})
	case "alphanum":
		return v.checkStringRule(rule, rv, path, func(s string) bool {
			for _, r := range s {
				if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
					return false
				}
			}
			return s != ""
		})
	case "numeric":
		return v.checkStringRule(rule, rv, path, func(s string) bool {
			if s == "" {
				return false
			}
			for _, r := range s {
				if !unicode.IsDigit(r) {
					return false
				}
			}
			return true
		})
	case "boolean":
		if rv.Kind() != reflect.String {
			return v.fieldErr(path, rule.name, "boolean 仅适用于字符串,当前类型 %s", rv.Kind())
		}
		if _, err := strconv.ParseBool(rv.String()); err != nil {
			return v.fieldErr(path, rule.name, "不是合法布尔字符串")
		}
	case "uuid":
		return v.checkStringRule(rule, rv, path, uuidPattern.MatchString)
	case "url":
		return v.checkStringRule(rule, rv, path, func(s string) bool {
			u, err := url.Parse(s)
			return err == nil && u.IsAbs() && u.Host != ""
		})
	case "ip":
		return v.checkStringRule(rule, rv, path, func(s string) bool {
			return net.ParseIP(s) != nil
		})
	case "datetime":
		if rv.Kind() != reflect.String {
			return v.fieldErr(path, rule.name, "datetime 仅适用于字符串,当前类型 %s", rv.Kind())
		}
		if _, err := time.Parse(rule.param, rv.String()); err != nil {
			return v.fieldErr(path, rule.name, "时间格式不匹配 %q", rule.param)
		}
	case "gt", "lt", "gte", "lte":
		return v.evalCompareRule(rule, rv, path)
	default:
		if fn, ok := v.customFn(rule.name); ok {
			if err := fn(rv.Interface(), rule.param, path); err != nil {
				if _, isErr := errx.As(err); isErr {
					return err
				}
				return errx.Wrap(err, errx.KindInvalid, CodeValidationFailed, "自定义规则校验失败").
					WithField("field", path).
					WithField("rule", rule.name)
			}
			return nil
		}
		return errx.Newf(errx.KindInvalid, CodeInvalidRule, "未知规则 %q", rule.name)
	}
	return nil
}

// checkStringRule 校验字符串规则,非字符串类型直接失败。
func (v *Validator) checkStringRule(rule Rule, rv reflect.Value, path string,
	check func(string) bool) error {
	if rv.Kind() != reflect.String {
		return v.fieldErr(path, rule.name, "%s 仅适用于字符串,当前类型 %s", rule.name, rv.Kind())
	}
	if !check(rv.String()) {
		return v.fieldErr(path, rule.name, "不满足规则 %s", rule.name)
	}
	return nil
}

// evalCompareRule 执行 gt / lt / gte / lte 规则(严格/非严格数值比较)。
func (v *Validator) evalCompareRule(rule Rule, rv reflect.Value, path string) error {
	limit, err := strconv.ParseInt(rule.param, 10, 64)
	if err != nil {
		return errx.Newf(errx.KindInvalid, CodeInvalidRule,
			"规则 %s 参数必须是整数:%q", rule.name, rule.param)
	}
	n, ok := numericLength(rv)
	if !ok {
		return v.fieldErr(path, rule.name, "规则不适用于类型 %s", rv.Kind())
	}
	switch rule.name {
	case "gt":
		if n <= limit {
			return v.fieldErr(path, rule.name, "值 %d 必须大于 %d", n, limit)
		}
	case "lt":
		if n >= limit {
			return v.fieldErr(path, rule.name, "值 %d 必须小于 %d", n, limit)
		}
	case "gte":
		if n < limit {
			return v.fieldErr(path, rule.name, "值 %d 必须大于等于 %d", n, limit)
		}
	case "lte":
		if n > limit {
			return v.fieldErr(path, rule.name, "值 %d 必须小于等于 %d", n, limit)
		}
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
	n, ok := numericLength(rv)
	if !ok {
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

// numericLength 将值归一为可比较的数值:
// 数值类型取数值,字符串取字符数,容器取元素数。
func numericLength(rv reflect.Value) (int64, bool) {
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv.Int(), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		u := rv.Uint()
		if u > uint64(^uint64(0)>>1) {
			return int64(^uint64(0) >> 1), true
		}
		return int64(u), true
	case reflect.Float32, reflect.Float64:
		return int64(rv.Float()), true
	case reflect.String:
		return int64(utf8.RuneCountInString(rv.String())), true
	case reflect.Slice, reflect.Array, reflect.Map:
		return int64(rv.Len()), true
	default:
		return 0, false
	}
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
