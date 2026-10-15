package httperr

import (
	"errors"
	"log"

	"github.com/TogetherBros/azit/proto"
	"github.com/gofiber/fiber/v3"
)

type Error struct {
	Status  int
	Code    string
	Message string
}

func (e *Error) Error() string {
	return e.Code + ": " + e.Message
}

func New(status int, code, message string) *Error {
	return &Error{Status: status, Code: code, Message: message}
}

func Unauthorized(message string) *Error {
	return New(fiber.StatusUnauthorized, proto.ErrCodeUnauthorized, message)
}

// Handler 는 핸들러가 돌려준 에러를 명세의 에러 응답으로 바꾼다.
// fiber.New(fiber.Config{ErrorHandler: httperr.Handler}) 로 등록한다.
func Handler(c fiber.Ctx, err error) error {
	// 1. 우리가 만든 에러면 그대로 응답
	var he *Error
	if errors.As(err, &he) {
		return write(c, he.Status, he.Code, he.Message)
	}

	// 2. Fiber 가 만든 에러면 (ErrUnauthorized, ErrBadRequest ...) 번호로 코드를 고른다
	var fe *fiber.Error
	if errors.As(err, &fe) {
		return write(c, fe.Code, codeFor(fe.Code), fe.Message)
	}

	// 3. 모르는 에러 = 서버 문제. 원인은 로그에만 남기고 응답에는 숨긴다
	log.Printf("internal error: %s %s: %v", c.Method(), c.Path(), err)
	return write(c, fiber.StatusInternalServerError, proto.ErrCodeInternal, "잠시 후 다시 시도해 주세요")
}

func write(c fiber.Ctx, status int, code, message string) error {
	return c.Status(status).JSON(proto.ErrorResponse{
		Error: proto.ErrorBody{
			Code:    code,
			Message: message,
		},
	})
}

func codeFor(status int) string {
	switch status {
	case fiber.StatusUnauthorized:
		return proto.ErrCodeUnauthorized
	case fiber.StatusBadRequest:
		return proto.ErrCodeBadRequest
	case fiber.StatusForbidden:
		return proto.ErrCodeForbidden
	case fiber.StatusNotFound:
		return proto.ErrCodeNotFound
	default:
		return proto.ErrCodeInternal
	}
}
