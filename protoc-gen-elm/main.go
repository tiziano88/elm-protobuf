package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"unicode"
	"unicode/utf8"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/golang/protobuf/protoc-gen-go/generator"
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
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

func hasMapEntries(inFile *descriptor.FileDescriptorProto) bool {
	for _, m := range inFile.GetMessageType() {
		if hasMapEntriesInMessage(m) {
			return true
		}
	}

	return false
}

func hasMapEntriesInMessage(inMessage *descriptor.DescriptorProto) bool {
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

func templateFile(inFile *descriptor.FileDescriptorProto, p parameters) (string, error) {
	t := template.New("t")

	t, err := t.Parse(`
{{- define "enum-type" }}


type {{ .Name }}
{{- range $i, $v := .Values }}
    {{ if not $i }}={{ else }}|{{ end }} {{ $v.ElmName }} -- {{ $v.Number }}
{{- end }}
{{- end }}
{{- define "enum-decoder" }}


{{ .DecoderName }} : JD.Decoder {{ .Name }}
{{ .DecoderName }} =
    let
        lookup s =
            case s of
{{- range .Values }}
                "{{ .JSONName }}" ->
                    {{ .ElmName }}
{{ end }}
                _ ->
                    {{ .DefaultEnumValue }}
    in
        JD.map lookup JD.string


{{ .DefaultValueName }} : {{ .Name }}
{{ .DefaultValueName }} = {{ .DefaultEnumValue }}
{{- end }}
{{- define "enum-encoder" }}


{{ .EncoderName }} : {{ .Name }} -> JE.Value
{{ .EncoderName }} v =
    let
        lookup s =
            case s of
{{- range .Values }}
                {{ .ElmName }} ->
                    "{{ .JSONName }}"
{{ end }}
    in
        JE.string <| lookup v
{{- end }}
{{- define "oneof-type" }}


type {{ .Name }}
    = {{ .Name }}Unspecified
{{- range .Fields }}
    | {{ .Type }} {{ .ElmType }}
{{- end }}


{{ .DecoderName }} : JD.Decoder {{ .Name }}
{{ .DecoderName }} =
    JD.lazy <| \_ -> JD.oneOf
        [{{ range $i, $v := .Fields }}{{ if $i }},{{ end }} JD.map {{ .Type }} (JD.field "{{ .Name }}" {{ .Decoder }})
        {{ end }}, JD.succeed {{ .Name }}Unspecified
        ]


{{ .EncoderName }} : {{ .Name }} -> Maybe ( String, JE.Value )
{{ .EncoderName }} v =
    case v of
        {{ .Name }}Unspecified ->
            Nothing
        {{- range .Fields }}
        {{ .Type }} x ->
            Just ( "{{ .Name }}", {{ .Encoder }} x )
        {{- end }}
{{- end }}
{{- define "enum" }}
{{- template "enum-type" . }}
{{- template "enum-decoder" . }}
{{- template "enum-encoder" . }}
{{- end }}
{{- define "message" }}


type alias {{ .Type }} =
    { {{ range $i, $v := .Fields }}
        {{- if $i }}, {{ end }}{{ .Name }} : {{ .Type }}{{ if .Number }} -- {{ .Number }}{{ end }}
    {{ end }}}
{{- range .NestedEnums }}{{ template "enum-type" . }}{{ end }}
{{- range .OneOfs }}{{ template "oneof-type" . }}{{ end }}


{{ .DecoderName }} : JD.Decoder {{ .Type }}
{{ .DecoderName }} =
    JD.lazy <| \_ -> decode {{ .Type }}{{ range .Fields }}
        |> {{ .Decoder.Preface }}
			{{- if .JSONName }} "{{ .JSONName }}"{{ end }} {{ .Decoder.Name }}
			{{- if .Decoder.HasDefaultValue }} {{ .Decoder.DefaultValue }}{{ end }}
        {{- end }}
{{- range .NestedEnums }}{{ template "enum-decoder" . }}{{ end }}


{{ .EncoderName }} : {{ .Type }} -> JE.Value
{{ .EncoderName }} v =
    JE.object <| List.filterMap identity <|
        [{{ range $i, $v := .Fields }}
            {{- if $i }},{{ end }} ({{ .Encoder }})
        {{ end }}]
{{- range .NestedEnums }}{{ template "enum-encoder" . }}{{ end }}
{{- range .NestedMessages }}{{ template "message" . }}{{ end }}
{{- end -}}
module {{ .ModuleName }} exposing (..)

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
{{- range .TopEnums }}{{ template "enum" . }}{{ end }}
{{- range .Messages }}{{ template "message" . }}{{ end }}
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
		TopEnums          []enum
		Messages          []message
	}{
		SourceFile:        inFile.GetName(),
		ModuleName:        moduleName(inFile.GetName()),
		ImportDict:        hasMapEntries(inFile),
		AdditionalImports: getAdditionalImports(inFile.GetDependency()),
		TopEnums:          enums([]string{}, inFile.GetEnumType(), p),
		Messages:          messages,
	}); err != nil {
		return "", err
	}

	return buff.String(), nil
}

// FieldNumber is a number assigned to each message, enum, and oneof fields that is unique to that element.
type FieldNumber int32

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

// FieldDecoderHelper is modifier for the field decoder based on properties of the field protobuf.
type FieldDecoderHelper string

const (
	OptionalFieldDecoderHelper FieldDecoderHelper = "optional"
	RepeatedFieldDecoderHelper FieldDecoderHelper = "repeated"
	RequiredFieldDecoderHelper FieldDecoderHelper = "required"
	MapFieldDecoderHelper      FieldDecoderHelper = "mapEntries"
	EnumFieldDecoderHelper     FieldDecoderHelper = "field"
)

type ElmEnumName string

type enumValue struct {
	JSONName string
	ElmName  ElmEnumName
	Number   FieldNumber
}

type enum struct {
	Name             string
	DecoderName      DecoderName
	EncoderName      string
	DefaultValueName string
	DefaultEnumValue ElmEnumName
	Values           []enumValue
}

type fieldDecoder struct {
	Preface         FieldDecoderHelper
	Name            DecoderName
	HasDefaultValue bool
	DefaultValue    string
}

type field struct {
	Name     string
	JSONName string
	Type     string
	Number   FieldNumber
	Decoder  fieldDecoder
	Encoder  string
}

type enumField struct {
	Name    string
	Decoder fieldDecoder
	Encoder string
}

type oneOf struct {
	Name        string
	DecoderName DecoderName
	EncoderName string
	Fields      []oneOfField
}

type oneOfField struct {
	Name    string
	Type    CustomElmType
	ElmType string
	Decoder DecoderName
	Encoder string
}

type nestedField struct {
	Key         string
	Value       string
	DecoderName DecoderName
	EncoderName string
}

type message struct {
	Name           string
	Type           CustomElmType
	DecoderName    DecoderName
	EncoderName    string
	Fields         []field
	EnumFields     []enumField
	OneOfs         []oneOf
	NestedEnums    []enum
	NestedMessages []message
}

func fullElmEnumName(preface []string, value *descriptorpb.EnumValueDescriptorProto) ElmEnumName {
	name := camelCase(strings.ToLower(value.GetName()))
	for _, p := range preface {
		name = fmt.Sprintf("%s_%s", camelCase(strings.ToLower(p)), name)
	}

	return ElmEnumName(name)
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

func enums(preface []string, enumPbs []*descriptor.EnumDescriptorProto, p parameters) []enum {
	var result []enum
	for _, enumPb := range enumPbs {
		if isDeprecated(enumPb.Options) && p.RemoveDeprecated {
			continue
		}

		var values []enumValue
		for _, value := range enumPb.GetValue() {
			if isDeprecated(value.Options) && p.RemoveDeprecated {
				continue
			}

			values = append(values, enumValue{
				JSONName: value.GetName(),
				ElmName:  fullElmEnumName(preface, value),
				Number:   FieldNumber(value.GetNumber()),
			})
		}

		fullName := enumPb.GetName()
		for _, p := range preface {
			fullName = fmt.Sprintf("%s_%s", p, fullName)
		}

		result = append(result, enum{
			Name:             fullName,
			DecoderName:      decoderName(fullName),
			EncoderName:      encoderName(fullName),
			DefaultValueName: defaultEnumValue(fullName),
			DefaultEnumValue: values[0].ElmName,
			Values:           values,
		})
	}

	return result
}

func messages(preface []string, messagePbs []*descriptor.DescriptorProto, p parameters) ([]message, error) {
	var result []message
	for _, messagePb := range messagePbs {

		if isDeprecated(messagePb.Options) && p.RemoveDeprecated {
			continue
		}

		var fields []field
		for _, fieldPb := range messagePb.GetField() {
			if isDeprecated(fieldPb.Options) && p.RemoveDeprecated {
				continue
			}

			if fieldPb.OneofIndex != nil {
				continue
			}

			basicType, err := fieldElmType(fieldPb)
			if err != nil {
				return nil, err
			}

			nestedField, err := nested(fieldPb, messagePb)
			if err != nil {
				return nil, err
			}

			var fieldType string
			var fieldDecoder fieldDecoder
			var fieldEncoder string
			if nestedField != nil {
				fieldType = fmt.Sprintf("Dict.Dict %s %s", nestedField.Key, nestedField.Value)
				fieldDecoder = mapFieldDecoder(*nestedField)
				fieldEncoder = fmt.Sprintf(
					"mapEntriesFieldEncoder %q %s v.%s",
					jsonFieldName(fieldPb),
					nestedField.EncoderName,
					elmFieldName(fieldPb.GetName()),
				)
			} else if isOptional(fieldPb) {
				fieldType = fmt.Sprintf("Maybe %s", basicType)
				fieldDecoder, err = optionalFieldDecoder(fieldPb)
				if err != nil {
					return nil, err
				}

				fieldEncoder = fmt.Sprintf(
					"optionalEncoder %q %s v.%s",
					jsonFieldName(fieldPb),
					fieldEncoderName(fieldPb),
					elmFieldName(fieldPb.GetName()),
				)
			} else if isRepeated(fieldPb) {
				fieldType = fmt.Sprintf("List %s", basicType)
				fieldDecoder, err = repeatedFieldDecoder(fieldPb)
				if err != nil {
					return nil, err
				}

				fieldEncoder = fmt.Sprintf(
					"repeatedFieldEncoder %q %s v.%s",
					jsonFieldName(fieldPb),
					fieldEncoderName(fieldPb),
					elmFieldName(fieldPb.GetName()),
				)
			} else {
				fieldType = basicType
				fieldDecoder, err = requiredFieldDecoder(fieldPb)
				if err != nil {
					return nil, err
				}

				defaultValue, err := fieldDefaultValue(fieldPb)
				if err != nil {
					return nil, err
				}

				fieldEncoder = fmt.Sprintf(
					"requiredFieldEncoder %q %s %s v.%s",
					jsonFieldName(fieldPb),
					fieldEncoderName(fieldPb),
					defaultValue,
					elmFieldName(fieldPb.GetName()),
				)
			}

			fields = append(fields, field{
				Name:     elmFieldName(fieldPb.GetName()),
				JSONName: jsonFieldName(fieldPb),
				Type:     fieldType,
				Decoder:  fieldDecoder,
				Encoder:  fieldEncoder,
				Number:   FieldNumber(fieldPb.GetNumber()),
			})
		}

		var oneOfs []oneOf
		for oneofIndex, oneOfPb := range messagePb.GetOneofDecl() {
			fields = append(fields, field{
				Name:    elmFieldName(oneOfPb.GetName()),
				Type:    elmTypeName(oneOfPb.GetName()),
				Decoder: oneOfFieldDecoder(oneOfPb),
				Encoder: fmt.Sprintf("%s v.%s", oneofEncoderName(oneOfPb), elmFieldName(oneOfPb.GetName())),
			})

			var oneOfFields []oneOfField
			for _, inField := range messagePb.GetField() {
				if inField.OneofIndex != nil && inField.GetOneofIndex() == int32(oneofIndex) {
					fieldDecoder, err := fieldDecoderName(inField)
					if err != nil {
						return nil, err
					}

					oneOfElmType, err := fieldElmType(inField)
					if err != nil {
						return nil, err
					}

					oneOfFields = append(oneOfFields, oneOfField{
						Name:    elmFieldName(inField.GetName()),
						Type:    customElmType(preface, inField.GetName()),
						ElmType: oneOfElmType,
						Decoder: fieldDecoder,
						Encoder: fieldEncoderName(inField),
					})

				}
			}

			oneOfs = append(oneOfs, oneOf{
				Name:        elmTypeName(oneOfPb.GetName()),
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

		result = append(result, message{
			Name:           camelCase(strings.ToLower(fullName)),
			Type:           customElmType(preface, messagePb.GetName()),
			DecoderName:    decoderName(fullName),
			EncoderName:    encoderName(fullName),
			Fields:         fields,
			OneOfs:         oneOfs,
			NestedEnums:    enums(newPreface, messagePb.GetEnumType(), p),
			NestedMessages: nestedMessages,
		})
	}

	return result, nil
}

func isOptional(inField *descriptor.FieldDescriptorProto) bool {
	return inField.GetLabel() == descriptor.FieldDescriptorProto_LABEL_OPTIONAL &&
		inField.GetType() == descriptor.FieldDescriptorProto_TYPE_MESSAGE
}

func isRepeated(inField *descriptor.FieldDescriptorProto) bool {
	return inField.GetLabel() == descriptor.FieldDescriptorProto_LABEL_REPEATED
}

func nested(inField *descriptor.FieldDescriptorProto, inMessage *descriptor.DescriptorProto) (*nestedField, error) {
	fullyQualifiedTypeName := inField.GetTypeName()
	splitName := strings.Split(fullyQualifiedTypeName, ".")
	localTypeName := splitName[len(splitName)-1]

	for _, nested := range inMessage.GetNestedType() {
		if nested.GetName() == localTypeName && nested.GetOptions().GetMapEntry() {
			keyField := nested.GetField()[0]
			valueField := nested.GetField()[1]

			fieldDecoder, err := fieldDecoderName(valueField)
			if err != nil {
				return nil, err
			}

			keyType, err := fieldElmType(keyField)
			if err != nil {
				return nil, err
			}

			valueType, err := fieldElmType(valueField)
			if err != nil {
				return nil, err
			}

			nest := nestedField{
				Key:         keyType,
				Value:       valueType,
				DecoderName: fieldDecoder,
				EncoderName: fieldEncoderName(valueField),
			}

			return &nest, nil
		}
	}

	return nil, nil
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

func elmTypeName(in string) string {
	n := camelCase(in)
	if reservedKeywords[n] {
		n += "_"
	}
	return n
}

func appendUnderscoreToReservedKeywords(in string) string {
	if reservedKeywords[in] {
		return fmt.Sprintf("%s_", in)
	}

	return in
}

func elmFieldName(in string) string {
	return appendUnderscoreToReservedKeywords(firstLower(camelCase(in)))
}

func defaultEnumValue(typeName string) string {
	return firstLower(typeName) + "Default"
}

func encoderName(typeName string) string {
	return firstLower(typeName) + "Encoder"
}

func mapFieldDecoder(nestedField nestedField) fieldDecoder {
	return fieldDecoder{
		Preface:         MapFieldDecoderHelper,
		Name:            nestedField.DecoderName,
		HasDefaultValue: false,
	}
}

func optionalFieldDecoder(fieldPb *descriptor.FieldDescriptorProto) (fieldDecoder, error) {
	basicDecoder, err := fieldDecoderName(fieldPb)
	if err != nil {
		return fieldDecoder{}, err
	}

	return fieldDecoder{
		Preface:         OptionalFieldDecoderHelper,
		Name:            basicDecoder,
		HasDefaultValue: false,
	}, nil
}

func repeatedFieldDecoder(fieldPb *descriptor.FieldDescriptorProto) (fieldDecoder, error) {
	basicDecoder, err := fieldDecoderName(fieldPb)
	if err != nil {
		return fieldDecoder{}, err
	}

	return fieldDecoder{
		Preface:         RepeatedFieldDecoderHelper,
		Name:            basicDecoder,
		HasDefaultValue: false,
	}, nil
}

func requiredFieldDecoder(fieldPb *descriptor.FieldDescriptorProto) (fieldDecoder, error) {
	basicDecoder, err := fieldDecoderName(fieldPb)
	if err != nil {
		return fieldDecoder{}, err
	}

	defaultValue, err := fieldDefaultValue(fieldPb)
	if err != nil {
		return fieldDecoder{}, err
	}

	return fieldDecoder{
		Preface:         RequiredFieldDecoderHelper,
		Name:            basicDecoder,
		HasDefaultValue: true,
		DefaultValue:    defaultValue,
	}, nil
}

func oneOfFieldDecoder(oneOfPb *descriptor.OneofDescriptorProto) fieldDecoder {
	return fieldDecoder{
		Preface:         EnumFieldDecoderHelper,
		Name:            oneofDecoderName(oneOfPb),
		HasDefaultValue: false,
	}
}

func fieldDecoderName(inField *descriptor.FieldDescriptorProto) (DecoderName, error) {
	switch inField.GetType() {
	case descriptor.FieldDescriptorProto_TYPE_INT32,
		descriptor.FieldDescriptorProto_TYPE_INT64,
		descriptor.FieldDescriptorProto_TYPE_UINT32,
		descriptor.FieldDescriptorProto_TYPE_UINT64,
		descriptor.FieldDescriptorProto_TYPE_SINT32,
		descriptor.FieldDescriptorProto_TYPE_SINT64,
		descriptor.FieldDescriptorProto_TYPE_FIXED32,
		descriptor.FieldDescriptorProto_TYPE_FIXED64,
		descriptor.FieldDescriptorProto_TYPE_SFIXED32,
		descriptor.FieldDescriptorProto_TYPE_SFIXED64:
		return IntDecoder, nil
	case descriptor.FieldDescriptorProto_TYPE_FLOAT,
		descriptor.FieldDescriptorProto_TYPE_DOUBLE:
		return FloatDecoder, nil
	case descriptor.FieldDescriptorProto_TYPE_BOOL:
		return BoolDecoder, nil
	case descriptor.FieldDescriptorProto_TYPE_STRING:
		return StringDecoder, nil

	// TODO: This is an unsupported stub (throw error)
	case descriptor.FieldDescriptorProto_TYPE_BYTES:
		return BytesDecoder, nil
	case descriptor.FieldDescriptorProto_TYPE_ENUM:
		return enumDecoderName(inField), nil
	case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
		return messageDecoderName(inField), nil
	default:
		return "", fmt.Errorf("error generating decoder for field %s", inField.GetType())
	}
}

func enumDecoderName(inField *descriptor.FieldDescriptorProto) DecoderName {
	_, messageName := convert(inField.GetTypeName())
	return decoderName(messageName)
}

func messageDecoderName(inField *descriptor.FieldDescriptorProto) DecoderName {
	// Well Known Types.
	if n, ok := excludedDecoders[inField.GetTypeName()]; ok {
		return n
	}

	_, messageName := convert(inField.GetTypeName())
	return decoderName(messageName)
}

func oneofDecoderName(inOneof *descriptor.OneofDescriptorProto) DecoderName {
	typeName := elmTypeName(inOneof.GetName())
	return decoderName(typeName)
}

func decoderName(typeName string) DecoderName {
	return DecoderName(firstLower(typeName) + "Decoder")
}

func elmFieldType(field *descriptor.FieldDescriptorProto) string {
	inFieldName := field.GetTypeName()
	_, messageName := convert(inFieldName)

	return messageName
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

func jsonFieldName(field *descriptor.FieldDescriptorProto) string {
	return field.GetJsonName()
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

func oneofEncoderName(inOneof *descriptor.OneofDescriptorProto) string {
	typeName := elmTypeName(inOneof.GetName())
	return encoderName(typeName)
}

// mapEntries detects whether a field is implemented using a `map<,>` field or not.
// In order for `map<,>` fields to be supported by proto2 format, they
// get parsed as a backwards compatible form:
//
//     map<KeyType, ValueType> map_field = 1;
// Gets parsed as:
//     message MapFieldEntry {
//         option map_entry = true;
//         optional KeyType key = 1;
//         optional ValueType value = 2;
//     }
//     repeated MapFieldEntry map_field = 1;
//
// our code looks for the `map_entry` option to detect `map<,>` fields, and generate Dict's for them
// https://github.com/golang/protobuf/blob/882cf97a83ad205fd22af574246a3bc647d7a7d2/protoc-gen-go/descriptor/descriptor.proto#L474-L495
func mapEntries(inField *descriptor.FieldDescriptorProto, inMessage *descriptor.DescriptorProto) (isMap bool, keyFieldDescriptor *descriptor.FieldDescriptorProto, valueFieldDescriptor *descriptor.FieldDescriptorProto) {
	isRepeated :=
		inField.GetLabel() == descriptor.FieldDescriptorProto_LABEL_REPEATED &&
			inField.GetType() == descriptor.FieldDescriptorProto_TYPE_MESSAGE

	if !isRepeated {
		return false, nil, nil
	}

	fullyQualifiedTypeName := inField.GetTypeName()
	splitName := strings.Split(fullyQualifiedTypeName, ".")
	localTypeName := splitName[len(splitName)-1]

	for _, nested := range inMessage.GetNestedType() {
		if nested.GetName() == localTypeName && nested.GetOptions().GetMapEntry() {
			return true, nested.GetField()[0], nested.GetField()[1]
		}
	}
	return false, nil, nil
}

func fieldElmType(inField *descriptor.FieldDescriptorProto) (string, error) {
	switch inField.GetType() {
	case descriptor.FieldDescriptorProto_TYPE_INT32,
		descriptor.FieldDescriptorProto_TYPE_INT64,
		descriptor.FieldDescriptorProto_TYPE_UINT32,
		descriptor.FieldDescriptorProto_TYPE_UINT64,
		descriptor.FieldDescriptorProto_TYPE_SINT32,
		descriptor.FieldDescriptorProto_TYPE_SINT64,
		descriptor.FieldDescriptorProto_TYPE_FIXED32,
		descriptor.FieldDescriptorProto_TYPE_FIXED64,
		descriptor.FieldDescriptorProto_TYPE_SFIXED32,
		descriptor.FieldDescriptorProto_TYPE_SFIXED64:
		return "Int", nil
	case descriptor.FieldDescriptorProto_TYPE_FLOAT,
		descriptor.FieldDescriptorProto_TYPE_DOUBLE:
		return "Float", nil
	case descriptor.FieldDescriptorProto_TYPE_BOOL:
		return "Bool", nil
	case descriptor.FieldDescriptorProto_TYPE_STRING:
		return "String", nil
	case descriptor.FieldDescriptorProto_TYPE_BYTES:
		return "Bytes", nil
	case descriptor.FieldDescriptorProto_TYPE_MESSAGE,
		descriptor.FieldDescriptorProto_TYPE_ENUM:
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

func fieldEncoderName(inField *descriptor.FieldDescriptorProto) string {
	switch inField.GetType() {
	case descriptor.FieldDescriptorProto_TYPE_INT32,
		descriptor.FieldDescriptorProto_TYPE_UINT32,
		descriptor.FieldDescriptorProto_TYPE_SINT32,
		descriptor.FieldDescriptorProto_TYPE_FIXED32,
		descriptor.FieldDescriptorProto_TYPE_SFIXED32:
		return "JE.int"
	case descriptor.FieldDescriptorProto_TYPE_INT64,
		descriptor.FieldDescriptorProto_TYPE_UINT64,
		descriptor.FieldDescriptorProto_TYPE_SINT64,
		descriptor.FieldDescriptorProto_TYPE_FIXED64,
		descriptor.FieldDescriptorProto_TYPE_SFIXED64:
		return "numericStringEncoder"
	case descriptor.FieldDescriptorProto_TYPE_FLOAT,
		descriptor.FieldDescriptorProto_TYPE_DOUBLE:
		return "JE.float"
	case descriptor.FieldDescriptorProto_TYPE_BOOL:
		return "JE.bool"
	case descriptor.FieldDescriptorProto_TYPE_STRING:
		return "JE.string"
	case descriptor.FieldDescriptorProto_TYPE_ENUM:
		// TODO: Default enum value.
		// Remove leading ".".
		_, messageName := convert(inField.GetTypeName())
		return encoderName(messageName)
	case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
		// Well Known Types.
		if n, ok := excludedEncoders[inField.GetTypeName()]; ok {
			return n
		}
		_, messageName := convert(inField.GetTypeName())
		return encoderName(messageName)
	case descriptor.FieldDescriptorProto_TYPE_BYTES:
		return "bytesFieldEncoder"
	default:
		return fmt.Sprintf("Error generating decoder for field %s", inField.GetType())
	}
}

func fieldDefaultValue(inField *descriptor.FieldDescriptorProto) (string, error) {
	if inField.GetLabel() == descriptor.FieldDescriptorProto_LABEL_REPEATED {
		return "[]", nil
	}

	switch inField.GetType() {
	case descriptor.FieldDescriptorProto_TYPE_INT32,
		descriptor.FieldDescriptorProto_TYPE_INT64,
		descriptor.FieldDescriptorProto_TYPE_UINT32,
		descriptor.FieldDescriptorProto_TYPE_UINT64,
		descriptor.FieldDescriptorProto_TYPE_SINT32,
		descriptor.FieldDescriptorProto_TYPE_SINT64,
		descriptor.FieldDescriptorProto_TYPE_FIXED32,
		descriptor.FieldDescriptorProto_TYPE_FIXED64,
		descriptor.FieldDescriptorProto_TYPE_SFIXED32,
		descriptor.FieldDescriptorProto_TYPE_SFIXED64:
		return "0", nil
	case descriptor.FieldDescriptorProto_TYPE_FLOAT,
		descriptor.FieldDescriptorProto_TYPE_DOUBLE:
		return "0.0", nil
	case descriptor.FieldDescriptorProto_TYPE_BOOL:
		return "False", nil
	case descriptor.FieldDescriptorProto_TYPE_STRING:
		return "\"\"", nil
	case descriptor.FieldDescriptorProto_TYPE_ENUM:
		// TODO: Default enum value.
		_, messageName := convert(inField.GetTypeName())
		return defaultEnumValue(messageName), nil
	case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
		return "xxx", nil
	case descriptor.FieldDescriptorProto_TYPE_BYTES:
		return "[]", nil
	default:
		return "", fmt.Errorf("error generating decoder for field %s", inField.GetType())
	}
}
