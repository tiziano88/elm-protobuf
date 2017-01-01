module Protobuf exposing (..)

{-| Runtime library for Google Protocol Buffers.

# Operators

@docs (<$>), (<*>)

# Decoder Helpers

@docs requiredFieldDecoder, optionalFieldDecoder, repeatedFieldDecoder, bytesFieldDecoder

@docs withDefault

# Encoder Helpers

@docs requiredFieldEncoder, optionalEncoder, repeatedFieldEncoder, bytesFieldEncoder

-}

import Json.Decode as JD
import Json.Encode as JE


{-| Applicative initial application.
-}
(<$>) : (a -> b) -> JD.Decoder a -> JD.Decoder b
(<$>) =
    JD.map


{-| Applicative continued application.
-}
(<*>) : JD.Decoder (a -> b) -> JD.Decoder a -> JD.Decoder b
(<*>) f v =
    f |> JD.andThen (\x -> x <$> v)


{-| Decodes a required field.
-}
requiredFieldDecoder : String -> a -> JD.Decoder a -> JD.Decoder a
requiredFieldDecoder name default decoder =
    withDefault default (JD.field name decoder)


{-| Decodes an optional field.
-}
optionalFieldDecoder : String -> JD.Decoder a -> JD.Decoder (Maybe a)
optionalFieldDecoder name decoder =
    JD.maybe (JD.field name decoder)


{-| Decodes an repeated field.
-}
repeatedFieldDecoder : String -> JD.Decoder a -> JD.Decoder (List a)
repeatedFieldDecoder name decoder =
    withDefault [] (JD.field name (JD.list decoder))


{-| Provides a default value for a field.
-}
withDefault : a -> JD.Decoder a -> JD.Decoder a
withDefault default decoder =
    JD.oneOf
        [ decoder
        , JD.succeed default
        ]


{-| Encodes an optional field.
-}
optionalEncoder : String -> (a -> JE.Value) -> Maybe a -> Maybe ( String, JE.Value )
optionalEncoder name encoder v =
    case v of
        Just x ->
            Just ( name, encoder x )

        Nothing ->
            Nothing


{-| Encodes a required field.
-}
requiredFieldEncoder : String -> (a -> JE.Value) -> a -> a -> Maybe ( String, JE.Value )
requiredFieldEncoder name encoder default v =
    if v == default then
        Nothing
    else
        Just ( name, encoder v )


{-| Encodes a repeated field.
-}
repeatedFieldEncoder : String -> (a -> JE.Value) -> List a -> Maybe ( String, JE.Value )
repeatedFieldEncoder name encoder v =
    case v of
        [] ->
            Nothing

        _ ->
            Just ( name, JE.list <| List.map encoder v )


{-| Decodes a bytes field.
TODO: Implement.
-}
bytesFieldDecoder : JD.Decoder (List Int)
bytesFieldDecoder =
    JD.succeed []


{-| Encodes a bytes field.
TODO: Implement.
-}
bytesFieldEncoder : List Int -> JE.Value
bytesFieldEncoder v =
    JE.list []
