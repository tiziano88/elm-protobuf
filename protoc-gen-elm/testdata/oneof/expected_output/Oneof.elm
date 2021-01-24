module Oneof exposing (..)

-- DO NOT EDIT
-- AUTOGENERATED BY THE ELM PROTOCOL BUFFER COMPILER
-- https://github.com/tiziano88/elm-protobuf
-- source file: oneof.proto

import Protobuf exposing (..)

import Json.Decode as JD
import Json.Encode as JE


uselessDeclarationToPreventErrorDueToEmptyOutputFile = 42


type alias Foo =
    { firstOneof : FirstOneof
    , secondOneof : SecondOneof
    }


fooDecoder : JD.Decoder Foo
fooDecoder =
    JD.lazy <| \_ -> decode Foo
        |> field firstOneofDecoder
        |> field secondOneofDecoder


fooEncoder : Foo -> JE.Value
fooEncoder v =
    JE.object <| List.filterMap identity <|
        [ (firstOneofEncoder v.firstOneof)
        , (secondOneofEncoder v.secondOneof)
        ]


type FirstOneof
    = FirstOneofUnspecified
    | StringField String
    | IntField Int


firstOneofDecoder : JD.Decoder FirstOneof
firstOneofDecoder =
    JD.lazy <| \_ -> JD.oneOf
        [ JD.map StringField (JD.field "stringField" JD.string)
        , JD.map IntField (JD.field "intField" intDecoder)
        , JD.succeed FirstOneofUnspecified
        ]


firstOneofEncoder : FirstOneof -> Maybe ( String, JE.Value )
firstOneofEncoder v =
    case v of
        FirstOneofUnspecified ->
            Nothing
        StringField x ->
            Just ( "stringField", JE.string x )
        IntField x ->
            Just ( "intField", JE.int x )


type SecondOneof
    = SecondOneofUnspecified
    | BoolField Bool
    | OtherStringField String


secondOneofDecoder : JD.Decoder SecondOneof
secondOneofDecoder =
    JD.lazy <| \_ -> JD.oneOf
        [ JD.map BoolField (JD.field "boolField" JD.bool)
        , JD.map OtherStringField (JD.field "otherStringField" JD.string)
        , JD.succeed SecondOneofUnspecified
        ]


secondOneofEncoder : SecondOneof -> Maybe ( String, JE.Value )
secondOneofEncoder v =
    case v of
        SecondOneofUnspecified ->
            Nothing
        BoolField x ->
            Just ( "boolField", JE.bool x )
        OtherStringField x ->
            Just ( "otherStringField", JE.string x )
