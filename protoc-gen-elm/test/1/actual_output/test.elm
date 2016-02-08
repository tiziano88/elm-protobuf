import Json.Decode

type alias Foo =
  { int64_field : Int
  , bool_field : Bool
  , string_field : String
  }

foo : Decoder Foo
foo =
  objectN Foo
    (int64_field := int)
    (bool_field := bool)
    (string_field := string)
