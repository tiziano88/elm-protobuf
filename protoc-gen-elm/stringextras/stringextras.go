package stringextras

import (
	"strings"

	"github.com/gogo/protobuf/protoc-gen-gogo/generator"
)

func UpperCamelCase(in string) string {
	return FirstUpper(CamelCase(in))
}

func LowerCamelCase(in string) string {
	return FirstLower(CamelCase(in))
}

func CamelCase(in string) string {
	// Remove any additional underscores, e.g. convert `foo_1` into `foo1`.
	return strings.Replace(generator.CamelCase(in), "_", "", -1)
}

func FirstUpper(in string) string {
	if len(in) < 2 {
		return strings.ToUpper(in)
	}

	return strings.ToUpper(string(in[0])) + string(in[1:])
}

func FirstLower(in string) string {
	if len(in) < 2 {
		return strings.ToLower(in)
	}

	return strings.ToLower(string(in[0])) + string(in[1:])
}
