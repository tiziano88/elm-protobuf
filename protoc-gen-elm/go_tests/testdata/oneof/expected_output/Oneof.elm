module Oneof exposing (..)


import Json.Decode as JD exposing ((:=))
import Json.Encode as JE


(<$>) : (a -> b) -> JD.Decoder a -> JD.Decoder b
(<$>) =
  JD.map


(<*>) : JD.Decoder (a -> b) -> JD.Decoder a -> JD.Decoder b
(<*>) f v =
  f `JD.andThen` \x -> x <$> v


optionalDecoder : JD.Decoder a -> JD.Decoder (Maybe a)
optionalDecoder decoder =
  JD.oneOf
    [ JD.map Just decoder
    , JD.succeed Nothing
    ]


requiredFieldDecoder : String -> a -> JD.Decoder a -> JD.Decoder a
requiredFieldDecoder name default decoder =
  withDefault default (name := decoder)


optionalFieldDecoder : String -> JD.Decoder a -> JD.Decoder (Maybe a)
optionalFieldDecoder name decoder =
  optionalDecoder (name := decoder)


repeatedFieldDecoder : String -> JD.Decoder a -> JD.Decoder (List a)
repeatedFieldDecoder name decoder =
  withDefault [] (name := (JD.list decoder))


withDefault : a -> JD.Decoder a -> JD.Decoder a
withDefault default decoder =
  JD.oneOf
    [ decoder
    , JD.succeed default
    ]


optionalEncoder : (a -> JE.Value) -> Maybe a -> JE.Value
optionalEncoder encoder v =
  case v of
    Just x ->
      encoder x
    
    Nothing ->
      JE.null


repeatedFieldEncoder : (a -> JE.Value) -> List a -> JE.Value
repeatedFieldEncoder encoder v =
  JE.list <| List.map encoder v


type alias Foo =
  { firstOneof : FirstOneof
  , secondOneof : SecondOneof
  }


type FirstOneof
  = StringField String
  | IntField Int


firstOneofDecoder : JD.Decoder FirstOneof
firstOneofDecoder =
  JD.oneOf
    [ JD.map StringField ("stringField" := JD.string)
    , JD.map IntField ("intField" := JD.int)
    ]


firstOneofEncoder : FirstOneof -> (String, JE.Value)
firstOneofEncoder v =
  case v of
    StringField x -> ("stringField", JE.string x)
    IntField x -> ("intField", JE.int x)


type SecondOneof
  = BoolField Bool
  | OtherStringField String


secondOneofDecoder : JD.Decoder SecondOneof
secondOneofDecoder =
  JD.oneOf
    [ JD.map BoolField ("boolField" := JD.bool)
    , JD.map OtherStringField ("otherStringField" := JD.string)
    ]


secondOneofEncoder : SecondOneof -> (String, JE.Value)
secondOneofEncoder v =
  case v of
    BoolField x -> ("boolField", JE.bool x)
    OtherStringField x -> ("otherStringField", JE.string x)


fooDecoder : JD.Decoder Foo
fooDecoder =
  Foo
    <$> firstOneofDecoder
    <*> secondOneofDecoder


fooEncoder : Foo -> JE.Value
fooEncoder v =
  JE.object
    [ firstOneofEncoder v.firstOneof
    , secondOneofEncoder v.secondOneof
    ]
