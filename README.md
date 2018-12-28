# elm-protobuf

[![Build Status](https://travis-ci.org/tiziano88/elm-protobuf.svg?branch=master)](https://travis-ci.org/tiziano88/elm-protobuf)
[![Go Report Card](https://goreportcard.com/badge/github.com/tiziano88/elm-protobuf)](https://goreportcard.com/report/github.com/tiziano88/elm-protobuf)

[Join #elm-protobuf on Slack](https://elmlang.slack.com/messages/elm-protobuf/details/)

Experimental protobuf plugin generating elm code to decode/encode [Proto3/JSON](https://developers.google.com/protocol-buffers/docs/proto3#json) from proto definitions.

The plugin itself is written in Go, and it requires the base `protoc` protobuf
compiler to be installed on the system.

For a sample generated output file, see
[Repeated.elm](https://github.com/tiziano88/elm-protobuf/blob/master/protoc-gen-elm/go_tests/testdata/repeated/expected_output/Repeated.elm).

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
-   [x] `oneof`
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

`elm install tiziano88/elm-protobuf`


## Options

Options can be passed to the plugin in the --elm_out value, with the following
syntax:

`protoc "--elm_out=option=value;option=value1,value2:."`

The valid options are:

- `excludeFile`: A list of files that should be ignored. Usefull to ignore a
  proto2 file that is a dependency of the compiled file.

- `options`: A file to customize the output, see below

## Customize the output

The elm type used for each field is automatically determined, but can be
overridden by passing an option file to the plugin.

For example, if a message has a Uuid field, the proto definition will probably
be something like this:

```proto
message MyObject {
    string id = 1;
    string name = 2;
}
```

By adding an option file, it is possible to get a 'Maybe Uuid' field instead
of a string.

Write a yaml file with all the options, 'elm-proto-options.yaml':

```yaml
types:
  Uuid.Uuid:
    json:
      encoder: Uuid.encode
      decoder: Uuid.decoder

files:
  my_proto_file.proto:
    imports:
    - Uuid
    fields:
      MyObject.id:
        type: Uuid.Uuid
        required: false
```

Pass the option file as a parameter to the plugin:

```
`protoc --elm_out='options=./elm-proto-options.yaml:.' my_proto_file.proto`
```
## References

https://developers.google.com/protocol-buffers/

https://developers.google.com/protocol-buffers/docs/proto3#json

https://developers.google.com/protocol-buffers/docs/reference/cpp/google.protobuf.compiler.plugin.pb

https://github.com/google/protobuf/wiki/Third-Party-Add-ons
