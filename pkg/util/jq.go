package util

import (
	"encoding/json"
	"fmt"
	"github.com/itchyny/gojq"
	"reflect"
)

// JQ transform the input by jq expression
// in the error case returns the original input
func JQ(jqExpr string, input []byte) ([]byte, []byte, error) {
	if len(input) == 0 || len(jqExpr) == 0 {
		return input, input, nil
	}

	query, err := gojq.Parse(jqExpr)
	if err != nil {
		return input, input, fmt.Errorf("%s: %w", GetFuncName(gojq.Parse), err)
	}

	var inputObject any
	err = json.Unmarshal(input, &inputObject)
	if err != nil {
		return input, input, fmt.Errorf("%s: %w", GetFuncName(json.Unmarshal), err)
	}

	iter := query.Run(inputObject)
	var transformed any
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok = v.(error); ok {
			return input, input, fmt.Errorf("%s: %w", GetFuncName(iter.Next), err)
		}
		transformed = v
	}
	var transformedJSON []byte
	if reflect.TypeOf(transformed).Kind() == reflect.Map {
		transformedJSON, err = json.Marshal(transformed)
		if err != nil {
			return input, input, fmt.Errorf("%s: %w", GetFuncName(json.Marshal), err)
		}
	} else {
		transformedJSON = []byte(fmt.Sprintf("%v", transformed))
	}

	return transformedJSON, input, nil
}
