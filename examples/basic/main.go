// basic 示例:用户注册参数校验(含跨字段确认)。
package main

import (
	"fmt"

	"github.com/lcylpzls/validx"
)

// RegisterReq 是注册请求参数。
type RegisterReq struct {
	Username string `validate:"required,min=3,max=32,alphanum"`
	Email    string `validate:"required,email"`
	Password string `validate:"required,min=8,max=64"`
	Confirm  string `validate:"required,eqfield=Password"`
	Age      int    `validate:"gte=0,lte=150"`
	Role     string `validate:"oneof=admin user guest"`
}

func main() {
	v, err := validx.New()
	if err != nil {
		panic(err)
	}
	req := RegisterReq{
		Username: "zhangsan",
		Email:    "zhangsan@example.com",
		Password: "secret123",
		Confirm:  "secret123",
		Age:      30,
		Role:     "user",
	}
	if err := v.Validate(req); err != nil {
		fmt.Println("校验失败:", err)
		return
	}
	fmt.Println("注册参数合法")
}
