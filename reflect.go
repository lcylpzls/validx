package validx

import (
	"fmt"
	"reflect"

	"github.com/lcylpzls/errx"
)

// validateValue 校验单个值:应用字段规则、dive 元素、嵌套递归。
// path 为错误路径前缀(顶层为空串)。
func (v *Validator) validateValue(rv reflect.Value, path string, rules []Rule, dive bool, diveRules []Rule, nested bool) error {
	// 解引用指针:空指针只应用字段规则(required 可拦截),不递归。
	for rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return v.applyRules(rv, path, rules)
		}
		rv = rv.Elem()
	}
	if err := v.applyRules(rv, path, rules); err != nil {
		return err
	}
	if dive {
		switch rv.Kind() {
		case reflect.Slice, reflect.Array:
			for i := 0; i < rv.Len(); i++ {
				if err := v.validateElement(rv.Index(i), fmt.Sprintf("%s[%d]", path, i), diveRules); err != nil {
					return err
				}
			}
		case reflect.Map:
			iter := rv.MapRange()
			for iter.Next() {
				if err := v.validateElement(iter.Value(),
					fmt.Sprintf("%s[%v]", path, iter.Key()), diveRules); err != nil {
					return err
				}
			}
		default:
			return v.fieldErr(path, "dive", "dive 仅适用于切片/数组/map,当前类型 %s", rv.Kind())
		}
	}
	if nested && rv.Kind() == reflect.Struct {
		return v.validateStruct(rv, path)
	}
	return nil
}

// validateElement 校验容器元素:dive 规则 + 结构体递归。
func (v *Validator) validateElement(rv reflect.Value, path string, diveRules []Rule) error {
	for rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return nil // 空元素无规则可应用(required 由 dive 规则处理)
		}
		rv = rv.Elem()
	}
	if err := v.applyRules(rv, path, diveRules); err != nil {
		return err
	}
	if rv.Kind() == reflect.Struct {
		return v.validateStruct(rv, path)
	}
	return nil
}

// applyRules 顺序执行规则,omitempty 且为空时跳过后续规则。
func (v *Validator) applyRules(rv reflect.Value, path string, rules []Rule) error {
	for _, rule := range rules {
		if rule.name == "omitempty" && isEmpty(rv) {
			break
		}
		if err := v.evalRule(rule, rv, path); err != nil {
			return err
		}
	}
	return nil
}

// validateStruct 校验结构体全部字段,聚合失败错误。
func (v *Validator) validateStruct(rv reflect.Value, path string) error {
	sr, err := v.compileType(rv.Type())
	if err != nil {
		return err
	}
	var errs []error
	for _, fr := range sr.fields {
		if fr.skip {
			continue
		}
		fpath := fr.name
		if path != "" {
			fpath = path + "." + fr.name
		}
		if err := v.validateValue(rv.Field(fr.index), fpath, fr.rules, fr.dive, fr.diveRules, fr.nested); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) == 0 {
		return nil
	}
	if len(errs) == 1 {
		return errs[0]
	}
	return errx.Join(errs...)
}
