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


type alias Foo =
  { int64_field : Int
  , bool_field : Bool
  , string_field : String
  , enum_field : Enum
  }


foo : JD.Decoder Foo
foo =
  JD.object4 Foo
    ("int64_field" := JD.int)
    ("bool_field" := JD.bool)
    ("string_field" := JD.string)
    ("enum_field" := enum)


