package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"protoc-gen-elm/elm"
	"strings"
	"text/template"
	"unicode"
	"unicode/utf8"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/generator"
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/types/descriptorpb"
)

var (
	// Well Known Types.
	excludedFiles = map[string]bool{
		"google/protobuf/timestamp.proto": true,
		"google/protobuf/wrappers.proto":  true,
	}
	excludedTypes = map[string]string{
		".google.protobuf.Timestamp":   "Timestamp",
		".google.protobuf.Int32Value":  "Int",
		".google.protobuf.Int64Value":  "Int",
		".google.protobuf.UInt32Value": "Int",
		".google.protobuf.UInt64Value": "Int",
		".google.protobuf.DoubleValue": "Float",
		".google.protobuf.FloatValue":  "Float",
		".google.protobuf.StringValue": "String",
		".google.protobuf.BytesValue":  "Bytes",
		".google.protobuf.BoolValue":   "Bool",
	}
	excludedDecoders = map[string]DecoderName{
		".google.protobuf.Timestamp":   "timestampDecoder",
		".google.protobuf.Int32Value":  "intValueDecoder",
		".google.protobuf.Int64Value":  "intValueDecoder",
		".google.protobuf.UInt32Value": "intValueDecoder",
		".google.protobuf.UInt64Value": "intValueDecoder",
		".google.protobuf.DoubleValue": "floatValueDecoder",
		".google.protobuf.FloatValue":  "floatValueDecoder",
		".google.protobuf.StringValue": "stringValueDecoder",
		".google.protobuf.BytesValue":  "bytesValueDecoder",
		".google.protobuf.BoolValue":   "boolValueDecoder",
	}
	excludedEncoders = map[string]string{
		".google.protobuf.Timestamp":   "timestampEncoder",
		".google.protobuf.Int32Value":  "intValueEncoder",
		".google.protobuf.Int64Value":  "numericStringEncoder",
		".google.protobuf.UInt32Value": "intValueEncoder",
		".google.protobuf.UInt64Value": "numericStringEncoder",
		".google.protobuf.DoubleValue": "floatValueEncoder",
		".google.protobuf.FloatValue":  "floatValueEncoder",
		".google.protobuf.StringValue": "stringValueEncoder",
		".google.protobuf.BytesValue":  "bytesValueEncoder",
		".google.protobuf.BoolValue":   "boolValueEncoder",
	}

	// Avoid collisions with reserved keywords by appending a single underscore after the name.
	// This does not guarantee that collisions are avoided, but makes them less likely to
	// happen.
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

type parameters struct {
	Debug            bool
	RemoveDeprecated bool
}

func parseParameters(input *string) (parameters, error) {
	var result parameters
	var err error

	if input == nil {
		return result, nil
	}

	for _, i := range strings.Split(*input, ",") {
		switch i {
		case "remove-deprecated":
			result.RemoveDeprecated = true
		case "debug":
			result.Debug = true
		default:
			err = fmt.Errorf("Unknown parameter: \"%s\"", i)
		}
	}

	return result, err
}

func main() {
	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatalf("Could not read request from STDIN: %v", err)
	}

	req := &plugin.CodeGeneratorRequest{}

	err = proto.Unmarshal(data, req)
	if err != nil {
		log.Fatalf("Could not unmarshal request: %v", err)
	}

	parameters, err := parseParameters(req.Parameter)
	if err != nil {
		log.Fatalf("Failed to parse parameters: %v", err)
	}

	if parameters.Debug {
		// Remove useless source code data.
		for _, inFile := range req.GetProtoFile() {
			inFile.SourceCodeInfo = nil
		}

		log.Printf("Input data: %v", proto.MarshalTextString(req))
	}

	resp := &plugin.CodeGeneratorResponse{}

	for _, inFile := range req.GetProtoFile() {
		log.Printf("Processing file %s", inFile.GetName())
		// Well Known Types.
		if excludedFiles[inFile.GetName()] {
			log.Printf("Skipping well known type")
			continue
		}

		name := fileName(inFile.GetName())
		content, err := templateFile(inFile, parameters)
		if err != nil {
			log.Fatalf("Could not template file: %v", err)
		}

		resp.File = append(resp.File, &plugin.CodeGeneratorResponse_File{
			Name:    &name,
			Content: &content,
		})
	}

	data, err = proto.Marshal(resp)
	if err != nil {
		log.Fatalf("Could not marshal response: %v [%v]", err, resp)
	}

	_, err = os.Stdout.Write(data)
	if err != nil {
		log.Fatalf("Could not write response to STDOUT: %v", err)
	}
}

