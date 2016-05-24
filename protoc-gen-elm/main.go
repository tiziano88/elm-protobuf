package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/golang/protobuf/proto"

	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/golang/protobuf/protoc-gen-go/generator"
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
)

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

	// Remove useless source code data.
	for _, inFile := range req.GetProtoFile() {
		inFile.SourceCodeInfo = nil
	}

	log.Printf("Input data: %v", proto.MarshalTextString(req))

	resp := &plugin.CodeGeneratorResponse{}

	for _, inFile := range req.GetProtoFile() {
		log.Printf("Processing file %s", inFile.GetName())
		outFile, err := processFile(inFile)
		if err != nil {
			log.Fatalf("Could not process file: %v", err)
		}
		resp.File = append(resp.File, outFile)
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

func processFile(inFile *descriptor.FileDescriptorProto) (*plugin.CodeGeneratorResponse_File, error) {
	if inFile.GetSyntax() != "proto3" {
		return nil, fmt.Errorf("Only proto3 syntax is supported")
	}

	outFile := &plugin.CodeGeneratorResponse_File{}

	inFileName := inFile.GetName()

	inFileDir, inFileFile := filepath.Split(inFileName)
	shortModuleName := firstUpper(strings.TrimSuffix(inFileFile, ".proto"))
	fullModuleName := strings.Replace(inFileDir, "/", ".", -1) + shortModuleName
	outFileName := filepath.Join(inFileDir, shortModuleName+".elm")
	outFile.Name = proto.String(outFileName)

	fg := NewFileGenerator()

	fg.GenerateModule(fullModuleName)
	fg.GenerateImports()
	fg.GenerateRuntime()

	var err error

	// Top-level enums.
	for _, inEnum := range inFile.GetEnumType() {
		err = fg.GenerateEnumThings("", inEnum)
		if err != nil {
			return nil, err
		}
	}

	// Top-level messages.
	for _, inMessage := range inFile.GetMessageType() {
		err = fg.GenerateEverything("", inMessage)
		if err != nil {
			return nil, err
		}
	}

	outFile.Content = proto.String(fg.out.String())

	return outFile, nil
}

type FileGenerator struct {
	out    bytes.Buffer
	indent uint
}

func NewFileGenerator() *FileGenerator {
	return &FileGenerator{}
}

func (fg *FileGenerator) In() {
	fg.indent += 1
}

func (fg *FileGenerator) Out() {
	fg.indent -= 1
}

func (fg *FileGenerator) P(format string, a ...interface{}) error {
	var err error

	_, err = fg.out.WriteString(strings.Repeat("  ", int(fg.indent)))
	if err != nil {
		return err
	}

	s := fmt.Sprintf(format, a...)
	_, err = fg.out.WriteString(s)
	if err != nil {
		return err
	}
	_, err = fg.out.WriteString("\n")
	if err != nil {
		return err
	}
	return nil
}

func (fg *FileGenerator) GenerateEnum(prefix string, inEnum *descriptor.EnumDescriptorProto) error {
	typeName := prefix + inEnum.GetName()
	fg.P("type %s", typeName)
	fg.In()
	leading := "="
	for _, enumValue := range inEnum.GetValue() {
		// TODO: Convert names to CamelCase.
		fg.P("%s %s -- %d", leading, prefix+elmEnumValueName(enumValue.GetName()), enumValue.GetNumber())
		leading = "|"
	}
	fg.Out()
	return nil
}

func (fg *FileGenerator) GenerateEnumDecoder(prefix string, inEnum *descriptor.EnumDescriptorProto) error {
	typeName := prefix + inEnum.GetName()
	decoderName := decoderName(typeName)
	fg.P("%s : JD.Decoder %s", decoderName, typeName)
	fg.P("%s =", decoderName)
	fg.In()
	fg.P("let")
	fg.In()
	fg.P("lookup s = case s of")
	fg.In()
	for _, enumValue := range inEnum.GetValue() {
		fg.P("%q -> %s", enumValue.GetName(), prefix+elmEnumValueName(enumValue.GetName()))
	}
	// TODO: This should fail instead.
	fg.P("_ -> %s", prefix+elmEnumValueName(inEnum.GetValue()[0].GetName()))
	fg.Out()
	fg.Out()
	fg.P("in")
	fg.In()
	fg.P("JD.map lookup JD.string")
	fg.Out()
	fg.Out()

	fg.P("")
	fg.P("")

	defaultName := defaultEnumValue(typeName)
	fg.P("%s : %s", defaultName, typeName)
	fg.P("%s = %s", defaultName, prefix+elmEnumValueName(inEnum.GetValue()[0].GetName()))
	return nil
}

func (fg *FileGenerator) GenerateEnumEncoder(prefix string, inEnum *descriptor.EnumDescriptorProto) error {
	typeName := prefix + inEnum.GetName()
	argName := "v"
	fg.P("%s : %s -> JE.Value", encoderName(typeName), typeName)
	fg.P("%s %s =", encoderName(typeName), argName)
	fg.In()
	fg.P("let")
	fg.In()
	fg.P("lookup s = case s of")
	fg.In()
	for _, enumValue := range inEnum.GetValue() {
		fg.P("%s -> %q", prefix+elmEnumValueName(enumValue.GetName()), enumValue.GetName())
	}
	fg.Out()
	fg.Out()
	fg.P("in")
	fg.In()
	fg.P("JE.string <| lookup %s", argName)
	fg.Out()
	fg.Out()
	return nil
}

func (fg *FileGenerator) GenerateModule(moduleName string) {
	fg.P("module %s exposing (..)", moduleName)

	fg.P("")
	fg.P("")
}

func (fg *FileGenerator) GenerateImports() {
	fg.P("import Json.Decode as JD exposing ((:=))")
	fg.P("import Json.Encode as JE")

	fg.P("")
	fg.P("")
}

func (fg *FileGenerator) GenerateRuntime() {
	// Applicative-style decoders. This is fine as long as this is the only Applicative in the
	// package, otherwise operator will clash, since Elm does not have support to generalise
	// them via HKTs.

	fg.P("(<$>) : (a -> b) -> JD.Decoder a -> JD.Decoder b")
	fg.P("(<$>) =")
	fg.In()
	fg.P("JD.map")
	fg.Out()

	fg.P("")
	fg.P("")

	fg.P("(<*>) : JD.Decoder (a -> b) -> JD.Decoder a -> JD.Decoder b")
	fg.P("(<*>) f v =")
	fg.In()
	fg.P("f `JD.andThen` \\x -> x <$> v")
	fg.Out()

	fg.P("")
	fg.P("")

	fg.P("optionalDecoder : JD.Decoder a -> JD.Decoder (Maybe a)")
	fg.P("optionalDecoder decoder =")
	fg.In()
	fg.P("JD.oneOf")
	fg.In()
	fg.P("[ JD.map Just decoder")
	fg.P(", JD.succeed Nothing")
	fg.P("]")
	fg.Out()
	fg.Out()

	fg.P("")
	fg.P("")

	fg.P("requiredFieldDecoder : String -> a -> JD.Decoder a -> JD.Decoder a")
	fg.P("requiredFieldDecoder name default decoder =")
	fg.In()
	fg.P("withDefault default (name := decoder)")
	fg.Out()

	fg.P("")
	fg.P("")

	fg.P("optionalFieldDecoder : String -> JD.Decoder a -> JD.Decoder (Maybe a)")
	fg.P("optionalFieldDecoder name decoder =")
	fg.In()
	fg.P("optionalDecoder (name := decoder)")
	fg.Out()

	fg.P("")
	fg.P("")

	fg.P("repeatedFieldDecoder : String -> JD.Decoder a -> JD.Decoder (List a)")
	fg.P("repeatedFieldDecoder name decoder =")
	fg.In()
	fg.P("withDefault [] (name := (JD.list decoder))")
	fg.Out()

	fg.P("")
	fg.P("")

	fg.P("withDefault : a -> JD.Decoder a -> JD.Decoder a")
	fg.P("withDefault default decoder =")
	fg.In()
	fg.P("JD.oneOf")
	fg.In()
	fg.P("[ decoder")
	fg.P(", JD.succeed default")
	fg.P("]")
	fg.Out()
	fg.Out()

	fg.P("")
	fg.P("")

	fg.P("optionalEncoder : (a -> JE.Value) -> Maybe a -> JE.Value")
	fg.P("optionalEncoder encoder v =")
	fg.In()
	fg.P("case v of")
	fg.In()
	fg.P("Just x ->")
	fg.In()
	fg.P("encoder x")
	fg.Out()
	fg.P("")
	fg.P("Nothing ->")
	fg.In()
	fg.P("JE.null")
	fg.Out()
	fg.Out()
	fg.Out()

	fg.P("")
	fg.P("")

	fg.P("repeatedFieldEncoder : (a -> JE.Value) -> List a -> JE.Value")
	fg.P("repeatedFieldEncoder encoder v =")
	fg.In()
	fg.P("JE.list <| List.map encoder v")
	fg.Out()

	fg.P("")
	fg.P("")
}

func (fg *FileGenerator) GenerateEnumThings(prefix string, inEnum *descriptor.EnumDescriptorProto) error {
	var err error

	err = fg.GenerateEnum(prefix, inEnum)
	if err != nil {
		return err
	}

	fg.P("")
	fg.P("")

	err = fg.GenerateEnumDecoder(prefix, inEnum)
	if err != nil {
		return err
	}

	fg.P("")
	fg.P("")

	err = fg.GenerateEnumEncoder(prefix, inEnum)
	if err != nil {
		return err
	}

	fg.P("")
	fg.P("")

	return nil
}

func (fg *FileGenerator) GenerateEverything(prefix string, inMessage *descriptor.DescriptorProto) error {
	var err error

	err = fg.GenerateMessage(prefix, inMessage)
	if err != nil {
		return err
	}

	fg.P("")
	fg.P("")

	err = fg.GenerateMessageDecoder(prefix, inMessage)
	if err != nil {
		return err
	}

	fg.P("")
	fg.P("")

	err = fg.GenerateMessageEncoder(prefix, inMessage)
	if err != nil {
		return err
	}

	fg.P("")
	fg.P("")

	newPrefix := prefix + inMessage.GetName() + "_"

	// Nested enums.
	for _, inEnum := range inMessage.GetEnumType() {
		err = fg.GenerateEnumThings(newPrefix, inEnum)
		if err != nil {
			return err
		}
	}

	// Nested messages.
	for _, nested := range inMessage.GetNestedType() {
		fg.GenerateEverything(newPrefix, nested)
	}

	return nil
}

func fieldElmType(inField *descriptor.FieldDescriptorProto) string {
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
		return "Int"
	case descriptor.FieldDescriptorProto_TYPE_FLOAT,
		descriptor.FieldDescriptorProto_TYPE_DOUBLE:
		return "Float"
	case descriptor.FieldDescriptorProto_TYPE_BOOL:
		return "Bool"
	case descriptor.FieldDescriptorProto_TYPE_STRING:
		return "String"
	case descriptor.FieldDescriptorProto_TYPE_ENUM:
		// XXX
		return elmFieldType(inField)
	case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
		// XXX
		return elmFieldType(inField)
	case descriptor.FieldDescriptorProto_TYPE_BYTES:
		// XXX
		return "Bytes"
	default:
		// TODO: Return error.
		return fmt.Sprintf("Error generating type for field %q %s", inField.GetName(), inField.GetType())
	}
}

