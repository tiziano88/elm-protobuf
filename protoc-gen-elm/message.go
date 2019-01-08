package main

import (
	"fmt"

	"github.com/golang/protobuf/protoc-gen-go/descriptor"
)

func (fg *FileGenerator) isOptional(
	msgName string, inField *descriptor.FieldDescriptorProto, fieldOptions *FieldOptions,
) bool {
	return (inField.GetLabel() == descriptor.FieldDescriptorProto_LABEL_OPTIONAL) &&
		(inField.GetType() == descriptor.FieldDescriptorProto_TYPE_MESSAGE) ||
		(fieldOptions != nil && !fieldOptions.Required)

}

func (fg *FileGenerator) GenerateMessageDefinition(prefix string, inMessage *descriptor.DescriptorProto) error {
	typeName := prefix + inMessage.GetName()

	fg.P("")
	fg.P("")
	fg.P("type alias %s =", typeName)
	{
		fg.In()

		leading := "{"

		if len(inMessage.GetField()) == 0 {
			fg.P(leading)
		}

		for _, inField := range inMessage.GetField() {
			if inField.OneofIndex != nil {
				// Handled in the oneof only.
				continue
			}
			fieldOptions := fg.FieldOptions(typeName + "." + inField.GetName())

			optional := fg.isOptional(inMessage.GetName(), inField, fieldOptions)
			repeated := inField.GetLabel() == descriptor.FieldDescriptorProto_LABEL_REPEATED

			fType := fieldElmType(inField)
			if fieldOptions != nil {
				fType = fieldOptions.Type
			}

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
	}

	for i, _ := range inMessage.GetOneofDecl() {
		fg.GenerateOneofDefinition(prefix, inMessage, i)
		fg.GenerateOneofDecoder(prefix, inMessage, i)
		fg.GenerateOneofEncoder(prefix, inMessage, i)
	}

	return nil
}

func (fg *FileGenerator) GenerateMessageDecoder(prefix string, inMessage *descriptor.DescriptorProto) error {
	typeName := prefix + inMessage.GetName()

	fg.P("")
	fg.P("")
	fg.P("%s : JD.Decoder %s", decoderName(typeName), typeName)
	fg.P("%s =", decoderName(typeName))
	{
		fg.In()
		fg.P("JD.lazy <| \\_ -> decode %s", typeName)
		{
			fg.In()

			for _, inField := range inMessage.GetField() {
				if inField.OneofIndex != nil {
					// Handled in the oneof only.
					continue
				}

				fieldOptions := fg.FieldOptions(typeName + "." + inField.GetName())

				optional := fg.isOptional(inMessage.GetName(), inField, fieldOptions)

				repeated := inField.GetLabel() == descriptor.FieldDescriptorProto_LABEL_REPEATED
				d := fieldDecoderName(inField)
				def := fieldDefaultValue(inField)

				if fieldOptions != nil {
					if t, ok := fg.options.Types[fieldOptions.Type]; ok && t.JSON.Decoder != "" {
						d = t.JSON.Decoder
					}
				}

				if repeated {
					fg.P("|> repeated %q %s", jsonFieldName(inField), d)
				} else {
					if optional {
						fg.P("|> optional %q %s", jsonFieldName(inField), d)
					} else {
						fg.P("|> required %q %s %s", jsonFieldName(inField), d, def)
					}
				}
			}

			for _, inOneof := range inMessage.GetOneofDecl() {
				oneofDecoderName := oneofDecoderName(inOneof)
				fg.P("|> field %s", oneofDecoderName)
			}

			fg.Out()
		}
		fg.Out()
	}
	return nil
}

func (fg *FileGenerator) GenerateMessageEncoder(prefix string, inMessage *descriptor.DescriptorProto) error {
	typeName := prefix + inMessage.GetName()
	argName := "v"

	fg.P("")
	fg.P("")
	fg.P("%s : %s -> JE.Value", encoderName(typeName), typeName)
	fg.P("%s %s =", encoderName(typeName), argName)
	{
		fg.In()
		fg.P("JE.object <| List.filterMap identity <|")
		{
			fg.In()

			leading := "["

			if len(inMessage.GetField()) == 0 {
				fg.P(leading)
			}

			for _, inField := range inMessage.GetField() {
				if inField.OneofIndex != nil {
					// Handled in the oneof only.
					continue
				}

				fieldOptions := fg.FieldOptions(typeName + "." + inField.GetName())

				optional := fg.isOptional(inMessage.GetName(), inField, fieldOptions)
				repeated := inField.GetLabel() == descriptor.FieldDescriptorProto_LABEL_REPEATED
				d := fieldEncoderName(inField)
				if fieldOptions != nil {
					if t, ok := fg.options.Types[fieldOptions.Type]; ok && t.JSON.Encoder != "" {
						d = t.JSON.Encoder
					}
				}
				val := argName + "." + elmFieldName(inField.GetName())
				def := fieldDefaultValue(inField)
				if repeated {
					fg.P("%s (repeatedFieldEncoder %q %s %s)", leading, jsonFieldName(inField), d, val)
				} else {
					if optional {
						fg.P("%s (optionalEncoder %q %s %s)", leading, jsonFieldName(inField), d, val)
					} else {
						fg.P("%s (requiredFieldEncoder %q %s %s %s)", leading, jsonFieldName(inField), d, def, val)
					}
				}

				leading = ","
			}

			for _, inOneof := range inMessage.GetOneofDecl() {
				val := argName + "." + elmFieldName(inOneof.GetName())
				oneofEncoderName := oneofEncoderName(inOneof)
				fg.P("%s (%s %s)", leading, oneofEncoderName, val)
				leading = ","
			}

			fg.P("]")

			fg.Out()
		}
		fg.Out()
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
	case descriptor.FieldDescriptorProto_TYPE_MESSAGE,
		descriptor.FieldDescriptorProto_TYPE_ENUM:
		// Well known types.
		if n, ok := excludedTypes[inField.GetTypeName()]; ok {
			return n
		}
		_, messageName := convert(inField.GetTypeName())
		return messageName
	case descriptor.FieldDescriptorProto_TYPE_BYTES:
		// XXX
		return "Bytes"
	default:
		// TODO: Return error.
		return fmt.Sprintf("Error generating type for field %q %s", inField.GetName(), inField.GetType())
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

func fieldDecoderName(inField *descriptor.FieldDescriptorProto) string {
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
		return "intDecoder"
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
		_, messageName := convert(inField.GetTypeName())
		return decoderName(messageName)
	case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
		// Well Known Types.
		if n, ok := excludedDecoders[inField.GetTypeName()]; ok {
			return n
		}
		_, messageName := convert(inField.GetTypeName())
		return decoderName(messageName)
	case descriptor.FieldDescriptorProto_TYPE_BYTES:
		return "bytesFieldDecoder"
	default:
		return fmt.Sprintf("Error generating decoder for field %s", inField.GetType())
	}
}

func fieldDefaultValue(inField *descriptor.FieldDescriptorProto) string {
	if inField.GetLabel() == descriptor.FieldDescriptorProto_LABEL_REPEATED {
		return "[]"
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
		_, messageName := convert(inField.GetTypeName())
		return defaultEnumValue(messageName)
	case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
		return "xxx"
	case descriptor.FieldDescriptorProto_TYPE_BYTES:
		return "[]"
	default:
		return fmt.Sprintf("Error generating decoder for field %s", inField.GetType())
	}
}
