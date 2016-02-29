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
	shortModuleName := firstUpper(strings.TrimSuffix(inFileFile, ".proto"))
	fullModuleName := strings.Replace(inFileDir, "/", ".", -1) + shortModuleName
	outFileName := filepath.Join(inFileDir, shortModuleName+".elm")
	outFile.Name = proto.String(outFileName)

	fg := NewFileGenerator()

	fg.GenerateModule(fullModuleName)
	fg.GenerateImports()
	fg.GenerateRuntime()

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

	fg.P("optionalFieldDecoder : JD.Decoder a -> String -> JD.Decoder (Maybe a)")
	fg.P("optionalFieldDecoder decoder name =")
	fg.In()
	fg.P("optionalDecoder (name := decoder)")
	fg.Out()

	fg.P("")
	fg.P("")

	fg.P("repeatedFieldDecoder : JD.Decoder a -> String -> JD.Decoder (List a)")
	fg.P("repeatedFieldDecoder decoder name =")
	fg.In()
	fg.P("JD.list (name := decoder)")
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

	fg.P("intFieldDecoder : String -> JD.Decoder Int")
	fg.P("intFieldDecoder name =")
	fg.In()
	fg.P("withDefault 0 (name := JD.int)")
	fg.Out()

	fg.P("")
	fg.P("")

	fg.P("floatFieldDecoder : String -> JD.Decoder Float")
	fg.P("floatFieldDecoder name =")
	fg.In()
	fg.P("withDefault 0.0 (name := JD.float)")
	fg.Out()

	fg.P("")
	fg.P("")

	fg.P("boolFieldDecoder : String -> JD.Decoder Bool")
	fg.P("boolFieldDecoder name =")
	fg.In()
	fg.P("withDefault False (name := JD.bool)")
	fg.Out()

	fg.P("")
	fg.P("")

	fg.P("stringFieldDecoder : String -> JD.Decoder String")
	fg.P("stringFieldDecoder name =")
	fg.In()
	fg.P("withDefault \"\" (name := JD.string)")
	fg.Out()

	fg.P("")
	fg.P("")

	fg.P("enumFieldDecoder : JD.Decoder a -> String -> JD.Decoder a")
	fg.P("enumFieldDecoder decoder name =")
	fg.In()
	fg.P("(name := decoder)")
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

