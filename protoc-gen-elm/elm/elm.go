package elm

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/gogo/protobuf/protoc-gen-gogo/generator"
	"google.golang.org/protobuf/types/descriptorpb"
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
	fullName := camelCase(name)
	for _, p := range preface {
		fullName = fmt.Sprintf("%s_%s", p, fullName)
	}

	return Type(firstUpper(fullName))
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

func BasicFieldEncoder(inField *descriptorpb.FieldDescriptorProto) VariableName {
	switch inField.GetType() {
	case descriptorpb.FieldDescriptorProto_TYPE_INT32,
		descriptorpb.FieldDescriptorProto_TYPE_UINT32,
		descriptorpb.FieldDescriptorProto_TYPE_SINT32,
		descriptorpb.FieldDescriptorProto_TYPE_FIXED32,
		descriptorpb.FieldDescriptorProto_TYPE_SFIXED32:
		return "JE.int"
	case descriptorpb.FieldDescriptorProto_TYPE_INT64,
		descriptorpb.FieldDescriptorProto_TYPE_UINT64,
		descriptorpb.FieldDescriptorProto_TYPE_SINT64,
		descriptorpb.FieldDescriptorProto_TYPE_FIXED64,
		descriptorpb.FieldDescriptorProto_TYPE_SFIXED64:
		return "numericStringEncoder"
	case descriptorpb.FieldDescriptorProto_TYPE_FLOAT,
		descriptorpb.FieldDescriptorProto_TYPE_DOUBLE:
		return "JE.float"
	case descriptorpb.FieldDescriptorProto_TYPE_BOOL:
		return "JE.bool"
	case descriptorpb.FieldDescriptorProto_TYPE_STRING:
		return "JE.string"
	case descriptorpb.FieldDescriptorProto_TYPE_ENUM,
		descriptorpb.FieldDescriptorProto_TYPE_MESSAGE:
		if n, ok := WellKnownTypeMap[inField.GetTypeName()]; ok {
			return n.Encoder
		}

		return EncoderName(Type(convert(inField.GetTypeName())))
	case descriptorpb.FieldDescriptorProto_TYPE_BYTES:
		return "bytesFieldEncoder"
	default:
		panic(fmt.Errorf("Error generating decoder for field %s", inField.GetType()))
	}
}

func BasicFieldDecoder(inField *descriptorpb.FieldDescriptorProto) VariableName {
	switch inField.GetType() {
	case descriptorpb.FieldDescriptorProto_TYPE_INT32,
		descriptorpb.FieldDescriptorProto_TYPE_INT64,
		descriptorpb.FieldDescriptorProto_TYPE_UINT32,
		descriptorpb.FieldDescriptorProto_TYPE_UINT64,
		descriptorpb.FieldDescriptorProto_TYPE_SINT32,
		descriptorpb.FieldDescriptorProto_TYPE_SINT64,
		descriptorpb.FieldDescriptorProto_TYPE_FIXED32,
		descriptorpb.FieldDescriptorProto_TYPE_FIXED64,
		descriptorpb.FieldDescriptorProto_TYPE_SFIXED32,
		descriptorpb.FieldDescriptorProto_TYPE_SFIXED64:
		return "intDecoder"
	case descriptorpb.FieldDescriptorProto_TYPE_FLOAT,
		descriptorpb.FieldDescriptorProto_TYPE_DOUBLE:
		return "JD.float"
	case descriptorpb.FieldDescriptorProto_TYPE_BOOL:
		return "JD.bool"
	case descriptorpb.FieldDescriptorProto_TYPE_STRING:
		return "JD.string"
	case descriptorpb.FieldDescriptorProto_TYPE_BYTES:
		return "bytesFieldDecoder"
	case descriptorpb.FieldDescriptorProto_TYPE_ENUM,
		descriptorpb.FieldDescriptorProto_TYPE_MESSAGE:
		if n, ok := WellKnownTypeMap[inField.GetTypeName()]; ok {
			return n.Decoder
		}

		return DecoderName(Type(convert(inField.GetTypeName())))
	default:
		panic(fmt.Errorf("error generating decoder for field %s", inField.GetType()))
	}
}