func (fg *FileGenerator) GenerateMessage(prefix string, inMessage *descriptor.DescriptorProto) error {
	typeName := prefix + inMessage.GetName()
	fg.P("type alias %s =", typeName)
	fg.In()

	leading := "{"
	for _, inField := range inMessage.GetField() {
		optional := (inField.GetLabel() == descriptor.FieldDescriptorProto_LABEL_OPTIONAL) &&
			(inField.GetType() == descriptor.FieldDescriptorProto_TYPE_MESSAGE)
		repeated := inField.GetLabel() == descriptor.FieldDescriptorProto_LABEL_REPEATED

		fType := fieldElmType(inField)

		fName := elmFieldName(inField.GetName())
		fNumber := inField.GetNumber()

		if repeated {
			fg.P("%s %s : List %s -- %d", leading, fName, fType, fNumber)
		} else {
			if optional {
				fg.P("%s %s : Maybe %s -- %d", leading, fName, fType, fNumber)
			} else {
				fg.P("%s %s : %s -- %d", leading, fName, fType, fNumber)
			}
		}

		leading = ","
	}

	fg.P("")

	leading = "{"
	for _, inOneof := range inMessage.GetOneofDecl() {

		oneofName := elmFieldName(inOneof.GetName())
		// TODO: Prefix with message name to avoid collisions.
		oneofTypeName := elmTypeName(inOneof.GetName())
		fg.P("%s %s : %s", leading, oneofName, oneofTypeName)

		leading = ","
	}

	fg.P("}")
	fg.Out()

	fg.P("")

	for i, _ := range inMessage.GetOneofDecl() {
		fg.GenerateOneofDefinition(prefix, inMessage, i)
		fg.GenerateOneofDecoder(prefix, inMessage, i)
		fg.GenerateOneofEncoder(prefix, inMessage, i)
	}

	return nil
}