func (fg *FileGenerator) GenerateMessage(inMessage *descriptor.DescriptorProto) error {
	typeName := inMessage.GetName()
	fg.P("type alias %s =", typeName)
	fg.In()

	first := true
	for _, inField := range inMessage.GetField() {
		optional := (inField.GetLabel() == descriptor.FieldDescriptorProto_LABEL_OPTIONAL) &&
			(inField.GetType() == descriptor.FieldDescriptorProto_TYPE_MESSAGE)
		repeated := inField.GetLabel() == descriptor.FieldDescriptorProto_LABEL_REPEATED

		t := ""
		switch inField.GetType() {
		case descriptor.FieldDescriptorProto_TYPE_INT64,
			descriptor.FieldDescriptorProto_TYPE_INT32,
			descriptor.FieldDescriptorProto_TYPE_UINT32,
			descriptor.FieldDescriptorProto_TYPE_UINT64,
			descriptor.FieldDescriptorProto_TYPE_SINT32,
			descriptor.FieldDescriptorProto_TYPE_SINT64:
			t = "Int"
		case descriptor.FieldDescriptorProto_TYPE_FLOAT:
			t = "Float"
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
		default:
			t = ">>>ERROR: " + inField.GetType().String()
		}

		leading := ""
		if first {
			leading = "{"
		} else {
			leading = ","
		}

		if repeated {
			fg.P("%s %s : List %s", leading, elmFieldName(inField.GetName()), t)
		} else {
			if optional {
				fg.P("%s %s : Maybe %s", leading, elmFieldName(inField.GetName()), t)
			} else {
				fg.P("%s %s : %s", leading, elmFieldName(inField.GetName()), t)
			}
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
		optional := (inField.GetLabel() == descriptor.FieldDescriptorProto_LABEL_OPTIONAL) &&
			(inField.GetType() == descriptor.FieldDescriptorProto_TYPE_MESSAGE)
		repeated := inField.GetLabel() == descriptor.FieldDescriptorProto_LABEL_REPEATED
		d := ""
		switch inField.GetType() {
		case descriptor.FieldDescriptorProto_TYPE_INT64,
			descriptor.FieldDescriptorProto_TYPE_INT32,
			descriptor.FieldDescriptorProto_TYPE_UINT32,
			descriptor.FieldDescriptorProto_TYPE_UINT64,
			descriptor.FieldDescriptorProto_TYPE_SINT32,
			descriptor.FieldDescriptorProto_TYPE_SINT64:
			// TODO: Handle parsing from string (for 64 bit types).
			d = "intFieldDecoder"
		case descriptor.FieldDescriptorProto_TYPE_FLOAT:
			d = "floatFieldDecoder"
		case descriptor.FieldDescriptorProto_TYPE_BOOL:
			d = "boolFieldDecoder"
		case descriptor.FieldDescriptorProto_TYPE_STRING:
			d = "stringFieldDecoder"
		case descriptor.FieldDescriptorProto_TYPE_ENUM:
			// TODO: Default enum value.
			// Remove leading ".".
			d = "(enumFieldDecoder " + decoderName(inField.GetTypeName()[1:]) + ")"
		case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
			// Remove leading ".".
			d = decoderName(inField.GetTypeName()[1:])
		default:
			d = ">>>ERROR: " + inField.GetType().String()
		}

		if repeated {
			fg.P("(repeatedFieldDecoder %s %q)", d, jsonFieldName(inField.GetName()))
		} else {
			if optional {
				fg.P("(optionalFieldDecoder %s %q)", d, jsonFieldName(inField.GetName()))
			} else {
				fg.P("(%s %q)", d, jsonFieldName(inField.GetName()))
			}
		}
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
		optional := (inField.GetLabel() == descriptor.FieldDescriptorProto_LABEL_OPTIONAL) &&
			(inField.GetType() == descriptor.FieldDescriptorProto_TYPE_MESSAGE)
		repeated := inField.GetLabel() == descriptor.FieldDescriptorProto_LABEL_REPEATED
		d := ""
		switch inField.GetType() {
		case descriptor.FieldDescriptorProto_TYPE_INT64,
			descriptor.FieldDescriptorProto_TYPE_INT32,
			descriptor.FieldDescriptorProto_TYPE_UINT32,
			descriptor.FieldDescriptorProto_TYPE_UINT64,
			descriptor.FieldDescriptorProto_TYPE_SINT32,
			descriptor.FieldDescriptorProto_TYPE_SINT64:
			d = "JE.int"
		case descriptor.FieldDescriptorProto_TYPE_FLOAT:
			d = "JE.float"
		case descriptor.FieldDescriptorProto_TYPE_BOOL:
			d = "JE.bool"
		case descriptor.FieldDescriptorProto_TYPE_STRING:
			d = "JE.string"
		case descriptor.FieldDescriptorProto_TYPE_ENUM:
			// Remove leading ".".
			d = encoderName(inField.GetTypeName()[1:])
		case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
			// Remove leading ".".
			d = encoderName(inField.GetTypeName()[1:])
		default:
			d = ">>>ERROR: " + inField.GetType().String()
		}

		leading := ""
		if first {
			leading = "["
		} else {
			leading = ","
		}
		first = false

		val := argName + "." + elmFieldName(inField.GetName())
		if repeated {
			fg.P("%s (%q, repeatedFieldEncoder %s %s)", leading, jsonFieldName(inField.GetName()), d, val)
		} else {
			if optional {
				fg.P("%s (%q, optionalEncoder %s %s)", leading, jsonFieldName(inField.GetName()), d, val)
			} else {
				fg.P("%s (%q, %s %s)", leading, jsonFieldName(inField.GetName()), d, val)
			}
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