func hasMapEntries(inFile *descriptorpb.FileDescriptorProto) bool {
	for _, m := range inFile.GetMessageType() {
		if hasMapEntriesInMessage(m) {
			return true
		}
	}

	return false
}

func hasMapEntriesInMessage(inMessage *descriptorpb.DescriptorProto) bool {
	if inMessage.GetOptions().GetMapEntry() {
		return true
	}

	for _, m := range inMessage.GetNestedType() {
		if hasMapEntriesInMessage(m) {
			return true
		}
	}

	return false
}

func templateFile(inFile *descriptorpb.FileDescriptorProto, p parameters) (string, error) {
	t := template.New("t")

	t, err := elm.CustomTypeTemplate(t)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse custom type template")
	}

	t, err = elm.TypeAliasTemplate(t)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse type alias template")
	}

	t, err = t.Parse(`
{{- define "oneof-type" -}}
type {{ .Name }}
    = {{ .Name }}Unspecified
{{- range .Fields }}
    | {{ .Type }} {{ .ElmType }}
{{- end }}


{{ .DecoderName }} : JD.Decoder {{ .Name }}
{{ .DecoderName }} =
    JD.lazy <| \_ -> JD.oneOf
        [{{ range $i, $v := .Fields }}{{ if $i }},{{ end }} JD.map {{ .Type }} (JD.field "{{ .JSONName }}" {{ .Decoder }})
        {{ end }}, JD.succeed {{ .Name }}Unspecified
        ]


{{ .EncoderName }} : {{ .Name }} -> Maybe ( String, JE.Value )
{{ .EncoderName }} v =
    case v of
        {{ .Name }}Unspecified ->
            Nothing
        {{- range .Fields }}
        {{ .Type }} x ->
            Just ( "{{ .JSONName }}", {{ .Encoder }} x )
        {{- end }}
{{- end -}}
`)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse one-of template")
	}

	t, err = t.Parse(`
{{- define "nested-message" -}}
{{ template "type-alias" .TypeAlias }}
{{- range .OneOfs }}


{{ template "oneof-type" . }}
{{- end }}
{{- range .NestedCustomTypes }}


{{ template "custom-type" . }}
{{- end }}
{{- range .NestedMessages }}


{{ template "nested-message" . }}
{{- end }}
{{- end -}}
`)

	if err != nil {
		return "", errors.Wrap(err, "failed to parse nested PB message template")
	}

	t, err = t.Parse(`module {{ .ModuleName }} exposing (..)

-- DO NOT EDIT
-- AUTOGENERATED BY THE ELM PROTOCOL BUFFER COMPILER
-- https://github.com/tiziano88/elm-protobuf
-- source file: {{ .SourceFile }}

import Protobuf exposing (..)

import Json.Decode as JD
import Json.Encode as JE
{{- if .ImportDict }}
import Dict
{{- end }}
{{- range .AdditionalImports }}
import {{ . }} exposing (..)
{{ end }}


uselessDeclarationToPreventErrorDueToEmptyOutputFile = 42
{{- range .TopEnums }}


{{ template "custom-type" . }}
{{- end }}
{{- range .Messages }}


{{ template "nested-message" . }}
{{- end }}
`)
	if err != nil {
		return "", err
	}

	messages, err := messages([]string{}, inFile.GetMessageType(), p)
	if err != nil {
		return "", err
	}

	buff := &bytes.Buffer{}
	if err = t.Execute(buff, struct {
		SourceFile        string
		ModuleName        string
		ImportDict        bool
		AdditionalImports []string
		TopEnums          []elm.CustomType
		Messages          []pbMessage
	}{
		SourceFile:        inFile.GetName(),
		ModuleName:        moduleName(inFile.GetName()),
		ImportDict:        hasMapEntries(inFile),
		AdditionalImports: getAdditionalImports(inFile.GetDependency()),
		TopEnums:          enumsToCustomTypes([]string{}, inFile.GetEnumType(), p),
		Messages:          messages,
	}); err != nil {
		return "", err
	}

	return buff.String(), nil
}

