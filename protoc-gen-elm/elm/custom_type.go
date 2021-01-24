package elm

import (
	"fmt"
	"strings"
	"text/template"

	"google.golang.org/protobuf/types/descriptorpb"
)

// CustomType - defines an Elm custom type (sometimes called union type)
// https://guide.elm-lang.org/types/custom_types.html
type CustomType struct {
	Name                   Type
	Decoder                VariableName
	Encoder                VariableName
	DefaultVariantVariable VariableName
	DefaultVariantValue    VariantName
	Variants               []CustomTypeVariant
}

// VariantName - unique camelcase identifier used for custom type variants
// https://guide.elm-lang.org/types/custom_types.html
type VariantName string

// VariantJSONName - unique JSON identifier, uppercase snake case, for a custom type variant
type VariantJSONName string

// CustomTypeVariant - a possible variant of a CustomType
// https://guide.elm-lang.org/types/custom_types.html
type CustomTypeVariant struct {
	Name     VariantName
	Number   ProtobufFieldNumber
	JSONName VariantJSONName
}

// NestedType - top level Elm type for a possibly nested PB definition
func NestedType(name string, preface []string) Type {
	fullName := name
	for _, p := range preface {
		fullName = fmt.Sprintf("%s_%s", p, fullName)
	}

	return Type(fullName)
}

// NestedVariantName - Elm variant name for a possibly nested PB definition
func NestedVariantName(name string, preface []string) VariantName {
	fullName := camelCase(strings.ToLower(name))
	for _, p := range preface {
		fullName = fmt.Sprintf("%s_%s", camelCase(strings.ToLower(p)), fullName)
	}

	return VariantName(fullName)
}

// EnumDefaultVariantVariableName - convenient identifier for a enum custom types default variant
func EnumDefaultVariantVariableName(name string, preface []string) VariableName {
	fullName := name
	for _, p := range preface {
		fullName = fmt.Sprintf("%s_%s", p, fullName)
	}

	return VariableName(firstLower(fmt.Sprintf("%sDefault", fullName)))
}

// EnumVariantJSONName - JSON identifier for variant decoder/encoding
func EnumVariantJSONName(pb *descriptorpb.EnumValueDescriptorProto) VariantJSONName {
	return VariantJSONName(pb.GetName())
}

// OneOfDefaultVariantName - convenient identifier for a one of custom types default variant
func OneOfDefaultVariantName(name VariableName) VariableName {
	return VariableName(firstLower(fmt.Sprintf("%sUnspecified", name)))
}

// CustomTypeTemplate - defines templates for custom types
func CustomTypeTemplate(t *template.Template) (*template.Template, error) {
	return t.Parse(`
{{ define "custom-type" -}}
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
{{- end }}
`)
}
