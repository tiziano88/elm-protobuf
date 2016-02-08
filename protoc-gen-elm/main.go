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
	for _, inMessage := range inFile.GetMessageType() {
		fg.ProcessMessage(inMessage)
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

func (fg *FileGenerator) ProcessMessage(inMessage *descriptor.DescriptorProto) error {
	var err error

	fg.GenerateImports()

	fg.P("")

	err = fg.GenerateType(inMessage)
	if err != nil {
		return err
	}

	fg.P("")

	err = fg.GenerateDecoder(inMessage)
	if err != nil {
		return err
	}

	return nil
}

func (fg *FileGenerator) GenerateImports() {
	fg.P("import Json.Decode")
}

func (fg *FileGenerator) GenerateDecoder(inMessage *descriptor.DescriptorProto) error {
	messageName := inMessage.GetName()
	decoderName := strings.ToLower(messageName)
	typeName := messageName
	fg.P("%s : Decoder %s", decoderName, typeName)
	fg.P("%s =", decoderName)
	fg.In()
	fg.P("objectN %s", typeName)
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
			d = "int"
		case descriptor.FieldDescriptorProto_TYPE_BOOL:
			d = "bool"
		case descriptor.FieldDescriptorProto_TYPE_STRING:
			d = "string"
		default:
			d = "----"
		}
		fg.P("(%s := %s)", inField.GetName(), d)
	}
	fg.Out()
	fg.Out()
	return nil
}

func (fg *FileGenerator) GenerateType(inMessage *descriptor.DescriptorProto) error {
	typeName := inMessage.GetName()
	first := true
	fg.P("type alias %s =", typeName)
	fg.In()
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
		default:
			t = "----"
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
