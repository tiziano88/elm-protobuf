module File2 exposing (..)


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


type alias File2Message =
  { field : Bool -- 1
  }


file2MessageDecoder : JD.Decoder File2Message
file2MessageDecoder =
  File2Message
    <$> (requiredFieldDecoder "field" False JD.bool)


file2MessageEncoder : File2Message -> JE.Value
file2MessageEncoder v =
  JE.object <| List.filterMap identity <|
    [ (requiredFieldEncoder "field" JE.bool False v.field)
    ]
