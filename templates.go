package goagen_js

import (
	"fmt"
	"strings"
	"text/template"
)

func newTemplate(args ...string) (*template.Template, error) {
	tmpl := template.New("")
	tmpl.Funcs(template.FuncMap{
		"join": func(value []string, arg string) string {
			return strings.Join(value, arg)
		},
	})
	for _, t := range args {
		var err error
		tmpl, err = tmpl.Parse(t)
		if err != nil {
			return nil, fmt.Errorf("tmpl.Parse() failed: %+v, %v", t, err)
		}
	}
	return tmpl, nil
}

const headerT = `{{- define "header" -}}
syntax = "proto3";

package proto;

message Empty {
}

{{ end }}
`

const serviceT = `{{ define "service" -}}
service {{ .Name }} {
{{- range $i, $p := .Rpcs }}
  {{ $p }}
{{- end }}
}
{{ end }}

`

const messageT = `{{- define "message" -}}
message {{ .Name}} {
{{- range $i, $p := .Definition }}
  {{ $p }}
{{- end }}
}
{{ end }}
`