func elmOneofDecoderName(inOneof *descriptor.OneofDescriptorProto) string {
	typeName := elmTypeName(inOneof.GetName())
	return decoderName(typeName)
}

func elmOneofTypeName(inOneof *descriptor.OneofDescriptorProto) string {
	return elmTypeName(inOneof.GetName())
}

func (fg *FileGenerator) GenerateOneofDefinition(prefix string, inMessage *descriptor.DescriptorProto, oneofIndex int) error {
	inOneof := inMessage.GetOneofDecl()[oneofIndex]

	// TODO: Prefix with message name to avoid collisions.
	oneofTypeName := elmOneofTypeName(inOneof)
	fg.P("type %s", oneofTypeName)

	fg.In()

	leading := "="
	for _, inField := range inMessage.GetField() {
		if inField.OneofIndex != nil && inField.GetOneofIndex() == int32(oneofIndex) {

			oneofVariantName := elmTypeName(inField.GetName())
			oneofArgumentType := fieldElmType(inField)
			fg.P("%s %s %s", leading, oneofVariantName, oneofArgumentType)

			leading = "|"
		}
	}
	fg.Out()
	fg.P("")

	return nil
}

func (fg *FileGenerator) GenerateOneofDecoder(prefix string, inMessage *descriptor.DescriptorProto, oneofIndex int) error {
	inOneof := inMessage.GetOneofDecl()[oneofIndex]

	// TODO: Prefix with message name to avoid collisions.
	typeName := elmOneofTypeName(inOneof)
	decoderName := elmOneofDecoderName(inOneof)

	fg.P("%s : JD.Decoder %s", decoderName, typeName)
	fg.P("%s =", decoderName)

	fg.In()

	fg.P("JD.oneOf")
	fg.In()

	leading := "["
	for _, inField := range inMessage.GetField() {
		if inField.OneofIndex != nil && inField.GetOneofIndex() == int32(oneofIndex) {
			oneofVariantName := elmTypeName(inField.GetName())
			decoderName := fieldElmDecoderName(inField)
			fg.P("%s JD.map %s (%q := %s)", leading, oneofVariantName, inField.GetJsonName(), decoderName)
			leading = ","
		}
	}
	fg.P("]")
	fg.Out()
	fg.Out()
	fg.P("")

	return nil
}

