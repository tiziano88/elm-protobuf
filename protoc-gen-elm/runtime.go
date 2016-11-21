package main

func (fg *FileGenerator) GenerateRuntime() {
	// Applicative-style decoders. This is fine as long as this is the only Applicative in the
	// package, otherwise operator will clash, since Elm does not have support to generalise
	// them via HKTs.

	fg.P("")
	fg.P("")

	fg.P("(<$>) : (a -> b) -> JD.Decoder a -> JD.Decoder b")
	fg.P("(<$>) =")
	fg.In()
	fg.P("JD.map")
	fg.Out()

	fg.P("")
	fg.P("")

	fg.P("(<*>) : JD.Decoder (a -> b) -> JD.Decoder a -> JD.Decoder b")
	fg.P("(<*>) f v =")
	fg.In()
	fg.P("f |> JD.andThen (\\x -> x <$> v)")
	fg.Out()

	fg.P("")
	fg.P("")

	fg.P("optionalDecoder : JD.Decoder a -> JD.Decoder (Maybe a)")
	fg.P("optionalDecoder decoder =")
	fg.In()
	fg.P("JD.oneOf")
	fg.In()
	fg.P("[ JD.map Just decoder")
	fg.P(", JD.succeed Nothing")
	fg.P("]")
	fg.Out()
	fg.Out()

	fg.P("")
	fg.P("")

	fg.P("requiredFieldDecoder : String -> a -> JD.Decoder a -> JD.Decoder a")
	fg.P("requiredFieldDecoder name default decoder =")
	fg.In()
	fg.P("withDefault default (JD.field name decoder)")
	fg.Out()

	fg.P("")
	fg.P("")

	fg.P("optionalFieldDecoder : String -> JD.Decoder a -> JD.Decoder (Maybe a)")
	fg.P("optionalFieldDecoder name decoder =")
	fg.In()
	fg.P("optionalDecoder (JD.field name decoder)")
	fg.Out()

	fg.P("")
	fg.P("")

	fg.P("repeatedFieldDecoder : String -> JD.Decoder a -> JD.Decoder (List a)")
	fg.P("repeatedFieldDecoder name decoder =")
	fg.In()
	fg.P("withDefault [] (JD.field name (JD.list decoder))")
	fg.Out()

	fg.P("")
	fg.P("")

	fg.P("withDefault : a -> JD.Decoder a -> JD.Decoder a")
	fg.P("withDefault default decoder =")
	fg.In()
	fg.P("JD.oneOf")
	fg.In()
	fg.P("[ decoder")
	fg.P(", JD.succeed default")
	fg.P("]")
	fg.Out()
	fg.Out()

	fg.P("")
	fg.P("")

	fg.P("optionalEncoder : String -> (a -> JE.Value) -> Maybe a -> Maybe (String, JE.Value)")
	fg.P("optionalEncoder name encoder v =")
	fg.In()
	fg.P("case v of")
	fg.In()
	fg.P("Just x ->")
	fg.In()
	fg.P("Just ( name, encoder x )")
	fg.Out()
	fg.P("")
	fg.P("Nothing ->")
	fg.In()
	fg.P("Nothing")
	fg.Out()
	fg.Out()
	fg.Out()

	fg.P("")
	fg.P("")

	fg.P("requiredFieldEncoder : String -> (a -> JE.Value) -> a -> a -> Maybe ( String, JE.Value )")
	fg.P("requiredFieldEncoder name encoder default v =")
	fg.In()
	fg.P("if v == default then")
	fg.In()
	fg.P("Nothing")
	fg.Out()
	fg.P("else")
	fg.In()
	fg.P("Just ( name, encoder v )")
	fg.Out()
	fg.Out()

	fg.P("")
	fg.P("")

	fg.P("repeatedFieldEncoder : String -> (a -> JE.Value) -> List a -> Maybe (String, JE.Value)")
	fg.P("repeatedFieldEncoder name encoder v =")
	fg.In()
	fg.P("case v of")
	fg.In()
	fg.P("[] ->")
	fg.In()
	fg.P("Nothing")
	fg.Out()
	fg.P("_ ->")
	fg.In()
	fg.P("Just (name, JE.list <| List.map encoder v)")
	fg.Out()
	fg.Out()
	fg.Out()
}
