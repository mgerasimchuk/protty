package config

import (
	"protty/pkg/util"
	"strings"
)

type OptionValueType interface {
	int | string | float64
}

type Option[T OptionValueType] struct {
	Name  string
	Value T
}

func (o Option[T]) GetEnvName() string {
	return strings.ToUpper(strings.ReplaceAll(o.GetKebabCaseName(), "-", "_"))
}

func (o Option[T]) GetHeaderName() string {
	return strings.ToUpper("x-protty-" + o.GetKebabCaseName())
}

func (o Option[T]) GetFlagName() string {
	return o.GetKebabCaseName()
}

func (o Option[T]) GetKebabCaseName() string {
	return util.ToKebabCase(o.Name)
}
