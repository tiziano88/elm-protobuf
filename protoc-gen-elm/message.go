package main

import (
	"fmt"

	"github.com/golang/protobuf/protoc-gen-go/descriptor"
)

func (fg *FileGenerator) GenerateMessageDefinition(prefix string, inMessage *descriptor.DescriptorProto) error {
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

	for _, inOneof := range inMessage.GetOneofDecl() {

		oneofName := elmFieldName(inOneof.GetName())
		// TODO: Prefix with message name to avoid collisions.
		oneofTypeName := elmTypeName(inOneof.GetName())
		fg.P("%s %s : %s", leading, oneofName, oneofTypeName)

		leading = ","
	}

	fg.P("}")
	fg.Out()

	for i, _ := range inMessage.GetOneofDecl() {
		fg.P("")
		fg.GenerateOneofDefinition(prefix, inMessage, i)
		fg.GenerateOneofDecoder(prefix, inMessage, i)
		fg.GenerateOneofEncoder(prefix, inMessage, i)
	}

	return nil
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