// CustomElmType is the type alias for custom message, enum, and oneof elements.
// This value is camel case with the first letter capitalized.
type CustomElmType string

// DecoderName is the Elm decoder for the FieldDescriptorProto_TYPE_****.
// Simple custom decoders or prefaced with JD, the alias for `Json.Decode`.
type DecoderName string

const (
	IntDecoder    DecoderName = "intDecoder"
	BytesDecoder  DecoderName = "bytesFieldDecoder"
	FloatDecoder  DecoderName = "JD.float"
	BoolDecoder   DecoderName = "JD.bool"
	StringDecoder DecoderName = "JD.string"
)

// TODO: Convert PB OneOf to an Elm custom type struct
//       Differences in decoders/encoders will be an issue
type oneOf struct {
	Name        string
	DecoderName DecoderName
	EncoderName string
	Fields      []oneOfField
}

type oneOfField struct {
	JSONName string
	Type     CustomElmType
	ElmType  string
	Decoder  DecoderName
	Encoder  string
}

type pbMessage struct {
	TypeAlias         elm.TypeAlias
	OneOfs            []oneOf
	NestedCustomTypes []elm.CustomType
	NestedMessages    []pbMessage
}

func isDeprecated(options interface{}) bool {
	switch v := options.(type) {
	case *descriptorpb.MessageOptions:
		return v != nil && v.Deprecated != nil && *v.Deprecated
	case *descriptorpb.FieldOptions:
		return v != nil && v.Deprecated != nil && *v.Deprecated
	case *descriptorpb.EnumOptions:
		return v != nil && v.Deprecated != nil && *v.Deprecated
	case *descriptorpb.EnumValueOptions:
		return v != nil && v.Deprecated != nil && *v.Deprecated
	default:
		return false
	}
}

func enumsToCustomTypes(preface []string, enumPbs []*descriptorpb.EnumDescriptorProto, p parameters) []elm.CustomType {
	var result []elm.CustomType
	for _, enumPb := range enumPbs {
		if isDeprecated(enumPb.Options) && p.RemoveDeprecated {
			continue
		}

		var values []elm.CustomTypeVariant
		for _, value := range enumPb.GetValue() {
			if isDeprecated(value.Options) && p.RemoveDeprecated {
				continue
			}

			values = append(values, elm.CustomTypeVariant{
				Name:     elm.NestedVariantName(value.GetName(), preface),
				Number:   elm.ProtobufFieldNumber(value.GetNumber()),
				JSONName: elm.EnumVariantJSONName(value),
			})
		}

		enumType := elm.NestedType(enumPb.GetName(), preface)

		result = append(result, elm.CustomType{
			Name:                   enumType,
			Decoder:                elm.DecoderName(enumType),
			Encoder:                elm.EncoderName(enumType),
			DefaultVariantVariable: elm.EnumDefaultVariantVariableName(enumPb.GetName(), preface),
			DefaultVariantValue:    values[0].Name,
			Variants:               values,
		})
	}

	return result
}