func (fg *FileGenerator) GenerateOneofEncoder(prefix string, inMessage *descriptor.DescriptorProto, oneofIndex int) error {
	return nil
}

func fieldElmDefaultValue(inField *descriptor.FieldDescriptorProto) string {
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
		return "0"
	case descriptor.FieldDescriptorProto_TYPE_FLOAT,
		descriptor.FieldDescriptorProto_TYPE_DOUBLE:
		return "0.0"
	case descriptor.FieldDescriptorProto_TYPE_BOOL:
		return "False"
	case descriptor.FieldDescriptorProto_TYPE_STRING:
		return "\"\""
	case descriptor.FieldDescriptorProto_TYPE_ENUM:
		// TODO: Default enum value.
		return defaultEnumValue(elmFieldType(inField))
	case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
		return "xxx"
	case descriptor.FieldDescriptorProto_TYPE_BYTES:
		return "xxx"
	default:
		return fmt.Sprintf("Error generating decoder for field %s", inField.GetType())
	}
}

func fieldElmDecoderName(inField *descriptor.FieldDescriptorProto) string {
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
		// TODO: Handle parsing from string (for 64 bit types).
		return "JD.int"
	case descriptor.FieldDescriptorProto_TYPE_FLOAT,
		descriptor.FieldDescriptorProto_TYPE_DOUBLE:
		return "JD.float"
	case descriptor.FieldDescriptorProto_TYPE_BOOL:
		return "JD.bool"
	case descriptor.FieldDescriptorProto_TYPE_STRING:
		return "JD.string"
	case descriptor.FieldDescriptorProto_TYPE_ENUM:
		// TODO: Default enum value.
		// Remove leading ".".
		return decoderName(elmFieldType(inField))
	case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
		// Remove leading ".".
		return decoderName(elmFieldType(inField))
	case descriptor.FieldDescriptorProto_TYPE_BYTES:
		return "bytesFieldDecoder"
	default:
		return fmt.Sprintf("Error generating decoder for field %s", inField.GetType())
	}
}

