package elm

import (
	"fmt"
	"protoc-gen-elm/pkg/stringextras"
	"strings"
	"text/template"

	"google.golang.org/protobuf/types/descriptorpb"
)

// EnumCustomType - defines an Elm custom type (sometimes called union type) for a PB enum
// https://guide.elm-lang.org/types/custom_types.html
type EnumCustomType struct {
	Name                   Type
	Decoder                VariableName
	Encoder                VariableName
	DefaultVariantVariable VariableName
	DefaultVariantValue    VariantName
	Variants               []EnumVariant
}

// VariantName - unique camelcase identifier used for custom type variants
// https://guide.elm-lang.org/types/custom_types.html
type VariantName string

// EnumVariant - a possible variant of an enum CustomType
// https://guide.elm-lang.org/types/custom_types.html
type EnumVariant struct {
	Name     VariantName
	Number   ProtobufFieldNumber
	JSONName VariantJSONName
}

// OneOfCustomType - defines an Elm custom type (sometimes called union type) for a PB one-of
// https://guide.elm-lang.org/types/custom_types.html
type OneOfCustomType struct {
	Name     Type
	Decoder  VariableName
	Encoder  VariableName
	Variants []OneOfVariant
}

// OneOfVariant - a possible variant of a one-of CustomType
// https://guide.elm-lang.org/types/custom_types.html
type OneOfVariant struct {
	Name     VariantName
	Type     Type
	JSONName VariantJSONName
	Decoder  VariableName
	Encoder  VariableName
}

// NestedVariantName - Elm variant name for a possibly nested PB definition
func NestedVariantName(name string, preface []string) VariantName {
	fullName := stringextras.CamelCase(strings.ToLower(name))
	for _, p := range preface {
		fullName = fmt.Sprintf("%s_%s", stringextras.CamelCase(p), fullName)
	}

	return VariantName(fullName)
}

// EnumDefaultVariantVariableName - convenient identifier for a enum custom types default variant
func EnumDefaultVariantVariableName(t Type) VariableName {
	return VariableName(stringextras.FirstLower(fmt.Sprintf("%sDefault", t)))
}

// EnumVariantJSONName - JSON identifier for variant decoder/encoding
func EnumVariantJSONName(pb *descriptorpb.EnumValueDescriptorProto) VariantJSONName {
	return VariantJSONName(pb.GetName())
}

// OneOfVariantJSONName - JSON identifier for variant decoder/encoding
func OneOfVariantJSONName(pb *descriptorpb.FieldDescriptorProto) VariantJSONName {
	return VariantJSONName(pb.GetJsonName())
}

// EnumCustomTypeTemplate - defines template for an enum custom type
func EnumCustomTypeTemplate(t *template.Template) (*template.Template, error) {
	return t.Parse(`
{{- define "enum-custom-type" -}}
type {{ .Name }}
{{- range $i, $v := .Variants }}
    {{ if not $i }}={{ else }}|{{ end }} {{ $v.Name }} -- {{ $v.Number }}
{{- end }}


{{ .Decoder }} : JD.Decoder {{ .Name }}
{{ .Decoder }} =
    let
        lookup s =
            case s of
{{- range .Variants }}
                "{{ .JSONName }}" ->
                    {{ .Name }}
{{ end }}
                _ ->
                    {{ .DefaultVariantValue }}
    in
        JD.map lookup JD.string


{{ .DefaultVariantVariable }} : {{ .Name }}
{{ .DefaultVariantVariable }} = {{ .DefaultVariantValue }}


{{ .Encoder }} : {{ .Name }} -> JE.Value
{{ .Encoder }} v =
    let
        lookup s =
            case s of
{{- range .Variants }}
                {{ .Name }} ->
                    "{{ .JSONName }}"
{{ end }}
    in
        JE.string <| lookup v
{{- end -}}
`)
}

// OneOfCustomTypeTemplate - defines template for a one-of custom type
func OneOfCustomTypeTemplate(t *template.Template) (*template.Template, error) {
	return t.Parse(`
{{- define "oneof-custom-type" -}}
type {{ .Name }}
    = {{ .Name }}Unspecified
{{- range .Variants }}
    | {{ .Name }} {{ .Type }}
{{- end }}


{{ .Decoder }} : JD.Decoder {{ .Name }}
{{ .Decoder }} =
    JD.lazy <| \_ -> JD.oneOf
        [{{ range $i, $v := .Variants }}{{ if $i }},{{ end }} JD.map {{ .Name }} (JD.field "{{ .JSONName }}" {{ .Decoder }})
        {{ end }}, JD.succeed {{ .Name }}Unspecified
        ]


{{ .Encoder }} : {{ .Name }} -> Maybe ( String, JE.Value )
{{ .Encoder }} v =
    case v of
        {{ .Name }}Unspecified ->
            Nothing
        {{- range .Variants }}

        {{ .Name }} x ->
            Just ( "{{ .JSONName }}", {{ .Encoder }} x )
        {{- end }}
{{- end -}}
`)
}
