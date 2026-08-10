package validx

// defaultValidator 是包级默认校验器，全局规则注册与快捷校验共用。
// 默认配置是编译期确定的合法值，New 必然成功；失败属于编程错误，
// 忽略错误后由首次使用时的空指针 panic 暴露。
var defaultValidator, _ = New()

// Validate 使用包级默认校验器校验结构体。
func Validate(value any) error {
	return defaultValidator.Validate(value)
}

// ValidateField 使用包级默认校验器校验单个值（不支持 dive）。
func ValidateField(value any, rules string) error {
	return defaultValidator.ValidateField(value, rules)
}

// RegisterRule 全局注册自定义规则（线程安全）。
// 规则名须合法且不与内置规则冲突；重复注册同名规则以最后一次为准。
func RegisterRule(name string, fn ValidationFunc) error {
	return defaultValidator.RegisterValidation(name, fn)
}
