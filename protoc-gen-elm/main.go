package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

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
		outFile, _ := processFile(inFile)
		resp.File = append(resp.File, outFile)
	}

	data, err = proto.Marshal(resp)
	if err != nil {
		log.Fatalf("Could not marshal response: %v", err)
	}

	_, err = os.Stdout.Write(data)
	if err != nil {
		log.Fatalf("Could not write response to STDOUT: %v", err)
	}
}

func processFile(inFile *descriptor.FileDescriptorProto) (*plugin.CodeGeneratorResponse_File, error) {
	outFile := &plugin.CodeGeneratorResponse_File{}

	inFileName := inFile.GetName()

	inFileDir, inFileFile := filepath.Split(inFileName)
	moduleName := firstUpper(strings.TrimSuffix(inFileFile, ".proto"))
	outFileName := filepath.Join(inFileDir, moduleName+".elm")
	outFile.Name = proto.String(outFileName)

	fg := NewFileGenerator()

	fg.GenerateModule(moduleName)

	fg.P("")
	fg.P("")

	fg.GenerateImports()

	fg.P("")
	fg.P("")

	fg.GenerateRuntime()

	fg.P("")
	fg.P("")

	var err error

	for _, inEnum := range inFile.GetEnumType() {
		err = fg.GenerateEnum(inEnum)
		if err != nil {
			return nil, err
		}

		fg.P("")
		fg.P("")

		err = fg.GenerateEnumDecoder(inEnum)
		if err != nil {
			return nil, err
		}

		fg.P("")
		fg.P("")

		err = fg.GenerateEnumEncoder(inEnum)
		if err != nil {
			return nil, err
		}

		fg.P("")
		fg.P("")
	}

	for _, inMessage := range inFile.GetMessageType() {
		err = fg.GenerateMessage(inMessage)
		if err != nil {
			return nil, err
		}

		fg.P("")
		fg.P("")

		err = fg.GenerateMessageDecoder(inMessage)
		if err != nil {
			return nil, err
		}

		fg.P("")
		fg.P("")

		err = fg.GenerateMessageEncoder(inMessage)
		if err != nil {
			return nil, err
		}

		fg.P("")
		fg.P("")
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

func (fg *FileGenerator) GenerateEnum(inEnum *descriptor.EnumDescriptorProto) error {
	typeName := inEnum.GetName()
	fg.P("type %s", typeName)
	fg.In()
	first := true
	for _, enumValue := range inEnum.GetValue() {
		leading := ""
		if first {
			leading = "="
		} else {
			leading = "|"
		}
		first = false
		// TODO: Convert names to CamelCase.
		fg.P("%s %s -- %d", leading, elmEnumValueName(enumValue.GetName()), enumValue.GetNumber())
	}
	fg.Out()
	return nil
}

func (fg *FileGenerator) GenerateEnumDecoder(inEnum *descriptor.EnumDescriptorProto) error {
	typeName := inEnum.GetName()
	decoderName := decoderName(typeName)
	fg.P("%s : JD.Decoder %s", decoderName, typeName)
	fg.P("%s =", decoderName)
	fg.In()
	fg.P("let")
	fg.In()
	fg.P("lookup s = case s of")
	fg.In()
	for _, enumValue := range inEnum.GetValue() {
		fg.P("%q -> %s", enumValue.GetName(), elmEnumValueName(enumValue.GetName()))
	}
	// TODO: This should fail instead.
	fg.P("_ -> %s", elmEnumValueName(inEnum.GetValue()[0].GetName()))
	fg.Out()
	fg.Out()
	fg.P("in")
	fg.In()
	fg.P("JD.map lookup JD.string")
	fg.Out()
	fg.Out()
	// TODO: Implement this.
	return nil
}

func (fg *FileGenerator) GenerateEnumEncoder(inEnum *descriptor.EnumDescriptorProto) error {
	typeName := inEnum.GetName()
	argName := "v"
	fg.P("%s : %s -> JE.Value", encoderName(typeName), typeName)
	fg.P("%s %s =", encoderName(typeName), argName)
	fg.In()
	fg.P("let")
	fg.In()
	fg.P("lookup s = case s of")
	fg.In()
	for _, enumValue := range inEnum.GetValue() {
		fg.P("%s -> %q", elmEnumValueName(enumValue.GetName()), enumValue.GetName())
	}
	fg.Out()
	fg.Out()
	fg.P("in")
	fg.In()
	fg.P("JE.string <| lookup %s", argName)
	fg.Out()
	fg.Out()
	// TODO: Implement this.
	return nil
}

func (fg *FileGenerator) GenerateModule(moduleName string) {
	fg.P("module %s where", moduleName)
}

func (fg *FileGenerator) GenerateImports() {
	fg.P("import Json.Decode as JD exposing ((:=))")
	fg.P("import Json.Encode as JE")
}

func (fg *FileGenerator) GenerateRuntime() {
	fg.P("optional : JD.Decoder a -> JD.Decoder (Maybe a)")
	fg.P("optional decoder =")
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

	fg.P("intField : String -> JD.Decoder Int")
	fg.P("intField name =")
	fg.In()
	fg.P("withDefault 0 (name := JD.int)")
	fg.Out()

	fg.P("")
	fg.P("")

	fg.P("boolField : String -> JD.Decoder Bool")
	fg.P("boolField name =")
	fg.In()
	fg.P("withDefault False (name := JD.bool)")
	fg.Out()

	fg.P("")
	fg.P("")

	fg.P("stringField : String -> JD.Decoder String")
	fg.P("stringField name =")
	fg.In()
	fg.P("withDefault \"\" (name := JD.string)")
	fg.Out()

	fg.P("")
	fg.P("")

	fg.P("messageField : JD.Decoder a -> String -> JD.Decoder (Maybe a)")
	fg.P("messageField decoder name =")
	fg.In()
	fg.P("optional (name := decoder)")
	fg.Out()

	fg.P("")
	fg.P("")

	fg.P("enumField : JD.Decoder a -> String -> JD.Decoder a")
	fg.P("enumField decoder name =")
	fg.In()
	fg.P("(name := decoder)")
	fg.Out()
}

func (fg *FileGenerator) GenerateMessage(inMessage *descriptor.DescriptorProto) error {
	typeName := inMessage.GetName()
	fg.P("type alias %s =", typeName)
	fg.In()

	first := true
	for _, inField := range inMessage.GetField() {
		optional := false
		t := ""
		switch inField.GetType() {
		case descriptor.FieldDescriptorProto_TYPE_INT64,
			descriptor.FieldDescriptorProto_TYPE_INT32,
			descriptor.FieldDescriptorProto_TYPE_UINT32,
			descriptor.FieldDescriptorProto_TYPE_UINT64,
			descriptor.FieldDescriptorProto_TYPE_SINT32,
			descriptor.FieldDescriptorProto_TYPE_SINT64:
			t = "Int"
		case descriptor.FieldDescriptorProto_TYPE_BOOL:
			t = "Bool"
		case descriptor.FieldDescriptorProto_TYPE_STRING:
			t = "String"
		case descriptor.FieldDescriptorProto_TYPE_ENUM:
			// XXX
			t = inField.GetTypeName()[1:]
		case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
			// XXX
			t = inField.GetTypeName()[1:]
			optional = true
		default:
			t = ">>>ERROR" + inField.GetType().String()
		}

		leading := ""
		if first {
			leading = "{"
		} else {
			leading = ","
		}

		if optional {
			fg.P("%s %s : Maybe %s", leading, elmFieldName(inField.GetName()), t)
		} else {
			fg.P("%s %s : %s", leading, elmFieldName(inField.GetName()), t)
		}
		first = false
	}
	fg.P("}")
	fg.Out()
	return nil
}

func (fg *FileGenerator) GenerateMessageDecoder(inMessage *descriptor.DescriptorProto) error {
	typeName := inMessage.GetName()
	fg.P("%s : JD.Decoder %s", decoderName(typeName), typeName)
	fg.P("%s =", decoderName(typeName))
	fg.In()
	fg.P("JD.object%d %s", len(inMessage.GetField()), typeName)
	fg.In()
	for _, inField := range inMessage.GetField() {
		d := ""
		switch inField.GetType() {
		case descriptor.FieldDescriptorProto_TYPE_INT64,
			descriptor.FieldDescriptorProto_TYPE_INT32,
			descriptor.FieldDescriptorProto_TYPE_UINT32,
			descriptor.FieldDescriptorProto_TYPE_UINT64,
			descriptor.FieldDescriptorProto_TYPE_SINT32,
			descriptor.FieldDescriptorProto_TYPE_SINT64:
			// TODO: Handle parsing from string (for 64 bit types).
			d = "intField"
		case descriptor.FieldDescriptorProto_TYPE_BOOL:
			d = "boolField"
		case descriptor.FieldDescriptorProto_TYPE_STRING:
			d = "stringField"
		case descriptor.FieldDescriptorProto_TYPE_ENUM:
			// TODO: Default enum value.
			// Remove leading ".".
			d = "enumField " + decoderName(inField.GetTypeName()[1:])
		case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
			// Remove leading ".".
			d = "messageField " + decoderName(inField.GetTypeName()[1:])
		default:
			d = "xxx"
		}
		fg.P("(%s %q)", d, jsonFieldName(inField.GetName()))
	}
	fg.Out()
	fg.Out()
	return nil
}

func (fg *FileGenerator) GenerateMessageEncoder(inMessage *descriptor.DescriptorProto) error {
	typeName := inMessage.GetName()
	argName := "v"
	fg.P("%s : %s -> JE.Value", encoderName(typeName), typeName)
	fg.P("%s %s =", encoderName(typeName), argName)
	fg.In()
	fg.P("JE.object")
	fg.In()

	first := true
	for _, inField := range inMessage.GetField() {
		optional := false
		d := ""
		switch inField.GetType() {
		case descriptor.FieldDescriptorProto_TYPE_INT64,
			descriptor.FieldDescriptorProto_TYPE_INT32,
			descriptor.FieldDescriptorProto_TYPE_UINT32,
			descriptor.FieldDescriptorProto_TYPE_UINT64,
			descriptor.FieldDescriptorProto_TYPE_SINT32,
			descriptor.FieldDescriptorProto_TYPE_SINT64:
			d = "JE.int"
		case descriptor.FieldDescriptorProto_TYPE_BOOL:
			d = "JE.bool"
		case descriptor.FieldDescriptorProto_TYPE_STRING:
			d = "JE.string"
		case descriptor.FieldDescriptorProto_TYPE_ENUM:
			// Remove leading ".".
			d = encoderName(inField.GetTypeName()[1:])
		case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
			optional = true
			// Remove leading ".".
			d = encoderName(inField.GetTypeName()[1:])
		default:
			d = "xxx"
		}

		leading := ""
		if first {
			leading = "["
		} else {
			leading = ","
		}
		first = false

		val := argName + "." + elmFieldName(inField.GetName())
		if optional {
			// TODO
		} else {
			fg.P("%s (%q, %s %s)", leading, jsonFieldName(inField.GetName()), d, val)
		}
	}
	fg.P("]")
	fg.Out()
	fg.Out()
	return nil
}

func elmFieldName(in string) string {
	return firstLower(camelCase(in))
}

func elmEnumValueName(in string) string {
	return camelCase(strings.ToLower(in))
}

func decoderName(typeName string) string {
	return firstLower(typeName) + "Decoder"
}

func encoderName(typeName string) string {
	return firstLower(typeName) + "Encoder"
}

func jsonFieldName(fieldName string) string {
	// TODO: Make sure this is fine.
	return firstLower(camelCase(fieldName))
}

func firstLower(in string) string {
	return strings.ToLower(string(in[0])) + string(in[1:])
}

func firstUpper(in string) string {
	return strings.ToUpper(string(in[0])) + string(in[1:])
}

func camelCase(in string) string {
	// Remove any additional underscores, e.g. convert `foo_1` into `foo1`.
	return strings.Replace(generator.CamelCase(in), "_", "", -1)
}
