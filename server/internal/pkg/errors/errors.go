package errors

import (
	"fmt"
	"net/http"
)

// 错误码定义
const (
	// 通用错误码 1000-1999
	CodeSuccess          = 0
	CodeInternalError    = 1000
	CodeInvalidParam     = 1001
	CodeUnauthorized     = 1002
	CodeForbidden        = 1003
	CodeNotFound         = 1004
	CodeAlreadyExists    = 1005
	CodeDatabaseError    = 1006
	CodeCacheError       = 1007
	CodeExternalAPIError = 1008

	// 用户服务错误码 2000-2999
	CodeUserNotFound      = 2000
	CodeUserAlreadyExists = 2001
	CodePasswordError     = 2002
	CodeTokenExpired      = 2003
	CodeTokenInvalid      = 2004

	// 商品服务错误码 3000-3999
	CodeProductNotFound  = 3000
	CodeProductOffline   = 3001
	CodeSkuNotFound      = 3002
	CodeCategoryNotFound = 3003

	// 订单服务错误码 4000-4999
	CodeOrderNotFound    = 4000
	CodeOrderStatusError = 4001
	CodeOrderCanceled    = 4002
	CodeOrderPaid        = 4003

	// 库存服务错误码 5000-5999
	CodeStockInsufficient = 5000
	CodeStockLocked       = 5001

	// 支付服务错误码 6000-6999
	CodePaymentFailed  = 6000
	CodePaymentExpired = 6001
	CodeRefundFailed   = 6002
)

// 错误信息映射
var errorMessages = map[int]string{
	CodeSuccess:          "成功",
	CodeInternalError:    "内部服务器错误",
	CodeInvalidParam:     "参数错误",
	CodeUnauthorized:     "未授权",
	CodeForbidden:        "禁止访问",
	CodeNotFound:         "资源不存在",
	CodeAlreadyExists:    "资源已存在",
	CodeDatabaseError:    "数据库错误",
	CodeCacheError:       "缓存错误",
	CodeExternalAPIError: "外部API调用失败",

	CodeUserNotFound:      "用户不存在",
	CodeUserAlreadyExists: "用户已存在",
	CodePasswordError:     "密码错误",
	CodeTokenExpired:      "Token已过期",
	CodeTokenInvalid:      "Token无效",

	CodeProductNotFound:  "商品不存在",
	CodeProductOffline:   "商品已下架",
	CodeSkuNotFound:      "SKU不存在",
	CodeCategoryNotFound: "类目不存在",

	CodeOrderNotFound:    "订单不存在",
	CodeOrderStatusError: "订单状态错误",
	CodeOrderCanceled:    "订单已取消",
	CodeOrderPaid:        "订单已支付",

	CodeStockInsufficient: "库存不足",
	CodeStockLocked:       "库存已锁定",

	CodePaymentFailed:  "支付失败",
	CodePaymentExpired: "支付已过期",
	CodeRefundFailed:   "退款失败",
}

// BusinessError 业务错误
type BusinessError struct {
	Code     int    `json:"code"`
	Message  string `json:"message"`
	HTTPCode int    `json:"-"`
}

// Error 实现 error 接口
func (e *BusinessError) Error() string {
	return fmt.Sprintf("code: %d, message: %s", e.Code, e.Message)
}

// NewError 创建业务错误
func NewError(code int, message ...string) *BusinessError {
	msg := errorMessages[code]
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}

	httpCode := http.StatusOK
	switch {
	case code == CodeUnauthorized:
		httpCode = http.StatusUnauthorized
	case code == CodeForbidden:
		httpCode = http.StatusForbidden
	case code == CodeNotFound:
		httpCode = http.StatusNotFound
	case code >= CodeInternalError && code < CodeUserNotFound:
		httpCode = http.StatusInternalServerError
	default:
		httpCode = http.StatusBadRequest
	}

	return &BusinessError{
		Code:     code,
		Message:  msg,
		HTTPCode: httpCode,
	}
}

// NewInternalError 创建内部错误
func NewInternalError(message string) *BusinessError {
	return NewError(CodeInternalError, message)
}

// NewInvalidParamError 创建参数错误
func NewInvalidParamError(message string) *BusinessError {
	return NewError(CodeInvalidParam, message)
}

// NewNotFoundError 创建资源不存在错误
func NewNotFoundError(message string) *BusinessError {
	return NewError(CodeNotFound, message)
}
