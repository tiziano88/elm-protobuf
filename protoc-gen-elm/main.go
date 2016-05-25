package main

import (
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
		err = fg.GenerateEverything("", inMessage)
		if err != nil {
			return nil, err
		}
	}

	outFile.Content = proto.String(fg.out.String())

	return outFile, nil
}

func (fg *FileGenerator) GenerateModule(moduleName string) {
	fg.P("module %s exposing (..)", moduleName)
}

func (fg *FileGenerator) GenerateImports() {
	fg.P("")
	fg.P("")
	fg.P("import Json.Decode as JD exposing ((:=))")
	fg.P("import Json.Encode as JE")
}

func (fg *FileGenerator) GenerateEverything(prefix string, inMessage *descriptor.DescriptorProto) error {
	newPrefix := prefix + inMessage.GetName() + "_"
	var err error

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
		fg.GenerateEverything(newPrefix, nested)
	}

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
