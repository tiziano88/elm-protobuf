package main

import "github.com/golang/protobuf/protoc-gen-go/descriptor"

func (fg *FileGenerator) GenerateOneofDefinition(prefix string, inMessage *descriptor.DescriptorProto, oneofIndex int) error {
	inOneof := inMessage.GetOneofDecl()[oneofIndex]

	// TODO: Prefix with message name to avoid collisions.
	oneofTypeName := elmOneofTypeName(inOneof)
	fg.P("type %s", oneofTypeName)

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
	fg.P("")

	return nil
}

func (fg *FileGenerator) GenerateOneofDecoder(prefix string, inMessage *descriptor.DescriptorProto, oneofIndex int) error {
	inOneof := inMessage.GetOneofDecl()[oneofIndex]

	// TODO: Prefix with message name to avoid collisions.
	typeName := elmOneofTypeName(inOneof)
	decoderName := elmOneofDecoderName(inOneof)

	fg.P("%s : JD.Decoder %s", decoderName, typeName)
	fg.P("%s =", decoderName)

	fg.In()

	fg.P("JD.oneOf")
	fg.In()

	leading := "["
	for _, inField := range inMessage.GetField() {
		if inField.OneofIndex != nil && inField.GetOneofIndex() == int32(oneofIndex) {
			oneofVariantName := elmTypeName(inField.GetName())
			decoderName := fieldElmDecoderName(inField)
			fg.P("%s JD.map %s (%q := %s)", leading, oneofVariantName, inField.GetJsonName(), decoderName)
			leading = ","
		}
	}
	fg.P("]")
	fg.Out()
	fg.Out()
	fg.P("")

	return nil
}

func (fg *FileGenerator) GenerateOneofEncoder(prefix string, inMessage *descriptor.DescriptorProto, oneofIndex int) error {
	return nil
}
