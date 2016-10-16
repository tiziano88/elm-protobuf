package main

import "github.com/golang/protobuf/protoc-gen-go/descriptor"

func (fg *FileGenerator) GenerateOneofDefinition(prefix string, inMessage *descriptor.DescriptorProto, oneofIndex int) error {
	inOneof := inMessage.GetOneofDecl()[oneofIndex]

	// TODO: Prefix with message name to avoid collisions.
	oneofType := oneofType(inOneof)

	fg.P("")
	fg.P("")
	fg.P("type %s", oneofType)
	{
		fg.In()

		leading := "="
		{
			oneofVariantName := oneofUnspecifiedValue(inOneof)
			fg.P("%s %s", leading, oneofVariantName)
			leading = "|"
		}
		for _, inField := range inMessage.GetField() {
			if inField.OneofIndex != nil && inField.GetOneofIndex() == int32(oneofIndex) {

				oneofVariantName := elmTypeName(inField.GetName())
				oneofArgumentType := fieldElmType(inField)
				fg.P("%s %s %s", leading, oneofVariantName, oneofArgumentType)

				leading = "|"
			}
		}
		fg.Out()
	}

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
	{
		fg.In()
		fg.P("lazy <| \\_ -> JD.oneOf")
		{
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
			fg.P("%s JD.succeed %s", leading, oneofUnspecifiedValue(inOneof))
			fg.P("]")
			fg.Out()
		}
		fg.Out()
	}

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
	fg.P("%s : %s -> Maybe (String, JE.Value)", encoderName, oneofType)
	fg.P("%s %s =", encoderName, argName)
	{
		fg.In()
		fg.P("case %s of", argName)
		{
			fg.In()

			valueName := "x"
			{
				oneofVariantName := oneofUnspecifiedValue(inOneof)
				fg.P("%s -> Nothing", oneofVariantName)
			}
			// TODO: Evaluate them in reverse order, as per
			// https://developers.google.com/protocol-buffers/docs/proto3#oneof
			for _, inField := range inMessage.GetField() {
				if inField.OneofIndex != nil && inField.GetOneofIndex() == int32(oneofIndex) {
					oneofVariantName := elmTypeName(inField.GetName())
					e := fieldEncoderName(inField)
					fg.P("%s %s -> Just (%q, %s %s)", oneofVariantName, valueName, inField.GetJsonName(), e, valueName)
				}
			}
			fg.Out()
		}
		fg.Out()
	}

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

func oneofUnspecifiedValue(inOneof *descriptor.OneofDescriptorProto) string {
	return elmTypeName(inOneof.GetName() + "_unspecified")
}
