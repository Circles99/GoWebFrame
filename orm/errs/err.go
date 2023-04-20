package errs

import (
	"fmt"
)

func NewErrInvalidTagContent(val string) error {
	return fmt.Errorf("xxxx %s", val)
}
