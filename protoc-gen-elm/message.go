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
		d := fieldElmDecoderName(inField)
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
