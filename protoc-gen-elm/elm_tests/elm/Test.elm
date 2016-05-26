import Json.Decode as JD
import Json.Encode as JE
import Result
import String
import Task

import ElmTest exposing (..)

import Simple as T


main =
  runSuite tests


tests : Test
tests =
  suite "JSON"
    [ test "JSON encode" <| assertEqual msgJson (JE.encode 2 (T.simpleEncoder msg))
    , test "JSON decode" <| assertDecode T.simpleDecoder msgJson msg
    , test "JSON decode extra field" <| assertDecode T.simpleDecoder msgExtraFieldJson msg
    , test "JSON decode empty JSON" <| assertDecode T.simpleDecoder emptyJson msgDefault
    , test "JSON encode message with repeated field" <| assertEqual (JE.encode 2 (T.fooEncoder foo)) fooJson
    , suite "oneof"
      [ test "encode" <| assertEqual fooJson (JE.encode 2 (T.fooEncoder foo))
      , test "decode empty JSON" <| assertDecode T.fooDecoder emptyJson fooDefault
      , test "decode oo1" <| assertDecode T.fooDecoder oo1SetJson oo1Set
      , test "decode oo2" <| assertDecode T.fooDecoder oo2SetJson oo2Set
      ]
    ]


assertDecode : JD.Decoder a -> String -> a -> Assertion
assertDecode decoder json msg =
  assertEqual
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
  }


msgJson : String
msgJson = String.trim """
{
  "int32Field": 123
}
"""


msgExtraFieldJson : String
msgExtraFieldJson = String.trim """
{
  "int32Field": 123,
  "extraField": "abc"
}
"""


emptyJson = String.trim """
{
}
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
  }


fooJson : String
fooJson = String.trim """
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
oo1SetJson = String.trim """
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
oo2SetJson = String.trim """
{
  "oo2": true
}
"""
