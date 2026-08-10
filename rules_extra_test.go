package validx

import (
	"testing"

	"github.com/lcylpzls/testx"
)

func TestLowercaseRule(t *testing.T) {
	testx.RequireNoError(t, ValidateField("abc123", "lowercase"))
	testx.RequireError(t, ValidateField("Abc", "lowercase"))
	testx.RequireError(t, ValidateField("", "lowercase"))
	testx.RequireError(t, ValidateField(123, "lowercase"))
}

func TestUppercaseRule(t *testing.T) {
	testx.RequireNoError(t, ValidateField("ABC123", "uppercase"))
	testx.RequireError(t, ValidateField("Abc", "uppercase"))
	testx.RequireError(t, ValidateField("", "uppercase"))
	testx.RequireError(t, ValidateField(123, "uppercase"))
}

func TestMinMaxBytesRules(t *testing.T) {
	// 中文按字节计：2 个字符 = 6 字节。
	testx.RequireNoError(t, ValidateField("你好", "minbytes=6,maxbytes=6"))
	testx.RequireError(t, ValidateField("你好", "maxbytes=5"))
	testx.RequireError(t, ValidateField("你好", "minbytes=7"))
	// 与字符数规则区分：len=2 通过，maxbytes=2 失败。
	testx.RequireNoError(t, ValidateField("你好", "len=2"))
	testx.RequireError(t, ValidateField("你好", "maxbytes=2"))
	testx.RequireError(t, ValidateField("abc", "minbytes=4"))
	testx.RequireError(t, ValidateField(123, "maxbytes=3"))
}

func TestByteRulesInStruct(t *testing.T) {
	type S struct {
		Key string `validate:"required,minbytes=1,maxbytes=1024"`
	}
	testx.RequireNoError(t, Validate(S{Key: "对象键"}))
	bad := make([]byte, 1025)
	for i := range bad {
		bad[i] = 'a'
	}
	testx.RequireError(t, Validate(S{Key: string(bad)}))
}

func TestByteRulesInvalidParam(t *testing.T) {
	_, err := newTestValidator(t).compileFieldRules("maxbytes=abc")
	testx.RequireErrCode(t, err, CodeInvalidRule)
}
