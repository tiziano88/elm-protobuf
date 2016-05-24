module Simple exposing (..)


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


type Colour
  = ColourUnspecified -- 0
  | Red -- 1
  | Green -- 2
  | Blue -- 3


colourDecoder : JD.Decoder Colour
colourDecoder =
  let
    lookup s = case s of
      "COLOUR_UNSPECIFIED" -> ColourUnspecified
      "RED" -> Red
      "GREEN" -> Green
      "BLUE" -> Blue
      _ -> ColourUnspecified
  in
    JD.map lookup JD.string


colourDefault : Colour
colourDefault = ColourUnspecified


colourEncoder : Colour -> JE.Value
colourEncoder v =
  let
    lookup s = case s of
      ColourUnspecified -> "COLOUR_UNSPECIFIED"
      Red -> "RED"
      Green -> "GREEN"
      Blue -> "BLUE"
  in
    JE.string <| lookup v


type alias Simple =
  { int32Field : Int -- 1
  
  }



simpleDecoder : JD.Decoder Simple
simpleDecoder =
  Simple
    <$> (requiredFieldDecoder "int32Field" 0 JD.int)


simpleEncoder : Simple -> JE.Value
simpleEncoder v =
  JE.object
    [ ("int32Field", JE.int v.int32Field)
    ]


type alias Foo =
  { s : Maybe Simple -- 1
  , ss : List Simple -- 2
  , colour : Colour -- 3
  , colours : List Colour -- 4
  , singleIntField : Int -- 5
  , repeatedIntField : List Int -- 6
  , oo1 : Int -- 7
  , oo2 : Bool -- 8
  
  , oo : Oo
  }

type Oo
  = Oo1
  | Oo2


fooDecoder : JD.Decoder Foo
fooDecoder =
  Foo
    <$> (optionalFieldDecoder "s" simpleDecoder)
    <*> (repeatedFieldDecoder "ss" simpleDecoder)
    <*> (requiredFieldDecoder "colour" colourDefault colourDecoder)
    <*> (repeatedFieldDecoder "colours" colourDecoder)
    <*> (requiredFieldDecoder "singleIntField" 0 JD.int)
    <*> (repeatedFieldDecoder "repeatedIntField" JD.int)
    <*> (requiredFieldDecoder "oo1" 0 JD.int)
    <*> (requiredFieldDecoder "oo2" False JD.bool)


fooEncoder : Foo -> JE.Value
fooEncoder v =
  JE.object
    [ ("s", optionalEncoder simpleEncoder v.s)
    , ("ss", repeatedFieldEncoder simpleEncoder v.ss)
    , ("colour", colourEncoder v.colour)
    , ("colours", repeatedFieldEncoder colourEncoder v.colours)
    , ("singleIntField", JE.int v.singleIntField)
    , ("repeatedIntField", repeatedFieldEncoder JE.int v.repeatedIntField)
    , ("oo1", JE.int v.oo1)
    , ("oo2", JE.bool v.oo2)
    ]


