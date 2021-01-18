package elm

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/types/descriptorpb"
)

// CustomType - defines an Elm custom type (sometimes called union type)
// The default value is the variant with
// https://guide.elm-lang.org/types/custom_types.html
type CustomType struct {
	Name                   VariableName
	Decoder                VariableName
	Encoder                VariableName
	DefaultVariantVariable VariableName
	DefaultVariantValue    VariantName
	Variants               []CustomTypeVariant
}

// VariantName - unique camelcase identifier used for custom type variants
// https://guide.elm-lang.org/types/custom_types.html
type VariantName string

// VariantJSONName - unique JSON identifier, uppercase snake case, for a custom type variant
type VariantJSONName string

// CustomTypeVariant - a possible variant of a CustomType
// https://guide.elm-lang.org/types/custom_types.html
type CustomTypeVariant struct {
	Name     VariantName
	Number   ProtobufFieldNumber
	JSONName VariantJSONName
}

// NestedVariableName - top level Elm variable name for a possibly nested PB definition
func NestedVariableName(name string, preface []string) VariableName {
	fullName := name
	for _, p := range preface {
		fullName = fmt.Sprintf("%s_%s", p, fullName)
	}

	return VariableName(fullName)
}

// NestedVariantName - Elm variant name for a possibly nested PB definition
func NestedVariantName(name string, preface []string) VariantName {
	fullName := camelCase(strings.ToLower(name))
	for _, p := range preface {
		fullName = fmt.Sprintf("%s_%s", camelCase(strings.ToLower(p)), fullName)
	}

	return VariantName(fullName)
}

// DefaultVariantVariableName - convenient identifier for a custom types default variant
func DefaultVariantVariableName(name string, preface []string) VariableName {
	variableName := NestedVariableName(name, preface)
	return VariableName(firstLower(fmt.Sprintf("%sDefault", variableName)))
}

// GetVariantJSONName - JSON identifier for variant decoder/encoding
func GetVariantJSONName(pb *descriptorpb.EnumValueDescriptorProto) VariantJSONName {
	return VariantJSONName(pb.GetName())
}
