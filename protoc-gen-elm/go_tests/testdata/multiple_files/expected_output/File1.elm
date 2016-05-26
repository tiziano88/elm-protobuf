module File1 exposing (..)


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


type alias File1Message =
  { field : Bool -- 1
  }


file1MessageDecoder : JD.Decoder File1Message
file1MessageDecoder =
  File1Message
    <$> (requiredFieldDecoder "field" False JD.bool)


file1MessageEncoder : File1Message -> JE.Value
file1MessageEncoder v =
  JE.object <| List.filterMap identity <|
    [ Just ("field", JE.bool v.field)
    ]
