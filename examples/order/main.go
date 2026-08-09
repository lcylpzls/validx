// order 示例:订单参数校验(dive 元素 + 数值范围)。
package main

import (
	"fmt"

	"github.com/lcylpzls/validx"
)

// OrderItem 是订单条目。
type OrderItem struct {
	SKU   string  `validate:"required,alphanum"`
	Qty   int     `validate:"gt=0,lte=100"`
	Price float64 `validate:"gte=0"`
}

// OrderReq 是下单请求参数。
type OrderReq struct {
	OrderNo string      `validate:"required,len=16"`
	Status  string      `validate:"oneof=pending paid shipped cancelled"`
	Items   []OrderItem `validate:"required,dive"`
	Remark  string      `validate:"omitempty,max=200"`
}

func main() {
	v, err := validx.New()
	if err != nil {
		panic(err)
	}
	req := OrderReq{
		OrderNo: "2026080900000001",
		Status:  "pending",
		Items: []OrderItem{
			{SKU: "SKU001", Qty: 2, Price: 99.9},
		},
	}
	if err := v.Validate(req); err != nil {
		fmt.Println("校验失败:", err)
		return
	}
	fmt.Println("订单参数合法")
}
