package main

import "github.com/golang/protobuf/protoc-gen-go/descriptor"

func oneofEncoderName(inOneof *descriptor.OneofDescriptorProto) string {
	typeName := elmTypeName(inOneof.GetName())
	return encoderName(typeName)
}