func messages(preface []string, messagePbs []*descriptorpb.DescriptorProto, p parameters) ([]pbMessage, error) {
	var result []pbMessage
	for _, messagePb := range messagePbs {

		if isDeprecated(messagePb.Options) && p.RemoveDeprecated {
			continue
		}

		var newFields []elm.TypeAliasField
		for _, fieldPb := range messagePb.GetField() {
			if isDeprecated(fieldPb.Options) && p.RemoveDeprecated {
				continue
			}

			if fieldPb.OneofIndex != nil {
				continue
			}

			nested := getNestedType(fieldPb, messagePb)
			if nested != nil {
				newFields = append(newFields, elm.TypeAliasField{
					Name:    elm.FieldName(fieldPb.GetName()),
					Type:    elm.MapType(nested),
					Number:  elm.ProtobufFieldNumber(fieldPb.GetNumber()),
					Encoder: elm.MapEncoder(fieldPb, nested),
					Decoder: elm.MapDecoder(fieldPb, nested),
				})
			} else if isOptional(fieldPb) {
				newFields = append(newFields, elm.TypeAliasField{
					Name:    elm.FieldName(fieldPb.GetName()),
					Type:    elm.MaybeType(elm.BasicFieldType(fieldPb)),
					Number:  elm.ProtobufFieldNumber(fieldPb.GetNumber()),
					Encoder: elm.MaybeEncoder(fieldPb),
					Decoder: elm.MaybeDecoder(fieldPb),
				})
			} else if isRepeated(fieldPb) {
				newFields = append(newFields, elm.TypeAliasField{
					Name:    elm.FieldName(fieldPb.GetName()),
					Type:    elm.ListType(elm.BasicFieldType(fieldPb)),
					Number:  elm.ProtobufFieldNumber(fieldPb.GetNumber()),
					Encoder: elm.ListEncoder(fieldPb),
					Decoder: elm.ListDecoder(fieldPb),
				})
			} else {
				newFields = append(newFields, elm.TypeAliasField{
					Name:    elm.FieldName(fieldPb.GetName()),
					Type:    elm.BasicFieldType(fieldPb),
					Number:  elm.ProtobufFieldNumber(fieldPb.GetNumber()),
					Encoder: elm.RequiredFieldEncoder(fieldPb),
					Decoder: elm.RequiredFieldDecoder(fieldPb),
				})
			}
		}

		for _, oneOfPb := range messagePb.GetOneofDecl() {
			newFields = append(newFields, elm.TypeAliasField{
				Name:    elm.FieldName(oneOfPb.GetName()),
				Type:    elm.Type(reserveWordProtectedCamelCase(oneOfPb.GetName())),
				Encoder: elm.OneOfEncoder(oneOfPb),
				Decoder: elm.OneOfDecoder(oneOfPb),
			})
		}

		var oneOfs []oneOf
		for oneofIndex, oneOfPb := range messagePb.GetOneofDecl() {
			var oneOfFields []oneOfField
			for _, inField := range messagePb.GetField() {
				if inField.OneofIndex == nil || inField.GetOneofIndex() != int32(oneofIndex) {
					continue
				}

				fieldDecoder, err := fieldDecoderName(inField)
				if err != nil {
					return nil, err
				}

				oneOfElmType, err := fieldElmType(inField)
				if err != nil {
					return nil, err
				}

				oneOfFields = append(oneOfFields, oneOfField{
					JSONName: inField.GetJsonName(),
					Type:     customElmType(preface, inField.GetName()),
					ElmType:  oneOfElmType,
					Decoder:  fieldDecoder,
					Encoder:  fieldEncoderName(inField),
				})
			}

			oneOfs = append(oneOfs, oneOf{
				Name:        reserveWordProtectedCamelCase(oneOfPb.GetName()),
				DecoderName: oneofDecoderName(oneOfPb),
				EncoderName: oneofEncoderName(oneOfPb),
				Fields:      oneOfFields,
			})
		}

		fullName := messagePb.GetName()
		for _, p := range preface {
			fullName = fmt.Sprintf("%s_%s", p, fullName)
		}

		newPreface := append([]string{messagePb.GetName()}, preface...)
		nestedMessages, err := messages(newPreface, messagePb.GetNestedType(), p)
		if err != nil {
			return nil, err
		}

		name := elm.NestedType(messagePb.GetName(), preface)
		result = append(result, pbMessage{
			TypeAlias: elm.TypeAlias{
				Name:    name,
				Decoder: elm.DecoderName(name),
				Encoder: elm.EncoderName(name),
				Fields:  newFields,
			},
			OneOfs:            oneOfs,
			NestedCustomTypes: enumsToCustomTypes(newPreface, messagePb.GetEnumType(), p),
			NestedMessages:    nestedMessages,
		})
	}

	return result, nil
}

