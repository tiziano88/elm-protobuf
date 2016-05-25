package main

import "github.com/golang/protobuf/protoc-gen-go/descriptor"

func (fg *FileGenerator) GenerateEnumDefinition(prefix string, inEnum *descriptor.EnumDescriptorProto) error {
	typeName := prefix + inEnum.GetName()
	fg.P("")
	fg.P("")
	fg.P("type %s", typeName)
	fg.In()
	leading := "="
	for _, enumValue := range inEnum.GetValue() {
		// TODO: Convert names to CamelCase.
		fg.P("%s %s -- %d", leading, prefix+elmEnumValueName(enumValue.GetName()), enumValue.GetNumber())
		leading = "|"
	}
	fg.Out()
	return nil
}

func (fg *FileGenerator) GenerateEnumDecoder(prefix string, inEnum *descriptor.EnumDescriptorProto) error {
	typeName := prefix + inEnum.GetName()
	decoderName := decoderName(typeName)
	fg.P("")
	fg.P("")
	fg.P("%s : JD.Decoder %s", decoderName, typeName)
	fg.P("%s =", decoderName)
	fg.In()
	fg.P("let")
	fg.In()
	fg.P("lookup s = case s of")
	fg.In()
	for _, enumValue := range inEnum.GetValue() {
		fg.P("%q -> %s", enumValue.GetName(), prefix+elmEnumValueName(enumValue.GetName()))
	}
	// TODO: This should fail instead.
	fg.P("_ -> %s", prefix+elmEnumValueName(inEnum.GetValue()[0].GetName()))
	fg.Out()
	fg.Out()
	fg.P("in")
	fg.In()
	fg.P("JD.map lookup JD.string")
	fg.Out()
	fg.Out()

	defaultName := defaultEnumValue(typeName)
	fg.P("")
	fg.P("")
	fg.P("%s : %s", defaultName, typeName)
	fg.P("%s = %s", defaultName, prefix+elmEnumValueName(inEnum.GetValue()[0].GetName()))
	return nil
}

func (fg *FileGenerator) GenerateEnumEncoder(prefix string, inEnum *descriptor.EnumDescriptorProto) error {
	typeName := prefix + inEnum.GetName()
	argName := "v"
	fg.P("")
	fg.P("")
	fg.P("%s : %s -> JE.Value", encoderName(typeName), typeName)
	fg.P("%s %s =", encoderName(typeName), argName)
	fg.In()
	fg.P("let")
	fg.In()
	fg.P("lookup s = case s of")
	fg.In()
	for _, enumValue := range inEnum.GetValue() {
		fg.P("%s -> %q", prefix+elmEnumValueName(enumValue.GetName()), enumValue.GetName())
	}
	fg.Out()
	fg.Out()
	fg.P("in")
	fg.In()
	fg.P("JE.string <| lookup %s", argName)
	fg.Out()
	fg.Out()
	return nil
}
