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

func fieldDefaultValue(inField *descriptor.FieldDescriptorProto) (string, error) {
	if inField.GetLabel() == descriptor.FieldDescriptorProto_LABEL_REPEATED {
		return "[]", nil
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
		return "0", nil
	case descriptor.FieldDescriptorProto_TYPE_FLOAT,
		descriptor.FieldDescriptorProto_TYPE_DOUBLE:
		return "0.0", nil
	case descriptor.FieldDescriptorProto_TYPE_BOOL:
		return "False", nil
	case descriptor.FieldDescriptorProto_TYPE_STRING:
		return "\"\"", nil
	case descriptor.FieldDescriptorProto_TYPE_ENUM:
		// TODO: Default enum value.
		_, messageName := convert(inField.GetTypeName())
		return defaultEnumValue(messageName), nil
	case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
		return "xxx", nil
	case descriptor.FieldDescriptorProto_TYPE_BYTES:
		return "[]", nil
	default:
		return "", fmt.Errorf("error generating decoder for field %s", inField.GetType())
	}
}
