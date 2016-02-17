module Test where


import Json.Decode as JD exposing ((:=))
import Json.Encode as JE


optional : JD.Decoder a -> JD.Decoder (Maybe a)
optional decoder =
  JD.oneOf
    [ JD.map Just decoder
    , JD.succeed Nothing
    ]


withDefault : a -> JD.Decoder a -> JD.Decoder a
withDefault default decoder =
  JD.oneOf
    [ decoder
    , JD.succeed default
    ]


intField : String -> JD.Decoder Int
intField name =
  withDefault 0 (name := JD.int)


boolField : String -> JD.Decoder Bool
boolField name =
  withDefault False (name := JD.bool)


stringField : String -> JD.Decoder String
stringField name =
  withDefault "" (name := JD.string)


messageField : JD.Decoder a -> String -> JD.Decoder (Maybe a)
messageField decoder name =
  optional (name := decoder)


enumField : JD.Decoder a -> String -> JD.Decoder a
enumField decoder name =
  (name := decoder)


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
  { int32Field : Int
  }


subMessageDecoder : JD.Decoder SubMessage
subMessageDecoder =
  JD.object1 SubMessage
    (intField "int32Field")


subMessageEncoder : SubMessage -> JE.Value
subMessageEncoder v =
  JE.object
    [ ("int32Field", JE.int v.int32Field)
    ]


type alias Foo =
  { int64Field : Int
  , boolField : Bool
  , stringField : String
  , enumField : Enum
  , subMessage : Maybe SubMessage
  }


fooDecoder : JD.Decoder Foo
fooDecoder =
  JD.object5 Foo
    (intField "int64Field")
    (boolField "boolField")
    (stringField "stringField")
    (enumField enumDecoder "enumField")
    (messageField subMessageDecoder "subMessage")


fooEncoder : Foo -> JE.Value
fooEncoder v =
  JE.object
    [ ("int64Field", JE.int v.int64Field)
    , ("boolField", JE.bool v.boolField)
    , ("stringField", JE.string v.stringField)
    , ("enumField", enumEncoder v.enumField)
    ]


