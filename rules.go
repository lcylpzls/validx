package validx

import (
	"encoding/base64"
	"encoding/json"
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
	// v0.3.0 跨字段
	"eqfield": {needsParam: true},
	"nefield": {needsParam: true},
	"eq":      {needsParam: true},
	"ne":      {needsParam: true},
	// v0.7.0 条件必填与格式
	"required_if":     {needsParam: true},
	"required_unless": {needsParam: true},
	"excluded_if":     {needsParam: true},
	"base64":          {},
	"json":            {},
	"hexadecimal":     {},
	"mac":             {},
	"semver":          {},
	"contains":        {needsParam: true},
	"excludes":        {needsParam: true},
	"startswith":      {needsParam: true},
	"endswith":        {needsParam: true},
	// v0.8.0 网络规则
	"ipv4":     {},
	"ipv6":     {},
	"hostname": {},
	"fqdn":     {},
	"port":     {},
}

// uuidPattern 是标准 UUID 格式(8-4-4-4-12)。
var uuidPattern = regexp.MustCompile(
	`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// macPattern 是 MAC 地址格式(冒号或连字符分隔)。
var macPattern = regexp.MustCompile(`^([0-9A-Fa-f]{2}[:-]){5}[0-9A-Fa-f]{2}$`)

// semverPattern 是语义化版本格式(可带 v 前缀/预发布/构建元数据)。
var semverPattern = regexp.MustCompile(
	`^v?\d+\.\d+\.\d+(?:-[0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*)?(?:\+[0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*)?$`)

// Rule 是编译后的单条规则。
type Rule struct {
	name    string
	param   string
	limit   float64  // min/max/len/gt/lt/gte/lte 预解析参数(支持小数)
	options []string // oneof 预拆分枚举
	regexp  *regexp.Regexp
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
		switch name {
		case "regexp":
			re, err := regexp.Compile(param)
			if err != nil {
				return nil, errx.Wrap(err, errx.KindInvalid, CodeInvalidRule, "正则表达式非法")
			}
			rule.regexp = re
		case "min", "max", "len", "gt", "lt", "gte", "lte":
			limit, err := strconv.ParseFloat(param, 64)
			if err != nil {
				return nil, errx.Newf(errx.KindInvalid, CodeInvalidRule,
					"规则 %s 参数必须是数值:%q", name, param)
			}
			rule.limit = limit
		case "oneof":
			rule.options = strings.Fields(param)
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
func (v *Validator) evalRule(rule Rule, rv reflect.Value, path string, parent reflect.Value) error {
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
		matched := false
		for _, opt := range rule.options {
			if opt == val {
				matched = true
				break
			}
		}
		if !matched {
			return v.fieldErr(path, rule.name, "取值必须在 %q 中", rule.param)
		}
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
	case "ipv4":
		return v.checkStringRule(rule, rv, path, func(s string) bool {
			ip := net.ParseIP(s)
			return ip != nil && ip.To4() != nil
		})
	case "ipv6":
		return v.checkStringRule(rule, rv, path, func(s string) bool {
			ip := net.ParseIP(s)
			return ip != nil && ip.To4() == nil
		})
	case "hostname":
		return v.checkStringRule(rule, rv, path, isValidHostname)
	case "fqdn":
		return v.checkStringRule(rule, rv, path, func(s string) bool {
			return strings.Contains(s, ".") && isValidHostname(s)
		})
	case "port":
		if rv.Kind() == reflect.String {
			n, err := strconv.ParseUint(rv.String(), 10, 16)
			if err != nil || n > 65535 {
				return v.fieldErr(path, rule.name, "端口必须在 0-65535")
			}
			return nil
		}
		n, ok := numericLength(rv)
		if !ok || n < 0 || n > 65535 {
			return v.fieldErr(path, rule.name, "端口必须在 0-65535")
		}
		return nil
	case "datetime":
		if rv.Kind() != reflect.String {
			return v.fieldErr(path, rule.name, "datetime 仅适用于字符串,当前类型 %s", rv.Kind())
		}
		if _, err := time.Parse(rule.param, rv.String()); err != nil {
			return v.fieldErr(path, rule.name, "时间格式不匹配 %q", rule.param)
		}
	case "gt", "lt", "gte", "lte":
		return v.evalCompareRule(rule, rv, path)
	case "eqfield", "nefield":
		if !parent.IsValid() {
			return errx.Newf(errx.KindInvalid, CodeInvalidRule,
				"规则 %s 需要结构体上下文(不适用于 dive 元素)", rule.name)
		}
		other := derefValue(parent.FieldByName(rule.param))
		if !other.IsValid() {
			return errx.Newf(errx.KindInvalid, CodeInvalidRule,
				"规则 %s 引用了不存在的字段 %q", rule.name, rule.param)
		}
		equal := reflect.DeepEqual(other.Interface(), rv.Interface())
		if (rule.name == "eqfield" && !equal) || (rule.name == "nefield" && equal) {
			return v.fieldErr(path, rule.name, "字段与 %s 不满足 %s", rule.param, rule.name)
		}
	case "eq", "ne":
		equal := stringify(rv) == rule.param
		if (rule.name == "eq" && !equal) || (rule.name == "ne" && equal) {
			return v.fieldErr(path, rule.name, "取值不满足 %s=%s", rule.name, rule.param)
		}
	case "required_if", "required_unless":
		if !parent.IsValid() {
			return errx.Newf(errx.KindInvalid, CodeInvalidRule,
				"规则 %s 需要结构体上下文(不适用于 dive 元素)", rule.name)
		}
		parts := strings.Fields(rule.param)
		if len(parts) != 2 {
			return errx.Newf(errx.KindInvalid, CodeInvalidRule,
				"规则 %s 参数格式应为 字段名 值,got %q", rule.name, rule.param)
		}
		other := derefValue(parent.FieldByName(parts[0]))
		if !other.IsValid() {
			return errx.Newf(errx.KindInvalid, CodeInvalidRule,
				"规则 %s 引用了不存在的字段 %q", rule.name, parts[0])
		}
		cond := stringify(other) == parts[1]
		needRequired := (rule.name == "required_if" && cond) ||
			(rule.name == "required_unless" && !cond)
		if needRequired && isEmpty(rv) {
			return v.fieldErr(path, rule.name, "字段在 %s=%s 条件下为必填项", parts[0], parts[1])
		}
	case "excluded_if":
		if !parent.IsValid() {
			return errx.Newf(errx.KindInvalid, CodeInvalidRule,
				"规则 %s 需要结构体上下文(不适用于 dive 元素)", rule.name)
		}
		parts := strings.Fields(rule.param)
		if len(parts) != 2 {
			return errx.Newf(errx.KindInvalid, CodeInvalidRule,
				"规则 %s 参数格式应为 字段名 值,got %q", rule.name, rule.param)
		}
		other := derefValue(parent.FieldByName(parts[0]))
		if !other.IsValid() {
			return errx.Newf(errx.KindInvalid, CodeInvalidRule,
				"规则 %s 引用了不存在的字段 %q", rule.name, parts[0])
		}
		if stringify(other) == parts[1] && !isEmpty(rv) {
			return v.fieldErr(path, rule.name, "字段在 %s=%s 条件下必须为空", parts[0], parts[1])
		}
	case "base64":
		return v.checkStringRule(rule, rv, path, func(s string) bool {
			_, err := base64.StdEncoding.DecodeString(s)
			return err == nil
		})
	case "json":
		return v.checkStringRule(rule, rv, path, func(s string) bool {
			return json.Valid([]byte(s))
		})
	case "hexadecimal":
		return v.checkStringRule(rule, rv, path, func(s string) bool {
			if s == "" {
				return false
			}
			for _, r := range s {
				if !(r >= '0' && r <= '9' || r >= 'a' && r <= 'f' || r >= 'A' && r <= 'F') {
					return false
				}
			}
			return true
		})
	case "mac":
		return v.checkStringRule(rule, rv, path, macPattern.MatchString)
	case "semver":
		return v.checkStringRule(rule, rv, path, semverPattern.MatchString)
	case "contains":
		return v.checkStringRule(rule, rv, path, func(s string) bool {
			return strings.Contains(s, rule.param)
		})
	case "excludes":
		return v.checkStringRule(rule, rv, path, func(s string) bool {
			return !strings.Contains(s, rule.param)
		})
	case "startswith":
		return v.checkStringRule(rule, rv, path, func(s string) bool {
			return strings.HasPrefix(s, rule.param)
		})
	case "endswith":
		return v.checkStringRule(rule, rv, path, func(s string) bool {
			return strings.HasSuffix(s, rule.param)
		})
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

// derefValue 解引用指针到最终值(空指针原样返回)。
func derefValue(v reflect.Value) reflect.Value {
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return v
		}
		v = v.Elem()
	}
	return v
}

// isValidHostname 校验 RFC 1123 主机名:
// 点分标签,每标签 1-63 字符,字母数字与连字符,首尾不能为连字符。
func isValidHostname(s string) bool {
	if s == "" || len(s) > 253 {
		return false
	}
	for _, label := range strings.Split(s, ".") {
		if label == "" || len(label) > 63 {
			return false
		}
		if label[0] == '-' || label[len(label)-1] == '-' {
			return false
		}
		for i := 0; i < len(label); i++ {
			c := label[i]
			if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' ||
				c >= '0' && c <= '9' || c == '-') {
				return false
			}
		}
	}
	return true
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
	n, ok := numericLength(rv)
	if !ok {
		return v.fieldErr(path, rule.name, "规则不适用于类型 %s", rv.Kind())
	}
	limit := rule.limit
	switch rule.name {
	case "gt":
		if n <= limit {
			return v.fieldErr(path, rule.name, "值 %v 必须大于 %v", n, limit)
		}
	case "lt":
		if n >= limit {
			return v.fieldErr(path, rule.name, "值 %v 必须小于 %v", n, limit)
		}
	case "gte":
		if n < limit {
			return v.fieldErr(path, rule.name, "值 %v 必须大于等于 %v", n, limit)
		}
	case "lte":
		if n > limit {
			return v.fieldErr(path, rule.name, "值 %v 必须小于等于 %v", n, limit)
		}
	}
	return nil
}

// evalLengthRule 执行 min / max / len 规则:
// 数值比较数值大小,字符串比较字符数,容器比较元素数。
func (v *Validator) evalLengthRule(rule Rule, rv reflect.Value, path string) error {
	n, ok := numericLength(rv)
	if !ok {
		return v.fieldErr(path, rule.name, "规则不适用于类型 %s", rv.Kind())
	}
	limit := rule.limit
	switch rule.name {
	case "min":
		if n < limit {
			return v.fieldErr(path, rule.name, "值 %v 小于下限 %v", n, limit)
		}
	case "max":
		if n > limit {
			return v.fieldErr(path, rule.name, "值 %v 超过上限 %v", n, limit)
		}
	case "len":
		if n != limit {
			return v.fieldErr(path, rule.name, "长度 %v 不等于 %v", n, limit)
		}
	}
	return nil
}

// numericLength 将值归一为可比较的数值:
// 数值类型取数值,字符串取字符数,容器取元素数。
func numericLength(rv reflect.Value) (float64, bool) {
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(rv.Int()), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return float64(rv.Uint()), true
	case reflect.Float32, reflect.Float64:
		return rv.Float(), true
	case reflect.String:
		return float64(utf8.RuneCountInString(rv.String())), true
	case reflect.Slice, reflect.Array, reflect.Map:
		return float64(rv.Len()), true
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
	return errx.Newf(errx.KindInvalid, CodeValidationFailed,
		"字段 %s:%s", path, fmt.Sprintf(format, args...)).
		WithField("field", path).
		WithField("rule", rule)
}
