package elm

import (
	"fmt"
	"strings"

	"github.com/gogo/protobuf/protoc-gen-gogo/generator"
)

// VariableName - unique camelcase identifier starting with lowercase letter.
// Used for both constants and function declarations
type VariableName string

// ProtobufFieldNumber - unique identifier required for protobuff field declarations
// Used only for commentsin Elm code generation
type ProtobufFieldNumber int32

// DecoderName - decoder function name for Elm variable
func DecoderName(name VariableName) VariableName {
	return VariableName(firstLower(fmt.Sprintf("%sDecoder", name)))
}

// EncoderName - encoder function name for Elm variable
func EncoderName(name VariableName) VariableName {
	return VariableName(firstLower(fmt.Sprintf("%sEncoder", name)))
}

func camelCase(in string) string {
	// Remove any additional underscores, e.g. convert `foo_1` into `foo1`.
	return strings.Replace(generator.CamelCase(in), "_", "", -1)
}

func firstLower(in string) string {
	if len(in) < 2 {
		return strings.ToLower(in)
	}

	return strings.ToLower(string(in[0])) + string(in[1:])
}
