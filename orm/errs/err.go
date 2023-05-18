package errs

import (
	"errors"
	"fmt"
)

var (
	ErrInsertZeroRow = errors.New("insert zero row")
)

func NewErrInvalidTagContent(val string) error {
	return fmt.Errorf("xxxx %s", val)
}

func NewErrUnknownField(val string) error {
	return fmt.Errorf("unkown field %s", val)
}

func NewErrUnsupportedAssignableType(exp any) error {
	return fmt.Errorf("orm: 不支持的 Assignable 表达式 %v", exp)
}
