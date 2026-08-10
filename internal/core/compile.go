package core

import (
	"reflect"
	"strings"

	"github.com/lcylpzls/errx"
)

// fieldRules 是单个字段的编译结果。
type fieldRules struct {
	index     int    // 字段索引
	name      string // 字段名(错误路径)
	rules     []Rule // 当前字段规则
	dive      bool   // 是否进入容器元素
	diveRules []Rule // dive 之后的元素规则
	nested    bool   // 结构体字段自动递归
	skip      bool   // "-" 跳过
}

// structRules 是结构体类型的编译结果。
type structRules struct {
	fields []fieldRules
}

// compileType 解析并缓存结构体类型的规则。
// 非法规则不缓存,保证配置错误每次暴露。
func (v *Validator) compileType(t reflect.Type) (*structRules, error) {
	if cached, ok := v.cache.Load(t); ok {
		return cached.(*structRules), nil
	}
	rules, err := v.parseType(t)
	if err != nil {
		return nil, err
	}
	v.cache.Store(t, rules)
	return rules, nil
}

// parseType 遍历结构体字段并编译各自规则。
func (v *Validator) parseType(t reflect.Type) (*structRules, error) {
	sr := &structRules{}
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if !f.IsExported() {
			continue // 未导出字段不参与校验
		}
		tag := f.Tag.Get(v.cfg.tagName)
		fr := fieldRules{index: i, name: f.Name}
		if tag == "-" {
			fr.skip = true
			sr.fields = append(sr.fields, fr)
			continue
		}
		if tag == "" {
			// 无显式规则:结构体字段自动递归。
			if isStructType(f.Type) {
				fr.nested = true
				sr.fields = append(sr.fields, fr)
			}
			continue
		}
		rules, err := v.compileRules(tag)
		if err != nil {
			return nil, err
		}
		for _, rule := range rules {
			if rule.name == "dive" {
				if fr.dive {
					return nil, errx.NewCodef(CodeInvalidRule,
						"字段 %s 出现多个 dive", f.Name)
				}
				fr.dive = true
				continue
			}
			if fr.dive {
				fr.diveRules = append(fr.diveRules, rule)
			} else {
				fr.rules = append(fr.rules, rule)
			}
		}
		// 跨字段规则:引用字段必须存在且已导出。
		for _, rule := range fr.rules {
			if rule.name == "eqfield" || rule.name == "nefield" ||
				rule.name == "required_if" || rule.name == "required_unless" ||
				rule.name == "excluded_if" {
				fieldName := rule.param
				if rule.name == "required_if" || rule.name == "required_unless" ||
					rule.name == "excluded_if" {
					parts := strings.Fields(rule.param)
					if len(parts) != 2 {
						return nil, errx.NewCodef(CodeInvalidRule,
							"字段 %s 的 %s 参数格式应为 字段名 值", f.Name, rule.name)
					}
					fieldName = parts[0]
				}
				other, ok := t.FieldByName(fieldName)
				if !ok {
					return nil, errx.NewCodef(CodeInvalidRule,
						"字段 %s 的 %s 引用了不存在的字段 %q", f.Name, rule.name, fieldName)
				}
				if !other.IsExported() {
					return nil, errx.NewCodef(CodeInvalidRule,
						"字段 %s 的 %s 不能引用未导出字段 %q", f.Name, rule.name, fieldName)
				}
			}
		}
		if isStructType(f.Type) {
			fr.nested = true
		}
		sr.fields = append(sr.fields, fr)
	}
	return sr, nil
}

// isStructType 判断类型(解引用指针后)是否为结构体。
func isStructType(t reflect.Type) bool {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Kind() == reflect.Struct
}

// ruleMeta 查询规则元信息:先内置,后自定义(v0.2.0 起)。
// ruleMeta 查询规则元信息:内置规则与自定义规则。
// 返回是否命中及是否为自定义规则(自定义规则参数可选)。
func (v *Validator) ruleMeta(name string) (ruleMeta, bool, bool) {
	if m, ok := builtinRules[name]; ok {
		return m, true, false
	}
	if _, ok := v.customFn(name); ok {
		return ruleMeta{}, true, true
	}
	return ruleMeta{}, false, false
}
