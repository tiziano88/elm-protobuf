module Repeated exposing (..)


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


enumDefault : Enum
enumDefault = EnumValueDefault


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
    <$> (requiredFieldDecoder "int32Field" 0 JD.int)


subMessageEncoder : SubMessage -> JE.Value
subMessageEncoder v =
  JE.object <| List.filterMap identity <|
    [ (requiredFieldEncoder "int32Field" JE.int 0 v.int32Field)
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
  , nestedMessageField : Maybe Foo_NestedMessage -- 19
  , nestedEnumField : Foo_NestedEnum -- 20
  }


type Foo_NestedEnum
  = Foo_EnumValueDefault -- 0


fooDecoder : JD.Decoder Foo
fooDecoder =
  Foo
    <$> (requiredFieldDecoder "doubleField" 0.0 JD.float)
    <*> (requiredFieldDecoder "floatField" 0.0 JD.float)
    <*> (requiredFieldDecoder "int32Field" 0 JD.int)
    <*> (requiredFieldDecoder "int64Field" 0 JD.int)
    <*> (requiredFieldDecoder "uint32Field" 0 JD.int)
    <*> (requiredFieldDecoder "uint64Field" 0 JD.int)
    <*> (requiredFieldDecoder "sint32Field" 0 JD.int)
    <*> (requiredFieldDecoder "sint64Field" 0 JD.int)
    <*> (requiredFieldDecoder "fixed32Field" 0 JD.int)
    <*> (requiredFieldDecoder "fixed64Field" 0 JD.int)
    <*> (requiredFieldDecoder "sfixed32Field" 0 JD.int)
    <*> (requiredFieldDecoder "sfixed64Field" 0 JD.int)
    <*> (requiredFieldDecoder "boolField" False JD.bool)
    <*> (requiredFieldDecoder "stringField" "" JD.string)
    <*> (requiredFieldDecoder "enumField" enumDefault enumDecoder)
    <*> (optionalFieldDecoder "subMessage" subMessageDecoder)
    <*> (repeatedFieldDecoder "repeatedInt64Field" JD.int)
    <*> (repeatedFieldDecoder "repeatedEnumField" enumDecoder)
    <*> (optionalFieldDecoder "nestedMessageField" foo_NestedMessageDecoder)
    <*> (requiredFieldDecoder "nestedEnumField" foo_NestedEnumDefault foo_NestedEnumDecoder)


foo_NestedEnumDecoder : JD.Decoder Foo_NestedEnum
foo_NestedEnumDecoder =
  let
    lookup s = case s of
      "ENUM_VALUE_DEFAULT" -> Foo_EnumValueDefault
      _ -> Foo_EnumValueDefault
  in
    JD.map lookup JD.string


foo_NestedEnumDefault : Foo_NestedEnum
foo_NestedEnumDefault = Foo_EnumValueDefault


fooEncoder : Foo -> JE.Value
fooEncoder v =
  JE.object <| List.filterMap identity <|
    [ (requiredFieldEncoder "doubleField" JE.float 0.0 v.doubleField)
    , (requiredFieldEncoder "floatField" JE.float 0.0 v.floatField)
    , (requiredFieldEncoder "int32Field" JE.int 0 v.int32Field)
    , (requiredFieldEncoder "int64Field" JE.int 0 v.int64Field)
    , (requiredFieldEncoder "uint32Field" JE.int 0 v.uint32Field)
    , (requiredFieldEncoder "uint64Field" JE.int 0 v.uint64Field)
    , (requiredFieldEncoder "sint32Field" JE.int 0 v.sint32Field)
    , (requiredFieldEncoder "sint64Field" JE.int 0 v.sint64Field)
    , (requiredFieldEncoder "fixed32Field" JE.int 0 v.fixed32Field)
    , (requiredFieldEncoder "fixed64Field" JE.int 0 v.fixed64Field)
    , (requiredFieldEncoder "sfixed32Field" JE.int 0 v.sfixed32Field)
    , (requiredFieldEncoder "sfixed64Field" JE.int 0 v.sfixed64Field)
    , (requiredFieldEncoder "boolField" JE.bool False v.boolField)
    , (requiredFieldEncoder "stringField" JE.string "" v.stringField)
    , (requiredFieldEncoder "enumField" enumEncoder enumDefault v.enumField)
    , (optionalEncoder "subMessage" subMessageEncoder v.subMessage)
    , (repeatedFieldEncoder "repeatedInt64Field" JE.int v.repeatedInt64Field)
    , (repeatedFieldEncoder "repeatedEnumField" enumEncoder v.repeatedEnumField)
    , (optionalEncoder "nestedMessageField" foo_NestedMessageEncoder v.nestedMessageField)
    , (requiredFieldEncoder "nestedEnumField" foo_NestedEnumEncoder foo_NestedEnumDefault v.nestedEnumField)
    ]


