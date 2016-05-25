package main

import "github.com/golang/protobuf/protoc-gen-go/descriptor"

func (fg *FileGenerator) GenerateOneofDefinition(prefix string, inMessage *descriptor.DescriptorProto, oneofIndex int) error {
	inOneof := inMessage.GetOneofDecl()[oneofIndex]

	// TODO: Prefix with message name to avoid collisions.
	oneofType := oneofType(inOneof)

	fg.P("")
	fg.P("")
	fg.P("type %s", oneofType)

	fg.In()

	leading := "="
	for _, inField := range inMessage.GetField() {
		if inField.OneofIndex != nil && inField.GetOneofIndex() == int32(oneofIndex) {

			oneofVariantName := elmTypeName(inField.GetName())
			oneofArgumentType := fieldElmType(inField)
			fg.P("%s %s %s", leading, oneofVariantName, oneofArgumentType)

			leading = "|"
		}
	}
	fg.Out()

	return nil
}

func (fg *FileGenerator) GenerateOneofDecoder(prefix string, inMessage *descriptor.DescriptorProto, oneofIndex int) error {
	inOneof := inMessage.GetOneofDecl()[oneofIndex]

	// TODO: Prefix with message name to avoid collisions.
	oneofType := oneofType(inOneof)
	decoderName := oneofDecoderName(inOneof)

	fg.P("")
	fg.P("")
	fg.P("%s : JD.Decoder %s", decoderName, oneofType)
	fg.P("%s =", decoderName)

	fg.In()

	fg.P("JD.oneOf")
	fg.In()

	leading := "["
	for _, inField := range inMessage.GetField() {
		if inField.OneofIndex != nil && inField.GetOneofIndex() == int32(oneofIndex) {
			oneofVariantName := elmTypeName(inField.GetName())
			decoderName := fieldDecoderName(inField)
			fg.P("%s JD.map %s (%q := %s)", leading, oneofVariantName, inField.GetJsonName(), decoderName)
			leading = ","
		}
	}
	fg.P("]")
	fg.Out()
	fg.Out()

	return nil
}

func (fg *FileGenerator) GenerateOneofEncoder(prefix string, inMessage *descriptor.DescriptorProto, oneofIndex int) error {
	inOneof := inMessage.GetOneofDecl()[oneofIndex]

	// TODO: Prefix with message name to avoid collisions.
	oneofType := oneofType(inOneof)
	encoderName := oneofEncoderName(inOneof)
	argName := "v"

	fg.P("")
	fg.P("")
	fg.P("%s : %s -> (String, JE.Value)", encoderName, oneofType)
	fg.P("%s %s =", encoderName, argName)

	fg.In()
	fg.P("case %s of", argName)
	fg.In()

	valueName := "x"
	for _, inField := range inMessage.GetField() {
		if inField.OneofIndex != nil && inField.GetOneofIndex() == int32(oneofIndex) {
			oneofVariantName := elmTypeName(inField.GetName())
			e := fieldEncoderName(inField)
			fg.P("%s %s -> (%q, %s %s)", oneofVariantName, valueName, inField.GetJsonName(), e, valueName)
		}
	}
	fg.Out()
	fg.Out()

	return nil
}

func oneofDecoderName(inOneof *descriptor.OneofDescriptorProto) string {
	typeName := elmTypeName(inOneof.GetName())
	return decoderName(typeName)
}

func oneofEncoderName(inOneof *descriptor.OneofDescriptorProto) string {
	typeName := elmTypeName(inOneof.GetName())
	return encoderName(typeName)
}

func oneofType(inOneof *descriptor.OneofDescriptorProto) string {
	return elmTypeName(inOneof.GetName())
}
