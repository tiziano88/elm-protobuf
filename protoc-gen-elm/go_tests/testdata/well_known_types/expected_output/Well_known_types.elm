module Well_known_types exposing (..)


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


optionalFieldDecoder : JD.Decoder a -> String -> JD.Decoder (Maybe a)
optionalFieldDecoder decoder name =
  optionalDecoder (name := decoder)


repeatedFieldDecoder : JD.Decoder a -> JD.Decoder (List a)
repeatedFieldDecoder decoder =
  withDefault [] (JD.list decoder)


withDefault : a -> JD.Decoder a -> JD.Decoder a
withDefault default decoder =
  JD.oneOf
    [ decoder
    , JD.succeed default
    ]


intFieldDecoder : String -> JD.Decoder Int
intFieldDecoder name =
  withDefault 0 (name := JD.int)


floatFieldDecoder : String -> JD.Decoder Float
floatFieldDecoder name =
  withDefault 0.0 (name := JD.float)


boolFieldDecoder : String -> JD.Decoder Bool
boolFieldDecoder name =
  withDefault False (name := JD.bool)


stringFieldDecoder : String -> JD.Decoder String
stringFieldDecoder name =
  withDefault "" (name := JD.string)


enumFieldDecoder : JD.Decoder a -> String -> JD.Decoder a
enumFieldDecoder decoder name =
  (name := decoder)


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


type alias Message =
  { doubleValueField : Maybe Google.Protobuf.DoubleValue -- 1
  }


messageDecoder : JD.Decoder Message
messageDecoder =
  Message
    <$> (optionalFieldDecoder google_Protobuf_DoubleValueDecoder "doubleValueField")


messageEncoder : Message -> JE.Value
messageEncoder v =
  JE.object
    [ ("doubleValueField", optionalEncoder google_Protobuf_DoubleValueEncoder v.doubleValueField)
    ]


