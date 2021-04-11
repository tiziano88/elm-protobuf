package elm

import (
	"fmt"
	"protoc-gen-elm/pkg/stringextras"
	"text/template"

	"google.golang.org/protobuf/types/descriptorpb"
)

// WellKnownType - information to handle Google well known types
type WellKnownType struct {
	Type    Type
	Encoder VariableName
	Decoder VariableName
}

var (
	// WellKnownTypeMap - map of Google well known type PB identifier to encoder/decoder info
	WellKnownTypeMap = map[string]WellKnownType{
		".google.protobuf.Timestamp": {
			Type:    "Timestamp",
			Decoder: "timestampDecoder",
			Encoder: "timestampEncoder",
		},
		".google.protobuf.Int32Value": {
			Type:    intType,
			Decoder: "intValueDecoder",
			Encoder: "intValueEncoder",
		},
		".google.protobuf.Int64Value": {
			Type:    intType,
			Decoder: "intValueDecoder",
			Encoder: "numericStringEncoder",
		},
		".google.protobuf.UInt32Value": {
			Type:    intType,
			Decoder: "intValueDecoder",
			Encoder: "intValueEncoder",
		},
		".google.protobuf.UInt64Value": {
			Type:    intType,
			Decoder: "intValueDecoder",
			Encoder: "numericStringEncoder",
		},
		".google.protobuf.DoubleValue": {
			Type:    floatType,
			Decoder: "floatValueDecoder",
			Encoder: "floatValueEncoder",
		},
		".google.protobuf.FloatValue": {
			Type:    floatType,
			Decoder: "floatValueDecoder",
			Encoder: "floatValueEncoder",
		},
		".google.protobuf.StringValue": {
			Type:    stringType,
			Decoder: "stringValueDecoder",
			Encoder: "stringValueEncoder",
		},
		".google.protobuf.BytesValue": {
			Type:    bytesType,
			Decoder: "bytesValueDecoder",
			Encoder: "bytesValueEncoder",
		},
		".google.protobuf.BoolValue": {
			Type:    boolType,
			Decoder: "boolValueDecoder",
			Encoder: "boolValueEncoder",
		},
	}

	reservedKeywords = map[string]bool{
		"module":   true,
		"exposing": true,
		"import":   true,
		"type":     true,
		"let":      true,
		"in":       true,
		"if":       true,
		"then":     true,
		"else":     true,
		"where":    true,
		"case":     true,
		"of":       true,
		"port":     true,
		"as":       true,
	}
)

// TypeAlias - defines an Elm type alias (somtimes called a record)
// https://guide.elm-lang.org/types/type_aliases.html
type TypeAlias struct {
	Name    Type
	Decoder VariableName
	Encoder VariableName
	Fields  []TypeAliasField
}

// FieldDecoder used in type alias decdoer (ex. )
type FieldDecoder string

// FieldEncoder used in type alias decdoer (ex. )
type FieldEncoder string

// TypeAliasField - type alias field definition
type TypeAliasField struct {
	Name    VariableName
	Type    Type
	Number  ProtobufFieldNumber
	Decoder FieldDecoder
	Encoder FieldEncoder
}

func appendUnderscoreToReservedKeywords(in string) string {
	if reservedKeywords[in] {
		return fmt.Sprintf("%s_", in)
	}

	return in
}

// FieldName - simple camelcase variable name with first letter lower
func FieldName(in string) VariableName {
	return VariableName(appendUnderscoreToReservedKeywords(stringextras.LowerCamelCase(in)))
}

// FieldJSONName - JSON identifier for field decoder/encoding
func FieldJSONName(pb *descriptorpb.FieldDescriptorProto) VariantJSONName {
	return VariantJSONName(pb.GetJsonName())
}

func RequiredFieldEncoder(pb *descriptorpb.FieldDescriptorProto) FieldEncoder {
	return FieldEncoder(fmt.Sprintf(
		"requiredFieldEncoder \"%s\" %s %s v.%s",
		FieldJSONName(pb),
		BasicFieldEncoder(pb),
		BasicFieldDefaultValue(pb),
		FieldName(pb.GetName()),
	))
}

func RequiredFieldDecoder(pb *descriptorpb.FieldDescriptorProto) FieldDecoder {
	return FieldDecoder(fmt.Sprintf(
		"required \"%s\" %s %s",
		FieldJSONName(pb),
		BasicFieldDecoder(pb),
		BasicFieldDefaultValue(pb),
	))
}