func (fg *FileGenerator) GenerateMessageDecoder(prefix string, inMessage *descriptor.DescriptorProto) error {
	typeName := prefix + inMessage.GetName()
	fg.P("%s : JD.Decoder %s", decoderName(typeName), typeName)
	fg.P("%s =", decoderName(typeName))
	fg.In()
	fg.P("%s", typeName)
	fg.In()

	leading := "<$>"
	for _, inField := range inMessage.GetField() {
		optional := (inField.GetLabel() == descriptor.FieldDescriptorProto_LABEL_OPTIONAL) &&
			(inField.GetType() == descriptor.FieldDescriptorProto_TYPE_MESSAGE)
		repeated := inField.GetLabel() == descriptor.FieldDescriptorProto_LABEL_REPEATED
		d := fieldElmDecoderName(inField)
		def := fieldElmDefaultValue(inField)

		if repeated {
			fg.P("%s (repeatedFieldDecoder %q %s)", leading, jsonFieldName(inField), d)
		} else {
			if optional {
				fg.P("%s (optionalFieldDecoder %q %s)", leading, jsonFieldName(inField), d)
			} else {
				fg.P("%s (requiredFieldDecoder %q %s %s)", leading, jsonFieldName(inField), def, d)
			}
		}

		leading = "<*>"
	}

	for _, inOneof := range inMessage.GetOneofDecl() {
		oneofDecoderName := elmOneofDecoderName(inOneof)
		fg.P("%s %s", leading, oneofDecoderName)

		leading = "<*>"
	}

	fg.Out()
	fg.Out()
	return nil
}

