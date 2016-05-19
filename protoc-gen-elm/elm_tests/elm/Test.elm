import Json.Decode as JD
import Json.Encode as JE
import Result
import String
import Task

import ElmTest exposing (..)

import Simple as T


main =
  runSuiteHtml tests


tests : Test
tests =
  suite "A Test Suite"
    [ test "JSON encode" (assertEqual (JE.encode 2 (T.simpleEncoder msg)) msgJson)
    , test "JSON decode" <| assertDecode T.simpleDecoder msgJson msg
    , test "JSON decode extra field" <| assertDecode T.simpleDecoder msgExtraFieldJson msg
    , test "JSON decode empty message" <| assertDecode T.simpleDecoder msgEmptyJson msgDefault
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
