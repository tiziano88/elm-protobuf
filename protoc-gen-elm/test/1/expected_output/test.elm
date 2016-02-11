import Json.Decode as JD exposing ((:=))
import Json.Encode as JE


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
    ("int32Field" := JD.int)


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
  , subMessage : SubMessage
  }


fooDecoder : JD.Decoder Foo
fooDecoder =
  JD.object5 Foo
    ("int64Field" := JD.int)
    ("boolField" := JD.bool)
    ("stringField" := JD.string)
    ("enumField" := enumDecoder)
    ("subMessage" := subMessageDecoder)


fooEncoder : Foo -> JE.Value
fooEncoder v =
  JE.object
    [ ("int64Field", JE.int v.int64Field)
    , ("boolField", JE.bool v.boolField)
    , ("stringField", JE.string v.stringField)
    , ("enumField", enumEncoder v.enumField)
    , ("subMessage", subMessageEncoder v.subMessage)
    ]


