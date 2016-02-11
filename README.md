# elm-protobuf

Experimental protobuf plugin generating elm code from proto definitions.

The plugin itself is written in Go.

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
https://developers.google.com/protocol-buffers/docs/reference/cpp/google.protobuf.compiler.plugin.pb
https://github.com/google/protobuf/wiki/Third-Party-Add-ons