func isOptional(inField *descriptorpb.FieldDescriptorProto) bool {
	return inField.GetLabel() == descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL &&
		inField.GetType() == descriptorpb.FieldDescriptorProto_TYPE_MESSAGE
}

func isRepeated(inField *descriptorpb.FieldDescriptorProto) bool {
	return inField.GetLabel() == descriptorpb.FieldDescriptorProto_LABEL_REPEATED
}

func getNestedType(inField *descriptorpb.FieldDescriptorProto, inMessage *descriptorpb.DescriptorProto) *descriptorpb.DescriptorProto {

	// TODO: Is there a better way?  Convert to a function with descriptive name.
	fullyQualifiedTypeName := inField.GetTypeName()
	splitName := strings.Split(fullyQualifiedTypeName, ".")
	localTypeName := splitName[len(splitName)-1]

	for _, nested := range inMessage.GetNestedType() {
		if nested.GetName() == localTypeName && nested.GetOptions().GetMapEntry() {
			return nested
		}
	}

	return nil
}

func fileName(inFilePath string) string {
	inFileDir, inFileName := filepath.Split(inFilePath)

	trimmed := strings.TrimSuffix(inFileName, ".proto")
	shortFileName := firstUpper(trimmed)

	fullFileName := ""
	for _, segment := range strings.Split(inFileDir, "/") {
		if segment == "" {
			continue
		}

		fullFileName += firstUpper(segment) + "/"
	}

	return fullFileName + shortFileName + ".elm"
}

func moduleName(inFilePath string) string {
	inFileDir, inFileName := filepath.Split(inFilePath)

	trimmed := strings.TrimSuffix(inFileName, ".proto")
	shortModuleName := firstUpper(trimmed)

	fullModuleName := ""
	for _, segment := range strings.Split(inFileDir, "/") {
		if segment == "" {
			continue
		}

		fullModuleName += firstUpper(segment) + "."
	}

	return fullModuleName + shortModuleName
}

func getAdditionalImports(dependencies []string) []string {
	var additions []string
	for _, d := range dependencies {
		if excludedFiles[d] {
			continue
		}

		fullModuleName := ""
		for _, segment := range strings.Split(strings.TrimSuffix(d, ".proto"), "/") {
			if segment == "" {
				continue
			}
			fullModuleName += firstUpper(segment) + "."
		}

		additions = append(additions, strings.TrimSuffix(fullModuleName, "."))
	}
	return additions
}

func customElmType(preface []string, in string) CustomElmType {
	fullType := camelCase(in)
	for _, p := range preface {
		fullType = fmt.Sprintf("%s_%s", camelCase(p), fullType)
	}

	return CustomElmType(appendUnderscoreToReservedKeywords(fullType))
}

func reserveWordProtectedCamelCase(in string) string {
	return appendUnderscoreToReservedKeywords(camelCase(in))
}

func appendUnderscoreToReservedKeywords(in string) string {
	if reservedKeywords[in] {
		return fmt.Sprintf("%s_", in)
	}

	return in
}

func encoderName(typeName string) string {
	return firstLower(typeName) + "Encoder"
}

func fieldDecoderName(inField *descriptorpb.FieldDescriptorProto) (DecoderName, error) {
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
		return IntDecoder, nil
	case descriptorpb.FieldDescriptorProto_TYPE_FLOAT,
		descriptorpb.FieldDescriptorProto_TYPE_DOUBLE:
		return FloatDecoder, nil
	case descriptorpb.FieldDescriptorProto_TYPE_BOOL:
		return BoolDecoder, nil
	case descriptorpb.FieldDescriptorProto_TYPE_STRING:
		return StringDecoder, nil

	// TODO: This is an unsupported stub (throw error)
	case descriptorpb.FieldDescriptorProto_TYPE_BYTES:
		return BytesDecoder, nil
	case descriptorpb.FieldDescriptorProto_TYPE_ENUM:
		return enumDecoderName(inField), nil
	case descriptorpb.FieldDescriptorProto_TYPE_MESSAGE:
		return messageDecoderName(inField), nil
	default:
		return "", fmt.Errorf("error generating decoder for field %s", inField.GetType())
	}
}

