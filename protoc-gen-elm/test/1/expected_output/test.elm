import Json.Decode as JD exposing ((:=))


type Enum
  = EnumValueDefault -- 0
  | EnumValue1 -- 1
  | EnumValue2 -- 2
  | EnumValue123 -- 123


enum : JD.Decoder Enum
enum =
  let
    lookup s = case s of
      "ENUM_VALUE_DEFAULT" -> EnumValueDefault
      "ENUM_VALUE_1" -> EnumValue1
      "ENUM_VALUE_2" -> EnumValue2
      "ENUM_VALUE_123" -> EnumValue123
      _ -> EnumValueDefault
  in
    JD.map lookup JD.string


type alias SubMessage =
  { int32Field : Int
  }


subMessage : JD.Decoder SubMessage
subMessage =
  JD.object1 SubMessage
    ("int32Field" := JD.int)


type alias Foo =
  { int64Field : Int
  , boolField : Bool
  , stringField : String
  , enumField : Enum
  , subMessage : SubMessage
  }


foo : JD.Decoder Foo
foo =
  JD.object5 Foo
    ("int64Field" := JD.int)
    ("boolField" := JD.bool)
    ("stringField" := JD.string)
    ("enumField" := enum)
    ("subMessage" := subMessage)


