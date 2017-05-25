# elm-protobuf

[![Build Status](https://travis-ci.org/tiziano88/elm-protobuf.svg?branch=master)](https://travis-ci.org/tiziano88/elm-protobuf)

[Join #elm-protobuf on Slack](https://elmlang.slack.com/messages/elm-protobuf/details/)

Experimental protobuf plugin generating elm code from proto definitions.

The plugin itself is written in Go, and it requires the base `protoc` protobuf
compiler to be installed on the system.

For a sample generated output file, see
https://github.com/tiziano88/elm-protobuf/blob/master/protoc-gen-elm/go_tests/testdata/repeated/expected_output/Repeated.elm

## Supported features

-   [x] `double`/`float` fields
-   [x]
    `int32`/`int64`/`uint32`/`uint64`/`sint32`/`sint64`/`fixed32`/`fixed64`/`sfixed32`/`sfixed64`
    fields
-   [x] `bool` fields
-   [x] `string` fields
-   [ ] `bytes` fields
-   [x] message fields
-   [x] enum fields
-   [x] imports
-   [x] nested types
-   [ ] `Any` type
-   [x] `Timestamp` type
-   [ ] `Duration` type
-   [ ] `Struct` type
-   [x] wrapper types
-   [ ] `FieldMask` type
-   [ ] `ListValue` type
-   [ ] `Value` type
-   [ ] `NullValue` type
-   [ ] `oneof`
-   [ ] `map`
-   [ ] packages
-   [ ] options

## How to install

### Release

The simplest way to install the plugin is to download a pre-compiled binary from
https://github.com/tiziano88/elm-protobuf/releases , then unpack it and copy or
move the `protoc-gen-elm` binary somewhere in your `$PATH`.

### From source

-   Make sure that you have a Go environment correctly set up, and that
    `$GOPATH/bin` is included in your `$PATH`. See
    https://golang.org/doc/install for info.

-   Install a recent `protoc` compiler version from
    https://github.com/google/protobuf (it must have support for `proto3`
    format).

-   Obtain the `protoc-gen-elm` binary using `go get`:

    ```
    go get github.com/tiziano88/elm-protobuf/protoc-gen-elm
    ```

## How to run

Run the `protoc` compiler specifying the `--elm_out` flag:

`protoc --elm_out=. *.proto`

`protoc` will automatically detect the `protoc-gen-elm` binary from your `$PATH`
and use it to generate the output elm code.

Then, in your project, add a dependency on the runtime library:

`elm package install tiziano88/elm-protobuf`

## References

https://developers.google.com/protocol-buffers/

https://developers.google.com/protocol-buffers/docs/proto3#json

https://developers.google.com/protocol-buffers/docs/reference/cpp/google.protobuf.compiler.plugin.pb

https://github.com/google/protobuf/wiki/Third-Party-Add-ons
