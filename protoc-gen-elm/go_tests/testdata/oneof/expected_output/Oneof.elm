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
  { stringField : String -- 1
  , intField : Int -- 2
  , boolField : Bool -- 3
  , otherStringField : String -- 4
  }


fooDecoder : JD.Decoder Foo
fooDecoder =
  Foo
    <$> (requiredFieldDecoder "stringField" "" JD.string)
    <*> (requiredFieldDecoder "intField" 0 JD.int)
    <*> (requiredFieldDecoder "boolField" False JD.bool)
    <*> (requiredFieldDecoder "otherStringField" "" JD.string)


fooEncoder : Foo -> JE.Value
fooEncoder v =
  JE.object
    [ ("stringField", JE.string v.stringField)
    , ("intField", JE.int v.intField)
    , ("boolField", JE.bool v.boolField)
    , ("otherStringField", JE.string v.otherStringField)
    ]


