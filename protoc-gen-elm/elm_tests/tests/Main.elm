module Main exposing (..)

import Json.Decode as JD
import Json.Encode as JE
import Date
import Fuzzer as F
import Result
import String
import Task
import Test exposing (..)
import Fuzz exposing (..)
import Expect exposing (..)
import Simple as T
import Wrappers as W
import Keywords as K
import Recursive as R
import Protobuf exposing (..)
import ISO8601


suite : Test
suite =
    describe "JSON"
        [ test "JSON encode" <| \() -> encode T.simpleEncoder msg |> equal msgJson
        , test "JSON decode" <| \() -> decode T.simpleDecoder msgJson |> equal (Ok msg)
        , test "JSON decode extra field" <| \() -> decode T.simpleDecoder msgExtraFieldJson |> equal (Ok msg)
        , test "JSON encode empty message" <| \() -> encode T.fooEncoder fooDefault |> equal emptyJson
        , test "JSON decode empty JSON" <| \() -> decode T.simpleDecoder emptyJson |> equal (Ok msgDefault)
        , test "JSON encode message with repeated field" <| \_ -> encode T.fooEncoder foo |> equal fooJson
          -- TODO: Should fail.
        , test "JSON decode wrong type" <| \() -> decode T.simpleDecoder wrongTypeJson |> equal (Ok msgDefault)
        , test "JSON decode null" <| \() -> decode T.simpleDecoder nullJson |> equal (Ok msgDefault)
        , describe "oneof"
            [ test "encode" <| \() -> encode T.fooEncoder foo |> equal fooJson
            , describe "decode"
                [ test "empty" <| \() -> decode T.fooDecoder emptyJson |> equal (Ok fooDefault)
                , test "oo1" <| \() -> decode T.fooDecoder oo1SetJson |> equal (Ok oo1Set)
                , test "oo2" <| \() -> decode T.fooDecoder oo2SetJson |> equal (Ok oo2Set)
                ]
            ]
        , describe "recursion"
            [ test "decode empty JSON" <| \() -> decode R.recDecoder emptyJson |> equal (Ok recDefault)
            , describe "decode"
                [ test "1-level JSON" <| \() -> decode R.recDecoder recJson1 |> equal (Ok rec1)
                , test "2-level JSON" <| \() -> decode R.recDecoder recJson2 |> equal (Ok rec2)
                ]
            ]
        , describe "timestamp"
            [ test "encode" <| \() -> encode T.fooEncoder timestampFoo |> equal timestampJson
            , test "decode" <| \() -> decode T.fooDecoder timestampJson |> equal (Ok timestampFoo)
            ]
        , describe "wrappers"
            -- TODO: Preserve nulls.
            [ test "encodeEmpty" <| \() -> encode W.wrappersEncoder wrappersEmpty |> equal wrappersJsonEmpty
            , describe "decode"
                [ test "Empty" <| \() -> decode W.wrappersDecoder wrappersJsonEmpty |> equal (Ok wrappersEmpty)
                , test "Zero" <| \() -> decode W.wrappersDecoder wrappersJsonZero |> equal (Ok wrappersZero)
                , test "Set" <| \() -> decode W.wrappersDecoder wrappersJsonSet |> equal (Ok wrappersSet)
                ]
            ]
        , describe "encode / decode"
            [ fuzz (tuple5 ( string, int, maybe string, maybe int, maybe float )) "fuzzer" <|
                assertEncodeDecode F.fuzzEncoder F.fuzzDecoder
                    << genFuzz
            ]
        ]


fuzz : Fuzzer a -> String -> (a -> Expectation) -> Test
fuzz =
    fuzzWith { runs = 2000 }


encode : (a -> JE.Value) -> a -> String
encode encoder msg =
    JE.encode 2 (encoder msg)


decode : JD.Decoder a -> String -> Result String a
decode decoder json =
    JD.decodeString decoder json


assertEncodeDecode : (a -> JE.Value) -> JD.Decoder a -> a -> Expectation
assertEncodeDecode encoder decoder msg =
    let
        encoded =
            encode encoder msg

        decoded =
            decode decoder encoded
    in
        decoded |> equal (Ok msg)


genFuzz : ( String, Int, Maybe String, Maybe Int, Maybe Float ) -> F.Fuzz
genFuzz ( s1, i1, s2, i2, t ) =
    { stringField = s1
    , int32Field = i1
    , stringValueField = s2
    , int32ValueField = i2
    , timestampField = Maybe.map Date.fromTime t
    }


msg : T.Simple
msg =
    { int32Field = 123
    }


msgDefault : T.Simple
msgDefault =
    { int32Field = 0
    }


fooDefault : T.Foo
fooDefault =
    { s = Nothing
    , ss = []
    , colour = T.ColourUnspecified
    , colours = []
    , singleIntField = 0
    , repeatedIntField = []
    , oo = T.OoUnspecified
    , bytesField = []
    , stringValueField = Nothing
    , otherField = Nothing
    , otherDirField = Nothing
    , timestampField = Nothing
    }


recDefault : R.Rec
recDefault =
    { int32Field = 0
    , r = R.RUnspecified
    , stringField = ""
    }


msgJson : String
msgJson =
    String.trim """
{
  "int32Field": 123
}
"""


msgExtraFieldJson : String
msgExtraFieldJson =
    String.trim """
{
  "int32Field": 123,
  "extraField": "abc"
}
"""


