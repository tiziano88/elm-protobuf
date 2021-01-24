package elm

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/gogo/protobuf/protoc-gen-gogo/generator"
)

// Type - Basic Elm type, custom type, or type alias
type Type string

var (
	intType    Type = "Int"
	floatType  Type = "Float"
	stringType Type = "String"
	bytesType  Type = "Bytes"
	boolType   Type = "Bool"
)

// VariableName - unique camelcase identifier starting with lowercase letter.
// Used for both constants and function declarations
type VariableName string

// VariantJSONName - unique JSON identifier, uppercase snake case, for a custom type variant
type VariantJSONName string

// ProtobufFieldNumber - unique identifier required for protobuff field declarations
// Used only for commentsin Elm code generation
type ProtobufFieldNumber int32

// DecoderName - decoder function name for Elm type
func DecoderName(t Type) VariableName {
	return VariableName(firstLower(fmt.Sprintf("%sDecoder", t)))
}

// EncoderName - encoder function name for Elm type
func EncoderName(t Type) VariableName {
	return VariableName(firstLower(fmt.Sprintf("%sEncoder", t)))
}

// NestedType - top level Elm type for a possibly nested PB definition
func NestedType(name string, preface []string) Type {
	fullName := name
	for _, p := range preface {
		fullName = fmt.Sprintf("%s_%s", p, fullName)
	}

	return Type(fullName)
}

func camelCase(in string) string {
	// Remove any additional underscores, e.g. convert `foo_1` into `foo1`.
	return strings.Replace(generator.CamelCase(in), "_", "", -1)
}

// TODO: Move these functions to a string package
func firstUpper(in string) string {
	if len(in) < 2 {
		return strings.ToUpper(in)
	}

	return strings.ToUpper(string(in[0])) + string(in[1:])
}

func firstLower(in string) string {
	if len(in) < 2 {
		return strings.ToLower(in)
	}

	return strings.ToLower(string(in[0])) + string(in[1:])
}

// TODO: Rename this function
func convert(inType string) string {
	outMessageSegments := []string{}
	for _, s := range strings.Split(inType, ".") {
		if s == "" {
			continue
		}

		if r, _ := utf8.DecodeRuneInString(s); !unicode.IsLower(r) {
			outMessageSegments = append(outMessageSegments, firstUpper(s))
		}
	}
	return strings.Join(outMessageSegments, "_")
}