func enumDecoderName(inField *descriptorpb.FieldDescriptorProto) DecoderName {
	_, messageName := convert(inField.GetTypeName())
	return decoderName(messageName)
}

func messageDecoderName(inField *descriptorpb.FieldDescriptorProto) DecoderName {
	// Well Known Types.
	if n, ok := excludedDecoders[inField.GetTypeName()]; ok {
		return n
	}

	_, messageName := convert(inField.GetTypeName())
	return decoderName(messageName)
}

func oneofDecoderName(inOneof *descriptorpb.OneofDescriptorProto) DecoderName {
	typeName := reserveWordProtectedCamelCase(inOneof.GetName())
	return decoderName(typeName)
}

func decoderName(typeName string) DecoderName {
	return DecoderName(firstLower(typeName) + "Decoder")
}

// Returns package name and message name.
func convert(inType string) (string, string) {
	segments := strings.Split(inType, ".")
	outPackageSegments := []string{}
	outMessageSegments := []string{}
	for _, s := range segments {
		if s == "" {
			continue
		}
		r, _ := utf8.DecodeRuneInString(s)
		if unicode.IsLower(r) {
			// Package name.
			outPackageSegments = append(outPackageSegments, firstUpper(s))
		} else {
			// Message name.
			outMessageSegments = append(outMessageSegments, firstUpper(s))
		}
	}
	return strings.Join(outPackageSegments, "."), strings.Join(outMessageSegments, "_")
}

func firstLower(in string) string {
	if len(in) < 2 {
		return strings.ToLower(in)
	}

	return strings.ToLower(string(in[0])) + string(in[1:])
}

func firstUpper(in string) string {
	if len(in) < 2 {
		return strings.ToUpper(in)
	}

	return strings.ToUpper(string(in[0])) + string(in[1:])
}

func camelCase(in string) string {
	// Remove any additional underscores, e.g. convert `foo_1` into `foo1`.
	return strings.Replace(generator.CamelCase(in), "_", "", -1)
}

func oneofEncoderName(inOneof *descriptorpb.OneofDescriptorProto) string {
	typeName := reserveWordProtectedCamelCase(inOneof.GetName())
	return encoderName(typeName)
}

func fieldElmType(inField *descriptorpb.FieldDescriptorProto) (string, error) {
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
		return "Int", nil
	case descriptorpb.FieldDescriptorProto_TYPE_FLOAT,
		descriptorpb.FieldDescriptorProto_TYPE_DOUBLE:
		return "Float", nil
	case descriptorpb.FieldDescriptorProto_TYPE_BOOL:
		return "Bool", nil
	case descriptorpb.FieldDescriptorProto_TYPE_STRING:
		return "String", nil
	case descriptorpb.FieldDescriptorProto_TYPE_BYTES:
		return "Bytes", nil
	case descriptorpb.FieldDescriptorProto_TYPE_MESSAGE,
		descriptorpb.FieldDescriptorProto_TYPE_ENUM:
		// Well known types.
		if n, ok := excludedTypes[inField.GetTypeName()]; ok {
			return n, nil
		}
		_, messageName := convert(inField.GetTypeName())
		return messageName, nil
	default:
		return "", fmt.Errorf("Error generating type for field %q %s", inField.GetName(), inField.GetType())
	}
}

func fieldEncoderName(inField *descriptorpb.FieldDescriptorProto) string {
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
	case descriptorpb.FieldDescriptorProto_TYPE_ENUM:
		// TODO: Default enum value.
		// Remove leading ".".
		_, messageName := convert(inField.GetTypeName())
		return encoderName(messageName)
	case descriptorpb.FieldDescriptorProto_TYPE_MESSAGE:
		// Well Known Types.
		if n, ok := excludedEncoders[inField.GetTypeName()]; ok {
			return n
		}
		_, messageName := convert(inField.GetTypeName())
		return encoderName(messageName)
	case descriptorpb.FieldDescriptorProto_TYPE_BYTES:
		return "bytesFieldEncoder"
	default:
		return fmt.Sprintf("Error generating decoder for field %s", inField.GetType())
	}
}
