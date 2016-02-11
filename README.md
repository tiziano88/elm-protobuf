# elm-protobuf

[![Build Status]
(https://travis-ci.org/tiziano88/elm-protobuf.svg?branch=master)]
(https://travis-ci.org/tiziano88/elm-protobuf)

Experimental protobuf plugin generating elm code from proto definitions.

The plugin itself is written in Go.

For a sample generated output file, see
https://github.com/tiziano88/elm-protobuf/blob/master/protoc-gen-elm/test/1/expected_output/test.elm

# How to install

First obtain the binary using `go get`:

`go get github.com/tiziano88/elm-protobuf/protoc-gen-elm`

This assumes that you have a Go environment correctly set up, and `$GOPATH/bin`
is included in your `$PATH`.

Then run the proto compiler specifying the `--elm-out` flag:

`protoc --elm-out=. *.proto`

`protoc` will automatically detect the `protoc-gen-elm` binary from your `$PATH`
and use it to generate the output elm code.

# References

https://developers.google.com/protocol-buffers/

https://developers.google.com/protocol-buffers/docs/proto3#json

https://developers.google.com/protocol-buffers/docs/reference/cpp/google.protobuf.compiler.plugin.pb

https://github.com/google/protobuf/wiki/Third-Party-Add-ons
