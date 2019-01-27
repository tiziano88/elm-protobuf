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

var (
	// Maps each type to the file in which it was defined.
	typeToFile = map[string]string{}

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
	excludedDecoders = map[string]string{
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
		// Well Known Types.
		if excludedFiles[inFile.GetName()] {
			log.Printf("Skipping well known type")
			continue
		}
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

func processFile(inFile *descriptor.FileDescriptorProto) (*plugin.CodeGeneratorResponse_File, error) {
	if inFile.GetSyntax() != "proto3" {
		return nil, fmt.Errorf("Only proto3 syntax is supported")
	}

	outFile := &plugin.CodeGeneratorResponse_File{}

	inFilePath := inFile.GetName()
	inFileDir, inFileName := filepath.Split(inFilePath)

	shortModuleName := firstUpper(strings.TrimSuffix(inFileName, ".proto"))

	fullModuleName := ""
	outFileName := ""
	for _, segment := range strings.Split(inFileDir, "/") {
		if segment == "" {
			continue
		}
		fullModuleName += firstUpper(segment) + "."
		outFileName += firstUpper(segment) + "/"
	}
	fullModuleName += shortModuleName
	outFileName += shortModuleName + ".elm"

	outFile.Name = proto.String(outFileName)

	b := &bytes.Buffer{}
	fg := NewFileGenerator(b, inFileName)

	fg.GenerateModule(fullModuleName)
	fg.GenerateComments(inFile)

	includeDictImport := hasMapEntries(inFile)
	fg.GenerateImports(includeDictImport)

	// Generate additional imports.
	for _, d := range inFile.GetDependency() {
		// Well Known Types.
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
		fullModuleName = strings.TrimSuffix(fullModuleName, ".")
		// TODO: Do not expose everything.
		fg.P("import %s exposing (..)", fullModuleName)
	}

	var err error

	// Top-level enums.
	for _, inEnum := range inFile.GetEnumType() {
		typeToFile[strings.TrimPrefix(inFile.GetPackage()+"."+inEnum.GetName(), ".")] = inFile.GetName()

		err = fg.GenerateEnumDefinition("", inEnum)
		if err != nil {
			return nil, err
		}

		err = fg.GenerateEnumDecoder("", inEnum)
		if err != nil {
			return nil, err
		}

		err = fg.GenerateEnumEncoder("", inEnum)
		if err != nil {
			return nil, err
		}
	}

	// Top-level messages.
	for _, inMessage := range inFile.GetMessageType() {
		typeToFile[strings.TrimPrefix(inFile.GetPackage()+"."+inMessage.GetName(), ".")] = inFile.GetName()

		err = fg.GenerateEverything("", inMessage)
		if err != nil {
			return nil, err
		}
	}

	outFile.Content = proto.String(b.String())

	return outFile, nil
}

func (fg *FileGenerator) GenerateModule(moduleName string) {
	fg.P("module %s exposing (..)", moduleName)
}

func (fg *FileGenerator) GenerateComments(inFile *descriptor.FileDescriptorProto) {
	fg.P("")
	fg.P("-- DO NOT EDIT")
	fg.P("-- AUTOGENERATED BY THE ELM PROTOCOL BUFFER COMPILER")
	fg.P("-- https://github.com/tiziano88/elm-protobuf")
	fg.P("-- source file: %s", inFile.GetName())
}

func (fg *FileGenerator) GenerateImports(includeDictImport bool) {
	fg.P("")
	fg.P("import Protobuf exposing (..)")
	fg.P("")
	fg.P("import Json.Decode as JD")
	fg.P("import Json.Encode as JE")
	if includeDictImport {
		fg.P("import Dict")
	}
}

func (fg *FileGenerator) GenerateEverything(prefix string, inMessage *descriptor.DescriptorProto) error {
	newPrefix := prefix + inMessage.GetName() + "_"
	var err error

	if inMessage.Options.GetMapEntry() {
		if len(inMessage.Field) != 2 {
			return fmt.Errorf("map entry must have exactly two fields")
		}

		keyField := inMessage.Field[0]
		if keyField.GetName() != "key" {
			return fmt.Errorf("first map entry field must be called `key`")
		}
		if keyField.GetType() != descriptor.FieldDescriptorProto_TYPE_STRING &&
			keyField.GetType() != descriptor.FieldDescriptorProto_TYPE_INT32 {
			return fmt.Errorf("map key must have type `string`")
		}

		valueField := inMessage.Field[1]
		if valueField.GetName() != "value" {
			return fmt.Errorf("second map entry field must be called `value`")
		}
	}

	err = fg.GenerateMessageDefinition(prefix, inMessage)
	if err != nil {
		return err
	}

	for _, inEnum := range inMessage.GetEnumType() {
		err = fg.GenerateEnumDefinition(newPrefix, inEnum)
		if err != nil {
			return err
		}
	}

	err = fg.GenerateMessageDecoder(prefix, inMessage)
	if err != nil {
		return err
	}

	for _, inEnum := range inMessage.GetEnumType() {
		err = fg.GenerateEnumDecoder(newPrefix, inEnum)
		if err != nil {
			return err
		}
	}

	err = fg.GenerateMessageEncoder(prefix, inMessage)
	if err != nil {
		return err
	}

	for _, inEnum := range inMessage.GetEnumType() {
		err = fg.GenerateEnumEncoder(newPrefix, inEnum)
		if err != nil {
			return err
		}
	}

	// Nested messages.
	for _, nested := range inMessage.GetNestedType() {
		err = fg.GenerateEverything(newPrefix, nested)
		if err != nil {
			return err
		}
	}

	return nil
}

func elmTypeName(in string) string {
	n := camelCase(in)
	if reservedKeywords[n] {
		n += "_"
	}
	return n
}

func elmFieldName(in string) string {
	n := firstLower(camelCase(in))
	if reservedKeywords[n] {
		n += "_"
	}
	return n
}

func elmEnumValueName(in string) string {
	return camelCase(strings.ToLower(in))
}

func defaultEnumValue(typeName string) string {
	return firstLower(typeName) + "Default"
}

func encoderName(typeName string) string {
	return firstLower(typeName) + "Encoder"
}

func decoderName(typeName string) string {
	return firstLower(typeName) + "Decoder"
}

func elmFieldType(field *descriptor.FieldDescriptorProto) string {
	inFieldName := field.GetTypeName()
	packageName, messageName := convert(inFieldName)

	// Since we are exposing everything from imported modules, we do not use the package name at
	// all here.
	// TODO: Change this.
	packageName = ""

	if packageName == "" {
		return messageName
	} else {
		return packageName + "." + messageName
	}
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
