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


optionalEncoder : String -> (a -> JE.Value) -> Maybe a -> Maybe (String, JE.Value)
optionalEncoder name encoder v =
  case v of
    Just x ->
      Just (name, encoder x)
    
    Nothing ->
      Nothing


requiredFieldEncoder : String -> (a -> JE.Value) -> a -> a -> Maybe (String, JE.Value)
requiredFieldEncoder name encoder default v =
  if
    v == default
  then
    Nothing
  else
    Just (name, encoder v)


repeatedFieldEncoder : String -> (a -> JE.Value) -> List a -> Maybe (String, JE.Value)
repeatedFieldEncoder name encoder v =
  case v of
    [] ->
      Nothing
    _ ->
      Just (name, JE.list <| List.map encoder v)


type alias Foo =
  { firstOneof : FirstOneof
  , secondOneof : SecondOneof
  }


type FirstOneof
  = FirstOneofUnspecified
  | StringField String
  | IntField Int


firstOneofDecoder : JD.Decoder FirstOneof
firstOneofDecoder =
  JD.oneOf
    [ JD.map StringField ("stringField" := JD.string)
    , JD.map IntField ("intField" := JD.int)
    , JD.succeed FirstOneofUnspecified
    ]


firstOneofEncoder : FirstOneof -> Maybe (String, JE.Value)
firstOneofEncoder v =
  case v of
    FirstOneofUnspecified -> Nothing
    StringField x -> Just ("stringField", JE.string x)
    IntField x -> Just ("intField", JE.int x)


type SecondOneof
  = SecondOneofUnspecified
  | BoolField Bool
  | OtherStringField String


secondOneofDecoder : JD.Decoder SecondOneof
secondOneofDecoder =
  JD.oneOf
    [ JD.map BoolField ("boolField" := JD.bool)
    , JD.map OtherStringField ("otherStringField" := JD.string)
    , JD.succeed SecondOneofUnspecified
    ]


secondOneofEncoder : SecondOneof -> Maybe (String, JE.Value)
secondOneofEncoder v =
  case v of
    SecondOneofUnspecified -> Nothing
    BoolField x -> Just ("boolField", JE.bool x)
    OtherStringField x -> Just ("otherStringField", JE.string x)


fooDecoder : JD.Decoder Foo
fooDecoder =
  Foo
    <$> firstOneofDecoder
    <*> secondOneofDecoder


fooEncoder : Foo -> JE.Value
fooEncoder v =
  JE.object <| List.filterMap identity <|
    [ (firstOneofEncoder v.firstOneof)
    , (secondOneofEncoder v.secondOneof)
    ]