func OneOfEncoder(pb *descriptorpb.OneofDescriptorProto) FieldEncoder {
	return FieldEncoder(fmt.Sprintf("%s v.%s",
		EncoderName(Type(stringextras.CamelCase(pb.GetName()))),
		FieldName(pb.GetName()),
	))
}

func OneOfDecoder(pb *descriptorpb.OneofDescriptorProto) FieldDecoder {
	return FieldDecoder(fmt.Sprintf(
		"field %s",
		DecoderName(Type(stringextras.CamelCase(pb.GetName()))),
	))
}

func MapType(messagePb *descriptorpb.DescriptorProto) Type {
	keyField := messagePb.GetField()[0]
	valueField := messagePb.GetField()[1]

	return Type(fmt.Sprintf(
		"Dict.Dict %s %s",
		BasicFieldType(keyField),
		BasicFieldType(valueField),
	))
}

func MapEncoder(
	fieldPb *descriptorpb.FieldDescriptorProto,
	messagePb *descriptorpb.DescriptorProto,
) FieldEncoder {
	valueField := messagePb.GetField()[1]

	return FieldEncoder(fmt.Sprintf(
		"mapEntriesFieldEncoder \"%s\" %s v.%s",
		FieldJSONName(fieldPb),
		BasicFieldEncoder(valueField),
		FieldName(fieldPb.GetName()),
	))
}

func MapDecoder(
	fieldPb *descriptorpb.FieldDescriptorProto,
	messagePb *descriptorpb.DescriptorProto,
) FieldDecoder {
	valueField := messagePb.GetField()[1]

	return FieldDecoder(fmt.Sprintf(
		"mapEntries \"%s\" %s",
		FieldJSONName(fieldPb),
		BasicFieldDecoder(valueField),
	))
}

func MaybeType(t Type) Type {
	return Type(fmt.Sprintf("Maybe %s", t))
}

func MaybeEncoder(pb *descriptorpb.FieldDescriptorProto) FieldEncoder {
	return FieldEncoder(fmt.Sprintf(
		"optionalEncoder \"%s\" %s v.%s",
		FieldJSONName(pb),
		BasicFieldEncoder(pb),
		FieldName(pb.GetName()),
	))
}

func MaybeDecoder(pb *descriptorpb.FieldDescriptorProto) FieldDecoder {
	return FieldDecoder(fmt.Sprintf(
		"optional \"%s\" %s",
		FieldJSONName(pb),
		BasicFieldDecoder(pb),
	))
}

func ListType(t Type) Type {
	return Type(fmt.Sprintf("List %s", t))
}

func ListEncoder(pb *descriptorpb.FieldDescriptorProto) FieldEncoder {
	return FieldEncoder(fmt.Sprintf(
		"repeatedFieldEncoder \"%s\" %s v.%s",
		FieldJSONName(pb),
		BasicFieldEncoder(pb),
		FieldName(pb.GetName()),
	))
}

func ListDecoder(pb *descriptorpb.FieldDescriptorProto) FieldDecoder {
	return FieldDecoder(fmt.Sprintf(
		"repeated \"%s\" %s",
		FieldJSONName(pb),
		BasicFieldDecoder(pb),
	))
}

func OneOfType(in string) Type {
	return Type(appendUnderscoreToReservedKeywords(stringextras.UpperCamelCase(in)))
}

// TypeAliasTemplate - defines templates for self contained type aliases
func TypeAliasTemplate(t *template.Template) (*template.Template, error) {
	return t.Parse(`
{{- define "type-alias" -}}
type alias {{ .Name }} =
    { {{ range $i, $v := .Fields }}
        {{- if $i }}, {{ end }}{{ .Name }} : {{ .Type }}{{ if .Number }} -- {{ .Number }}{{ end }}
    {{ end }}}


{{ .Decoder }} : JD.Decoder {{ .Name }}
{{ .Decoder }} =
    JD.lazy <| \_ -> decode {{ .Name }}{{ range .Fields }}
        |> {{ .Decoder }}{{ end }}


{{ .Encoder }} : {{ .Name }} -> JE.Value
{{ .Encoder }} v =
    JE.object <| List.filterMap identity <|
        [{{ range $i, $v := .Fields }}
            {{- if $i }},{{ end }} ({{ .Encoder }})
        {{ end }}]
{{- end -}}
`)
}
