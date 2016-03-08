module Test where


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


repeatedFieldDecoder : JD.Decoder a -> String -> JD.Decoder (List a)
repeatedFieldDecoder decoder name =
  JD.list (name := decoder)


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


type Enum
  = EnumValueDefault -- 0
  | EnumValue1 -- 1
  | EnumValue2 -- 2
  | EnumValue123 -- 123


enumDecoder : JD.Decoder Enum
enumDecoder =
  let
    lookup s = case s of
      "ENUM_VALUE_DEFAULT" -> EnumValueDefault
      "ENUM_VALUE_1" -> EnumValue1
      "ENUM_VALUE_2" -> EnumValue2
      "ENUM_VALUE_123" -> EnumValue123
      _ -> EnumValueDefault
  in
    JD.map lookup JD.string


enumEncoder : Enum -> JE.Value
enumEncoder v =
  let
    lookup s = case s of
      EnumValueDefault -> "ENUM_VALUE_DEFAULT"
      EnumValue1 -> "ENUM_VALUE_1"
      EnumValue2 -> "ENUM_VALUE_2"
      EnumValue123 -> "ENUM_VALUE_123"
  in
    JE.string <| lookup v


type alias SubMessage =
  { int32Field : Int -- 1
  }


subMessageDecoder : JD.Decoder SubMessage
subMessageDecoder =
  SubMessage
    <$> (intFieldDecoder "int32Field")


subMessageEncoder : SubMessage -> JE.Value
subMessageEncoder v =
  JE.object
    [ ("int32Field", JE.int v.int32Field)
    ]


type alias Foo =
  { doubleField : Float -- 1
  , floatField : Float -- 2
  , int32Field : Int -- 3
  , int64Field : Int -- 4
  , uint32Field : Int -- 5
  , uint64Field : Int -- 6
  , sint32Field : Int -- 7
  , sint64Field : Int -- 8
  , fixed32Field : Int -- 9
  , fixed64Field : Int -- 10
  , sfixed32Field : Int -- 11
  , sfixed64Field : Int -- 12
  , boolField : Bool -- 13
  , stringField : String -- 14
  , enumField : Enum -- 15
  , subMessage : Maybe SubMessage -- 16
  , repeatedInt64Field : List Int -- 17
  , repeatedEnumField : List Enum -- 18
  }


fooDecoder : JD.Decoder Foo
fooDecoder =
  Foo
    <$> (floatFieldDecoder "doubleField")
    <*> (floatFieldDecoder "floatField")
    <*> (intFieldDecoder "int32Field")
    <*> (intFieldDecoder "int64Field")
    <*> (intFieldDecoder "uint32Field")
    <*> (intFieldDecoder "uint64Field")
    <*> (intFieldDecoder "sint32Field")
    <*> (intFieldDecoder "sint64Field")
    <*> (intFieldDecoder "fixed32Field")
    <*> (intFieldDecoder "fixed64Field")
    <*> (intFieldDecoder "sfixed32Field")
    <*> (intFieldDecoder "sfixed64Field")
    <*> (boolFieldDecoder "boolField")
    <*> (stringFieldDecoder "stringField")
    <*> ((enumFieldDecoder enumDecoder) "enumField")
    <*> (optionalFieldDecoder subMessageDecoder "subMessage")
    <*> (repeatedFieldDecoder intFieldDecoder "repeatedInt64Field")
    <*> (repeatedFieldDecoder (enumFieldDecoder enumDecoder) "repeatedEnumField")


fooEncoder : Foo -> JE.Value
fooEncoder v =
  JE.object
    [ ("doubleField", JE.float v.doubleField)
    , ("floatField", JE.float v.floatField)
    , ("int32Field", JE.int v.int32Field)
    , ("int64Field", JE.int v.int64Field)
    , ("uint32Field", JE.int v.uint32Field)
    , ("uint64Field", JE.int v.uint64Field)
    , ("sint32Field", JE.int v.sint32Field)
    , ("sint64Field", JE.int v.sint64Field)
    , ("fixed32Field", JE.int v.fixed32Field)
    , ("fixed64Field", JE.int v.fixed64Field)
    , ("sfixed32Field", JE.int v.sfixed32Field)
    , ("sfixed64Field", JE.int v.sfixed64Field)
    , ("boolField", JE.bool v.boolField)
    , ("stringField", JE.string v.stringField)
    , ("enumField", enumEncoder v.enumField)
    , ("subMessage", optionalEncoder subMessageEncoder v.subMessage)
    , ("repeatedInt64Field", repeatedFieldEncoder JE.int v.repeatedInt64Field)
    , ("repeatedEnumField", repeatedFieldEncoder enumEncoder v.repeatedEnumField)
    ]


