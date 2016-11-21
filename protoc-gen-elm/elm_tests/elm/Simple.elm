module Simple exposing (..)


import Json.Decode as JD
import Json.Encode as JE


(<$>) : (a -> b) -> JD.Decoder a -> JD.Decoder b
(<$>) =
    JD.map


(<*>) : JD.Decoder (a -> b) -> JD.Decoder a -> JD.Decoder b
(<*>) f v =
    f |> JD.andThen (\x -> x <$> v)


optionalDecoder : JD.Decoder a -> JD.Decoder (Maybe a)
optionalDecoder decoder =
    JD.oneOf
        [ JD.map Just decoder
        , JD.succeed Nothing
        ]


requiredFieldDecoder : String -> a -> JD.Decoder a -> JD.Decoder a
requiredFieldDecoder name default decoder =
    withDefault default (JD.field name decoder)


optionalFieldDecoder : String -> JD.Decoder a -> JD.Decoder (Maybe a)
optionalFieldDecoder name decoder =
    optionalDecoder (JD.field name decoder)


repeatedFieldDecoder : String -> JD.Decoder a -> JD.Decoder (List a)
repeatedFieldDecoder name decoder =
    withDefault [] (JD.field name (JD.list decoder))


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
            Just ( name, encoder x )

        Nothing ->
            Nothing


requiredFieldEncoder : String -> (a -> JE.Value) -> a -> a -> Maybe ( String, JE.Value )
requiredFieldEncoder name encoder default v =
    if v == default then
        Nothing
    else
        Just ( name, encoder v )


repeatedFieldEncoder : String -> (a -> JE.Value) -> List a -> Maybe (String, JE.Value)
repeatedFieldEncoder name encoder v =
    case v of
        [] ->
            Nothing
        _ ->
            Just (name, JE.list <| List.map encoder v)


type Colour
    = ColourUnspecified -- 0
    | Red -- 1
    | Green -- 2
    | Blue -- 3


colourDecoder : JD.Decoder Colour
colourDecoder =
    let
        lookup s =
            case s of
                "COLOUR_UNSPECIFIED" ->
                    ColourUnspecified

                "RED" ->
                    Red

                "GREEN" ->
                    Green

                "BLUE" ->
                    Blue

                _ ->
                    ColourUnspecified
    in
        JD.map lookup JD.string


colourDefault : Colour
colourDefault = ColourUnspecified


colourEncoder : Colour -> JE.Value
colourEncoder v =
    let
        lookup s =
            case s of
                ColourUnspecified ->
                    "COLOUR_UNSPECIFIED"

                Red ->
                    "RED"

                Green ->
                    "GREEN"

                Blue ->
                    "BLUE"

    in
        JE.string <| lookup v


type alias Simple =
    { int32Field : Int -- 1
    }


simpleDecoder : JD.Decoder Simple
simpleDecoder =
    JD.lazy <| \_ -> Simple
        <$> (requiredFieldDecoder "int32Field" 0 JD.int)


simpleEncoder : Simple -> JE.Value
simpleEncoder v =
    JE.object <| List.filterMap identity <|
        [ (requiredFieldEncoder "int32Field" JE.int 0 v.int32Field)
        ]


type alias Foo =
    { s : Maybe Simple -- 1
    , ss : List Simple -- 2
    , colour : Colour -- 3
    , colours : List Colour -- 4
    , singleIntField : Int -- 5
    , repeatedIntField : List Int -- 6
    , oo : Oo
    }


type Oo
    = OoUnspecified
    | Oo1 Int
    | Oo2 Bool


ooDecoder : JD.Decoder Oo
ooDecoder =
    JD.lazy <| \_ -> JD.oneOf
        [ JD.map Oo1 (JD.field "oo1" JD.int)
        , JD.map Oo2 (JD.field "oo2" JD.bool)
        , JD.succeed OoUnspecified
        ]


ooEncoder : Oo -> Maybe ( String, JE.Value )
ooEncoder v =
    case v of
        OoUnspecified -> Nothing
        Oo1 x ->
            Just ( "oo1", JE.int x )
        Oo2 x ->
            Just ( "oo2", JE.bool x )


fooDecoder : JD.Decoder Foo
fooDecoder =
    JD.lazy <| \_ -> Foo
        <$> (optionalFieldDecoder "s" simpleDecoder)
        <*> (repeatedFieldDecoder "ss" simpleDecoder)
        <*> (requiredFieldDecoder "colour" colourDefault colourDecoder)
        <*> (repeatedFieldDecoder "colours" colourDecoder)
        <*> (requiredFieldDecoder "singleIntField" 0 JD.int)
        <*> (repeatedFieldDecoder "repeatedIntField" JD.int)
        <*> ooDecoder


fooEncoder : Foo -> JE.Value
fooEncoder v =
    JE.object <| List.filterMap identity <|
        [ (optionalEncoder "s" simpleEncoder v.s)
        , (repeatedFieldEncoder "ss" simpleEncoder v.ss)
        , (requiredFieldEncoder "colour" colourEncoder colourDefault v.colour)
        , (repeatedFieldEncoder "colours" colourEncoder v.colours)
        , (requiredFieldEncoder "singleIntField" JE.int 0 v.singleIntField)
        , (repeatedFieldEncoder "repeatedIntField" JE.int v.repeatedIntField)
        , (ooEncoder v.oo)
        ]
