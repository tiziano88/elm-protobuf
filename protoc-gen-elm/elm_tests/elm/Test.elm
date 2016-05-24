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
  suite "A Test Suite"
    [ test "JSON encode" (assertEqual (JE.encode 2 (T.simpleEncoder msg)) msgJson)
    , test "JSON decode" <| assertDecode T.simpleDecoder msgJson msg
    , test "JSON decode extra field" <| assertDecode T.simpleDecoder msgExtraFieldJson msg
    , test "JSON decode empty message" <| assertDecode T.simpleDecoder msgEmptyJson msgDefault
    , test "JSON encode message with repeated field" <| assertEqual (JE.encode 2 (T.fooEncoder foo)) fooJson
    ]


assertDecode : JD.Decoder a -> String -> a -> Assertion
assertDecode decoder json msg =
  assertEqual
    (Result.Ok msg)
    (JD.decodeString decoder json)


msg : T.Simple
msg =
  { int32Field = 123
  }


msgDefault : T.Simple
msgDefault =
  { int32Field = 0
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


msgEmptyJson = String.trim """
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
  , oo1 = 1
  , oo2 = False
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
  "oo1": 1,
  "oo2": false
}
"""