func (fg *FileGenerator) GenerateMessageEncoder(prefix string, inMessage *descriptor.DescriptorProto) error {
	typeName := prefix + inMessage.GetName()
	argName := "v"
	fg.P("%s : %s -> JE.Value", encoderName(typeName), typeName)
	fg.P("%s %s =", encoderName(typeName), argName)
	fg.In()
	fg.P("JE.object")
	fg.In()

	leading := "["
	for _, inField := range inMessage.GetField() {
		optional := (inField.GetLabel() == descriptor.FieldDescriptorProto_LABEL_OPTIONAL) &&
			(inField.GetType() == descriptor.FieldDescriptorProto_TYPE_MESSAGE)
		repeated := inField.GetLabel() == descriptor.FieldDescriptorProto_LABEL_REPEATED
		d := ""
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
			d = "JE.int"
		case descriptor.FieldDescriptorProto_TYPE_FLOAT,
			descriptor.FieldDescriptorProto_TYPE_DOUBLE:
			d = "JE.float"
		case descriptor.FieldDescriptorProto_TYPE_BOOL:
			d = "JE.bool"
		case descriptor.FieldDescriptorProto_TYPE_STRING:
			d = "JE.string"
		case descriptor.FieldDescriptorProto_TYPE_ENUM:
			// Remove leading ".".
			d = encoderName(elmFieldType(inField))
		case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
			// Remove leading ".".
			d = encoderName(elmFieldType(inField))
		case descriptor.FieldDescriptorProto_TYPE_BYTES:
			d = "bytesFieldEncoder"
		default:
			return fmt.Errorf("Error generating encoder for field %s", inField.GetType())
		}

		val := argName + "." + elmFieldName(inField.GetName())
		if repeated {
			fg.P("%s (%q, repeatedFieldEncoder %s %s)", leading, jsonFieldName(inField), d, val)
		} else {
			if optional {
				fg.P("%s (%q, optionalEncoder %s %s)", leading, jsonFieldName(inField), d, val)
			} else {
				fg.P("%s (%q, %s %s)", leading, jsonFieldName(inField), d, val)
			}
		}

		leading = ","
	}
	fg.P("]")
	fg.Out()
	fg.Out()
	return nil
}

func elmTypeName(in string) string {
	return camelCase(in)
}

func elmFieldName(in string) string {
	return firstLower(camelCase(in))
}

func elmEnumValueName(in string) string {
	return camelCase(strings.ToLower(in))
}

func decoderName(typeName string) string {
	packageName, messageName := convert(typeName)

	if packageName == "" {
		return firstLower(messageName) + "Decoder"
	} else {
		return packageName + "." + firstLower(messageName) + "Decoder"
	}
}

func defaultEnumValue(typeName string) string {
	packageName, messageName := convert(typeName)

	if packageName == "" {
		return firstLower(messageName) + "Default"
	} else {
		return packageName + "." + firstLower(messageName) + "Default"
	}
}

func encoderName(typeName string) string {
	packageName, messageName := convert(typeName)

	if packageName == "" {
		return firstLower(messageName) + "Encoder"
	} else {
		return packageName + "." + firstLower(messageName) + "Encoder"
	}
}

func elmFieldType(field *descriptor.FieldDescriptorProto) string {
	inFieldName := field.GetTypeName()
	packageName, messageName := convert(inFieldName)

	if packageName == "" {
		return messageName
	} else {
		return packageName + "." + messageName
	}
}

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
	if in == "" {
		return ""
	}
	if len(in) == 1 {
		return strings.ToLower(in)
	}
	return strings.ToLower(string(in[0])) + string(in[1:])
}

func firstUpper(in string) string {
	if in == "" {
		return ""
	}
	if len(in) == 1 {
		return strings.ToUpper(in)
	}
	return strings.ToUpper(string(in[0])) + string(in[1:])
}

func camelCase(in string) string {
	// Remove any additional underscores, e.g. convert `foo_1` into `foo1`.
	return strings.Replace(generator.CamelCase(in), "_", "", -1)
}
