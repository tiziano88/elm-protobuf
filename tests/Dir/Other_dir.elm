module Dir.Other_dir exposing (..)

-- DO NOT EDIT
-- AUTOGENERATED BY THE ELM PROTOCOL BUFFER COMPILER
-- https://github.com/tiziano88/elm-protobuf
-- source file: dir/other_dir.proto

import Protobuf exposing (..)

import Json.Decode as JD
import Json.Encode as JE


uselessDeclarationToPreventErrorDueToEmptyOutputFile = 42


type alias OtherDir =
    { stringField : String -- 1
    }


otherDirDecoder : JD.Decoder OtherDir
otherDirDecoder =
    JD.lazy <| \_ -> decode OtherDir
        |> required "stringField" JD.string ""


otherDirEncoder : OtherDir -> JE.Value
otherDirEncoder v =
    JE.object <| List.filterMap identity <|
        [ (requiredFieldEncoder "stringField" JE.string "" v.stringField)
        ]
