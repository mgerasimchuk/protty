package util

import (
	"fmt"
	"github.com/rwtodd/Go.Sed/sed"
	"strings"
)

// SED replace the input by sed expression
// in the error case returns the original input
func SED(sedExpr string, input []byte) ([]byte, []byte, error) {
	engine, err := sed.New(strings.NewReader(sedExpr))
	if err != nil {
		return input, input, fmt.Errorf("%s: %w", GetFuncName(sed.New), err)
	}
	output, err := engine.RunString(string(input))
	if err != nil {
		return input, input, fmt.Errorf("%s: %w", GetFuncName(engine.RunString), err)
	}
	return []byte(output), input, nil
}