func BasicFieldType(inField *descriptorpb.FieldDescriptorProto) Type {
	switch inField.GetType() {
	case descriptorpb.FieldDescriptorProto_TYPE_INT32,
		descriptorpb.FieldDescriptorProto_TYPE_INT64,
		descriptorpb.FieldDescriptorProto_TYPE_UINT32,
		descriptorpb.FieldDescriptorProto_TYPE_UINT64,
		descriptorpb.FieldDescriptorProto_TYPE_SINT32,
		descriptorpb.FieldDescriptorProto_TYPE_SINT64,
		descriptorpb.FieldDescriptorProto_TYPE_FIXED32,
		descriptorpb.FieldDescriptorProto_TYPE_FIXED64,
		descriptorpb.FieldDescriptorProto_TYPE_SFIXED32,
		descriptorpb.FieldDescriptorProto_TYPE_SFIXED64:
		return intType
	case descriptorpb.FieldDescriptorProto_TYPE_FLOAT,
		descriptorpb.FieldDescriptorProto_TYPE_DOUBLE:
		return floatType
	case descriptorpb.FieldDescriptorProto_TYPE_BOOL:
		return boolType
	case descriptorpb.FieldDescriptorProto_TYPE_STRING:
		return stringType
	case descriptorpb.FieldDescriptorProto_TYPE_BYTES:
		return bytesType
	case descriptorpb.FieldDescriptorProto_TYPE_ENUM,
		descriptorpb.FieldDescriptorProto_TYPE_MESSAGE:
		if n, ok := WellKnownTypeMap[inField.GetTypeName()]; ok {
			return n.Type
		}
		return Type(convert(inField.GetTypeName()))
	default:
		panic(fmt.Errorf("Error generating type for field %q %s", inField.GetName(), inField.GetType()))
	}
}

type DefaultValue string

func BasicFieldDefaultValue(inField *descriptorpb.FieldDescriptorProto) DefaultValue {
	if inField.GetLabel() == descriptorpb.FieldDescriptorProto_LABEL_REPEATED {
		return "[]"
	}

	switch inField.GetType() {
	case descriptorpb.FieldDescriptorProto_TYPE_INT32,
		descriptorpb.FieldDescriptorProto_TYPE_INT64,
		descriptorpb.FieldDescriptorProto_TYPE_UINT32,
		descriptorpb.FieldDescriptorProto_TYPE_UINT64,
		descriptorpb.FieldDescriptorProto_TYPE_SINT32,
		descriptorpb.FieldDescriptorProto_TYPE_SINT64,
		descriptorpb.FieldDescriptorProto_TYPE_FIXED32,
		descriptorpb.FieldDescriptorProto_TYPE_FIXED64,
		descriptorpb.FieldDescriptorProto_TYPE_SFIXED32,
		descriptorpb.FieldDescriptorProto_TYPE_SFIXED64:
		return "0"
	case descriptorpb.FieldDescriptorProto_TYPE_FLOAT,
		descriptorpb.FieldDescriptorProto_TYPE_DOUBLE:
		return "0.0"
	case descriptorpb.FieldDescriptorProto_TYPE_BOOL:
		return "False"
	case descriptorpb.FieldDescriptorProto_TYPE_STRING:
		return "\"\""
	case descriptorpb.FieldDescriptorProto_TYPE_MESSAGE:

		// TODO: What is this?
		return "xxx"
	case descriptorpb.FieldDescriptorProto_TYPE_BYTES:
		return "[]"
	case descriptorpb.FieldDescriptorProto_TYPE_ENUM:
		return DefaultValue(EnumDefaultVariantVariableName(
			convert(inField.GetTypeName()),
			[]string{},
		))
	default:
		panic(fmt.Errorf("error - no known default value for field %s", inField.GetType()))
	}
}
