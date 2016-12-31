port module Main exposing (..)

import Json.Decode as JD
import Json.Encode as JE
import Result
import String
import Task
import Test exposing (..)
import Test.Runner.Node exposing (run, TestProgram)
import Expect exposing (..)
import Simple as T


main : TestProgram
main =
    run emit suite


port emit : ( String, JE.Value ) -> Cmd msg


suite : Test
suite =
    describe "JSON"
        [ test "JSON encode" <| \_ -> equal msgJson (JE.encode 2 (T.simpleEncoder msg))
        , test "JSON decode" <| \_ -> assertDecode T.simpleDecoder msgJson msg
        , test "JSON decode extra field" <| \_ -> assertDecode T.simpleDecoder msgExtraFieldJson msg
        , test "JSON encode empty message" <| \_ -> equal emptyJson (JE.encode 2 (T.fooEncoder fooDefault))
        , test "JSON decode empty JSON" <| \_ -> assertDecode T.simpleDecoder emptyJson msgDefault
        , test "JSON encode message with repeated field" <| \_ -> equal (JE.encode 2 (T.fooEncoder foo)) fooJson
        , describe "oneof"
            [ test "encode" <| \_ -> equal fooJson (JE.encode 2 (T.fooEncoder foo))
            , test "decode empty JSON" <| \_ -> assertDecode T.fooDecoder emptyJson fooDefault
            , test "decode oo1" <| \_ -> assertDecode T.fooDecoder oo1SetJson oo1Set
            , test "decode oo2" <| \_ -> assertDecode T.fooDecoder oo2SetJson oo2Set
            ]
        ]


assertDecode : JD.Decoder a -> String -> a -> Expectation
assertDecode decoder json msg =
    equal
        (JD.decodeString decoder json)
        (Result.Ok msg)


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
    , otherField = Nothing
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
  "oo1": 1
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
