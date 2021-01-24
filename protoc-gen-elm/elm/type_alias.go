package elm

import "text/template"

// TODO: group struct and functions for defining an Elm alias type

// TypeAliasTemplate - defines templates for type aliases
// TODO: Is the mixing of templates an issue?  Maybe this cannot be split out?
func TypeAliasTemplate(t *template.Template) (*template.Template, error) {
	return t.Parse(`
{{ define "type-alias" -}}
type alias {{ .Type }} =
    { {{ range $i, $v := .Fields }}
        {{- if $i }}, {{ end }}{{ .Name }} : {{ .Type }}{{ if .Number }} -- {{ .Number }}{{ end }}
    {{ end }}}
{{- range .OneOfs }}{{ template "oneof-type" . }}{{ end }}


{{ .DecoderName }} : JD.Decoder {{ .Type }}
{{ .DecoderName }} =
    JD.lazy <| \_ -> decode {{ .Type }}{{ range .Fields }}
        |> {{ .Decoder.Preface }}
            {{- if .JSONName }} "{{ .JSONName }}"{{ end }} {{ .Decoder.Name }}
            {{- if .Decoder.HasDefaultValue }} {{ .Decoder.DefaultValue }}{{ end }}
        {{- end }}


{{ .EncoderName }} : {{ .Type }} -> JE.Value
{{ .EncoderName }} v =
    JE.object <| List.filterMap identity <|
        [{{ range $i, $v := .Fields }}
            {{- if $i }},{{ end }} ({{ .Encoder }})
        {{ end }}]
{{- range .NestedCustomTypes }}


{{ template "custom-type" . }}{{ end }}
{{- range .NestedMessages }}


{{ template "type-alias" . }}
{{- end }}
{{- end }}
`)
}
