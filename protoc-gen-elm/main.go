package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/golang/protobuf/proto"

	"github.com/golang/protobuf/protoc-gen-go/descriptor"
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
	outFile.Name = proto.String(strings.TrimSuffix(inFile.GetName(), ".proto") + ".elm")

	fg := NewFileGenerator()

	fg.GenerateImports()

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
	}

	fg.P("")
	fg.P("")

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
	}

	fg.P("")
	fg.P("")

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
	fg.P("type %s", inEnum.GetName())
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
		fg.P("%s %s -- %d", leading, underscoreToCamel(enumValue.GetName()), enumValue.GetNumber())
	}
	fg.Out()
	return nil
}

func (fg *FileGenerator) GenerateEnumDecoder(inEnum *descriptor.EnumDescriptorProto) error {
	decoderName := strings.ToLower(inEnum.GetName())
	typeName := inEnum.GetName()
	fg.P("%s : JD.Decoder %s", decoderName, typeName)
	fg.P("%s =", decoderName)
	fg.In()
	fg.P("let")
	fg.In()
	fg.P("lookup s = case s of")
	fg.In()
	for _, enumValue := range inEnum.GetValue() {
		fg.P("%q -> %s", enumValue.GetName(), underscoreToCamel(enumValue.GetName()))
	}
	fg.P("_ -> %s", underscoreToCamel(inEnum.GetValue()[0].GetName()))
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

func (fg *FileGenerator) GenerateImports() {
	fg.P("import Json.Decode as JD exposing ((:=))")
}

func (fg *FileGenerator) GenerateMessageDecoder(inMessage *descriptor.DescriptorProto) error {
	messageName := inMessage.GetName()
	decoderName := strings.ToLower(messageName)
	typeName := messageName
	fg.P("%s : JD.Decoder %s", decoderName, typeName)
	fg.P("%s =", decoderName)
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
			d = "JD.int"
		case descriptor.FieldDescriptorProto_TYPE_BOOL:
			d = "JD.bool"
		case descriptor.FieldDescriptorProto_TYPE_STRING:
			d = "JD.string"
		case descriptor.FieldDescriptorProto_TYPE_ENUM:
			// XXX
			d = strings.ToLower(inField.GetTypeName()[1:])
		default:
			d = "xxx"
		}
		fg.P("(%q := %s)", inField.GetName(), d)
	}
	fg.Out()
	fg.Out()
	return nil
}

func (fg *FileGenerator) GenerateMessage(inMessage *descriptor.DescriptorProto) error {
	typeName := inMessage.GetName()
	fg.P("type alias %s =", typeName)
	fg.In()

	first := true
	for _, inField := range inMessage.GetField() {
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
		default:
			t = "----" + inField.GetType().String()
		}
		leading := ""
		if first {
			leading = "{"
		} else {
			leading = ","
		}
		fg.P("%s %s : %s", leading, inField.GetName(), t)
		first = false
	}
	fg.P("}")
	fg.Out()
	return nil
}

func underscoreToCamel(in string) string {
	out := ""
	for _, segment := range strings.Split(in, "_") {
		out += camel(segment)
	}
	return out
}

func camel(in string) string {
	return strings.ToUpper(string(in[0])) + strings.ToLower(string(in[1:]))
}
