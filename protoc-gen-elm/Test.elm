import Json.Decode as JD
import Json.Encode as JE
import Result
import String
import Task

import Console
import ElmTest exposing (..)

import Repeated as T


tests : Test
tests =
    suite "A Test Suite"
        [ test "JSON encode" (assertEqual (JE.encode 2 (T.subMessageEncoder msg)) msgJson)
        , test "JSON decode" (assertEqual (JD.decodeString T.subMessageDecoder msgJson) (Result.Ok msg))
        ]

msg : T.SubMessage
msg =
  { int32Field = 123
  }

msgJson : String
msgJson = String.trim """
{
  "int32Field": 123
}
"""


port runner : Signal (Task.Task x ())
port runner =
    Console.run (consoleRunner tests)
