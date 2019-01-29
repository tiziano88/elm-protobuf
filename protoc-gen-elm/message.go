package main

import (
	"fmt"
	"strings"

	"github.com/golang/protobuf/protoc-gen-go/descriptor"
)

// mapEntries detects whether a field is implemented using a `map<,>` field or not.
// In order for `map<,>` fields to be supported by proto2 format, they
// get parsed as a backwards compatible form:
//
//     map<KeyType, ValueType> map_field = 1;
// Gets parsed as:
//     message MapFieldEntry {
//         option map_entry = true;
//         optional KeyType key = 1;
//         optional ValueType value = 2;
//     }
//     repeated MapFieldEntry map_field = 1;
//
// our code looks for the `map_entry` option to detect `map<,>` fields, and generate Dict's for them
// https://github.com/golang/protobuf/blob/882cf97a83ad205fd22af574246a3bc647d7a7d2/protoc-gen-go/descriptor/descriptor.proto#L474-L495
func mapEntries(inField *descriptor.FieldDescriptorProto, inMessage *descriptor.DescriptorProto) (isMap bool, keyFieldDescriptor *descriptor.FieldDescriptorProto, valueFieldDescriptor *descriptor.FieldDescriptorProto) {
	isRepeated :=
		inField.GetLabel() == descriptor.FieldDescriptorProto_LABEL_REPEATED &&
			inField.GetType() == descriptor.FieldDescriptorProto_TYPE_MESSAGE

	if !isRepeated {
		return false, nil, nil
	}

	fullyQualifiedTypeName := inField.GetTypeName()
	splitName := strings.Split(fullyQualifiedTypeName, ".")
	localTypeName := splitName[len(splitName)-1]

	for _, nested := range inMessage.GetNestedType() {
		if nested.GetName() == localTypeName && nested.GetOptions().GetMapEntry() {
			return true, nested.GetField()[0], nested.GetField()[1]
		}
	}
	return false, nil, nil
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

			optional := (inField.GetLabel() == descriptor.FieldDescriptorProto_LABEL_OPTIONAL) &&
				(inField.GetType() == descriptor.FieldDescriptorProto_TYPE_MESSAGE)
			repeated := inField.GetLabel() == descriptor.FieldDescriptorProto_LABEL_REPEATED

			isMapEntries, mapKeyFieldDescriptor, mapValueFieldDescriptor := mapEntries(inField, inMessage)

			fType := fieldElmType(inField)

			fName := elmFieldName(inField.GetName())
			fNumber := inField.GetNumber()

			if isMapEntries {
				fg.P("%s %s : Dict.Dict %s %s -- %d", leading, fName, fieldElmType(mapKeyFieldDescriptor), fieldElmType(mapValueFieldDescriptor), fNumber)
			} else if repeated {
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

				optional := (inField.GetLabel() == descriptor.FieldDescriptorProto_LABEL_OPTIONAL) &&
					(inField.GetType() == descriptor.FieldDescriptorProto_TYPE_MESSAGE)
				repeated := inField.GetLabel() == descriptor.FieldDescriptorProto_LABEL_REPEATED
				isMapEntries, _, mapValueFieldDescriptor := mapEntries(inField, inMessage)
				d := fieldDecoderName(inField)
				def := fieldDefaultValue(inField)

				if isMapEntries {
					fg.P("|> mapEntries %q %s", jsonFieldName(inField), fieldDecoderName(mapValueFieldDescriptor))
				} else if repeated {
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

				optional := (inField.GetLabel() == descriptor.FieldDescriptorProto_LABEL_OPTIONAL) &&
					(inField.GetType() == descriptor.FieldDescriptorProto_TYPE_MESSAGE)
				repeated := inField.GetLabel() == descriptor.FieldDescriptorProto_LABEL_REPEATED
				isMapEntries, _, mapValueFieldDescriptor := mapEntries(inField, inMessage)
				d := fieldEncoderName(inField)
				val := argName + "." + elmFieldName(inField.GetName())
				def := fieldDefaultValue(inField)

				if isMapEntries {
					fg.P("%s (mapEntriesFieldEncoder %q %s %s)", leading, jsonFieldName(inField), fieldEncoderName(mapValueFieldDescriptor), val)
				} else if repeated {
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