emptyJson =
    String.trim """
{}
"""


foo : T.Foo
foo =
    { s =
        Just
            { int32Field = 11
            }
    , ss =
        [ { int32Field = 111
          }
        , { int32Field = 222
          }
        ]
    , colour = T.Red
    , colours =
        [ T.Red
        , T.Red
        ]
    , singleIntField = 123
    , repeatedIntField =
        [ 111
        , 222
        , 333
        ]
    , oo = T.Oo1 1
    , bytesField = []
    , stringValueField = Nothing
    , otherField =
        Just
            { stringField = "xxx"
            }
    , otherDirField =
        Just
            { stringField = "yyy"
            }
    , timestampField = Nothing
    }


fooJson : String
fooJson =
    String.trim """
{
  "s": {
    "int32Field": 11
  },
  "ss": [
    {
      "int32Field": 111
    },
    {
      "int32Field": 222
    }
  ],
  "colour": "RED",
  "colours": [
    "RED",
    "RED"
  ],
  "singleIntField": 123,
  "repeatedIntField": [
    111,
    222,
    333
  ],
  "otherField": {
    "stringField": "xxx"
  },
  "otherDirField": {
    "stringField": "yyy"
  },
  "oo1": 1
}
"""


nullJson : String
nullJson =
    String.trim """
{
  "singleIntField": null
}
"""


wrongTypeJson : String
wrongTypeJson =
    String.trim """
{
  "singleIntField": "invalid-value"
}
"""


oo1Set : T.Foo
oo1Set =
    { fooDefault
        | oo = T.Oo1 123
    }


oo1SetJson : String
oo1SetJson =
    String.trim """
{
  "oo1": 123
}
"""


oo2Set : T.Foo
oo2Set =
    { fooDefault
        | oo = T.Oo2 True
    }


oo2SetJson : String
oo2SetJson =
    String.trim """
{
  "oo2": true
}
"""


recJson1 : String
recJson1 =
    String.trim """
{
  "recField": {}
}
"""


rec1 : R.Rec
rec1 =
    { int32Field = 0
    , r =
        R.RecField
            { int32Field = 0
            , r = R.RUnspecified
            , stringField = ""
            }
    , stringField = ""
    }


recJson2 : String
recJson2 =
    String.trim """
{
  "recField": {
    "recField": {}
  }
}
"""


rec2 : R.Rec
rec2 =
    { int32Field = 0
    , r =
        R.RecField
            { int32Field = 0
            , r =
                R.RecField
                    { int32Field = 0
                    , r = R.RUnspecified
                    , stringField = ""
                    }
            , stringField = ""
            }
    , stringField = ""
    }


timestampJson : String
timestampJson =
    String.trim """
{
  "timestampField": "1988-12-14T01:23:45.678Z"
}
"""


timestampFoo : T.Foo
timestampFoo =
    { fooDefault
        | timestampField =
            Result.toMaybe <| Date.fromString "1988-12-14T01:23:45.678Z"
    }


wrappersJsonEmpty : String
wrappersJsonEmpty =
    String.trim """
{}
"""


wrappersJsonNull : String
wrappersJsonNull =
    String.trim """
{
  "int32ValueField": null,
  "int64ValueField": null,
  "uInt32ValueField": null,
  "uInt64ValueField": null,
  "doubleValueField": null,
  "floatValueField": null,
  "boolValueField": null,
  "stringValueField": null,
  "bytesValueField": null
}
"""


wrappersEmpty : W.Wrappers
wrappersEmpty =
    { int32ValueField = Nothing
    , int64ValueField = Nothing
    , uInt32ValueField = Nothing
    , uInt64ValueField = Nothing
    , doubleValueField = Nothing
    , floatValueField = Nothing
    , boolValueField = Nothing
    , stringValueField = Nothing
    , bytesValueField = Nothing
    }


wrappersJsonZero : String
wrappersJsonZero =
    String.trim """
{
  "int32ValueField": 0,
  "int64ValueField": 0,
  "uInt32ValueField": 0,
  "uInt64ValueField": 0,
  "doubleValueField": 0.0,
  "floatValueField": 0.0,
  "boolValueField": false,
  "stringValueField" : "",
  "bytesValueField" : ""
}
"""


wrappersZero : W.Wrappers
wrappersZero =
    { int32ValueField = Just 0
    , int64ValueField = Just 0
    , uInt32ValueField = Just 0
    , uInt64ValueField = Just 0
    , doubleValueField = Just 0.0
    , floatValueField = Just 0.0
    , boolValueField = Just False
    , stringValueField = Just ""
    , bytesValueField = Just []
    }


wrappersJsonSet : String
wrappersJsonSet =
    String.trim """
{
  "int32ValueField": 111,
  "int64ValueField": 222,
  "uInt32ValueField": 333,
  "uInt64ValueField": 444,
  "doubleValueField": 5.5,
  "floatValueField": 6.6,
  "boolValueField": true,
  "stringValueField" : "888",
  "bytesValueField" : ""
}
"""


wrappersSet : W.Wrappers
wrappersSet =
    { int32ValueField = Just 111
    , int64ValueField = Just 222
    , uInt32ValueField = Just 333
    , uInt64ValueField = Just 444
    , doubleValueField = Just 5.5
    , floatValueField = Just 6.6
    , boolValueField = Just True
    , stringValueField = Just "888"
    , bytesValueField = Just []
    }
