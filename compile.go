package validx

import (
	"reflect"

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
					return nil, errx.Newf(errx.KindInvalid, CodeInvalidRule,
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
func (v *Validator) ruleMeta(name string) (ruleMeta, bool) {
	m, ok := builtinRules[name]
	return m, ok
}
