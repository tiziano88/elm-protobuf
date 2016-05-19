module google.protobuf.Wrappers exposing (..)


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


type alias DoubleValue =
  { value : Float -- 1
  }


doubleValueDecoder : JD.Decoder DoubleValue
doubleValueDecoder =
  DoubleValue
    <$> (floatFieldDecoder "value")


doubleValueEncoder : DoubleValue -> JE.Value
doubleValueEncoder v =
  JE.object
    [ ("value", JE.float v.value)
    ]


type alias FloatValue =
  { value : Float -- 1
  }


floatValueDecoder : JD.Decoder FloatValue
floatValueDecoder =
  FloatValue
    <$> (floatFieldDecoder "value")


floatValueEncoder : FloatValue -> JE.Value
floatValueEncoder v =
  JE.object
    [ ("value", JE.float v.value)
    ]


type alias Int64Value =
  { value : Int -- 1
  }


int64ValueDecoder : JD.Decoder Int64Value
int64ValueDecoder =
  Int64Value
    <$> (intFieldDecoder "value")


int64ValueEncoder : Int64Value -> JE.Value
int64ValueEncoder v =
  JE.object
    [ ("value", JE.int v.value)
    ]


type alias UInt64Value =
  { value : Int -- 1
  }


uInt64ValueDecoder : JD.Decoder UInt64Value
uInt64ValueDecoder =
  UInt64Value
    <$> (intFieldDecoder "value")


uInt64ValueEncoder : UInt64Value -> JE.Value
uInt64ValueEncoder v =
  JE.object
    [ ("value", JE.int v.value)
    ]


type alias Int32Value =
  { value : Int -- 1
  }


int32ValueDecoder : JD.Decoder Int32Value
int32ValueDecoder =
  Int32Value
    <$> (intFieldDecoder "value")


int32ValueEncoder : Int32Value -> JE.Value
int32ValueEncoder v =
  JE.object
    [ ("value", JE.int v.value)
    ]


type alias UInt32Value =
  { value : Int -- 1
  }


uInt32ValueDecoder : JD.Decoder UInt32Value
uInt32ValueDecoder =
  UInt32Value
    <$> (intFieldDecoder "value")


uInt32ValueEncoder : UInt32Value -> JE.Value
uInt32ValueEncoder v =
  JE.object
    [ ("value", JE.int v.value)
    ]


type alias BoolValue =
  { value : Bool -- 1
  }


boolValueDecoder : JD.Decoder BoolValue
boolValueDecoder =
  BoolValue
    <$> (boolFieldDecoder "value")


boolValueEncoder : BoolValue -> JE.Value
boolValueEncoder v =
  JE.object
    [ ("value", JE.bool v.value)
    ]


type alias StringValue =
  { value : String -- 1
  }


stringValueDecoder : JD.Decoder StringValue
stringValueDecoder =
  StringValue
    <$> (stringFieldDecoder "value")


stringValueEncoder : StringValue -> JE.Value
stringValueEncoder v =
  JE.object
    [ ("value", JE.string v.value)
    ]


type alias BytesValue =
  { value : Bytes -- 1
  }


bytesValueDecoder : JD.Decoder BytesValue
bytesValueDecoder =
  BytesValue
    <$> (bytesFieldDecoder "value")


bytesValueEncoder : BytesValue -> JE.Value
bytesValueEncoder v =
  JE.object
    [ ("value", bytesFieldEncoder v.value)
    ]