foo_NestedEnumEncoder : Foo_NestedEnum -> JE.Value
foo_NestedEnumEncoder v =
  let
    lookup s = case s of
      Foo_EnumValueDefault -> "ENUM_VALUE_DEFAULT"
  in
    JE.string <| lookup v


type alias Foo_NestedMessage =
  { int32Field : Int -- 1
  }


foo_NestedMessageDecoder : JD.Decoder Foo_NestedMessage
foo_NestedMessageDecoder =
  Foo_NestedMessage
    <$> (requiredFieldDecoder "int32Field" 0 JD.int)


foo_NestedMessageEncoder : Foo_NestedMessage -> JE.Value
foo_NestedMessageEncoder v =
  JE.object <| List.filterMap identity <|
    [ (requiredFieldEncoder "int32Field" JE.int 0 v.int32Field)
    ]


type alias Foo_NestedMessage_NestedNestedMessage =
  { int32Field : Int -- 1
  }


foo_NestedMessage_NestedNestedMessageDecoder : JD.Decoder Foo_NestedMessage_NestedNestedMessage
foo_NestedMessage_NestedNestedMessageDecoder =
  Foo_NestedMessage_NestedNestedMessage
    <$> (requiredFieldDecoder "int32Field" 0 JD.int)


foo_NestedMessage_NestedNestedMessageEncoder : Foo_NestedMessage_NestedNestedMessage -> JE.Value
foo_NestedMessage_NestedNestedMessageEncoder v =
  JE.object <| List.filterMap identity <|
    [ (requiredFieldEncoder "int32Field" JE.int 0 v.int32Field)
    ]


type alias FooRepeated =
  { doubleField : List Float -- 1
  , floatField : List Float -- 2
  , int32Field : List Int -- 3
  , int64Field : List Int -- 4
  , uint32Field : List Int -- 5
  , uint64Field : List Int -- 6
  , sint32Field : List Int -- 7
  , sint64Field : List Int -- 8
  , fixed32Field : List Int -- 9
  , fixed64Field : List Int -- 10
  , sfixed32Field : List Int -- 11
  , sfixed64Field : List Int -- 12
  , boolField : List Bool -- 13
  , stringField : List String -- 14
  , enumField : List Enum -- 15
  , subMessage : List SubMessage -- 16
  }


fooRepeatedDecoder : JD.Decoder FooRepeated
fooRepeatedDecoder =
  FooRepeated
    <$> (repeatedFieldDecoder "doubleField" JD.float)
    <*> (repeatedFieldDecoder "floatField" JD.float)
    <*> (repeatedFieldDecoder "int32Field" JD.int)
    <*> (repeatedFieldDecoder "int64Field" JD.int)
    <*> (repeatedFieldDecoder "uint32Field" JD.int)
    <*> (repeatedFieldDecoder "uint64Field" JD.int)
    <*> (repeatedFieldDecoder "sint32Field" JD.int)
    <*> (repeatedFieldDecoder "sint64Field" JD.int)
    <*> (repeatedFieldDecoder "fixed32Field" JD.int)
    <*> (repeatedFieldDecoder "fixed64Field" JD.int)
    <*> (repeatedFieldDecoder "sfixed32Field" JD.int)
    <*> (repeatedFieldDecoder "sfixed64Field" JD.int)
    <*> (repeatedFieldDecoder "boolField" JD.bool)
    <*> (repeatedFieldDecoder "stringField" JD.string)
    <*> (repeatedFieldDecoder "enumField" enumDecoder)
    <*> (repeatedFieldDecoder "subMessage" subMessageDecoder)


fooRepeatedEncoder : FooRepeated -> JE.Value
fooRepeatedEncoder v =
  JE.object <| List.filterMap identity <|
    [ (repeatedFieldEncoder "doubleField" JE.float v.doubleField)
    , (repeatedFieldEncoder "floatField" JE.float v.floatField)
    , (repeatedFieldEncoder "int32Field" JE.int v.int32Field)
    , (repeatedFieldEncoder "int64Field" JE.int v.int64Field)
    , (repeatedFieldEncoder "uint32Field" JE.int v.uint32Field)
    , (repeatedFieldEncoder "uint64Field" JE.int v.uint64Field)
    , (repeatedFieldEncoder "sint32Field" JE.int v.sint32Field)
    , (repeatedFieldEncoder "sint64Field" JE.int v.sint64Field)
    , (repeatedFieldEncoder "fixed32Field" JE.int v.fixed32Field)
    , (repeatedFieldEncoder "fixed64Field" JE.int v.fixed64Field)
    , (repeatedFieldEncoder "sfixed32Field" JE.int v.sfixed32Field)
    , (repeatedFieldEncoder "sfixed64Field" JE.int v.sfixed64Field)
    , (repeatedFieldEncoder "boolField" JE.bool v.boolField)
    , (repeatedFieldEncoder "stringField" JE.string v.stringField)
    , (repeatedFieldEncoder "enumField" enumEncoder v.enumField)
    , (repeatedFieldEncoder "subMessage" subMessageEncoder v.subMessage)
    ]
